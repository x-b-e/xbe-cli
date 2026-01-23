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
│   ├── equipment-suppliers  Manage equipment suppliers
│   │   ├── create           Create an equipment supplier
│   │   ├── update           Update an equipment supplier
│   │   └── delete           Delete an equipment supplier
│   ├── driver-day-adjustments Manage driver day adjustments
│   │   ├── create           Create a driver day adjustment
│   │   ├── update           Update a driver day adjustment
│   │   └── delete           Delete a driver day adjustment
│   ├── job-site-times        Manage job site times
│   │   ├── create           Create a job site time
│   │   ├── update           Update a job site time
│   │   └── delete           Delete a job site time
│   ├── time-card-approval-audits Manage time card approval audits
│   │   ├── create           Create a time card approval audit
│   │   ├── update           Update a time card approval audit
│   │   └── delete           Delete a time card approval audit
│   ├── time-card-payroll-certifications Manage time card payroll certifications
│   │   ├── create           Create a time card payroll certification
│   │   └── delete           Delete a time card payroll certification
│   ├── time-card-unsubmissions Unsubmit time cards
│   │   └── create           Unsubmit a time card
│   ├── time-sheet-submissions Submit time sheets
│   │   └── create           Submit a time sheet
│   ├── time-sheet-unsubmissions Unsubmit time sheets
│   │   └── create           Unsubmit a time sheet
│   ├── time-sheet-cost-code-allocations Manage time sheet cost code allocations
│   │   ├── create           Create a time sheet cost code allocation
│   │   ├── update           Update a time sheet cost code allocation
│   │   └── delete           Delete a time sheet cost code allocation
│   ├── service-events        Manage service events
│   │   ├── create           Create a service event
│   │   ├── update           Update a service event
│   │   └── delete           Delete a service event
│   ├── hos-annotations      Manage HOS annotations
│   │   └── delete           Delete a HOS annotation
│   ├── driver-managers      Manage driver managers
│   │   ├── create           Create a driver manager
│   │   ├── update           Update a driver manager
│   │   └── delete           Delete a driver manager
│   ├── lane-summary         Generate lane (cycle) summaries
│   │   └── create           Create a lane summary
│   ├── material-transaction-summary  Generate material transaction summaries
│   │   └── create           Create a material transaction summary
│   ├── material-transaction-denials  Deny material transactions
│   │   └── create           Deny a material transaction
│   ├── material-transaction-inspection-rejections Manage material transaction inspection rejections
│   │   ├── create           Create a material transaction inspection rejection
│   │   ├── update           Update a material transaction inspection rejection
│   │   └── delete           Delete a material transaction inspection rejection
│   ├── material-unit-of-measure-quantities Manage material unit of measure quantities
│   │   ├── create           Create a material unit of measure quantity
│   │   ├── update           Update a material unit of measure quantity
│   │   └── delete           Delete a material unit of measure quantity
│   ├── material-purchase-orders Manage material purchase orders
│   │   ├── create           Create a material purchase order
│   │   ├── update           Update a material purchase order
│   │   └── delete           Delete a material purchase order
│   ├── material-site-unavailabilities Manage material site unavailabilities
│   │   ├── create           Create a material site unavailability
│   │   ├── update           Update a material site unavailability
│   │   └── delete           Delete a material site unavailability
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
│   ├── material-purchase-orders Browse material purchase orders
│   │   ├── list            List material purchase orders with filtering
│   │   └── show <id>       Show material purchase order details
│   ├── material-site-unavailabilities Browse material site unavailabilities
│   │   ├── list            List material site unavailabilities with filtering
│   │   └── show <id>       Show material site unavailability details
│   ├── material-transaction-inspection-rejections Browse material transaction inspection rejections
│   │   ├── list            List material transaction inspection rejections with filtering
│   │   └── show <id>       Show material transaction inspection rejection details
│   ├── material-unit-of-measure-quantities Browse material unit of measure quantities
│   │   ├── list            List material unit of measure quantities with filtering
│   │   └── show <id>       Show material unit of measure quantity details
│   ├── material-transaction-status-changes Browse material transaction status changes
│   │   ├── list            List material transaction status changes with filtering
│   │   └── show <id>       Show material transaction status change details
│   ├── equipment-suppliers Browse equipment suppliers
│   │   ├── list            List equipment suppliers with filtering
│   │   └── show <id>       Show equipment supplier details
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
│   ├── driver-day-adjustments Browse driver day adjustments
│   │   ├── list            List driver day adjustments with filtering
│   │   └── show <id>       Show driver day adjustment details
│   ├── job-site-times       Browse job site times
│   │   ├── list            List job site times with filtering
│   │   └── show <id>       Show job site time details
│   ├── time-card-approval-audits Browse time card approval audits
│   │   ├── list            List time card approval audits with filtering
│   │   └── show <id>       Show time card approval audit details
│   ├── time-card-status-changes Browse time card status changes
│   │   ├── list            List time card status changes with filtering
│   │   └── show <id>       Show time card status change details
│   ├── time-card-payroll-certifications Browse time card payroll certifications
│   │   ├── list            List time card payroll certifications with filtering
│   │   └── show <id>       Show time card payroll certification details
│   ├── time-sheet-cost-code-allocations Browse time sheet cost code allocations
│   │   ├── list            List time sheet cost code allocations with filtering
│   │   └── show <id>       Show time sheet cost code allocation details
│   ├── time-sheet-unsubmissions Browse time sheet unsubmissions
│   │   ├── list            List time sheet unsubmissions
│   │   └── show <id>       Show time sheet unsubmission details
│   ├── service-events       Browse service events
│   │   ├── list            List service events with filtering
│   │   └── show <id>       Show service event details
│   ├── driver-managers     Browse driver managers
│   │   ├── list            List driver managers with filtering
│   │   └── show <id>       Show driver manager details
│   ├── hos-annotations     Browse HOS annotations
│   │   ├── list            List HOS annotations with filtering
│   │   └── show <id>       Show HOS annotation details
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

