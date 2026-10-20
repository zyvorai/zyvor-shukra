package agent

import (
	"testing"
	"time"

	"github.com/zyvorai/shukra/internal/identity"
	"github.com/zyvorai/shukra/internal/state"
)

// recordAttach must not panic when it runs before noteSeen has created a.seen.
func TestRecordAttachNilSeen(t *testing.T) {
	a := &Agent{State: state.New(""), covered: map[string]bool{"lo0": true, "lo": true}}
	a.recordAttach(time.Now(), []identity.VM{{Name: "vm", Taps: []string{"lo0", "lo"}}})
}
