# SSE Hub Patterns

## Register Before Enqueue

```go
// 1. Client connects and registers
ch := make(chan []byte, 16)
hub.Register(clientID, ch)

// 2. Handler enqueues a job
q.Enqueue(ctx, jobData)

// 3. Worker sends result via hub
hub.Send(clientID, result)
```

## Replay Buffer

If a client registers after an event was already sent, the hub replays from a buffer.

```go
// Buffer up to 64 messages per client
const maxBuffer = 64
```

## Backpressure

If a client is slow (channel full), the hub drops the event instead of blocking.