### Material Purchase Orders

```bash
# List purchase orders for a supplier
xbe view material-purchase-orders list --material-supplier 123

# View a purchase order
xbe view material-purchase-orders show 456

# Create a purchase order
xbe do material-purchase-orders create \
  --broker 123 \
  --material-supplier 456 \
  --material-type 789 \
  --unit-of-measure 10 \
  --quantity 500
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

### Material Transaction Denials

Material transaction denials set a transaction's status to denied. Comments are required
for submitted or accepted transactions.

```bash
# Deny a material transaction
xbe do material-transaction-denials create \
  --material-transaction 123 \
  --comment "Load contaminated"
```

### Material Transaction Inspection Rejections

Material transaction inspection rejections record rejected quantities and notes for inspections.

```bash
# List rejections for an inspection
xbe view material-transaction-inspection-rejections list --material-transaction-inspection 123

# Create a rejection
xbe do material-transaction-inspection-rejections create \
  --material-transaction-inspection 123 \
  --unit-of-measure 45 \
  --quantity 10 \
  --note "Excess moisture"

# Update a rejection note
xbe do material-transaction-inspection-rejections update 456 --note "Updated note"
```

### Material Transaction Status Changes

Material transaction status changes record status history for a transaction.

```bash
# List status changes for a material transaction
xbe view material-transaction-status-changes list --material-transaction 123

# Filter status changes by status
xbe view material-transaction-status-changes list --status accepted

# Show a specific status change
xbe view material-transaction-status-changes show 456
```

### Material Mix Design Matches

Material mix design matches find the best mix designs for a material type
and optional material sites as of a timestamp.

```bash
# Match material mix designs for a material type
xbe do material-mix-design-matches create --material-type 123 --as-of "2026-01-23T00:00:00Z"

# Match with material sites
xbe do material-mix-design-matches create \
  --material-type 123 \
  --as-of "2026-01-23T00:00:00Z" \
  --material-sites 456,789

# JSON output
xbe do material-mix-design-matches create --material-type 123 --as-of "2026-01-23T00:00:00Z" --json
```

### Shift Scope Matches

Shift scope matches evaluate whether a rate or shift set time card constraint
matches a tender's shifts, returning a summary and match details.

```bash
# Match a tender against a rate
xbe do shift-scope-matches create --tender 123 --rate 456

# Match using a shift set time card constraint
xbe do shift-scope-matches create --tender 123 --shift-set-time-card-constraint 789

# Include match SQL when permitted
xbe do shift-scope-matches create --tender 123 --rate 456 --show-matching-shift-sql

# JSON output
xbe do shift-scope-matches create --tender 123 --rate 456 --json
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

### Driver Assignment Acknowledgements

