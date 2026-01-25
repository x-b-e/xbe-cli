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
│   ├── action-item-line-items Manage action item line items
│   │   ├── create           Create an action item line item
│   │   ├── update           Update an action item line item
│   │   └── delete           Delete an action item line item
│   ├── application-settings Manage global application settings
│   │   ├── create           Create an application setting
│   │   ├── update           Update an application setting
│   │   └── delete           Delete an application setting
│   ├── administrative-incidents Manage administrative incidents
│   │   ├── create           Create an administrative incident
│   │   ├── update           Update an administrative incident
│   │   └── delete           Delete an administrative incident
│   ├── efficiency-incidents Manage efficiency incidents
│   │   ├── create           Create an efficiency incident
│   │   ├── update           Update an efficiency incident
│   │   └── delete           Delete an efficiency incident
│   ├── incident-participants Manage incident participants
│   │   ├── create           Create an incident participant
│   │   ├── update           Update an incident participant
│   │   └── delete           Delete an incident participant
│   ├── incident-requests     Manage incident requests
│   │   ├── create           Create an incident request
│   │   ├── update           Update an incident request
│   │   └── delete           Delete an incident request
│   ├── customer-incident-default-assignees Manage customer incident default assignees
│   │   ├── create           Create a customer incident default assignee
│   │   ├── update           Update a customer incident default assignee
│   │   └── delete           Delete a customer incident default assignee
│   ├── deere-integrations  Manage Deere integrations
│   │   ├── create           Create a Deere integration
│   │   ├── update           Update a Deere integration
│   │   └── delete           Delete a Deere integration
│   ├── samsara-integrations Manage Samsara integrations
│   │   ├── create           Create a Samsara integration
│   │   ├── update           Update a Samsara integration
│   │   └── delete           Delete a Samsara integration
│   ├── exporter-configurations Manage exporter configurations
│   │   ├── create           Create an exporter configuration
│   │   ├── update           Update an exporter configuration
│   │   └── delete           Delete an exporter configuration
│   ├── importer-configurations Manage importer configurations
│   │   ├── create           Create an importer configuration
│   │   ├── update           Update an importer configuration
│   │   └── delete           Delete an importer configuration
│   ├── device-diagnostics  Manage device diagnostics
│   │   └── create           Create a device diagnostic
│   ├── file-attachment-signed-urls Generate signed URLs for file attachments
│   │   └── create           Generate a signed URL for a file attachment
│   ├── login-code-requests  Request login codes
│   │   └── create           Request a login code
│   ├── post-routers         Manage post routers
│   │   └── create           Create a post router
│   ├── user-device-location-tracking-requests Request user device location tracking
│   │   └── create           Send a location tracking request
│   ├── user-location-requests Request user location
│   │   └── create           Create a user location request
│   ├── user-post-feed-posts Manage user post feed posts
│   │   └── update           Update a user post feed post
│   ├── saml-code-redemptions Redeem SAML login codes
│   │   └── create           Redeem a SAML login code
│   ├── sourcing-searches    Find matching truckers and trailers for a customer tender
│   │   └── create           Run a sourcing search
│   ├── file-imports         Manage file imports
│   │   ├── create           Create a file import
│   │   ├── update           Update a file import
│   │   └── delete           Delete a file import
│   ├── ticket-reports       Manage ticket reports
│   │   ├── create           Create a ticket report
│   │   ├── update           Update a ticket report
│   │   └── delete           Delete a ticket report
│   ├── organization-invoices-batches Manage organization invoices batches
│   │   └── create           Create an organization invoices batch
│   ├── organization-invoices-batch-files Manage organization invoices batch files
│   │   └── create           Create an organization invoices batch file
│   ├── organization-invoices-batch-invoice-unbatchings Unbatch organization invoices batch invoices
│   │   └── create           Unbatch an organization invoices batch invoice
│   ├── organization-invoices-batch-pdf-generations Manage organization invoices batch PDF generations
│   │   └── create           Create an organization invoices batch PDF generation
│   ├── glossary-terms       Manage glossary terms
│   │   ├── create           Create a glossary term
│   │   ├── update           Update a glossary term
│   │   └── delete           Delete a glossary term
│   ├── native-app-releases  Manage native app releases
│   │   ├── create           Create a native app release
│   │   ├── update           Update a native app release
│   │   └── delete           Delete a native app release
│   ├── gauge-vehicles       Manage gauge vehicles
│   │   └── update           Update gauge vehicle assignments
│   ├── geotab-vehicles      Manage geotab vehicles
│   │   └── update           Update geotab vehicle assignments
│   ├── gps-insight-vehicles Manage GPS Insight vehicles
│   │   └── update           Update GPS Insight vehicle assignments
│   ├── one-step-gps-vehicles Manage One Step GPS vehicles
│   │   └── update           Update One Step GPS vehicle assignments
│   ├── samsara-vehicles     Manage samsara vehicles
│   │   └── update           Update samsara vehicle assignments
│   ├── temeda-vehicles      Manage temeda vehicles
│   │   └── update           Update temeda vehicle assignments
│   ├── t3-equipmentshare-vehicles Manage T3 EquipmentShare vehicles
│   │   └── update           Update T3 EquipmentShare vehicle assignments
│   ├── verizon-reveal-vehicles Manage Verizon Reveal vehicles
│   │   └── update           Update Verizon Reveal vehicle assignments
│   ├── keep-truckin-users  Manage KeepTruckin users
│   │   └── update           Update a KeepTruckin user assignment
│   ├── platform-statuses    Manage platform status updates
│   │   ├── create           Create a platform status
│   │   ├── update           Update a platform status
│   │   └── delete           Delete a platform status
│   ├── commitment-items     Manage commitment items
│   │   ├── create           Create a commitment item
│   │   ├── update           Update a commitment item
│   │   └── delete           Delete a commitment item
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
│   ├── invoice-approvals    Approve invoices
│   │   └── create           Approve an invoice
│   ├── invoice-sends Send invoices
│   │   └── create           Send an invoice
│   ├── invoice-pdf-emails   Email invoice PDFs
│   │   └── create           Email an invoice PDF
│   ├── time-sheet-approvals Approve time sheets
│   │   └── create           Approve a time sheet
│   ├── time-sheet-unapprovals Unapprove time sheets
│   │   └── create           Unapprove a time sheet
│   ├── time-sheet-no-shows  Manage time sheet no-shows
│   │   ├── create           Create a time sheet no-show
│   │   ├── update           Update a time sheet no-show
│   │   └── delete           Delete a time sheet no-show
│   ├── resource-unavailabilities Manage resource unavailabilities
│   │   ├── create           Create a resource unavailability
│   │   ├── update           Update a resource unavailability
│   │   └── delete           Delete a resource unavailability
│   ├── tractor-odometer-readings Manage tractor odometer readings
│   │   ├── create           Create a tractor odometer reading
│   │   ├── update           Update a tractor odometer reading
│   │   └── delete           Delete a tractor odometer reading
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
│   ├── project-bid-location-material-types Manage project bid location material types
│   │   ├── create           Create a project bid location material type
│   │   ├── update           Update a project bid location material type
│   │   └── delete           Delete a project bid location material type
│   ├── project-material-types Manage project material types
│   │   ├── create           Create a project material type
│   │   ├── update           Update a project material type
│   │   └── delete           Delete a project material type
│   ├── project-revenue-items Manage project revenue items
│   │   ├── create           Create a project revenue item
│   │   ├── update           Update a project revenue item
│   │   └── delete           Delete a project revenue item
│   ├── project-customers   Manage project customers
│   │   ├── create           Create a project customer
│   │   └── delete           Delete a project customer
│   ├── project-truckers   Manage project truckers
│   │   ├── create           Create a project trucker
│   │   ├── update           Update a project trucker
│   │   └── delete           Delete a project trucker
│   ├── project-transport-organizations Manage project transport organizations
│   │   ├── create           Create a project transport organization
│   │   ├── update           Update a project transport organization
│   │   └── delete           Delete a project transport organization
│   ├── project-transport-plan-events Manage project transport plan events
│   │   ├── create           Create a project transport plan event
│   │   ├── update           Update a project transport plan event
│   │   └── delete           Delete a project transport plan event
│   ├── project-transport-plan-stop-insertions Manage project transport plan stop insertions
│   │   └── create           Create a project transport plan stop insertion
│   ├── project-transport-plan-segment-drivers Manage project transport plan segment drivers
│   │   ├── create           Create a project transport plan segment driver
│   │   └── delete           Delete a project transport plan segment driver
│   ├── project-transport-plan-segment-trailers Manage project transport plan segment trailers
│   │   ├── create           Create a project transport plan segment trailer
│   │   └── delete           Delete a project transport plan segment trailer
│   ├── project-project-cost-classifications Manage project project cost classifications
│   │   ├── create           Create a project project cost classification
│   │   ├── update           Update a project project cost classification
│   │   └── delete           Delete a project project cost classification
│   ├── project-import-file-verifications Verify project import files
│   │   └── create           Create a project import file verification
│   ├── geofence-restrictions Manage geofence restrictions
│   │   ├── create           Create a geofence restriction
│   │   ├── update           Update a geofence restriction
│   │   └── delete           Delete a geofence restriction
│   ├── lane-summary         Generate lane (cycle) summaries
│   │   └── create           Create a lane summary
│   ├── job-production-plan-abandonments Abandon job production plans
│   │   └── create           Abandon a job production plan
│   ├── job-production-plan-completions Complete job production plans
│   │   └── create           Complete a job production plan
│   ├── job-production-plan-unscrappages Unscrap job production plans
│   │   └── create           Unscrap a job production plan
│   ├── key-result-unscrappages Unscrap key results
│   │   └── create           Unscrap a key result
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
│   ├── resource-classification-project-cost-classifications Manage resource classification project cost classifications
│   │   ├── create           Create a resource classification project cost classification
│   │   └── delete           Delete a resource classification project cost classification
│   ├── objective-stakeholder-classifications Manage objective stakeholder classifications
│   │   ├── create           Create an objective stakeholder classification
│   │   ├── update           Update an objective stakeholder classification
│   │   └── delete           Delete an objective stakeholder classification
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
│   ├── prediction-knowledge-bases Manage prediction knowledge bases
│   │   └── create           Create a prediction knowledge base
│   ├── prediction-subjects Manage prediction subjects
│   │   ├── create           Create a prediction subject
│   │   ├── update           Update a prediction subject
│   │   └── delete           Delete a prediction subject
│   ├── prediction-subject-recap-generations Generate prediction subject recaps
│   │   └── create           Generate a prediction subject recap
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
│   ├── production-measurements Manage production measurements
│   │   ├── create           Create a production measurement
│   │   ├── update           Update a production measurement
│   │   └── delete           Delete a production measurement
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
│   ├── trucker-invoices    Manage trucker invoices
│   │   ├── create           Create a trucker invoice
│   │   ├── update           Update a trucker invoice
│   │   └── delete           Delete a trucker invoice
│   ├── trucker-shift-sets  Manage trucker shift sets
│   │   └── update           Update a trucker shift set
│   ├── time-card-time-changes Manage time card time changes
│   │   ├── create           Create a time card time change
│   │   ├── update           Update a time card time change
│   │   └── delete           Delete a time card time change
│   ├── time-sheet-line-items Manage time sheet line items
│   │   ├── create           Create a time sheet line item
│   │   ├── update           Update a time sheet line item
│   │   └── delete           Delete a time sheet line item
│   ├── lineups             Manage lineups
│   │   ├── create           Create a lineup
│   │   ├── update           Update a lineup
│   │   └── delete           Delete a lineup
│   ├── meetings           Manage meetings
│   │   ├── create           Create a meeting
│   │   ├── update           Update a meeting
│   │   └── delete           Delete a meeting
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
│   ├── material-transaction-field-scopes Manage material transaction field scopes
│   │   └── create           Create a material transaction field scope
│   ├── tender-job-schedule-shifts-material-transactions-checksums Generate tender job schedule shift material transaction checksums
│   │   └── create           Create a checksum
│   ├── tender-rejections    Reject tenders
│   │   └── create           Reject a tender
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
│   ├── profit-improvement-subscriptions Manage profit improvement subscriptions
│   │   ├── create           Create a profit improvement subscription
│   │   ├── update           Update a profit improvement subscription
│   │   └── delete           Delete a profit improvement subscription
│   ├── memberships          Manage user-organization memberships
│   │   ├── create           Create a membership
│   │   ├── update           Update a membership
│   │   └── delete           Delete a membership
│   ├── retainers            Manage retainers
│   │   ├── create           Create a retainer
│   │   ├── update           Update a retainer
│   │   └── delete           Delete a retainer
│   ├── rate-agreements-copiers Copy rate agreements from a template
│   │   └── create           Copy a template rate agreement
│   ├── work-order-service-codes Manage work order service codes
│   │   ├── create           Create a work order service code
│   │   ├── update           Update a work order service code
│   │   └── delete           Delete a work order service code
│   └── transport-order-stop-materials Manage transport order stop materials
│       ├── create           Create a transport order stop material
│       ├── update           Update a transport order stop material
│       └── delete           Delete a transport order stop material
├── view                    Browse and view XBE content
│   ├── action-item-line-items Browse action item line items
│   │   ├── list            List action item line items
│   │   └── show <id>       Show action item line item details
│   ├── application-settings Browse application settings
│   │   ├── list            List application settings
│   │   └── show <id>       Show application setting details
│   ├── commitment-items    Browse commitment items
│   │   ├── list            List commitment items
│   │   └── show <id>       Show commitment item details
│   ├── commitments         Browse commitments
│   │   ├── list            List commitments
│   │   └── show <id>       Show commitment details
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
│   ├── invoice-approvals   Browse invoice approvals
│   │   ├── list            List invoice approvals
│   │   └── show <id>       Show invoice approval details
│   ├── invoice-sends       Browse invoice sends
│   │   ├── list            List invoice sends
│   │   └── show <id>       Show invoice send details
│   ├── invoice-status-changes Browse invoice status changes
│   │   ├── list            List invoice status changes
│   │   └── show <id>       Show invoice status change details
│   ├── invoice-revisionizing-invoice-revisions Browse invoice revisionizing invoice revisions
│   │   ├── list            List invoice revisionizing invoice revisions
│   │   └── show <id>       Show invoice revisionizing invoice revision details
│   ├── time-sheet-approvals Browse time sheet approvals
│   │   ├── list            List time sheet approvals
│   │   └── show <id>       Show time sheet approval details
│   ├── time-sheet-unapprovals Browse time sheet unapprovals
│   │   ├── list            List time sheet unapprovals
│   │   └── show <id>       Show time sheet unapproval details
│   ├── time-sheet-no-shows Browse time sheet no-shows
│   │   ├── list            List time sheet no-shows
│   │   └── show <id>       Show time sheet no-show details
│   ├── time-sheet-status-changes Browse time sheet status changes
│   │   ├── list            List time sheet status changes
│   │   └── show <id>       Show time sheet status change details
│   ├── resource-unavailabilities Browse resource unavailabilities
│   │   ├── list            List resource unavailabilities
│   │   └── show <id>       Show resource unavailability details
│   ├── retainers           Browse retainers
│   │   ├── list            List retainers
│   │   └── show <id>       Show retainer details
│   ├── rate-agreements-copiers Browse rate agreements copiers
│   │   ├── list            List rate agreements copiers
│   │   └── show <id>       Show rate agreements copier details
│   ├── tractor-odometer-readings Browse tractor odometer readings
│   │   ├── list            List tractor odometer readings
│   │   └── show <id>       Show tractor odometer reading details
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
│   ├── project-bid-location-material-types Browse project bid location material types
│   │   ├── list            List project bid location material types
│   │   └── show <id>       Show project bid location material type details
│   ├── project-material-types Browse project material types
│   │   ├── list            List project material types
│   │   └── show <id>       Show project material type details
│   ├── project-revenue-items Browse project revenue items
│   │   ├── list            List project revenue items
│   │   └── show <id>       Show project revenue item details
│   ├── project-customers   Browse project customers
│   │   ├── list            List project customers
│   │   └── show <id>       Show project customer details
│   ├── project-truckers   Browse project truckers
│   │   ├── list            List project truckers
│   │   └── show <id>       Show project trucker details
│   ├── project-transport-organizations Browse project transport organizations
│   │   ├── list            List project transport organizations
│   │   └── show <id>       Show project transport organization details
│   ├── project-transport-plan-events Browse project transport plan events
│   │   ├── list            List project transport plan events
│   │   └── show <id>       Show project transport plan event details
│   ├── project-transport-plan-stop-insertions Browse project transport plan stop insertions
│   │   ├── list            List project transport plan stop insertions
│   │   └── show <id>       Show project transport plan stop insertion details
│   ├── project-transport-plan-segment-drivers Browse project transport plan segment drivers
│   │   ├── list            List project transport plan segment drivers
│   │   └── show <id>       Show project transport plan segment driver details
│   ├── project-transport-plan-segment-trailers Browse project transport plan segment trailers
│   │   ├── list            List project transport plan segment trailers
│   │   └── show <id>       Show project transport plan segment trailer details
│   ├── project-transport-plan-strategy-steps Browse project transport plan strategy steps
│   │   ├── list            List project transport plan strategy steps
│   │   └── show <id>       Show project transport plan strategy step details
│   ├── project-transport-plan-event-location-prediction-autopsies Browse project transport plan event location prediction autopsies
│   │   ├── list            List project transport plan event location prediction autopsies
│   │   └── show <id>       Show project transport plan event location prediction autopsy details
│   ├── project-project-cost-classifications Browse project project cost classifications
│   │   ├── list            List project project cost classifications
│   │   └── show <id>       Show project project cost classification details
│   ├── project-import-file-verifications Browse project import file verifications
│   │   └── list            List project import file verifications
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
│   ├── post-routers        Browse post routers
│   │   ├── list            List post routers with filtering
│   │   └── show <id>       Show post router details
│   ├── post-router-jobs    Browse post router jobs
│   │   ├── list            List post router jobs with filtering
│   │   └── show <id>       Show post router job details
│   ├── places              Lookup place details
│   │   └── show <place-id> Show place details
│   ├── proffer-likes       Browse proffer likes
│   │   ├── list            List proffer likes with filtering
│   │   └── show <id>       Show proffer like details
│   ├── public-praise-culture-values Browse public praise culture values
│   │   ├── list            List public praise culture values with filtering
│   │   └── show <id>       Show public praise culture value details
│   ├── brokers             Browse broker/branch information
│   │   └── list            List brokers with filtering
│   ├── administrative-incidents Browse administrative incidents
│   │   ├── list            List administrative incidents with filtering
│   │   └── show <id>       Show administrative incident details
│   ├── efficiency-incidents Browse efficiency incidents
│   │   ├── list            List efficiency incidents with filtering
│   │   └── show <id>       Show efficiency incident details
│   ├── incident-participants Browse incident participants
│   │   ├── list            List incident participants
│   │   └── show <id>       Show incident participant details
│   ├── incident-requests    Browse incident requests
│   │   ├── list            List incident requests
│   │   └── show <id>       Show incident request details
│   ├── customer-incident-default-assignees Browse customer incident default assignees
│   │   ├── list            List customer incident default assignees
│   │   └── show <id>       Show customer incident default assignee details
│   ├── device-diagnostics  Browse device diagnostics
│   │   ├── list            List device diagnostics
│   │   └── show <id>       Show device diagnostic details
│   ├── user-location-estimates Browse user location estimates
│   │   └── list            List user location estimates
│   ├── user-location-requests Browse user location requests
│   │   ├── list            List user location requests
│   │   └── show <id>       Show user location request details
│   ├── user-post-feed-posts Browse user post feed posts
│   │   ├── list            List user post feed posts
│   │   └── show <id>       Show user post feed post details
│   ├── file-imports        Browse file imports
│   │   ├── list            List file imports with filtering
│   │   └── show <id>       Show file import details
│   ├── ticket-reports      Browse ticket reports
│   │   ├── list            List ticket reports with filtering
│   │   └── show <id>       Show ticket report details
│   ├── organization-invoices-batches Browse organization invoices batches
│   │   ├── list            List organization invoices batches with filtering
│   │   └── show <id>       Show organization invoices batch details
│   ├── organization-invoices-batch-files Browse organization invoices batch files
│   │   ├── list            List organization invoices batch files with filtering
│   │   └── show <id>       Show organization invoices batch file details
│   ├── organization-invoices-batch-invoice-unbatchings Browse organization invoices batch invoice unbatchings
│   │   ├── list            List organization invoices batch invoice unbatchings
│   │   └── show <id>       Show organization invoices batch invoice unbatching details
│   ├── organization-invoices-batch-pdf-generations Browse organization invoices batch PDF generations
│   │   ├── list            List organization invoices batch PDF generations with filtering
│   │   ├── show <id>       Show organization invoices batch PDF generation details
│   │   └── download-all <id> Download all completed PDFs for a PDF generation
│   ├── integration-invoices-batch-exports Browse integration invoices batch exports
│   │   ├── list            List integration invoices batch exports with filtering
│   │   └── show <id>       Show integration invoices batch export details
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
│   ├── material-transaction-field-scopes Browse material transaction field scopes
│   │   ├── list            List material transaction field scopes
│   │   └── show <id>       Show material transaction field scope details
│   ├── tender-job-schedule-shifts-material-transactions-checksums Browse tender job schedule shift material transaction checksums
│   │   ├── list            List checksum records
│   │   └── show <id>       Show checksum details
│   ├── tender-rejections   Browse tender rejections
│   │   ├── list            List tender rejections
│   │   └── show <id>       Show tender rejection details
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
│   ├── profit-improvement-subscriptions Browse profit improvement subscriptions
│   │   ├── list            List profit improvement subscriptions
│   │   └── show <id>       Show profit improvement subscription details
│   ├── inventory-changes   Browse and view inventory changes
│   │   ├── list            List inventory changes with filtering
│   │   └── show <id>       Show inventory change details
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
│   ├── trucker-brokerages  Browse trucker brokerages
│   │   ├── list            List trucker brokerages
│   │   └── show <id>       Show trucker brokerage details
│   ├── trucker-referrals   Browse trucker referrals
│   │   ├── list            List trucker referrals
│   │   └── show <id>       Show trucker referral details
│   ├── trucker-shift-sets  Browse trucker shift sets
│   │   ├── list            List trucker shift sets with filtering
│   │   └── show <id>       Show trucker shift set details
│   ├── memberships         Browse user-organization memberships
│   │   ├── list            List memberships with filtering
│   │   └── show <id>       Show membership details
│   ├── transport-order-stop-materials Browse transport order stop materials
│   │   ├── list            List transport order stop materials
│   │   └── show <id>       Show transport order stop material details
│   ├── work-order-service-codes Browse work order service codes
│   │   ├── list            List work order service codes
│   │   └── show <id>       Show work order service code details
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
│   ├── version-events      Browse version events
│   │   ├── list            List version events with filtering
│   │   └── show <id>       Show version event details
│   ├── gauge-vehicles       Browse gauge vehicles
│   │   ├── list            List gauge vehicles with filtering
│   │   └── show <id>       Show gauge vehicle details
│   ├── geotab-vehicles      Browse geotab vehicles
│   │   ├── list            List geotab vehicles with filtering
│   │   └── show <id>       Show geotab vehicle details
│   ├── gps-insight-vehicles Browse GPS Insight vehicles
│   │   ├── list            List GPS Insight vehicles with filtering
│   │   └── show <id>       Show GPS Insight vehicle details
│   ├── one-step-gps-vehicles Browse One Step GPS vehicles
│   │   ├── list            List One Step GPS vehicles with filtering
│   │   └── show <id>       Show One Step GPS vehicle details
│   ├── samsara-vehicles     Browse samsara vehicles
│   │   ├── list            List samsara vehicles with filtering
│   │   └── show <id>       Show samsara vehicle details
│   ├── temeda-vehicles      Browse temeda vehicles
│   │   ├── list            List temeda vehicles with filtering
│   │   └── show <id>       Show temeda vehicle details
│   ├── t3-equipmentshare-vehicles Browse T3 EquipmentShare vehicles
│   │   ├── list            List T3 EquipmentShare vehicles with filtering
│   │   └── show <id>       Show T3 EquipmentShare vehicle details
│   ├── verizon-reveal-vehicles Browse Verizon Reveal vehicles
│   │   ├── list            List Verizon Reveal vehicles with filtering
│   │   └── show <id>       Show Verizon Reveal vehicle details
│   ├── keep-truckin-users  Browse KeepTruckin users
│   │   ├── list            List KeepTruckin users with filtering
│   │   └── show <id>       Show KeepTruckin user details
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

