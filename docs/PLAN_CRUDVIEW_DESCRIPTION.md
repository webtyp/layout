> **Parked** — `docs/PLAN.md` already has a plan running (H2a, capability types in `platformd`). Merge this
> into `docs/PLAN.md` as its next stage once H2a closes (**agents-workflow** → «A repo that already has a
> plan»), or dispatch it standalone once H2a's slot frees up — do not dispatch two plans in this repo at once.

---
PLAN: "feat(crudview): show the view's Description when nothing is selected, EmptyMessage when the list is empty"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change (same repo as `platformd`: no conditional mounting,
> visibility is a `*SignalBool`/`*SignalString` driving `BindStateFunc`/`Show`, never a Go `if` around a
> `Div`). This plan is complete: do not guess; if something does not fit, stop and ask in the PR. Depends on
> `webtyp.com/view` v0.6.40 (already published): `view.Presenter.Description() lang.Text` and
> `EmptyMessage() lang.Text`. Design history: `webtyp/docs/MODULE_GUIDANCE_MASTER_PLAN.md` (D4 closes P4).

## Goal

Two independent, additive displays in `crudview.CrudView.Render()` (`crudview/crudview.go`):

1. **Description**, shown in the article/form area exactly when nothing is selected and nothing is being
   composed (`v.selected.Get() == "" && !v.composing.Get()`) — the same condition `hasSelection()`
   (`crudview/crudview.go:~226-229`) already computes; reuse it, do not duplicate the check.
2. **EmptyMessage**, shown in the list area exactly when the list has no rows — the existing `v.hasRows`
   signal (`crudview/crudview.go:~862-863`, already set by `filter()`).

Both texts come from `v.Presenter.Description()`/`EmptyMessage()` (new in `view.Presenter` as of v0.6.40);
`""` means "nothing to show" — render nothing, not an empty box.

## Git rules (strict)

- Work on the branch codejob created. Never `git reset`/`checkout -- <file>`/`clean`/`stash`. Commit often.
- **Never pause to ask permission or to report progress.** Delivery check passes → commit, push, open the PR.
  Still failing after 30 minutes → open the PR anyway, list exactly what fails.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Stage 0 — dependency

```bash
go get webtyp.com/view@v0.6.40
go mod tidy
```

## Stage 1 — Description in the article area

In `crudview/crudview.go`'s `Render()` (`~993-1002`), where `boxContent` is built:

```go
boxContent := Div().Set(clsBoxContent.AsAttr())
if v.Form != nil {
	boxContent.Child(v.Form)
}
```

Add a description element as a sibling, always present, whose CSS visibility is driven by a bound state —
follow whichever of this repo's own existing patterns already toggles a sibling element by a condition on
`v.selected`/`v.composing` (search `hasSelection` for its call sites first; if a `BindStateFunc`-style
toggle on a `widget.State` is already used nearby for a similar "something vs. nothing selected" display,
copy that exact construction — do not invent a new one). The element's text is bound to
`v.Presenter.Description()` (a static value at `Render` time is fine: `Description()` does not change after
construction, unlike `Items()`). If `v.Presenter.Description()` is `""`, do not add the element at all (no
empty box, no wasted DOM node) — this one case **is** a plain Go `if`, because it depends on data fixed at
construction time, not on a signal that changes later.

## Stage 2 — EmptyMessage in the list area

Where `v.panel.Aside` is built (`crudview/crudview.go:~1009-1010`, `Child(v.list)`), add a sibling element
whose visibility is bound to `v.hasRows` (`BindStateFunc(widget.Open, func() bool { return !v.hasRows.Get() })`
or whatever state/attribute this repo's existing empty-state convention uses — grep for any other place
that already shows/hides on `hasRows`/`itemsLoaded` before inventing one). Its text is
`v.Presenter.EmptyMessage()`. Same `""` rule as Stage 1: if empty, do not add the element.

## Stage 3 — tests

Through the exported API, `crudview/tests/` (grep an existing file for this package's exact `package` name
before writing — do not assume `crudview_test`), a fake `Presenter` with `Description()`/`EmptyMessage()`
set:

- Nothing selected, no draft composing → the description element is present with the right text.
- A record selected → the description element is absent (or hidden — match whichever your Stage 1
  implementation chose; assert on the same observable the rest of this repo's tests use for a hidden vs.
  absent element).
- An empty list → the empty-message element is present with the right text; adding one row removes it.
- `Description()`/`EmptyMessage()` both `""` → neither element exists at all (regression guard: no empty
  boxes).

## Stage 4 — docs

`crudview/docs/ARCHITECTURE.md`/`README.md` (or this repo's equivalent location for `crudview`'s own docs —
check the actual path first): document both displays next to `CrudView`'s existing behaviour description.

## Delivery check

- `gotest` green (vet, race, tests, wasm).
- `git diff --stat main --diff-filter=D` empty.
- A `Presenter` that does not set either option (both return `""`) renders byte-for-byte as before — no
  behaviour change for every existing consumer of `crudview`.

## Stages

| # | File(s) | What |
|---|---|---|
| 0 | `go.mod` | bump `webtyp.com/view` to v0.6.40 |
| 1 | `crudview/crudview.go` | description element, bound to `hasSelection`'s existing condition |
| 2 | `crudview/crudview.go` | empty-message element, bound to `v.hasRows` |
| 3 | `crudview/tests/*` | presence/absence across all four states |
| 4 | `crudview/docs/ARCHITECTURE.md`/`README.md` | document both displays |
