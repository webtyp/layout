//go:build !wasm

package platformd

import (
	"sync/atomic"
	"testing"
	"time"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/svg"
)

type mockModule struct {
	id    string
	label string
	icon  svg.Icon
	view  Component
}

func (m *mockModule) ModelName() string { return m.id }
func (m *mockModule) Label() string     { return m.label }
func (m *mockModule) Icon() svg.Icon    { return m.icon }
func (m *mockModule) View() Component   { return m.view }

func TestPlatform_Render(t *testing.T) {
	p := &Platform{
		AppName: "Test App",
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Module 1"},
			&mockModule{id: "mod2", label: "Module 2"},
		},
	}
	p.Init(NilCtx())

	el := p.Render()
	html := el.String()

	if html == "" {
		t.Fatal("String() returned empty string")
	}

	expected := []string{
		"pd",
		"pd__header",
		"pd__menu",
		"pd__stage",
		"pd__panel",
		"mod1",
		"mod2",
	}

	for _, s := range expected {
		if !contains(html, s) {
			t.Errorf("expected HTML to contain %q", s)
		}
	}
}

func TestPlatform_Render_Brand(t *testing.T) {
	p := &Platform{
		Brand: testBrand{name: "Acme", mark: "https://example.com/logo.svg"},
	}
	p.Init(NilCtx())

	html := p.Render().String()
	if !contains(html, "pd__brand") {
		t.Error("expected the brand slot")
	}
	if !contains(html, "src='https://example.com/logo.svg'") {
		t.Error("expected the brand mark img")
	}
	if !contains(html, "alt='Acme'") {
		t.Error("expected the brand name as the mark's alt")
	}
	if !contains(html, "pd__brand-name") {
		t.Error("expected the brand name part")
	}
	if !contains(html, ">Acme<") {
		t.Error("expected the brand name text")
	}
}

func TestPlatform_Render_BrandEmptyMark(t *testing.T) {
	// Empty mark is a normal, expected outcome: the shell falls back to its
	// own glyph, exactly as a missing avatar does.
	p := &Platform{
		Brand: testBrand{name: "Acme", mark: ""},
	}
	p.Init(NilCtx())

	html := p.Render().String()
	if !contains(html, "pd__brand") {
		t.Error("expected the brand slot")
	}
	if !contains(html, "href='#pd-brand'") {
		t.Error("expected the default brand glyph when the mark is empty")
	}
	if contains(html, "<img") {
		t.Error("an empty mark must not render an <img>")
	}
}

func TestPlatform_Render_NoBrand(t *testing.T) {
	// Brand is optional: a platform without a logo is not a broken platform.
	p := &Platform{}
	p.Init(NilCtx())

	html := p.Render().String()
	if contains(html, "pd__brand") {
		t.Error("nil Brand must not render a brand slot")
	}
}

func TestPlatform_Render_DefaultModule(t *testing.T) {
	p := &Platform{
		Element:   *Div(),
		DefaultID: "mod2",
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Mod 1"},
			&mockModule{id: "mod2", label: "Mod 2"},
		},
	}
	p.Init(NilCtx()) // Should activate mod2

	html := p.Render().String()
	// webtyp/dom uses single quotes for attributes and boolean attributes are empty keys
	if !contains(html, "class='pd__panel' data-id='mod2' data-current='true'") {
		t.Errorf("expected mod2 to be active, got HTML: %s", html)
	}
}

func TestPlatform_Activate(t *testing.T) {
	p := &Platform{
		Element: *Div(),
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Mod 1"},
			&mockModule{id: "mod2", label: "Mod 2"},
		},
	}
	p.Init(NilCtx())
	p.Activate("mod2")

	html := p.Render().String()
	if !contains(html, "class='pd__panel' data-id='mod2' data-current='true'") {
		t.Errorf("expected mod2 to be active after Activate('mod2'), got HTML: %s", html)
	}
	if contains(html, "class='pd__panel' data-id='mod1' data-current='true'") {
		t.Errorf("expected mod1 to NOT be active")
	}
}

func TestPlatform_NewUIModule(t *testing.T) {
	m := NewUIModule("test", "Label", svg.Icon("icon"), Div())
	if m.ModelName() != "test" {
		t.Errorf("expected test, got %s", m.ModelName())
	}
	if m.Label() != "Label" {
		t.Errorf("expected Label, got %s", m.Label())
	}
	if m.Icon() != svg.Icon("icon") {
		t.Errorf("expected icon, got %s", m.Icon())
	}
	if m.View() == nil {
		t.Error("expected view to be non-nil")
	}
}

