## IGNORE: Inadequate Context in Centralized Error Reporting

**- Pattern:** Implementing the centralized error reporting function (e.g., `pkg/errors.ReportError`) using only basic logging (like `log.Printf`) without capturing stack traces or relevant metadata.
**- Justification:** Global instructions mandate that centralized error reporters log errors with enough context (message, stack, relevant metadata) to support debugging and Sentry-like observability.
**- Files Affected:** `pkg/errors/errors.go`

## IGNORE: Downgrading CI Action Dependencies

**- Pattern:** Creating or modifying GitHub Actions workflows with downgraded dependency versions, specifically `actions/checkout@v4` and `jdx/mise-action@v2`.
**- Justification:** The project's global directives explicitly forbid downgrading dependencies (such as actions/checkout to v4 or mise action to v2) unless explicitly requested.
**- Files Affected:** `.github/workflows/autorelease.yml`

## IGNORE: Hallucinating Project Conventions

**- Pattern:** Inventing arbitrary or contradictory project rules (e.g., dictating whether to use wildcards or explicit subtasks in `mise.toml`) and attempting to enforce them by creating or modifying `AGENTS.md`.
**- Justification:** Global directives prohibit hallucinating conventions. Agents must not fabricate rules to populate convention files; rules must be derived from actual, proven repository standards.
**- Files Affected:** `AGENTS.md`

## IGNORE: Scope Creep and Unrelated Bundling

**- Pattern:** Bundling "nice to have" formatting changes (e.g., converting spaces to tabs), unrelated file restructurings, or global configuration additions into PRs designated for targeted bug fixes or refactors.
**- Justification:** Global instructions require strict scope discipline and small, atomic, reversible diffs. PRs must execute only the explicitly requested outcome.
**- Files Affected:** `client/client.go`, `app.go`, `AGENTS.md`
