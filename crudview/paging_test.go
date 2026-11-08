package crudview

import (
	"strings"
	"testing"

	"webtyp.com/crud"
	"webtyp.com/view"
)

type pagedLister struct {
	rows     []*Device
	page     view.Page
	calls    int
	lastArgs crud.ReadArgs
	records  *view.Records
}

func (l *pagedLister) Records() *view.Records {
	if l.records == nil {
		l.records = view.NewRecords()
	}
	return l.records
}

func (l *pagedLister) Page() view.Page {
	return l.page
}

func (l *pagedLister) List(args crud.ReadArgs, done func([]view.Record, error)) {
	l.calls++
	l.lastArgs = args
	var filtered []view.Record
	for _, d := range l.rows {
		if args.After != "" && d.Id <= args.After {
			continue
		}
		if args.Search != "" && !strings.Contains(strings.ToLower(d.Name), strings.ToLower(args.Search)) {
			continue
		}
		filtered = append(filtered, d)
		if args.Limit > 0 && len(filtered) >= int(args.Limit) {
			break
		}
	}
	done(filtered, nil)
}

func TestCrudView_SearchAsksTheServer(t *testing.T) {
	devices := []*Device{
		{Id: "1", Name: "maria a"},
		{Id: "2", Name: "maria b"},
		{Id: "3", Name: "maria c"},
		{Id: "4", Name: "maria d"},
		{Id: "5", Name: "maria e"},
		{Id: "6", Name: "pedro"},
	}
	lister := &pagedLister{
		rows: devices,
		page: view.Page{First: 0, Size: 2, More: true},
	}
	p := view.New(lister, &Device{})
	filter := &fakeFilter{}
	v := &CrudView{Presenter: p, Filter: filter}
	v.Init(&fakeCtx{})
	v.Activate()

	v.Reload(func(err error) {
		if err != nil {
			t.Fatalf("Reload failed: %v", err)
		}
	})

	if len(cardLabels(v)) != 0 {
		t.Fatalf("expected list empty, got %d rows", len(cardLabels(v)))
	}
	if lister.calls != 0 {
		t.Fatalf("expected 0 calls on Reload with First: 0, got %d", lister.calls)
	}

	filter.sink("maria")

	if lister.lastArgs.Search != "maria" {
		t.Errorf("expected Search == 'maria', got %q", lister.lastArgs.Search)
	}
	if lister.lastArgs.Limit != 2 {
		t.Errorf("expected Limit == 2, got %d", lister.lastArgs.Limit)
	}
	if len(cardLabels(v)) != 2 {
		t.Errorf("expected 2 rows, got %d", len(cardLabels(v)))
	}
}

func TestCrudView_ShowMoreAppendsTheNextPage(t *testing.T) {
	devices := []*Device{
		{Id: "1", Name: "maria a"},
		{Id: "2", Name: "maria b"},
		{Id: "3", Name: "maria c"},
		{Id: "4", Name: "maria d"},
		{Id: "5", Name: "maria e"},
	}
	lister := &pagedLister{
		rows: devices,
		page: view.Page{First: 0, Size: 2, More: true},
	}
	p := view.New(lister, &Device{})
	filter := &fakeFilter{}
	v := &CrudView{Presenter: p, Filter: filter}
	v.Init(&fakeCtx{})
	v.Activate()

	filter.sink("maria")

	if !v.hasMore.Get() {
		t.Fatal("expected hasMore to be true after search with more rows")
	}

	v.Presenter.More(func(err error) {
		if err != nil {
			t.Fatalf("More: %v", err)
		}
		v.paint()
	})

	if lister.lastArgs.After != "2" {
		t.Errorf("expected After == '2', got %q", lister.lastArgs.After)
	}
	if len(cardLabels(v)) != 4 {
		t.Errorf("expected 4 rows after first More, got %d", len(cardLabels(v)))
	}

	v.Presenter.More(func(err error) {
		if err != nil {
			t.Fatalf("More: %v", err)
		}
		v.paint()
	})

	if len(cardLabels(v)) != 5 {
		t.Errorf("expected 5 rows after second More, got %d", len(cardLabels(v)))
	}
	if v.hasMore.Get() {
		t.Error("expected hasMore to be false when all rows are loaded")
	}
}

func TestCrudView_ReloadDoesNotSearchTwice(t *testing.T) {
	devices := []*Device{
		{Id: "1", Name: "Device 1"},
	}
	lister := &pagedLister{
		rows: devices,
		page: view.Page{First: view.All},
	}
	p := view.New(lister, &Device{})
	v := &CrudView{Presenter: p}
	v.Init(&fakeCtx{})

	v.Reload(func(err error) {
		if err != nil {
			t.Fatalf("Reload: %v", err)
		}
	})

	if lister.calls != 1 {
		t.Errorf("expected exactly 1 List call on Reload, got %d", lister.calls)
	}
}
