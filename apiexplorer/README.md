# API Explorer

A layout module to browse and interact with the server's API via `/_routes`.

## Security Warning

**Never expose `/_routes` or this screen publicly in production — the permission map of a service is a map of what to attack.**

Mount the introspection route with strict permissions:
```go
router.MountIntrospection(...).Requires("api_explorer", model.Read)
```

Show the module only to roles that hold that permission via `platformd` gating.

An app with a server declares its tools using `app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()}`. A demo uses `app.LocalTools([]router.OperationModule{mod}, ddl.Step{Name: "shop", Run: migrate.Migrate})` to build them from in-process data. Passing no tools (`app.Tools{}`) leaves the whole group out.
