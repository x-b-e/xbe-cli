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
│   ├── administrative-incidents Manage administrative incidents
│   │   ├── create           Create an administrative incident
│   │   ├── update           Update an administrative incident
│   │   └── delete           Delete an administrative incident
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
│   ├── job-production-plan-abandonments Abandon job production plans
│   │   └── create           Abandon a job production plan
│   ├── job-production-plan-completions Complete job production plans
│   │   └── create           Complete a job production plan
│   ├── job-production-plan-unscrappages Unscrap job production plans
│   │   └── create           Unscrap a job production plan
│   ├── project-abandonments Abandon projects
│   │   └── create           Abandon a project
│   ├── project-cancellations Cancel projects
│   │   └── create           Cancel a project
│   ├── project-completions Complete projects
│   │   └── create           Complete a project
│   ├── project-duplications Duplicate projects
│   │   └── create           Duplicate a project
│   ├── project-trailer-classifications Manage project trailer classifications
│   │   ├── create           Create a project trailer classification
│   │   ├── update           Update a project trailer classification
│   │   └── delete           Delete a project trailer classification
│   ├── project-transport-plan-assignment-rules Manage project transport plan assignment rules
│   │   ├── create           Create a project transport plan assignment rule
│   │   ├── update           Update a project transport plan assignment rule
│   │   └── delete           Delete a project transport plan assignment rule
│   ├── project-transport-plan-event-times Manage project transport plan event times
│   │   ├── create           Create a project transport plan event time
│   │   ├── update           Update a project transport plan event time
│   │   └── delete           Delete a project transport plan event time
│   ├── project-transport-plan-stops Manage project transport plan stops
│   │   ├── create           Create a project transport plan stop
│   │   ├── update           Update a project transport plan stop
│   │   └── delete           Delete a project transport plan stop
│   ├── project-transport-plan-tractors Manage project transport plan tractors
│   │   ├── create           Create a project transport plan tractor
│   │   ├── update           Update a project transport plan tractor
│   │   └── delete           Delete a project transport plan tractor
│   ├── project-transport-plan-segment-tractors Manage project transport plan segment tractors
│   │   ├── create           Create a project transport plan segment tractor
│   │   └── delete           Delete a project transport plan segment tractor
│   ├── project-margin-matrices Manage project margin matrices
│   │   ├── create           Create a project margin matrix
│   │   └── delete           Delete a project margin matrix
│   ├── project-phase-cost-items Manage project phase cost items
│   │   ├── create           Create a project phase cost item
│   │   ├── update           Update a project phase cost item
│   │   └── delete           Delete a project phase cost item
│   ├── project-phase-revenue-items Manage project phase revenue items
│   │   ├── create           Create a project phase revenue item
│   │   ├── update           Update a project phase revenue item
│   │   └── delete           Delete a project phase revenue item
│   ├── project-phase-cost-item-price-estimates Manage project phase cost item price estimates
│   │   ├── create           Create a project phase cost item price estimate
│   │   ├── update           Update a project phase cost item price estimate
│   │   └── delete           Delete a project phase cost item price estimate
│   ├── job-production-plan-driver-movements Generate job production plan driver movements
│   │   └── create           Generate driver movement details
│   ├── job-production-plan-job-site-changes Update job production plan job sites
│   │   └── create           Create a job site change
│   ├── job-production-plan-segments Manage job production plan segments
│   │   ├── create           Create a job production plan segment
│   │   ├── update           Update a job production plan segment
│   │   └── delete           Delete a job production plan segment
│   ├── job-production-plan-project-phase-revenue-items Manage job production plan project phase revenue items
│   │   ├── create           Create a job production plan project phase revenue item
│   │   ├── update           Update a job production plan project phase revenue item
│   │   └── delete           Delete a job production plan project phase revenue item
│   ├── job-schedule-shift-start-at-changes Reschedule job schedule shifts
│   │   └── create           Create a start-at change
│   ├── invoice-addresses    Address rejected invoices
│   │   └── create           Address a rejected invoice
│   ├── invoice-rejections   Reject sent invoices
│   │   └── create           Reject a sent invoice
│   ├── invoice-revisionables Mark invoices as revisionable
│   │   └── create           Mark an invoice as revisionable
│   ├── invoice-revisionizings Revise invoices
│   │   └── create           Revise an invoice
│   ├── time-card-time-changes Manage time card time changes
│   │   ├── create           Create a time card time change
│   │   ├── update           Update a time card time change
│   │   └── delete           Delete a time card time change
│   ├── time-sheet-line-items Manage time sheet line items
│   │   ├── create           Create a time sheet line item
│   │   ├── update           Update a time sheet line item
│   │   └── delete           Delete a time sheet line item
│   ├── lineup-job-schedule-shifts Manage lineup job schedule shifts
│   │   ├── create           Create a lineup job schedule shift
│   │   ├── update           Update a lineup job schedule shift
│   │   └── delete           Delete a lineup job schedule shift
│   ├── lineup-scenario-trailer-lineup-job-schedule-shifts Manage lineup scenario trailer lineup job schedule shifts
│   │   ├── create           Create a lineup scenario trailer lineup job schedule shift
│   │   ├── update           Update a lineup scenario trailer lineup job schedule shift
│   │   └── delete           Delete a lineup scenario trailer lineup job schedule shift
│   ├── shift-scope-tenders  Find tenders for a shift scope
│   │   └── create           Find tenders for a shift scope
│   ├── tender-returns       Return tenders
│   │   └── create           Return a tender
│   ├── driver-day-shortfall-allocations Allocate driver day shortfall quantities
│   │   └── create           Allocate driver day shortfall quantities
│   ├── material-transaction-summary  Generate material transaction summaries
│   │   └── create           Create a material transaction summary
│   ├── material-transaction-cost-code-allocations Manage material transaction cost code allocations
│   │   ├── create           Create a material transaction cost code allocation
│   │   ├── update           Update a material transaction cost code allocation
│   │   └── delete           Delete a material transaction cost code allocation
│   ├── material-transaction-preloads Manage material transaction preloads
│   │   ├── create           Create a material transaction preload
│   │   └── delete           Delete a material transaction preload
│   ├── material-purchase-order-releases Manage material purchase order releases
│   │   ├── create           Create a material purchase order release
│   │   ├── update           Update a material purchase order release
│   │   └── delete           Delete a material purchase order release
│   ├── material-site-readings Manage material site readings
│   │   ├── create           Create a material site reading
│   │   ├── update           Update a material site reading
│   │   └── delete           Delete a material site reading
│   ├── tractor-fuel-consumption-readings Manage tractor fuel consumption readings
│   │   ├── create           Create a tractor fuel consumption reading
│   │   ├── update           Update a tractor fuel consumption reading
│   │   └── delete           Delete a tractor fuel consumption reading
│   ├── material-type-conversions Manage material type conversions
│   │   ├── create           Create a material type conversion
│   │   ├── update           Update a material type conversion
│   │   └── delete           Delete a material type conversion
│   ├── memberships          Manage user-organization memberships
│   │   ├── create           Create a membership
│   │   ├── update           Update a membership
│   │   └── delete           Delete a membership
│   ├── broker-memberships  Manage broker memberships
│   │   ├── create           Create a broker membership
│   │   ├── update           Update a broker membership
│   │   └── delete           Delete a broker membership
│   ├── broker-commitments   Manage broker commitments
│   │   ├── create           Create a broker commitment
│   │   ├── update           Update a broker commitment
│   │   └── delete           Delete a broker commitment
│   ├── proffer-likes        Manage proffer likes
│   │   ├── create           Create a proffer like
│   │   └── delete           Delete a proffer like
│   ├── work-order-assignments Manage work order assignments
│   │   ├── create           Create a work order assignment
│   │   ├── update           Update a work order assignment
│   │   └── delete           Delete a work order assignment
│   ├── action-item-team-members Manage action item team members
│   │   ├── create           Create an action item team member
│   │   ├── update           Update an action item team member
│   │   └── delete           Delete an action item team member
│   ├── service-type-unit-of-measure-cohorts Manage service type unit of measure cohorts
│   │   ├── create           Create a service type unit of measure cohort
│   │   ├── update           Update a service type unit of measure cohort
│   │   └── delete           Delete a service type unit of measure cohort
│   ├── rate-agreement-copier-works Manage rate agreement copier works
│   │   ├── create           Create a rate agreement copier work
│   │   └── update           Update a rate agreement copier work
│   ├── retainer-payments   Manage retainer payments
│   │   ├── create           Create a retainer payment
│   │   ├── update           Update a retainer payment
│   │   └── delete           Delete a retainer payment
│   ├── maintenance-requirement-parts Manage maintenance requirement parts
│   │   ├── create           Create a maintenance requirement part
│   │   ├── update           Update a maintenance requirement part
│   │   └── delete           Delete a maintenance requirement part
│   └── maintenance-requirement-set-maintenance-requirements Manage maintenance requirement set maintenance requirements
│       ├── create           Create a maintenance requirement set maintenance requirement
│       ├── update           Update a maintenance requirement set maintenance requirement
│       └── delete           Delete a maintenance requirement set maintenance requirement
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
│   ├── proffer-likes       Browse proffer likes
│   │   ├── list            List proffer likes with filtering
│   │   └── show <id>       Show proffer like details
│   ├── brokers             Browse broker/branch information
│   │   └── list            List brokers with filtering
│   ├── administrative-incidents Browse administrative incidents
│   │   ├── list            List administrative incidents with filtering
│   │   └── show <id>       Show administrative incident details
│   ├── users               Browse users (for creator lookup)
│   │   └── list            List users with filtering
│   ├── material-suppliers  Browse material suppliers
│   │   └── list            List suppliers with filtering
│   ├── material-purchase-order-releases Browse material purchase order releases
│   │   ├── list            List material purchase order releases
│   │   └── show <id>       Show material purchase order release details
│   ├── material-transaction-cost-code-allocations Browse material transaction cost code allocations
│   │   ├── list            List material transaction cost code allocations
│   │   └── show <id>       Show material transaction cost code allocation details
│   ├── material-transaction-preloads Browse material transaction preloads
│   │   ├── list            List material transaction preloads
│   │   └── show <id>       Show material transaction preload details
│   ├── material-site-readings Browse material site readings
│   │   ├── list            List material site readings
│   │   └── show <id>       Show material site reading details
│   ├── tractor-fuel-consumption-readings Browse tractor fuel consumption readings
│   │   ├── list            List tractor fuel consumption readings
│   │   └── show <id>       Show tractor fuel consumption reading details
│   ├── material-type-conversions Browse material type conversions
│   │   ├── list            List material type conversions
│   │   └── show <id>       Show material type conversion details
│   ├── customers           Browse customers
│   │   └── list            List customers with filtering
│   ├── truckers            Browse trucking companies
│   │   └── list            List truckers with filtering
│   ├── memberships         Browse user-organization memberships
│   │   ├── list            List memberships with filtering
│   │   └── show <id>       Show membership details
│   ├── broker-memberships  Browse broker memberships
│   │   ├── list            List broker memberships with filtering
│   │   └── show <id>       Show broker membership details
│   ├── broker-commitments  Browse broker commitments
│   │   ├── list            List broker commitments
│   │   └── show <id>       Show broker commitment details
│   ├── work-order-assignments Browse work order assignments
│   │   ├── list            List work order assignments
│   │   └── show <id>       Show work order assignment details
│   ├── action-item-team-members Browse action item team members
│   │   ├── list            List action item team members
│   │   └── show <id>       Show action item team member details
│   ├── service-type-unit-of-measure-cohorts Browse service type unit of measure cohorts
│   │   ├── list            List service type unit of measure cohorts
│   │   └── show <id>       Show service type unit of measure cohort details
│   ├── maintenance-requirement-parts Browse maintenance requirement parts
│   │   ├── list            List maintenance requirement parts
│   │   └── show <id>       Show maintenance requirement part details
│   ├── maintenance-requirement-set-maintenance-requirements Browse maintenance requirement set maintenance requirements
│   │   ├── list            List maintenance requirement set maintenance requirements
│   │   └── show <id>       Show maintenance requirement set maintenance requirement details
│   ├── driver-movement-segments Browse driver movement segments
│   │   ├── list            List movement segments with filtering
│   │   └── show <id>       Show movement segment details
│   ├── job-production-plan-job-site-changes Browse job production plan job site changes
│   │   └── show <id>       Show job site change details
│   ├── job-production-plan-segments Browse job production plan segments
│   │   ├── list            List job production plan segments
│   │   └── show <id>       Show job production plan segment details
│   ├── job-production-plan-supply-demand-balances Browse job production plan supply/demand balances
│   │   ├── list            List supply/demand balances
│   │   └── show <id>       Show supply/demand balance details
│   ├── job-production-plan-schedule-change-works Browse job production plan schedule change works
│   │   ├── list            List schedule change works with filtering
│   │   └── show <id>       Show schedule change work details
│   ├── rate-agreement-copier-works Browse rate agreement copier works
│   │   ├── list            List copier works with filtering and pagination
│   │   └── show <id>       Show copier work details
│   ├── retainer-earning-statuses Browse retainer earning statuses
│   │   ├── list            List retainer earning statuses with filtering
│   │   └── show <id>       Show retainer earning status details
│   ├── retainer-payments  Browse retainer payments
│   │   ├── list            List retainer payments with filtering
│   │   └── show <id>       Show retainer payment details
│   ├── project-abandonments Browse project abandonments
│   │   ├── list            List project abandonments
│   │   └── show <id>       Show project abandonment details
│   ├── tender-returns      Browse tender returns
│   │   ├── list            List tender returns
│   │   └── show <id>       Show tender return details
│   ├── project-cancellations Browse project cancellations
│   │   ├── list            List project cancellations
│   │   └── show <id>       Show project cancellation details
│   ├── project-completions Browse project completions
│   │   ├── list            List project completions
│   │   └── show <id>       Show project completion details
│   ├── project-status-changes Browse project status changes
│   │   ├── list            List project status changes with filtering
│   │   └── show <id>       Show project status change details
│   ├── project-trailer-classifications Browse project trailer classifications
│   │   ├── list            List project trailer classifications
│   │   └── show <id>       Show project trailer classification details
│   ├── project-transport-plan-assignment-rules Browse project transport plan assignment rules
│   │   ├── list            List project transport plan assignment rules
│   │   └── show <id>       Show project transport plan assignment rule details
│   ├── project-transport-plan-event-times Browse project transport plan event times
│   │   ├── list            List project transport plan event times with filtering
│   │   └── show <id>       Show project transport plan event time details
│   ├── project-transport-plan-stops Browse project transport plan stops
│   │   ├── list            List project transport plan stops
│   │   └── show <id>       Show project transport plan stop details
│   ├── project-transport-plan-tractors Browse project transport plan tractors
│   │   ├── list            List project transport plan tractors with filtering
│   │   └── show <id>       Show project transport plan tractor details
│   ├── project-transport-plan-segment-tractors Browse project transport plan segment tractors
│   │   ├── list            List project transport plan segment tractors
│   │   └── show <id>       Show project transport plan segment tractor details
│   ├── project-margin-matrices Browse project margin matrices
│   │   ├── list            List project margin matrices
│   │   └── show <id>       Show project margin matrix details
│   ├── project-phase-cost-items Browse project phase cost items
│   │   ├── list            List project phase cost items
│   │   └── show <id>       Show project phase cost item details
│   ├── project-phase-revenue-items Browse project phase revenue items
│   │   ├── list            List project phase revenue items
│   │   └── show <id>       Show project phase revenue item details
│   ├── project-phase-cost-item-price-estimates Browse project phase cost item price estimates
│   │   ├── list            List project phase cost item price estimates
│   │   └── show <id>       Show project phase cost item price estimate details
│   ├── job-schedule-shift-start-at-changes Browse job schedule shift start-at changes
│   │   ├── list            List start-at changes
│   │   └── show <id>       Show start-at change details
│   ├── time-card-invoices  Browse time card invoices
│   │   ├── list            List time card invoices with filtering
│   │   └── show <id>       Show time card invoice details
│   ├── time-card-time-changes Browse time card time changes
│   │   ├── list            List time card time changes with filtering
│   │   └── show <id>       Show time card time change details
│   ├── time-sheet-line-items Browse time sheet line items
│   │   ├── list            List time sheet line items with filtering
│   │   └── show <id>       Show time sheet line item details
│   ├── lineup-job-schedule-shifts Browse lineup job schedule shifts
│   │   ├── list            List lineup job schedule shifts
│   │   └── show <id>       Show lineup job schedule shift details
│   ├── lineup-scenario-trailer-lineup-job-schedule-shifts Browse lineup scenario trailer lineup job schedule shifts
│   │   ├── list            List lineup scenario trailer lineup job schedule shifts
│   │   └── show <id>       Show lineup scenario trailer lineup job schedule shift details
│   ├── job-production-plan-project-phase-revenue-items Browse job production plan project phase revenue items
│   │   ├── list            List job production plan project phase revenue items
│   │   └── show <id>       Show job production plan project phase revenue item details
│   ├── hos-availability-snapshots Browse HOS availability snapshots
│   │   ├── list            List availability snapshots with filtering
│   │   └── show <id>       Show availability snapshot details
│   ├── hos-violations      Browse HOS violations
│   │   ├── list            List HOS violations with filtering
│   │   └── show <id>       Show HOS violation details
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

