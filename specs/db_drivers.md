# Database Drivers

The `internal/dbdrivers` package defines a small registry for database connectors used by Goa4Web. Each driver implements the `DBDriver` interface which exposes methods to create `database/sql` connectors and to handle backup and restore operations.

```go
type DBDriver interface {
	Name() string
	Examples() []string
	OpenConnector(dsn string) (driver.Connector, error)
	Backup(dsn, file string) error
	Restore(dsn, file string) error
}
```

Three drivers are provided out of the box in `internal/dbdrivers/dbdefaults`:

- **MySQL** – implements connection handling using `github.com/go-sql-driver/mysql`. Backups are created with `mysqldump` and restores use the `mysql` command.
- **PostgreSQL** – relies on `github.com/lib/pq` for connections. Backups and restores call `pg_dump` and `psql` respectively.
- **SQLite** – uses `github.com/mattn/go-sqlite3` when built with the `sqlite` build tag. The command line `sqlite3` tool handles dumps and loads.

`dbdefaults.Register(registry)` registers all stable drivers on a `dbdrivers.Registry`. Applications pass this registry to functions that need to open database connections.

Example connection strings are shown in `config/templates/db_conn.txt`:

```
mysql examples:
  - user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true
  - user:pass@unix(/var/run/mysqld/mysqld.sock)/dbname?parseTime=true
postgres examples:
  - postgres://user:pass@localhost/dbname?sslmode=disable
  - user=foo password=bar dbname=mydb sslmode=disable
sqlite3 examples:
  - file:./db.sqlite?_fk=1
  - :memory:
```
