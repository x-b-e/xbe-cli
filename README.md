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
│   ├── bidders             Manage bidders
│   │   ├── create           Create a bidder
│   │   ├── update           Update a bidder
│   │   └── delete           Delete a bidder
│   ├── contractors         Manage contractors
│   │   ├── create           Create a contractor
│   │   ├── update           Update a contractor
│   │   └── delete           Delete a contractor
│   ├── application-settings Manage global application settings
│   │   ├── create           Create an application setting
│   │   ├── update           Update an application setting
│   │   └── delete           Delete an application setting
│   ├── base-summary-templates Manage base summary templates
│   │   ├── create           Create a base summary template
│   │   └── delete           Delete a base summary template
│   ├── email-address-statuses Check email address status
│   │   └── create           Check an email address
│   ├── glossary-terms       Manage glossary terms
│   │   ├── create           Create a glossary term
│   │   ├── update           Update a glossary term
│   │   └── delete           Delete a glossary term
│   ├── file-attachments     Manage file attachments
│   │   ├── create           Create a file attachment
│   │   ├── update           Update a file attachment
│   │   └── delete           Delete a file attachment
│   ├── platform-statuses    Manage platform status updates
│   │   ├── create           Create a platform status
│   │   ├── update           Update a platform status
│   │   └── delete           Delete a platform status
│   ├── device-location-events Record device location events
│   │   └── create           Create a device location event
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
│   ├── customer-applications Manage customer applications
│   │   ├── create           Create a customer application
│   │   ├── update           Update a customer application
│   │   └── delete           Delete a customer application
│   ├── customer-commitments Manage customer commitments
│   │   ├── create           Create a customer commitment
│   │   ├── update           Update a customer commitment
│   │   └── delete           Delete a customer commitment
│   ├── customer-retainers   Manage customer retainers
│   │   ├── create           Create a customer retainer
│   │   ├── update           Update a customer retainer
│   │   └── delete           Delete a customer retainer
│   ├── action-item-key-results Manage action item key result links
│   │   ├── create           Create an action item key result link
│   │   └── delete           Delete an action item key result link
│   ├── action-item-trackers Manage action item trackers
│   │   ├── create           Create an action item tracker
│   │   ├── update           Update an action item tracker
│   │   └── delete           Delete an action item tracker
│   ├── job-production-plan-approvals Approve job production plans
│   │   └── create           Approve a job production plan
│   ├── project-approvals    Approve projects
│   │   └── create           Approve a project
│   ├── project-unabandonments Unabandon projects
│   │   └── create           Unabandon a project
│   ├── project-submissions  Submit projects
│   │   └── create           Submit a project
│   ├── project-bid-locations Manage project bid locations
│   │   ├── create           Create a project bid location
│   │   ├── update           Update a project bid location
│   │   └── delete           Delete a project bid location
│   ├── project-estimate-file-imports Import project estimate files
│   │   └── create           Import a project estimate file
│   ├── project-labor-classifications Manage project labor classifications
│   │   ├── create           Create a project labor classification
│   │   ├── update           Update a project labor classification
│   │   └── delete           Delete a project labor classification
│   ├── project-transport-location-event-types Manage project transport location event types
│   │   ├── create           Create a project transport location event type
│   │   └── delete           Delete a project transport location event type
│   ├── project-transport-plan-event-location-predictions Manage project transport plan event location predictions
│   │   ├── create           Create a location prediction
│   │   ├── update           Update a location prediction
│   │   └── delete           Delete a location prediction
│   ├── project-transport-plan-planned-event-time-schedules Generate planned event time schedules
│   │   └── create           Generate a planned event time schedule
│   ├── project-transport-plan-segments Manage project transport plan segments
│   │   ├── create           Create a project transport plan segment
│   │   ├── update           Update a project transport plan segment
│   │   └── delete           Delete a project transport plan segment
│   ├── project-transport-plan-trailers Manage project transport plan trailers
│   │   ├── create           Create a project transport plan trailer assignment
│   │   ├── update           Update a project transport plan trailer assignment
│   │   └── delete           Delete a project transport plan trailer assignment
│   ├── project-transport-plan-driver-assignment-recommendations Generate driver assignment recommendations
│   │   └── create           Generate driver assignment recommendations
│   ├── project-transport-plan-trailer-assignment-recommendations Generate trailer assignment recommendations
│   │   └── create           Generate trailer assignment recommendations
│   ├── project-phase-cost-item-actuals Manage project phase cost item actuals
│   │   ├── create           Create a project phase cost item actual
│   │   ├── update           Update a project phase cost item actual
│   │   └── delete           Delete a project phase cost item actual
│   ├── project-phase-revenue-item-actuals Manage project phase revenue item actuals
│   │   ├── create           Create a project phase revenue item actual
│   │   ├── update           Update a project phase revenue item actual
│   │   └── delete           Delete a project phase revenue item actual
│   ├── project-revenue-item-price-estimates Manage project revenue item price estimates
│   │   ├── create           Create a project revenue item price estimate
│   │   ├── update           Update a project revenue item price estimate
│   │   └── delete           Delete a project revenue item price estimate
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
│   ├── incident-request-approvals Approve incident requests
│   │   └── create           Approve an incident request
│   ├── incident-request-rejections Reject incident requests
│   │   └── create           Reject an incident request
│   ├── incident-subscriptions Manage incident subscriptions
│   │   ├── create           Create an incident subscription
│   │   ├── update           Update an incident subscription
│   │   └── delete           Delete an incident subscription
│   ├── tender-acceptances   Accept tenders
│   │   └── create           Accept a tender
│   ├── tender-offers        Offer tenders
│   │   └── create           Offer a tender
│   ├── tender-re-rates      Re-rate tenders
│   │   └── create           Re-rate tenders
│   ├── tender-job-schedule-shift-time-card-reviews Manage tender job schedule shift time card reviews
│   │   ├── create           Create a time card review
│   │   └── delete           Delete a time card review
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
│   ├── memberships          Manage user-organization memberships
│   │   ├── create           Create a membership
│   │   ├── update           Update a membership
│   │   └── delete           Delete a membership
│   ├── rate-agreements      Manage rate agreements
│   │   ├── create           Create a rate agreement
│   │   ├── update           Update a rate agreement
│   │   └── delete           Delete a rate agreement
│   ├── proffers             Manage proffers
│   │   ├── create           Create a proffer
│   │   ├── update           Update a proffer
│   │   └── delete           Delete a proffer
│   └── developer-trucker-certification-multipliers Manage developer trucker certification multipliers
│       ├── create           Create a developer trucker certification multiplier
│       ├── update           Update a developer trucker certification multiplier
│       └── delete           Delete a developer trucker certification multiplier
├── view                    Browse and view XBE content
│   ├── application-settings Browse application settings
│   │   ├── list            List application settings
│   │   └── show <id>       Show application setting details
│   ├── file-attachments     Browse file attachments
│   │   ├── list            List file attachments with filtering
│   │   └── show <id>       Show file attachment details
│   ├── newsletters         Browse and view newsletters
│   │   ├── list            List newsletters with filtering
│   │   └── show <id>       Show newsletter details
│   ├── posts               Browse and view posts
│   │   ├── list            List posts with filtering
│   │   └── show <id>       Show post details
│   ├── proffers            Browse and view proffers
│   │   ├── list            List proffers with filtering
│   │   └── show <id>       Show proffer details
│   ├── brokers             Browse broker/branch information
│   │   └── list            List brokers with filtering
│   ├── bidders             Browse bidders
│   │   ├── list            List bidders with filtering
│   │   └── show <id>       Show bidder details
│   ├── contractors         Browse contractors
│   │   ├── list            List contractors with filtering
│   │   └── show <id>       Show contractor details
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
│   ├── customer-applications Browse customer applications
│   │   ├── list            List customer applications with filtering
│   │   └── show <id>       Show customer application details
│   ├── invoice-revisionizing-works Browse invoice revisionizing work
│   │   ├── list            List invoice revisionizing work with filtering
│   │   └── show <id>       Show invoice revisionizing work details
│   ├── service-sites       Browse service sites
│   │   ├── list            List service sites with filtering
│   │   └── show <id>       Show service site details
│   ├── customers           Browse customers
│   │   └── list            List customers with filtering
│   ├── rate-agreements     Browse rate agreements
│   │   ├── list            List rate agreements with filtering
│   │   └── show <id>       Show rate agreement details
│   ├── customer-retainers  Browse customer retainers
│   │   ├── list            List customer retainers with filtering
│   │   └── show <id>       Show customer retainer details
│   ├── retainer-payment-deductions Browse retainer payment deductions
│   │   ├── list            List retainer payment deductions
│   │   └── show <id>       Show retainer payment deduction details
│   ├── customer-commitments Browse customer commitments
│   │   ├── list            List customer commitments with filtering
│   │   └── show <id>       Show customer commitment details
│   ├── action-item-key-results Browse action item key result links
│   │   ├── list            List action item key result links with filtering
│   │   └── show <id>       Show action item key result details
│   ├── action-item-trackers Browse action item trackers
│   │   ├── list            List action item trackers
│   │   └── show <id>       Show action item tracker details
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
│   ├── project-approvals   Browse project approvals
│   │   ├── list            List project approvals
│   │   └── show <id>       Show project approval details
│   ├── project-unabandonments Browse project unabandonments
│   │   ├── list            List project unabandonments
│   │   └── show <id>       Show project unabandonment details
│   ├── project-submissions Browse project submissions
│   │   ├── list            List project submissions
│   │   └── show <id>       Show project submission details
│   ├── project-bid-locations Browse project bid locations
│   │   ├── list            List project bid locations with filtering
│   │   └── show <id>       Show project bid location details
│   ├── project-estimate-file-imports Browse project estimate file imports
│   │   └── list            List project estimate file imports
│   ├── project-labor-classifications Browse project labor classifications
│   │   ├── list            List project labor classifications with filtering
│   │   └── show <id>       Show project labor classification details
│   ├── project-transport-location-event-types Browse project transport location event types
│   │   ├── list            List project transport location event types with filtering
│   │   └── show <id>       Show project transport location event type details
│   ├── project-transport-plan-event-location-predictions Browse project transport plan event location predictions
│   │   ├── list            List project transport plan event location predictions
│   │   └── show <id>       Show project transport plan event location prediction details
│   ├── project-transport-plan-planned-event-time-schedules Browse project transport plan planned event time schedules
│   │   ├── list            List project transport plan planned event time schedules
│   │   └── show <id>       Show project transport plan planned event time schedule details
│   ├── project-transport-plan-segments Browse project transport plan segments
│   │   ├── list            List project transport plan segments with filtering
│   │   └── show <id>       Show project transport plan segment details
│   ├── project-transport-plan-trailers Browse project transport plan trailers
│   │   ├── list            List project transport plan trailers with filtering
│   │   └── show <id>       Show project transport plan trailer details
│   ├── project-transport-plan-strategies Browse project transport plan strategies
│   │   ├── list            List project transport plan strategies with filtering
│   │   └── show <id>       Show project transport plan strategy details
│   ├── project-transport-plan-driver-assignment-recommendations Browse project transport plan driver assignment recommendations
│   │   ├── list            List project transport plan driver assignment recommendations
│   │   └── show <id>       Show project transport plan driver assignment recommendation details
│   ├── project-transport-plan-trailer-assignment-recommendations Browse project transport plan trailer assignment recommendations
│   │   ├── list            List project transport plan trailer assignment recommendations
│   │   └── show <id>       Show project transport plan trailer assignment recommendation details
│   ├── project-phase-cost-item-actuals Browse project phase cost item actuals
│   │   ├── list            List project phase cost item actuals with filtering
│   │   └── show <id>       Show project phase cost item actual details
│   ├── project-phase-revenue-item-actuals Browse project phase revenue item actuals
│   │   ├── list            List project phase revenue item actuals with filtering
│   │   └── show <id>       Show project phase revenue item actual details
│   ├── project-revenue-item-price-estimates Browse project revenue item price estimates
│   │   ├── list            List project revenue item price estimates with filtering
│   │   └── show <id>       Show project revenue item price estimate details
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
│   ├── incident-request-approvals Browse incident request approvals
│   │   ├── list            List incident request approvals
│   │   └── show <id>       Show incident request approval details
│   ├── incident-request-rejections Browse incident request rejections
│   │   ├── list            List incident request rejections
│   │   └── show <id>       Show incident request rejection details
│   ├── incident-subscriptions Browse incident subscriptions
│   │   ├── list            List incident subscriptions with filtering
│   │   └── show <id>       Show incident subscription details
│   ├── tender-acceptances  Browse tender acceptances
│   │   ├── list            List tender acceptances
│   │   └── show <id>       Show tender acceptance details
│   ├── tender-offers       Browse tender offers
│   │   ├── list            List tender offers
│   │   └── show <id>       Show tender offer details
│   ├── tender-re-rates     Browse tender re-rates
│   │   ├── list            List tender re-rates
│   │   └── show <id>       Show tender re-rate details
│   ├── tender-job-schedule-shift-time-card-reviews Browse tender job schedule shift time card reviews
│   │   ├── list            List tender job schedule shift time card reviews
│   │   └── show <id>       Show tender job schedule shift time card review details
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
│   ├── transport-routes    Browse transport routes
│   │   ├── list            List transport routes with filtering
│   │   └── show <id>       Show transport route details
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
│   ├── base-summary-templates Browse base summary templates
│   │   ├── list            List base summary templates with filtering
│   │   └── show <id>       Show base summary template details
│   ├── glossary-terms      Browse glossary terms
│   │   ├── list            List glossary terms with filtering
│   │   └── show <id>       Show glossary term details
│   └── developer-trucker-certification-multipliers Browse developer trucker certification multipliers
│       ├── list            List developer trucker certification multipliers with filtering
│       └── show <id>       Show developer trucker certification multiplier details
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

