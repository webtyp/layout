---
PLAN: "feat!: lists and selects follow the shared records (view.Records); no reload after save; selects from Ref/Options"
TAG: v0.3.78
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 11831450411581804107
---

> Read AGENTS.md at the repo root before any change.

<!-- PARKED: becomes docs/PLAN.md when the running plan (menu: Component, explorer, closed by default — v0.3.77)
     closes. Dispatch only after webtyp.com/view v0.6.47, webtyp.com/components v0.8.56 and webtyp.com/model v0.2.7
     are published. -->

# Plan — `webtyp.com/layout`: what is saved shows at once in the list and in every select

## Why

`webtyp.com/view` v0.6.47 keeps **one** copy of each record per page (`view.Records`) and tells every presenter of a
model when it changes (`Presenter.OnChange`); a save puts the server's answer there. So `crudview` no longer
reloads the whole list after a save, and a select (a field with `Ref` or `Options`, `webtyp.com/model` v0.2.7) is
filled from the presenter of that model and follows it live: renaming a specialty changes the services' select at
once. `view.Lookup` and `Screen.Lookups` are gone; `layout/app` finds the source by the field's model.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/view@v0.6.47
go get webtyp.com/components@v0.8.56
go get webtyp.com/model@v0.2.7
```

If a tag lacks a symbol of «APIs used», stop and say so in the PR.

## Ecosystem search

`view.Records`, `Presenter.OnChange`, `Lister.Records` (view v0.6.47); `model.Field.Ref`, `model.Field.Options`
(model v0.2.7); `crudview.SetOptions` (this repo) → reuse. Nothing new is created.

## Design gate

1. **Prior art.** Apollo / Ember Data / Relay: every view of an entity re-renders from the normalised store; React
   Admin `<ReferenceInput>` fills a select from the referenced resource and refreshes with it.
2. **Novice-name test.** No new exported name.
3. **Complexity ledger.** Code in every module −1 `Lookups` block per select; ways to refresh a list after save −1
   (the store, not a reload).
4. **Where it belongs.** Drawing lists and selects is `crudview` / `layout/app`.
5. **What it deletes.** The reload + reorder after a save (`crudview.go:755-772`), the reload after a delete,
   `onFirstActivation`/`fillOptions`/`lookups` in `layout/app/menu.go`.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `view.Presenter.OnChange` | `OnChange(fn func())` | view v0.6.47 | `view.go` |
| `view.Presenter.Record` | `Record() model.Model` | view v0.6.47 | `view.go` |
| `view.Lister.Records` | `Records() *Records` | view v0.6.47 | `lister.go` |
| `view.Records` | `NewRecords() *Records`; `Put(rows ...model.Model)`; `Remove(modelName string, ids ...string)` | view v0.6.47 | `records.go` |
| `view.Item` | `struct { ID, Label, Description, Badge string }` | view v0.6.47 | `view.go` |
| `model.Field.Ref`, `model.Field.Options` | `*Definition` (its `Name` is the model name) | model v0.2.7 | `field.go` |
| `(*crudview.CrudView).SetOptions` | `func (v *CrudView) SetOptions(fieldName string, opts ...fmt.KeyValue)` | this repo | `crudview.go:525` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — `crudview` follows the shared records (`crudview/crudview.go`)

1. Where the CrudView is constructed (in `New`, after the presenter is set), subscribe once:
   `cfg.Presenter.OnChange(func() { if v.list != nil { v.filter() } })`.
2. `saveAction`, success branch: delete the whole `v.Reload(func(err error) { … reorder … })` block. Keep
   `v.form.MarkPristine()`, the composing reset and `v.reportSave(nil)`. A new record appears first by itself
   (`Presenter.Items()` puts records it never listed first), an edited one keeps its place.
3. `deleteAction` and `confirmDeleteAction`: delete the `v.Reload(nil)` after a successful delete (the lister
   removes the ids from the shared records, which runs `filter()`). Keep the success notice and the selection reset.
4. Bulk edit (`bulkEditAction`) keeps its `v.Reload(nil)` (a patch of N rows answers no records).

Check: `grep -n "reordered\|v.Reload(" crudview/crudview.go` → only `Activate`/`Reload` itself and `bulkEditAction`.

## Stage 2 — selects from `Ref`/`Options` (`app/menu.go`)

Delete `onFirstActivation`, `fillOptions` and every use of `s.Lookups`. After all screens are built (before the
explorer is added), wire the selects:

```go
const errNoOptionSource = "layout/app: a field takes its options from a model no screen lists: "

