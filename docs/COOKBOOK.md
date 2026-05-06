# zd-cli Cookbook

Worked examples for common workflows. Every snippet assumes you have a
profile configured (`zd config init`) and a token in
`ZENDUTY_API_TOKEN`.

## 1. Triage an open incident from the command line

```sh
# Find the highest-urgency open incidents
zd incident list --status open | jq '[.[] | select(.urgency=="high")] | sort_by(.creation_date) | .[0]'

# Acknowledge it
zd incident ack 4815

# Add a timeline note
zd incident note add 4815 -m "investigating; rolling back deploy abc123"

# Page a teammate
zd incident responder add-user 4815 --user grace.hopper

# Resolve once the rollback completes
zd incident resolve 4815
```

## 2. Idempotent alert from CI

The `--entity-id` flag is the deduplication key. Two events with the
same `entity_id` reference the same alert (and incident) on Zenduty's
side, so retries from a flaky CI step do not produce duplicate
incidents.

```sh
zd event fire \
  --integration-key "$ZD_INTEGRATION_KEY" \
  --message "build $CI_BUILD_ID failed" \
  --summary "main branch is red" \
  --entity-id "build-$CI_BUILD_ID" \
  --url "https://ci.example.com/build/$CI_BUILD_ID" \
  --wait
```

`--wait` blocks until Zenduty's ingestion reaches a terminal state
(`completed` or `failed`), which is useful when you want CI logs to
record the actual outcome rather than just the trace ID.

## 3. Auto-resolve when a remediation completes

Pair `--entity-id` with `zd event resolve`:

```sh
zd event resolve \
  --integration-key "$ZD_INTEGRATION_KEY" \
  --message "deploy abc123 healthy for 5min" \
  --entity-id "build-$CI_BUILD_ID"
```

## 4. Quickly find who is on call

```sh
# Default profile, default team
zd oncall now

# A different team, only primaries
zd oncall who --team 03f7b1c2-...

# Full roster including delayed escalations
zd oncall list --team 03f7b1c2-... --output table
```

## 5. Multi-environment workflows

```yaml
# ~/.config/zd/config.yaml
default_profile: prod
profiles:
  prod:
    host: https://www.zenduty.com
    token_env: ZENDUTY_API_TOKEN
    account_id: ABCDE
    default_team: 03f7b1c2-...
  staging:
    host: https://www.zenduty.com
    token_env: ZENDUTY_STAGING_TOKEN
    account_id: VWXYZ
```

```sh
# Run a command against staging
zd --profile staging incident list

# Switch the default profile
zd config use staging
```

## 6. Custom output template

Render a list of incident numbers + titles, one per line:

```sh
zd incident list --status open \
  --output template \
  --template '{{range .}}{{.incident_number}}\t{{.title}}\n{{end}}'
```

## 7. Drive zd from a Cursor / AI skill

A skill that triages alerts can chain `zd-cli` with `jq` because the
pipe-to-stdout default is JSON:

```sh
# Acknowledge anything assigned to me older than 10min
zd incident list --user "$ZD_USERNAME" --status triggered \
  | jq -r --arg now "$(date -u +%s)" \
      '.[] | select((($now|tonumber) - (.creation_date | sub("\\..*$"; "Z") | fromdate)) > 600) | .incident_number' \
  | xargs -I{} zd incident ack {}
```

The exit-code contract (0 success, 1 API, 2 usage, 3 config, 4 network)
is documented in the README so a skill can branch on `?` cleanly.

## 8. Watch a fired alert until it terminates

```sh
trace=$(zd event fire --integration-key K --message "test" --output json | jq -r '.[0].trace_id')
zd event status "$trace" --watch
```

## 9. ximr aliasing

Whether you typed `zd incident list` or `ximr incident list`, the
binary is identical and produces the same output. The `--debug` flag
even shows you which name was invoked.

## 10. Find a service unique_id (until v0.2 ships service commands)

Most CRUDs land in v0.2 (see the README roadmap). Until then, use
`curl` + the typed JSON output:

```sh
curl -fsSL -H "Authorization: Token $ZENDUTY_API_TOKEN" \
  "https://www.zenduty.com/api/account/teams/$TEAM_ID/services/" \
  | jq '.[] | {unique_id, name}'
```

Track the v0.2 milestone here:
https://github.com/hackath0r/zd-cli/issues?q=label%3Aapi-coverage
