# Translation Dictionary — Consumer Guide

`layout` renders its own UI chrome (dialog titles, button labels, empty-state
texts) through `webtyp.com/lang`. Its translations ship as data in
[`lang.json`](../lang.json) at the repo root; there is no Go API to register
words. Without a page dictionary everything renders in English (the keys pass
through unchanged).

## What is translatable

| Component | Text | Keys |
|---|---|---|
| `crudview` delete dialog — title | "Confirm" | `"Confirm"` |
| `crudview` delete dialog — message | "Delete `%s`? This action cannot be undone." | `"Delete"`, `"This"`, `"action"`, `"cannot"`, `"be"`, `"undone."` (`"%s?"` passes through) |
| `crudview` delete dialog — buttons | "Cancel" / "Delete" | `"Cancel"`, `"Delete"` |
| `chatview` | tab labels, empty states, compose bar | `"Conversations"`, `"People"`, `"Send"`, `"Write a message"`, … (see `lang.json`) |
| `CrudView.Title`, `RightPanel.Title`, `Login.Title`/`Subtitle` | your text | the `lang.Text` you pass, exactly as written |

Fixed UI text is typed `lang.Text` and translated when the component renders;
data stays `string`. Each argument of `lang.Translate` is one key, exactly as
written: the author chooses word or phrase with the comma.

## How your app gets the translations

1. Add your languages to the project's `config/lang.json` (`default` and
   `languages` in its header).
2. Run `langc sync` — it keeps your file in step with the code and reports
   untranslated keys; `layout`'s own `lang.json` already covers its chrome.
3. The build inlines the merged dictionary (libraries + project, the project
   wins) in `index.html` as `<script type="application/json" id="webtyp-lang">`.
   The browser picks the user's language, then the project default, then English.
