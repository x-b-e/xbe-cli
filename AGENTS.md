# AGENTS.md

## Shell
- Use `bash` for commands (not zsh).

## Go
- Go version: 1.22+
- Run `gofmt -w` on modified Go files.
- Run `go test ./...` before committing changes.

## CLI conventions
- Use Cobra for commands.
- Default output is human-readable; `--json` outputs simplified JSON for scripting.
- Keep commands discoverable via `--help` at each level.
- Avoid adding TUI/auth flows unless explicitly requested.

## Flag Display in Help Output
Subcommand help shows only command-specific flags. Global flags (`--json`, `--limit`, `--offset`, `--sort`, `--base-url`, `--token`, `--no-auth`) are documented in `xbe --help` and referenced with a one-liner.

Flag categorization is defined in `internal/cli/help.go`. When adding new global flags, update the flag maps there.

## Structure
- Entrypoint: `cmd/xbe/main.go`
- Commands: `internal/cli`
- Version: `internal/version`

## Build
- `make build` should produce `./xbe`.
