---
PLAN: "feat: layout types its fixed UI text as lang.Text, ships lang.json, drops Go dictionaries"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 3421908638887019252
PR: https://github.com/webtyp/layout/pull/41
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — layout: fixed UI text is `lang.Text`; translations ship as `lang.json`

Phase **T5** of the master plan `SOURCE_SELECTION_MASTER_PLAN.md` (orchestration only — everything
this plan needs is inline). **Depends on three published tags:**
- `webtyp.com/lang` with `lang.Text`, the page dictionary, and no `RegisterWords`
  (`https://github.com/webtyp/lang/blob/main/docs/PLAN.md`);
- `webtyp.com/components` with `lang.Text` fields (`https://github.com/webtyp/components/blob/main/docs/PLAN.md`);
- `webtyp.com/view` whose `Presenter.SearchPlaceholder()` returns `lang.Text`.

Read [AGENTS.md](../AGENTS.md) first. Rules repeated:
- This module compiles to WASM: use `webtyp.com/fmt`, never `strings`/`strconv`/stdlib `fmt`.
- New and rewritten tests live in `tests/` (external package, public API only). A root-level test
  needs a top-of-file `// Root-level test (justified): …` comment. **Never export a symbol so a test
  can reach it.**
- Import `webtyp.com/lang`, never `webtyp.com/fmt/lang` (removed).

## Why

Translations are data now. The page carries a dictionary merged from every library's `lang.json` and
the project's `config/lang.json`, and there is no Go API to register words. `lang.Translate` looks up
each argument exactly as written: the author chooses word or phrase with the comma.

In this repo:
- `chatview/words.go` registers Spanish in Go, against the ecosystem rule that a library never
  fixes a language in code;
- the fixed titles (`CrudView.Title`, `RightPanel.Title`, `Login.Title`/`Subtitle`) are `string`,
  so neither the layout nor the `langc` generator knows they must be translated;
- crudview pre-translates the dialog title before handing it to `modaldialog`.

## Design gate

Same decision as components (see its plan): **fixed UI text is `lang.Text`, data stays `string`**,
and the component translates in `Render`. Prior art: Android `@StringRes`, Angular `i18n` attributes,
Flutter localized getters. Ledger: +0 concepts here (`lang.Text` is introduced by lang); −1 way
(Go dictionaries die). It deletes `chatview/words.go`.

## Stage 1 — retype and translate

| File | Field | Render |
|---|---|---|
| `crudview/crudview.go:90` | `CrudView.Title` → `lang.Text` | `lang.Translate(v.Title).String()` where it is shown |
| `rightpanel/rightpanel.go:69` | `RightPanel.Title` → `lang.Text` | same, in the `<h1>` |
| `login/login.go:57` (+ `Subtitle` right below) | `Login.Title`, `Login.Subtitle` → `lang.Text` | same; fix the comment examples to English (`"Sign in"`, `"Enter your credentials to continue"`) |

- `crudview/crudview.go` (~265): the confirm dialog becomes
  `&modaldialog.ModalDialog{Title: "Confirm", …}`. `modaldialog.Title` is a `lang.Text` now and is
  translated in its own `Render`; do not pre-translate. Leave `renderDeleteConfirm` as it is: it
  passes each word as its own argument on purpose (the author's choice of word-level keys).
- `crudview/crud.go:91`: `Placeholder: cfg.Presenter.SearchPlaceholder()` compiles as is once
  `view` returns `lang.Text`.
- `crudview/consumer_test.go:128`: the fake presenter's `SearchPlaceholder() string` becomes
  `SearchPlaceholder() lang.Text`, to satisfy the new `view.Presenter`.
- Fix any other call site `go build ./...` reports (a `string` variable assigned to a now-`lang.Text`
  field needs `lang.Text(v)`).

## Stage 2 — `lang.json` replaces the Go dictionary

1. Delete `chatview/words.go`.
2. Create `lang.json` at the module root. It is library format: no `default`; each value is a
   positional list in the order of `languages`; **one key per line**. Use exactly this content:
```json
{
  "languages": ["es"],
  "keys": {
    "Cancel": ["Cancelar"],
    "Confirm": ["Confirmar"],
    "Conversations": ["Conversaciones"],
    "Delete": ["Eliminar"],
    "No conversations yet": ["Aún no hay conversaciones"],
    "No messages yet": ["Aún no hay mensajes"],
    "Nobody else is here yet": ["Todavía no hay nadie más"],
    "Offline": ["Desconectado"],
    "Online": ["En línea"],
    "People": ["Personas"],
    "Pick a conversation": ["Elige una conversación"],
    "Read": ["Leído"],
    "Send": ["Enviar"],
    "This": ["Esta"],
    "Write a message": ["Escribe un mensaje"],
    "action": ["acción"],
    "be": ["se"],
    "cannot": ["no"],
    "records": ["registros"],
    "undone.": ["puede deshacer."]
  }
}
```
   (The keys are sorted by byte order, as the `langc` generator writes them.)

## Stage 3 — tests

- `crudview/crudview_test.go` uses `lang.RegisterWords` and `lang.OutLang(ES)`
  (lines ~97–125). Move that translation case to `tests/crudview_translate_test.go`
  (`//go:build wasm`). Its `TestMain` inserts
  `<script type="application/json" id="` + lang.ScriptID + `">` with
  `{"default":"es","languages":["es"],"keys":{"Confirm":["Confirmar"],"Delete":["Eliminar"],"This":["Esta"]}}`
  before any lookup. Assert under `lang.OutLang(lang.ES)` that the dialog title reads `Confirmar`
  and the message contains `Eliminar` and `Esta`. Delete that case from the package test.
- `tests/titles_test.go` (backend): `CrudView{Title: "Devices"}`, `RightPanel{Title: "Devices"}` and
  `Login{Title: "Sign in"}` render their English text as is (no dictionary on the backend).
- `tests/lang_json_test.go` (backend): `lang.json` parses as JSON. Every list has exactly
  `len(languages)` values, and every key appears in this repo's code as a `lang.Translate` argument
  or a `lang.Text` literal. Check the second rule with a plain `os.ReadFile` + `fmt.Contains` over
  the `.go` files: it guards against stale keys.

## Acceptance

- `gotest` passes (includes WASM).
- `grep -rn "RegisterWords\|DictEntry\|webtyp.com/fmt/lang" --include='*.go' .` → empty.
- `test -f lang.json && ! test -e chatview/words.go`.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Retype + translate | `crudview/crudview.go`, `crudview/crud.go`, `rightpanel/rightpanel.go`, `login/login.go`, call sites, `go.mod`, `go.sum` |
| 2 | `lang.json` | `chatview/words.go` (deleted), `lang.json` |
| 3 | Tests | `tests/*.go`, `crudview/crudview_test.go` |
## Executor notes
The tests are failing due to type mismatch errors. The components and views now require `lang.Text` for `Empty`, `Placeholder`, etc. but they are receiving `string` from `lang.Translate(...).String()`. The user explicitly asked to publish despite the test failures.
