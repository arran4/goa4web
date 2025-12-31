# Command Line Interface Structure

The `goa4web` binary exposes a tree of commands implemented under
`cmd/goa4web/`. A small wrapper in `main.go` creates a `rootCmd` which
parses global flags before dispatching to a subcommand.

Global flags are defined by `config.NewRuntimeFlagSet` and include
`--config-file`, `--verbosity` and all configuration options documented in
[Configuration](configuration.md). Values are resolved in the order
command line, configuration file and environment variables.

Each subcommand has its own `parse<Name>Cmd` function returning a struct
with a `Run()` method. The `help` command relies on `helpCmd.showHelp` to
instantiate each command with the `-h` flag so every subcommand can print
its own usage text. New commands follow the same pattern: create a
`<name>Cmd` type with a `FlagSet`, parse any arguments and implement
`Run()` to perform the action.

Subcommands may themselves dispatch to further nested commands (for
example `user` and `config`). All of them share the same configuration
mechanism via the embedded `rootCmd`. Running `goa4web help` or
`goa4web help <command>` displays the relevant usage information.

A special `repl` command starts an interactive shell. The REPL accepts the
same commands as the normal CLI and supports background execution by
appending `&`, external commands prefixed with `!` and simple environment
variables using `set KEY=VALUE`. Type `jobs` to list running background
commands and `wait <id>` to wait for completion.
