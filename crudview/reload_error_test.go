//go:build !wasm

package crudview

import (
	"testing"

	. "webtyp.com/fmt"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// TestCrudView_Reload_ReportsErrorViaOnLoadError reproduces a real bug found
// while manually testing a downstream app (mjosefa-cms): a list screen
// showed "0 / 0" even though the server held real rows for the current
// tenant, and nothing anywhere — not the browser console error surface, not
// the UI — gave any indication that Reload had actually failed underneath.
//
// Root cause: CrudView.Init calls v.Reload(...) with an inline closure that
// only does `Log(err.Error())` on failure (console-only, invisible to an end
// user), and that is the ONLY call site — there is no exported hook a host
// app can wire to observe a load failure, unlike Save (OnSaved) and Delete
// (OnDeleted), which both already report their errors to a host-settable
// field. See docs/PLAN.md.
//
// This test drives Init() (which performs the first Reload internally, the
// same as production) against a Presenter whose List always fails, and
// expects the new OnLoadError hook — mirroring OnSaved/OnDeleted's existing
// shape — to receive that error. It cannot compile until OnLoadError exists
// on CrudView; adding it and wiring it inside Reload() is stage 1 of the
// plan.
func TestCrudView_Reload_ReportsErrorViaOnLoadError(t *testing.T) {
	wantErr := Err("presenter list failed")
	fb := &conformance.FakeLister{Err: wantErr}
	p := view.New(fb, &Device{})

	var gotErr error
	v := &CrudView{
		Title:     "Test CRUD",
		Presenter: p,
		OnLoadError: func(err error) {
			gotErr = err
		},
	}
	v.Init(&mockCtx{})

	if gotErr == nil {
		t.Fatal("OnLoadError was never called even though the Presenter's List failed on Init's " +
			"own Reload — a load failure is invisible today: Log() only reaches the browser " +
			"console, which no end user ever opens, and there is no host-observable signal at all")
	}
}
