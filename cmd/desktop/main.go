// Command desktop builds the Wails v3 native app shell around the exact
// same Go backend the web app uses (PocketBase + goqite queue + router +
// handlers). It boots the server via internal/server.Run, serves
// PocketBase in a goroutine, and points the Wails webview at it through a
// reverse proxy — so the desktop app and the web app share 100% of the
// business logic.
//
// Phase A (this file): desktop shell only, no edge sync. Phase B adds an
// embedded NATS Leaf Node; Phase C adds the Loro CRDT collab layer. Those
// are wired in subsequent sessions and do not change this boot shape.
package main

import (
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/calionauta/gogogo-fullstack-template/config"
	"github.com/calionauta/gogogo-fullstack-template/internal/server"
)

func main() {
	cfg := config.Load()

	// Boot the shared backend (PocketBase + queue + router + handlers).
	// js is nil here (no realtime broadcaster in Phase A); the router
	// falls back to in-process SSE. Phase B passes a Leaf-Node JetStream.
	pb, _, shutdown, err := server.Run(cfg, nil)
	if err != nil {
		log.Fatalf("boot failed: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// PocketBase blocks on pb.Start(), so run it on a goroutine. The
	// Wails webview reaches it via the reverse proxy below.
	go func() {
		args := []string{"serve", "--http", addr}
		if len(os.Args) > 1 {
			args = os.Args[1:]
		}
		pb.RootCmd.SetArgs(args)
		if startErr := pb.Start(); startErr != nil {
			log.Fatalf("pocketbase start: %v", startErr)
		}
	}()

	// Reverse proxy: the Wails webview loads http://localhost:<port>
	// (where PocketBase serves the UI) through this handler.
	target, err := url.Parse(fmt.Sprintf("http://%s", addr))
	if err != nil {
		log.Fatalf("parse target url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	app := application.New(application.Options{
		Name: "Gogogo Template",
		Assets: application.AssetOptions{
			Handler: proxy,
		},
		// Shut the backend down when the Wails app closes (window quit),
		// not via defer — defer would be skipped on a fatal error and the
		// gocritic linter flags it. OnShutdown runs on a clean exit.
		OnShutdown: func() {
			shutdown()
		},
	})

	log.Printf("desktop: PocketBase on %s, webview proxied", addr)
	if runErr := app.Run(); runErr != nil {
		log.Fatalf("wails run: %v", runErr)
	}
}