### Project Material Types

```bash
# List material types for a project
xbe view project-material-types list --project 123

# Filter by pickup-at-min window
xbe view project-material-types list \
  --pickup-at-min-min 2026-01-01T00:00:00Z \
  --pickup-at-min-max 2026-01-02T00:00:00Z
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

### Broker Tenders

```bash
# List broker tenders
xbe view broker-tenders list

# Filter by broker, job, or status
xbe view broker-tenders list --broker 123 --status editing
xbe view broker-tenders list --job 456

# Create a broker tender
xbe do broker-tenders create --job 456 --broker 123 --trucker 789

# Show broker tender details
xbe view broker-tenders show 789
```

### Broker Tender Cancelled Seller Notifications

```bash
# List notifications
xbe view broker-tender-cancelled-seller-notifications list

# Filter by read status
xbe view broker-tender-cancelled-seller-notifications list --read false

# Show notification details
xbe view broker-tender-cancelled-seller-notifications show 123

# Mark as read
xbe do broker-tender-cancelled-seller-notifications update 123 --read
```

### Broker Tender Returned Buyer Notifications

```bash
# List notifications
xbe view broker-tender-returned-buyer-notifications list

# Filter by read status
xbe view broker-tender-returned-buyer-notifications list --read false

# Show notification details
xbe view broker-tender-returned-buyer-notifications show 123

