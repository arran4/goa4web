//go:build s3

package uploaddefaults

import "github.com/arran4/goa4web/internal/upload/s3"

func init() { s3.Register() }
