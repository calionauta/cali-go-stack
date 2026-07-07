//go:build jetstream

package nats

import (
	"log/slog"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var (
	NS *server.Server
	NC *nats.Conn
	JS nats.JetStreamContext
)

// StartEmbedded starts an embedded NATS server and connects to it.
func StartEmbedded(storeDir string) error {
	ns, err := server.NewServer(&server.Options{
		Port:     -1,
		NoLog:    true,
		NoSigs:   true,
		StoreDir: storeDir,
	})
	if err != nil {
		return err
	}
	ns.Start()
	NS = ns

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		return err
	}
	NC = nc

	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	JS = js

	slog.Info("NATS embedded server started", "url", ns.ClientURL())
	return nil
}

func Stop() {
	if NC != nil {
		NC.Close()
	}
	if NS != nil {
		NS.Shutdown()
	}
}

func ClientURL() string {
	if NS == nil {
		return ""
	}
	return NS.ClientURL()
}