# Mark as read
xbe do broker-tender-returned-buyer-notifications update 123 --read
```

### Customer Tender Offered Buyer Notifications

```bash
# List notifications
xbe view customer-tender-offered-buyer-notifications list

# Filter by read status
xbe view customer-tender-offered-buyer-notifications list --read false

# Show notification details
xbe view customer-tender-offered-buyer-notifications show 123

# Mark as read
xbe do customer-tender-offered-buyer-notifications update 123 --read
```

### Tender Job Schedule Shift Cancelled Trucker Contact Notifications

```bash
# List notifications
xbe view tender-job-schedule-shift-cancelled-trucker-contact-notifications list

# Filter by read status
xbe view tender-job-schedule-shift-cancelled-trucker-contact-notifications list --read false

# Show notification details
xbe view tender-job-schedule-shift-cancelled-trucker-contact-notifications show 123

# Mark as read
xbe do tender-job-schedule-shift-cancelled-trucker-contact-notifications update 123 --read
```

### Tender Job Schedule Shift Starting Seller Notifications

```bash
# List notifications
xbe view tender-job-schedule-shift-starting-seller-notifications list

# Filter by read status
xbe view tender-job-schedule-shift-starting-seller-notifications list --read false

# Show notification details
xbe view tender-job-schedule-shift-starting-seller-notifications show 123

