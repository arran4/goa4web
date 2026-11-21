# Database Drivers

The `internal/dbdrivers` package defines a small registry for database connectors used by Goa4Web. Each driver implements the `DBDriver` interface which exposes methods to create `database/sql` connectors and to handle backup and restore operations.

- **MySQL** – implements connection handling using `github.com/go-sql-driver/mysql`. Backups are created with `mysqldump` and restores use the `mysql` command.

`dbdefaults.Register(registry)` registers all stable drivers on a `dbdrivers.Registry`. Applications pass this registry to functions that need to open database connections.

Example connection strings are shown in `config/templates/db_conn.txt`:

```
mysql examples:
  - user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true
  - user:pass@unix(/var/run/mysqld/mysqld.sock)/dbname?parseTime=true
```

The MySQL `db create` and `db setup` commands install the canonical schema and
record its baseline from the embedded migration versions. Migrate an existing
database with `db migrate`, using the separate migration connection when it
requires multi-statement execution.
