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
│   ├── driver-day-constraints Manage driver day constraints
│   │   ├── create           Create a driver day constraint
│   │   ├── update           Update a driver day constraint
│   │   └── delete           Delete a driver day constraint
│   ├── shift-set-time-card-constraints Manage shift set time card constraints
│   │   ├── create           Create a shift set time card constraint
│   │   ├── update           Update a shift set time card constraint
│   │   └── delete           Delete a shift set time card constraint
│   ├── time-card-cost-code-allocations Manage time card cost code allocations
│   │   ├── create           Create a time card cost code allocation
│   │   ├── update           Update a time card cost code allocation
│   │   └── delete           Delete a time card cost code allocation
│   ├── time-card-scrappages Scrap time cards
│   │   └── create           Scrap a time card
│   ├── time-card-unapprovals Unapprove time cards
│   │   └── create           Unapprove a time card
│   ├── resource-unavailabilities Manage resource unavailabilities
│   │   ├── create           Create a resource unavailability
│   │   ├── update           Update a resource unavailability
│   │   └── delete           Delete a resource unavailability
│   ├── equipment-movement-requirements Manage equipment movement requirements
│   │   ├── create           Create an equipment movement requirement
│   │   ├── update           Update an equipment movement requirement
│   │   └── delete           Delete an equipment movement requirement
│   ├── maintenance-requirement-rules Manage maintenance requirement rules
│   │   ├── create           Create a maintenance requirement rule
│   │   ├── update           Update a maintenance requirement rule
│   │   └── delete           Delete a maintenance requirement rule
│   ├── equipment-movement-trip-job-production-plans Manage equipment movement trip job production plans
│   │   ├── create           Create an equipment movement trip job production plan link
│   │   └── delete           Delete an equipment movement trip job production plan link
│   ├── labor-requirements   Manage labor requirements
│   │   ├── create           Create a labor requirement
│   │   ├── update           Update a labor requirement
│   │   └── delete           Delete a labor requirement
│   ├── job-production-plan-cancellations Cancel job production plans
│   │   └── create           Cancel a job production plan
│   ├── job-production-plan-rejections Reject job production plans
│   │   └── create           Reject a job production plan
│   ├── job-production-plan-unabandonments Unabandon job production plans
│   │   └── create           Unabandon a job production plan
│   ├── job-production-plan-recap-generations Generate job production plan recaps
│   │   └── create           Generate a job production plan recap
│   ├── job-production-plan-schedule-changes Apply schedule changes to job production plans
│   │   └── create           Apply a job production plan schedule change
│   ├── job-schedule-shift-splits Split job schedule shifts
│   │   └── create           Split a job schedule shift
│   ├── lineup-scenario-solutions Solve lineup scenarios
│   │   └── create           Create a lineup scenario solution
│   ├── lineup-summary-requests Request lineup summaries
│   │   └── create           Create a lineup summary request
│   ├── job-production-plan-display-unit-of-measures Manage job production plan display unit of measures
│   │   ├── create           Add a display unit of measure
│   │   ├── update           Update a display unit of measure
│   │   └── delete           Delete a display unit of measure
│   ├── job-production-plan-service-type-unit-of-measures Manage job production plan service type unit of measures
│   │   ├── create           Add a service type unit of measure
│   │   ├── update           Update a service type unit of measure
│   │   └── delete           Delete a service type unit of measure
│   ├── job-production-plan-trailer-classifications Manage job production plan trailer classifications
│   │   ├── create           Add a trailer classification
│   │   ├── update           Update a trailer classification
│   │   └── delete           Delete a trailer classification
│   ├── job-production-plan-locations Manage job production plan locations
│   │   ├── create           Create a job production plan location
│   │   ├── update           Update a job production plan location
│   │   └── delete           Delete a job production plan location
│   ├── geofence-restrictions Manage geofence restrictions
│   │   ├── create           Create a geofence restriction
│   │   ├── update           Update a geofence restriction
│   │   └── delete           Delete a geofence restriction
│   ├── lane-summary         Generate lane (cycle) summaries
│   │   └── create           Create a lane summary
│   ├── material-transaction-summary  Generate material transaction summaries
│   │   └── create           Create a material transaction summary
│   ├── material-transaction-field-scopes Manage material transaction field scopes
│   │   └── create           Create a material transaction field scope
│   ├── material-transaction-rejections Reject material transactions
│   │   └── create           Reject a material transaction
│   ├── material-transaction-submissions Submit material transactions
│   │   └── create           Submit a material transaction
│   ├── material-transaction-ticket-generators Manage material transaction ticket generators
│   │   ├── create           Create a material transaction ticket generator
│   │   ├── update           Update a material transaction ticket generator
│   │   └── delete           Delete a material transaction ticket generator
│   ├── inventory-changes   Manage inventory changes
│   │   ├── create           Create an inventory change
│   │   └── delete           Delete an inventory change
│   ├── material-site-inventory-locations Manage material site inventory locations
│   │   ├── create           Create a material site inventory location
│   │   ├── update           Update a material site inventory location
│   │   └── delete           Delete a material site inventory location
│   ├── material-site-subscriptions Manage material site subscriptions
│   │   ├── create           Create a material site subscription
│   │   ├── update           Update a material site subscription
│   │   └── delete           Delete a material site subscription
│   └── memberships          Manage user-organization memberships
│       ├── create           Create a membership
│       ├── update           Update a membership
│       └── delete           Delete a membership
├── view                    Browse and view XBE content
│   ├── application-settings Browse application settings
│   │   ├── list            List application settings
│   │   └── show <id>       Show application setting details
│   ├── driver-day-constraints Browse driver day constraints
│   │   ├── list            List driver day constraints
│   │   └── show <id>       Show driver day constraint details
│   ├── shift-set-time-card-constraints Browse shift set time card constraints
│   │   ├── list            List shift set time card constraints
│   │   └── show <id>       Show shift set time card constraint details
│   ├── time-card-cost-code-allocations Browse time card cost code allocations
│   │   ├── list            List time card cost code allocations
│   │   └── show <id>       Show time card cost code allocation details
│   ├── time-card-scrappages Browse time card scrappages
│   │   ├── list            List time card scrappages
│   │   └── show <id>       Show time card scrappage details
│   ├── time-card-unapprovals Browse time card unapprovals
│   │   ├── list            List time card unapprovals
│   │   └── show <id>       Show time card unapproval details
│   ├── resource-unavailabilities Browse resource unavailabilities
│   │   ├── list            List resource unavailabilities
│   │   └── show <id>       Show resource unavailability details
│   ├── driver-movement-segment-sets Browse driver movement segment sets
│   │   ├── list            List driver movement segment sets
│   │   └── show <id>       Show driver movement segment set details
│   ├── equipment-movement-requirements Browse equipment movement requirements
│   │   ├── list            List equipment movement requirements
│   │   └── show <id>       Show equipment movement requirement details
│   ├── maintenance-requirement-rules Browse maintenance requirement rules
│   │   ├── list            List maintenance requirement rules
│   │   └── show <id>       Show maintenance requirement rule details
│   ├── equipment-movement-trip-job-production-plans Browse equipment movement trip job production plans
│   │   ├── list            List equipment movement trip job production plans
│   │   └── show <id>       Show equipment movement trip job production plan details
│   ├── labor-requirements   Browse labor requirements
│   │   ├── list            List labor requirements
│   │   └── show <id>       Show labor requirement details
│   ├── job-production-plan-cancellations Browse job production plan cancellations
│   │   ├── list            List job production plan cancellations
│   │   └── show <id>       Show job production plan cancellation details
│   ├── job-production-plan-rejections Browse job production plan rejections
│   │   ├── list            List job production plan rejections
│   │   └── show <id>       Show job production plan rejection details
│   ├── job-production-plan-unabandonments Browse job production plan unabandonments
│   │   ├── list            List job production plan unabandonments
│   │   └── show <id>       Show job production plan unabandonment details
│   ├── job-production-plan-display-unit-of-measures Browse job production plan display unit of measures
│   │   ├── list            List job production plan display unit of measures
│   │   └── show <id>       Show job production plan display unit of measure details
│   ├── job-production-plan-service-type-unit-of-measures Browse job production plan service type unit of measures
│   │   ├── list            List job production plan service type unit of measures
│   │   └── show <id>       Show job production plan service type unit of measure details
│   ├── service-type-unit-of-measures Browse service type unit of measures
│   │   ├── list            List service type unit of measures
│   │   └── show <id>       Show service type unit of measure details
│   ├── job-production-plan-trailer-classifications Browse job production plan trailer classifications
│   │   ├── list            List job production plan trailer classifications
│   │   └── show <id>       Show job production plan trailer classification details
│   ├── job-production-plan-locations Browse job production plan locations
│   │   ├── list            List job production plan locations
│   │   └── show <id>       Show job production plan location details
│   ├── job-production-plan-job-site-location-estimates Browse job site location estimates
│   │   ├── list            List job site location estimates
│   │   └── show <id>       Show job site location estimate details
│   ├── job-schedule-shift-splits Browse job schedule shift splits
│   │   ├── list            List job schedule shift splits
│   │   └── show <id>       Show job schedule shift split details
│   ├── lineup-scenario-solutions Browse lineup scenario solutions
│   │   ├── list            List lineup scenario solutions
│   │   └── show <id>       Show lineup scenario solution details
│   ├── lineup-summary-requests Browse lineup summary requests
│   │   ├── list            List lineup summary requests
│   │   └── show <id>       Show lineup summary request details
│   ├── geofence-restrictions Browse geofence restrictions
│   │   ├── list            List geofence restrictions
│   │   └── show <id>       Show geofence restriction details
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
│   ├── material-transaction-field-scopes Browse material transaction field scopes
│   │   ├── list            List material transaction field scopes
│   │   └── show <id>       Show material transaction field scope details
│   ├── material-transaction-rejections Browse material transaction rejections
│   │   ├── list            List material transaction rejections
│   │   └── show <id>       Show material transaction rejection details
│   ├── material-transaction-submissions Browse material transaction submissions
│   │   ├── list            List material transaction submissions
│   │   └── show <id>       Show material transaction submission details
│   ├── material-transaction-ticket-generators Browse material transaction ticket generators
│   │   ├── list            List material transaction ticket generators
│   │   └── show <id>       Show material transaction ticket generator details
│   ├── material-site-inventory-locations Browse material site inventory locations
│   │   ├── list            List material site inventory locations
│   │   └── show <id>       Show material site inventory location details
│   ├── material-site-subscriptions Browse material site subscriptions
│   │   ├── list            List material site subscriptions
│   │   └── show <id>       Show material site subscription details
│   ├── inventory-changes   Browse and view inventory changes
│   │   ├── list            List inventory changes with filtering
│   │   └── show <id>       Show inventory change details
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