# Mark as read
xbe do tender-job-schedule-shift-starting-seller-notifications update 123 --read
```

### Notification Subscriptions

```bash
# List notification subscriptions
xbe view notification-subscriptions list

# Filter by user
xbe view notification-subscriptions list --user 123

# Show notification subscription details
xbe view notification-subscriptions show 456
```

### Site Wait Time Notification Triggers

```bash
# List site wait time notification triggers
xbe view site-wait-time-notification-triggers list

# Filter by job production plan
xbe view site-wait-time-notification-triggers list --job-production-plan 123

# Filter by site type
xbe view site-wait-time-notification-triggers list --site-type job_site

# Show trigger details
xbe view site-wait-time-notification-triggers show 456
```

### OpenAI Vector Stores

```bash
# List vector stores
xbe view open-ai-vector-stores list

# Filter by purpose
xbe view open-ai-vector-stores list --purpose user_post_feed

# Filter by scope
xbe view open-ai-vector-stores list --scope "UserPostFeed|123"

# Show vector store details
xbe view open-ai-vector-stores show 789
```

### Prompters

```bash
# List prompters
xbe view prompters list

# Filter by active status
xbe view prompters list --is-active true

# Show prompter details
xbe view prompters show 123

# Create a prompter
xbe do prompters create --name "Release Notes" --is-active=false

