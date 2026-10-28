---
PLAN: "refactor!: crudview and the menu speak model.Item and view.Record"
TAG: v0.3.79
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

<!-- Dispatch after webtyp.com/model v0.2.8, webtyp.com/view v0.6.48 and webtyp.com/components v0.8.57 are published. -->

# Plan — `webtyp.com/layout`: one list-item type

## Why

`webtyp.com/model` v0.2.8 replaces `model.Row`/`model.RowOf` (a runtime walk that `view` re-ran O(N²) times per
list) with ONE type, `model.Item{ID, Label, Description, Badge}`, and a field mark that names its slot
(`Item: model.ItemLabel` / `ItemDescription` / `ItemBadge`). `ormc` v0.1.26 generates a typed `func (m *X) Item()
model.Item` from those marks, and `webtyp.com/view` v0.6.48 lists only records that have it (`view.Record`),
computing each item once. `view.Item` is gone: everything speaks `model.Item`. Breaking change, no compatibility layer.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/model@v0.2.9
go get webtyp.com/view@v0.6.48
go get webtyp.com/components@v0.8.57
```

## Ecosystem search / Design gate

Reuses `model.Item` and `view.Record`; no new name; deletes every `view.Item` in this repo (see `webtyp/view`
v0.6.48 README for the full gate).

## Stage 1 — `crudview`

Every `view.Item` in `crudview/*.go` (`ListView.SetItems`, `ListView.Items`, `Config.List`, `Config.OnSelect`,
`selectAction`, `editAction`) becomes `model.Item`. Comments that still mention `LeadTop/Main/Bottom` say
«the item's `Badge`». `Presenter.Record()` is now `view.Record`: where the code passes it as `model.Model`/`Fielder`
it still works (a `view.Record` is a `model.Model`).

## Stage 2 — `app/menu.go`

`pres.Record().Schema()` and `.ModelName()` keep working. No other change; check it compiles.

## Stage 3 — tests

- `crudview/consumer_test.go` `Device` and every test record: drop `Row:` marks; implement `Item() model.Item` by
  hand (a test double): `return model.Item{ID: d.Id, Label: d.Name, Description: d.Ip}`.
- Every test double's `List` delivers `[]view.Record`; `view.Item` → `model.Item` everywhere.
- `TestRecordIdentity_EditedValueShowsInItsRow`, `TestCrudView_OtherScreenSaveShowsHere`,
  `TestBuildMenu_SelectFollowsSource` stay green.

Check: `grep -rn "view.Item\b\|RowOf\|model.Row" --include=*.go .` → empty.

## Delivery

- `gotest` green (also `-tags browser` tests). A PR that does not compile, or whose tests could not run, is a failed
  delivery. Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`, `go mod tidy` before `ormc`.
- Never edit a generated file (`model_orm.go`) by hand: regenerate with `ormc`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `crudview/*.go`, `go.mod`, `go.sum` | `model.Item` |
| 2 | `app/menu.go` | compiles with `view.Record` |
| 3 | `*_test.go` | test doubles with a hand-written `Item()` |