Driver assignment acknowledgements record when a driver acknowledges a tender job schedule shift assignment.

```bash
# List acknowledgements
xbe view driver-assignment-acknowledgements list

# Filter by tender job schedule shift
xbe view driver-assignment-acknowledgements list --tender-job-schedule-shift 123

# Filter by driver
xbe view driver-assignment-acknowledgements list --driver 456

# Show acknowledgement details
xbe view driver-assignment-acknowledgements show 789

# Create an acknowledgement
xbe do driver-assignment-acknowledgements create --tender-job-schedule-shift 123 --driver 456
```

### Driver Managers

Driver managers link manager memberships to managed memberships within a trucker.

```bash
# List driver managers
xbe view driver-managers list

# Filter by trucker
xbe view driver-managers list --trucker 123

# Filter by manager membership
xbe view driver-managers list --manager-membership 456

# Filter by managed membership
xbe view driver-managers list --managed-membership 789

# Filter by manager user
xbe view driver-managers list --manager-user 654

# Show driver manager details
xbe view driver-managers show 321

# Create a driver manager
xbe do driver-managers create --trucker 123 --manager-membership 456 --managed-membership 789

# Update a driver manager
xbe do driver-managers update 321 --manager-membership 456

# Delete a driver manager (requires --confirm)
xbe do driver-managers delete 321 --confirm
```

### HOS Annotations

HOS annotations capture comments and metadata for hours-of-service days and events.

```bash
# List HOS annotations
xbe view hos-annotations list

# Filter by HOS day
xbe view hos-annotations list --hos-day 123

# Filter by HOS event
xbe view hos-annotations list --hos-event 456

# Show annotation details
xbe view hos-annotations show 789

# Delete a HOS annotation (requires --confirm)
xbe do hos-annotations delete 789 --confirm
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

### Equipment Location Estimates

Equipment location estimates return the most recent known location for equipment.

```bash
# Estimate location for a specific equipment ID
xbe view equipment-location-estimates list --equipment 123

# Estimate location as of a specific time
xbe view equipment-location-estimates list --equipment 123 --as-of 2026-01-23T12:00:00Z

# Constrain the event window
xbe view equipment-location-estimates list --equipment 123 \
  --earliest-event-at 2026-01-22T00:00:00Z \
  --latest-event-at 2026-01-23T00:00:00Z

# Output as JSON
xbe view equipment-location-estimates list --equipment 123 --json
```

### Equipment Movement Stop Completions

Equipment movement stop completions record when a movement stop was completed.

```bash
# List stop completions
xbe view equipment-movement-stop-completions list

# Filter by stop
xbe view equipment-movement-stop-completions list --stop 123

# Create a completion
xbe do equipment-movement-stop-completions create \
  --stop 123 \
  --completed-at 2026-01-23T12:34:56Z

# Update a completion
xbe do equipment-movement-stop-completions update 456 \
  --completed-at 2026-01-23T13:00:00Z \
  --note "Gate locked"

# Delete a completion (requires --confirm)
xbe do equipment-movement-stop-completions delete 456 --confirm
```

### Maintenance Requirement Rule Maintenance Requirement Sets

Maintenance requirement rule maintenance requirement sets link maintenance requirement rules
to template maintenance requirement sets.

```bash
# List maintenance requirement rule maintenance requirement sets
xbe view maintenance-requirement-rule-maintenance-requirement-sets list

# Filter by maintenance requirement rule
xbe view maintenance-requirement-rule-maintenance-requirement-sets list --maintenance-requirement-rule 123

# Filter by maintenance requirement set
xbe view maintenance-requirement-rule-maintenance-requirement-sets list --maintenance-requirement-set 456

# Show maintenance requirement rule maintenance requirement set details
xbe view maintenance-requirement-rule-maintenance-requirement-sets show 789

# Create a maintenance requirement rule maintenance requirement set
xbe do maintenance-requirement-rule-maintenance-requirement-sets create \
  --maintenance-requirement-rule 123 \
  --maintenance-requirement-set 456

# Delete a maintenance requirement rule maintenance requirement set (requires --confirm)
xbe do maintenance-requirement-rule-maintenance-requirement-sets delete 789 --confirm
```

### Job Production Plan Submissions

Job production plan submissions move plans from editing or rejected to submitted.

```bash
# Submit a job production plan
xbe do job-production-plan-submissions create --job-production-plan 123

