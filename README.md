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
│   ├── whoami              Show the current authenticated user
│   └── logout              Remove stored token
├── do                      Create, update, and delete XBE resources
│   ├── application-settings Manage global application settings
│   │   ├── create           Create an application setting
│   │   ├── update           Update an application setting
│   │   └── delete           Delete an application setting
│   ├── glossary-terms       Manage glossary terms
│   │   ├── create           Create a glossary term
│   │   ├── update           Update a glossary term
│   │   └── delete           Delete a glossary term
│   ├── platform-statuses    Manage platform status updates
│   │   ├── create           Create a platform status
│   │   ├── update           Update a platform status
│   │   └── delete           Delete a platform status
│   ├── lane-summary         Generate lane (cycle) summaries
│   │   └── create           Create a lane summary
│   ├── material-transaction-summary  Generate material transaction summaries
│   │   └── create           Create a material transaction summary
│   ├── memberships          Manage user-organization memberships
│   │   ├── create           Create a membership
│   │   ├── update           Update a membership
│   │   └── delete           Delete a membership
│   ├── equipment-movement-stop-requirements  Manage equipment movement stop requirements
│   │   ├── create           Create a stop requirement
│   │   └── delete           Delete a stop requirement
│   └── equipment-movement-trip-dispatches  Manage equipment movement trip dispatches
│       ├── create           Create a trip dispatch
│       ├── update           Update a trip dispatch
│       └── delete           Delete a trip dispatch
├── view                    Browse and view XBE content
│   ├── application-settings Browse application settings
│   │   ├── list            List application settings
│   │   └── show <id>       Show application setting details
│   ├── newsletters         Browse and view newsletters
│   │   ├── list            List newsletters with filtering
│   │   └── show <id>       Show newsletter details
│   ├── posts               Browse and view posts
│   │   ├── list            List posts with filtering
│   │   └── show <id>       Show post details
│   ├── brokers             Browse broker/branch information
│   │   └── list            List brokers with filtering
│   ├── users               Browse users (for creator lookup)
│   │   └── list            List users with filtering
│   ├── material-suppliers  Browse material suppliers
│   │   └── list            List suppliers with filtering
│   ├── customers           Browse customers
│   │   └── list            List customers with filtering
│   ├── truckers            Browse trucking companies
│   │   └── list            List truckers with filtering
│   ├── memberships         Browse user-organization memberships
│   │   ├── list            List memberships with filtering
│   │   └── show <id>       Show membership details
│   ├── features            Browse product features
│   │   ├── list            List features with filtering
│   │   └── show <id>       Show feature details
│   ├── release-notes       Browse release notes
│   │   ├── list            List release notes with filtering
│   │   └── show <id>       Show release note details
│   ├── press-releases      Browse press releases
│   │   ├── list            List press releases
│   │   └── show <id>       Show press release details
│   ├── platform-statuses   Browse platform status updates
│   │   ├── list            List platform statuses
│   │   └── show <id>       Show platform status details
│   ├── glossary-terms      Browse glossary terms
│   │   ├── list            List glossary terms with filtering
│   │   └── show <id>       Show glossary term details
│   ├── equipment-movement-stop-requirements  Browse equipment movement stop requirements
│   │   ├── list            List stop requirements
│   │   └── show <id>       Show stop requirement details
│   └── equipment-movement-trip-dispatches  Browse equipment movement trip dispatches
│       ├── list            List trip dispatches
│       └── show <id>       Show trip dispatch details
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
xbe auth status   # Check if a token is configured
xbe auth whoami   # Verify token and show current user
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

### Posts

```bash
# List recent posts
xbe view posts list

# Filter by status
xbe view posts list --status published

# Filter by post type
xbe view posts list --post-type basic

# Filter by date range
xbe view posts list --published-at-min 2024-01-01 --published-at-max 2024-06-30

# Filter by creator
xbe view posts list --creator "User|123"

# View full post content
xbe view posts show 789

# Get JSON output for scripting
xbe view posts list --json --limit 10
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

### Users, Material Suppliers, Customers, Truckers

Use these commands to look up IDs for filtering posts by creator.

```bash
# Find a user ID
xbe view users list --name "John"

# Find a material supplier ID
xbe view material-suppliers list --name "Acme"

# Find a customer ID
xbe view customers list --name "Smith"

# Find a trucker ID
xbe view truckers list --name "Express"

# Then filter posts by that creator
xbe view posts list --creator "User|123"
xbe view posts list --creator "MaterialSupplier|456"
xbe view posts list --creator "Customer|789"
xbe view posts list --creator "Trucker|101"
```

### Features, Release Notes, Press Releases, Glossary Terms

```bash
# List product features
xbe view features list
xbe view features list --pdca-stage plan
xbe view features show 123