# Update a prompt template
xbe do prompters update 123 --release-note-headline-suggestions-prompt-template "New template"

# Delete a prompter
xbe do prompters delete 123 --confirm
```

### Communications

```bash
# List communications
xbe view communications list

# Filter by subject or delivery status
xbe view communications list --subject-type Project --subject-id 123
xbe view communications list --delivery-status incoming_received

# Show communication details
xbe view communications show 123
```

### Text Messages

```bash
# List today's text messages
xbe view text-messages list

# Filter by recipient and date
xbe view text-messages list --to +15551234567 --date-sent 2025-01-20

# Show text message details
xbe view text-messages show SM123
```

### Tenders

```bash
# List tenders
xbe view tenders list

# Filter by buyer, job number, or status
xbe view tenders list --buyer 123 --status editing
xbe view tenders list --job-number "JOB-1001"

# Show tender details
xbe view tenders show 789
```

### Customer Tenders

```bash
# List customer tenders
xbe view customer-tenders list

# Filter by broker, job, or status
xbe view customer-tenders list --broker 123 --status editing
xbe view customer-tenders list --job 456

# Create a customer tender
xbe do customer-tenders create --job 456 --customer 123 --broker 789

# Show customer tender details
xbe view customer-tenders show 789
```

### Broker Retainer Payment Forecasts

```bash
# Forecast upcoming broker retainer payments starting today
xbe do broker-retainer-payment-forecasts create --broker 123