# Submit with a comment
xbe do job-production-plan-submissions create \
  --job-production-plan 123 \
  --comment "Ready for review"

# Submit while suppressing notifications
  xbe do job-production-plan-submissions create \
    --job-production-plan 123 \
    --suppress-status-change-notifications
```

### Job Production Plan Uncancellations

Job production plan uncancellations restore cancelled plans to their previous status.

```bash
# Uncancel a job production plan
xbe do job-production-plan-uncancellations create --job-production-plan 123

# Uncancel with a comment
xbe do job-production-plan-uncancellations create \
  --job-production-plan 123 \
  --comment "Reopen plan"

# Uncancel while suppressing notifications
xbe do job-production-plan-uncancellations create \
  --job-production-plan 123 \
  --suppress-status-change-notifications
```

### Lineup Job Production Plans

Lineup job production plans connect lineups to job production plans.

```bash
# List lineup job production plans
xbe view lineup-job-production-plans list

# Filter by lineup
xbe view lineup-job-production-plans list --lineup 123

# Filter by job production plan
xbe view lineup-job-production-plans list --job-production-plan 456

# Show lineup job production plan details
xbe view lineup-job-production-plans show 789

# Create a lineup job production plan
xbe do lineup-job-production-plans create --lineup 123 --job-production-plan 456

# Delete a lineup job production plan (requires --confirm)
xbe do lineup-job-production-plans delete 789 --confirm
```

### Lineup Scenario Lineup Job Schedule Shifts

Lineup scenario lineup job schedule shifts connect lineup scenarios to lineup job schedule shifts.

```bash
# List lineup scenario lineup job schedule shifts
xbe view lineup-scenario-lineup-job-schedule-shifts list

# Filter by lineup scenario
xbe view lineup-scenario-lineup-job-schedule-shifts list --lineup-scenario 123

# Filter by lineup job schedule shift
xbe view lineup-scenario-lineup-job-schedule-shifts list --lineup-job-schedule-shift 456

# Show lineup scenario lineup job schedule shift details
xbe view lineup-scenario-lineup-job-schedule-shifts show 789

# Create a lineup scenario lineup job schedule shift
xbe do lineup-scenario-lineup-job-schedule-shifts create --lineup-scenario 123 --lineup-job-schedule-shift 456

# Delete a lineup scenario lineup job schedule shift (requires --confirm)
xbe do lineup-scenario-lineup-job-schedule-shifts delete 789 --confirm
```

### Lineup Scenario Trailers

Lineup scenario trailers attach trailers to lineup scenario truckers.

```bash
# List lineup scenario trailers
xbe view lineup-scenario-trailers list

# Filter by lineup scenario trucker
xbe view lineup-scenario-trailers list --lineup-scenario-trucker 123

# Filter by lineup scenario
xbe view lineup-scenario-trailers list --lineup-scenario 456

# Filter by trailer
xbe view lineup-scenario-trailers list --trailer 789

# Show lineup scenario trailer details
xbe view lineup-scenario-trailers show 123

# Create a lineup scenario trailer
xbe do lineup-scenario-trailers create --lineup-scenario-trucker 123 --trailer 456

# Update a lineup scenario trailer
xbe do lineup-scenario-trailers update 123 --last-assigned-on 2024-01-01

# Delete a lineup scenario trailer (requires --confirm)
xbe do lineup-scenario-trailers delete 123 --confirm
```

### Job Schedule Shift Managed Toggles

Job schedule shift managed toggles flip a shift between managed and unmanaged.
Shifts must not be cancelled, tied to unmanaged tenders, or have time cards.

```bash
# Toggle managed status for a job schedule shift
xbe do job-schedule-shift-is-managed-toggles create --job-schedule-shift 123
```

### Job Production Plan Alarms

Job production plan alarms notify subscribers when production reaches
specified tonnage thresholds or exceeds latency targets.

```bash
# List alarms
xbe view job-production-plan-alarms list

# Filter by job production plan
xbe view job-production-plan-alarms list --job-production-plan 123

# Show alarm details
xbe view job-production-plan-alarms show 456

# Create an alarm
xbe do job-production-plan-alarms create \
  --job-production-plan 123 \
  --tons 150 \
  --base-material-type-fully-qualified-name "Asphalt Mixture" \
  --max-latency-minutes 45 \
  --note "Alert at 150 tons"

