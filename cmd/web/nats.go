//go:build jetstream

package main

import (
	"log"

	"github.com/calionauta/cali-go-stack/config"
	"github.com/calionauta/cali-go-stack/internal/nats"
)

func startNATS(cfg *config.Config) {
	if !cfg.NATS.Enabled {
		return
	}
	if err := nats.StartEmbedded(cfg.NATS.StoreDir); err != nil {
		log.Fatalf("NATS startup: %v", err)
	}
}
