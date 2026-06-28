# Contributing

Thanks for contributing to `terraform-provider-mtncloud`.

## Before you start

Read [ARCHITECTURE.md](ARCHITECTURE.md). It defines the layering (provider →
client → API), the shared helpers you must reuse, and the conventions every
resource follows. PRs that re-implement boilerplate already in `configure.go`,
`conversions.go`, `diagnostics.go`, or `client/query.go` will be asked to use them.

## Adding a resource or data source

Follow the **"Adding a resource — checklist"** in ARCHITECTURE.md. In short:

1. Client method(s) in `internal/client/<area>.go` using the `query.go` generics.
2. `resource_<x>.go` embedding `resourceBase`; `data_source_<x>.go` embedding
   `dataSourceBase`.
3. Register the `New*` constructors in `internal/provider/provider.go`.
4. Add `examples/` `.tf`, then `make docs`.
5. Add a client unit test, an `*_acc_test.go` (model it on
   `resource_credential_acc_test.go`), and a sweeper entry in `sweep_test.go`.

## Local workflow

```sh
make check      # build + vet + golangci-lint + unit tests + gofmt — must pass
make docs       # regenerate docs/ from schema + examples (CI fails if stale)
make testacc    # acceptance tests — needs MTN_CLOUD_TOKEN (tokens expire ~hourly)
make install-local   # install the provider for manual terraform testing
```

## Conventions that reviewers check

- New resources reuse the shared seams (`resourceBase`, `parseID`,
  `handleReadError`, `opError`, the client generics).
- Write-only secret fields are `Sensitive` and never read back from the API.
- Optional+Computed fields use `mergeAPI*` to avoid perpetual diffs.
- Immutable fields use `RequiresReplace()` and say so in their description.
- `docs/` is regenerated and committed; `make check` is green.

## Commit & PR

- Keep PRs focused; one resource (or one concern) per PR where practical.
- CI must be green: build, vet, lint, unit tests, gofmt, and the docs-drift check.