### Proffers

```bash
# List proffers
xbe view proffers list

# Filter by kind
xbe view proffers list --kind hot_feed_post

# Find similar proffers (uses embeddings)
xbe view proffers list --similar-to-text "export to CSV"

# View full proffer details
xbe view proffers show 123

# Create a proffer
xbe do proffers create --title "Add CSV export"
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

### Business Unit Customers

```bash
# List business unit customer links
xbe view business-unit-customers list

# Filter by business unit
xbe view business-unit-customers list --business-unit 123

# Filter by customer
xbe view business-unit-customers list --customer 456

# Show a business unit customer link
xbe view business-unit-customers show 789

# Create a business unit customer link
xbe do business-unit-customers create --business-unit 123 --customer 456

# Delete a business unit customer link
xbe do business-unit-customers delete 789 --confirm
```

### Business Unit Laborers

```bash
# List business unit laborer links
xbe view business-unit-laborers list

# Filter by business unit
xbe view business-unit-laborers list --business-unit 123

# Filter by laborer
xbe view business-unit-laborers list --laborer 456

# Show a business unit laborer link
xbe view business-unit-laborers show 789

# Create a business unit laborer link
xbe do business-unit-laborers create --business-unit 123 --laborer 456

# Delete a business unit laborer link
xbe do business-unit-laborers delete 789 --confirm
```

### Broker Customers

```bash
# List broker-customer relationships
xbe view broker-customers list

