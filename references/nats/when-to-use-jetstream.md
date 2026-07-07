# NATS vs JetStream: When to Use Each

## NATS Core (Pub/Sub)

**Use when:**
- Fire-and-forget messaging (no persistence needed)
- Simple broadcasts (one publisher to multiple subscribers)
- Low latency is critical
- No need for message history
- Lost messages are acceptable if no subscriber

```go
nc.Publish("events.users", data)
```

## JetStream (Persistence)

**Use when:**
- Need history (consumers can read old messages)
- Delivery guarantee (at-least-once or exactly-once)
- Work queues (multiple consumers process messages)
- Persistent streams (event log, audit trail)
- Message replay
- Horizontal scaling of consumers

```go
js, _ := nc.JetStream()
js.AddStream(&nats.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"events.>"},
    Storage:  nats.FileStorage,
})
```

## Quick Summary

| Feature | NATS Core | JetStream |
|---------|-----------|-----------|
| Persistence | No | Yes |
| History | No | Yes |
| Replay | No | Yes |
| Throughput | Higher | High |
| Latency | Lower | Low |
| Complexity | Lower | Higher |

## Decision Tree

```
Need messages after restart?
  NO → NATS Core
  YES → JetStream

Need Key-Value store?
  → JetStream Key-Value (wrapper over streams)
```
