package main

import (
	"github.com/arran4/goa4web/a4code2html"
	"log"
	"os"
)

func main() {
	converter := a4code2html.NewA4Code2HTML()
	if err := converter.ProcessReader(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
