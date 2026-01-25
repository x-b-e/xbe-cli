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
│   ├── api-tokens           Manage API tokens
│   │   ├── create           Create an API token
│   │   └── update           Update an API token
│   ├── action-item-tracker-update-requests  Manage action item tracker update requests
│   │   ├── create           Create an update request
│   │   ├── update           Update an update request
│   │   └── delete           Delete an update request
│   ├── change-requests       Manage change requests
│   │   ├── create           Create a change request
│   │   ├── update           Update a change request
│   │   └── delete           Delete a change request
│   ├── hos-days              Manage HOS days
│   │   └── update            Update an HOS day
│   ├── answer-related-contents  Manage answer related contents
│   │   ├── create           Create an answer related content link
│   │   ├── update           Update an answer related content link
│   │   └── delete           Delete an answer related content link
│   ├── glossary-terms       Manage glossary terms
│   │   ├── create           Create a glossary term
│   │   ├── update           Update a glossary term
│   │   └── delete           Delete a glossary term
│   ├── ui-tour-steps        Manage UI tour steps
│   │   ├── create           Create a UI tour step
│   │   ├── update           Update a UI tour step
│   │   └── delete           Delete a UI tour step
│   ├── user-ui-tours        Manage user UI tours
│   │   ├── create           Create a user UI tour
│   │   ├── update           Update a user UI tour
│   │   └── delete           Delete a user UI tour
│   ├── platform-statuses    Manage platform status updates
│   │   ├── create           Create a platform status
│   │   ├── update           Update a platform status
│   │   └── delete           Delete a platform status
│   ├── post-children        Manage post child links
│   │   ├── create           Create a post child link
│   │   └── delete           Delete a post child link
│   ├── post-views           Record post views
│   │   └── create           Record a post view
│   ├── marketing-metrics    Refresh marketing metrics
│   │   └── create           Refresh marketing metrics
│   ├── lane-summary         Generate lane (cycle) summaries
│   │   └── create           Create a lane summary
│   ├── material-transaction-summary  Generate material transaction summaries
│   │   └── create           Create a material transaction summary
│   ├── job-production-plan-broadcast-messages  Manage job production plan broadcast messages
│   │   ├── create           Create a broadcast message
│   │   └── update           Update a broadcast message
│   ├── job-production-plan-safety-risk-communication-suggestions  Generate safety risk communication suggestions
│   │   ├── create           Generate a safety risk communication suggestion
│   │   └── delete           Delete a safety risk communication suggestion
│   ├── job-production-plan-material-site-changes  Manage job production plan material site changes
│   │   └── create           Create a material site change
│   ├── job-production-plan-status-changes  Manage job production plan status changes
│   │   └── update           Update a job production plan status change
│   ├── job-production-plan-duplications  Duplicate job production plan templates
│   │   └── create           Duplicate a job production plan template
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
│   ├── action-item-tracker-update-requests  Browse action item tracker update requests
│   │   ├── list            List update requests with filtering
│   │   └── show <id>       Show update request details
│   ├── change-requests       Browse change requests
│   │   ├── list            List change requests with filtering
│   │   └── show <id>       Show change request details
│   ├── incident-headline-suggestions  Browse incident headline suggestions
│   │   ├── list            List incident headline suggestions
│   │   └── show <id>       Show incident headline suggestion details
│   ├── prompt-prescriptions  Browse prompt prescriptions
│   │   ├── list            List prompt prescriptions
│   │   └── show <id>       Show prompt prescription details
│   ├── incident-tag-incidents  Browse incident tag incident links
│   │   ├── list            List incident tag incident links with filtering
│   │   └── show <id>       Show incident tag incident details
│   ├── safety-incidents     Browse safety incidents
│   │   ├── list            List safety incidents with filtering
│   │   └── show <id>       Show safety incident details
│   ├── production-incident-detectors  Browse production incident detector runs
│   │   ├── list            List detector runs
│   │   └── show <id>       Show detector run details
│   ├── hos-days             Browse HOS days
│   │   ├── list            List HOS days with filtering
│   │   └── show <id>       Show HOS day details
│   ├── answer-related-contents  Browse answer related contents
│   │   ├── list            List answer related contents with filtering
│   │   └── show <id>       Show answer related content details
│   ├── newsletters         Browse and view newsletters
│   │   ├── list            List newsletters with filtering
│   │   └── show <id>       Show newsletter details
│   ├── posts               Browse and view posts
│   │   ├── list            List posts with filtering
│   │   └── show <id>       Show post details
│   ├── job-production-plan-broadcast-messages  Browse job production plan broadcast messages
│   │   ├── list            List broadcast messages
│   │   └── show <id>       Show broadcast message details
│   ├── job-production-plan-safety-risk-communication-suggestions  Browse safety risk communication suggestions
│   │   ├── list            List safety risk communication suggestions
│   │   └── show <id>       Show safety risk communication suggestion details
│   ├── job-production-plan-recaps  Browse job production plan recaps
│   │   ├── list            List job production plan recaps
│   │   └── show <id>       Show job production plan recap details
│   ├── job-production-plan-material-site-changes  Browse job production plan material site changes
│   │   ├── list            List material site changes
│   │   └── show <id>       Show material site change details
│   ├── job-production-plan-status-changes  Browse job production plan status changes
│   │   ├── list            List job production plan status changes
│   │   └── show <id>       Show job production plan status change details
│   ├── brokers             Browse broker/branch information
│   │   └── list            List brokers with filtering
│   ├── broker-certification-types  Browse broker certification types
│   │   ├── list            List broker certification types with filtering
│   │   └── show <id>       Show broker certification type details
│   ├── customer-certification-types  Browse customer certification types
│   │   ├── list            List customer certification types with filtering
│   │   └── show <id>       Show customer certification type details
│   ├── broker-equipment-classifications  Browse broker equipment classifications
│   │   ├── list            List broker equipment classifications with filtering
│   │   └── show <id>       Show broker equipment classification details
│   ├── cost-code-trucking-cost-summaries  Browse cost code trucking cost summaries
│   │   ├── list            List cost code trucking cost summaries with filtering
│   │   └── show <id>       Show cost code trucking cost summary details
│   ├── pave-frame-actual-statistics  Browse pave frame actual statistics
│   │   ├── list            List pave frame actual statistics with filtering
│   │   └── show <id>       Show pave frame actual statistic details
│   ├── business-unit-equipments  Browse business unit equipment links
│   │   ├── list            List business unit equipment links with filtering
│   │   └── show <id>       Show business unit equipment details
│   ├── broker-trucker-ratings  Browse broker trucker ratings
│   │   ├── list            List broker trucker ratings with filtering
│   │   └── show <id>       Show broker trucker rating details
│   ├── expected-time-of-arrivals  Browse expected time of arrival updates
│   │   ├── list            List expected time of arrivals with filtering
│   │   └── show <id>       Show expected time of arrival details
│   ├── prediction-agents   Browse prediction agents
│   │   ├── list            List prediction agents with filtering
│   │   └── show <id>       Show prediction agent details
│   ├── prediction-knowledge-base-questions  Browse prediction knowledge base questions
│   │   ├── list            List prediction knowledge base questions with filtering
│   │   └── show <id>       Show prediction knowledge base question details
│   ├── prediction-subject-memberships  Browse prediction subject memberships
│   │   ├── list            List prediction subject memberships with filtering
│   │   └── show <id>       Show prediction subject membership details
│   ├── prediction-subject-recaps  Browse prediction subject recaps
│   │   ├── list            List prediction subject recaps with filtering
│   │   └── show <id>       Show prediction subject recap details
│   ├── prediction-subject-gap-portions  Browse prediction subject gap portions
│   │   ├── list            List prediction subject gap portions with filtering
│   │   └── show <id>       Show prediction subject gap portion details
│   ├── lowest-losing-bid-prediction-subject-details  Browse lowest losing bid prediction subject details
│   │   ├── list            List details with filtering
│   │   └── show <id>       Show detail record
│   ├── users               Browse users (for creator lookup)
│   │   └── list            List users with filtering
│   ├── user-creator-feeds   Browse user creator feeds
│   │   ├── list            List user creator feeds with filtering
│   │   └── show <id>       Show user creator feed details
│   ├── user-post-feeds      Browse user post feeds
│   │   ├── list            List user post feeds with filtering
│   │   └── show <id>       Show user post feed details
│   ├── user-searches        Browse user searches
│   │   └── list            List user searches
│   ├── api-tokens          Browse API tokens
│   │   ├── list            List API tokens with filtering
│   │   └── show <id>       Show API token details
│   ├── material-suppliers  Browse material suppliers
│   │   └── list            List suppliers with filtering
│   ├── material-site-reading-material-types  Browse material site reading material types
│   │   ├── list            List material site reading material types with filtering
│   │   └── show <id>       Show material site reading material type details
│   ├── material-type-material-site-inventory-locations  Browse material type material site inventory locations
│   │   ├── list            List material type material site inventory locations with filtering
│   │   └── show <id>       Show material type material site inventory location details
│   ├── material-purchase-order-release-redemptions  Browse material purchase order release redemptions
│   │   ├── list            List release redemptions with filtering
│   │   └── show <id>       Show release redemption details
│   ├── haskell-lemon-inbound-material-transaction-exports  Browse Haskell Lemon inbound material transaction exports
│   │   ├── list            List Haskell Lemon inbound material transaction exports with filtering
│   │   └── show <id>       Show Haskell Lemon inbound material transaction export details
│   ├── integration-configs  Browse integration configs
│   │   ├── list            List integration configs with filtering
│   │   └── show <id>       Show integration config details
│   ├── integration-exports  Browse integration exports
│   │   ├── list            List integration exports with filtering
│   │   └── show <id>       Show integration export details
│   ├── material-transaction-diversions  Browse material transaction diversions
│   │   ├── list            List material transaction diversions with filtering
│   │   └── show <id>       Show diversion details
│   ├── material-transaction-shift-assignments  Browse material transaction shift assignments
│   │   ├── list            List material transaction shift assignments with filtering
│   │   └── show <id>       Show assignment details
│   ├── raw-material-transaction-sales-customers  Browse raw material transaction sales customers
│   │   ├── list            List raw material transaction sales customers with filtering
│   │   └── show <id>       Show raw material transaction sales customer details
│   ├── raw-transport-tractors  Browse raw transport tractors
│   │   ├── list            List raw transport tractors with filtering
│   │   └── show <id>       Show raw transport tractor details
│   ├── raw-transport-exports  Browse raw transport exports
│   │   ├── list            List raw transport exports with filtering
│   │   └── show <id>       Show raw transport export details
│   ├── raw-records         Browse ingest raw records
│   │   ├── list            List raw records with filtering
│   │   └── show <id>       Show raw record details
│   ├── transport-order-materials  Browse transport order materials
│   │   ├── list            List transport order materials with filtering
│   │   └── show <id>       Show transport order material details
│   ├── transport-references  Browse transport references
│   │   ├── list            List transport references with filtering
│   │   └── show <id>       Show transport reference details
│   ├── service-type-unit-of-measure-quantities  Browse service type unit of measure quantities
│   │   ├── list            List service type unit of measure quantities with filtering
│   │   └── show <id>       Show service type unit of measure quantity details
│   ├── customers           Browse customers
│   │   └── list            List customers with filtering
│   ├── commitment-simulation-sets  Browse commitment simulation sets
│   │   ├── list            List commitment simulation sets with filtering
│   │   └── show <id>       Show commitment simulation set details
│   ├── commitment-simulation-periods  Browse commitment simulation periods
│   │   ├── list            List commitment simulation periods with filtering
│   │   └── show <id>       Show commitment simulation period details
│   ├── truckers            Browse trucking companies
│   │   └── list            List truckers with filtering
│   ├── trucker-applications  Browse trucker applications
│   │   ├── list            List trucker applications with filtering
│   │   └── show <id>       Show trucker application details
│   ├── trucker-referral-codes  Browse trucker referral codes
│   │   ├── list            List trucker referral codes with filtering
│   │   └── show <id>       Show trucker referral code details
│   ├── customer-memberships  Browse customer memberships
│   │   ├── list            List customer memberships with filtering
│   │   └── show <id>       Show customer membership details
│   ├── customer-truckers     Browse customer trucker links
│   │   ├── list            List customer truckers with filtering
│   │   └── show <id>       Show customer trucker details
│   ├── developer-certified-weighers  Browse developer certified weighers
│   │   ├── list            List developer certified weighers with filtering
│   │   └── show <id>       Show developer certified weigher details
│   ├── developer-trucker-certifications  Browse developer trucker certifications
│   │   ├── list            List developer trucker certifications with filtering
│   │   └── show <id>       Show developer trucker certification details
│   ├── deere-equipments     Browse Deere equipment
│   │   ├── list            List Deere equipment with filtering
│   │   └── show <id>       Show Deere equipment details
│   ├── digital-fleet-trucks  Browse digital fleet trucks
│   │   ├── list            List digital fleet trucks with filtering
│   │   └── show <id>       Show digital fleet truck details
│   ├── go-motive-integrations  Browse GoMotive integrations
│   │   ├── list            List GoMotive integrations with filtering
│   │   └── show <id>       Show GoMotive integration details
│   ├── keep-truckin-vehicles  Browse KeepTruckin vehicles
│   │   ├── list            List KeepTruckin vehicles with filtering
│   │   └── show <id>       Show KeepTruckin vehicle details
│   ├── tenna-vehicles       Browse Tenna vehicles
│   │   ├── list            List Tenna vehicles with filtering
│   │   └── show <id>       Show Tenna vehicle details
│   ├── teletrac-navman-vehicles  Browse Teletrac Navman vehicles
│   │   ├── list            List Teletrac Navman vehicles with filtering
│   │   └── show <id>       Show Teletrac Navman vehicle details
│   ├── memberships         Browse user-organization memberships
│   │   ├── list            List memberships with filtering
│   │   └── show <id>       Show membership details
│   ├── geofence-restriction-violations  Browse geofence restriction violations
│   │   ├── list            List geofence restriction violations
│   │   └── show <id>       Show geofence restriction violation details
│   ├── hos-day-regulation-sets  Browse HOS day regulation sets
│   │   ├── list            List HOS day regulation sets
│   │   └── show <id>       Show HOS day regulation set details
│   ├── features            Browse product features
│   │   ├── list            List features with filtering
│   │   └── show <id>       Show feature details
│   ├── release-notes       Browse release notes
│   │   ├── list            List release notes with filtering
│   │   └── show <id>       Show release note details
│   ├── press-releases      Browse press releases
│   │   ├── list            List press releases
│   │   └── show <id>       Show press release details
│   ├── place-predictions   Browse place predictions
│   │   └── list            List place predictions by query
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

