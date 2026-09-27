//go:build wasm

package chatview

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/inboxlist"
	. "webtyp.com/dom"
)

func TestChatView_WASM_Mount(t *testing.T) {
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "test-cv-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() {
		doc.Get("body").Call("removeChild", root)
	})

	src := &fakeSource{
		roomsData: []inboxlist.Row{
			{ID: "r1", Title: "General", Unread: 1},
		},
	}

	v, err := New(Config{Source: src, MaxBodyLength: 2000})
	if err != nil {
		t.Fatalf("failed to create ChatView: %v", err)
	}
	v.Init(NilCtx())

	if err := Render("test-cv-root", v); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	cvs := doc.Call("querySelectorAll", ".chatview")
	if cvs.Get("length").Int() != 1 {
		t.Fatalf("expected 1 .chatview element, got %d", cvs.Get("length").Int())
	}
}
