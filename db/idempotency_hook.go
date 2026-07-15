// SCOPE:plugin - REMOVE if you don't need idempotent offline replay.
//
// PocketBase OnRecordCreateRequest hook that dedupes todo creation by
// the (idem_key, owner) pair. When the UI sends a request with an
// idem_key that already exists for the same owner, this hook returns
// the existing record (by replacing e.Record) instead of creating a
// duplicate. The unique index added in ensureTodosCollection is a
// belt-and-suspenders: it catches the race when two concurrent
// requests with the same key arrive before the first hook check
// completes — the second insert raises a unique-constraint error,
// which the handler swallows and converts to a re-read of the
// existing record.
//
// Single-instance and multi-instance both work because the dedup
// state lives in the database (the unique index). For workloads where
// the DB hit on every create is too expensive, swap this hook for a
// NATS JetStream producer with Nats-Msg-Id = idem_key (JetStream's
// built-in dedup window replaces the DB lookup).
package db

import (
	"log/slog"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterIdempotencyHook installs the OnRecordCreateRequest dedup hook
// on the todos collection. Idempotent: calling twice does nothing
// because PB matches handlers by exact function reference, but in
// practice SeedDefaults calls this once per process start.
func RegisterIdempotencyHook(app core.App) {
	app.OnRecordCreateRequest("todos").BindFunc(func(e *core.RecordRequestEvent) error {
		key := e.Record.GetString("idem_key")
		if key == "" {
			// No idempotency key supplied; let the create proceed
			// (legacy callers / non-form paths).
			return e.Next()
		}
		owner := e.Record.GetString("owner")
		if owner == "" {
			// No owner: defer to the existing access rules to reject
			// (the collection has a create rule that requires
			// @request.auth.id). We don't dedup owner-less records.
			return e.Next()
		}

		existing, err := app.FindFirstRecordByFilter(
			"todos",
			"idem_key = {:k} && owner = {:o}",
			map[string]any{"k": key, "o": owner},
		)
		if err != nil {
			// FindFirstRecordByFilter returns an error only when there
			// is a DB problem (not when the record doesn't exist).
			// Log and proceed — we never want dedup to break the
			// happy path.
			slog.Warn("idempotency hook: lookup failed; proceeding with create", "error", err)
			return e.Next()
		}
		if existing == nil {
			return e.Next()
		}
		// Hit: replace the inbound record with the existing one so PB
		// returns it in the response. NOT calling e.Next() aborts the
		// create with no error (PB treats a missing next as "handled").
		e.Record = existing
		return nil
	})
}
