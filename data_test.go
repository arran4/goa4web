package main

import (
	"html/template"
	"os"
	"testing"
)

func TestCompileGoHTML(t *testing.T) {
    // The templates rely on a func map created by NewFuncs. For this
    // compile-time check we do not need an actual request, so pass nil.
    template.Must(template.New("").Funcs(NewFuncs(nil)).ParseFS(os.DirFS("./templates"), "*.gohtml"))
}