// leafView pairs a built list screen with its CrudView, so selects can be wired after
// every screen exists. buildScreen appends one per list leaf.
type leafView struct {
	view *crudview.CrudView
	pres view.Presenter
}

// wireOptions fills every select of every list screen from the presenter that lists the
// field's Ref/Options model, and keeps it in step: the source's OnChange refills it.
func wireOptions(leaves []leafView) error {
	for _, target := range leaves {
		for _, f := range target.pres.Record().Schema() {
			src := f.Options
			if src == nil {
				src = f.Ref
			}
			if src == nil {
				continue
			}
			source, ok := sourceFor(leaves, src.Name)
			if !ok {
				return fmt.Err(errNoOptionSource + target.pres.Record().ModelName() + "." + f.Name + " → " + src.Name)
			}
			fill := func(v *crudview.CrudView, field string, s view.Presenter) func() {
				return func() {
					items := s.Items()
					opts := make([]fmt.KeyValue, 0, len(items))
					for _, it := range items {
						opts = append(opts, fmt.KeyValue{Key: it.ID, Value: it.Label})
					}
					v.SetOptions(field, opts...)
				}
			}(target.view, f.Name, source)
			source.OnChange(fill)
			fill()
		}
	}
	return nil
}

func sourceFor(leaves []leafView, modelName string) (view.Presenter, bool) {
	for _, l := range leaves {
		if l.pres.Record().ModelName() == modelName {
			return l.pres, true
		}
	}
	return nil, false
}
```

A source screen loads its rows when it is first activated (crudview `Activate` → `Reload`); additionally, when a
screen with selects is activated, reload each source presenter once (`source.Reload(nil)`) so its select is filled
even if the source screen was never opened — put that in the `onActivate` func passed to `layout.NewUIModule` for
the target screen. `buildScreen` returns the `*crudview.CrudView` it builds so `BuildMenu` collects `leaves`.
`BuildMenu` returns `Menu{}, err` when `wireOptions` fails.

Only fields with an input Kind (`f.Type` implements `input.Input`) are selects; skip the rest
(`if _, isInput := f.Type.(input.Input); !isInput { continue }`, import `webtyp.com/input`).

## Stage 3 — tests

Every test double that implements `view.Lister` gains `Records() *view.Records` (a field set to
`view.NewRecords()`), puts what it saves and removes what it deletes (`memoryBackend`, `laterBackend` in
`crudview/record_identity_test.go`, the doubles of `crudview/consumer_test.go` and `tests/menu_test.go`). The test
record `Device` (`crudview/consumer_test.go:41`) loses `Item()`; its schema marks `name` `Row: model.RowTitle`,
`ip` `Row: model.RowDetail`, `id` as PK.

| Test | Expected |
|---|---|
| `TestRecordIdentity_EditedValueShowsInItsRow` (exists) | still green, now without a reload |
| `TestRecordIdentity_CreatedItemAppearsFirstInList` (exists) | still green |
| `TestCrudView_OtherScreenSaveShowsHere` (new, `crudview/`) | two CrudViews over two listers sharing one `view.Records`, same model; save a renamed record in the first → the second's `list.Items()` shows it, with no call to its `Reload` |
| `TestBuildMenu_SelectFollowsSource` (new, `tests/menu_test.go`) | screens «specialty» and «service» (service schema: `specialty_id` `Type: input.Select()`, `Ref: &specialtyDef`); rename a specialty through the specialty presenter's lister → the service CrudView's options for `specialty_id` carry the new label |
| `TestBuildMenu_SelectWithoutSourceFails` (new) | a field with `Options` naming a model no screen lists → error containing `a field takes its options from a model no screen lists` |

## Stage 4 — docs and translations

- `docs/ARCHITECTURE.md` («CRUD Layout» and «Chassis contracts and menu assembly»): the list follows
  `Presenter.OnChange`, no reload after save; selects come from the field's `Ref`/`Options` and follow their source.
- `langc sync .`; fill every empty value; zero untranslated keys, zero reuse warnings.

## Delivery

- `gotest` green (also `-tags browser` in headless Chrome). A PR that does not compile, or whose tests could not
  run, is a failed delivery. Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `crudview/crudview.go`, `go.mod`, `go.sum` | follow `OnChange`; no reload after save/delete |
| 2 | `app/menu.go` | selects from `Ref`/`Options`; `Lookups` gone |
| 3 | `crudview/*_test.go`, `tests/menu_test.go` | doubles and behaviour tests |
| 4 | `docs/ARCHITECTURE.md`, `lang.json` | document, translate |
