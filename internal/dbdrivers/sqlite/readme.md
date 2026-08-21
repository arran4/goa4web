# internal/dbdrivers/sqlite

## Purpose

Package `sqlite` implements the `dbdrivers.DBDriver` interface for SQLite using `modernc.org/sqlite`.

## Why It Exists

To provide an opt-in, CGo-free SQLite database driver suitable for local development, fast automated testing, and CI environments without requiring an external MySQL server.

## Features

- Pure Go (CGo-free via `modernc.org/sqlite`).
- Implements `dbdrivers.DBDriver` for `"sqlite3"` and `"sqlite"`.
- Support for Backup and Restore via `sqlite3` CLI tool.
- Opt-in via build tags `sqlite` or `sqlite3`.
