// SCOPE:core - Browser smoke test (headless). Builds the binary, boots it
// on an ephemeral port + data dir, creates a test user, then loads the
// real app pages in headless Chromium and FAILS if any uncaught client-side
// JS exception (ReferenceError / TypeError / pageerror) fires. This is the
// gate that catches the "offlineSync is not defined" class of error BEFORE
// a deploy reaches production — `make ci-local` / CI run it after the unit
// tests.
//
// No project cache is involved: a fresh browser context is used every run,
// so it loads the genuine served HTML (not a stale browser/Cloudflare copy).
//
// Requires: `npx playwright install chromium` (run by `make smoke`).

import { chromium } from "playwright";
import { spawn, spawnSync, execSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const PORT = Number(process.env.SMOKE_PORT || 8099);
const BASE = `http://127.0.0.1:${PORT}`;
const SU_EMAIL = "smoke-superuser@local.dev";
const SU_PASS = "SmokeSuperuserPass!123";
const USER_EMAIL = "smoke-user@local.dev";
const USER_PASS = "SmokeUserPass!123";

const ROUTES = ["/todos", "/whiteboard", "/login"];

const fail = (msg) => {
  console.error("❌ " + msg);
  process.exitCode = 1;
};

const tmp = mkdtempSync(join(tmpdir(), "gogogo-smoke-"));
const pbDir = join(tmp, "pb");
const bin = join(tmp, "web");

let server = null;
let browser = null;

async function api(method, path, { body, token } = {}) {
  const res = await fetch(BASE + path, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: "Bearer " + token } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  let json = null;
  try {
    json = text ? JSON.parse(text) : null;
  } catch {
    /* non-JSON */
  }
  return { status: res.status, json };
}

async function waitForHealth(timeoutMs = 30000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const r = await fetch(BASE + "/health");
      if ((await r.text()).trim() === "ok") return;
    } catch {
      /* server not up yet */
    }
    await new Promise((r) => setTimeout(r, 500));
  }
  throw new Error("server did not become healthy in " + timeoutMs + "ms");
}

try {
  console.log("→ Building binary (./cmd/web)…");
  execSync(`go build -o ${JSON.stringify(bin)} ./cmd/web`, {
    stdio: "inherit",
    timeout: 180000,
  });

  console.log(`→ Starting server on ${BASE} (data dir ${pbDir})…`);
  server = spawn(bin, ["serve", "--http", `127.0.0.1:${PORT}`, "--dir", pbDir], {
    stdio: "ignore",
  });
  server.on("exit", (code) => {
    if (code && code !== 0) console.error(`server exited with code ${code}`);
  });

  await waitForHealth();

  console.log("→ Creating superuser + test user…");
  spawnSync(bin, ["superuser", "upsert", "--dir", pbDir, SU_EMAIL, SU_PASS], {
    stdio: "ignore",
  });
  const su = await api("POST", "/api/collections/_superusers/auth-with-password", {
    body: { identity: SU_EMAIL, password: SU_PASS },
  });
  if (!su.json?.token) throw new Error("superuser auth failed: " + JSON.stringify(su.json));
  const suToken = su.json.token;

  // Create a regular app user (retry briefly in case seeding is still in flight).
  let created = false;
  for (let i = 0; i < 10 && !created; i++) {
    const r = await api("POST", "/api/collections/users/records", {
      token: suToken,
      body: { email: USER_EMAIL, password: USER_PASS, passwordConfirm: USER_PASS },
    });
    if (r.status === 200 || r.status === 400) created = true; // 400 = already exists
    else await new Promise((res) => setTimeout(res, 500));
  }
  const auth = await api("POST", "/api/collections/users/auth-with-password", {
    body: { identity: USER_EMAIL, password: USER_PASS },
  });
  if (!auth.json?.token) throw new Error("user auth failed: " + JSON.stringify(auth.json));
  const userToken = auth.json.token;

  console.log(`→ Launching headless Chromium; testing routes: ${ROUTES.join(", ")}`);
  browser = await chromium.launch({ headless: true, args: ["--no-sandbox"] });
  const context = await browser.newContext();
  await context.addCookies([
    { name: "gogogo_auth", value: userToken, url: BASE + "/" },
  ]);

  const pageErrors = [];
  const consoleErrors = [];
  const page = await context.newPage();
  page.on("pageerror", (err) =>
    pageErrors.push({ msg: String(err), stack: err.stack || "", url: page.url() }),
  );
  page.on("console", (msg) => {
    if (msg.type() === "error") consoleErrors.push(msg.text());
  });

  for (const route of ROUTES) {
    pageErrors.length = 0;
    consoleErrors.length = 0;
    try {
      await page.goto(BASE + route, { waitUntil: "load", timeout: 20000 });
    } catch (e) {
      console.error(`  ! navigation to ${route} failed: ${e.message}`);
    }
    if (route === "/todos") {
      await page.content().then((c) => writeFileSync("/tmp/served_todos.html", c));
    }
    // Give inline scripts + SSE a moment to execute.
    await page.waitForTimeout(1000);
    if (pageErrors.length > 0) {
      fail(`uncaught JS error on ${route}: ${pageErrors.map((e) => e.msg).join(" | ")}`);
      for (const e of pageErrors) {
        console.error(`      at ${e.url}`);
        if (e.stack) console.error(e.stack.split("\n").slice(0, 4).join("\n"));
      }
    }
    const tag = pageErrors.length ? "❌" : "✓";
    console.log(`  ${tag} ${route} (console errors: ${consoleErrors.length})`);
    for (const ce of consoleErrors) console.log(`      console.error: ${ce}`);
  }

  await context.close();
} catch (e) {
  fail(e.stack || String(e));
} finally {
  if (browser) await browser.close().catch(() => {});
  if (server) server.kill("SIGKILL");
  rmSync(tmp, { recursive: true, force: true });
}

if (process.exitCode === 1) {
  console.error("\n❌ Browser smoke test FAILED");
} else {
  console.log("\n✅ Browser smoke test passed (no uncaught client JS errors)");
}