### Post Children

```bash
# Link a child post to a parent post
xbe do post-children create --parent-post 123 --child-post 456

# List post child links for a parent post
xbe view post-children list --parent-post 123

# Show post child link details
xbe view post-children show 789

# Delete a post child link
xbe do post-children delete 789 --confirm
```

### Post Views

```bash
# List post views
xbe view post-views list

# Filter by post and viewer
xbe view post-views list --post 123 --viewer 456

# Show post view details
xbe view post-views show 789

# Record a post view
xbe do post-views create --post 123 --viewer 456 --viewed-at 2025-01-01T12:00:00Z
```

### Follows

```bash
# List follows
xbe view follows list

# Filter by follower
xbe view follows list --follower 123

# Follow a creator
xbe do follows create --follower 123 --creator-type projects --creator-id 456
```

### Marketing Metrics

```bash
# Refresh the marketing metrics snapshot
xbe do marketing-metrics create

# View the latest snapshot in a table
xbe view marketing-metrics list

# View full metrics details
xbe view marketing-metrics show

# Get JSON output for scripting
xbe view marketing-metrics show --json
```

### Incident Headline Suggestions

```bash
# List recent incident headline suggestions
xbe view incident-headline-suggestions list --limit 10

# Filter by incident
xbe view incident-headline-suggestions list --incident 123

# Create a headline suggestion
xbe do incident-headline-suggestions create --incident 123

# Create with custom options
xbe do incident-headline-suggestions create --incident 123 --options '{"temperature":0.4,"max_tokens":256}'

# Show full suggestion details
xbe view incident-headline-suggestions show 456
```

