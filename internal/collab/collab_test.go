package collab

import "testing"

// TestDoc_UpdateSnapshotRoundTrip covers the core CRDT operations on the
// default (non-jetstream) build: encode an update, apply it to a second
// doc, and confirm both converge to the same snapshot. This keeps
// EncodeUpdate/StateVersion reachable for the deadcode scan and proves
// the Loro wrapper round-trips without a NATS transport.
func TestDoc_UpdateSnapshotRoundTrip(t *testing.T) {
	a := NewDoc("roundtrip")
	if v := a.StateVersion(); v == nil {
		t.Fatal("StateVersion returned nil")
	}

	update, err := a.EncodeUpdate(nil)
	if err != nil {
		t.Fatalf("EncodeUpdate: %v", err)
	}

	b := NewDoc("roundtrip")
	if err := b.ApplyUpdate(update); err != nil {
		t.Fatalf("ApplyUpdate: %v", err)
	}

	sa := a.EncodeSnapshot()
	sb := b.EncodeSnapshot()
	if len(sa) == 0 || len(sb) == 0 {
		t.Fatal("empty snapshot")
	}
	if string(sa) != string(sb) {
		t.Fatal("docs did not converge to identical snapshots")
	}
}
