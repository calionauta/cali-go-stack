package queue

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"maragu.dev/goqite"
)

type WorkerPool struct {
	q      *goqite.Queue
	hub    *SSEHub
	count  int
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewWorkerPool(q *goqite.Queue, hub *SSEHub, count int) *WorkerPool {
	return &WorkerPool{
		q:      q,
		hub:    hub,
		count:  count,
		stopCh: make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	for i := range wp.count {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	slog.Info("queue workers started", "count", wp.count)
}

func (wp *WorkerPool) Stop() {
	close(wp.stopCh)
	wp.wg.Wait()
	slog.Info("queue workers stopped")
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	ctx := context.Background()

	for {
		select {
		case <-wp.stopCh:
			return
		default:
		}

		msg, err := wp.q.Receive(ctx)
		if err != nil {
			select {
			case <-wp.stopCh:
				return
			default:
				slog.Warn("queue worker: receive error", "worker_id", id, "error", err)
				time.Sleep(time.Second)
				continue
			}
		}

		wp.processMessage(msg)

		if err := wp.q.Delete(ctx, msg.ID); err != nil {
			slog.Warn("queue worker: delete error", "worker_id", id, "error", err)
		}
	}
}

func (wp *WorkerPool) processMessage(msg *goqite.Message) {
	wp.hub.Broadcast(msg.Body)
}
