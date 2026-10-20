# AI review (sukra)

_Generated 2026-10-20 by claude-opus-5-5 from a snapshot of tracked sources; verify before acting._

# Code review: Shukra (`sukra`)

## Summary

Shukra is a host-side observability and enforcement daemon for KVM hypervisors, written in Go with eBPF programs in C. The daemon is `shukrad` and the CLI is `shukractl`. It attaches eBPF programs (kvm, sched, block, net, mem, vmm, drops, tap) and attributes traffic to a VM only when it is seen on that VM's tap. It can isolate a VM or enforce a per-VM egress policy through TCX, and it serves an authenticated HTTP API and a React console. The engineering is unusually careful: bounded tables, fail-closed enforcement, sanitised guest text, and atomic persistence. Most findings below are edge cases.

## Bugs and correctness risks

**1. One crash hides all later snapshots in a file (`internal/persist/rollup.go`, `internal/persist/log.go`).**
- In `Rollup.scan`, the first record that fails to decode `break`s out of the file, on the assumption that a torn line can only be the last one.
- But `Log.open` deliberately terminates a torn last line with `\n` and keeps appending after it.
- So after one crash mid-write, every snapshot appended afterwards is invisible to `Around` until the file rolls.
- Effect: `explain --at` silently loses history.
- Fix: skip the bad line (scan line by line, as `ReadTail` does) instead of stopping.

**2. Possible nil-map panic (`internal/agent/coverage.go`).**
- `recordAttach` initialises `a.covered` but not `a.seen`, then writes `a.seen[tap] = now` when a tap stops being ready.
- `a.seen` is only created in `noteSeen`.
- If `recordAttach` ever runs before `noteSeen`, the daemon panics. The call order is not visible in the reviewed files.
- These maps also have no visible lock. Check that they are only touched from one goroutine.

**3. Lock-convention violation (`internal/state/contention.go`).**
- `Contention` calls `s.cpuVendorLocked()` without taking `s.mu`.
- Elsewhere in the package, the `…Locked` suffix means "caller holds s.mu" (see `recordLocked`, `baseLocked`).
- This is a likely data race, unless that function is misnamed.
- Separately, `out.Window` is overwritten per VM. If some VMs have windowed history and others fall back to lifetime, the label reflects whichever VM came last.

**4. Hand-written tracepoint layouts are not CO-RE (`bpf/kvm.bpf.c`, `bpf/net.bpf.c`).**
- `shukra_kvm_exit` assumes `exit_reason` is the first payload field.
- `shukra_tcp_retrans` hard-codes the 6.8 layout.
- The drops program was moved to CO-RE precisely because Linux 6.9 moved fields. These two have no equivalent format check like `checkKfreeSkbFormat`.
- On arm64, `kvm_exit`'s first field is not an x86-style exit reason. Per-reason counts would be meaningless, and the HLT exclusion (12 / 120) could drop real exits from the latency histogram.
- Recommendation: validate against `/sys/kernel/tracing/events/.../format` at attach time. I have not verified the exact field order on each kernel.

**5. Fixed temp-file name (`internal/persist/log.go`, `WriteFileAtomic`).**
- It always uses `path + ".tmp"`.
- Two concurrent writers to the same path would truncate each other's temp file.
- It is safe only if every caller (policies, baselines, bundles, recorder snapshot) serialises its own writes. That is not enforced here.

**6. `Recorder.Window("")` is not globally ordered (`internal/recorder/recorder.go`).**
- With `vm` empty or `"_"`, it concatenates per-VM rings in map order.
- The doc comment says "oldest first", but that is only true within one VM.

**7. Smaller issues.**
- **`internal/persist/log.go` (`ReadTail`):** a line over 1 MiB makes the scanner stop, and every later line in that file is lost.
- **`cmd/shukractl/install_cli.go`:** leaves `shukractl.tmp` behind on a copy or rename failure.
- **`internal/persist/policy.go` (`OpenPolicies`):** ignores a failed rename to `.corrupt`, so the next `Save` overwrites the evidence. Each new corrupt file also replaces the previous `.corrupt`.
- **`web/src/pages/Security.tsx` (`Detections`):** keys on `d.ts + d.dst`, where `dst` is often undefined. Two detections in the same second collide.
- **`internal/bpfgen/drops.go`:** the comment on `tryDrops` still describes the old field-layout check ("reading a moved field would count nonsense"), which no longer matches the CO-RE program.

## Security concerns

**`cmd/shukractl/client.go`**
- `loadConfig` silently sets `SHUKRA_TLS_INSECURE=true` for any `https://` loopback URL, which disables certificate verification. That is reasonable on loopback, but it is surprising and undocumented in `help.go`.
- If `SHUKRA_CA_FILE` cannot be read, `httpClient` silently falls back to the system roots instead of failing.
- The admin bearer is sent to any `SHUKRA_URL`, including plain `http://` to a remote host, with no warning.
- `~/.shukra/env` can set arbitrary environment variables, including `SHUKRA_URL` and `SHUKRA_TLS_INSECURE`. This is low risk, since it is the user's own file.

**`web/src/components/Login.tsx`**
- `autoComplete="current-password"` invites the browser's password manager to store the admin API token.
- That undercuts the "kept in memory only" stance in `api.ts`. Consider `autoComplete="off"` or `"one-time-code"`.

**`deploy/install.sh`**
- The generated admin key is printed to stdout. It will land in CI logs or the logs that `deploy-remote.sh` or the `.deb` produce.
- It is documented, but consider writing it only to `/etc/shukra/env` and telling the operator where to find it.

**`internal/api/who.go` and `internal/state/caller.go`**
- The actor label is client-supplied (`X-Shukra-Actor`). This is documented as not a proof, which is good.
- The audit `Principal()` shows `keyID (label)`. Make sure UIs do not render the label as an identity.

**`cmd/shukrad/listen_test.go`**
- `localhost` is treated as loopback.
- If `/etc/hosts` maps it elsewhere, the plain-HTTP guard is bypassed. This is an edge case.

## Test and CI gaps

- **`internal/api/emptylists_test.go`:** `continue`s on any non-200 response or non-JSON body. A route that starts erroring, or a broken route, silently passes. List the routes that are expected to return non-200 explicitly.
- **`internal/api/fuzz_test.go`:** `FuzzAPIJSON` only unmarshals into two structs. It never exercises a handler, auth, or validation, so it adds little. Fuzz the handlers through `httptest`.
- **`internal/observe/tap_size_test.go`:** `TestTapRecordIs56Bytes` compares a constant to a literal and protects nothing. The real guard is the C `_Static_assert`.
- **No test covers:**
  - a torn line in the middle of `snapshots.jsonl` (bug 1);
  - `recordAttach` being called before `noteSeen` (b