# Filter by broker and customer
xbe view broker-customers list --broker 123
xbe view broker-customers list --partner "Customer|456"

# Show a broker-customer relationship
xbe view broker-customers show 789

# Create a broker-customer relationship
xbe do broker-customers create --broker 123 --customer 456 \
  --external-accounting-broker-customer-id "ACCT-42"

# Update external accounting ID
xbe do broker-customers update 789 --external-accounting-broker-customer-id "ACCT-43"

# Delete a broker-customer relationship
xbe do broker-customers delete 789 --confirm
```

### Broker Vendors

```bash
# List broker-vendor relationships
xbe view broker-vendors list

# Filter by broker and vendor
xbe view broker-vendors list --broker 123
xbe view broker-vendors list --partner "Trucker|456"

# Show a broker-vendor relationship
xbe view broker-vendors show 789

# Create a broker-vendor relationship
xbe do broker-vendors create --broker 123 --vendor "Trucker|456" \
  --external-accounting-broker-vendor-id "ACCT-42"

# Update external accounting ID
xbe do broker-vendors update 789 --external-accounting-broker-vendor-id "ACCT-43"

# Delete a broker-vendor relationship
xbe do broker-vendors delete 789 --confirm
```

### Customer Vendors

```bash
# List customer-vendor relationships
xbe view customer-vendors list

