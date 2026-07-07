package queue

import (
	"context"
	"database/sql"
	"time"

	"github.com/calionauta/cali-go-stack/config"
	_ "github.com/ncruces/go-sqlite3/driver"
	"maragu.dev/goqite"
)

type Queue struct {
	q   *goqite.Queue
	hub *SSEHub
	db  *sql.DB
}

func New(cfg *config.Config) (*Queue, error) {
	dbPath := cfg.DataDir + "/queue.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: "default",
	})

	hub := NewSSEHub()

	return &Queue{
		q:   q,
		db:  db,
		hub: hub,
	}, nil
}

func (q *Queue) Enqueue(ctx context.Context, data []byte) error {
	return q.q.Send(ctx, goqite.Message{Body: data})
}

func (q *Queue) Receive(ctx context.Context) (*goqite.Message, error) {
	return q.q.Receive(ctx)
}

func (q *Queue) ReceiveAndWait(ctx context.Context, timeout time.Duration) (*goqite.Message, error) {
	return q.q.ReceiveAndWait(ctx, timeout)
}

func (q *Queue) Delete(ctx context.Context, id goqite.ID) error {
	return q.q.Delete(ctx, id)
}

func (q *Queue) Hub() *SSEHub {
	return q.hub
}

func (q *Queue) StartWorkers() {
	wp := NewWorkerPool(q.q, q.hub, 4)
	wp.Start()
}

func (q *Queue) Close() {
	q.q = nil
	q.db.Close()
}
