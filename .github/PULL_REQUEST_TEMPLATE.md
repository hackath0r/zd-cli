<!--
Thanks for opening a PR! Please:
- Use a Conventional Commits-style title (e.g. feat(incident): support --include-resolved)
- Confirm the checks below before requesting review
-->

## What

Briefly describe the change.

## Why

Link the issue or describe the user-visible problem this solves.

## How

A short summary of the implementation, especially for non-trivial changes.

## Checklist

- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] If the OpenAPI spec changed, I ran `make generate` and committed the result
- [ ] If a public CLI flag changed, I updated `docs/COMMANDS.md` (or it will be regenerated)
- [ ] Commit messages follow Conventional Commits
