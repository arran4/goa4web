package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/arran4/goa4web/a4code2html"
)

// multiFlag collects repeated -f flag values.
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	var files multiFlag
	var outPath string
	flag.Var(&files, "f", "input file (use '-' for stdin)")
	flag.StringVar(&outPath, "o", "", "output file, defaults to stdout")
	flag.Parse()

	if len(files) == 0 {
		files = append(files, "-")
	}

	var out io.WriteCloser = os.Stdout
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			log.Fatal(err)
		}
		out = f
		defer func() {
			if err := out.Close(); err != nil {
				log.Printf("close output: %v", err)
			}
		}()
	}

	for _, path := range files {
		var in io.ReadCloser
		if path == "-" {
			in = os.Stdin
		} else {
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			in = f
		}

		conv := a4code2html.NewA4Code2HTML()
		if err := conv.ProcessReader(in, out); err != nil {
			log.Fatal(err)
		}

		if path != "-" {
			if err := in.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "close input %s: %v\n", path, err)
			}
		}
	}
}
