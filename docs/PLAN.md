---
PLAN: "fix(platformd): the «?» help button looks like a button (style.Button recipe)"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# PLAN — the help button looks like a button

## Why

The «?» next to a screen title (`platformd`, part `help-btn`) renders as a tiny loose character
with no button box (screenshot in mjosefa-cms, 2026-10-24): its style composes `style.As(style.Bare)`
+ `style.Interactive(style.Bare)` + `style.KeepSize()` by hand. `webtyp.com/widget/style` has the
recipe for anything the user presses, `style.Button(s Surface)` (`widget/style/except.go`), which
exists precisely because hand-composed buttons came out wrong. `crudview/css.go` already uses it
(`style.Button(style.Primary)`, `style.Button(style.Secondary)`).

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
```

## Stage 1 — `platformd/css.go`

The `Part(widget.Part("help-btn"), …)` block (around line 655) becomes exactly:

```go
		Part(widget.Part("help-btn"),
			style.Button(style.Secondary),
		).
```

Delete its five previous options (`Row`, `CenterContent`, `As(style.Bare)`, `Interactive(style.Bare)`,
`KeepSize`). Change nothing else. If a test in this repo asserts the old options of `help-btn`
(look in `platformd/consumer_stylesheet_test.go`), update that assertion to the new recipe and
list it in the PR.

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission or report progress: when the check passes, commit, push and open
  the PR. Stuck for 30 minutes: open the PR anyway listing what fails.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `platformd/css.go` (+ the stylesheet test if it asserts `help-btn`) | `help-btn` uses `style.Button(style.Secondary)` |
