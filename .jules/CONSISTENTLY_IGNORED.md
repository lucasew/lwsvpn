## IGNORE: Downgrading Dependencies

**- Pattern:** Downgrading GitHub Action dependencies, such as using `actions/checkout@v4` or `jdx/mise-action@v2`.
**- Justification:** The project global directives explicitly prohibit downgrading dependencies unless explicitly requested. PRs consistently fail when `v4` and `v2` are used instead of the expected versions.
**- Files Affected:** `.github/workflows/autorelease.yml`

## IGNORE: Scope Creep and Bundling Global Setup Files

**- Pattern:** Modifying or creating repository-wide structural or configuration files (e.g., `AGENTS.md`, `.github/workflows/autorelease.yml`, `mise.toml`, `.jules/sentinel.md`) in PRs meant for specific functional bugs or refactoring, or performing actions outside of the main goal.
**- Justification:** PRs should adhere strictly to their intended execution scope. Including unrelated global files or out-of-scope changes violates boundaries, making PRs noisy and causing them to be rejected.
**- Files Affected:** `AGENTS.md`, `.github/workflows/autorelease.yml`, `mise.toml`, `.jules/sentinel.md`
