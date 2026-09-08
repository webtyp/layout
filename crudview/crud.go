package crudview

import (
	"webtyp.com/components/searchbar"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/model"
	"webtyp.com/view"
)

// Config is what a renderer needs to draw a view.Presenter — nothing about ops, transport, or
// codec: that is ALL inside the Presenter already. The module builds the Presenter via
// view.New(...) (importing view+model+router, never layout) and hands it here.
type Config struct {
	// ParentID is the DOM id the form mounts under.
	ParentID string
	// Presenter is built by the module via view.New(...). Required.
	Presenter view.Presenter
	// IDs supplies new ids for hidden text-PK auto-assignment on submit (the
	// form's composition-root injection, see form.New). Required — New fails
	// if nil. Pass unixid.NewUnixID() at your composition root.
	IDs model.IDGenerator

	// Filter is the control that narrows the list. Optional: nil renders no
	// controls band. When nil, New installs a searchbar.SearchBar carrying the
	// presenter's placeholder — the ergonomic default, not a decision imposed:
	// pass any widget.Filterable to replace it.
	Filter dom.Component

	// Context is a control rendered in the title band (below the H1, above
	// the article), for a context/scope selector: e.g. "which professional /
	// which área". Optional: nil paints no band. It is the mirror of the
	// legacy Pa100T professional dropdown that re-scoped the whole screen.
	//
	// Unlike Filter, Context NEVER wires itself to the list filter even if it
	// satisfies widget.Filterable: changing the context re-scopes the data
	// (the composing module calls its presenter + CrudView.Reload), it does
	// not filter a search term. crudview only renders it.
	Context dom.Component

	// List builds the row-rendering widget. Optional: nil installs a
	// targetlist.TargetList factory — the ergonomic default, not a decision
	// imposed: pass a targetdate.TargetDate (or anything satisfying
	// ListView) factory instead when the data wants a leading date/time
	// badge (view.Item's LeadTop/Main/Bottom) rather than a plain label.
	List func(selected *dom.SignalString, onSelect func(view.Item)) ListView

	// OnAfterReload se invoca al final de Reload(), justo después de que la lista
	// se re-llena con los items del presenter (list.SetItems). Recibe el list
	// concreto ya pintado, para que el consumidor acomode detalles que el widget no
	// puede derivar — p. ej. type-assert a *targetdate.TargetDate / *targethour.
	// TargetHour y setear sus campos (FreeSlots). nil = hook ausente.
	//
	// Un solo argumento a propósito: items[] no viajan (list.Items() /
	// Presenter.Items() ya los dan); lo único que NO se puede conseguir de otro
	// lado es el list concrete que Config.List construyó.
	OnAfterReload func(list ListView)
}

// New builds the renderer around an already-constructed Presenter. It generates the form from
// Presenter.Record().Schema(), and wires save/delete to sync the form INTO Record() before
// calling Presenter.Save/Delete — crudview never talks to a Caller directly anymore.
func New(cfg Config) (*CrudView, error) {
	if cfg.Presenter == nil {
		return nil, fmt.Errf("crudview.New: Presenter is required")
	}
	if cfg.IDs == nil {
		return nil, fmt.Errf("crudview.New: IDs is required (e.g. unixid.NewUnixID())")
	}

	f, err := form.New(cfg.ParentID, cfg.Presenter.Record(), cfg.IDs)
	if err != nil {
		return nil, err // a record with no widgets fails HERE, loudly
	}
	f.HideSubmit() // there is no Save button — auto-save (OnFieldChange) replaces it

	// The default filter keeps every existing consumer compiling: a host that
	// does not care which control it gets still gets a working search bar with
	// the presenter's placeholder. The demo injects its own — see platformd.
	filter := cfg.Filter
	if filter == nil {
		filter = &searchbar.SearchBar{Placeholder: cfg.Presenter.SearchPlaceholder()}
	}

	v := &CrudView{
		Title:     cfg.Presenter.Title(),
		Form:      f,
		form:      f,
		Presenter: cfg.Presenter,
		Filter:    filter,
		Context:   cfg.Context,
		// List is passed through as-is, nil included: Init resolves the
		// same targetlist.TargetList default that a nil List would get
		// here, so there is exactly one place that decision lives.
		List:          cfg.List,
		OnAfterReload: cfg.OnAfterReload,
	}

	// Auto-save: every field commit (blur/change) persists immediately — see
	// docs/ROADMAP.md "Save: auto-save". Only wired when the presenter can save;
	// autoSaveAction is a no-op otherwise.
	f.OnFieldChange(func() { v.autoSaveAction() })

	return v, nil
}
