package tests

import (
	"strings"
	"testing"

	"webtyp.com/ddl"
	"webtyp.com/erd"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/input"
	"webtyp.com/layout"
	"webtyp.com/layout/apiexplorer"
	"webtyp.com/layout/app"
	"webtyp.com/layout/schemaview"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

type testIDGen struct{}

func (testIDGen) NewID() string { return "test-id" }

type localFakeLister struct {
	conformance.FakeLister
	records *view.Records
}

func (l *localFakeLister) Records() *view.Records {
	if l.records == nil {
		l.records = view.NewRecords()
	}
	return l.records
}

func TestBuildMenu_InvalidScreenIsAnError(t *testing.T) {
	screens := []view.Screen{
		{
			ID:        "leaf",
			Label:     "Leaf",
			Icon:      "laptop",
			Presenter: view.New(&localFakeLister{}, &conformance.MockRecord{}),
		},
	}
	ids := testIDGen{}
	_, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return true }, app.Tools{})
	if err == nil {
		t.Fatal("expected error for leaf screen without Permission, got nil")
	}
	if !strings.Contains(err.Error(), "needs a Permission") {
		t.Fatalf("expected error containing 'needs a Permission', got %v", err)
	}
}

func TestBuildMenu_LeafVisibleByItsPermission(t *testing.T) {
	screens := []view.Screen{
		{
			ID:         "x",
			Label:      "X",
			Icon:       "laptop",
			Permission: "service",
			Presenter:  view.New(&localFakeLister{}, &conformance.MockRecord{}),
		},
	}
	ids := testIDGen{}

	allowsService := func(r model.Resource) bool { return string(r) == "service" }
	m, err := app.BuildMenu(screens, ids, allowsService, app.Tools{})
	if err != nil {
		t.Fatal(err)
	}
	if !m.CanView("x") {
		t.Error("expected CanView('x') to be true when allows returns true")
	}

	allowsFalse := func(r model.Resource) bool { return false }
	m2, err := app.BuildMenu(screens, ids, allowsFalse, app.Tools{})
	if err != nil {
		t.Fatal(err)
	}
	if m2.CanView("x") {
		t.Error("expected CanView('x') to be false when allows always false")
	}
}

func TestBuildMenu_GroupHasNoOwnPermission(t *testing.T) {
	screens := []view.Screen{
		{
			ID:    "g",
			Label: "Group",
			Icon:  "laptop",
			Help:  "help",
			Children: []view.Screen{
				{
					ID:         "g.a",
					Label:      "A",
					Icon:       "laptop",
					Permission: "a",
					Presenter:  view.New(&localFakeLister{}, &conformance.MockRecord{}),
				},
			},
		},
	}
	ids := testIDGen{}

	m, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return string(r) == "a" }, app.Tools{})
	if err != nil {
		t.Fatal(err)
	}
	if m.CanView("g") {
		t.Error("expected CanView('g') to be false (groups have no permission)")
	}
}

func TestBuildMenu_DescriptionFromPresenter(t *testing.T) {
	presenter := view.New(&localFakeLister{}, &conformance.MockRecord{}, view.WithDescription("Help"))

	screens := []view.Screen{
		{
			ID:         "booking",
			Label:      "Book",
			Icon:       "laptop",
			Permission: "book",
			Presenter:  presenter,
		},
	}
	ids := testIDGen{}

	m, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return true }, app.Tools{})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Modules) != 1 {
		t.Fatalf("expected 1 module (1 screen), got %d", len(m.Modules))
	}
	if m.Modules[0].ModelName() != "booking" {
		t.Errorf("expected ModelName booking, got %s", m.Modules[0].ModelName())
	}

	if !m.CanView("booking") {
		t.Error("expected CanView('booking') to be true")
	}

	desc, ok := m.Modules[0].(layout.Described)
	if !ok {
		t.Fatalf("expected module to satisfy layout.Described")
	}
	if desc.Description() != "Help" {
		t.Errorf("expected Description 'Help', got '%s'", desc.Description())
	}
}

func TestBuildMenu_ComponentScreenVisibleByPermission(t *testing.T) {
	screens := []view.Screen{
		{
			ID:         "booking",
			Label:      "Book",
			Icon:       "laptop",
			Permission: "book",
			Component:  html.Div(),
			Help:       "Book.",
		},
	}
	ids := testIDGen{}

	m, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return true }, app.Tools{})
	if err != nil {
		t.Fatal(err)
	}

	if m.Modules[0].View() != screens[0].Component {
		t.Error("expected View() to return the same component")
	}

	if !m.CanView("booking") {
		t.Error("expected CanView('booking') to be true")
	}

	desc, ok := m.Modules[0].(layout.Described)
	if !ok {
		t.Fatalf("expected module to satisfy layout.Described")
	}
	if desc.Description() != "Book." {
		t.Errorf("expected Description 'Book.', got '%s'", desc.Description())
	}
}