# Forecast starting on a specific date
xbe do broker-retainer-payment-forecasts create --broker 123 --date 2025-02-01

# Output as JSON
xbe do broker-retainer-payment-forecasts create --broker 123 --json
```

### Customer Application Approvals

```bash
# Approve a customer application
xbe do customer-application-approvals create --customer-application 123 --credit-limit 1000000

# Output as JSON
xbe do customer-application-approvals create --customer-application 123 --credit-limit 1000000 --json
```

### Trucker Application Approvals

```bash
# Approve a trucker application
xbe do trucker-application-approvals create --trucker-application 123

# Also add the application user as a trucker manager
xbe do trucker-application-approvals create --trucker-application 123 --add-application-user-as-trucker-manager

# Output as JSON
xbe do trucker-application-approvals create --trucker-application 123 --json
```

### Customer Incident Default Assignees

```bash
# List default assignees for a customer
xbe view customer-incident-default-assignees list --customer 123

# Create a default assignee
xbe do customer-incident-default-assignees create --customer 123 --default-assignee 456 --kind safety
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

### Key Result Changes

```bash
# List key result changes
xbe view key-result-changes list

# Filter by key result
xbe view key-result-changes list --key-result 123

# Filter by objective
xbe view key-result-changes list --objective 456

# Show key result change details
xbe view key-result-changes show 789
```

