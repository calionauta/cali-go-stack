//go:build !jetstream

package main

import "github.com/calionauta/cali-go-stack/config"

func startNATS(cfg *config.Config) {
	// NATS not available without -tags jetstream
	_ = cfg
}
