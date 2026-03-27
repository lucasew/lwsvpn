# AGENTS.md - Project Conventions & Rules

## Core Tooling
- The project requires the use of `mise` for all task execution and tool management. Versions are explicitly pinned in `mise.toml`.
- Parent tasks like `install`, `test`, and `codegen` in `mise.toml` must only depend on wildcards (e.g., `["install:*"]`). Do not manually list specific subtasks.

## Architecture & Code Layout
- The project is a Go application relying on Go modules.
- **cmd/server -> Server entrypoint:** Hosts the server instance multiplexing traffic and serving logs.
- **cmd/client -> Client entrypoint:** Handles local proxy listening and connection routing back to the server.
- **pkg/errors -> Centralized error reporting:** All unexpected errors must be funneled through `errors.ReportError`.
- Ensure directory grouping is driven by domain and responsibility (e.g. `cmd/` for entrypoints, `pkg/` for shared libraries). Avoid sparse directories.

## Error Handling Conventions
- **No silent failures.** All error code paths must be reported.
- Unexpected errors must be handled via a centralized error-reporting function (`pkg/errors.ReportError`).
- Raw usage of `console.error`, standard `log.Printf` for errors, or unchecked `panic` without reporting is strictly prohibited.

## External Libraries
- The repository relies on `hashicorp/yamux` to multiplex multiple SOCKS5 streams over a single persistent WebSocket connection between the client and server.

## CI/CD
- The repository enforces exactly one GitHub Actions workflow file located at `.github/workflows/autorelease.yml`. It handles setup, codegen, PR creation for diffs, CI checks, and conditional release/artifacts.
