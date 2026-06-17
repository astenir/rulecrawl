# RuleCrawl

RuleCrawl is a small rule-based crawler framework prototype in Go. It provides:

- task-based crawling rules
- concurrent request scheduling with priority support
- optional rate limiting
- HTTP fetching with cookie, proxy, timeout, random user-agent, and charset handling
- parser output through `spider.DataCell`
- MySQL storage with automatic table creation and batched inserts
- in-memory storage for tests and local demos
- a small go-micro/gRPC gateway demo service

## Requirements

- Go `1.23.3`, as declared in `go.mod`
- MySQL, when SQL storage is enabled
- etcd, when running the go-micro service registry in `main.go`

## Quick Start

Edit `config.toml` for your local MySQL, proxy, task, and server settings.

Run checks:

```bash
go test ./...
go build ./...
```

Run with the default local config:

```bash
go run .
```

Run with an explicit config file:

```bash
go run . -config ./config.toml
```

or:

```bash
CRAWLER_CONFIG=./config.toml go run .
```

## Configuration

The application reads `./config.toml` by default. The path can be overridden by:

1. `-config /path/to/config.toml`
2. `CRAWLER_CONFIG=/path/to/config.toml`

Important sections:

- `logLevel`: zap log level, for example `debug` or `info`
- `Tasks`: seed task settings; `Name` must match a registered task
- `fetcher.timeout`: request timeout in milliseconds
- `fetcher.proxy`: optional proxy list; leave empty to fetch directly
- `storage.sqlURL`: MySQL DSN
- `GRPCServer`: go-micro and HTTP gateway addresses

## Example Tasks

The engine package does not register site-specific tasks by itself. The application entrypoint registers the current Douban examples through `examples/douban`:

- `douban_book_list`
- Douban group examples
- Douban group JavaScript rule example

To add a Go task, define a `*spider.Task` with:

- `Options.Name`
- `Rule.Root`
- `Rule.Trunk`
- `Rule.ItemFields` for fields that should be stored
- `Rule.ParseFunc` for parsing fetched pages

Then register it with `engine.Store.Add`.

## Notes

This repository is still a prototype. Before deploying it beyond local use, review:

- database credentials in `config.toml`
- proxy availability
- SQL table/column names generated from task metadata
- final flush behavior for batched storage on shutdown
- crawler-specific response validation, such as the current minimum body length check in the engine
