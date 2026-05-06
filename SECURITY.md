# Security policy

## Supported versions

`zd-cli` is pre-1.0; only the latest minor release receives security fixes.
After 1.0 we intend to support the latest two minor releases.

## Reporting a vulnerability

Please report security issues privately rather than opening a public issue.

Use GitHub's
[Security Advisories](https://github.com/hackath0r/zd-cli/security/advisories/new)
flow ("Report a vulnerability"). I will acknowledge within 5 business days
and aim to ship a fix within 30 days for high-severity issues.

If GitHub Security Advisories isn't an option, email
`hackath0r@users.noreply.github.com` with the subject `zd-cli security`.

## Secrets handling

`zd-cli` reads API tokens from:

1. `--token` flag
2. `ZENDUTY_API_TOKEN` env var (or the per-profile `token_env`)
3. `~/.config/zd/config.yaml` (file mode `0600`)

The `zd config show` command redacts inline tokens before printing.

Tokens are never logged at the default log level. With `--debug-curl`
enabled the rendered curl line will contain the token, so do not paste
that output into bug reports without redacting it.

## Supply chain

- Releases are built reproducibly via [GoReleaser](https://goreleaser.com).
- Archives, deb/rpm packages, and a `checksums.txt` are published to
  GitHub Releases for every tagged release.
- `cosign` keyless signing is enabled in the release pipeline so each
  artifact can be verified.
- The OpenAPI spec is vendored at `api/openapi.yaml`; any drift is
  caught by the weekly `openapi-sync` workflow.
