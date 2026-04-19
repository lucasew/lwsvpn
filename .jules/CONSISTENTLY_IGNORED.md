## IGNORE: Downgrading Dependencies

**- Pattern:** Downgrading GitHub Action dependencies, such as using `actions/checkout@v4` or `jdx/mise-action@v2`.
**- Justification:** The project global directives explicitly prohibit downgrading dependencies unless explicitly requested.
**- Files Affected:** `.github/workflows/autorelease.yml`

## IGNORE: Missing Context in Centralized Error Reporter

**- Pattern:** Implementing the centralized error reporting function (e.g., `pkg/errors.ReportError`) with basic logging (like `log.Printf`) without capturing stack traces and relevant metadata.
**- Justification:** The project requires that the centralized error reporter log errors with enough context (message, stack, relevant metadata) to be useful, as per the Sentry-aware instructions.
**- Files Affected:** `pkg/errors/errors.go`

## IGNORE: Manually Listing Subtasks in mise.toml

**- Pattern:** Manually listing specific subtasks in the `depends` array of parent tasks (e.g., `depends = ["lint", "test"]`) in `mise.toml`.
**- Justification:** Project rules strictly require that parent tasks like `install`, `test`, and `codegen` must only depend on wildcards (e.g., `["install:*"]`).
**- Files Affected:** `mise.toml`

## IGNORE: Scope Creep and Bundling

**- Pattern:** Bundling repository-wide structural or configuration file creations (e.g., `AGENTS.md`, `.github/workflows/autorelease.yml`, `mise.toml`, `.jules/sentinel.md`) into PRs meant for specific bugs or refactoring.
**- Justification:** PRs must maintain strict scope discipline and execute only the explicitly requested outcome without bundling unrelated changes.
**- Files Affected:** `AGENTS.md`, `.github/workflows/autorelease.yml`, `mise.toml`, `.jules/sentinel.md`

## IGNORE: Hallucinated Diff and Title Mismatch

**- Pattern:** Submitting a PR where the diff does not align with the PR title (e.g., claiming to "Fix ignored errors" but only modifying a `Dockerfile`).
**- Justification:** Changes must actually implement what is promised in the title and description, without hallucinating unrelated modifications.
**- Files Affected:** `Dockerfile`, `app.go`
