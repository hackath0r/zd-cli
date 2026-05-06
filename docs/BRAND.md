# Brand & host migration notes

## Zenduty -> Xurrent IMR

Zenduty was acquired by [Xurrent](https://www.xurrent.com) in February 2025
and rebranded to **Xurrent Incident Management & Response** (IMR) in
November 2025. The product, the REST API, and the CLI surface all stayed
the same; only the legal entity and the marketing language changed.

`zd-cli` reflects this transition in two places:

1. The same binary is shipped as both `zd` and `ximr`. Pick whichever
   reads better in your shell history. `ximr` is the new
   "Xurrent IMR" name.
2. The default API host is still `https://www.zenduty.com` because the
   public REST API has not (yet) been migrated to a `xurrent.com`
   subdomain. When that migration ships, `zd-cli` will release a minor
   version that switches the default and keeps the old host as an
   override (`--host` flag, profile field, or `ZENDUTY_HOST` env var).

## What stays the same

- Authentication: `Authorization: Token <api_key>` (NOT `Bearer`)
- Endpoint paths: `/api/...`, `/api/v2/...`, `/integration/{account_id}/generic/{key}/`
- Status code mapping: `1=triggered`, `2=acknowledged`, `3=resolved`
- OpenAPI spec location: `https://apidocs.zenduty.com/openapi.json`

## Trademarks

"Zenduty" and "Xurrent" are trademarks of their respective owners. This
project is not affiliated with or endorsed by Zenduty or Xurrent. Use of
the names is solely for describing what this software interoperates with.