# Update an alarm
xbe do job-production-plan-alarms update 456 --tons 200 --note "Updated trigger"

# Delete an alarm (requires --confirm)
xbe do job-production-plan-alarms delete 456 --confirm
```

### Job Production Plan Cost Codes

Job production plan cost codes map cost codes to job production plans.

```bash
# List job production plan cost codes
xbe view job-production-plan-cost-codes list

# Filter by job production plan
xbe view job-production-plan-cost-codes list --job-production-plan 123

# Show job production plan cost code details
xbe view job-production-plan-cost-codes show 456

# Create a job production plan cost code
xbe do job-production-plan-cost-codes create --job-production-plan 123 --cost-code 789

# Update a job production plan cost code
xbe do job-production-plan-cost-codes update 456 --project-resource-classification 321

# Delete a job production plan cost code (requires --confirm)
xbe do job-production-plan-cost-codes delete 456 --confirm
```

### Job Production Plan Segment Sets

Job production plan segment sets group plan segments and define offsets.

```bash
# List job production plan segment sets
xbe view job-production-plan-segment-sets list

# Filter by job production plan
xbe view job-production-plan-segment-sets list --job-production-plan 123

# Show job production plan segment set details
xbe view job-production-plan-segment-sets show 456

# Create a job production plan segment set
xbe do job-production-plan-segment-sets create --job-production-plan 123 --name "AM shift"

# Update a job production plan segment set
xbe do job-production-plan-segment-sets update 456 --start-offset-minutes 15 --is-default

# Delete a job production plan segment set (requires --confirm)
xbe do job-production-plan-segment-sets delete 456 --confirm
```

### Job Production Plan Inspectors

Job production plan inspectors assign inspectors (users) to job production plans.

```bash
# List job production plan inspectors
xbe view job-production-plan-inspectors list

# Filter by job production plan
xbe view job-production-plan-inspectors list --job-production-plan-id 123

# Filter by user
xbe view job-production-plan-inspectors list --user 456

# Show job production plan inspector details
xbe view job-production-plan-inspectors show 789

# Create a job production plan inspector
xbe do job-production-plan-inspectors create --job-production-plan-id 123 --user 456

# Delete a job production plan inspector (requires --confirm)
xbe do job-production-plan-inspectors delete 789 --confirm
```

### Job Production Plan Safety Risks Suggestions

Job production plan safety risks suggestions generate AI safety risk lists for a
job production plan.

```bash
# List safety risks suggestions
xbe view job-production-plan-safety-risks-suggestions list

# Filter by job production plan
xbe view job-production-plan-safety-risks-suggestions list --job-production-plan 123

# Show suggestion details
xbe view job-production-plan-safety-risks-suggestions show 456

# Generate safety risks suggestions
xbe do job-production-plan-safety-risks-suggestions create --job-production-plan 123

# Generate with options
xbe do job-production-plan-safety-risks-suggestions create \
  --job-production-plan 123 \
  --options '{"include_other_incidents":true}'

  # Generate synchronously (wait for risks)
  xbe do job-production-plan-safety-risks-suggestions create \
    --job-production-plan 123 \
    --is-async=false
```

### Job Site Times

Job site times track the time a user spent at a job site for a job production plan.

```bash
# List job site times
xbe view job-site-times list

# Filter by job production plan
xbe view job-site-times list --job-production-plan 123

# Filter by user
xbe view job-site-times list --user 456

# Show job site time details
xbe view job-site-times show 789

# Create a job site time
xbe do job-site-times create \
  --job-production-plan 123 \
  --user 456 \
  --start-at 2026-01-23T08:00:00Z \
  --end-at 2026-01-23T10:00:00Z \
  --description "On site"

# Update a job site time
xbe do job-site-times update 789 --description "Updated notes"

# Delete a job site time (requires --confirm)
xbe do job-site-times delete 789 --confirm
```

### Time Sheet Cost Code Allocations

Time sheet cost code allocations split a time sheet's costs across cost codes.

```bash
# List time sheet cost code allocations
xbe view time-sheet-cost-code-allocations list

# Filter by time sheet
xbe view time-sheet-cost-code-allocations list --time-sheet 123

# Show allocation details
xbe view time-sheet-cost-code-allocations show 456

# Create a time sheet cost code allocation
xbe do time-sheet-cost-code-allocations create \
  --time-sheet 123 \
  --details '[{"cost_code_id":"789","percentage":1}]'