### Prompt Prescriptions

```bash
# List prompt prescriptions
xbe view prompt-prescriptions list --limit 10

# Filter by email address
xbe view prompt-prescriptions list --email-address "name@example.com"

# Submit a prompt prescription request
xbe do prompt-prescriptions create \
  --email-address "name@example.com" \
  --name "Alex Builder" \
  --organization-name "Concrete Co" \
  --location-name "Austin, TX" \
  --role "Operations Manager" \
  --symptoms "Rising costs and scheduling delays"

# Show full prompt prescription details
xbe view prompt-prescriptions show 789
```

### Incident Tag Incidents

```bash
# List incident tag incident links
xbe view incident-tag-incidents list --limit 10

# Filter by incident
xbe view incident-tag-incidents list --incident 123

# Filter by incident tag
xbe view incident-tag-incidents list --incident-tag 456

# Create a tag link
xbe do incident-tag-incidents create --incident 123 --incident-tag 456

# Delete a tag link
xbe do incident-tag-incidents delete 789 --confirm
```

### Safety Incidents

```bash
# List safety incidents
xbe view safety-incidents list --limit 10

# Filter by status
xbe view safety-incidents list --status open

# Show safety incident details
xbe view safety-incidents show 123

# Create a safety incident
xbe do safety-incidents create \
  --subject-type brokers \
  --subject-id 123 \
  --start-at 2025-01-15T10:00:00Z \
  --status open \
  --kind near_miss \
  --headline "Near miss at plant"

# Update a safety incident
xbe do safety-incidents update 123 --status closed --end-at 2025-01-15T12:00:00Z

# Delete a safety incident
xbe do safety-incidents delete 123 --confirm
```

