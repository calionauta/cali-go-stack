// Package collab wraps the Loro CRDT (github.com/aholstenson/loro-go) for
// the template's collaborative features (whiteboard, shared docs). Loro
// gives conflict-free offline merges: two edges editing the same doc
// offline both converge when their updates meet, with no Last-Writer-Wins
// data loss.
//
// This package is the local CRDT model. The TRANSPORT (how updates reach
// the central server) is NATS Leaf Node (Phase B) — collab produces the
// bytes, the SyncWorker (sync.go) publishes/applies them over JetStream.
package collab

import (
	"sync"

	loro "github.com/aholstenson/loro-go"
)

// Doc is a single collaborative document (whiteboard, shared list, etc.)
// backed by a Loro CRDT. It is safe for concurrent use: all mutations go
// through the embedded mutex so the desktop UI thread and the sync
// publisher never race on the LoroDoc.
type Doc struct {
	mu   sync.Mutex
	id   string
	loro *loro.LoroDoc
}

// NewDoc creates an empty collaborative doc with the given ID (UUID v7 in
// practice). The ID is the JetStream subject key (app.sync.<id>) so
// updates route to the right doc on every peer.
func NewDoc(id string) *Doc {
	return &Doc{id: id, loro: loro.NewLoroDoc()}
}

// ID returns the document identifier.
func (d *Doc) ID() string {
	return d.id
}

// ApplyUpdate merges a Loro update (bytes from EncodeUpdate) into this
// doc. Concurrent-safe. Returns an error if the bytes are not a valid
// Loro update.
func (d *Doc) ApplyUpdate(update []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, err := d.loro.Import(update); err != nil {
		return err
	}
	return nil
}

// EncodeSnapshot returns a self-contained snapshot of the full doc state.
// Used to bootstrap a peer or to persist the resolved state to PocketBase.
func (d *Doc) EncodeSnapshot() []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := d.loro.Export(loro.SnapshotMode())
	if err != nil {
		// Export only errors on a malformed doc; with a freshly created
		// LoroDoc this is unreachable. Propagate so callers can log.
		panic(err)
	}
	return b
}

// EncodeUpdate returns the delta from the given peer version vector to
// the current state — i.e. "what changed since this peer last synced".
// Use an empty VersionVector (loro.NewVersionVector) to export all changes
// from the doc's origin (the common case for small collaborative docs).
func (d *Doc) EncodeUpdate(since *loro.VersionVector) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if since == nil {
		since = loro.NewVersionVector()
	}
	return d.loro.Export(loro.UpdatesMode(since))
}

// StateVersion returns this doc's current version vector, used by the
// sync layer to compute deltas for the next publish.
func (d *Doc) StateVersion() *loro.VersionVector {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.loro.StateVv()
}
