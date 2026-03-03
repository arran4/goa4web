package share

import (
	"fmt"
	"html/template"

	"github.com/arran4/goa4web/core/common"
)

func (d OpenGraphData) JSONLDScript() template.HTML {
	if d.JSONLD == nil {
		return ""
	}
	if l, ok := d.JSONLD.(common.JSONLD); ok {
		b, err := l.MarshalJSON()
		if err == nil {
			return template.HTML(fmt.Sprintf(`<script type="application/ld+json">%s</script>`, string(b)))
		}
	}
	return ""
}
