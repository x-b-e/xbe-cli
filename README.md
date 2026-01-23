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
│   ├── driver-day-adjustment-plans Manage driver day adjustment plans
│   │   ├── create           Create a driver day adjustment plan
│   │   ├── update           Update a driver day adjustment plan
│   │   └── delete           Delete a driver day adjustment plan
│   ├── driver-day-shortfall-calculations Calculate driver day shortfall allocations
│   │   └── create           Create a driver day shortfall calculation
│   ├── shift-counters       Count accepted shifts
│   │   └── create           Create a shift counter
│   ├── inventory-estimates  Manage inventory estimates
│   │   ├── create           Create an inventory estimate
│   │   ├── update           Update an inventory estimate
│   │   └── delete           Delete an inventory estimate
│   ├── job-production-plan-approvals Approve job production plans
│   │   └── create           Approve a job production plan
│   ├── job-production-plan-uncompletions Uncomplete job production plans
│   │   └── create           Uncomplete a job production plan
│   ├── job-schedule-shift-start-site-changes Manage job schedule shift start site changes
│   │   └── create           Create a job schedule shift start site change
│   ├── site-events          Manage site events
│   │   ├── create           Create a site event
│   │   ├── update           Update a site event
│   │   └── delete           Delete a site event
│   ├── service-sites       Manage service sites
│   │   ├── create           Create a service site
│   │   ├── update           Update a service site
│   │   └── delete           Delete a service site
│   ├── lineup-dispatches    Manage lineup dispatches
│   │   ├── create           Create a lineup dispatch
│   │   ├── update           Update a lineup dispatch
│   │   └── delete           Delete a lineup dispatch
│   ├── lineup-scenario-generators Generate lineup scenarios
│   │   ├── create           Create a lineup scenario generator
│   │   └── delete           Delete a lineup scenario generator
│   ├── lineup-scenarios     Manage lineup scenarios
│   │   ├── create           Create a lineup scenario
│   │   ├── update           Update a lineup scenario
│   │   └── delete           Delete a lineup scenario
│   ├── job-production-plan-material-types Manage job production plan material types
│   │   ├── create           Create a job production plan material type
│   │   ├── update           Update a job production plan material type
│   │   └── delete           Delete a job production plan material type
│   ├── job-production-plan-time-card-approvers Manage job production plan time card approvers
│   │   ├── create           Create a job production plan time card approver
│   │   └── delete           Delete a job production plan time card approver
│   ├── time-card-pre-approvals Manage time card pre-approvals
│   │   ├── create           Create a time card pre-approval
│   │   ├── update           Update a time card pre-approval
│   │   └── delete           Delete a time card pre-approval
│   ├── time-card-unscrappages Manage time card unscrappages
│   │   └── create           Create a time card unscrappage
│   ├── time-sheets           Manage time sheets
│   │   ├── create           Create a time sheet
│   │   ├── update           Update a time sheet
│   │   └── delete           Delete a time sheet
│   ├── time-sheet-rejections Manage time sheet rejections
│   │   └── create           Reject a time sheet
│   ├── time-sheet-line-item-equipment-requirements Manage time sheet line item equipment requirements
│   │   ├── create           Create a time sheet line item equipment requirement
│   │   ├── update           Update a time sheet line item equipment requirement
│   │   └── delete           Delete a time sheet line item equipment requirement
│   ├── job-production-plan-service-type-unit-of-measure-cohorts Manage job production plan service type unit of measure cohorts
│   │   ├── create           Create a job production plan service type unit of measure cohort link
│   │   └── delete           Delete a job production plan service type unit of measure cohort link
│   ├── maintenance-requirement-sets Manage maintenance requirement sets
│   │   ├── create           Create a maintenance requirement set
│   │   ├── update           Update a maintenance requirement set
│   │   └── delete           Delete a maintenance requirement set
│   ├── maintenance-requirement-maintenance-requirement-parts Manage maintenance requirement parts
│   │   ├── create           Create a maintenance requirement part link
│   │   ├── update           Update a maintenance requirement part link
│   │   └── delete           Delete a maintenance requirement part link
│   ├── lane-summary         Generate lane (cycle) summaries
│   │   └── create           Create a lane summary
│   ├── material-transaction-summary  Generate material transaction summaries
│   │   └── create           Create a material transaction summary
│   ├── material-transaction-inspections Manage material transaction inspections
│   │   ├── create           Create a material transaction inspection
│   │   ├── update           Update a material transaction inspection
│   │   └── delete           Delete a material transaction inspection
│   ├── material-type-unavailabilities Manage material type unavailabilities
│   │   ├── create           Create a material type unavailability
│   │   ├── update           Update a material type unavailability
│   │   └── delete           Delete a material type unavailability
│   ├── material-supplier-memberships Manage material supplier memberships
│   │   ├── create           Create a material supplier membership
│   │   ├── update           Update a material supplier membership
│   │   └── delete           Delete a material supplier membership
│   └── memberships          Manage user-organization memberships
│       ├── create           Create a membership
│       ├── update           Update a membership
│       └── delete           Delete a membership
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
│   ├── material-supplier-memberships Browse material supplier memberships
│   │   ├── list            List material supplier memberships with filtering
│   │   └── show <id>       Show material supplier membership details
│   ├── material-site-measures Browse material site measures
│   │   ├── list            List material site measures with filtering
│   │   └── show <id>       Show material site measure details
│   ├── material-site-mixing-lots Browse material site mixing lots
│   │   ├── list            List material site mixing lots with filtering
│   │   └── show <id>       Show material site mixing lot details
│   ├── material-type-unavailabilities Browse material type unavailabilities
│   │   ├── list            List material type unavailabilities with filtering
│   │   └── show <id>       Show material type unavailability details
│   ├── material-transaction-inspections Browse material transaction inspections
│   │   ├── list            List material transaction inspections with filtering
│   │   └── show <id>       Show material transaction inspection details
│   ├── inventory-estimates Browse inventory estimates
│   │   ├── list            List inventory estimates with filtering
│   │   └── show <id>       Show inventory estimate details
│   ├── service-sites       Browse service sites
│   │   ├── list            List service sites with filtering
│   │   └── show <id>       Show service site details
│   ├── customers           Browse customers
│   │   └── list            List customers with filtering
│   ├── truckers            Browse trucking companies
│   │   └── list            List truckers with filtering
│   ├── driver-day-adjustment-plans Browse driver day adjustment plans
│   │   ├── list            List driver day adjustment plans
│   │   └── show <id>       Show driver day adjustment plan details
│   ├── driver-day-shortfall-calculations Browse driver day shortfall calculations
│   │   ├── list            List driver day shortfall calculations
│   │   └── show <id>       Show driver day shortfall calculation details
│   ├── shift-counters      Browse shift counters
│   │   └── list            List shift counters
│   ├── job-production-plan-duplication-works Browse job production plan duplication work
│   │   ├── list            List duplication work with filtering
│   │   └── show <id>       Show duplication work details
│   ├── job-production-plan-material-types Browse job production plan material types
│   │   ├── list            List job production plan material types with filtering
│   │   └── show <id>       Show job production plan material type details
│   ├── job-production-plan-service-type-unit-of-measure-cohorts Browse job production plan service type unit of measure cohorts
│   │   ├── list            List job production plan service type unit of measure cohorts with filtering
│   │   └── show <id>       Show job production plan service type unit of measure cohort details
│   ├── job-production-plan-time-card-approvers Browse job production plan time card approvers
│   │   ├── list            List job production plan time card approvers with filtering
│   │   └── show <id>       Show job production plan time card approver details
│   ├── time-card-pre-approvals Browse time card pre-approvals
│   │   ├── list            List time card pre-approvals with filtering
│   │   └── show <id>       Show time card pre-approval details
│   ├── time-card-unscrappages Browse time card unscrappages
│   │   ├── list            List time card unscrappages
│   │   └── show <id>       Show time card unscrappage details
│   ├── time-sheets         Browse time sheets
│   │   ├── list            List time sheets with filtering
│   │   └── show <id>       Show time sheet details
│   ├── time-sheet-rejections Browse time sheet rejections
│   │   ├── list            List time sheet rejections
│   │   └── show <id>       Show time sheet rejection details
│   ├── time-sheet-line-item-equipment-requirements Browse time sheet line item equipment requirements
│   │   ├── list            List time sheet line item equipment requirements with filtering
│   │   └── show <id>       Show time sheet line item equipment requirement details
│   ├── maintenance-requirement-sets Browse maintenance requirement sets
│   │   ├── list            List maintenance requirement sets with filtering
│   │   └── show <id>       Show maintenance requirement set details
│   ├── maintenance-requirement-maintenance-requirement-parts Browse maintenance requirement parts
│   │   ├── list            List maintenance requirement parts with filtering
│   │   └── show <id>       Show maintenance requirement part details
│   ├── job-schedule-shift-start-site-changes Browse job schedule shift start site changes
│   │   ├── list            List job schedule shift start site changes
│   │   └── show <id>       Show job schedule shift start site change details
│   ├── lineup-dispatches   Browse lineup dispatches
│   │   ├── list            List lineup dispatches with filtering
│   │   └── show <id>       Show lineup dispatch details
│   ├── lineup-scenario-generators Browse lineup scenario generators
│   │   ├── list            List lineup scenario generators with filtering
│   │   └── show <id>       Show lineup scenario generator details
│   ├── lineup-scenarios    Browse lineup scenarios
│   │   ├── list            List lineup scenarios with filtering
│   │   └── show <id>       Show lineup scenario details
│   ├── lineup-dispatch-shifts Browse lineup dispatch shifts
│   │   ├── list            List lineup dispatch shifts with filtering
│   │   └── show <id>       Show lineup dispatch shift details
│   ├── hos-events          Browse hours-of-service (HOS) events
│   │   ├── list            List HOS events with filtering
│   │   └── show <id>       Show HOS event details
│   ├── site-events         Browse site events
│   │   ├── list            List site events with filtering
│   │   └── show <id>       Show site event details
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
│   └── glossary-terms      Browse glossary terms
│       ├── list            List glossary terms with filtering
│       └── show <id>       Show glossary term details
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

