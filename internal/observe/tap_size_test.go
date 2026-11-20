// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: LicenseRef-Zyvor-Production-1.0

package observe

import "testing"

func TestTapRecordIs56Bytes(t *testing.T) {
	if tapEventSize != 56 {
		t.Fatalf("tap record %d", tapEventSize)
	}
}
