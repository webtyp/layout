---
PLAN: "feat(layout): chatview; platformd rail badge and notifier capabilities"
TAG: v0.3.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Stage F2b of** `veltylabs/mjosefa-cms/docs/MASTER.md` (private — context restated here).
> **Depends on** `webtyp.com/components` with `inboxlist`, `bubblethread`, `composebar`,
> `presencelist` (stage F2a). First line of work: `go get webtyp.com/components@latest` and
> confirm those four packages exist. If they do not, stop: never add a `replace`, never copy them
> here.
>
> Consumer: `github.com/veltylabs/chat_room/ui` (internal staff chat).

# Plan — `chatview` and two optional `platformd` capabilities

## 0. Context (verified 2026-09-26)

- A chat screen needs a layout that arranges the four chat components and talks to a data source.
  None exists. `crudview` is the model to follow: it wraps `rightpanel.RightPanel`
  (list in `Aside`, work area in `Article`) and gets its data through an interface, never a
  concrete `router.Caller`.
- `platformd.UIModule` is `ModelName() + Label() + Icon() + View()`. The rail
  (`platformd.go`, loop over `p.Modules` building `A("#"+id)` links) cannot show a count over a
  module's entry. There is **no way for a module to show a notification**: `Platform.Notify`
  exists, but no module holds the `*Platform`, and nothing in `webtyp/*` or `veltylabs/*`
  calls `Notify` today.
- `platformd` has no optional-capability checks today. The house pattern for adding one is a
  small interface + a type assertion at the seam (`if x, ok := m.(Cap); ok`), exactly how
  `router.APIModule` is detected.
- Texts go through `webtyp.com/fmt/lang` (`lang.Translate("Cancel")`, as `crudview` does).
  A package adds its words with `lang.RegisterWords([]lang.DictEntry{...})` in an `init()`.

## 1. Rules

- `platformd`, `chatview`: WASM-compiled. `webtyp.com/fmt` instead of `fmt`/`strings`/`strconv`/
  `errors`; no `map[K]V`; no `reflect`; no `encoding/json`.
- Components implement only `Render() *dom.Element` (+ optional `Init(ctx dom.Ctx)`); no manual
  update calls; state through `dom` signals. Classes only via `widget.Name`/`widget.Part` +
  `.Set(cls.AsAttr())`; styles only via `webtyp.com/widget/style` in `css.go`
  (`//go:build !wasm`). A missing `style` recipe → stop and report it in the PR; never
  hand-compose CSS.
- Ids are minted by `dom`; never compose one. Dot-import `webtyp.com/dom` and `webtyp.com/html`.
- Read `AGENTS.md` at the repo root (construction harness): typed over `any`, one way per intent,
  minimal surface, fail loudly.
- `gotest` runs native + browser; **both lanes must pass**.

## 2. Design gate

**1. Prior art.** Slack, WhatsApp Web and Element lay out a chat the same way: conversation list
+ people on the side, open thread with a send box in the main area, and a badge on the app's nav
entry. Desktop shells (VS Code activity bar, macOS dock) put the count on the nav icon, owned by
the extension/app and drawn by the shell: the module owns the number, the shell owns where it is
painted. That is the split below.

**2. Novice-name test.**
- `chatview.New(chatview.Config{Source: src, MaxBodyLength: 2000})`, `v.Refresh()`,
  `v.RefreshPeople()`, `v.Unread()` read as sentences.