### Meetings

```bash
# List meetings
xbe view meetings list

# Filter by organization (Type|ID)
xbe view meetings list --organization "Broker|123"

# Filter by organizer
xbe view meetings list --organizer 456

# Show meeting details
xbe view meetings show 789

# Create a meeting
xbe do meetings create --organization-type brokers --organization-id 123 \
  --subject "Weekly Safety Meeting" \
  --start-at 2025-01-15T14:00:00Z --end-at 2025-01-15T15:00:00Z

# Update a meeting summary
xbe do meetings update 789 --summary "Reviewed safety topics and action items"

# Delete a meeting
xbe do meetings delete 789 --confirm
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

### Features, Native App Releases, Release Notes, Press Releases, Glossary Terms

```bash
# List product features
xbe view features list
xbe view features list --pdca-stage plan
xbe view features show 123

# List native app releases
xbe view native-app-releases list
xbe view native-app-releases list --release-channel apple-app-store
xbe view native-app-releases show 321

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

### Tender Job Schedule Shift Material Transaction Checksums

```bash
# Generate checksum diagnostics for a job number and time window
xbe do tender-job-schedule-shifts-material-transactions-checksums create \
  --raw-job-number 3882 \
  --transaction-at-min 2025-01-01T00:00:00Z \
  --transaction-at-max 2025-01-02T00:00:00Z

# Include material sites and a job production plan context
xbe do tender-job-schedule-shifts-material-transactions-checksums create \
  --raw-job-number 3882 \
  --transaction-at-min 2025-01-01T00:00:00Z \
  --transaction-at-max 2025-01-02T00:00:00Z \
  --material-site-ids 101,102 \
  --job-production-plan-id 555

# View a checksum record
xbe view tender-job-schedule-shifts-material-transactions-checksums show 123
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

### Developer Memberships

Developer memberships define the relationship between users and developer organizations.

```bash
# List developer memberships
xbe view developer-memberships list --organization "Developer|123"

