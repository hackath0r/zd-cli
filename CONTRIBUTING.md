# Contributing to zd-cli

Thanks for your interest in `zd-cli`. This project is small, so onboarding
is short. Please skim this document before opening a pull request.

## Quickstart

```sh
git clone https://github.com/hackath0r/zd-cli.git
cd zd-cli
go mod download
make build
./zd version
```

You will need:

- Go 1.22 or later (1.25+ recommended)
- `golangci-lint` for `make lint`
- `goreleaser` only if you want to test the release pipeline locally

## Project layout

```
api/                      vendored OpenAPI spec + codegen config
cmd/zd/                   cobra commands; one file per resource group
internal/zenduty/         hand-written wrapper + generated typed client
internal/{config,output,errors,version}
                          shared CLI plumbing
internal/tools/           one-shot Go programs used by the Makefile
scripts/                  install.sh, install.ps1
```

## Workflow

1. Fork and create a branch from `main`.
2. Make your change. Keep PRs focused; one feature or fix per PR.
3. If you touched the OpenAPI spec, run `make openapi-pull && make generate`
   so the generated client stays in sync.
4. Run the standard checks:
   ```sh
   make lint
   make test
   make build
   ```
5. Commit using [Conventional Commits](https://www.conventionalcommits.org/).
   Examples: `feat(incident): add --include-resolved`, `fix(retry): honour
   Retry-After when seconds is 0`. The release notes are generated from
   commit messages, so a clean history matters.
6. Push and open a PR. The `ci` workflow runs lint, tests on
   linux/macOS/windows, and verifies the generated client is up to date.

## Adding a new command

1. Find the right file under `cmd/zd/`. If the command is for an existing
   resource group (incident, event, oncall, etc.), append to that file.
   New top-level groups get a new file plus a registration line in
   `cmd/zd/root.go`.
2. Implement the command using the `callAPI` generic helper for clean
   error/context handling.
3. Provide a `*output.TableSpec` so `--output table` does the right thing.
   Tables use `text/tabwriter`.
4. Add a focused test (httptest server in `internal/zenduty/auth_test.go`
   shows the pattern). Aim for a useful assertion, not coverage padding.
5. If your command shells out to a new endpoint, glance at the generated
   code in `internal/zenduty/zenduty.gen.go` to confirm the request body
   shape and JSON200 vs JSON201 wrapper. Some upstream endpoints declare
   list responses as single objects; use `decodeList[T]` / `decodeOne[T]`
   to work around those.

## Reporting bugs

Please open an issue with:

- The command you ran (redact your token)
- Expected vs actual behaviour
- `zd version` output
- Output of `zd <command> --debug` if relevant

Security vulnerabilities should go to `SECURITY.md`, not public issues.

## Code style

- `gofmt` + `goimports` (run `make fmt`).
- Errors propagate as `*internal/errors.ExitError` so the exit code
  contract works.
- Comments explain *why*, not what. Avoid restating the next line in a
  comment.
- Keep cobra command help readable; long descriptions belong in `Long:`,
  one-liners in `Short:`.

Thanks again. Even a typo fix is a contribution.