### Broker Commitments

```bash
# List broker commitments
xbe view broker-commitments list

# Filter by broker or trucker
xbe view broker-commitments list --broker 123
xbe view broker-commitments list --trucker 456

# Create a broker commitment
xbe do broker-commitments create --status active --broker 123 --trucker 456 --label "Q1"

# Show broker commitment details
xbe view broker-commitments show 789
```

### Project Status Changes

```bash
# List project status changes
xbe view project-status-changes list

# Filter by project
xbe view project-status-changes list --project 123

# Filter by status
xbe view project-status-changes list --status active

# Show status change details
xbe view project-status-changes show 456
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

### Broker Memberships

Broker memberships define the relationship between users and broker organizations.

```bash
# List broker memberships
xbe view broker-memberships list --broker 123

# Search broker memberships by user name
xbe view broker-memberships list --q "Jane"

# Show broker membership details
xbe view broker-memberships show 456

# Create a broker membership
xbe do broker-memberships create --user 123 --broker 456 --kind manager

# Update a broker membership
xbe do broker-memberships update 789 --title "Dispatcher" --is-rate-editor true

# Delete a broker membership (requires --confirm)
xbe do broker-memberships delete 789 --confirm
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

### Driver Assignment Refusals

Driver assignment refusals record when a driver declines a tender job schedule shift assignment.