### Resource Unavailabilities

Resource unavailabilities define time ranges when a user, equipment, trailer, or tractor is unavailable.

```bash
# List resource unavailabilities
xbe view resource-unavailabilities list

# Filter by resource
xbe view resource-unavailabilities list --resource-type User --resource-id 123

# Filter by organization
xbe view resource-unavailabilities list --organization "Broker|456"

# Show unavailability details
xbe view resource-unavailabilities show 789

# Create a resource unavailability
xbe do resource-unavailabilities create \
  --resource-type User \
  --resource-id 123 \
  --start-at "2025-01-01T08:00:00Z" \
  --end-at "2025-01-01T17:00:00Z" \
  --description "PTO"

# Update an unavailability
xbe do resource-unavailabilities update 789 --end-at "2025-01-01T18:00:00Z"

# Delete an unavailability (requires --confirm)
xbe do resource-unavailabilities delete 789 --confirm
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

### Driver Assignment Rules

Driver assignment rules define constraints or guidance used when assigning drivers.

```bash
# List driver assignment rules
xbe view driver-assignment-rules list

# Filter by level
xbe view driver-assignment-rules list --level-type Broker --level-id 123

# Create a broker-level rule
xbe do driver-assignment-rules create \
  --rule "Drivers must be assigned by 6am" \
  --level-type Broker \
  --level-id 123 \
  --is-active

