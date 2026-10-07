---
PLAN: "feat(crudview): render view actions — buttons, confirmation, OnAction"
EXECUTOR: jules
REVIEWER: none
---

> Executed LOCALLY on 2026-10-07: the dispatched Jules session failed and left no branch.
> Deviations from the text below, all forced by existing tests of this package:
> - CSS parts are `action-bar` / `action-bar-btn` / `control-stack`, not `actions` / `action-btn`:
>   `.crudview__actions` is a retired class a stylesheet test forbids, and any `.crudview__aside…`
>   name collides with another forbidden one.
> - The bar lives in the aside **controls** band (below the filter, in a `control-stack` when a
>   filter exists), not as a sibling of the list: that band keeps its size and the list keeps its fill.
> - Enabled state reads the presenter's items through a new `itemsLoaded` signal, never `hasRows`.
>
> Phase F6b of the network administration master plan (private repo `veltylabs/mjosefa-cms`; you do
> not need it). **Depends on `webtyp.com/view` with actions** (phase F6a). First line of work:
> `go get webtyp.com/view@latest`; if `view.Actioner` does not exist in what you get, STOP and
> report — never add a `replace`, never declare the types locally.

# Plan — `webtyp.com/layout/crudview`: actions

**The spec is [docs/ARCHITECTURE.md → "Actions — commands on the whole list"](ARCHITECTURE.md)** plus
`webtyp.com/view`'s SPECS §9 (in the module cache: `$(go env GOMODCACHE)/webtyp.com/view@<version>/docs/SPECS.md`).
Follow [AGENTS.md](../AGENTS.md) (styling with the `widget/style` DSL, value-embedded `dom.Element`,
`lang` dictionary keys, CSS in `css.go` with `//go:build !wasm`).

## Development rules

- Pure renderer: crudview never talks to a `router.Caller`; it calls `Presenter.(view.Actioner)`.
- Reuse what exists — do not invent parallel mechanisms:
  - the confirmation modal: copy the **pattern** of `confirmDelete` / `renderDeleteConfirm` /
    `confirmDeleteAction` in `crudview.go` (same `modaldialog.ModalDialog`, `HideClose: true`, same
    button classes for Cancel and the confirm button; the confirm button is **not** the danger
    variant — an action is not a delete);
  - the outcome hook: mirror `OnSaved` / `reportSave` (`Log` on error, then the hook, exactly once);
  - styling: add parts/classes in `css.go` with the existing DSL and recipes; no raw CSS strings,
    no hex colours.
- New dictionary keys go where crudview's existing keys are declared (`lang.json` / DICTIONARY.md
  per AGENTS.md); the button text itself is the module's `action.Label`.
- Tests in `crudview` follow the package's existing convention (its `*_test.go` and
  `*_wasm_test.go` files live next to the code — keep that convention for this package). Runner
  `gotest ./...` including WASM.

## Design gate

**1. Prior art.** Django admin action bar + intermediate confirmation page; React-Admin list toolbar
buttons; GitHub's "Merge" button with a confirm step. Adopted: a visible text button per action,
optional confirmation naming the action, disabled while nothing to act on / while running.

**2. Novice-name test.** `CrudView.OnAction(op, err)` next to `OnSaved(err)`; CSS part
`actions` / `action-btn` read as what they are.

**3. Complexity ledger.**
```
Concepts the developer must learn   +1 hook (OnAction); the buttons appear by themselves
Files they must touch to do X       0 in the app (the module declares the action)
Lines at the call site              0 (optional: 1 OnAction assignment)
Ways to do the same thing           0
```

**4. Where it belongs.** Drawing presenter capabilities is crudview's job; the action contract is
`view`'s.

**5. What it deletes.** Nothing — new capability.

## Stage 1 — state and hook (`crudview.go`)

- `CrudView` fields: `OnAction func(op string, err error)` (exported, doc like `OnSaved`);
  unexported: the running op (a signal, so buttons can bind `disabled`), the op awaiting
  confirmation, `confirmAction *modaldialog.ModalDialog`.
- `Init`: build `confirmAction` like `confirmDelete` (message bound to the pending action's
  `Confirm`, buttons Cancel / `<Label>`).
- `actionRequest(a view.Action)`: if `a.Confirm != ""` → set pending, open modal; else
  `runAction(a.Op)`.
- `runAction(op string)`: set running; `actioner.Run(op, func(err error){ clear running; if err == nil { repaint the list from the reloaded presenter (same path Reload's done uses) }; reportAction(op, err) })`.
- `reportAction(op, err)`: `Log(err.Error())` on error, then `OnAction(op, err)` if set — exactly
  once per run.

## Stage 2 — render (`crudview.go` `Render`, `css.go`)

- When the presenter implements `view.Actioner` **and** `len(Actions()) > 0`: an actions band
  (`Div` with the new class) as the first child of the aside, before the list box; one `Button` per
  action, `Attr("name", "cv-action-"+op)`, text `lang.Translate(a.Label).String()`,
  `BindAttrBool("disabled", …)` true when `len(Presenter.Items()) == 0` or that op is running;
  `OnClick` → `actionRequest(a)`.
- No band when there are no actions (no empty element in the DOM).
- Mount the `confirmAction` modal next to the delete confirmation mount.
- `css.go`: classes for the band and the button. The band is `style.Row(style.Space1)` — `Row`
  already wraps (`flex-wrap: wrap`); there is no `style.Wrap`. The button reuses the text-button
  recipe the confirmation buttons use. Import path: `webtyp.com/widget/style` (never
  `github.com/tinywasm/...`).

## Stage 3 — tests

- `conformance_test.go`: fill the new `Driver` fields:
  `ActionLabels` (texts of the `cv-action-*` buttons in order), `ActionEnabled(op)` (evaluates the
  same condition the binding uses), `ClickAction(op)` (calls `actionRequest` for that action),
  `ConfirmAction` (calls the confirm handler when a pending action exists). All five new
  `view/conformance` clauses must pass.
- `actions_wasm_test.go` (`//go:build wasm`): a presenter over `conformance.FakeLister` with one
  action with `Confirm`: the button renders with its label; it is disabled before rows load and
  enabled after; click opens the modal and does **not** run; confirm runs (`FakeLister.Ran`) and
  `OnAction` fires once with `nil`. A second presenter without actions renders no `cv-action-*`
  button and no actions band.
- `actions_test.go` (non-WASM): `OnAction` receives the error when the runner fails and the list is
  not reloaded (counting lister).

## Stage 4 — docs

Verify the ARCHITECTURE "Actions" section against the code and **remove its STATUS note**. Add the
`OnAction` hook to wherever the README/ARCHITECTURE lists `OnSaved`/`OnDeleted`/`OnUpdated`.

## Acceptance criteria

- `gotest ./...` green, WASM included; all `view/conformance` clauses pass.
- `grep -rn "#[0-9a-fA-F]\{3,6\}" crudview/css.go` shows no new hex colours.
- `grep -n "STATUS (remove" docs/ARCHITECTURE.md` → empty.

| Stage | Files | Done when |
|---|---|---|
| 1 | `crudview/crudview.go` | state + hook |
| 2 | `crudview/crudview.go`, `crudview/css.go` | band + modal render |
| 3 | `crudview/conformance_test.go`, `crudview/actions_wasm_test.go`, `crudview/actions_test.go` | green |
| 4 | `docs/ARCHITECTURE.md`, README | STATUS removed |