```bash
# List refusals
xbe view driver-assignment-refusals list

# Filter by tender job schedule shift
xbe view driver-assignment-refusals list --tender-job-schedule-shift 123

# Filter by driver
xbe view driver-assignment-refusals list --driver 456

# Show refusal details
xbe view driver-assignment-refusals show 789

# Create a refusal
xbe do driver-assignment-refusals create \
  --tender-job-schedule-shift 123 \
  --driver 456 \
  --comment "Unable to cover the shift"
```

### Tender Job Schedule Shift Drivers

Tender job schedule shift drivers link drivers to tendered job schedule shifts.

```bash
# List shift drivers
xbe view tender-job-schedule-shift-drivers list

# Filter by tender job schedule shift
xbe view tender-job-schedule-shift-drivers list --tender-job-schedule-shift 123

# Filter by user
xbe view tender-job-schedule-shift-drivers list --user 456

# Show shift driver details
xbe view tender-job-schedule-shift-drivers show 789

# Create a shift driver
xbe do tender-job-schedule-shift-drivers create \
  --tender-job-schedule-shift 123 \
  --user 456 \
  --is-primary

# Update a shift driver
xbe do tender-job-schedule-shift-drivers update 789 --is-primary true

# Delete a shift driver
xbe do tender-job-schedule-shift-drivers delete 789 --confirm
```