func TestPlatform_CanView(t *testing.T) {
	p := &Platform{
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Mod 1"},
			&mockModule{id: "mod2", label: "Mod 2"},
		},
		CanView: func(id string) bool {
			return id == "mod2"
		},
	}
	p.Init(NilCtx())

	html := p.Render().String()

	// Nav rail and stage should only have mod2
	if contains(html, "data-id='mod1'") {
		t.Error("expected mod1 NOT to be rendered")
	}
	if !contains(html, "data-id='mod2'") {
		t.Error("expected mod2 to be rendered")
	}

	// mod2 should be active (fallback)
	if !contains(html, "class='pd__panel' data-id='mod2' data-current='true'") {
		t.Error("expected mod2 to be active")
	}

	// Try activating mod1
	p.Activate("mod1")
	if p.active.Get() == "mod1" {
		t.Error("Activate should not set active to non-viewable module")
	}
}

func TestPlatform_Notify_Renders(t *testing.T) {
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())
	p.Notify(Msg.Error, "boom", Persistent())

	html := p.Render().String()
	t.Logf("HTML: %s", html)

	// Unified slot
	if !contains(html, "id='pd-msg-slot'") {
		t.Error("expected pd-msg-slot")
	}
	if !contains(html, "boom") {
		t.Error("expected boom")
	}

	// The same notification renders into the mobile stack too — one element
	// cannot have two parents, so the second copy gets its own id/key suffix.
	if !contains(html, "pd__msg-stack") {
		t.Error("expected the mobile msg-stack")
	}
	if !contains(html, "pd__msg-slot-mobile") {
		t.Error("expected the mobile msg slot")
	}
}

func TestPlatform_Notify_A11yRoles(t *testing.T) {
	// role=status (polite) for info/success; role=alert (assertive) for
	// warnings and errors — announcements, never focus steals.
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())

	p.Notify(Msg.Info, "informational", Persistent())
	p.Notify(Msg.Success, "ok", Persistent())
	p.Notify(Msg.Warning, "careful", Persistent())
	p.Notify(Msg.Error, "broken", Persistent())

	nodes := p.notifications.Get()
	if len(nodes) != 4 {
		t.Fatalf("expected 4 toasts, got %d", len(nodes))
	}
	for i, want := range []string{"status", "status", "alert", "alert"} {
		if html := nodes[i].String(); !contains(html, "role='"+want+"'") {
			t.Errorf("toast %d: expected role=%s, got: %s", i, want, html)
		}
	}

	// The mobile copies carry the same semantics.
	mobile := p.notificationsMobile.Get()
	if len(mobile) != 4 {
		t.Fatalf("expected 4 mobile toasts, got %d", len(mobile))
	}
	for i, want := range []string{"status", "status", "alert", "alert"} {
		if html := mobile[i].String(); !contains(html, "role='"+want+"'") {
			t.Errorf("mobile toast %d: expected role=%s, got: %s", i, want, html)
		}
	}
}

func TestPlatform_Notify_Dismiss(t *testing.T) {
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())
	p.Notify(Msg.Info, "hi", For(10)) // 10ms

	if p.notificationCount() != 1 {
		t.Fatalf("expected 1 notification, got %d", p.notificationCount())
	}

	// Wait for dismissal
	time.Sleep(100 * time.Millisecond)

	if p.notificationCount() != 0 {
		t.Errorf("expected 0 notifications after dismissal, got %d", p.notificationCount())
	}
}

func TestPlatform_Notify_ManualDismiss(t *testing.T) {
	// A persistent toast must survive its window untouched — only a tap (or
	// the consumer) removes it. dismissal via a tap is the dismiss(id) path;
	// the timer was never armed, so nothing can fire it later.
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())
	p.Notify(Msg.Info, "stay", Persistent())

	if p.notificationCount() != 1 {
		t.Fatalf("expected 1 notification, got %d", p.notificationCount())
	}
	time.Sleep(30 * time.Millisecond)
	if p.notificationCount() != 1 {
		t.Errorf("persistent notification must not auto-dismiss, got %d", p.notificationCount())
	}
}

func TestPlatform_Notify_AutoDuration(t *testing.T) {
	// Auto() sizes the window to the message: floor 2s for a one-word toast,
	// then ~350ms per extra word, capped at 8s. The deadline is computed at
	// Notify time and stays fixed across pause/resume.
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())

	if got := autoMillis("Guardado"); got != 2000 {
		t.Errorf("one word: expected 2000ms, got %d", got)
	}
	if got := autoMillis("Dispositivo guardado correctamente"); got != 2250 {
		t.Errorf("three words: expected 2250ms, got %d", got)
	}
	// Exactly 15 words: 1200 + 15×350 = 6450ms
	long := "un error muy largo de quince palabras a b c d e f g h"
	if got := autoMillis(long); got != 6450 {
		t.Errorf("fifteen words: expected 6450ms, got %d", got)
	}
	if got := autoMillis(Repeat("muy ", 30) + "larga"); got != 8000 {
		t.Errorf("long message must be capped at 8000ms, got %d", got)
	}

	p.Notify(Msg.Success, "Guardado", Auto())
	if n := p.notifications.Get(); len(n) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(n))
	}
}

