# Contributing

Thanks for contributing to `go-robokassa-sdk`.

## Prerequisites

- Go `1.25+`

## Local setup

```bash
go mod download
```

## Validation commands

Run tests:

```bash
go test ./...
```

Run cyclomatic complexity check (same policy as CI):

```bash
go run github.com/fzipp/gocyclo/cmd/gocyclo@latest -over 12 -ignore '.*_test\.go$' .
```

This repository enforces complexity **<= 12** for production Go files (`*_test.go` excluded).

## Code style and design expectations

- Preserve public API behavior unless a change is explicitly requested.
- Prefer small, focused functions.
- If a method starts combining multiple logical concerns (validation, normalization, mapping, transport), extract helper functions.
- Reuse existing helpers before adding new logic paths.
- Keep error messages and validation semantics stable when refactoring.

## Documentation expectations

If you add or change behavior in Invoice API, Payment Interface, XML interfaces, or Refund API flows:

1. Update the relevant file in `docs/`.
2. Update `README.md` links/overview if discoverability changes.

## Pull request checklist

1. Code is formatted (`gofmt`).
2. Tests pass (`go test ./...`).
3. Complexity check passes.
4. Documentation is updated when behavior or usage changes.
