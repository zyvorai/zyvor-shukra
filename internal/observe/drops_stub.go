// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: LicenseRef-Zyvor-Production-2.0

//go:build !shukrabpf

package observe

// DropSample has nothing to read in a build without BPF.
func DropSample() []DropCounters { return nil }
