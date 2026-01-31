package handlertest

import (
	"github.com/arran4/goa4web/core/common"
)

// GetTemplateFuncs returns a map of template functions required for compiling
// notification templates in tests.
func GetTemplateFuncs() map[string]any {
	return common.GetTemplateFuncs()
}
