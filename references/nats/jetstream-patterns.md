# JetStream Patterns: Multi-User Real-Time

Always compiled in the unified build; opt out via `NATS_ENABLED=false`.

## Embedded NATS Server

```go
// internal/nats/embedded.go
ns, _ := server.NewServer(&server.Options{Port: -1, NoLog: true, NoSigs: true})
ns.Start()
nc, _ := nats.Connect(ns.ClientURL())
js, _ := nc.JetStream()
```

## KV Store (Shared State)

```go
kv, _ := js.CreateKeyValue(&nats.KeyValueConfig{
    Bucket: "canvas-state", Storage: nats.FileStorage,
})
kv.Put("room.123.canvas", data)
entry, _ := kv.Get("room.123.canvas")
```

Use for: canvas state, room status, user settings.

## Stream (Event History)

```go
js.AddStream(&nats.StreamConfig{
    Name: "room-123", Subjects: []string{"room.123.>"},
    MaxAge: 24 * time.Hour, Storage: nats.FileStorage,
})
js.Publish("room.123.stroke", data)
```

Use for: audit log, replay, late joiner catch-up.

## Presence

```go
// NATS Core for ephemeral join/leave events
NC.Publish("presence.room.123.join", event)

// JetStream KV for current state
kv.Put("presence.room.123.users", usersJSON)
```

Use for: who's online, "user is typing...", join/leave notifications.

## goqite + JetStream Coexistence

goqite handles fire-and-forget + SSE streaming to the current user.
JetStream handles multi-user real-time across all users in a room.
They communicate via NATS client — no shared memory, no conflict.