### Material Supplier Memberships

Material supplier memberships link users to material suppliers and control role settings and notifications.

```bash
# List material supplier memberships
xbe view material-supplier-memberships list

# Filter by material supplier
xbe view material-supplier-memberships list --material-supplier 123

# Show membership details
xbe view material-supplier-memberships show 456

# Create a membership
xbe do material-supplier-memberships create --user 123 --material-supplier 456

# Update a membership
xbe do material-supplier-memberships update 456 --kind manager --title "Operations Manager"

# Update notifications
xbe do material-supplier-memberships update 456 --enable-recap-notifications true

# Delete a membership (requires --confirm)
xbe do material-supplier-memberships delete 456 --confirm
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

### Equipment Movement Requirement Locations

Equipment movement requirement locations define named coordinates used as origins and destinations.

```bash
# List locations
xbe view equipment-movement-requirement-locations list

# Filter by broker
xbe view equipment-movement-requirement-locations list --broker 123

# Show a location
xbe view equipment-movement-requirement-locations show 456

# Create a location
xbe do equipment-movement-requirement-locations create \
  --broker 123 \
  --latitude 37.7749 \
  --longitude -122.4194 \
  --name "Main Yard"

# Update a location
xbe do equipment-movement-requirement-locations update 456 --name "Updated Yard"

