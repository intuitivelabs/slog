# slog

[![Go Reference](https://pkg.go.dev/badge/github.com/intuitivelabs/slog.svg)](https://pkg.go.dev/github.com/intuitivelabs/slog)

The slog package provides a simple, basic, fast leveled logger.

It supports logging to stderr and stdout, various levels, caller location
 information (file, line) and backtraces.

It does not use any locking so log messages might get "mixed" between the
threads.
