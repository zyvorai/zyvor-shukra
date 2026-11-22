// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: LicenseRef-Zyvor-Production-2.0

//go:build !shukrabpf

package observe

import "testing"

func TestSampleStubIsNil(t *testing.T) {
	if Sample() != nil {
		t.Fatal("stub build invented counters")
	}
}