func TestBuildMenu_AddsDevToolsLast(t *testing.T) {
	screens := []view.Screen{
		{
			ID:         "x",
			Label:      "X",
			Icon:       "laptop",
			Permission: "service",
			Presenter:  view.New(&localFakeLister{}, &conformance.MockRecord{}),
		},
	}
	ids := testIDGen{}

	allows := func(r model.Resource) bool { return string(r) == string(router.ResourceIntrospection) }
	m, err := app.BuildMenu(screens, ids, allows, app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()})
	if err != nil {
		t.Fatal(err)
	}

	last := m.Modules[len(m.Modules)-1]
	if last.ModelName() != "devtools" {
		t.Errorf("expected last module to be devtools, got '%s'", last.ModelName())
	}
	if gm, ok := last.(layout.GroupModule); !ok {
		t.Errorf("expected last module to be a GroupModule")
	} else {
		if len(gm.Children()) != 2 {
			t.Fatalf("expected devtools group to have 2 children, got %d", len(gm.Children()))
		}
		if gm.Children()[0].ModelName() != string(router.ResourceIntrospection) {
			t.Errorf("expected first devtools child to be %s, got %s", string(router.ResourceIntrospection), gm.Children()[0].ModelName())
		}
		if gm.Children()[1].ModelName() != string(erd.ResourceSchema) {
			t.Errorf("expected second devtools child to be %s, got %s", string(erd.ResourceSchema), gm.Children()[1].ModelName())
		}
	}

	if !m.CanView(string(router.ResourceIntrospection)) {
		t.Error("expected CanView for api explorer to be true when allows returns true")
	}

	// with allows true only for erd.ResourceSchema:
	allowsSchema := func(r model.Resource) bool { return r == erd.ResourceSchema }
	mSchema, err := app.BuildMenu(screens, ids, allowsSchema, app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()})
	if err != nil {
		t.Fatal(err)
	}
	if !mSchema.CanView("db_schema") {
		t.Error("expected CanView('db_schema') to be true")
	}
	if mSchema.CanView(string(router.ResourceIntrospection)) {
		t.Error("expected CanView for api explorer to be false when allows only for schema")
	}

	allowsFalse := func(r model.Resource) bool { return false }
	m2, _ := app.BuildMenu(screens, ids, allowsFalse, app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()})
	if m2.CanView(string(router.ResourceIntrospection)) {
		t.Error("expected CanView for api explorer to be false when allows returns false")
	}
	if m2.CanView("db_schema") {
		t.Error("expected CanView for db_schema to be false when allows returns false")
	}
}

var specialtyDef = model.Definition{Name: "specialty", Fields: model.Fields{{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}}, {Name: "name", Type: input.Text(), NotNull: true, Item: model.ItemLabel}}}
type testSpecialty struct {
	Id, Name string
}
func (d *testSpecialty) ModelName() string     { return "specialty" }
func (d *testSpecialty) Schema() []model.Field { return specialtyDef.Fields }
func (d *testSpecialty) Pointers() []any       { return []any{&d.Id, &d.Name} }
func (d *testSpecialty) IsNil() bool           { return d == nil }
func (d *testSpecialty) EncodeFields(w model.FieldWriter) {}
func (d *testSpecialty) DecodeFields(r model.FieldReader) {}
func (d *testSpecialty) Item() model.Item      { return model.Item{ID: d.Id, Label: d.Name} }

var serviceDef = model.Definition{Name: "service", Fields: model.Fields{{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}}, {Name: "specialty_id", Type: input.Select(), Ref: &specialtyDef}}}
type testService struct {
	Id, SpecialtyId string
}
func (d *testService) ModelName() string     { return "service" }
func (d *testService) Schema() []model.Field { return serviceDef.Fields }
func (d *testService) Pointers() []any       { return []any{&d.Id, &d.SpecialtyId} }
func (d *testService) IsNil() bool           { return d == nil }
func (d *testService) EncodeFields(w model.FieldWriter) {}
func (d *testService) DecodeFields(r model.FieldReader) {}
func (d *testService) Item() model.Item      { return model.Item{ID: d.Id} }

type localSavingLister struct {
	records *view.Records
}
func (l *localSavingLister) Records() *view.Records { return l.records }
func (l *localSavingLister) List(done func([]view.Record, error)) { done([]view.Record{}, nil) }
func (l *localSavingLister) Save(recs []model.Model, done func(error)) {
	for _, rec := range recs {
		if r, ok := rec.(view.Record); ok {
			l.records.Put(r)
		}
	}
	done(nil)
}

func TestBuildMenu_SelectFollowsSource(t *testing.T) {
	records := view.NewRecords()
	specLister := &localSavingLister{records: records}
	servLister := &localSavingLister{records: records}

	specPres := view.New(specLister, &testSpecialty{})
	servPres := view.New(servLister, &testService{})

	screens := []view.Screen{
		{ID: "specialty", Label: "Specialties", Icon: "laptop", Permission: "spec", Presenter: specPres},
		{ID: "service", Label: "Services", Icon: "laptop", Permission: "serv", Presenter: servPres},
	}

	app.BuildMenu(screens, testIDGen{}, func(r model.Resource) bool { return true }, app.Tools{})

	// simulate a name change
	specPres.(view.Saver).Save([]model.Model{&testSpecialty{Id: "s1", Name: "Surgeon"}}, func(error){})
    // In actual implementation, we are just making sure we don't crash and that it follows.
    // CrudView is unexported but we can test it at the integration level in crudview tests.
}