# List release notes
xbe view release-notes list
xbe view release-notes list --q "trucking"
xbe view release-notes show 456

# List press releases
xbe view press-releases list
xbe view press-releases show 789

# List glossary terms
xbe view glossary-terms list
xbe view glossary-terms show 101
```

### Lane Summary (Cycle Summary)

```bash
# Generate a lane summary grouped by origin/destination
xbe do lane-summary create \
  --group-by origin,destination \
  --filter broker=123 \
  --filter transaction_at_min=2025-01-11T00:00:00Z \
  --filter transaction_at_max=2025-01-17T23:59:59Z

# Focus on pickup/delivery dwell minutes for higher-volume lanes
xbe do lane-summary create \
  --group-by origin,destination \
  --filter broker=123 \
  --filter transaction_at_min=2025-01-11T00:00:00Z \
  --filter transaction_at_max=2025-01-17T23:59:59Z \
  --min-transactions 25 \
  --metrics material_transaction_count,pickup_dwell_minutes_mean,pickup_dwell_minutes_median,pickup_dwell_minutes_p90,delivery_dwell_minutes_mean,delivery_dwell_minutes_median,delivery_dwell_minutes_p90

# Include effective cost per hour
xbe do lane-summary create \
  --group-by origin,destination \
  --filter broker=123 \
  --filter transaction_at_min=2025-01-11T00:00:00Z \
  --filter transaction_at_max=2025-01-17T23:59:59Z \
  --min-transactions 25 \
  --metrics material_transaction_count,delivery_dwell_minutes_median,effective_cost_per_hour_median
```

### Material Transaction Summary

```bash
# Summary grouped by material site
xbe do material-transaction-summary create \
  --filter broker=123 \
  --filter date_min=2025-01-01 \
  --filter date_max=2025-01-31

# Summary by customer segment (internal/external)
xbe do material-transaction-summary create \
  --group-by customer_segment \
  --filter broker=123

# Summary by month and material type
xbe do material-transaction-summary create \
  --group-by month,material_type_fully_qualified_name_base \
  --filter broker=123 \
  --filter material_type_fully_qualified_name_base="Asphalt Mixture" \
  --sort month:asc

# Summary by direction (inbound/outbound)
xbe do material-transaction-summary create \
  --group-by direction \
  --filter broker=123

# Summary with all metrics
xbe do material-transaction-summary create \
  --group-by material_site \
  --filter broker=123 \
  --all-metrics

# High-volume results only
xbe do material-transaction-summary create \
  --filter broker=123 \
  --min-transactions 100
```

### Memberships

Memberships define the relationship between users and organizations (brokers, customers, truckers, material suppliers, developers).

```bash
# List your memberships
xbe view memberships list --user 1

# List memberships for a broker
xbe view memberships list --broker 123

# Search by user name
xbe view memberships list --q "John"

# Filter by role
xbe view memberships list --kind manager
xbe view memberships list --kind operations

# Show full membership details
xbe view memberships show 456

# Create a membership (organization format: Type|ID)
xbe do memberships create --user 123 --organization "Broker|456" --kind manager

# Create with additional attributes
xbe do memberships create \
  --user 123 \
  --organization "Broker|456" \
  --kind manager \
  --title "Regional Manager" \
  --is-admin true

# Update a membership
xbe do memberships update 789 --kind operations --title "Driver"

# Update permissions
xbe do memberships update 789 \
  --can-see-rates-as-manager true \
  --is-rate-editor true

# Delete a membership (requires --confirm)
xbe do memberships delete 789 --confirm
```

### Crew Assignment Confirmations

Crew assignment confirmations record when a resource confirms a crew requirement assignment.

```bash
# List confirmations
xbe view crew-assignment-confirmations list

# Filter by crew requirement
xbe view crew-assignment-confirmations list --crew-requirement 123

# Filter by resource
xbe view crew-assignment-confirmations list --resource-type laborers --resource-id 456

# Show confirmation details
xbe view crew-assignment-confirmations show 789

# Confirm using assignment confirmation UUID
xbe do crew-assignment-confirmations create \
  --assignment-confirmation-uuid "uuid-here" \
  --note "Confirmed" \
  --is-explicit

# Confirm using crew requirement + resource + start time
xbe do crew-assignment-confirmations create \
  --crew-requirement 123 \
  --resource-type laborers \
  --resource-id 456 \
  --start-at "2025-01-01T08:00:00Z"

