---
PLAN: "feat!: crudview lists paged and searched on the server (view v0.6.64: Search/HasMore/More)"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# crudview — paged lists searched on the server

`webtyp.com/view` v0.6.64 changed its contract: a list loads what the module declares (`view.Page`) and searches on
the server. `crudview` still calls the deleted `Presenter.Filter` (`crudview/crudview.go:835`), so this repo — and
every module that renders with it — does not compile against the new `view`. This plan moves `crudview` to the new
contract and adds the «Show more results» control.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/view@v0.6.64
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`, run `go mod tidy` after `go get`.

## APIs used (webtyp.com/view v0.6.64)

| Symbol | Exact signature | File |
|---|---|---|
| `view.Presenter.Reload` | `Reload(done func(error))` — loads `Page.First` rows (0 = none, the screen waits for a search; `All` = up to `MaxPage`) and clears the search | view.go |
| `view.Presenter.Search` | `Search(term string, done func(error))` — server search, `Page.Size` rows; `""` = same as `Reload` | view.go |
| `view.Presenter.HasMore` | `HasMore() bool` — the last page was full | view.go |
| `view.Presenter.More` | `More(done func(error))` — appends the rows after the last one shown, same search; does nothing when `!HasMore()` | view.go |
| `view.Presenter.Items` | `Items() []model.Item` — listed rows plus those saved since | view.go |
| `view.Lister.List` | `List(args view.ListArgs, done func(rows []view.Record, err error))` | lister.go |
| `view.Lister.Page` | `Page() view.Page` | lister.go |
| `view.Page` | `struct{ First, Size int; More bool }`; `view.All = -1`; `view.MaxPage = 500` | page.go |
| `view.ListArgs` | `struct{ Search, After string; Limit int64 }` | page.go |

`Presenter.Filter` no longer exists. If a signature in the published tag differs from this table, stop and name the
symbol in the PR; never adapt it by guessing.

## Stage 1 — the search goes to the server (`crudview/crudview.go`)

Replace the body of `func (v *CrudView) filter()` (line 834). Today it filters locally:

```go
v.list.SetItems(v.Presenter.Filter(v.search.Get()))
```

New behaviour: a term change asks the server, and the list is painted when the answer arrives. Split it in two:

```go
// filter asks the server for the current term; paint shows the result.
func (v *CrudView) filter() {
	v.Presenter.Search(v.search.Get(), func(err error) {
		if err != nil {
			Log(err.Error())
			if v.OnLoadError != nil {
				v.OnLoadError(err)
			}
			return
		}
		v.paint()
	})
}