### Production Incident Detectors

```bash
# Run detection for a job production plan
xbe do production-incident-detectors create --job-production-plan 123

# Run with custom thresholds
xbe do production-incident-detectors create \
  --job-production-plan 123 \
  --lookahead-offset 30 \
  --minutes-threshold 45 \
  --quantity-threshold 50

# List detector runs
xbe view production-incident-detectors list --limit 10

# Show detector run details
xbe view production-incident-detectors show 456
```

### Incident Request Cancellations

```bash
# Cancel an incident request
xbe do incident-request-cancellations create --incident-request 123 --comment "No longer needed"
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

### Material Purchase Order Release Redemptions

```bash
# List release redemptions
xbe view material-purchase-order-release-redemptions list --limit 5

# Filter by release or ticket number
xbe view material-purchase-order-release-redemptions list --release 123
xbe view material-purchase-order-release-redemptions list --ticket-number T-100

# Show a redemption
xbe view material-purchase-order-release-redemptions show 456

# Create a redemption
xbe do material-purchase-order-release-redemptions create --release 123 --ticket-number T-100
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

### Material Transaction Rate Summary

```bash
# Hourly rate summary for a material site
xbe do material-transaction-rate-summaries create \
  --material-site 123 \
  --start-at 2025-01-01T00:00:00Z \
  --end-at 2025-01-02T00:00:00Z

# Filter by material type hierarchy and return sparse results
xbe do material-transaction-rate-summaries create \
  --material-site 123 \
  --material-type-hierarchies "aggregate,asphalt" \
  --start-at 2025-01-01T00:00:00Z \
  --end-at 2025-01-02T00:00:00Z \
  --sparse
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

### Customer Memberships

Customer memberships focus on customer organizations only.

```bash
# List customer memberships for a customer
xbe view customer-memberships list --customer 123

