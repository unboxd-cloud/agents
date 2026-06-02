# Handover — extracting the community platform into `unboxd-cloud/agent-research`

**Status:** the community-development platform is **built, tested, and committed in
`platform`** (PR #1). Extraction into a separate repo (`unboxd-cloud/agent-research`)
is **blocked by session scope** and awaits a session that can write to that repo.
This document is the durable handover so the move is mechanical.

## Why it isn't already moved (the blocker, proven)

This session is hard-scoped to `unboxd-cloud/platform` at the infrastructure layer —
two independent walls, both confirmed empirically:

- **GitHub MCP tools:** every call to `agent-research` returns
  `Access denied: repository "unboxd-cloud/agent-research" is not configured for this
  session. Allowed repositories: unboxd-cloud/platform`. Repo creation returns
  `403 Resource not accessible by integration`.
- **Git proxy:** pushing/fetching `…/git/unboxd-cloud/agent-research` returns
  `remote: Proxy error: repository not authorized` (HTTP 502).
- **No scope tool:** there is no `list_repos` / `add_repo` in this session, so scope
  cannot be expanded from inside.

To unblock: add `agent-research` to the **session's repository scope** (the
web/app environment configuration that lists accessible repos — *not* GitHub repo
permissions), then run a session on that environment.

## Decisions locked (from the working session)

- **Repo:** `unboxd-cloud/agent-research` (name mirrors AGenNext/Agent-Research).
- **Scope of move:** *move* the engine out of `platform` (not copy) — `platform`'s
  core control plane must stay uncoupled and not overloaded.
- **Keep `platform` core untouched:** the community surface must not be baked into
  the platform's service build; the orchestrator (`platform.agent`, kernel) is not
  touched by it.

## What to move (already in `platform`, all stdlib-only, tests green)

| from `platform` | to `agent-research` |
|---|---|
| `internal/community/community.go` + `community_test.go` | `internal/community/` |
| `cmd/community/main.go` | `cmd/community/` (rewrite imports, see below) |
| `internal/ui/ui.go` | `internal/ui/` (copy; shared design-system helper) |
| `internal/server/server.go` | `internal/server/` (copy; shared HTTP helper) |
| `metamodels/community.agent` | `model/community.agent` |

The standalone module was scaffolded and **verified compiling + vetting + passing
tests** during the session (`go.mod` module `github.com/unboxd-cloud/agent-research`,
go 1.24; plus `Makefile`, `README.md`, `.github/workflows/ci.yml` running
vet/test/build).

### Exact transplant steps (reproducible)

```sh
# 1. In the agent-research checkout, lay down the tree:
mkdir -p cmd/community internal/community internal/ui internal/server model .github/workflows

# 2. Copy from a platform checkout (P=path to platform):
cp $P/internal/community/community.go      internal/community/
cp $P/internal/community/community_test.go internal/community/
cp $P/internal/ui/ui.go                    internal/ui/
cp $P/internal/server/server.go            internal/server/
cp $P/metamodels/community.agent           model/community.agent

# 3. Rewrite the command's import paths platform -> agent-research:
sed 's#unboxd-cloud/platform/internal#unboxd-cloud/agent-research/internal#g' \
    $P/cmd/community/main.go > cmd/community/main.go

# 4. Add go.mod (module github.com/unboxd-cloud/agent-research, go 1.24),
#    Makefile, README.md, .github/workflows/ci.yml (vet+test+build).

# 5. Verify:
go vet ./... && go test ./... -count=1 && go build ./...
```

`go.mod`, `Makefile`, `README.md`, and `ci.yml` contents are reproducible from the
session; if not preserved, regenerate them (single-module, no external deps).

## After the move — clean up `platform`

Once `agent-research` holds and builds the code, remove it from `platform` so the
two don't drift:

```sh
git rm -r internal/community cmd/community metamodels/community.agent
# remove `community` from the SERVICES list in Makefile
make model   # regenerate docs/datamodel.json without the community declarations
```

`internal/ui` and `internal/server` **stay** in `platform` (used by admin/orgconsole).

## Open follow-ups discussed (not built — awaiting explicit go)

- **Provenance catalog:** proposed-by / last-validated-by / last-published-by /
  last-sourced-by, surfaced as a theory/idea catalog (3 of 4 fields already in the
  engine's data; needs a `source`/`grounded-in` field + a read view).
- **Sustaining vs disruptive** idea-type → routes disruptive ideas to a sandboxed
  innovation-lab track; plus innovation metrics (adoption, time-to-release, hit rate).
- **MLCommons** as an external conformance authority in the eval/benchmark thread
  (separate from the community funnel).
- **Daily conformance workflow** (Tier 1): scheduled `make agents` + `go test` +
  `make bench` + `platform agent bench platform.agent`, publishing dated results.