# Update allocation details
xbe do time-sheet-cost-code-allocations update 456 \
  --details '[{"cost_code_id":"789","percentage":1}]'

# Delete a time sheet cost code allocation (requires --confirm)
xbe do time-sheet-cost-code-allocations delete 456 --confirm
```

### Time Sheet Submissions

Time sheet submissions move a time sheet into submitted status.

```bash
# Submit a time sheet
xbe do time-sheet-submissions create --time-sheet 123

# Submit with a comment
xbe do time-sheet-submissions create --time-sheet 123 --comment "Ready for approval"
```

### Time Sheet Unsubmissions

Time sheet unsubmissions move a time sheet back to editing status.

```bash
# List time sheet unsubmissions
xbe view time-sheet-unsubmissions list

# Show a time sheet unsubmission
xbe view time-sheet-unsubmissions show 456

# Unsubmit a time sheet
xbe do time-sheet-unsubmissions create --time-sheet 123

# Unsubmit with a comment
xbe do time-sheet-unsubmissions create --time-sheet 123 --comment "Needs edits"
```

### Material Site Unavailabilities

Material site unavailabilities track downtime windows for material sites.

```bash
# List material site unavailabilities
xbe view material-site-unavailabilities list

# Filter by material site
xbe view material-site-unavailabilities list --material-site 123

# Show material site unavailability details
xbe view material-site-unavailabilities show 456

# Create a material site unavailability
xbe do material-site-unavailabilities create \
  --material-site 123 \
  --start-at 2026-01-24T08:00:00Z \
  --end-at 2026-01-24T12:00:00Z \
  --description "Planned maintenance"

# Update a material site unavailability
xbe do material-site-unavailabilities update 456 --description "Updated notes"

# Delete a material site unavailability (requires --confirm)
xbe do material-site-unavailabilities delete 456 --confirm
```

### Job Production Plan Trucking Incident Detectors

Job production plan trucking incident detectors analyze material transactions
and identify potential trucking incidents based on ordering inconsistencies.

```bash
# List trucking incident detectors
xbe view job-production-plan-trucking-incident-detectors list

# Filter by job production plan
xbe view job-production-plan-trucking-incident-detectors list --job-production-plan 123

# Filter by performed status
xbe view job-production-plan-trucking-incident-detectors list --is-performed true

# Show detector details
xbe view job-production-plan-trucking-incident-detectors show 456

# Run a detector
xbe do job-production-plan-trucking-incident-detectors create --job-production-plan 123

# Run as of a timestamp and persist incident changes
xbe do job-production-plan-trucking-incident-detectors create \
  --job-production-plan 123 \
  --as-of "2026-01-23T00:00:00Z" \
  --persist-changes
```

### Job Production Plan Material Type Quality Control Requirements

Job production plan material type quality control requirements attach quality
control classifications to job production plan material types.

```bash
# List requirements
xbe view job-production-plan-material-type-quality-control-requirements list

# Filter by job production plan material type
xbe view job-production-plan-material-type-quality-control-requirements list --job-production-plan-material-type 123

# Filter by quality control classification
xbe view job-production-plan-material-type-quality-control-requirements list --quality-control-classification 456

# Show requirement details
xbe view job-production-plan-material-type-quality-control-requirements show 789

# Create a requirement
xbe do job-production-plan-material-type-quality-control-requirements create \
  --job-production-plan-material-type 123 \
  --quality-control-classification 456 \
  --note "Temperature check"

# Update a requirement
xbe do job-production-plan-material-type-quality-control-requirements update 789 --note "Updated note"

# Delete a requirement (requires --confirm)
xbe do job-production-plan-material-type-quality-control-requirements delete 789 --confirm
```

### Inventory Capacities

Inventory capacities define min/max storage levels and alert thresholds for a
material site and material type.

```bash
# List inventory capacities
xbe view inventory-capacities list

# Filter by material site
xbe view inventory-capacities list --material-site 123

# Show capacity details
xbe view inventory-capacities show 456

# Create an inventory capacity
xbe do inventory-capacities create --material-site 123 --material-type 789 \
  --min-capacity-tons 50 --max-capacity-tons 500 --threshold-tons 75

# Update capacity thresholds
xbe do inventory-capacities update 456 --threshold-tons 120

# Delete a capacity (requires --confirm)
xbe do inventory-capacities delete 456 --confirm
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
