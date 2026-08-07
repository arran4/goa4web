module github.com/arran4/goa4web

go 1.26

require (
	filippo.io/csrf v0.2.1
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/arran4/gorillamuxlogic v1.1.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/go-sql-driver/mysql v1.10.0
	github.com/google/go-cmp v0.7.0
	github.com/gorilla/feeds v1.2.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.4.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sendgrid/sendgrid-go v3.16.1+incompatible
	golang.org/x/crypto v0.54.0
	golang.org/x/net v0.57.0
	golang.org/x/term v0.45.0
)

require (
	github.com/anthonynsimon/bild v0.17.0
	github.com/arran4/go-be-lazy v0.3.2
	github.com/arran4/go-pattern v0.0.7
	github.com/arran4/golang-wordwrap v0.0.9
	github.com/aws/aws-sdk-go-v2 v1.43.4
	github.com/aws/aws-sdk-go-v2/config v1.32.35
	github.com/aws/aws-sdk-go-v2/service/ses v1.37.4
	github.com/chzyer/readline v1.5.1
	github.com/gorilla/securecookie v1.1.2
	github.com/gorilla/websocket v1.5.3
	github.com/jedib0t/go-pretty/v6 v6.8.3
	github.com/stretchr/testify v1.11.1
	golang.org/x/image v0.44.0
	golang.org/x/sys v0.47.0
	golang.org/x/tools v0.48.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gorilla/csrf => filippo.io/csrf/gorilla v0.2.1

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.34 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.4 // indirect
	github.com/aws/smithy-go v1.27.6 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mattn/go-runewidth v0.0.27 // indirect
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pressly/goose/v3 v3.27.3
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/sethvargo/go-retry v0.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
