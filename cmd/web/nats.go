//go:build jetstream

package main

import (
	"log"

	"github.com/calionauta/gogogo-fullstack-template/config"
	"github.com/calionauta/gogogo-fullstack-template/internal/nats"
)

func startNATS(cfg *config.Config) {
	if !cfg.NATS.Enabled {
		return
	}
	if err := nats.StartEmbedded(cfg.NATS.StoreDir); err != nil {
		// Don't take the whole app down if embedded NATS can't start
		// (e.g. a read-only or full store dir). Fall back to the
		// in-memory broadcaster so realtime still works within the
		// instance.
		log.Printf("WARN: NATS startup failed, falling back to in-memory broadcaster: %v", err)
		return
	}
}
