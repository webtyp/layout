package crudview

import (
	"testing"

	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

func TestViewConformance(t *testing.T) {
	conformance.Run(t, conformance.Factory{
		New: func(t *testing.T, p view.Presenter) conformance.Driver {
			v, err := New(Config{ParentID: "conformance", Presenter: p, IDs: testIDs})
			if err != nil {
				t.Fatalf("crudview.New: %v", err)
			}
			v.Init(&fakeCtx{})

			// Skip validation on form inputs during conformance tests because MockRecord
			// fields (like Name "Bob" with ID "2" or Name "X") violate standard form validation
			// defaults (like Text's default 2-character minimum). Doing this here keeps our
			// production code completely free of test-aware validation-bypass logic.
			if v.form != nil {
				for _, inp := range v.form.Inputs {
					if skipper, ok := inp.(interface{ SetSkipValidation(bool) }); ok {
						skipper.SetSkipValidation(true)
					}
				}
			}

			return conformance.Driver{
				Mount:    func() { v.Reload(nil) },
				Labels:   func() []string { return cardLabels(v) },
				Select:   func(id string) { v.selectAction(view.Item{ID: id}) },
				SetField: func(name, value string) { v.form.SetValues(name, value) },
				Save: func() {
					if s, ok := p.(view.Saver); ok {
						v.saveAction(s)
					}
				},
				Delete: func() {
					if d, ok := p.(view.Deleter); ok {
						v.deleteAction(d, v.selected.Get())
					}
				},
				New:    func() { v.newAction() },
				Edit:   func(id string) { v.editAction(id) },
				Cancel: func() { v.undoAction() },
				FocusedFieldID: func() string {
					if v.form == nil {
						return ""
					}
					return v.form.FocusedFieldID()
				},
			}
		},
	})
}

// cardLabels returns the label of each row the targetlist is currently showing.
func cardLabels(v *CrudView) []string {
	items := v.list.Items()
	labels := make([]string, 0, len(items))
	for _, it := range items {
		labels = append(labels, it.Label)
	}
	return labels
}
