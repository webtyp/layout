package crudview

import (
	"strings"
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// auditRowModel is an output-only record: base kinds, no form widget — the
// shape of a log, a plan or a "who is connected" row.
var auditRowModel = model.Definition{
	Name: "audit_row",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "what", Type: model.Text()},
	},
}

type AuditRow struct{ Id, What string }

func (r *AuditRow) ModelName() string     { return "audit_row" }
func (r *AuditRow) Schema() []model.Field { return auditRowModel.Fields }
func (r *AuditRow) Pointers() []any       { return []any{&r.Id, &r.What} }
func (r *AuditRow) IsNil() bool           { return r == nil }
func (r *AuditRow) EncodeFields(w model.FieldWriter) {
	w.String("id", r.Id)
	w.String("what", r.What)
}
func (r *AuditRow) DecodeFields(rd model.FieldReader) {
	r.Id, _ = rd.String("id")
	r.What, _ = rd.String("what")
}
func (r *AuditRow) Item() view.Item { return view.Item{ID: r.Id, Label: r.What} }

// listOnly is a Lister with no write capability.
type listOnly struct{ rows []model.Model }

func (l *listOnly) List(done func([]model.Model, error)) { done(l.rows, nil) }

// A read-only presenter over a widget-less record is a standalone list.
func TestReadOnlyWithoutWidgets_IsAStandaloneList(t *testing.T) {
	p := view.New(&listOnly{rows: []model.Model{&AuditRow{Id: "1", What: "applied"}}}, &AuditRow{})
	v, err := New(Config{ParentID: "audit", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if v.Form != nil || v.form != nil {
		t.Fatal("a read-only widget-less view must have no form")
	}
	v.Init(&fakeCtx{})
	v.Reload(nil)
	v.selectAction(view.Item{ID: "1"}) // must not touch a form
	if html := v.Render().String(); !strings.Contains(html, string(clsListaBox)) {
		t.Errorf("list box not rendered:\n%s", html)
	}
	if got := cardLabels(v); len(got) != 1 || got[0] != "applied" {
		t.Errorf("list rows = %v, want [applied]", got)
	}
}

// An editable presenter over a widget-less record still fails loudly.
func TestEditableWithoutWidgets_FailsLoudly(t *testing.T) {
	p := view.New(&conformance.FakeLister{}, &AuditRow{})
	if _, err := New(Config{ParentID: "audit", Presenter: p, IDs: testIDs}); err == nil {
		t.Fatal("an editable presenter over a record without widgets must fail in New")
	}
}
