package chatview

import (
	"webtyp.com/components/bubblethread"
	"webtyp.com/components/composebar"
	"webtyp.com/components/decktabs"
	"webtyp.com/components/inboxlist"
	"webtyp.com/components/presencelist"
	"webtyp.com/layout/rightpanel"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/lang"
)

const NameChatView widget.Name = "chatview"

const (
	partHeader widget.Part = "header"
	partWork   widget.Part = "work"
	partThread widget.Part = "thread"
)

var (
	clsRoot   = NameChatView.Root()
	clsHeader = NameChatView.Class(partHeader)
	clsWork   = NameChatView.Class(partWork)
	clsThread = NameChatView.Class(partThread)
)

func (v *ChatView) WidgetName() widget.Name { return NameChatView }
func (v *ChatView) WidgetKind() widget.Kind { return widget.Region }

type Config struct {
	Source        Source // required
	MaxBodyLength int    // required, > 0 — passed to composebar.MaxLength
}

type ChatView struct {
	Element
	OnError func(err error)

	source        Source
	maxBodyLength int

	unreadCount   *SignalString
	unreadVisible *SignalBool
	activeTab     *SignalString
	openRoomID    string
	headerTitle   *SignalString

	inbox    inboxlist.InboxList
	presence presencelist.PresenceList
	thread   bubblethread.BubbleThread
	bar      composebar.ComposeBar
	tabs     decktabs.DeckTabs

	composeDisabled *SignalBool
}

func New(cfg Config) (*ChatView, error) {
	if cfg.Source == nil {
		return nil, Err("chatview.New: Source is required")
	}
	if cfg.MaxBodyLength <= 0 {
		return nil, Err("chatview.New: MaxBodyLength must be greater than zero")
	}

	// The unread signals exist from New: the rail badge (platformd.Badged) is
	// wired when the module is built, before the chassis calls Init.
	v := &ChatView{
		source:        cfg.Source,
		maxBodyLength: cfg.MaxBodyLength,
		unreadCount:   NewString("0"),
		unreadVisible: NewBool(false),
	}
	return v, nil
}

func (v *ChatView) Init(ctx Ctx) {
	v.activeTab = NewString("conversations")
	v.headerTitle = NewString("")
	v.composeDisabled = NewBool(true)

	v.inbox.Empty = "No conversations yet"
	v.inbox.Selected = NewString("")
	v.inbox.OnSelect = func(id string) {
		v.openRoom(id)
	}

	v.presence.Empty = "Nobody else is here yet"
	v.presence.OnlineLabel = lang.Translate("Online").String()
	v.presence.OfflineLabel = lang.Translate("Offline").String()
	v.presence.OnSelect = func(personID string) {
		v.source.OpenDirect(personID, func(roomID string, err error) {
			if err != nil {
				if v.OnError != nil {
					v.OnError(err)
				}
				return
			}
			v.source.Rooms(func(rows []inboxlist.Row, rerr error) {
				if rerr == nil {
					v.updateRooms(rows)
				}
				v.openRoom(roomID)
				v.activeTab.Set("conversations")
			})
		})
	}

	v.thread.Empty = "Pick a conversation"
	v.thread.ReadLabel = lang.Translate("Read").String()

	v.bar.Placeholder = "Write a message"
	v.bar.SendLabel = "Send"
	v.bar.MaxLength = v.maxBodyLength
	v.bar.Disabled = v.composeDisabled
	v.bar.OnSend = func(body string) {
		if v.openRoomID == "" {
			return
		}
		v.composeDisabled.Set(true)
		v.source.Send(v.openRoomID, body, func(sent bubblethread.Bubble, err error) {
			v.composeDisabled.Set(false)
			if err != nil {
				if v.OnError != nil {
					v.OnError(err)
				}
				if len(v.thread.Bubbles()) == 0 {
					v.thread.Empty = lang.Text(err.Error())
				}
				return
			}
			v.thread.Append(sent)
		})
	}

	v.tabs = decktabs.DeckTabs{
		Active: v.activeTab,
		Items: []decktabs.Item{
			{
				ID:    "conversations",
				Label: "Conversations",
				Panel: v.inbox.Render(),
			},
			{
				ID:    "people",
				Label: "People",
				Panel: v.presence.Render(),
			},
		},
	}

	// Initial fetch
	v.source.Rooms(func(rows []inboxlist.Row, err error) {
		if err != nil {
			if v.OnError != nil {
				v.OnError(err)
			}
			return
		}
		v.updateRooms(rows)
		v.source.People(func(people []presencelist.Person, perr error) {
			if perr != nil {
				if v.OnError != nil {
					v.OnError(perr)
				}
				return
			}
			v.presence.SetPeople(people)
		})
	})
}