### Tender Job Schedule Shifts

Tender job schedule shifts represent tendered shifts tied to job schedule shifts.

```bash
# List shifts
xbe view tender-job-schedule-shifts list

# Filter by tender
xbe view tender-job-schedule-shifts list --tender 123

# Filter by driver
xbe view tender-job-schedule-shifts list --seller-operations-contact 456

# Show shift details
xbe view tender-job-schedule-shifts show 789

# Create a shift
xbe do tender-job-schedule-shifts create \
  --tender-type broker-tenders \
  --tender-id 123 \
  --job-schedule-shift 456 \
  --material-transaction-status open

# Update a shift
xbe do tender-job-schedule-shifts update 789 --seller-operations-contact 456

# Delete a shift
xbe do tender-job-schedule-shifts delete 789 --confirm
```

### Tender Returns

Tender returns record when accepted tenders are returned.

```bash
# List tender returns
xbe view tender-returns list

# Show tender return details
xbe view tender-returns show 123

# Return a tender
xbe do tender-returns create --tender-type broker-tenders --tender-id 123 --comment "Returned"
```

### Driver Movement Segments

Driver movement segments represent contiguous moving or stationary intervals for a driver day.

```bash
# List recent segments
xbe view driver-movement-segments list --limit 10

# Filter moving segments
xbe view driver-movement-segments list --is-moving true

# Filter by segment set
xbe view driver-movement-segments list --driver-movement-segment-set 123

# Show segment details
xbe view driver-movement-segments show 456
```