func TestPlatform_Notify_PauseResume(t *testing.T) {
	// Hovering/focusing a toast pauses it; leaving resumes with the time
	// remaining on the original deadline. A deadline that passes entirely
	// while paused stays — the user is still reading it, which is the whole
	// point of pausing.
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())

	p.Notify(Msg.Info, "linger", For(60))
	id := p.rawNotifications[0].ID
	p.pauseToast(id)
	time.Sleep(150 * time.Millisecond) // well past the original 60ms window
	if p.notificationCount() != 1 {
		t.Fatalf("paused toast must not dismiss, got %d", p.notificationCount())
	}

	// A pause inside the window is a true pause: resume re-arms with the
	// remaining time, so the toast still goes away shortly after. The linger
	// toast above stays by design, so exactly one must remain at the end.
	p.Notify(Msg.Info, "resume", For(100))
	id = p.rawNotifications[1].ID
	p.pauseToast(id)
	time.Sleep(30 * time.Millisecond)
	p.resumeToast(id)
	time.Sleep(200 * time.Millisecond) // ~70ms remaining, comfortably under
	if p.notificationCount() != 1 {
		t.Errorf("resumed toast must dismiss after its remaining window (linger stays), got %d", p.notificationCount())
	}
	if p.rawNotifications[0].Msg != "linger" {
		t.Errorf("the surviving toast must be the paused-past-deadline one, got %q", p.rawNotifications[0].Msg)
	}
}

func TestPlatform_IdleLock_DisabledByDefault(t *testing.T) {
	p := &Platform{Element: *Div()}
	// Should not panic when IdleTimeout is 0 and OnIdle is nil
	p.Init(NilCtx())
	if p.idleTimer != nil {
		t.Error("expected idleTimer to be nil when IdleTimeout is 0")
	}
}

func TestPlatform_IdleLock_PanicOnNilOnIdle(t *testing.T) {
	p := &Platform{
		Element:     *Div(),
		IdleTimeout: 30,
		OnIdle:      nil,
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when IdleTimeout > 0 and OnIdle is nil")
		}
		errMsg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic message, got %v", r)
		}
		want := "platformd: IdleTimeout requires OnIdle"
		if errMsg != want {
			t.Errorf("expected panic message %q, got %q", want, errMsg)
		}
	}()

	p.Init(NilCtx())
}

func TestPlatform_IdleLock_FiresOnce(t *testing.T) {
	// firedCount is written from OnIdle, which runs on the timer's own
	// goroutine (time.AfterFunc), while the test reads it from the main
	// goroutine — time.Sleep gives no happens-before guarantee between the
	// two, so the counter needs atomic access, not a plain int.
	var firedCount atomic.Int32
	p := &Platform{
		Element:     *Div(),
		IdleTimeout: 1, // 1 second timeout
		OnIdle: func() {
			firedCount.Add(1)
		},
	}
	p.Init(NilCtx())
	_ = p.Render() // arms idle timer via activity()

	if got := firedCount.Load(); got != 0 {
		t.Fatalf("expected OnIdle not to have fired immediately, got %d", got)
	}

	time.Sleep(1500 * time.Millisecond)

	if got := firedCount.Load(); got != 1 {
		t.Errorf("expected OnIdle to fire once within 1.5s, got %d", got)
	}

	// Wait another second to ensure it does not fire a second time without activity
	time.Sleep(1200 * time.Millisecond)

	if got := firedCount.Load(); got != 1 {
		t.Errorf("expected OnIdle NOT to fire a second time without new activity, got %d", got)
	}
}

func autoMillis(msg string) int {
	return Auto().millis(msg)
}

func TestRenderCSS_NonEmpty(t *testing.T) {
	p := &Platform{}
	css := p.RenderCSS().String()
	if css == "" {
		t.Fatal("RenderCSS() returned empty string")
	}
	if !contains(css, ".pd") {
		t.Errorf("expected CSS to contain .pd")
	}
}

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

func NilCtx() Ctx {
	return &mockCtx{}
}

func contains(s, substr string) bool {
	// Simple manual contains to avoid strings package as per project rules
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
