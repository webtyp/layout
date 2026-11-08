//go:build wasm

package tests

import (
	"strings"
	"syscall/js"
	"testing"
	"time"

	layoutapp "webtyp.com/layout/app"
	"webtyp.com/view"
)

func TestDemo_InvalidShowsReason(t *testing.T) {
	go layoutapp.Demo(func() (view.Demo, error) {
		return view.Demo{}, nil
	})

	time.Sleep(50 * time.Millisecond)

	body := js.Global().Get("document").Get("body").Get("textContent").String()
	if !strings.Contains(body, view.ErrDemoNoLabel.Error()) {
		t.Fatalf("expected body to contain %q, got %q", view.ErrDemoNoLabel.Error(), body)
	}
}
