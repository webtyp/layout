package crudview

import (
	"errors"
	"strings"
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// buttonTexts returns the text of every <button> in el's markup, in order.
func buttonTexts(el interface{ String() string }) []string {
	var out []string
	for _, chunk := range strings.Split(el.String(), "</button>") {
		open := strings.LastIndex(chunk, "<button")
		if open < 0 {
			continue
		}
		body := chunk[open:]
		if gt := strings.Index(body, ">"); gt >= 0 {
			out = append(out, body[gt+1:])
		}
	}
	return out
}

// countingLister counts List calls, so a test can prove a reload did (not) happen.
type countingLister struct {
	*conformance.FakeLister
	lists int
}

func (c *countingLister) List(done func([]model.Model, error)) {
	c.lists++
	c.FakeLister.List(done)
}

func newActionView(t *testing.T, fl *conformance.FakeLister) (*CrudView, *countingLister) {
	t.Helper()
	cl := &countingLister{FakeLister: fl}
	v, err := New(Config{ParentID: "actions", Presenter: view.New(cl, &Device{}), IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&fakeCtx{})
	return v, cl
}

func TestActions_BandRendersLabelsInOrder(t *testing.T) {
	fl := fakeListBackend()
	fl.ActionList = []view.Action{{Op: "apply", Label: "Apply"}, {Op: "resend", Label: "Resend"}}
	v, _ := newActionView(t, fl)

	html := v.Render().String()
	for _, want := range []string{"name='cv-action-apply'", "name='cv-action-resend'", string(clsActions)} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %q in markup", want)
		}
	}
	band, ok := v.renderActions()
	if !ok {
		t.Fatal("renderActions: no band for a presenter with actions")
	}
	if got := buttonTexts(band); len(got) != 2 || got[0] != "Apply" || got[1] != "Resend" {
		t.Errorf("button texts = %v, want [Apply Resend]", got)
	}
}

func TestActions_NoBandWithoutActions(t *testing.T) {
	v, _ := newActionView(t, fakeListBackend())
	html := v.Render().String()
	if strings.Contains(html, actionNamePrefix) || strings.Contains(html, string(clsActions)) {
		t.Errorf("a presenter without actions must render no actions band:\n%s", html)
	}
}

// The action runs over the PRESENTER's items, not the filtered list: a search
// term that hides every row must not disable it.
func TestActions_EnabledByPresenterItemsNotByFilter(t *testing.T) {
	fl := fakeListBackend()
	fl.ActionList = []view.Action{{Op: "apply", Label: "Apply"}}
	v, _ := newActionView(t, fl)

	if v.actionEnabled("apply") {
		t.Fatal("enabled before any items were loaded")
	}
	v.Reload(nil)
	if !v.actionEnabled("apply") {
		t.Fatal("disabled after items loaded")
	}
	v.search.Set("no-row-matches-this")
	v.filter()
	if v.hasRows.Get() {
		t.Fatal("test premise: the filter should hide every row")
	}
	if !v.actionEnabled("apply") {
		t.Error("a filter that hides every row disabled the action")
	}
}

func TestActions_FailureReportsOnceAndDoesNotReload(t *testing.T) {
	fl := fakeListBackend()
	fl.ActionList = []view.Action{{Op: "apply", Label: "Apply"}}
	v, cl := newActionView(t, fl)
	v.Reload(nil)

	var calls []error
	v.OnAction = func(op string, err error) {
		if op != "apply" {
			t.Errorf("OnAction op = %q, want apply", op)
		}
		calls = append(calls, err)
	}
	listsBefore := cl.lists
	fl.Err = errors.New("router unreachable")
	v.actionRequest(fl.ActionList[0])

	if len(calls) != 1 || calls[0] == nil {
		t.Fatalf("OnAction calls = %v, want exactly one error", calls)
	}
	if cl.lists != listsBefore {
		t.Errorf("a failed action reloaded the list (%d → %d List calls)", listsBefore, cl.lists)
	}
	if v.actionRunning.Get() != "" {
		t.Error("actionRunning not cleared after the failure")
	}
}

func TestActions_SuccessReloadsAndReportsOnce(t *testing.T) {
	fl := fakeListBackend()
	fl.ActionList = []view.Action{{Op: "apply", Label: "Apply", Confirm: "Apply these changes?"}}
	v, cl := newActionView(t, fl)
	v.Reload(nil)

	var calls int
	v.OnAction = func(op string, err error) {
		calls++
		if err != nil {
			t.Errorf("OnAction err = %v", err)
		}
	}
	listsBefore := cl.lists
	v.actionRequest(fl.ActionList[0]) // opens the confirmation only
	if len(fl.Ran) != 0 || calls != 0 {
		t.Fatalf("ran before confirmation: Ran=%v calls=%d", fl.Ran, calls)
	}
	if got := v.confirmLabel.Get(); got != "Apply" {
		t.Errorf("confirm button label = %q, want the action's own label", got)
	}
	v.confirmActionRun()
	if len(fl.Ran) != 1 || fl.Ran[0] != "apply" || calls != 1 {
		t.Fatalf("after confirm: Ran=%v calls=%d, want [apply] and 1", fl.Ran, calls)
	}
	if cl.lists != listsBefore+1 {
		t.Errorf("success must reload once: %d → %d List calls", listsBefore, cl.lists)
	}
	v.confirmActionRun() // nothing pending any more
	if len(fl.Ran) != 1 {
		t.Errorf("a second confirm without a pending action ran again: %v", fl.Ran)
	}
}
