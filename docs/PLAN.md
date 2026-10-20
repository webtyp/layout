---
PLAN: "feat(platformd): Described and AlertedModule capability interfaces (types only, no wiring)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 17491794384523155737
---

> Read AGENTS.md at the repo root before any change. This plan is complete: do not guess; if something does
> not fit, stop and ask in the PR. This is a deliberately small slice of a larger feature (two previous
> attempts at the full feature failed): **only new types in `platformd/capabilities.go`, nothing else.** Do
> not touch `platformd.go`, `css.go`, or any rendering code. Design history:
> `webtyp/docs/MODULE_GUIDANCE_MASTER_PLAN.md` (D1 closes P1, D6 closes P6, D19 closes P16).

## Goal

Two new optional `UIModule` capabilities, following `Badged`'s exact shape (`platformd/capabilities.go`):
pure interface declarations and their small value types, so that `alert_manager` and any describable module
can start implementing them. The chassis wiring (a "?" button, a welcome card, an alert strip) is separate,
later plans.

## Git rules (strict)

- Work on the branch codejob created. Never `git reset`/`checkout -- <file>`/`clean`/`stash`. Commit often.
- **Never pause to ask permission or to report progress.** Delivery check passes → commit, push, open the PR.
  Still failing after 30 minutes → open the PR anyway, list exactly what fails.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Stage 1 — `platformd/capabilities.go`

Add, right after the existing `Badged` interface (do not touch `Badged`, `Notifier`, `UsesNotifier`,
`GroupModule`, or anything else already in this file):

```go
// Described is an optional capability of a UIModule: what it is, for someone who has never used it, plus
// optional first-steps guidance. A future chassis feature draws it; this interface alone does not change
// anything rendered today.
type Described interface {
	Description() lang.Text // what this module is for, in plain words — never the internal structure
	Steps() []Step           // optional; nil/empty means no step list
}

// Step is one setup prerequisite shown with a checkmark once Done is true.
type Step struct {
	Label lang.Text
	Done  bool
}

// AlertedModule is an optional capability of a UIModule: it has alerts a person has not seen yet. A future
// chassis feature draws them; this interface alone does not change anything rendered today. ModuleAlert and
// this interface are deliberately minimal — alert_manager (a different repo) owns storage, "seen" state and
// delivery; this is only the shape a future chassis feature will need.
type AlertedModule interface {
	UnseenAlerts() []ModuleAlert
}

// ModuleAlert is one row a future alert strip will draw.
type ModuleAlert struct {
	ID      string
	Subject string
	OnView  func()
	OnSeen  func()
}
```

Add `"webtyp.com/lang"` to this file's imports if not already present (check first — `platformd.go` likely
already imports it, but `capabilities.go` may not).

## Stage 2 — tests

Through the exported API, `tests/` (check this repo's existing test package name — do not assume
`platformd_test`, grep an existing file in `tests/` first): a tiny fake `UIModule` implementing `Described`
and one implementing `AlertedModule`, asserting only that the interfaces compile and their methods return
what was set — these are type-contract tests, nothing rendered. No test touches `Platform`/`platformd.go` at
all in this plan.

## Stage 3 — docs

`docs/ARCHITECTURE.md`/`README.md`: document `Described`/`Step`/`AlertedModule`/`ModuleAlert` next to
`Badged`'s own documentation, noting explicitly that chassis wiring is a separate, later plan.

## Delivery check

- `gotest` green (vet, race, tests, wasm).
- `git diff --stat main --diff-filter=D` empty.
- `git diff --stat main` touches only `platformd/capabilities.go`, its test file, and docs — nothing in
  `platformd.go` or `css.go`.
- No behaviour change: nothing in the running app renders differently.

## Stages

| # | File(s) | What |
|---|---|---|
| 1 | `platformd/capabilities.go` | `Described`, `Step`, `AlertedModule`, `ModuleAlert` |
| 2 | `tests/*` | type-contract tests only |
| 3 | `docs/ARCHITECTURE.md`/`README.md` | document the two new capabilities |
