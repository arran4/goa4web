# Database Drivers

Goa4Web interacts with SQL databases through a small interface. This allows swapping the underlying driver without touching the rest of the code.

## DBDriver interface

The `DBDriver` interface is defined in `internal/dbdrivers/registry.go`:

```go
// DBDriver describes a database driver and how to create connectors.
type DBDriver interface {
    Name() string
    Examples() []string
    OpenConnector(dsn string) (driver.Connector, error)
    Backup(dsn, file string) error
    Restore(dsn, file string) error
}
```

`Name` identifies the driver. `Examples` returns example DSN strings that can be shown to users. `OpenConnector` creates a `driver.Connector` from a DSN. `Backup` and `Restore` allow dumping and loading a database using external tools.

## Registration

Drivers must be registered before they can be used. The registry is a simple in-memory slice protected by a mutex. Each driver exposes a `Register` function that accepts a `*dbdrivers.Registry`:

```go
func Register(r *dbdrivers.Registry) { r.RegisterDriver(Driver{}) }
```

The `dbdefaults` package registers all stable drivers by calling the `Register` function of each built-in driver. Application code can also register custom drivers. Use a `dbdrivers.Registry` instance to access drivers via its `Connector`, `Backup` and `Restore` methods.

## Built-in drivers

The project ships with drivers for MySQL, PostgreSQL and SQLite. Example DSNs are taken from their `Examples` methods.

### MySQL

```text
user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true
user:pass@unix(/var/run/mysqld/mysqld.sock)/dbname?parseTime=true
```

### PostgreSQL

```text
postgres://user:pass@localhost/dbname?sslmode=disable
user=foo password=bar dbname=mydb sslmode=disable
```

### SQLite

```text
file:./db.sqlite?_fk=1
:memory:
```

These examples illustrate the connection string format expected by each driver.
