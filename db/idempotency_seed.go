// SCOPE:plugin - REMOVE if you don't need idempotent offline replay.
//
// Adds the (idem_key, owner) field + unique index to a collection so
// the OnRecordCreateRequest hook in db/idempotency_hook.go can dedup
// replays of the SW queue. Kept separate from ensureTodosCollection
// so removing the dedup feature is a one-file delete: delete
// idempotency_seed.go + idempotency_hook.go, drop the call below,
// and remove the hidden `name="idem_key"` input from the form.
package db

import (
	"log/slog"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// enableTodosIdempotency adds the `idem_key` field + unique index to
// the todos collection. Called from SeedDefaults; idempotent (no-op
// if the field/index already exist).
func enableTodosIdempotency(col *core.Collection) {
	if col.Fields.GetByName("idem_key") == nil {
		col.Fields.Add(&core.TextField{Name: "idem_key", Max: 100})
		slog.Info("seed: ensured todos.idem_key field")
	}
	hasIdemKeyIndex := false
	for _, sql := range col.Indexes {
		// Indexes are stored as raw SQL CREATE INDEX statements in
		// PocketBase; match by the index name we set below.
		if strings.Contains(sql, `"idem_key_owner_idx"`) {
			hasIdemKeyIndex = true
			break
		}
	}
	if !hasIdemKeyIndex {
		col.AddIndex("idem_key_owner_idx", true, "idem_key,owner", "")
	}
}
