# cali-go-stack

## Project Overview

Go web application template with Datastar + Templ + PocketBase + goqite + NATS JetStream.
Module: `github.com/calionauta/cali-go-stack`

## Stack

Go 1.26 | Templ v0.3.1020 | Datastar v1.0.2 | Datastar-Go v1.2.2 | goai v0.7.6 | PocketBase v0.39.5 (embedded, ncruces/go-sqlite3) | DaisyUI v5.6.13 + TailwindCSS | goqite v0.4.0 | retry-go v4 | NATS JetStream (opt-in, build tag) | age v1.3.1

## DaisyUI Components

Before building any UI, read the full DaisyUI component catalog:
- https://daisyui.com/llms.txt

Use DaisyUI components for ALL HTML UI. Prefer: `table`, `input`, `btn`, `badge`, `card`, `modal`, `toast`, `loading`, `collapse`, `join`, `kbd`, `select`, `toggle`, `range`, `progress`, `chat`, `diff`, `stat`, `timeline`.

## Skills

Install via `npx skills add`:

```bash
npx skills add https://github.com/calionauta/agent-sync-public --skill cali-coding-go-standards --yes
npx skills add https://github.com/calionauta/agent-sync-public --skill cali-code-navigation --yes
```

| Skill | Source | When |
|-------|--------|------|
| [`cali-coding-go-standards`](../../../.qwen/skills/cali-coding-go-standards/SKILL.md) | `agent-sync-public` | Code quality: KISS, DRY, file size, functions, `slog`, error handling, tests, lints |
| [`cali-code-navigation`](../../../.qwen/skills/cali-code-navigation/SKILL.md) | `agent-sync-public` | Code nav: cymbal over grep/find |

For stack-specific patterns (Datastar, Templ, DaisyUI, goqite, JetStream, queues, workflows), see the project's `references/` directory and `docs/` directory.

## Commands

| Command | Description |
|---------|-------------|
| `make dev` | Live reload with Air |
| `make build` | Build binary |
| `make build-jetstream` | Build with JetStream support |
| `go test -race ./...` | Run tests with race detector |
| `go tool templ generate` | Generate Templ components |
| `golangci-lint run ./...` | Run linter |
| `deadcode -test ./...` | Find dead code |

## Don'ts

- Do NOT use HTMX/Alpine.js — use Datastar
- Do NOT use fmt.Sprintf for HTML — use Templ
- Do NOT remove goqite when adding JetStream (they are complementary)
- Do NOT use modernc.org/sqlite — driver is ncruces/go-sqlite3
- Do NOT use raw CSS class names when a DaisyUI component exists — check `daisyui.com/llms.txt`
- Do NOT use `log` package — use `slog` (`log/slog`, stdlib since Go 1.21)

## Architecture

```
cmd/web/main.go          → Entry point (PB embedded + goqite + SSE Hub)
  -tags jetstream        → Also starts embedded NATS server + JetStream
config/                  → Env-based config (dev/prod build tags)
db/                      → PocketBase setup + repositories
internal/
  secrets/               → age-decrypt loader
  queue/                 → goqite + SSE Hub + workers + retry
  nats/                  → NATS JetStream (build-tag gated)
  llm/                   → GoAI client
  datastar/              → Datastar render helpers
features/
  app/                   → Application core
  todo/                  → Example: Todo MVC (Datastar + DaisyUI + PocketBase + SSE Hub)
web/resources/           → Static assets (embedded)
router/                  → Route registration on PocketBase

Three async layers (complementary, not alternatives):
  goqite    → background jobs (fire-and-forget, SSE streaming)
  turbine   → durable workflows (multi-step, crash recovery)
  JetStream → multi-user real-time (KV, streams, presence)
```
