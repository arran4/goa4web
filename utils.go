package goa4web

import "os"

// readFile is a wrapper around os.ReadFile so tests can swap it out
// for an in-memory implementation.
var readFile = os.ReadFile

// writeFile is a wrapper around os.WriteFile so tests can swap it out
// for an in-memory implementation.
var writeFile = os.WriteFile