# Show customer membership details
xbe view customer-memberships show 456

# Create a customer membership
xbe do customer-memberships create --user 123 --customer 456 --kind manager

# Update a customer membership
xbe do customer-memberships update 789 --title "Dispatcher" --is-admin true

# Delete a customer membership (requires --confirm)
xbe do customer-memberships delete 789 --confirm
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

### Crew Requirements

Crew requirements schedule labor or equipment needs on job production plans.

```bash
# List requirements
xbe view crew-requirements list

# Filter by job production plan
xbe view crew-requirements list --job-production-plan 123

# Show requirement details
xbe view crew-requirements show 456

# Create an equipment requirement
xbe do crew-requirements create \
  --requirement-type equipment \
  --job-production-plan 123 \
  --resource-classification-type equipment-classifications \
  --resource-classification-id 456 \
  --resource-type equipment \
  --resource-id 789 \
  --start-at "2025-01-01T08:00:00Z" \
  --end-at "2025-01-01T16:00:00Z"

# Update a requirement
xbe do crew-requirements update 456 --note "Updated note" --requires-inbound-movement true

# Delete a requirement (requires --confirm)
xbe do crew-requirements delete 456 --confirm
```

### Equipment Requirements

Equipment requirements manage equipment-specific crew requirements.

```bash
# List equipment requirements
xbe view equipment-requirements list

# Filter by assignment candidate
xbe view equipment-requirements list --is-assignment-candidate-for 123

# Show equipment requirement details
xbe view equipment-requirements show 456

# Create an equipment requirement
xbe do equipment-requirements create \
  --job-production-plan 123 \
  --resource-classification-type equipment-classifications \
  --resource-classification-id 456 \
  --resource-type equipment \
  --resource-id 789 \
  --start-at "2025-01-01T08:00:00Z" \
  --end-at "2025-01-01T16:00:00Z"

# Update an equipment requirement
xbe do equipment-requirements update 456 --note "Updated note" --requires-inbound-movement true

# Delete an equipment requirement (requires --confirm)
xbe do equipment-requirements delete 456 --confirm
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

### Job Production Plan Broadcast Messages

Broadcast messages notify participants on a job production plan.

```bash
# List messages for a job production plan
xbe view job-production-plan-broadcast-messages list --job-production-plan 123