func (v *ChatView) openRoom(id string) {
	v.openRoomID = id
	v.inbox.Selected.Set(id)

	for _, r := range v.inbox.Rows() {
		if r.ID == id {
			v.headerTitle.Set(r.Title)
			break
		}
	}

	v.source.Messages(id, func(bubbles []bubblethread.Bubble, err error) {
		if err != nil {
			if v.OnError != nil {
				v.OnError(err)
			}
			if len(v.thread.Bubbles()) == 0 {
				v.thread.Empty = lang.Text(err.Error())
			}
		} else {
			v.thread.Empty = "No messages yet"
			v.thread.SetBubbles(bubbles)
		}

		v.source.MarkRead(id, func(merr error) {
			if merr == nil {
				v.source.Rooms(func(rows []inboxlist.Row, rerr error) {
					if rerr == nil {
						v.updateRooms(rows)
					}
				})
			}
		})
		v.composeDisabled.Set(false)
	})
}

func (v *ChatView) updateRooms(rows []inboxlist.Row) {
	v.inbox.SetRows(rows)
	total := 0
	for _, r := range rows {
		if r.ID != v.openRoomID {
			total += r.Unread
		}
	}
	v.unreadCount.Set(Sprint(total))
	v.unreadVisible.Set(total > 0)
}

func (v *ChatView) Refresh() {
	v.source.Rooms(func(rows []inboxlist.Row, err error) {
		if err != nil {
			if v.OnError != nil {
				v.OnError(err)
			}
			return
		}
		v.updateRooms(rows)

		if v.openRoomID != "" {
			v.source.Messages(v.openRoomID, func(bubbles []bubblethread.Bubble, merr error) {
				if merr != nil {
					if v.OnError != nil {
						v.OnError(merr)
					}
					return
				}
				v.thread.SetBubbles(bubbles)
				v.source.MarkRead(v.openRoomID, func(rerr error) {
					if rerr == nil {
						v.source.Rooms(func(newRows []inboxlist.Row, rerr2 error) {
							if rerr2 == nil {
								v.updateRooms(newRows)
							}
						})
					}
				})
			})
		}
	})
}

func (v *ChatView) RefreshPeople() {
	v.source.People(func(people []presencelist.Person, err error) {
		if err != nil {
			if v.OnError != nil {
				v.OnError(err)
			}
			return
		}
		v.presence.SetPeople(people)
	})
}

func (v *ChatView) Unread() (*SignalString, *SignalBool) {
	return v.unreadCount, v.unreadVisible
}

func (v *ChatView) Render() *Element {
	header := Div().Set(clsHeader.AsAttr()).BindText(v.headerTitle)
	if title := v.headerTitle.Get(); title != "" {
		header.Text(title)
	}

	threadBox := Div().Set(clsThread.AsAttr()).Child(v.thread.Render())

	workArea := Div().Set(clsWork.AsAttr()).Child(header, threadBox, v.bar.Render())

	panel := rightpanel.RightPanel{
		Aside:   v.tabs.Render(),
		Article: workArea,
	}

	return Div().Set(clsRoot.AsAttr()).Child(panel.Render())
}
