# Project Conventions & Rules

## Project Structure & Location
- `app.go` -> Main server application entrypoint.
- `client/client.go` -> Client application entrypoint.
- `pkg/errors/errors.go` -> Centralized error reporting function.
- `.github/workflows/autorelease.yml` -> CI/CD GitHub Actions workflow.
- `mise.toml` -> Tools definition and tasks runner.

## Error Handling Rule
- Always check for errors before deferencing pointers, particularly network connections.
- The project MUST have a single, centralized error-reporting function (`pkg/errors.ReportError`). All code paths that handle unexpected errors MUST funnel through this function. Never call `console.error`, `Sentry.captureException` or raw `log` output for unexpected errors directly at the call site unless there is a specific reason to not use the centralized one.
- No silent failures: Every error that is not an expected/recoverable condition MUST be reported via the centralized error reporting function.
- "Out of scope" is not an excuse to swallow an error.

## General
- The project is a Go application and relies on Go modules for dependency management.
- Always install tools pinning exactly their versions in `mise.toml`.
- Avoid wildcards in `mise.toml` subtasks.