# Show developer membership details
xbe view developer-memberships show 456

# Create a developer membership
xbe do developer-memberships create --user 123 --developer 456 --kind manager

# Update a developer membership
xbe do developer-memberships update 789 --title "Project Manager" --can-see-rates-as-manager true

# Delete a developer membership (requires --confirm)
xbe do developer-memberships delete 789 --confirm
```

### Business Unit Memberships

Business unit memberships associate broker memberships with specific business units.

```bash
# List business unit memberships
xbe view business-unit-memberships list --business-unit 123

# Show business unit membership details
xbe view business-unit-memberships show 456

# Create a business unit membership
xbe do business-unit-memberships create --business-unit 123 --membership 456 --kind technician

# Update a business unit membership
xbe do business-unit-memberships update 789 --kind general

# Delete a business unit membership (requires --confirm)
xbe do business-unit-memberships delete 789 --confirm
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

### Tractor Odometer Readings

Tractor odometer readings capture mileage readings for tractors.

```bash
# List tractor odometer readings
xbe view tractor-odometer-readings list

# Filter by tractor
xbe view tractor-odometer-readings list --tractor 123

# Show reading details
xbe view tractor-odometer-readings show 456

# Create a reading
xbe do tractor-odometer-readings create \
  --tractor 123 \
  --unit-of-measure 456 \
  --value 120345.6 \
  --state-code IL \
  --reading-on 2025-01-15 \
  --reading-time 08:30

# Update a reading
xbe do tractor-odometer-readings update 456 --value 120400 --state-code CA

# Delete a reading (requires --confirm)
xbe do tractor-odometer-readings delete 456 --confirm
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

### Rate Adjustments

Rate adjustments connect rates to cost indexes and define how pricing changes as index values move.

```bash
# List rate adjustments
xbe view rate-adjustments list

# Filter by rate or cost index
xbe view rate-adjustments list --rate 123 --cost-index 456

# Show rate adjustment details
xbe view rate-adjustments show 789

# Create a rate adjustment
xbe do rate-adjustments create --rate 123 --cost-index 456 \
  --zero-intercept-value 100 --zero-intercept-ratio 0.25 \
  --adjustment-min 1.00 --adjustment-max 5.00 \
  --prevent-rating-when-index-value-missing

# Update a rate adjustment
xbe do rate-adjustments update 789 --adjustment-max 6.00

# Delete a rate adjustment (requires --confirm)
xbe do rate-adjustments delete 789 --confirm
```

### Rate Agreements Copiers

Rate agreements copiers copy a template rate agreement to multiple customers or truckers.

```bash
# List rate agreements copiers
xbe view rate-agreements-copiers list

# Filter by broker or template
xbe view rate-agreements-copiers list --broker 123
xbe view rate-agreements-copiers list --rate-agreement-template 456

# Show copier details
xbe view rate-agreements-copiers show 789

# Copy a template to customers
xbe do rate-agreements-copiers create \
  --rate-agreement-template 456 \
  --target-customers 111,222 \
  --note "Annual renewal"

# Copy a template to truckers
xbe do rate-agreements-copiers create \
  --rate-agreement-template 456 \
  --target-truckers 333,444
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
