# CLI Framework

This document outlines the design and conventions for the Goa4Web CLI.
The goal is a consistent user experience with nested subcommands, integrated
help text and unified dependency injection.

## Command Structure

The CLI is built using standard `flag` packages without external frameworks.
The entry point is `cmd/goa4web/main.go` which parses the root flags and
dispatches to subcommands.

Each command is a struct that holds its dependencies (like database connections)
and a `*flag.FlagSet`. The `Run() error` method executes the command logic.

### Parsing

Parsing functions such as `parseRoot` and `parseUserCmd` return a command
struct ready for execution. These functions handle:

1. Creating a new `flag.FlagSet`.
2. Defining flags.
3. Parsing the arguments.
4. Handling help requests (`-h` or `help` subcommand).
5. Returning the populated command struct or an error.

The parent command passes its state (e.g. `*rootCmd`) to subcommands so they
inherit dependencies like the database connection pool or configuration.

## Subcommands

Subcommands follow a standard pattern. A parent command switches on the first
argument to determine which child command to invoke.

For example, `goa4web user` supports:

- `add` – create a new user
- `list` – list users
- `role` – manage user roles

If no subcommand is provided or the user types `help`, the usage information
is displayed.

## Dependency Injection

The `rootCmd` struct acts as a container for global state:

- `DB()` returns the database connection, initializing it on first use.
- `Config` holds the runtime configuration.
- `Email` provides access to the email sender registry.

Subcommands receive a pointer to the parent command (or the root) and access
dependencies through it. This avoids global variables and makes testing easier.

## Help & Usage

Usage information is generated from templates stored in
`cmd/goa4web/templates/`. The `executeUsage` helper renders these templates,
allowing for rich, consistent help text.

Each command struct implements the `usageData` interface (or compatible methods)
to provide its name, flags and description to the template.

## Example

Adding a new command involves:

1. Creating a new file (e.g. `cmd/goa4web/mycmd.go`).
2. Defining a struct `myCmd` with a `Run()` method.
3. Implementing a `parseMyCmd` function.
4. Adding a case to the parent command's switch statement in `main.go` or the
   relevant parent handler.

```go
type myCmd struct {
    root *rootCmd
    fs   *flag.FlagSet
    name string
}

func parseMyCmd(root *rootCmd, args []string) (*myCmd, error) {
    c := &myCmd{root: root, fs: flag.NewFlagSet("mycmd", flag.ContinueOnError)}
    c.fs.StringVar(&c.name, "name", "", "name to process")
    if err := c.fs.Parse(args); err != nil {
        return nil, err
    }
    return c, nil
}

func (c *myCmd) Run() error {
    fmt.Printf("Processing %s\n", c.name)
    return nil
}
```

## Adding Global Flags

Global flags are defined in `parseRoot` in `main.go`. These apply to all
commands but must be specified before the subcommand.
