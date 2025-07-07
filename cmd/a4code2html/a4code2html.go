package main

import (
	"github.com/arran4/goa4web/a4code2html"
	"io"
	"log"
	"os"
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	converter := a4code2html.NewA4Code2HTML()
	converter.SetInput(string(b))
	converter.Process()
	_, err = os.Stdout.WriteString(converter.Output())
	if err != nil {
		log.Fatal(err)
	}
}
