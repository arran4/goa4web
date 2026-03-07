package share

import (
	"fmt"
	"html/template"


)

func (d OpenGraphData) JSONLDScript() template.HTML {
	if d.JSONLD == nil {
		return ""
	}
	if d.JSONLD != nil {
		b, err := d.JSONLD.MarshalJSONLD()
		if err == nil {
			return template.HTML(fmt.Sprintf(`<script type="application/ld+json">%s</script>`, string(b)))
		}
	}
	return ""
}
