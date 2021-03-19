# slog

[![Go Reference](https://pkg.go.dev/badge/github.com/intuitivelabs/slog.svg)](https://pkg.go.dev/github.com/intuitivelabs/slog)

The slog package provides a simple, basic, fast leveled logger.

It supports logging to stderr and stdout, various levels, caller location
 information (file, line) and backtraces.

It tries to minimize memory allocations by using an internal lock-less 
pool of buffers.

It does not use any locking so log messages might get "mixed" between the
threads.

Note that enabling file location or backtracing will cause a significant
slowdown and lot of memory allocations.

