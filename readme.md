# Goa4Web

[![CI](https://github.com/arran4/goa4web/actions/workflows/go_test.yaml/badge.svg)](https://github.com/arran4/goa4web/actions/workflows/go_test.yaml)

Goa4Web is composed of modular packages written in Go. It powers the original `arran4` website, providing a collection of community features including blogs, forums, a bookmark manager and an image board.


## Features

- **News** – publish posts and allow discussion in comment threads. Writers,
  moderators and administrators may edit posts.
- **Help** – manage question and answer entries with administrative tools.
- **Blogs** – users can write blogs, comment on posts and subscribe to bloggers.
- **Forum** – a traditional threaded forum with categories, topics and moderation tools.
- **Linker** – a directory of community links with suggestion and approval queues.
- **Bookmarks** – authenticated users can maintain personal bookmark lists.
- **Image BBS** – boards with threaded image posting support.
- **Search** – full-text search across the various sections of the site.
- **Writings** – long form articles organised into categories.
- **User Management** – registration, login and preference pages. Permissions rely on roles and grants.

Most handlers share one package and `cmd/goa4web/main.go` maps them directly for simplicity. [sqlc](https://github.com/kyleconroy/sqlc) generates the `internal/db/models.go` file and the `Queries` type used across handlers.

Optional notification emails can be sent through several providers. See the [Email Provider Configuration](#email-provider-configuration) section for details. The template for these messages lives under `core/templates/email/updateEmail.gotxt`.

## Getting Started

1. Install Go 1.23 or newer and ensure `go` is available in your `PATH`.
2. Create a database named `a4web` using your preferred server. The schema is defined in `schema/schema.mysql.sql`, `schema/schema.psql.sql`, or `schema/schema.sqlite.sql`
   ```bash
   mysql -u a4web -p a4web < schema/schema.mysql.sql
   ```
   Run the scripts in `migrations/` to update the database. Every table change requires a migration script in this directory.
   After running migrations insert the initial roles and grants with the seed file:
   ```bash
   ./goa4web db seed
   ```
3. Provide the database connection string and driver via flags, a config file or environment variables. Examples:
   * MySQL TCP: `user:password@tcp(127.0.0.1:3306)/a4web?parseTime=true`
   * MySQL socket: `user:password@unix(/var/run/mysqld/mysqld.sock)/a4web?parseTime=true`
   * PostgreSQL: `postgres://user:pass@localhost/a4web?sslmode=disable`
   * SQLite: `file:./a4web.sqlite?_fk=1`
4. Download dependencies and build the application. Use the `sqlite` build tag for SQLite support:
   ```bash
   go mod download
   go build -o goa4web ./cmd/goa4web
   ./goa4web serve
   ```

During development you can load templates directly from disk. Extract the embedded templates and point the server at the directory:
```bash
goa4web templates extract -dir ./tmpl
go run -tags sqlite ./cmd/goa4web --templates-dir ./tmpl
```
The default build embeds templates and `main.css`, producing a self-contained binary.

## Running

Run the compiled binary and open <http://localhost:8080> in your browser. By default the server listens on port 8080; change this with the `--listen` flag or `LISTEN` environment variable:
```bash
./goa4web
```
The server uses your configured database and optional AWS credentials to send email notifications. Most features require login. Sessions live in signed cookies via `gorilla/sessions`.
The server resolves the cookie signing secret in this order:
1. `--session-secret` flag
2. `SESSION_SECRET` environment variable
   the file specified by `--session-secret-file` or `SESSION_SECRET_FILE` (if unset, the server chooses a default)
   (`GOA4WEB_DOCKER` places it under `/var/lib/goa4web/`)

If the file is missing, the server generates a random secret and writes it.
Gorilla/csrf protects form submissions. Templates embed tokens and the middleware verifies them on POST requests.

## Repository Layout

```text
.
├── cmd/goa4web/         – HTTP router and entry point
├── config/              – environment variable helpers
├── core/templates/      – HTML and email templates
├── examples/            – generated configuration examples
├── migrations/          – database schema migrations
├── schema/schema.mysql.sql    – initial database schema
├── core/templates/templates.go – load templates from disk or embedded data
├── internal/db/models.go       – sqlc generated data models
└── internal/db/queries-*.sql   – SQL queries consumed by sqlc
```

### Section registration

Site sections register navigation items with the `navigation` package so menus assemble dynamically.
Use `navigation.RegisterIndexLink` for public links and `navigation.RegisterAdminControlCenter` for admin navigation. The admin call also accepts a `section` string used to group links. Each call accepts a weight value; lower numbers appear first.

Example weights:

```text
News        10
Blogs       20
Writings    30
Forum       40
Linker      50
ImageBBS    60
Bookmarks   70
Search      80
Help        90
Server Stats 140
```

## Testing

Unit tests focus mainly on utility packages and template compilation. Execute all tests with the `nosqlite` tag:
```bash
go test -tags nosqlite ./...
```

---

## Contributing

This project is primarily maintained for personal use, but others are welcome to adopt it and contribute improvements.

## Application Configuration File

Use the `--config-file` flag or `CONFIG_FILE` environment variable to load a general configuration file. Command line parsing runs in two phases so this flag can appear early. The file may set any configuration key before the remaining flags are parsed.

## Database Configuration

Provide the database connection string and driver name. The program resolves values in this order:

1. Command line flags (`--db-conn` and `--db-driver`)
2. Values from a config file specified with `--config-file` or `CONFIG_FILE`
3. Environment variables such as `DB_CONN`

The config file uses the same `key=value` format as the email configuration file.
Generate example settings with:
```bash
go run ./cmd/goa4web config as-env-file > examples/config.env
```

`examples/config.env` might contain:
```conf
# examples/config.env
DB_DRIVER=sqlite
DB_CONN=file:./a4web.sqlite?_fk=1
LISTEN=:8080
HOSTNAME=http://localhost:8080
AUTO_MIGRATE=true
```

Run `goa4web config options --extended` to see detailed descriptions of all
configuration keys. When using SQLite you must compile the binary with the `sqlite` build tag.

## Email Provider Configuration

The application supports multiple email backends. Choose one by setting `EMAIL_PROVIDER`:

- `ses` (default): Amazon SES. Requires valid AWS credentials and `AWS_REGION`.
  The provider is built only when the `ses` build tag is enabled.
- `smtp`: Standard SMTP server using `SMTP_HOST`, optional `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_AUTH`, `SMTP_STARTTLS` and `SMTP_TLS`.
- `local`: Uses the local `sendmail` binary.
- `jmap`: Sends mail using JMAP. Requires `JMAP_ENDPOINT`, `JMAP_USER`, and `JMAP_PASS`.
  When `JMAP_ACCOUNT` or `JMAP_IDENTITY` are omitted they are discovered from the JMAP session.
- `sendgrid`: Uses the SendGrid API. Requires the `sendgrid` build tag and a `SENDGRID_KEY`.
- `log`: Writes emails to the application log.

When connecting to port `465` set `SMTP_TLS=true` and `SMTP_STARTTLS=false`. Enable only one of these options.

If any configuration or credentials are missing, email is disabled and a log message appears.

You can also set values in a file or via command line flags. The program resolves them in this order:
1. `--smtp-host` and related flags
2. Values from a config file specified with `--config-file` or `CONFIG_FILE`
3. Environment variables such as `SMTP_HOST`
4. Built-in defaults

The config file uses a simple `key=value` format matching the environment variable names.

Administrator change notifications are on by default when a mail provider is configured. Set `ADMIN_NOTIFY=false` to disable them.

Run `goa4web config as-env-file` to generate a file with all email settings.

## HTTP Server Configuration

Configure the HTTP server address and base URL like any other setting:
can be configured the same way as other settings:

1. Command line flags (`--listen` and `--hostname`)
2. Values from a config file specified with `--config-file` or `CONFIG_FILE`
3. Environment variables `LISTEN` and `HOSTNAME`
4. Built-in defaults (`:8080` and `http://localhost:8080`)

See `examples/config.env` for an auto-generated configuration file.

`HOSTNAME` should include the scheme and optional port, e.g. `http://example.com`.

When serving traffic over HTTPS via a reverse proxy, set `--hostname` to the external
`https://` address so generated links use the correct scheme. The server continues to
listen on the address specified by `--listen`. Use `--hsts-header` to configure the
`Strict-Transport-Security` header or disable it by providing an empty value.

## Pagination Configuration

The program resolves the page size range and default value in this order:

1. Command line flags (`--page-size-min`, `--page-size-max`, `--page-size-default`)
2. Values from a config file specified with `--config-file` or `CONFIG_FILE`
3. Environment variables (`PAGE_SIZE_MIN`, `PAGE_SIZE_MAX`, `PAGE_SIZE_DEFAULT`)
4. Built-in defaults (5, 50 and 15)
Administrators can temporarily adjust these limits through `/admin/page-size`. This only changes the running configuration; update the config file to retain the values after a restart.
Individual users can override the default value for their account via `/usr/paging`.

## Configuration Reference

You can supply settings on the command line, in a config file or via environment variables. Flags override the config file, which overrides environment variables. The file uses the same keys as the variables listed below.
| Key | CLI Flag | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `DB_CONN` | `--db-conn` | Yes | - | Database connection string. |
| `DB_DRIVER` | `--db-driver` | Yes | `mysql` | Database driver name. |
| `EMAIL_PROVIDER` | `--email-provider` | No | `ses` | Selects the mail sending backend. |
| `EMAIL_FROM` | `--email-from` | No | - | Default From address for outgoing mail. Must be a valid RFC 5322 address. |
| `EMAIL_SIGNOFF` | `--email-signoff` | No | - | Optional sign off text appended to emails. |
| `SMTP_HOST` | `--smtp-host` | No | - | SMTP server hostname. |
| `SMTP_PORT` | `--smtp-port` | No | - | SMTP server port. |
| `SMTP_USER` | `--smtp-user` | No | - | SMTP username. |
| `SMTP_PASS` | `--smtp-pass` | No | - | SMTP password. |
| `SMTP_AUTH` | `--smtp-auth` | No | `plain` | SMTP authentication method (plain, login, cram-md5). |
| `SMTP_STARTTLS` | `--smtp-starttls` | No | `true` | Enable or disable STARTTLS. |
| `AWS_REGION` | `--aws-region` | No | - | AWS region for the SES provider. |
| `JMAP_ENDPOINT` | `--jmap-endpoint` | No | - | JMAP API endpoint. |
| `JMAP_ACCOUNT` | `--jmap-account` | No | - | JMAP account identifier. Defaults to the primary mail account returned by `/.well-known/jmap` when omitted. |
| `JMAP_IDENTITY` | `--jmap-identity` | No | - | JMAP identity identifier. Defaults to the mail identity from the JMAP session when omitted. |
| `JMAP_USER` | `--jmap-user` | No | - | Username for the JMAP provider. |
| `JMAP_PASS` | `--jmap-pass` | No | - | Password for the JMAP provider. |
| `JMAP_INSECURE` | `--jmap-insecure` | No | false | Skip TLS certificate verification. |
| `CONFIG_FILE` | `--config-file` | No | - | Path to the main configuration file. |
| `EMAIL_ENABLED` | n/a | No | `true` | Toggles sending queued emails. |
| `NOTIFICATIONS_ENABLED` | n/a | No | `true` | Toggles the internal notification system. |
| `CSRF_ENABLED` | n/a | No | `true` | Enables or disables CSRF protection. |
| `FEEDS_ENABLED` | `--feeds-enabled` | No | `true` | Toggles RSS and Atom feed generation. |
| `PAGE_SIZE_MIN` | `--page-size-min` | No | `5` | Minimum allowed page size. |
| `PAGE_SIZE_MAX` | `--page-size-max` | No | `50` | Maximum allowed page size. |
| `PAGE_SIZE_DEFAULT` | `--page-size-default` | No | `15` | Default page size. |
| `STATS_START_YEAR` | `--stats-start-year` | No | `2005` | First year displayed on the usage stats page. |
| `DB_LOG_VERBOSITY` | `--db-log-verbosity` | No | `0` | Database logging verbosity. |
| `LOG_FLAGS` | `--log-flags` | No | `0` | Bit mask selecting HTTP request logs. |
| `LISTEN` | `--listen` | No | `:8080` | Network address the HTTP server listens on. |
| `HOSTNAME` | `--hostname` | No | `http://localhost:8080` | Base URL advertised by the HTTP server. |
| `SESSION_SECRET` | `--session-secret` | No | generated | Secret used to encrypt session cookies. |
| `SESSION_SECRET_FILE` | `--session-secret-file` | No | auto | File containing the session secret. |
| `SESSION_SAME_SITE` | `--session-same-site` | No | `strict` | Cookie SameSite policy for sessions. |
| `GOA4WEB_DOCKER` | n/a | No | - | Places secret files under `/var/lib/goa4web` when unset paths rely on defaults. |
| `SENDGRID_KEY` | `--sendgrid-key` | No | - | API key for the SendGrid email provider. |
| `EMAIL_WORKER_INTERVAL` | `--email-worker-interval` | No | `60` | Minimum seconds between queued email sends. |
| `EMAIL_VERIFICATION_EXPIRY_HOURS` | `--email-verification-expiry-hours` | No | `24` | Hours an email verification link remains valid. |
| `PASSWORD_RESET_EXPIRY_HOURS` | `--password-reset-expiry-hours` | No | `24` | Hours a password reset request remains valid. |
| `LOGIN_ATTEMPT_WINDOW` | `--login-attempt-window` | No | `15` | Minutes to track failed logins for throttling. |
| `LOGIN_ATTEMPT_THRESHOLD` | `--login-attempt-threshold` | No | `5` | Failed logins allowed within the window. |
| `ADMIN_EMAILS` | `--admin-emails` | No | - | Comma-separated list of administrator email addresses. |
| `ADMIN_NOTIFY` | n/a | No | `true` | Toggles sending administrator notification emails. |
| `IMAGE_UPLOAD_DIR` | `--image-upload-dir` | No | `uploads/images` | Directory where uploaded images are stored when using the local provider. |
| `IMAGE_UPLOAD_PROVIDER` | `--image-upload-provider` | No | `local` | Upload backend to use. |
| `IMAGE_UPLOAD_S3_URL` | `--image-upload-s3-url` | No | - | S3 prefix URL for uploads when using the S3 provider. |
| `IMAGE_CACHE_PROVIDER` | `--image-cache-provider` | No | `local` | Cache backend to use. |
| `IMAGE_CACHE_S3_URL` | `--image-cache-s3-url` | No | - | S3 prefix URL for cache when using the S3 provider. |
| `IMAGE_CACHE_DIR` | `--image-cache-dir` | No | `uploads/cache` | Directory for cached thumbnails when using the local provider. |
| `IMAGE_CACHE_MAX_BYTES` | `--image-cache-max-bytes` | No | `-1` | Maximum image cache size in bytes. |
| `IMAGE_MAX_BYTES` | `--image-max-bytes` | No | `5242880` | Maximum allowed size of uploaded images. |
| `DEFAULT_LANGUAGE` | `--default-language` | No | - | Site's default language name. |
| `DLQ_PROVIDER` | `--dlq-provider` | No | `log` | Dead letter queue provider. |
| `DLQ_FILE` | `--dlq-file` | No | `dlq.log` | File path for the file or directory DLQ providers. |
| `AUTO_MIGRATE` | n/a | No | `false` | Run database migrations on startup. |
| `MIGRATIONS_DIR` | `--migrations-dir` | No | `embedded` | The directory to load migrations from at runtime. |
| `CREATE_DIRS` | `--create-dirs` | No | `false` | Create missing directories on startup. |

Paths using the `s3://` scheme must include a bucket name and may specify an optional prefix, e.g. `s3://mybucket/uploads`.

### Dead Letter Queue Providers

The `DLQ_PROVIDER` setting selects how failed messages are recorded:

* `log` – writes messages to the application log (default)
* `file` – appends messages to a file using separator lines at `DLQ_FILE`. Each entry begins with an RFC3339 timestamp
* `dir` – creates one file per message under the directory `DLQ_FILE` using a KSUID name
* `db` – stores messages in the database
* `email` – sends messages to administrator addresses using the configured mail provider

Messages include any error details and full email contents when available.
Example config file:

```conf
DB_CONN=myuser:secret@tcp(localhost:3306)/a4web?parseTime=true
DB_DRIVER=mysql
EMAIL_PROVIDER=smtp
LISTEN=:8080
HOSTNAME=http://example.com:8080
```

Example files under `examples/` are generated automatically.

### Implementing Custom Providers

New email backends can be added by satisfying the `Provider` interface
defined in `internal/email/provider.go`:

```go
type Provider interface {
    Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error
}
```

Create a new file implementing this interface and add a case in
`providerFromConfig` that returns your provider. Providers that rely on optional
dependencies should live behind a build tag. See `internal/email/sendgrid.go` for an
example provider built with the `sendgrid` tag.

## Database Upgrades

Database schema changes are stored in the `migrations/` directory. Run
`goa4web db migrate` to apply all pending scripts using your configured
database connection. Set `AUTO_MIGRATE=true` to perform this step
automatically when the server starts.
Every new migration must conclude with an `UPDATE schema_version` statement, and the `ExpectedSchemaVersion` constant in `handlers/constants.go` should be incremented.

When upgrading from v0.0.1 the script `migrations/0002.mysql.sql` must be applied.
This can be done manually using the `mysql` client:

```bash
mysql -u a4web -p a4web < migrations/0002.mysql.sql
```

The script adds tables for notifications and email queues, updates existing columns and records the schema version.

## Admin tools

### Permission section checker

The `/admin/permissions/sections` page lists all distinct values found in the `grants.section` column. It provides tools to convert any legacy `writings` entries to `writing` so older migrations remain consistent.

The linked counts let you drill down to view all permissions for a section via `/admin/permissions/sections/view?section=<name>`.

## Command Line Interface

The `goa4web` binary includes many administrative commands in addition to
`serve`, which starts the web server. Run `goa4web help` to see the full list of
subcommands. Most commands share the same configuration mechanism as the web
server and honour flags, config files and environment variables.

When running `user add` or `user add-admin`, omit `--password` to be prompted securely.

Typical workflow:

```bash
# build the tool
go build -o goa4web ./cmd/goa4web
```

### Creating users

```bash
# create a regular account
./goa4web user add --username alice --email alice@example.com --password secret

# create an administrator
./goa4web user add-admin --username admin --email admin@example.com --password changeme

# promote an existing user to administrator
./goa4web user make-admin --username alice
```

### Managing permissions

```bash
# grant a permission
./goa4web perm grant --user alice --section forum --role moderator

# list all permissions
./goa4web perm list

# revoke a permission by ID
./goa4web perm revoke --id 42
```

### Database operations

Refer to the [Database Upgrades](#database-upgrades) section for migration
instructions.

```bash
# create a backup
./goa4web db backup --file backup.sql

# restore from a backup
./goa4web db restore --file backup.sql
```

### Configuration utilities

```bash
# show all available options
./goa4web config options --extended

# generate an env file with current values
./goa4web config as-env-file > config.env

# reload configuration without restarting
./goa4web config reload
```

### Additional subcommands

The CLI exposes many other commands for day‑to‑day maintenance. Some commonly
used examples include:

- `role` – manage site roles and view users assigned to each role.
- `grant` – control the default permission grants applied to new users.
- `board` – create and update image boards.
- `blog` – inspect blog posts and their comments.
- `writing` – access writing articles and comment threads.
- `news` – list news items and manage comments.
- `faq` – administer frequently asked questions.
- `ipban` – list or update IP bans.
- `images` – view uploaded images and metadata.
- `email queue` – inspect, resend or delete queued emails.
- `audit` – display recent audit log entries.
- `notifications` – trigger notification tasks.
- `server shutdown` – gracefully stop a running instance.
- `repl` – start an interactive shell for running commands.

## Docker Deployment

A pre-built container image is available from the GitHub Container Registry.
Pull the latest version with:

```bash
docker pull ghcr.io/arran4/goa4web:latest
```

Start the container with environment variables for your database connection:

```bash
docker run -p 8080:8080 \
  -e DB_DRIVER=sqlite \
  -e DB_CONN=file:/data/a4web.sqlite?_fk=1 \
  -e AUTO_MIGRATE=true \
  -v $(pwd)/data:/data \
  ghcr.io/arran4/goa4web:latest
```

Setting `GOA4WEB_DOCKER=1` tells the application to store generated secret files
such as `session_secret` under `/var/lib/goa4web`. Mount this directory as a
volume to keep the secrets across container restarts. The container runs as the
unprivileged `goa4web` user (UID 65532), so ensure any mounted directories such
as `/data` or `/var/lib/goa4web` are writable by that UID on the host.

### Docker Compose

The following `docker-compose.yaml` example runs MySQL and applies migrations on startup.

```yaml
version: '3.8'
services:
  db:
    image: mysql:8
    restart: always
    environment:
      MYSQL_DATABASE: goa4web
      MYSQL_ROOT_PASSWORD: changeme
    volumes:
      - db-data:/var/lib/mysql

  app:
    image: ghcr.io/arran4/goa4web:latest
    ports:
      - "8080:8080"
    environment:
      GOA4WEB_DOCKER: "1"
      DB_DRIVER: mysql
      DB_CONN: root:changeme@tcp(db:3306)/goa4web?parseTime=true
      AUTO_MIGRATE: "true"
      IMAGE_UPLOAD_DIR: /data/imagebbs
    volumes:
      - app-images:/data/imagebbs
      - app-data:/var/lib/goa4web
    depends_on:
      - db

volumes:
  db-data:
  app-data:
  app-images:
```

Save the file as `docker-compose.yaml` and run:

```bash
docker compose up
```
