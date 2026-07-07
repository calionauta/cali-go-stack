# NATS Core vs JetStream vs SSE Hub

## Three Tools, Three Jobs

| Tool | Scope | Persistence | Use Case |
|------|-------|-------------|----------|
| **goqite SSE Hub** | 1 user session | ❌ Events are ephemeral | Notify current user of background job completion |
| **NATS Core** | Multi-user, 1 server | ❌ | Ephemeral broadcasts (cursor positions, notifications) |
| **NATS JetStream** | Multi-user, multi-server | ✅ FileStorage | Shared state, event history, presence, late joiners |

## Decision Flow

```
Do you need to notify only the current user?
  YES → SSE Hub (built into goqite, no extra infra)
  NO → ↓

Do you need to share state between multiple users?
  NO → NATS Core (ephemeral broadcast)
  YES → ↓

Do users need to see past state (late joiners)?
  NO → NATS Core
  YES → NATS JetStream (KV for state + Stream for history)
```

## Quick Reference

| Need | Solution | Why |
|------|----------|-----|
| Single-user job notification | goqite SSE Hub | Built-in, zero config |
| Multi-user cursor sync | NATS Core | Ephemeral, low latency |
| Shared canvas (whiteboard) | JetStream KV | Persistent, atomic writes |
| User presence (who's online) | NATS Core + KV | Core for events, KV for state |
| Event history / audit | JetStream Stream | Durable, replayable |
| Late joiner catch-up | JetStream KV + Stream | KV for current, Stream for history |
| Multi-instance scaling | JetStream Consumer Groups | Load balancing across instances |
