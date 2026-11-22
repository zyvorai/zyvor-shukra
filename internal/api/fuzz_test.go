// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: LicenseRef-Zyvor-Production-2.0

package api

import (
	"encoding/json"
	"testing"
)

func FuzzAPIJSON(f *testing.F) {
	f.Add([]byte(`{"vm":"web","mode":"audit"}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		var policy policyBody
		_ = json.Unmarshal(b, &policy)
		var vm struct {
			VM string `json:"vm"`
		}
		_ = json.Unmarshal(b, &vm)
	})
}
