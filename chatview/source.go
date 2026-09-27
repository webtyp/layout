package chatview

import (
	"webtyp.com/components/bubblethread"
	"webtyp.com/components/inboxlist"
	"webtyp.com/components/presencelist"
)

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