### Equipment Movement Trips

Equipment movement trips track equipment transfers between stops.

```bash
# List trips
xbe view equipment-movement-trips list

# Filter by broker
xbe view equipment-movement-trips list --broker 123

# Show trip details
xbe view equipment-movement-trips show 456

# Create a trip
xbe do equipment-movement-trips create --broker 123 --job-number "EMT-100"

# Update mobilization timing
xbe do equipment-movement-trips update 456 --explicit-driver-day-mobilization-before-minutes 30

# Delete a trip (requires --confirm)
xbe do equipment-movement-trips delete 456 --confirm
```

### Equipment Movement Trip Dispatch Fulfillment Clerks

Equipment movement trip dispatch fulfillment clerks trigger the fulfillment workflow for a dispatch.

```bash
# Run fulfillment for a dispatch
xbe do equipment-movement-trip-dispatch-fulfillment-clerks create \
  --equipment-movement-trip-dispatch 123
```

### Lineup Dispatch Fulfillment Clerks

Lineup dispatch fulfillment clerks trigger the fulfillment workflow for a lineup dispatch.

```bash
# Run fulfillment for a dispatch
xbe do lineup-dispatch-fulfillment-clerks create \
  --lineup-dispatch 123
```

### Lineup Dispatch Statuses

Lineup dispatch statuses compute the offered tender percentage for a broker and lineup window.

```bash
# Check lineup dispatch status for a day window
xbe do lineup-dispatch-statuses create --broker 123 --window day --date 2025-01-23
```

### Shift Scope Tenders

Shift scope tenders return tender IDs that match a shift scope, optionally filtered by tender creation time.

```bash
# Find tenders for a shift scope
xbe do shift-scope-tenders create --shift-scope 123

# Filter by created_at window and limit results
xbe do shift-scope-tenders create --shift-scope 123 \
  --created-at-min 2025-01-01 --created-at-max 2025-01-31 --limit 5
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
