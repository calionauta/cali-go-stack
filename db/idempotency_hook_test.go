// SCOPE:plugin - E2E test for the (idem_key, owner) unique index
// installed by ensureTodosCollection.
//
// The OnRecordCreateRequest hook itself is exercised by the live
// site's manual offline-replay test (it intercepts HTTP API requests;
// direct app.Save calls bypass the hook and reach the DAO directly).
// This test verifies the underlying race-detection layer that protects
// against two concurrent requests racing the hook check.
package db

import (
	"os"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

// TestIdempotencyUniqueIndexRejectsDuplicates creates two records
// with the same (idem_key, owner) pair via the DAO. The first succeeds;
// the second is rejected by the unique index with a validation error
// mentioning `idem_key`. This proves the index is installed correctly
// and that concurrent or replayed POSTs cannot land a duplicate even
// if the hook check is bypassed.
func TestIdempotencyUniqueIndexRejectsDuplicates(t *testing.T) {
	app, demoUser, col := setupIdemTestApp(t)
	owner := demoUser.Id
	idemKey := "11111111-2222-3333-4444-555555555555"

	// First save: succeeds.
	first := core.NewRecord(col)
	first.Set("title", "first")
	first.Set("completed", false)
	first.Set("owner", owner)
	first.Set("idem_key", idemKey)
	if err := app.Save(first); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if first.Id == "" {
		t.Fatal("first save did not assign an id")
	}

	// Second save with the same (idem_key, owner): rejected by the
	// unique index. (Direct app.Save bypasses the OnRecordCreateRequest
	// hook — that's the HTTP-layer path. This test exercises the
	// DB-level race-detection layer that catches a hook bypass.)
	second := core.NewRecord(col)
	second.Set("title", "duplicate")
	second.Set("completed", false)
	second.Set("owner", owner)
	second.Set("idem_key", idemKey)
	err := app.Save(second)
	if err == nil {
		t.Fatal("expected unique-constraint violation on duplicate idem_key+owner, got nil")
	}
	if !strings.Contains(err.Error(), "idem_key") {
		t.Errorf("expected error to mention idem_key, got: %v", err)
	}

	// Original record is unchanged.
	existing, lookupErr := app.FindFirstRecordByFilter(
		"todos",
		"idem_key = {:k} && owner = {:o}",
		map[string]any{"k": idemKey, "o": owner},
	)
	if lookupErr != nil {
		t.Fatalf("lookup after duplicate: %v", lookupErr)
	}
	if existing == nil {
		t.Fatal("expected existing record to still be findable after rejected duplicate")
	}
	if existing.GetString("title") != "first" {
		t.Errorf("existing title = %q, want %q (duplicate must NOT have replaced it)",
			existing.GetString("title"), "first")
	}

	// Exactly one record with this (idem_key, owner) pair.
	all, listErr := app.FindRecordsByFilter(
		"todos",
		"idem_key = {:k} && owner = {:o}",
		"", 0, 0,
		map[string]any{"k": idemKey, "o": owner},
	)
	if listErr != nil {
		t.Fatalf("FindRecordsByFilter: %v", listErr)
	}
	if len(all) != 1 {
		t.Errorf("record count for (idem_key, owner) = %d, want 1", len(all))
	}
}

// setupIdemTestApp boots a fresh app, runs the seed + hook, and
// returns the demo user and todos collection. Extracted to keep
// TestIdempotencyUniqueIndexRejectsDuplicates under gocyclo's 12
// branch limit.
func setupIdemTestApp(t *testing.T) (core.App, *core.Record, *core.Collection) {
	t.Helper()
	tmpDir, mkErr := os.MkdirTemp("", "idem-*")
	if mkErr != nil {
		t.Fatal(mkErr)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	app := newSeedTestApp(t, tmpDir)
	if err := ensureTodosCollection(app, true); err != nil {
		t.Fatalf("ensureTodosCollection: %v", err)
	}
	if err := ensureDemoUser(app); err != nil {
		t.Fatalf("ensureDemoUser: %v", err)
	}
	RegisterIdempotencyHook(app)

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("users collection: %v", err)
	}
	demoUser, err := app.FindAuthRecordByEmail(usersCol.Name, DemoUserEmail)
	if err != nil || demoUser == nil {
		t.Fatalf("demo user lookup: err=%v user=%v", err, demoUser)
	}
	col, err := app.FindCollectionByNameOrId("todos")
	if err != nil {
		t.Fatalf("todos collection: %v", err)
	}
	return app, demoUser, col
}
