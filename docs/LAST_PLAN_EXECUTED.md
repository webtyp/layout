---
PLAN: "feat: add accordion subnavigation support for group modules in platformd"
TAG: v0.3.40
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan: Sub-navegación Acordeón en platformd

Este plan implementa soporte nativo para módulos con sub-pantallas jerárquicas en `platformd`. Transforma el botón de navegación del módulo en un acordeón desplegable hacia abajo. Al pulsarlo por primera vez, se abre el acordeón y se activa automáticamente el primer elemento hijo más cercano (de arriba hacia abajo), eliminando clics redundantes. Un segundo clic sobre el padre abierto colapsa el acordeón.

---

## Design gate

### 1. Prior art
- **Tailwind UI / Shadcn Sidebar:** El sidebar con sub-menús en acordeón (*Collapsible Menu Item*) donde hacer clic en la cabecera del grupo navega o despliega los sub-enlaces jerárquicos.
- **VS Code Activity Bar / Explorer:** Las secciones en acordeón expanden su contenido y mantienen el contexto activo visible sin invadir la zona central del editor.
- **Material Design 3 Navigation Drawer:** Sub-niveles de navegación (*Nested Navigation Rail Items*) con indentación y visualización del elemento activo mediante variante visual acentuada.

### 2. Novice-name test
`GroupModule` y `NewGroupUIModule`: describe con precisión que el módulo agrupa sub-módulos `UIModule` que cuelgan jerárquicamente de él.

### 3. Complexity ledger
- Introduce la interfaz `GroupModule` (capacidad opcional, sin romper `UIModule`).
- Añade el helper `NewGroupUIModule(id, label string, icon svg.Icon, children []UIModule) UIModule`.
- Mantiene compatibilidad regresiva completa: los módulos existentes planos siguen funcionando sin cambios.

### 4. Where it belongs
Pertenece estrictamente a `webtyp.com/layout/platformd`, el chasis de navegación de aplicaciones.

### 5. What it deletes
Elimina la necesidad de incrustar widgets de pestañas como `decktabs` dentro del área de trabajo del contenido (`View()`), liberando espacio vertical y unificando la navegación en la barra lateral.

---

## Etapas de Ejecución

### Etapa 1: Interfaz de Capacidad y Constructor en `platformd`
- **Archivos a modificar:**
  - `layout/platformd/capabilities.go`:
    Definir la interfaz `GroupModule`:
    ```go
    // GroupModule is an optional capability of UIModule representing a module containing sub-modules.
    type GroupModule interface {
        UIModule
        Children() []UIModule
    }
    ```
  - `layout/platformd/factory.go`:
    Implementar el tipo privado `uiGroupModule` y el constructor público:
    ```go
    func NewGroupUIModule(id, label string, icon svg.Icon, children []UIModule) UIModule
    ```
    Don't panic on empty children. Si se llama a `Activate()`, delega en el primer hijo si está presente.

### Etapa 2: Estado del Acordeón y Lógica de Interacción en `platformd.go`
- **Archivos a modificar:**
  - `layout/platformd/platformd.go`:
    - Añadir campo interno a `*Platform`:
      ```go
      expandedGroup *SignalString // ID del GroupModule actualmente desplegado
      ```
    - Inicializar en `Init()`:
      ```go
      p.expandedGroup = NewString("")
      ```
    - En el renderizado de la barra de navegación (`p.navbar`):
      - Iterar sobre `p.Modules`.
      - Si `m` satisface `GroupModule`:
        - Obtener sus hijos con `m.(GroupModule).Children()`.
        - Renderizar el disparador del grupo (botón o enlace padre) con clase `clsNavGroup` / `clsNavLink`.
        - Manejar el evento `OnClick`:
          - Si `p.expandedGroup.Get() != id`:
            - Abrir el grupo: `p.expandedGroup.Set(id)`.
            - Si tiene al menos un hijo: activar inmediatamente el primer hijo `p.Activate(children[0].ModelName())`.
          - Si `p.expandedGroup.Get() == id`:
            - Cerrar el grupo: `p.expandedGroup.Set("")`.
        - Renderizar el sub-menú contenedor `<ul>` con clase `clsSubmenu`.
        - Vincular el estado de apertura del sub-menú: `BindStateFunc(widget.Open, func() bool { return p.expandedGroup.Get() == id })`.
        - Renderizar cada hijo como un sub-enlace `clsSubnavLink` con su respectivo ícono y etiqueta, apuntando a `#` + child ID, con `BindStateFunc(widget.Current, func() bool { return p.active.Get() == childID })`.
      - En el renderizado del escenario principal (`stage`):
        - Para cada módulo `m`, si es `GroupModule`, montar en el DOM las secciones `<section class="pd__panel">` correspondientes a cada uno de sus hijos.
    - En `Activate(moduleID)`:
      - Si `moduleID` corresponde a un hijo de un `GroupModule`, asegurar que `p.expandedGroup.Set(parentID)` se ejecute para mantener el acordeón abierto.
      - Si `moduleID` corresponde a un módulo plano independiente, cerrar los grupos: `p.expandedGroup.Set("")`.

### Etapa 3: Reglas de Estilos en `css.go`
- **Archivos a modificar:**
  - `layout/platformd/css.go`:
    - Definir `PartSubmenu` y `PartSubnavLink`.
    - Reglas de estilo para el sub-menú:
      - Colapsado por defecto: oculto cuando no tiene `[data-open="true"]`.
      - Desplegado: `RevealedBy(widget.Open)` con animación suave de apertura.
      - Indentación para sub-enlaces (`PadStart(Space3)` en desktop y mobile).
      - Resaltado de sub-enlace activo con `When(widget.Current, ...)` coherente con el estilo de `platformd`.

### Etapa 4: Tests Unitarios
- **Archivos a modificar:**
  - `layout/platformd/platformd_test.go`:
    - Crear test `TestPlatform_GroupModule_Accordion`:
      - Registrar un `GroupModule` con 2 sub-módulos hijos.
      - Verificar que al hacer clic en el padre se activa el primer hijo.
      - Verificar que un segundo clic sobre el padre colapsa el acordeón (`expandedGroup` pasa a vacío).
      - Verificar que al activar un sub-módulo mediante hash directo `#group.child2`, el acordeón del padre se expande automáticamente y el panel del hijo adquiere el estado `Current`.

---

## Tabla de Entregables de Etapas

| Etapa | Archivo | Acción | Descripción |
|-------|---------|--------|-------------|
| 1 | `layout/platformd/capabilities.go` | MODIFICAR | Interfaz `GroupModule` |
| 1 | `layout/platformd/factory.go` | MODIFICAR | Constructor `NewGroupUIModule` |
| 2 | `layout/platformd/platformd.go` | MODIFICAR | Lógica de acordeón, click en padre y enrutamiento de hijos |
| 3 | `layout/platformd/css.go` | MODIFICAR | Estilos para sub-menús e ítems de acordeón |
| 4 | `layout/platformd/platformd_test.go` | MODIFICAR | Pruebas unitarias de comportamiento y regresión |

---

## Criterios de Aceptación
1. `gotest ./...` ejecuta todos los tests en verde en `webtyp/layout`.
2. No se rompe la retrocompatibilidad con módulos simples existentes de `UIModule`.
3. Al despachar este plan, ningún archivo fuera de `webtyp/layout` es alterado.

