# AI security (sukra)

_Generated 2026-10-20 by claude-opus-5-5 from a snapshot of tracked sources; verify before acting._

# Security audit: `sukra` (Shukra, KVM eBPF observability and enforcement daemon)

## Scope and method

I reviewed only the files supplied. Several security-relevant files were **not** included:

- `bpf/tap.bpf.c`
- `bpf/vmm.bpf.c`
- `internal/api/server.go` (where the `auth` function lives)
- the QEMU cmdline and fd identity scan
- `decodeVMM` and `decodeTLS`
- `writeMetrics`
- `printEvent`
- the `spa` handler
- the response and policy engines
- the release workflow

Where a finding depends on code I could not see, I say so and mark it **[needs verification]**. Severities assume the threat model in `SECURITY.md`:

- guest: untrusted
- VMM: watched, not trusted
- read-only key: may read, never change
- local unprivileged host user: not explicitly modelled, but realistic on a k3s/Cilium node

---

## High

### H1. The dev admin key `shukra` is not refused on a non-loopback listener

**Severity:** High (configuration-dependent)

**Where:** `cmd/shukrad`, `refuseInsecureListen(addr, insecure, noAuth)`, as exercised by `cmd/shukrad/listen_test.go`.

**Evidence:**
- The function takes no key argument, so it cannot know whether the dev key is in use.
- `SECURITY.md` says that when `SHUKRA_API_KEY` is unset, the daemon uses `shukra`. That is documented as "anyone who can reach the port can … isolate any VM".
- The only guard is `shukractl doctor`, which runs after the fact and only if someone runs it.

**Attack:**
1. The operator runs `-listen 0.0.0.0:30970 -tls-cert … -tls-key …`, or the documented container command, without exporting `SHUKRA_API_KEY`. Because TLS is on, `refuseInsecureListen` never triggers.
2. The same happens with `-allow-insecure-http`.
3. Any network peer sends `Authorization: Bearer shukra` and calls `POST /api/v1/isolate`, `/policy/apply` or `/actions/{id}/approve` for any VM.
4. That peer may be a guest, if the host address is reachable from guest networks, and the guest is not on the isolation allow list.

**Fix:**
- Refuse to start whenever the listen address is not loopback and the effective admin key is the dev default, regardless of TLS or `-allow-insecure-http`.
- Refuse keys shorter than 16 characters on non-loopback addresses.
- Consider removing the implicit dev key entirely and requiring `-dev-key` explicitly.

### H2. Interface identity is not validated, so enforcement can land on host interfaces

**Severity:** High if confirmed; Medium as