- `platformd.Badged` ("a module that is badged") with `Badge() *countbadge.CountBadge`;
  `platformd.Notifier` with `Notify(...)` (the existing method's shape); `platformd.UsesNotifier`
  with `UseNotifier(n Notifier)`.

**3. Complexity ledger.**
```
Concepts                          +1 layout (chatview) +3 small interfaces (Badged, Notifier, UsesNotifier)
Lines for a module to get a badge +3 (embed UIModule, one field, one method)
Ways to show a notification        still 1 (Platform.Notify); UsesNotifier only hands it to modules
Ways to do the same thing          0
```

**4. Where it belongs.** `chatview` here, beside `crudview`: a layout arranging components.
The capabilities in `platformd`, because the rail and the notification slots are the chassis's.
Neither leaks back into `components` (which never imports `layout`).

**5. What it deletes.** Nothing; new capability.

## 3. `platformd` — two optional capabilities

New file `platformd/capabilities.go` (no build tag):

```go
// Badged is an optional capability of a UIModule: a count the nav rail draws
// over the module's entry. The module owns the signals; the chassis only paints.
type Badged interface {
	Badge() *countbadge.CountBadge
}

// Notifier raises a notification in the chassis. *Platform satisfies it.
type Notifier interface {
	Notify(t MessageType, msg string, d Duration)
}

// UsesNotifier is an optional capability of a UIModule that raises
// notifications. The chassis hands itself over once, in Init, before any
// module renders.
type UsesNotifier interface {
	UseNotifier(n Notifier)
}

var _ Notifier = (*Platform)(nil)
```

Changes in `platformd.go`:
- In `Init`, after the existing signal setup and **before** the idle-timeout check: loop
  `p.Modules`; `if u, ok := m.(UsesNotifier); ok { u.UseNotifier(p) }`.
- In the rail loop that builds each `link`: after the label span,
  `if b, ok := m.(Badged); ok { link.Child(b.Badge()) }`.
- `css.go`: the nav link part becomes an anchor host (`style.Anchor()` added to its recipe) so
  the bubble sits on the link's corner (`countbadge`'s host contract). Verify the collapsed rail
  still shows the bubble over the icon (it is `OnEdge`, out of flow).

## 4. `chatview`

Files: `chatview/chatview.go`, `chatview/source.go`, `chatview/words.go`, `chatview/css.go`
(`//go:build !wasm`), `chatview/README.md`, tests (§6).

```go
// source.go
// Source is where chatview gets and sends its data. Every method is
// asynchronous: it must not block, and must call done exactly once.
type Source interface {
	Rooms(done func(rows []inboxlist.Row, err error))
	Messages(roomID string, done func(bubbles []bubblethread.Bubble, err error))
	Send(roomID, body string, done func(sent bubblethread.Bubble, err error))
	MarkRead(roomID string, done func(err error))
	People(done func(people []presencelist.Person, err error))
	OpenDirect(personID string, done func(roomID string, err error))
}

// chatview.go
type Config struct {
	Source        Source // required
	MaxBodyLength int    // required, > 0 — passed to composebar.MaxLength
}

func New(cfg Config) (*ChatView, error)

// Refresh reloads the room list and the open room's messages. The host calls
// it when its push channel says something changed.
func (v *ChatView) Refresh()

// RefreshPeople reloads the people list. The host calls it on its presence tick.
func (v *ChatView) RefreshPeople()

// Unread is the total unread count across rooms, as the text/visibility pair
// countbadge takes. Updated after every Rooms load.
func (v *ChatView) Unread() (count *SignalString, visible *SignalBool)

// OnError is called with every error a Source returns (nil = errors are shown
// inline in the thread's empty slot only).
// Field on ChatView, set after New:
//   OnError func(err error)
```

`New` errors, exact text: `chatview.New: Source is required`,
`chatview.New: MaxBodyLength must be greater than zero`.

Behaviour:
- **Arrangement:** `rightpanel.RightPanel` with `Aside` = a `decktabs.DeckTabs` holding two tabs,
  "Conversations" (`inboxlist.InboxList`) and "People" (`presencelist.PresenceList`), and
  `Article` = a header with the open room's title + `bubblethread.BubbleThread` +
  `composebar.ComposeBar`. Phone layout is whatever `rightpanel` already does for `crudview`; do
  not add breakpoints here.
- **Init:** `Rooms`, then `People`. No room is open at start: the thread shows the "Pick a
  conversation" empty text and the compose bar is disabled.