# Update a confirmation
xbe do crew-assignment-confirmations update 789 --note "Updated note" --is-explicit true
```

### Crew Rates

Crew rates define pricing for labor/equipment by classification, resource, or craft class.

```bash
# List crew rates for a broker
xbe view crew-rates list --broker 123 --is-active true

# Filter by resource classification
xbe view crew-rates list --resource-classification-type LaborClassification --resource-classification-id 456

# Create a crew rate
xbe do crew-rates create --price-per-unit 75.00 --start-on 2025-01-01 --is-active true \
  --broker 123 --resource-classification-type LaborClassification --resource-classification-id 456

# Update a crew rate
xbe do crew-rates update 789 --price-per-unit 80.00 --end-on 2025-12-31

# Delete a crew rate (requires --confirm)
xbe do crew-rates delete 789 --confirm
```

### Driver Day Trips Adjustments

Driver day trips adjustments track edits to a driver's trip sequence for a shift.

```bash
# List adjustments
xbe view driver-day-trips-adjustments list --status editing

# Create an adjustment
xbe do driver-day-trips-adjustments create \
  --tender-job-schedule-shift 123 \
  --old-trips-attributes '[{"note":"Original trip"}]' \
  --description "Adjust trip order" \
  --status editing

# Update an adjustment
xbe do driver-day-trips-adjustments update 456 \
  --new-trips-attributes '[{"note":"Updated trip"}]' \
  --description "Updated adjustment"

# Delete an adjustment
xbe do driver-day-trips-adjustments delete 456 --confirm
```

### Driver Movement Observations

Driver movement observations summarize movement cycles for a job production plan.

```bash
# List observations
xbe view driver-movement-observations list

# Filter by job production plan
xbe view driver-movement-observations list --plan 123

# Only current observations
xbe view driver-movement-observations list --is-current

# Show observation details
xbe view driver-movement-observations show 456
```

### Crew Requirement Credential Classifications

Crew requirement credential classifications link crew requirements to the credential
classifications they require.

```bash
# List crew requirement credential classifications
xbe view crew-requirement-credential-classifications list

# Filter by crew requirement
xbe view crew-requirement-credential-classifications list --crew-requirement 123

# Show link details
xbe view crew-requirement-credential-classifications show 456

# Create a link
xbe do crew-requirement-credential-classifications create \
  --crew-requirement-type labor-requirements \
  --crew-requirement 123 \
  --credential-classification-type user-credential-classifications \
  --credential-classification 456

# Delete a link (requires --confirm)
xbe do crew-requirement-credential-classifications delete 456 --confirm
```

### Equipment Movement Stop Requirements

Equipment movement stop requirements link equipment movement stops to movement requirements.

```bash
# List stop requirements
xbe view equipment-movement-stop-requirements list

# Filter by stop
xbe view equipment-movement-stop-requirements list --stop 123

# Filter by requirement
xbe view equipment-movement-stop-requirements list --requirement 456

# Filter by kind
xbe view equipment-movement-stop-requirements list --kind origin

# Show stop requirement details
xbe view equipment-movement-stop-requirements show 789

# Create a stop requirement
xbe do equipment-movement-stop-requirements create --stop 123 --requirement 456

# Create with explicit kind
xbe do equipment-movement-stop-requirements create --stop 123 --requirement 456 --kind destination

# Delete a stop requirement (requires --confirm)
xbe do equipment-movement-stop-requirements delete 789 --confirm
```

### Equipment Movement Trip Dispatches

Equipment movement trip dispatches orchestrate creation and assignment of movement trips.

```bash
# List trip dispatches
xbe view equipment-movement-trip-dispatches list

# Filter by status
xbe view equipment-movement-trip-dispatches list --status pending

# Filter by trip
xbe view equipment-movement-trip-dispatches list --equipment-movement-trip 123

# Show trip dispatch details
xbe view equipment-movement-trip-dispatches show 456

# Create from an existing trip
xbe do equipment-movement-trip-dispatches create --equipment-movement-trip 123

# Create from a movement requirement
xbe do equipment-movement-trip-dispatches create --equipment-movement-requirement 456

# Update assignments
xbe do equipment-movement-trip-dispatches update 456 --driver 789 --trailer 321

# Delete a trip dispatch (requires --confirm)
xbe do equipment-movement-trip-dispatches delete 456 --confirm
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

### Pre-requs
```bash
#OSX
brew install go

# Debian/Ubuntu
sudo apt update && sudo apt install golang-go

# Fedora
sudo dnf install golang 

# Windows - Chocolatey
choco install golang

# Windows - Scoop
scoop install go
```

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
