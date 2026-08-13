# `gen-permutations` command

This developer executable prints `goa4web gen-og-image` commands for every
supported pattern, color pair, and RPG-theme combination. It is a package `main`,
not an importable library.

Run from the repository root:

```sh
go run ./cmd/gen-permutations > /tmp/generate-og-images.sh
# Inspect the generated commands before choosing whether to execute them.
```

The command creates `examples/og-images` and prints commands; it does not execute
the image generator itself. Generated image sets can be large, so avoid committing
them accidentally.