- **Open a room** (click in `InboxList`): set it selected, header = row title, `Messages(id)` →
  `SetBubbles`, then `MarkRead(id)` → on success, reload `Rooms` (the row's unread drops to 0).
  Enable the compose bar.
- **Open a person** (click in `PresenceList`): `OpenDirect(personID)` → reload `Rooms` → open
  the returned room as above, and switch the tab to "Conversations".
- **Send** (`ComposeBar.OnSend`): disable the bar, `Send(openRoom, body)` → `Append(sent)`,
  re-enable. On error: re-enable, call `OnError`, leave the text lost (the bar already cleared)
  and show the error text in the thread's empty slot only if the thread is empty.
- **Refresh:** `Rooms` → `SetRows`; if a room is open, `Messages(open)` → `SetBubbles` (it
  replaces, so read marks update), then `MarkRead(open)`.
- **Unread total:** after every `Rooms`, sum `Unread` over the rows **except** the open one; set
  the count signal to that number and visible to `> 0`.
- Texts through `lang.Translate`. `words.go` registers, with EN and ES:
  `Conversations/Conversaciones`, `People/Personas`, `Send/Enviar`, `Write a message/Escribe un mensaje`,
  `Read/Leído`, `Online/En línea`, `Offline/Desconectado`, `No conversations yet/Aún no hay conversaciones`,
  `No messages yet/Aún no hay mensajes`, `Pick a conversation/Elige una conversación`,
  `Nobody else is here yet/Todavía no hay nadie más`. Before adding each, check the dictionary
  already registered in `webtyp.com/fmt/lang` and reuse an existing entry instead of adding a
  duplicate.
- `css.go`: only the header part and the stacking of thread + bar in the article
  (`style.Stack`, thread `style.Grow()`). Everything else is styled by the components themselves.

## 5. `platformd` README and `layout` docs

- `platformd/README.md`: a "Module capabilities" section with `Badged` and `UsesNotifier`, a
  three-line example of a module embedding `platformd.UIModule`:
  ```go
  type chatModule struct {
  	platformd.UIModule
  	badge *countbadge.CountBadge
  	notify platformd.Notifier
  }
  func (m *chatModule) Badge() *countbadge.CountBadge { return m.badge }
  func (m *chatModule) UseNotifier(n platformd.Notifier) { m.notify = n }
  ```
- `chatview/README.md`: `Source` contract, `Config`, the host's two duties (call `Refresh` on push,
  `RefreshPeople` on its presence tick), wiring `Unread()` into a `countbadge` for `Badged`.
- `docs/ARCHITECTURE.md`: add `chatview` next to `crudview`.

## 6. Tests

`platformd` (existing test files; add cases):
1. A module implementing `UsesNotifier` receives the `*Platform` exactly once, during `Init`,
   before `Render`.
2. A module implementing `Badged` gets its badge rendered inside its rail link; a plain module
   gets none. `Platform.Notify` called through the handed `Notifier` produces a notification node
   (consumer-shaped: a fake module calls it, the chassis shows it).

`chatview` (`chatview/chatview_test.go` `!wasm`, `chatview/chatview_wasm_test.go` `wasm`), with a
**fake `Source`** that records calls and answers from in-memory slices — and the **real** four
components (no component doubles):
3. `New` errors with the exact texts.
4. Init calls `Rooms` then `People`; compose bar disabled; empty text is "Pick a conversation"
   (translated).
5. Clicking a room calls `Messages(id)` then `MarkRead(id)` then `Rooms`, in that order; the
   thread shows the bubbles; compose bar enabled.
6. Clicking a person calls `OpenDirect`, then opens the returned room and switches to the
   "Conversations" tab.
7. Sending calls `Send(openRoom, body)` and appends the returned bubble; a `Send` error calls
   `OnError` and re-enables the bar.
8. `Unread()` equals the sum of unread over rows other than the open room, and is not visible at 0.
9. `Refresh` with a room open calls `Rooms`, `Messages(open)` and `MarkRead(open)`.

## 7. Stages

| # | Stage | Files | Acceptance |
|---|---|---|---|
| L1 | `platformd` capabilities | `platformd/capabilities.go`, `platformd/platformd.go`, `platformd/css.go`, tests | tests 1–2 green; existing `platformd` tests unchanged and green |
| L2 | `chatview` | `chatview/*` | tests 3–9 green |
| L3 | Docs + close | READMEs, `docs/ARCHITECTURE.md` | `grep -rn "map\[\|TODO\|\.Class(\"" chatview platformd/capabilities.go` empty; `gotest` green on both lanes |