# Filter by customer and vendor
xbe view customer-vendors list --customer 123
xbe view customer-vendors list --partner "Trucker|456"

# Show a customer-vendor relationship
xbe view customer-vendors show 789

# Create a customer-vendor relationship
xbe do customer-vendors create --customer 123 --vendor "Trucker|456" \
  --external-accounting-customer-vendor-id "ACCT-42"

# Update external accounting ID
xbe do customer-vendors update 789 --external-accounting-customer-vendor-id "ACCT-43"

# Delete a customer-vendor relationship
xbe do customer-vendors delete 789 --confirm
```

### Broker Project Transport Event Types

```bash
# List broker project transport event types
xbe view broker-project-transport-event-types list

# Filter by broker or event type
xbe view broker-project-transport-event-types list --broker 123
xbe view broker-project-transport-event-types list --project-transport-event-type 456

# Show a broker project transport event type
xbe view broker-project-transport-event-types show 789

# Create a broker project transport event type
xbe do broker-project-transport-event-types create \
  --broker 123 \
  --project-transport-event-type 456 \
  --code "PICK"

# Update the broker-specific code
xbe do broker-project-transport-event-types update 789 --code "DROP"

# Delete a broker project transport event type
xbe do broker-project-transport-event-types delete 789 --confirm
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

