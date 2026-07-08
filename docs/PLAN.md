# PLAN.md — multi-session roadmap for gogogo-fullstack-template

> Living document. Each numbered section is a **future session** — read
> "What we'll do" before opening the session to know the starting point,
> "Exit criteria" to know when the session is done.
>
> Update this file at the end of each session, not during. The session
> that closes work also commits the updated PLAN.md.

---

## 1. Wails v3 — desktop + mobile builds

### What we'll do

Add an optional second frontend target to the same project: a native
desktop app (Windows/macOS/Linux) and a mobile app (Android/iOS),
both built on top of the same Go business logic that already powers
the web app. No rewrite — Wails v3 wraps the existing handlers in
a webview and binds Go methods to JS; we just need a thin
`cmd/desktop/main.go` (and `cmd/mobile/main.go`) that:

1. Mounts the same router inside `application.New[options]`
2. Re-exports the same Datastar/Templ pages
3. Adds native-only stuff (menu bar, tray, file dialogs,
   notifications, biometric auth) via `application.OnStartup` hooks

### Why it matters

Right now the template is web-only. A user wanting to ship the same
admin tool (todos, PocketBase auth, LLM demos) as a desktop dashboard
or offline mobile app has to start a new project. Wails v3 lets us
ship both from one tree, one Go codebase, one set of tests.

CI strategy: **GitHub Actions runners build the platform artifacts,
the developer ships zero toolchains locally.** `macos-latest` builds
`.dmg` + `.app` (macOS), `.msi` (Windows cross from macOS), and
`.ipa` (iOS); `ubuntu-latest` builds `.deb` + `.AppImage` (Linux)
and `.apk` (Android). Artifacts are attached to the run, so a
forker who can push a tag gets a release with all binaries.

### Scope estimate

| Sub-step | Time | Output |
|----------|------|--------|
| Scaffold `cmd/desktop/main.go` + Wails v3 imports | 30 min | Compiles, opens blank window |
| Mount existing router via `application.Options.Assets` | 1 h | Webview shows the existing web UI |
| Add 2 build tags: `//go:build desktop` / `//go:build mobile` | 30 min | Clean separation, `cmd/web/` untouched |
| Write `Makefile` targets: `make desktop`, `make mobile-bundle` | 1 h | Repeatable local + CI build |
| GitHub Actions workflow `build-platforms.yml` (matrix: mac/win/linux × desktop/android/ios) | 3 h | 6 jobs, artifacts attached |
| Native menu bar + tray + biometric auth demo | 2 h | One proof-of-concept per platform |
| Tests: Wails test helpers + a smoke test per platform | 2 h | `make test-desktop` green |

**Total**: ~12 h focused work, fits in 2 sessions.

### Where to look when we start

- Wails v3 docs (https://wails.io) — `v3_preview` branch in their repo
- Existing `cmd/web/main.go` for the HTTP entry point to mirror
- The LLM `fakeserver` we added — desktop app can ship with the
  fake mode toggled by envvar so users get a working demo offline

### Exit criteria

- [ ] `make desktop` produces a runnable `.app` (macOS dev) on macOS
- [ ] `make android` produces a runnable `.apk` from any dev machine
- [ ] `.github/workflows/build-platforms.yml` runs on every push and
      uploads `.dmg`, `.msi`, `.deb`, `.AppImage`, `.apk`, `.ipa`
      artifacts — all green
- [ ] `cmd/desktop/main.go` and `cmd/mobile/main.go` are < 200 lines
      each and reuse the existing router/handlers as-is
- [ ] README has a "Desktop & Mobile" section with screenshot of
      the native wrapper around the web UI

---

## 2. Persistent Wails dev reload (deferred)

If Wails v3 lands cleanly, the next session will:

- Add `wailsjs/` codegen to `make codegen`
- Add a `--live-reload-from=web` flag that proxies to `localhost:8080`
  during development so the desktop app reloads when the web dev server
  does

---

## 3. Local-first repo CI as reusable workflow

The current `make check` / `make gate` works locally. The next
session extracts this into a reusable GitHub Actions workflow
(`.github/workflows/lint-test.yml`) that downstream forks can
`uses: calionauta/gogogo-fullstack-template/.github/workflows/lint-test.yml@master`
to inherit our quality bar with zero config.

---

## 4. Multi-project promotion

If we adopt `gogogo-fullstack-template` for `stelow` and
`datastar-lint` (sibling repos already in `~/Development/`), the
canonical `/opt/<name>/{bin,compose,env,secrets,data}` layout
becomes a shared convention. The next session:

- Extracts `scripts/deploy-prod.sh` to a separate `gogogo-deploy`
  repo (or template file) so the three projects share the same
  deploy runner
- Adds a `Makefile` target `make release` that versions + tags +
  builds + deploys in one command

---

## 5. Cleanup post-rename

Items left over from the `gogogo-template` → `gogogo-fullstack-template` rename:

- [ ] Server still has `/home/deploy/gogogo-template/` (old deploy dir)
      — next deploy will create `/home/deploy/gogogo-fullstack-template/`
      alongside it. Once the new one is verified, `rm -rf` the old.
- [ ] Cloudflare Tunnel ingress at `gogogo.calionauta.com` still points
      at the old container — needs a new tunnel entry for the new
      `<name>.example.com` host
- [ ] Both `v0.1.0` … `v0.4.0` tags exist on the renamed repo
- [ ] The `gh release` notes for v0.1.0 still mention the old name
      (cannot edit after fact; document in a CHANGELOG note)

---

## Notes for whoever picks this up

- The session that opened this file used `git filter-repo --mailmap`
  for the privacy pass — that pattern is reusable, see
  the backup repo at `~/backups/gogogo-template-*.git/` for the
  migration script.
- The `caveman ultra` user preference (terse chat, full prose in
  deliverable docs) is preserved in agentmemory — sessions don't
  need to re-ask.
- The deploy pipeline is **Pattern B** (shell key, image built on
  server) — see `~/.agents/skills/cali-ops-deploy-github-tailscale/`
  for the two patterns and when to use each.
