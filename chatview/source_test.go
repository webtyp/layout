package chatview

import (
	"webtyp.com/components/bubblethread"
	"webtyp.com/components/inboxlist"
	"webtyp.com/components/presencelist"
	. "webtyp.com/dom"
)

type fakeSource struct {
	calls []string

	roomsFunc      func(done func([]inboxlist.Row, error))
	messagesFunc   func(roomID string, done func([]bubblethread.Bubble, error))
	sendFunc       func(roomID, body string, done func(bubblethread.Bubble, error))
	markReadFunc   func(roomID string, done func(error))
	peopleFunc     func(done func([]presencelist.Person, error))
	openDirectFunc func(personID string, done func(string, error))

	roomsData      []inboxlist.Row
	messagesData   map[string][]bubblethread.Bubble
	peopleData     []presencelist.Person
	openDirectRoom string
}

func (f *fakeSource) record(call string) {
	f.calls = append(f.calls, call)
}

func (f *fakeSource) Rooms(done func([]inboxlist.Row, error)) {
	f.record("Rooms")
	if f.roomsFunc != nil {
		f.roomsFunc(done)
		return
	}
	done(f.roomsData, nil)
}

func (f *fakeSource) Messages(roomID string, done func([]bubblethread.Bubble, error)) {
	f.record("Messages:" + roomID)
	if f.messagesFunc != nil {
		f.messagesFunc(roomID, done)
		return
	}
	done(f.messagesData[roomID], nil)
}

func (f *fakeSource) Send(roomID, body string, done func(bubblethread.Bubble, error)) {
	f.record("Send:" + roomID + ":" + body)
	if f.sendFunc != nil {
		f.sendFunc(roomID, body, done)
		return
	}
	done(bubblethread.Bubble{ID: "b_new", Body: body, Mine: true}, nil)
}

func (f *fakeSource) MarkRead(roomID string, done func(error)) {
	f.record("MarkRead:" + roomID)
	if f.markReadFunc != nil {
		f.markReadFunc(roomID, done)
		return
	}
	done(nil)
}

func (f *fakeSource) People(done func([]presencelist.Person, error)) {
	f.record("People")
	if f.peopleFunc != nil {
		f.peopleFunc(done)
		return
	}
	done(f.peopleData, nil)
}

func (f *fakeSource) OpenDirect(personID string, done func(string, error)) {
	f.record("OpenDirect:" + personID)
	if f.openDirectFunc != nil {
		f.openDirectFunc(personID, done)
		return
	}
	done(f.openDirectRoom, nil)
}

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

func NilCtx() Ctx {
	return &mockCtx{}
}