### Base Summary Templates

```bash
# List base summary templates
xbe view base-summary-templates list

# Show a base summary template
xbe view base-summary-templates show 123

# Create a base summary template
xbe do base-summary-templates create \
  --label "Weekly Summary" \
  --broker 123 \
  --group-bys broker \
  --filter broker=123

# Delete a base summary template
xbe do base-summary-templates delete 123 --confirm
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

### Commitment Simulations

```bash
# List commitment simulations
xbe view commitment-simulations list

# Filter by commitment and status
xbe view commitment-simulations list --commitment 123 --status enqueued

# Create a commitment simulation
xbe do commitment-simulations create \
  --commitment-type commitments \
  --commitment-id 123 \
  --start-on 2026-01-23 \
  --end-on 2026-01-23 \
  --iteration-count 100
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

### Developer Trucker Certification Multipliers

```bash
# List multipliers
xbe view developer-trucker-certification-multipliers list

# Filter by developer trucker certification
xbe view developer-trucker-certification-multipliers list --developer-trucker-certification 123

# Filter by trailer
xbe view developer-trucker-certification-multipliers list --trailer 456

# Show multiplier details
xbe view developer-trucker-certification-multipliers show 789

# Create a multiplier
xbe do developer-trucker-certification-multipliers create \
  --developer-trucker-certification 123 \
  --trailer 456 \
  --multiplier 0.85

# Update a multiplier
xbe do developer-trucker-certification-multipliers update 789 --multiplier 0.9

# Delete a multiplier
xbe do developer-trucker-certification-multipliers delete 789 --confirm
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

### Project Transport Plan Driver Confirmations

Project transport plan driver confirmations track when drivers confirm or reject assignments.

```bash
# List confirmations
xbe view project-transport-plan-driver-confirmations list

