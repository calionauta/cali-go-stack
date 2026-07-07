package main

import (
	"fmt"
	"log"
	"os"

	"github.com/calionauta/cali-go-stack/config"
	"github.com/calionauta/cali-go-stack/db"
	"github.com/calionauta/cali-go-stack/internal/queue"
	"github.com/calionauta/cali-go-stack/router"
)

func main() {
	cfg := config.Load()

	// Initialize PocketBase (database, auth, REST, file storage)
	pb, err := db.Init(cfg)
	if err != nil {
		log.Fatalf("PocketBase init: %v", err)
	}

	// Seed default collections and data
	db.SeedDefaults(pb)

	// Initialize goqite task queue with SSE Hub
	q, err := queue.New(cfg)
	if err != nil {
		log.Fatalf("queue init: %v", err)
	}
	defer q.Close()

	// Start background workers
	q.StartWorkers()

	// Start NATS JetStream (only with -tags jetstream)
	startNATS(cfg)

	// Register routes on PocketBase's serve event
	router.Init(pb, q, cfg)

	// Configure PocketBase's HTTP address via command args
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	os.Args = append([]string{os.Args[0]}, "--http", addr)

	log.Printf("listening on %s", addr)
	if err := pb.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
