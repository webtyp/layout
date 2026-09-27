# webtyp/layout/chatview

Layout arranging chat components (`inboxlist`, `presencelist`, `bubblethread`, `composebar`) using `rightpanel.RightPanel` and data supplied by a `Source` interface.

## Usage

```go
v, err := chatview.New(chatview.Config{
	Source:        mySource,
	MaxBodyLength: 2000,
})
if err != nil {
	log.Fatal(err)
}
v.Init(ctx)
```

## Config

- `Source`: required implementation of the `Source` interface.
- `MaxBodyLength`: required integer > 0, passed to `composebar.MaxLength`.

## Source Contract

Every method on `Source` is asynchronous: it must not block, and must invoke `done` exactly once.

```go
type Source interface {
	Rooms(done func(rows []inboxlist.Row, err error))
	Messages(roomID string, done func(bubbles []bubblethread.Bubble, err error))
	Send(roomID, body string, done func(sent bubblethread.Bubble, err error))
	MarkRead(roomID string, done func(err error))
	People(done func(people []presencelist.Person, err error))
	OpenDirect(personID string, done func(roomID string, err error))
}
```

## Host Duties

The host application has two periodic duties:

1. **Push updates:** Call `v.Refresh()` when real-time push events indicate room or message changes.
2. **Presence updates:** Call `v.RefreshPeople()` on presence ticker pulses to keep the people list updated.

## Integration with `platformd.Badged`

To show unread notification counts in the `platformd` nav rail, wire `Unread()` into a `countbadge.CountBadge`:

```go
type chatModule struct {
	platformd.UIModule
	view  *chatview.ChatView
	badge *countbadge.CountBadge
}

func newChatModule(v *chatview.ChatView) *chatModule {
	count, visible := v.Unread()
	badge := &countbadge.CountBadge{
		Count:   count,
		Visible: visible,
	}
	return &chatModule{view: v, badge: badge}
}

func (m *chatModule) Badge() *countbadge.CountBadge {
	return m.badge
}
```