# Filter by status
xbe view project-transport-plan-driver-confirmations list --status pending

# Filter by project transport plan driver
xbe view project-transport-plan-driver-confirmations list --project-transport-plan-driver 123

# Show confirmation details
xbe view project-transport-plan-driver-confirmations show 456

# Update status
xbe do project-transport-plan-driver-confirmations update 456 --status confirmed

# Append a note
xbe do project-transport-plan-driver-confirmations update 456 --note "Reviewed"

# Update confirmation deadline
xbe do project-transport-plan-driver-confirmations update 456 --confirm-at-max "2025-01-01T12:00:00Z"
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

### Rate Agreements

Rate agreements define negotiated pricing between sellers (brokers or truckers) and buyers (customers or brokers).

```bash
# List rate agreements
xbe view rate-agreements list

# Filter by seller and buyer
xbe view rate-agreements list --seller "Broker|123" --buyer "Customer|456"

# Show a rate agreement
xbe view rate-agreements show 789

# Create a rate agreement
xbe do rate-agreements create --name "Standard" --status active --seller "Broker|123" --buyer "Customer|456"

# Update a rate agreement
xbe do rate-agreements update 789 --status inactive

# Delete a rate agreement (requires --confirm)
xbe do rate-agreements delete 789 --confirm
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

### Transport Routes

Transport routes represent computed paths between origin and destination coordinates.

```bash
# List routes near an origin location
xbe view transport-routes list --near-origin-location "40.7128|-74.0060|10"

# List routes near a destination location
xbe view transport-routes list --near-destination-location "34.0522|-118.2437|25"

# Show a route
xbe view transport-routes show 123

# JSON output
xbe view transport-routes list --json --limit 10
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

### Project Transport Plan Driver Assignment Recommendations

Project transport plan driver assignment recommendations rank candidate drivers for a project transport plan driver.

```bash
# Generate recommendations for a driver assignment
xbe do project-transport-plan-driver-assignment-recommendations create \
  --project-transport-plan-driver 123

# List recommendations (optionally filter by driver)
xbe view project-transport-plan-driver-assignment-recommendations list \
  --project-transport-plan-driver 123

# Show recommendation details
xbe view project-transport-plan-driver-assignment-recommendations show <id>
```

### Project Transport Plan Trailer Assignment Recommendations

Project transport plan trailer assignment recommendations rank candidate trailers for a project transport plan trailer.

```bash
# Generate recommendations for a trailer assignment
xbe do project-transport-plan-trailer-assignment-recommendations create \
  --project-transport-plan-trailer 123

# List recommendations (optionally filter by trailer)
xbe view project-transport-plan-trailer-assignment-recommendations list \
  --project-transport-plan-trailer 123

# Show recommendation details
xbe view project-transport-plan-trailer-assignment-recommendations show <id>
```

### Project Transport Plan Event Location Predictions

Project transport plan event location predictions rank candidate locations for a transport plan event.

```bash
# Create predictions for a project transport plan event
xbe do project-transport-plan-event-location-predictions create \
  --project-transport-plan-event 123 \
  --transport-order 456

# Create predictions with explicit context
xbe do project-transport-plan-event-location-predictions create \
  --transport-order 456 \
  --project-transport-event-type-id-explicit 789 \
  --event-position-explicit 1 \
  --broker-id-explicit 321

# List predictions filtered by event
xbe view project-transport-plan-event-location-predictions list \
  --project-transport-plan-event 123

# Show prediction details
xbe view project-transport-plan-event-location-predictions show <id>

# Delete a prediction (requires --confirm)
xbe do project-transport-plan-event-location-predictions delete <id> --confirm
```

