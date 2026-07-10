//go:build !jetstream

package router

import "github.com/pocketbase/pocketbase/core"

// registerCollabSync is a no-op when NATS/JetStream is not compiled in.
// The real implementation lives in collab_jetstream.go (build tag
// jetstream).
func registerCollabSync(_ *core.ServeEvent) {}
