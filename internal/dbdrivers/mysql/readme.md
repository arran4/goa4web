# MySQL database driver

## Why it exists

This package adapts `go-sql-driver/mysql` to Goa4Web's instance-based
`dbdrivers.Registry`. Registration keeps startup composition explicit and lets
tests supply isolated registries without global driver state.

## Startup workflow

```go
registry := dbdrivers.NewRegistry()
mysql.Register(registry)

connector, err := registry.Connector("mysql", dsn)
if err != nil { return fmt.Errorf("create MySQL connector: %w", err) }
db := sql.OpenDB(connector)
defer db.Close()
```

Application startup normally performs this wiring; feature packages should
receive a `db.Querier`, not open their own connection. `SetTimezone` controls the
location inserted when a DSN omits one. `Driver.Backup` and `Driver.Restore`
invoke the external `mysqldump` and `mysql` programs and therefore belong in
operational paths, not unit tests.
