# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test

```bash
make build      # Build binary to ./xbe with version info
make test       # Run all tests (go test ./...)
make fmt        # Format code (go fmt ./...)
make tidy       # Clean up dependencies (go mod tidy)
```

Run a single test:
```bash
go test -v ./internal/telemetry -run TestConfigFromEnv
```

Go version required: 1.25.6 (see go.mod)

## Architecture

XBE CLI is a Cobra-based command-line tool for the XBE platform.

### Directory Structure
- `cmd/xbe/main.go` - Entry point with telemetry initialization
- `internal/cli/` - All command implementations (~75 files)
- `internal/api/client.go` - HTTP client with JSON:API support
- `internal/auth/` - Token storage (keychain with file fallback)
- `internal/telemetry/` - OpenTelemetry instrumentation (optional)
- `internal/version/` - Version constant (injected at build)

### Command Groups
- `auth` - Authentication (login, status, logout, whoami)
- `view` - Read operations for various resources
- `do` - Write operations (create/update/delete)
- `summarize` - Data analysis commands

### Command Patterns

**List commands** (`*_list.go`):
- Options struct with flags for filtering, pagination, output format
- Factory function: `newXxxListCmd()`
- Execution function: `runXxxList()`

**Show commands** (`*_show.go`):
- Fetch single resource by ID
- Support both table and JSON output

**Create/Update/Delete** (`do_*.go`):
- Options struct with optional flags
- POST/PATCH/DELETE to API endpoints

### Adding New Commands

1. Create file in `internal/cli/` following naming convention
2. Use options struct pattern for flags
3. Create factory function `newXxxCmd()` returning `*cobra.Command`
4. Add to parent command in appropriate file (e.g., `view.go`, `do.go`)
5. Support `--json` flag for machine-readable output
6. Include `Use`, `Short`, `Long`, and `Example` in command definition

### Resource Decisions

See `RESOURCE_DECISIONS.md` for tracking which server resources to implement in the CLI.

When considering a new resource:
1. Check if it's already in the decisions file
2. If pending, discuss with the user before implementing
3. If implementing, move it to the "Implemented" section
4. If skipping, add it to "Skipped" with the reason and date
5. Check the server's policy file (`app/policies/*_policy.rb`) to see if the resource is read-only

Resources marked as `abstract` in the server are not real API endpoints and should be skipped.

### Flag Display in Help Output

Subcommand help shows only command-specific flags. Global flags (`--json`, `--limit`, `--offset`, `--sort`, `--base-url`, `--token`, `--no-auth`) are documented in `xbe --help` and referenced via a one-liner in subcommand help.

When adding new global flags, update the flag maps in `internal/cli/help.go`.

### API Client

The client in `internal/api/client.go` uses JSON:API format (`application/vnd.api+json`). Auth token resolution order: flag → env vars → keychain → file.

### Telemetry

Controlled by environment variables (`XBE_TELEMETRY_ENABLED`). Disabled by default. Set `XBE_TELEMETRY_ENABLED=0` to explicitly disable during testing.

## Shell

Use `bash` for commands (not zsh).

## Formatting

Run `gofmt -w` on modified Go files before committing.
