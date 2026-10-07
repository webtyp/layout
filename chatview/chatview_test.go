//go:build !wasm

package chatview

import (
	"errors"
	"testing"

	"webtyp.com/components/bubblethread"
	"webtyp.com/components/inboxlist"
	"webtyp.com/components/presencelist"
	"webtyp.com/lang"
)

func TestNew_Errors(t *testing.T) {
	src := &fakeSource{}

	_, err := New(Config{Source: nil, MaxBodyLength: 2000})
	if err == nil || err.Error() != "chatview.New: Source is required" {
		t.Fatalf("expected 'chatview.New: Source is required', got: %v", err)
	}

	_, err = New(Config{Source: src, MaxBodyLength: 0})
	if err == nil || err.Error() != "chatview.New: MaxBodyLength must be greater than zero" {
		t.Fatalf("expected 'chatview.New: MaxBodyLength must be greater than zero', got: %v", err)
	}

	_, err = New(Config{Source: src, MaxBodyLength: -10})
	if err == nil || err.Error() != "chatview.New: MaxBodyLength must be greater than zero" {
		t.Fatalf("expected 'chatview.New: MaxBodyLength must be greater than zero', got: %v", err)
	}
}

func TestInit_Lifecycle(t *testing.T) {
	src := &fakeSource{
		roomsData:  []inboxlist.Row{{ID: "r1", Title: "Room 1", Unread: 1}},
		peopleData: []presencelist.Person{{ID: "p1", Label: "Alice", Online: true}},
	}

	v, err := New(Config{Source: src, MaxBodyLength: 2000})
	if err != nil {
		t.Fatalf("unexpected New error: %v", err)
	}

	v.Init(NilCtx())

	if len(src.calls) != 2 || src.calls[0] != "Rooms" || src.calls[1] != "People" {
		t.Fatalf("expected Init to call Rooms then People, got: %v", src.calls)
	}

	if !v.composeDisabled.Get() {
		t.Error("expected compose bar to be disabled initially")
	}

	html := v.Render().String()
	wantPick := lang.Translate("Pick a conversation").String()
	if !contains(html, wantPick) {
		t.Errorf("expected HTML to contain %q, got: %s", wantPick, html)
	}
}

func TestOpenRoom(t *testing.T) {
	src := &fakeSource{
		roomsData: []inboxlist.Row{
			{ID: "r1", Title: "General", Unread: 2},
			{ID: "r2", Title: "Random", Unread: 3},
		},
		messagesData: map[string][]bubblethread.Bubble{
			"r1": {{ID: "m1", Body: "Hello", Author: "Bob"}},
		},
	}

	v, _ := New(Config{Source: src, MaxBodyLength: 2000})
	v.Init(NilCtx())

	src.calls = nil

	v.inbox.OnSelect("r1")

	expectedCalls := []string{"Messages:r1", "MarkRead:r1", "Rooms"}
	if len(src.calls) != 3 {
		t.Fatalf("expected %v, got %v", expectedCalls, src.calls)
	}
	for i, want := range expectedCalls {
		if src.calls[i] != want {
			t.Errorf("call %d: want %s, got %s", i, want, src.calls[i])
		}
	}

	if v.composeDisabled.Get() {
		t.Error("expected compose bar to be enabled after opening a room")
	}

	if v.headerTitle.Get() != "General" {
		t.Errorf("expected header title to be 'General', got %q", v.headerTitle.Get())
	}

	bubbles := v.thread.Bubbles()
	if len(bubbles) != 1 || bubbles[0].Body != "Hello" {
		t.Errorf("expected thread to hold 1 bubble 'Hello', got %v", bubbles)
	}
}

func TestOpenPerson(t *testing.T) {
	src := &fakeSource{
		roomsData: []inboxlist.Row{
			{ID: "r1", Title: "Direct Alice", Unread: 0},
		},
		peopleData: []presencelist.Person{
			{ID: "p1", Label: "Alice", Online: true},
		},
		openDirectRoom: "r1",
	}

	v, _ := New(Config{Source: src, MaxBodyLength: 2000})
	v.Init(NilCtx())

	src.calls = nil

	v.presence.OnSelect("p1")

	expectedCalls := []string{"OpenDirect:p1", "Rooms", "Messages:r1", "MarkRead:r1", "Rooms"}
	if len(src.calls) != len(expectedCalls) {
		t.Fatalf("expected calls %v, got %v", expectedCalls, src.calls)
	}
	for i, want := range expectedCalls {
		if src.calls[i] != want {
			t.Errorf("call %d: want %s, got %s", i, want, src.calls[i])
		}
	}

	if v.activeTab.Get() != "conversations" {
		t.Errorf("expected active tab to be 'conversations', got %q", v.activeTab.Get())
	}
}

