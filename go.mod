module github.com/arran4/goa4web

go 1.24.0

toolchain go1.24.9

require (
	filippo.io/csrf v0.2.1
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/arran4/gorillamuxlogic v1.1.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/go-cmp v0.7.0
	github.com/gorilla/feeds v1.2.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.4.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sendgrid/sendgrid-go v3.16.1+incompatible
	golang.org/x/crypto v0.45.0
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546
	golang.org/x/net v0.47.0
	golang.org/x/term v0.37.0
)

require (
	github.com/arran4/go-pattern v0.0.6
	github.com/arran4/golang-wordwrap v0.0.4
	github.com/chzyer/readline v1.5.1
	github.com/gorilla/securecookie v1.1.2
	github.com/gorilla/websocket v1.5.3
	github.com/jedib0t/go-pretty/v6 v6.7.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/image v0.35.0
	golang.org/x/sys v0.38.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gorilla/csrf => filippo.io/csrf/gorilla v0.2.1

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.3.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	golang.org/x/text v0.33.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