# Update a rule
xbe do driver-assignment-rules update 789 --rule "Updated rule" --is-active=false

# Delete a rule (requires --confirm)
xbe do driver-assignment-rules delete 789 --confirm
```

### Driver Movement Segment Sets

Driver movement segment sets summarize driver movement segments for a driver day.

```bash
# List driver movement segment sets
xbe view driver-movement-segment-sets list

# Filter by driver day
xbe view driver-movement-segment-sets list --driver-day 123

# Filter by driver
xbe view driver-movement-segment-sets list --driver 456

# Show full details
xbe view driver-movement-segment-sets show 789
```

### HOS Ruleset Assignments

HOS ruleset assignments track which HOS rule set applies to a driver at a given time.

```bash
# List HOS ruleset assignments
xbe view hos-ruleset-assignments list

# Filter by driver
xbe view hos-ruleset-assignments list --driver 123

# Filter by effective-at range
xbe view hos-ruleset-assignments list --effective-at-min 2025-01-01T00:00:00Z --effective-at-max 2025-01-31T23:59:59Z

# Show full details
xbe view hos-ruleset-assignments show 456
```

### Equipment Movement Requirements

Equipment movement requirements define equipment movement timing and locations.

```bash
# List requirements
xbe view equipment-movement-requirements list

# Filter by broker or equipment
xbe view equipment-movement-requirements list --broker 123
xbe view equipment-movement-requirements list --equipment 456

# Show details
xbe view equipment-movement-requirements show 789

# Create a requirement
xbe do equipment-movement-requirements create --broker 123 --equipment 456 \
  --origin-at-min "2025-01-01T08:00:00Z" --destination-at-max "2025-01-01T17:00:00Z" \
  --note "Move to yard"

# Update a requirement
xbe do equipment-movement-requirements update 789 --note "Updated note"

# Delete a requirement (requires --confirm)
xbe do equipment-movement-requirements delete 789 --confirm
```

### Maintenance Requirement Rules

Maintenance requirement rules define maintenance or inspection requirements for equipment,
equipment classifications, or business units.

```bash
# List rules
xbe view maintenance-requirement-rules list

# Filter by scope
xbe view maintenance-requirement-rules list --broker 111
xbe view maintenance-requirement-rules list --equipment 123
xbe view maintenance-requirement-rules list --equipment-classification 456
xbe view maintenance-requirement-rules list --business-unit 789
xbe view maintenance-requirement-rules list --is-active false

# Create a rule
xbe do maintenance-requirement-rules create \
  --rule "Service every 100 hours" \
  --broker 123 \
  --equipment-classification 456 \
  --is-active

# Update a rule
xbe do maintenance-requirement-rules update 789 --rule "Updated rule" --is-active=false

# Delete a rule (requires --confirm)
xbe do maintenance-requirement-rules delete 789 --confirm
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
