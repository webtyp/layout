package crudview

import (
	"testing"
	"webtyp.com/model"
	"webtyp.com/view"
)

type sharedBackend struct {
	records *view.Records
}

func (b *sharedBackend) Records() *view.Records { return b.records }
func (b *sharedBackend) List(done func([]view.Record, error)) { done(nil, nil) }
func (b *sharedBackend) Save(recs []model.Model, done func(error)) {
	for _, rec := range recs {
		if r, ok := rec.(view.Record); ok {
			b.records.Put(r)
		}
	}
	done(nil)
}

func TestCrudView_OtherScreenSaveShowsHere(t *testing.T) {
	records := view.NewRecords()
	b1 := &sharedBackend{records: records}
	b2 := &sharedBackend{records: records}

	p1 := view.New(b1, &Device{})
	p2 := view.New(b2, &Device{})

	v1, _ := New(Config{ParentID: "v1", Presenter: p1, IDs: testIDs})
	v2, _ := New(Config{ParentID: "v2", Presenter: p2, IDs: testIDs})

	v1.Init(&fakeCtx{})
	v1.Activate()
	v2.Init(&fakeCtx{})
	v2.Activate()

	if len(v2.list.Items()) != 0 {
		t.Fatalf("expected 0 items, got %d", len(v2.list.Items()))
	}

	p1.(view.Saver).Save([]model.Model{&Device{Id: "1", Name: "One"}}, func(error) {})

	items := v2.list.Items()
	if len(items) != 1 || items[0].Label != "One" {
		t.Fatalf("expected 1 item 'One', got %v", items)
	}
}
