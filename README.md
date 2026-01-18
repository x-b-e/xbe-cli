# XBE CLI

A command-line interface for the [XBE platform](https://www.x-b-e.com), providing programmatic access to newsletters, broker data, and platform services. Designed for both interactive use and automation by AI agents.

## What is XBE?

XBE is a business operations platform for the heavy materials, logistics, and construction industries. It provides end-to-end visibility from quarry to customer, managing materials (asphalt, concrete, aggregates), logistics coordination, and construction operations. The XBE CLI lets you access platform data programmatically.

## Quick Start

```bash
# 1. Install
curl -fsSL https://raw.githubusercontent.com/x-b-e/xbe-cli/main/scripts/install.sh | bash

# 2. Authenticate
xbe auth login

# 3. Browse newsletters
xbe view newsletters list

# 4. View a specific newsletter
xbe view newsletters show <id>
```

## Installation

### macOS and Linux

```bash
curl -fsSL https://raw.githubusercontent.com/x-b-e/xbe-cli/main/scripts/install.sh | bash
```

Installs to `/usr/local/bin` if writable, otherwise `~/.local/bin`.

To specify a custom location:

```bash
INSTALL_DIR=/usr/local/bin USE_SUDO=1 curl -fsSL https://raw.githubusercontent.com/x-b-e/xbe-cli/main/scripts/install.sh | bash
```

### Windows

Download the latest release from [GitHub Releases](https://github.com/x-b-e/xbe-cli/releases), extract `xbe.exe`, and add it to your PATH.

### Updating

```bash
xbe update
```

## Command Reference

```
xbe
├── auth                    Manage authentication credentials
│   ├── login               Store an access token
│   ├── status              Show authentication status
│   └── logout              Remove stored token
├── view                    Browse and view XBE content
│   ├── newsletters         Browse and view newsletters
│   │   ├── list            List newsletters with filtering
│   │   └── show <id>       Show newsletter details
│   └── brokers             Browse broker/branch information
│       └── list            List brokers with filtering
├── update                  Show update instructions
└── version                 Print the CLI version
```

Run `xbe --help` for comprehensive documentation, or `xbe <command> --help` for details on any command.

## Authentication

### Getting a Token

Create an API token in the XBE client: https://client.x-b-e.com/#/browse/users/me/api-tokens

### Storing Your Token

```bash
# Interactive (secure prompt, recommended)
xbe auth login

# Via flag
xbe auth login --token "YOUR_TOKEN"

# Via stdin (for password managers)
op read "op://Vault/XBE/token" | xbe auth login --token-stdin
```

Tokens are stored securely in your system's credential storage:
- **macOS**: Keychain
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

Fallback: `~/.config/xbe/config.json`

### Token Resolution Order

1. `--token` flag
2. `XBE_TOKEN` or `XBE_API_TOKEN` environment variable
3. System keychain
4. Config file

### Managing Authentication

```bash
xbe auth status   # Check if authenticated
xbe auth logout   # Remove stored token
```

## Usage Examples

### Newsletters

```bash
# List recent published newsletters
xbe view newsletters list

# Search by keyword
xbe view newsletters list --q "market analysis"

# Filter by broker
xbe view newsletters list --broker-id 123

# Filter by date range
xbe view newsletters list --published-on-min 2024-01-01 --published-on-max 2024-06-30

# View full newsletter content
xbe view newsletters show 456

# Get JSON output for scripting
xbe view newsletters list --json --limit 10
```

### Brokers

```bash
# List all brokers
xbe view brokers list

# Search by company name
xbe view brokers list --company-name "Acme"

# Get broker ID for use in newsletter filtering
xbe view brokers list --company-name "Acme" --json | jq '.[0].id'
```

## Output Formats

All `list` and `show` commands support two output formats:

| Format | Flag | Use Case |
|--------|------|----------|
| Table | (default) | Human-readable, interactive use |
| JSON | `--json` | Scripting, automation, AI agents |

## Configuration

| Setting | Default | Override |
|---------|---------|----------|
| Base URL | `https://app.x-b-e.com` | `--base-url` or `XBE_BASE_URL` |
| Config directory | `~/.config/xbe` | `XDG_CONFIG_HOME` |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `XBE_TOKEN` | API access token |
| `XBE_API_TOKEN` | API access token (alternative) |
| `XBE_BASE_URL` | API base URL |
| `XDG_CONFIG_HOME` | Config directory (default: `~/.config`) |

## For AI Agents

This CLI is designed for AI agents. To have an agent use it:

1. Install the CLI (see above)
2. Authenticate (see above)
3. Tell the agent to run `xbe --help` to learn what the CLI can do

That's it. The `--help` output contains everything the agent needs: available commands, authentication details, configuration options, and examples. The agent can drill down with `xbe <command> --help` for specifics.

All commands support `--json` for structured output that's easy for agents to parse.

## Development

### Build

```bash
make build
```

### Run

```bash
./xbe --help
./xbe version
```

### Test

```bash
make test
```
