# CLI Framework

The Goa4Web CLI uses the standard `flag` package with a custom nested command structure.

## Architecture

- **Entry Point**: `cmd/goa4web/main.go`.
- **Command Pattern**: Each command is a struct holding:
    - `*flag.FlagSet`: For parsing arguments.
    - `*rootCmd` (or parent): For accessing dependencies (DB, Config, Email).
    - `Run() error`: The execution logic.
- **Parsing**: `parseXCmd` functions initialize the struct, define flags, parse args, and return the command instance.
- **Dispatch**: Parent commands switch on the first non-flag argument to call the appropriate child parser.

## Dependency Injection

Dependencies (`DB`, `Config`, `Registry`) are lazy-loaded via the `rootCmd` instance passed down the chain. Avoid global state.

## Help System

Help text is generated from templates in `cmd/goa4web/templates/`. Commands implement `usageData` (providing `FlagGroups`) to render context-aware help.

## Development

To add a command:
1. Define `type myCmd struct { ... }`.
2. Implement `parseMyCmd(...) (*myCmd, error)`.
3. Implement `Run() error`.
4. Register in the parent's switch statement.
