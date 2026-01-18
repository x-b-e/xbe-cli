# xbe-cli

The CLI for the XBE platform, designed to make it easy for agents to interact with XBE. We'll add capabilities bit by bit, leveraging the existing API and various client component designs; today it supports authentication plus newsletters and brokers.

## Quick start
1) Install the CLI (see below).
2) Store a token:

```
xbe auth login --token "YOUR_TOKEN"
```

3) List newsletters:

```
xbe view newsletters list
```

## Install (copy/paste)
macOS + Linux (downloads the latest release):

```
curl -fsSL https://raw.githubusercontent.com/x-b-e/xbe-cli/main/scripts/install.sh | bash
```

By default this installs to `/usr/local/bin` if writable; otherwise it installs to `~/.local/bin` and prints a PATH hint.
To override:

```
INSTALL_DIR=/usr/local/bin USE_SUDO=1 curl -fsSL https://raw.githubusercontent.com/x-b-e/xbe-cli/main/scripts/install.sh | bash
```

To update later, rerun the command above or use:

```
xbe update
```

Windows: download the zip from GitHub Releases, extract `xbe.exe`, and place it somewhere on your PATH.

## Authentication
The CLI reads tokens in this order:

1) `--token`
2) `XBE_TOKEN` or `XBE_API_TOKEN`
3) Stored token from `xbe auth login`

You can create an API token in the XBE client:

```
https://client.x-b-e.com/#/browse/users/me/api-tokens
```

Commands:

- `xbe auth login` stores a token (keychain if available, otherwise config file).
- `xbe auth status` shows whether a token is set for the current base URL.
- `xbe auth logout` removes the stored token.

You can also read a token from stdin:

```
cat token.txt | xbe auth login --token-stdin
```

## Configuration
- Base URL default: `https://server.x-b-e.com`
- Override with `--base-url` or `XBE_BASE_URL` / `XBE_API_BASE_URL`
- File token storage path: `~/.config/xbe/config.json` (respects `XDG_CONFIG_HOME`)

## Output formats
- Default output is human-readable tables or details.
- Add `--json` to `list` and `show` commands for machine-readable output.

## Examples
List published newsletters for a broker:

```
xbe view newsletters list --broker-id 123
```

Search newsletters:

```
xbe view newsletters list --q "interest rates"
```

List brokers as JSON:

```
xbe view brokers list --json
```

Show a newsletter as JSON:

```
xbe view newsletters show 42 --json
```

## Build
```
make build
```

## Run
```
./xbe --help
./xbe version
```
