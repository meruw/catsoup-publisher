# ADR 001: Local desktop Publisher

Date: 2026-11-04. Decision: accepted following the owner's request to select the
stack for maintainability, scalability and security.

## Decision

Use Go for application logic, Wails v3.1.0 as the desktop adapter, Vue 3 with strict
TypeScript and Vite for the UI, and SQLite for local persistent state. No hosted
service, public HTTP API, Electron, React, Nuxt or SSR is required.

Wails is a delivery mechanism, not the business architecture. The frontend talks
to a deliberately small service facade. Internal Go packages must not import
Wails. They can later be driven by a CLI, worker or another desktop adapter.

## Why this choice

- Wails fits the already accepted Go backend and uses Windows WebView2. It supplies
  the window, bridge and desktop integration without adding a second backend language.
- Vue has first-class TypeScript support, established SFC conventions, official
  routing/state tools and explicit security guidance. These suit this application's
  forms, queues, tables and approval states. Add Router/Pinia only when needed.
- Svelte is also viable. Its compiler and concise components are attractive, but
  do not remove our largest costs: media I/O, API limits and durable state. Choosing
  Vue is an application-maintenance judgment, not a claim that Svelte is insecure
  or unable to scale.
- Tauri is viable but adds Rust or a Go sidecar/IPC boundary. Electron adds a bundled
  browser and Node runtime that this Go application does not need.
- SQLite fits one local application. Multi-machine collaboration would require an
  explicit server, identity model and likely PostgreSQL migration; it is not an
  automatic configuration switch.

## Boundaries

```
Vue components -> typed bridge client -> desktop service facade
                                           |
                                    Go application logic
                                           |
                    delivery reader / draft store / future job runner
                                           |
                    read-only delivery files / SQLite / future connectors
```

Publisher reads delivery contract v2. Production remains Python/Blender. Trends is
independent; Python is the proposed research implementation. Neither Trends nor a
running Claude session is required by Publisher.

Initially, the existing Python validator remains the authoritative technical
check through a fixed, timeout-bounded adapter. Python and imageio-ffmpeg are thus
still needed for importing deliveries, even though Publisher's backend is Go.
Removing that dependency later requires shared conformance fixtures; do not silently
implement a second incompatible contract. No contract changes in this milestone.

## Security design

- Bundle only our own UI. Never load social websites or untrusted iframes in the
  window with Go bindings. Future authentication uses the system browser.
- Restrict CSP; no remote scripts, raw HTML rendering or untrusted Vue templates.
- Go validates each bridge argument. Access files through episode/video identifiers,
  canonical paths and root containment checks (including symlinks/junctions).
- Permit legacy paths only for the known PoC and its frozen output folders.
- Deny unused webview capabilities, including camera, microphone and geolocation.
- No public listener in the desktop release. An internal Wails asset handler is not
  an internet-facing HTTP service. Development tools are not release artifacts.
- Do not expose shell commands, arbitrary URLs or generic read/write functions.
- Read-only integrity checks never approve content. Publishing must recheck current
  hashes and approvals in Go, not just disable a button in Vue.
- Future connector credentials use the OS credential store; no tokens exist yet.
- A local malicious process running as the same Windows user is outside this app's
  isolation guarantee. A desktop bridge is not a full sandbox.

## Version and maintenance policy

Pin Go/tooling and all runtime packages, commit lockfiles, review security advisories
and apply supported security fixes. Wails v3 is recently GA (2026-10-26); v3.1.0 is
a stable release, not a nightly. Newness is a maintenance tradeoff. Keep experimental
server/mobile/MCP/auto-update features out and test upgrades before accepting them.

Reviewed Wails advisories include notification command injection (fixed in
v3.0.0-beta.31), v2 origin matching (fixed in v2.13.0), and Linux child-frame access
(v2 fix v2.17.1). This is source review, not an independent security audit. Before
distribution, run dependency audits and test the actual packaged executable.

## Growth path

Desktop shell update: a small Vue title bar uses Wails' existing Window runtime
adapter and two maximize-state event subscriptions. The window is frameless; the
shell adds no Go service methods, timers or runtime package dependencies. Sidebar
modules are navigation structure only; the planned Trends module is disabled.
Avoid adding routing/state frameworks until an actual second functional screen
requires them. Measure asset and executable growth as shell features are added.

1. Desktop foundation: narrow bridge, read-only validated deliveries, media preview.
2. SQLite drafts, migrations, optimistic version checks and destination history.
3. Manual handoff with account verification; approved deliveries only.
4. Official API connectors where free and available; durable jobs, bounded concurrency,
   cancellation, backoff and reconciliation before retrying ambiguous uploads.

Jobs require a running PC/process. A platform-accepted scheduled publication is
different from a task waiting in our local queue. Never promise exactly-once posting
where a provider cannot support it. Keep audit events and external post identifiers.

## Sources reviewed

- https://wails.io/concepts/architecture/
- https://wails.io/contributing/release-policy/
- https://github.com/wailsapp/wails/releases
- https://github.com/wailsapp/wails/security/advisories
- https://wails.io/features/windows/permissions/
- https://vuejs.org/guide/typescript/overview.html
- https://vuejs.org/guide/best-practices/security.html
- https://svelte.dev/docs/svelte/overview
- https://www.sqlite.org/whentouse.html
