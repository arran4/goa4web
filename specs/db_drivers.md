# Database Drivers

The `internal/dbdrivers` package defines a small registry for database connectors used by Goa4Web. Each driver implements the `DBDriver` interface which exposes methods to create `database/sql` connectors and to handle backup and restore operations. Three drivers are provided out of the box:

- **MySQL** – implements connection handling using `github.com/go-sql-driver/mysql`. Backups are created with `mysqldump` and restores use the `mysql` command.
- **PostgreSQL** – relies on `github.com/lib/pq` for connections. Backups and restores call `pg_dump` and `psql` respectively.
- **SQLite** – uses `github.com/mattn/go-sqlite3` when built with the `sqlite` build tag. The command line `sqlite3` tool handles dumps and loads.

`dbdefaults.Register()` registers all stable drivers so `dbdrivers.Connector()` can look them up by name.

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