func TestSend(t *testing.T) {
	src := &fakeSource{
		roomsData: []inboxlist.Row{
			{ID: "r1", Title: "General", Unread: 0},
		},
	}

	v, _ := New(Config{Source: src, MaxBodyLength: 2000})
	v.Init(NilCtx())
	v.inbox.OnSelect("r1")

	src.calls = nil

	v.bar.OnSend("Test message")

	if len(src.calls) != 1 || src.calls[0] != "Send:r1:Test message" {
		t.Fatalf("expected Send call, got %v", src.calls)
	}

	bubbles := v.thread.Bubbles()
	if len(bubbles) == 0 || bubbles[len(bubbles)-1].Body != "Test message" {
		t.Errorf("expected sent bubble appended to thread, got %v", bubbles)
	}

	// Test Send error handling
	var errorLogged error
	v.OnError = func(err error) {
		errorLogged = err
	}

	src.sendFunc = func(roomID, body string, done func(bubblethread.Bubble, error)) {
		done(bubblethread.Bubble{}, errors.New("network failure"))
	}

	v.bar.OnSend("Failing message")

	if errorLogged == nil || errorLogged.Error() != "network failure" {
		t.Errorf("expected OnError to receive network failure, got %v", errorLogged)
	}
	if v.composeDisabled.Get() {
		t.Error("expected compose bar to be re-enabled after error")
	}
}

func TestUnread(t *testing.T) {
	src := &fakeSource{
		roomsData: []inboxlist.Row{
			{ID: "r1", Title: "Room 1", Unread: 2},
			{ID: "r2", Title: "Room 2", Unread: 3},
		},
	}

	v, _ := New(Config{Source: src, MaxBodyLength: 2000})
	v.Init(NilCtx())

	count, visible := v.Unread()

	if count.Get() != "5" {
		t.Errorf("expected unread count '5', got %q", count.Get())
	}
	if !visible.Get() {
		t.Error("expected unread visible to be true")
	}

	v.inbox.OnSelect("r1")

	if count.Get() != "3" {
		t.Errorf("expected unread count '3' after opening r1, got %q", count.Get())
	}
}

func TestRefresh(t *testing.T) {
	src := &fakeSource{
		roomsData: []inboxlist.Row{
			{ID: "r1", Title: "Room 1", Unread: 0},
		},
	}

	v, _ := New(Config{Source: src, MaxBodyLength: 2000})
	v.Init(NilCtx())
	v.inbox.OnSelect("r1")

	src.calls = nil

	v.Refresh()

	expectedCalls := []string{"Rooms", "Messages:r1", "MarkRead:r1", "Rooms"}
	if len(src.calls) != len(expectedCalls) {
		t.Fatalf("expected calls %v, got %v", expectedCalls, src.calls)
	}
	for i, want := range expectedCalls {
		if src.calls[i] != want {
			t.Errorf("call %d: want %s, got %s", i, want, src.calls[i])
		}
	}
}

func TestRefreshPeople(t *testing.T) {
	src := &fakeSource{
		peopleData: []presencelist.Person{
			{ID: "p1", Label: "Bob", Online: true},
		},
	}

	v, _ := New(Config{Source: src, MaxBodyLength: 2000})
	v.Init(NilCtx())

	src.calls = nil

	v.RefreshPeople()

	if len(src.calls) != 1 || src.calls[0] != "People" {
		t.Fatalf("expected 1 call to People, got %v", src.calls)
	}
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// The rail badge is wired when the module is built — before the chassis ever
// calls Init (see README "Integration with platformd.Badged"). Unread must hand
// out the live signals from New, and Init must keep those same signals, or the
// badge binds to nil / to signals nobody updates.
func TestUnread_BeforeInit_SameSignalsAfterInit(t *testing.T) {
	src := &fakeSource{roomsData: []inboxlist.Row{{ID: "r1", Title: "Room 1", Unread: 2}}}
	v, err := New(Config{Source: src, MaxBodyLength: 2000})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	count, visible := v.Unread()
	if count == nil || visible == nil {
		t.Fatal("Unread() before Init returned nil signals: a badge wired at construction binds to nothing")
	}
	v.Init(nil)
	if got := count.Get(); got != "2" {
		t.Errorf("badge count after Init = %q, want 2 (Init replaced the signals the badge holds)", got)
	}
	if !visible.Get() {
		t.Error("badge visible after Init = false, want true")
	}
}