# Delete a location (requires --confirm)
xbe do equipment-movement-requirement-locations delete 456 --confirm
```

### Equipment Movement Stops

Equipment movement stops define the ordered locations within an equipment movement trip.

```bash
# List stops
xbe view equipment-movement-stops list

# Filter by trip
xbe view equipment-movement-stops list --trip 123

# Show a stop
xbe view equipment-movement-stops show 456

# Create a stop
xbe do equipment-movement-stops create \
  --trip 123 \
  --location 456 \
  --sequence-position 1 \
  --scheduled-arrival-at "2025-01-01T08:00:00Z"

# Update a stop
xbe do equipment-movement-stops update 456 --sequence-position 2

# Delete a stop (requires --confirm)
xbe do equipment-movement-stops delete 456 --confirm
```

### Equipment Utilization Readings

Equipment utilization readings capture odometer and hourmeter values for equipment.

```bash
# List readings
xbe view equipment-utilization-readings list

# Filter by equipment
xbe view equipment-utilization-readings list --equipment 123

# Filter by reported-at range
xbe view equipment-utilization-readings list --reported-at-min 2025-01-01T00:00:00Z --reported-at-max 2025-01-31T23:59:59Z

# Show a reading
xbe view equipment-utilization-readings show 456

# Create a reading
xbe do equipment-utilization-readings create --equipment 123 --reported-at 2025-01-01T08:00:00Z --odometer 100

# Update a reading
xbe do equipment-utilization-readings update 456 --hourmeter 12

# Delete a reading (requires --confirm)
xbe do equipment-utilization-readings delete 456 --confirm
```

### Driver Day Adjustment Plans

Driver day adjustment plans define per-trucker adjustments applied to driver day recaps.

```bash
# List plans for a trucker
xbe view driver-day-adjustment-plans list --trucker 123

# Show a plan
xbe view driver-day-adjustment-plans show 456

# Create a plan
xbe do driver-day-adjustment-plans create --trucker 123 --content "Adjusted start time" \
  --start-at "2025-01-15T08:00:00Z"

# Update a plan
xbe do driver-day-adjustment-plans update 456 --content "Updated plan" \
  --start-at "2025-01-16T06:00:00Z"

# Delete a plan (requires --confirm)
xbe do driver-day-adjustment-plans delete 456 --confirm
```

### Driver Day Shortfall Calculations

Driver day shortfall calculations allocate shortfall quantities across time cards and constraints.

```bash
# Create a calculation
xbe do driver-day-shortfall-calculations create \
  --time-card-ids 101,102 \
  --driver-day-time-card-constraint-ids 55,56

# Exclude time cards from allocation
xbe do driver-day-shortfall-calculations create \
  --time-card-ids 101,102,103 \
  --unallocatable-time-card-ids 103 \
  --driver-day-time-card-constraint-ids 55,56

# Show a calculation (when available)
xbe view driver-day-shortfall-calculations show <id>
```

### Shift Counters

Shift counters return the number of accepted tender job schedule shifts after a minimum start time.

```bash
# Count accepted shifts (default start)
xbe do shift-counters create

# Count accepted shifts after a date
xbe do shift-counters create --start-at-min 2025-01-01T00:00:00Z

# List counters (typically empty)
xbe view shift-counters list
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