// paint repopulates the list from the presenter and keeps the invariants filter() kept before.
func (v *CrudView) paint() {
	v.list.SetItems(v.Presenter.Items())
	// … every line that followed the old SetItems call in filter() stays here unchanged:
	//     hasRows, hasMultiRows, itemsLoaded, dropSelectionOutOfScope, …
	if v.hasMore != nil {
		v.hasMore.Set(v.Presenter.HasMore())
	}
}
```

- In `Reload` (line ~552) replace the call `v.filter()` with `v.paint()`: `Presenter.Reload` already loaded the rows;
  calling `filter()` there would ask the server a second time.
- Every other caller of `v.filter()` that runs after a local change (save, delete, records changed) calls
  `v.paint()` instead; only the search control's `OnFilterChange` callback (line ~289) calls `v.filter()`.
  Verify with `grep -n "v.filter()" crudview/crudview.go` → exactly one hit, inside `OnFilterChange`.

## Stage 2 — «Show more results» at the end of the list

- New field `hasMore *SignalBool` on `CrudView`, created with `NewBool(false)` next to `v.hasRows` (line ~261).
- In the render code that builds `asideContent` (line ~993), after `Child(v.list)` and after the empty-message
  paragraph, add a button shown only while `v.hasMore` is true:

```go
asideContent.Child(
	Button().Set(clsMoreBtn.AsAttr()).Text(lang.Translate(textShowMore).String()).
		On(Click, func() {
			v.Presenter.More(func(err error) {
				if err != nil {
					Log(err.Error())
					return
				}
				v.paint()
			})
		}).
		BindStateFunc(widget.Open, func() bool { return v.hasMore.Get() }),
)
```

  Use the same element builders, `On(Click, …)` form and `BindStateFunc` pattern the file already uses for its other
  buttons; `clsMoreBtn` is a new class constant declared next to `clsListaBox`, styled in `crudview/css.go` like the
  existing secondary buttons (a full-width text button). Do not invent new style recipes.
- When the search returned rows but `HasMore()` is false and a term is set, show a muted line under the list with
  `textNoMore` (same `BindStateFunc` pattern: open when `v.search.Get() != "" && v.hasRows.Get() && !v.hasMore.Get()`).
- Text keys (English, in the code; translations in `lang.json`):

```go
const (
	textShowMore lang.Text = "Show more results"
	textNoMore   lang.Text = "No more results for this search"
)
```

## Stage 3 — test doubles and tests

Every type in this repo that implements `view.Lister` gets the new method set. The sites (grep
`func (.*) List(done` → these 8):

`crudview/actions_test.go:34`, `crudview/record_identity_test.go:31`, `crudview/record_identity_test.go:278`,
`crudview/consumer_test.go:82`, `crudview/readonly_test.go:45`, `tests/menu_test.go:281`, `crudview/bulk_test.go:346`,
`crudview/shared_records_test.go:14`.

For each: `List(done …)` → `List(_ view.ListArgs, done …)` and add `func (x *T) Page() view.Page { return
view.Page{First: view.All} }` (they list everything today; `All` keeps that). Replace every `.Filter(` call in tests
with `Search(term, nil)` followed by reading `Items()`.

New tests in `crudview/paging_test.go` (stdlib, no wasm), with a double `pagedLister` that answers like a server
(rows whose label contains `args.Search`, after `args.After`, at most `args.Limit`):

| Test | Setup → call → expected |
|---|---|
| `TestCrudView_SearchAsksTheServer` | double with 5 rows "maria a…e" + "pedro", `Page{First: 0, Size: 2, More: true}`; `Reload(nil)` → list empty and the double saw 0 calls; set the search term "maria" through the filter control → the double's last args `Search == "maria"`, `Limit == 2`; the list shows 2 rows |
| `TestCrudView_ShowMoreAppendsTheNextPage` | same double; after the search, `hasMore` is true; trigger the «Show more results» action → last args `After` = id of the 2nd row; the list shows 4; trigger again → 5 rows and `hasMore` false |
| `TestCrudView_ReloadDoesNotSearchTwice` | double with `Page{First: view.All}`; `Reload(nil)` → exactly 1 `List` call |

Add each test's row to `docs/TESTS.md` if the repo has one.

## Stage 4 — translations

Run `langc sync .`; fill the `es` (and every other column `languages` lists) for `Show more results` («Mostrar más
resultados») and `No more results for this search` («No hay más resultados para esta búsqueda»). The delivery check
requires 0 untranslated keys and 0 reuse warnings.

## Stage 5 — documentation

- `README.md`: in the step-by-step section, the list paragraph says the list is loaded as the module's
  `view.Page` declares, searched on the server, and shows «Show more results» while there are more rows; link
  `https://github.com/webtyp/view/blob/main/docs/LIST_PAGING.md`. Remove any sentence that says the search filters
  locally.
- `docs/ARCHITECTURE.md`: the crudview section describes `filter()` (asks the server) and `paint()` (repaints from
  `Items()`), and the «Show more results» control. Keep its 4 mandatory sections.

## Delivery check

```
gotest
langc sync .
grep -rn "\.Filter(" --include=*.go .        # → empty
grep -n "v.filter()" crudview/crudview.go    # → exactly 1 line (OnFilterChange)
```

A PR that does not compile, or whose tests could not run, is a failed delivery. Commit after every stage; never run
`git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work. Stuck for 30 minutes: open the
PR anyway listing what fails.

| Stage | Files |
|---|---|
| 1 | `crudview/crudview.go` |
| 2 | `crudview/crudview.go`, `crudview/css.go` |
| 3 | the 8 test files above, `crudview/paging_test.go` (new), `docs/TESTS.md` if present |
| 4 | `lang.json` |
| 5 | `README.md`, `docs/ARCHITECTURE.md` |
