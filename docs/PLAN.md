---
PLAN: "feat(crudview): Config.NewRecord — a new draft can start from a seeded record"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 13786302461332605150
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — un borrador nuevo puede arrancar con datos

## El problema, medido en una app real

`veltylabs/mjosefa-cms` tiene una pantalla de ficha clínica cuyo filtro es un
`selectsearch` con la agenda del día: el doctor elige al paciente de sus
reservas confirmadas. Al pulsar "+" para crear la ficha de esa atención,
`crudview` limpia el formulario y **el doctor tiene que escribir a mano el
`patient_id`, el nombre y el RUT del paciente que acaba de seleccionar**.

No es una preferencia de esa app: hoy **no existe forma** de que un host
siembre el borrador. `newAction` (crudview.go) hace:

```go
if v.form != nil {
    v.form.Reset()
    v.form.Focus()
}
...
if v.OnNew != nil {
    v.OnNew()          // notificación, sin argumentos, después del Reset
}
```

`OnNew` no recibe nada y no expone el formulario, así que no hay dónde
engancharse. El resultado es que una app que YA SABE el dato se lo pide igual
a la persona — y escribir a mano un RUT que el sistema tiene en pantalla es
exactamente donde aparecen los errores de tipeo en una ficha médica.

## La pieza ya está: `LoadValues`

`selectAction` puebla el formulario desde un registro con una sola línea:

```go
rec := v.Presenter.Select(it.ID)
if v.form != nil {
    _ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
}
```

Un borrador sembrado es ese mismo camino, con un registro que lo da el host en
vez del presenter. No hace falta mecanismo nuevo: hace falta **de dónde sacar
el registro**.

## Design gate (skill: api-design)

### 1. Prior art

- **React Hook Form** — `useForm({ defaultValues })`: el host entrega los
  valores de arranque; la librería no los inventa.
- **Django** — `MyForm(initial={...})`: un formulario no ligado arranca con
  los valores que el host declara.
- **Rails** — `form_with model: Post.new(author: current_user)`: el host
  construye el objeto ya sembrado y se lo pasa al formulario.

Los tres coinciden en lo mismo: **el host provee el objeto/valores semilla, la
capa de formulario solo los carga.** Este ecosistema es centrado en registros
(`Presenter.Record()`, `form.LoadValues(rec)`), así que la forma local de esa
misma idea es una fábrica de registro, no un mapa de valores sueltos.

### 2. La prueba del nombre novato

`NewRecord` — "el registro con el que arranca un borrador nuevo". Se lee en la
misma frase que `newAction`/`OnNew`, que es donde actúa, y usa la palabra que
esta librería ya usa para la cosa (`Presenter.Record()`). `DefaultValues` sería
vocabulario de React sobre un ecosistema que no habla de "values" sino de
registros; `Initial` no dice de qué.

### 3. El libro de complejidad

```
Conceptos que aprender              +1   (un campo opcional de Config)
Archivos que tocar para hacer X     +0   (el host ya construye su Config)
Líneas en el call site              +1   (NewRecord: func() model.Model {...})
Formas de hacer lo mismo             0   (hoy no existe NINGUNA)
```

### 4. Dónde vive

`webtyp/layout/crudview`: es quien posee `newAction` y el formulario. La app
consumidora no puede resolverlo — `v.form` no es alcanzable desde fuera, y
`OnNew` corre después del `Reset` sin recibir nada. Arreglarlo en el consumidor
exigiría reimplementar el botón "+", que es forkear la pantalla entera.

### 5. Qué borra este cambio

Borra el tipeo manual de datos que la app ya tiene en pantalla: en
`mjosefa-cms`, los tres campos (`patient_id`, `patient_name_snapshot`,
`patient_rut_snapshot`) que hoy el doctor copia a mano desde el selector de
agenda. No borra código de esta librería — es capacidad que no existía.

## Qué construir

Un campo opcional en `crudview.Config`:

```go
// NewRecord returns the record a NEW draft starts from — the "+" button's
// seed. nil (the default) starts from an empty form, which is what a screen
// with nothing to pre-fill wants.
//
// Called once per "+", and it must return a FRESH record every time: a
// shared instance would carry the previous draft's edits into the next one.
NewRecord func() model.Model
```

y en `newAction`, reemplazar el `Reset` incondicional:

```go
if v.form != nil {
    if v.NewRecord != nil {
        _ = v.form.LoadValues(v.NewRecord())
    } else {
        v.form.Reset()
    }
    v.form.Focus()
}
```

Decisiones ya tomadas — el ejecutor no elige:

- **`NewRecord` se copia de `Config` a `CrudView` en `New`**, igual que
  `Filter`, `Context`, `List` y `OnAfterReload`. Mismo patrón, ninguna
  excepción.
- **`nil` mantiene EXACTAMENTE el comportamiento de hoy** (`form.Reset()`).
  Ninguna pantalla existente cambia de conducta; este campo es aditivo.
- **`OnNew` NO cambia de firma.** Sigue siendo la notificación que ya es, y
  sigue corriendo al final de `newAction`. Cambiarla rompería a cada consumidor
  para resolver algo que este campo ya resuelve.
- **El orden se mantiene**: sembrar → `Focus()` → `composing.Set(true)` →
  `panel.ShowMain()` → `OnNew()`. El comentario que ya existe sobre por qué
  `Focus` va antes de `ShowMain` (el teclado de iOS Safari) sigue siendo
  válido y **no se toca**.
- **Si `NewRecord` devuelve `nil`**, `LoadValues(nil)` ya resetea — está
  documentado en el propio call site de `selectAction`. No agregar un guard
  extra: sería una segunda forma de decir lo mismo.

## Pasos

1. Agregar el campo a `crudview.Config` (con el comentario de arriba) y al
   struct `CrudView`, copiándolo en `New` junto a los demás.
2. Cambiar `newAction` exactamente como se muestra.
3. Tests en la suite de `crudview` (stdlib, sin navegador — es lógica, no DOM):
   - un `Config` con `NewRecord` devolviendo un registro con un campo poblado:
     tras `newAction`, el formulario tiene ese valor;
   - un `Config` SIN `NewRecord`: tras `newAction`, el formulario queda vacío
     (la conducta de hoy, que no debe cambiar);
   - `NewRecord` llamado dos veces devuelve instancias distintas y la edición
     de la primera no aparece en la segunda — la trampa que el comentario del
     campo advierte, fijada por un test y no solo por prosa.
4. `docs/` de crudview: si hay una tabla de `Config`, agregar la fila; si no,
   no crear documento nuevo.

## Criterios de aceptación

- `gotest` en verde (nunca `go test`).
- Una pantalla sin `NewRecord` se comporta byte por byte como antes.
- `OnNew` conserva su firma `func()`.
- Ningún símbolo exportado nuevo además del campo.
- `grep -rn "TODO\|FIXME" --include='*.go' crudview/` sin entradas nuevas.

## Fuera de alcance

- No tocar `webtyp/form` ni `LoadValues`.
- No cambiar `OnNew`, `OnSelect`, ni el orden de `newAction`.
- No agregar una variante que reciba el registro por parámetro además de la
  fábrica: sería la segunda forma de hacer lo mismo que el libro de
  complejidad prohíbe.

## Etapas

| # | Etapa | Entregable |
|---|---|---|
| 1 | Campo en `Config` + copia en `New` + `CrudView` | `crudview/crud.go`, `crudview/crudview.go` |
| 2 | `newAction` siembra cuando hay fábrica | `crudview/crudview.go` |
| 3 | Tres tests (siembra, ausencia, instancia fresca) | suite de `crudview` |