### Project Transport Plan Strategies

Project transport plan strategies define ordered steps and patterns used for transport planning.

```bash
# List strategies
xbe view project-transport-plan-strategies list

# Filter by name
xbe view project-transport-plan-strategies list --name "Default"

# Filter by step pattern
xbe view project-transport-plan-strategies list --step-pattern "pickup-dropoff"

# Show strategy details
xbe view project-transport-plan-strategies show <id>
```

### Project Transport Plan Segments

Project transport plan segments connect origin and destination stops within a plan.

```bash
# Create a segment
xbe do project-transport-plan-segments create \
  --project-transport-plan 123 \
  --origin 456 \
  --destination 789

# Update segment miles
xbe do project-transport-plan-segments update 101 --miles 12.5

# List segments for a plan
xbe view project-transport-plan-segments list --project-transport-plan 123

# Show segment details
xbe view project-transport-plan-segments show <id>

# Delete a segment (requires --confirm)
xbe do project-transport-plan-segments delete <id> --confirm
```

### Project Transport Plan Trailers

Project transport plan trailers assign trailers across a segment range within a plan.

```bash
# Create a trailer assignment
xbe do project-transport-plan-trailers create \
  --project-transport-plan 123 \
  --segment-start 456 \
  --segment-end 789

# Update status or trailer
xbe do project-transport-plan-trailers update 101 --status active --trailer 555

# List trailer assignments for a plan
xbe view project-transport-plan-trailers list --project-transport-plan 123

# Show trailer assignment details
xbe view project-transport-plan-trailers show <id>

# Delete a trailer assignment (requires --confirm)
xbe do project-transport-plan-trailers delete <id> --confirm
```

### Project Transport Plan Planned Event Time Schedules

Project transport plan planned event time schedules compute planned event times and warnings.

```bash
# Generate a schedule for a project transport plan
xbe do project-transport-plan-planned-event-time-schedules create \
  --project-transport-plan 123

# Generate a schedule from explicit event data
xbe do project-transport-plan-planned-event-time-schedules create \
  --transport-order 456 \
  --plan-data '{"events":[{"location_id":1,"event_type_id":2}]}'

# List schedules filtered by project transport plan
xbe view project-transport-plan-planned-event-time-schedules list \
  --project-transport-plan 123

# Show schedule details
xbe view project-transport-plan-planned-event-time-schedules show <id>
```

### Device Location Events

Device location events capture device-reported GPS activity for users. The API currently supports create-only access.

```bash
# Create an event from payload
xbe do device-location-events create --device-identifier "ios:ABC123" \
  --payload '{"uuid":"evt-1","timestamp":"2025-01-01T00:00:00Z","activity":{"type":"walking"},"coords":{"latitude":40.0,"longitude":-74.0}}'

# Create an event with explicit fields
xbe do device-location-events create --device-identifier "ios:ABC123" \
  --event-id "evt-2" --event-at 2025-01-01T00:05:00Z --event-description "moving" \
  --event-latitude 40.1 --event-longitude -74.1

# JSON output
xbe do device-location-events create --device-identifier "ios:ABC123" \
  --payload '{"uuid":"evt-3","timestamp":"2025-01-01T00:10:00Z","activity":{"type":"still"},"coords":{"latitude":40.2,"longitude":-74.2}}' \
  --json
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

### Email Address Statuses

Email address statuses report whether an email address is on the rejection list (create-only).

```bash
# Check an email address
xbe do email-address-statuses create --email-address "user@example.com"

# JSON output
xbe do email-address-statuses create --email-address "user@example.com" --json
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
