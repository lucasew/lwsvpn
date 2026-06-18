## IGNORE: Moving Entrypoints to cmd/ Directory

**- Pattern:** Moving or renaming the main entrypoints (`app.go` to `cmd/server/main.go` and `client/client.go` to `cmd/client/main.go`).
**- Justification:** The project specifically maintains `app.go` as the server entrypoint and `client/client.go` as the client entrypoint. They should not be moved to a `cmd/` structure.
**- Files Affected:** `app.go`, `client/client.go`, `cmd/server/main.go`, `cmd/client/main.go`

## IGNORE: Inadequate Context in Centralized Error Reporting

**- Pattern:** Implementing the centralized error reporting function (e.g., `pkg/errors.ReportError`) using basic logging (like `log.Printf`) without capturing stack traces or relevant metadata.
**- Justification:** Centralized error reporters must log errors with enough context (message, stack trace, relevant metadata) to support debugging and Sentry-like observability.
**- Files Affected:** `pkg/errors/errors.go`

## IGNORE: Downgrading Dependencies

**- Pattern:** Downgrading GitHub Actions dependencies, such as using `actions/checkout@v4` instead of a newer pinned version, or using `jdx/mise-action@v2`.
**- Justification:** The project's global directives explicitly forbid downgrading dependencies unless explicitly requested.
**- Files Affected:** `.github/workflows/autorelease.yml`

## IGNORE: Scope Creep and Unrelated Bundling

**- Pattern:** Bundling repository-wide global setup or convention files (e.g., `AGENTS.md`, `mise.toml`, `.github/workflows/autorelease.yml`) into PRs meant for specific bug fixes or refactors.
**- Justification:** PRs must maintain strict scope discipline and execute only the explicitly requested outcome without bundling global or unrelated structural changes.
**- Files Affected:** `AGENTS.md`, `mise.toml`, `.github/workflows/autorelease.yml`