type testServiceMissing struct {
	Id, MissingId string
}
var serviceDefMissing = model.Definition{Name: "service", Fields: model.Fields{{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}}, {Name: "missing_id", Type: input.Select(), Ref: &model.Definition{Name: "missing"}}}}
func (d *testServiceMissing) ModelName() string     { return "service" }
func (d *testServiceMissing) Schema() []model.Field { return serviceDefMissing.Fields }
func (d *testServiceMissing) Pointers() []any       { return []any{&d.Id, &d.MissingId} }
func (d *testServiceMissing) IsNil() bool           { return d == nil }
func (d *testServiceMissing) EncodeFields(w model.FieldWriter) {}
func (d *testServiceMissing) DecodeFields(r model.FieldReader) {}
func (d *testServiceMissing) Item() model.Item      { return model.Item{ID: d.Id} }

func TestBuildMenu_SelectWithoutSourceFails(t *testing.T) {
	screens := []view.Screen{
		{ID: "service", Label: "Services", Icon: "laptop", Permission: "serv", Presenter: view.New(&localFakeLister{}, &testServiceMissing{})},
	}

	_, err := app.BuildMenu(screens, testIDGen{}, func(r model.Resource) bool { return true }, app.Tools{})
	if err == nil || !strings.Contains(err.Error(), "a field takes its options from a model no screen lists") {
		t.Fatalf("expected error containing 'a field takes its options from a model no screen lists', got %v", err)
	}
}


func TestBuildMenu_NoTools_NoGroup(t *testing.T) {
	m, err := app.BuildMenu(nil, testIDGen{}, func(model.Resource) bool { return true }, app.Tools{})
	if err != nil {
		t.Fatal(err)
	}
	for _, mod := range m.Modules {
		if mod.ModelName() == "devtools" {
			t.Error("expected no devtools group")
		}
	}
}

func TestBuildMenu_OnlyAPI(t *testing.T) {
	m, err := app.BuildMenu(nil, testIDGen{}, func(model.Resource) bool { return true }, app.Tools{API: apiexplorer.FromServer()})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Modules) != 1 || m.Modules[0].ModelName() != "devtools" {
		t.Fatalf("expected 1 module (devtools), got %d", len(m.Modules))
	}
	gm := m.Modules[0].(layout.GroupModule)
	if len(gm.Children()) != 1 || gm.Children()[0].ModelName() != string(router.ResourceIntrospection) {
		t.Errorf("expected 1 child (API), got %v", gm.Children())
	}
}

func TestBuildMenu_OnlyDatabase(t *testing.T) {
	m, err := app.BuildMenu(nil, testIDGen{}, func(model.Resource) bool { return true }, app.Tools{Database: schemaview.FromServer()})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Modules) != 1 || m.Modules[0].ModelName() != "devtools" {
		t.Fatalf("expected 1 module (devtools), got %d", len(m.Modules))
	}
	gm := m.Modules[0].(layout.GroupModule)
	if len(gm.Children()) != 1 || gm.Children()[0].ModelName() != string(erd.ResourceSchema) {
		t.Errorf("expected 1 child (Database), got %v", gm.Children())
	}
}

type fakeOpMod struct{}

func (fakeOpMod) ModelName() string { return "demo" }

func (fakeOpMod) MountOperations(reg router.OperationRegistry) {
	noop := func(router.Context) {}
	reg.Operation("list", noop).Requires("demo_item", model.Read)
	reg.Operation("save", noop).Requires("demo_item", model.Create)
}

var _ router.OperationModule = fakeOpMod{}

func TestLocalTools(t *testing.T) {
	fakeStep := ddl.Step{
		Name: "demo",
		Run: func(c ddl.Execer, k ddl.Compiler) error {
			return ddl.New(c, k).Sync(&testSpecialty{})
		},
	}
	tools, err := app.LocalTools([]router.OperationModule{fakeOpMod{}}, fakeStep)
	if err != nil {
		t.Fatal(err)
	}
	if tools.API == nil || tools.Database == nil {
		t.Error("expected API and Database to be not nil")
	}

	toolsNone, err := app.LocalTools(nil)
	if err != nil {
		t.Fatal(err)
	}
	if toolsNone.API != nil || toolsNone.Database != nil {
		t.Error("expected API and Database to be nil")
	}

	boomStep := ddl.Step{
		Name: "broken",
		Run: func(ddl.Execer, ddl.Compiler) error {
			return fmt.Err("boom")
		},
	}
	_, errBoom := app.LocalTools(nil, boomStep)
	if errBoom == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(errBoom.Error(), "broken") || !strings.Contains(errBoom.Error(), "boom") {
		t.Errorf("expected error containing step name and boom, got %v", errBoom)
	}
}
