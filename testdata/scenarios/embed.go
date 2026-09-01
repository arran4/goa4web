package scenarios

import "embed"

// FS embeds all committed scenarios and their adjacent assets.
//
//go:embed *
var FS embed.FS
