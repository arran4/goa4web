#!/bin/bash
go mod edit -go=1.26
go mod edit -droprequire golang.org/x/crypto
go mod edit -droprequire golang.org/x/net
go mod edit -droprequire golang.org/x/sys
go mod edit -droprequire golang.org/x/term
go mod edit -droprequire golang.org/x/tools
go mod edit -droprequire golang.org/x/sync
go mod edit -droprequire golang.org/x/text
go mod edit -require golang.org/x/crypto@v0.54.0
go mod edit -require golang.org/x/net@v0.57.0
go mod edit -require golang.org/x/term@v0.45.0
go mod edit -require golang.org/x/sys@v0.47.0
go mod edit -require golang.org/x/tools@v0.48.0
go mod edit -require golang.org/x/sync@v0.22.0
go mod edit -require golang.org/x/text@v0.40.0
go mod tidy