# Include hidden messages
xbe view job-production-plan-broadcast-messages list --job-production-plan 123 --is-hidden true

# Show message details
xbe view job-production-plan-broadcast-messages show 456

# Create a broadcast message
xbe do job-production-plan-broadcast-messages create \
  --job-production-plan 123 \
  --message "Crew arrival moved to 7:30 AM" \
  --summary "Start time update"

# Hide a broadcast message
xbe do job-production-plan-broadcast-messages update 456 --is-hidden
```

### Job Production Plan Safety Risk Communication Suggestions

Safety risk communication suggestions generate draft plans for communicating
job safety risks and remediation strategies.

```bash
# List suggestions
xbe view job-production-plan-safety-risk-communication-suggestions list

# Filter by job production plan
xbe view job-production-plan-safety-risk-communication-suggestions list --job-production-plan 123

# Show suggestion details
xbe view job-production-plan-safety-risk-communication-suggestions show 456

# Create a suggestion (async by default)
xbe do job-production-plan-safety-risk-communication-suggestions create \
  --job-production-plan 123

# Create synchronously with options
xbe do job-production-plan-safety-risk-communication-suggestions create \
  --job-production-plan 123 \
  --is-async=false \
  --options '{"temperature":0.2}'

# Delete a suggestion
xbe do job-production-plan-safety-risk-communication-suggestions delete 456 --confirm
```

### Job Production Plan Recaps

Job production plan recaps provide generated markdown summaries for a plan.

```bash
# List recaps
xbe view job-production-plan-recaps list

# Filter by job production plan
xbe view job-production-plan-recaps list --plan 123

# Show recap details
xbe view job-production-plan-recaps show 456
```

### Job Production Plan Material Site Changes

Material site changes swap material sites (and optionally material types) on a job production plan.

```bash
# List material site changes
xbe view job-production-plan-material-site-changes list

# Show material site change details
xbe view job-production-plan-material-site-changes show 123

# Swap material sites
xbe do job-production-plan-material-site-changes create \
  --job-production-plan 123 \
  --old-material-site 456 \
  --new-material-site 789

# Swap material site and material type
xbe do job-production-plan-material-site-changes create \
  --job-production-plan 123 \
  --old-material-site 456 \
  --new-material-site 789 \
  --old-material-type 111 \
  --new-material-type 222
```

### Job Production Plan Status Changes

Status changes record lifecycle transitions for job production plans.

```bash
# List recent status changes
xbe view job-production-plan-status-changes list

# Show status change details
xbe view job-production-plan-status-changes show 123

# Set a cancellation reason type
xbe do job-production-plan-status-changes update 123 \
  --job-production-plan-cancellation-reason-type 456
```

### Job Production Plan Duplications

Job production plan duplications copy a template into a new plan or template.

```bash
# Duplicate a template into a new plan
xbe do job-production-plan-duplications create \
  --job-production-plan-template 123 \
  --start-on 2026-01-23

# Duplicate into another customer and skip copying shifts
xbe do job-production-plan-duplications create \
  --job-production-plan-template 123 \
  --start-on 2026-01-23 \
  --new-customer 456 \
  --skip-job-schedule-shifts
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
