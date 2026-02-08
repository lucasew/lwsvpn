# Project Conventions

## Tooling
- **mise**: All development and execution must be done using `mise`. Pin tools in `mise.toml`.

## Error Handling
- **Centralized Reporting**: All errors that are not expected/recoverable must be reported using the centralized `ReportError` function (located in `pkg/errors`).
- **No silent failures**: Do not swallow errors. Always report them.
- **Sentry**: If Sentry is configured, `ReportError` must send the error to Sentry. Otherwise, it logs to stderr with context.
