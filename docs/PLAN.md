---
PLAN: "feat(crudview): OnAfterReload hook — consumer personalizes the list widget after every load"
TAG: v0.2.21
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 7064163444445900914
---

# PLAN — `crudview.Config.OnAfterReload` (Etapa G del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` — la etapa G necesita que
`reservation` rellene `targethour.FreeSlots` (huecos libres) tras CADA Reload.

> **Diseño cerrado con el dueño.** Firma: `OnAfterReload func(list ListView)` —
> un solo argumento, el list concreto que `crudview` acaba de pintar. Los items
> NO se pasan: ya son consultables (`list.Items()` / `Presenter.Items()`), pasar
> un segundo arg duplicaría una fuente de verdad y confundiría al consumidor
> ("¿uso el arg o list.Items()?"). El único dato que el consumidor no puede
> conseguir solo es el **widget concreto** (el factory `Config.List` lo crea
> crudview internamente), por eso es el argumento.

## Cambio (`crudview.go` + `crud.go`)

### `crud.go` — `Config`

```go
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
```

### `crudview.go` — `CrudView`

- Campo `OnAfterReload func(ListView)` en el struct (doc espejo del de Config).
- `New(cfg)`: `v.OnAfterReload = cfg.OnAfterReload`.
- `Reload()`:

```go
func (v *CrudView) Reload() error {
	if v.Presenter == nil {
		return nil
	}
	if err := v.Presenter.Reload(); err != nil {
		return err
	}
	v.filter()
	if v.OnAfterReload != nil && v.list != nil {
		v.OnAfterReload(v.list)
	}
	return nil
}
```

Es el ÚNICO punto de disparo. `filter()` ya hace `list.SetItems(...)` y los
otros `Reload()` del módulo pasan por acá (los `_ = v.Reload()` de save/delete/
selection). El hook ve `v.list` ya pintado con `Items()` actuales.

No se toca `rightpanel`.

## Tests (`crudview/`, `gotest`)

- `TestOnAfterReload_FiresWithList` — config con `List` factory que devuelve un
  `stubList` y `OnAfterReload` que captura el `ListView` recibido; `Reload()` →
  el capturado es el `*stubList`, y su `Items()` == los del presenter.
- `TestOnAfterReload_AfterSetItems` — el hook observa `list.Items()` y ve los
  items recién rellenados (no vacíos/viejos): prueba el "justo después".
- `TestOnAfterReload_NilNoop` — sin hook, `Reload()` idéntico a hoy (regresión:
  snapshot/HTML no cambia).
- `nil` por defecto en `Config` → cero cambio de comportamiento para los
  consumidores actuales.

## Criterios de aceptación

- `gotest ./...` verde en `layout`.
- `GOOS=js GOARCH=wasm go build ./...` OK.
- `docs/ARCHITECTURE.md` (crudview) menciona el hook y por qué pasa solo el
  list (la regla "minimal surface" del AGENTS).
- La demo `app-demo` (Etapa G) puede usarlo para `FreeSlots`.

## Fuera de alcance

- El cálculo de huecos libres (Etapa G, app-demo).
- Un hook "antes de SetItems" o "por señal" — OnAfterReload es la única vía
  nueva; si un consumidor necesita otro momento, se evalúa aparte.