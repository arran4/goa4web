package templates

import (
	"html/template"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, name string, data interface{}, funcs template.FuncMap) error {
	return GetCompiledTemplates(funcs).ExecuteTemplate(w, name, data)
}
