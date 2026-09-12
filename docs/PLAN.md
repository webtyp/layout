---
PLAN: "feat(platformd): IdleTimeout + OnIdle — close the session when the user walks away"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 6417421674963655552
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Phase L** of
> [`LAN_RUT_AUTH_MASTER_PLAN.md`](https://github.com/tinywasm/app/blob/main/docs/LAN_RUT_AUTH_MASTER_PLAN.md).
> Parallel with phases A/B/S; the leaf app (phase D) waits for this tag only
> for its idle-lock acceptance criterion. Doctrine: `CONSTRUCTION_HARNESS.md`
> in `tinywasm/app` docs.
>
> ⚠ The working tree of this repo currently carries an **unrelated
> uncommitted change** (`rightpanel/css.go`). Do not touch, revert, or
> commit it — this plan is `platformd`-only.

# Plan — `webtyp.com/layout/platformd`: the idle lock

## 0. Context

Clinical apps on shared LAN devices must not stay open on an unattended
screen. The server-side enforcement is a sliding session TTL (phase A in
`webtyp.com/auth`); this is the client half: after N seconds without user
activity **inside the platform**, the shell fires `OnIdle` and the app logs
out + reloads. Every app would write the same listener/timer wiring → it
belongs in the shell, not in each `client.go`.

`webtyp/dom` is deliberately **not** modified: its pending typed-events plan
owns that surface. platformd attaches element-level `On(...)` listeners to
its own root (mouse/touch bubble from anywhere inside the app; key events
arrive while focus is inside the app, which is the case during any real
interaction) and reuses the existing document-level `OnScrollCapture`
(already wired in `Init`) as one more activity signal. The claim is exactly
what that delivers: *activity inside the platform resets the idle timer*.

## Design gate (api-design — five answers)

1. **Prior art.** **ng-idle / @ng-idle/core** (Angular): `Idle` + `Timeout`
   services over document-level mouse/keyboard events. **react-idle-timer**:
   one component, `timeout` + `onIdle` props, same event set. **OWASP Session
   Management Cheat Sheet**: recommends client-side idle logout *in addition
   to* server-side inactivity timeout — which is exactly the split phase A /
   phase L implement. We differ in shape: two fields on the existing
   `Platform` struct literal (no services, no providers, no new concepts),
   because platformd configures itself by struct fields already
   (`CanView`, `OnLogout`, `UserActions`).
2. **Novice-name test.** `IdleTimeout` — "how long idle before it fires";
   `OnIdle` — "what runs when idle" (mirrors the existing `OnLogout` field
   naming). Seconds as the unit, documented on the field (same unit as
   `auth.Config.TokenTTL`/`IdleTTL` — one vocabulary across the wave).
3. **Complexity ledger.** Concepts +2 fields / −0 (no timers-as-objects, no
   event lists — the event set is fixed and documented). Files an app touches
   to get idle logout +0 / −1 (two struct fields instead of a hand-written
   listener+timer block in every `client.go`). Call-site lines +3 / −25.
   Ways to do the same thing +0 / −0.
4. **Where it belongs.** The shell owns "the user is present" the same way it
   owns navigation and notifications; apps own the *consequence* (`OnIdle` →
   logout), which is policy. No second concern enters platformd: it detects
   idleness, it does not log out.
5. **What it deletes.** The would-be per-app idle detector (mjosefa-cms phase
   D no longer writes one; no existing code in this repo is removed).

## Stage 1 — the fields

**File:** `platformd/platformd.go`, `Platform` struct (public section, after
`DefaultID`):

```go
// IdleTimeout is the number of seconds without user activity inside the
// platform (mouse, keyboard, touch, or document scroll) before OnIdle fires
// once. 0 — the default — disables the idle lock entirely: nothing is
// armed, no listener changes behavior.
IdleTimeout int

// OnIdle is called when IdleTimeout seconds pass without activity. Required
// when IdleTimeout > 0 — Init panics otherwise (a configured idle lock with
// no consequence would be a silent failure). The platform does NOT re-arm
// after firing: the next activity starts a fresh period. Typical use: post
// the logout route and reload.
OnIdle func()
```

## Stage 2 — the wiring

**File:** `platformd/platformd.go`.

1. Internal state (private section of the struct): `idleTimer
   tintime.Timer` and `idleArmed bool` (`tintime` is this repo's existing
   import alias for `webtyp.com/time`; if platformd.go does not import it
   yet, add it — layout's go.mod already requires it).
2. In `Init(ctx Ctx)`, before the hash routing block:
   ```go
   if p.IdleTimeout > 0 && p.OnIdle == nil {
       panic("platformd: IdleTimeout requires OnIdle")
   }
   ```
   and inside the existing `OnScrollCapture` callback, call
   `p.activity()` (scroll counts as presence).
3. New private methods:
   ```go
   // activity rearms the idle timer; called from every presence signal.
   func (p *Platform) activity() {
       if p.IdleTimeout <= 0 { return }
       if p.idleTimer != nil { p.idleTimer.Stop() }
       p.idleTimer = tintime.AfterFunc(p.IdleTimeout*1000, func() {
           p.idleTimer = nil
           p.OnIdle()
       })
   }

   // armIdle attaches the presence listeners to the platform root. Called
   // once from Render (idleArmed guards re-renders).
   func (p *Platform) armIdle(root *Element) {
       if p.IdleTimeout <= 0 || p.idleArmed { return }
       p.idleArmed = true
       for _, ev := range []string{"mousemove", "keydown", "touchstart"} {
           root.On(ev, func(Event) { p.activity() })
       }
       p.activity() // opening the app counts as presence
   }
   ```
   (Event-set strings live in this one slice — no scattered literals. If the
   root element variable in `Render()` has another name, adapt the call site,
   not the method.)
4. In `Render()`: call `p.armIdle(root)` on the root element of the shell
   (the same element the existing render builds first), immediately before
   returning.

## Stage 3 — tests

**File:** `platformd/platformd_test.go` (extend; follow its existing
conventions and build lane).

1. `IdleTimeout: 0` → `Init` does not panic with `OnIdle` nil; no timer armed.
2. `IdleTimeout: 30, OnIdle: nil` → `Init` panics with exactly
   `platformd: IdleTimeout requires OnIdle`.
3. `IdleTimeout: 1` + recording `OnIdle`, no activity → fires within ~1.5 s
   (wasm lane: real `time.AfterFunc` provider); a second period does not
   fire without new activity.

## Stage 4 — docs

`platformd/README.md` (or the layout README section for platformd): two
rows in the field table + the 4-line usage example
(`IdleTimeout: 30 * 60, OnIdle: logoutAndReload`). VERIFY against the
implementation.

## Acceptance criteria

1. `go build ./...`, `go vet ./...`, `gotest ./...` green (both lanes).
2. `grep -rn "IdleTimeout" platformd/` → struct field, Init guard, armIdle/activity, tests, README — no other surface.
3. `git status` shows `rightpanel/css.go` still modified and untouched by this plan's commits.

| Stage | File | Action |
|---|---|---|
| 1 | `platformd/platformd.go` | `IdleTimeout`, `OnIdle` fields |
| 2 | `platformd/platformd.go` | panic guard, `activity`/`armIdle`, scroll signal |
| 3 | `platformd/platformd_test.go` | three cases |
| 4 | README | verify docs |
