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

## Knowledge Explorer

The knowledge explorer lets you query a local graph of resources, commands, fields,
summaries, and neighborhood relationships built by the Cartographer pipeline. It is
built for AI agents and power users who need to orient quickly without prior context.

### How to use it (fast mental model)

- **Resources** are the core entities (jobs, projects, brokers, etc.).
- **Fields** are attributes/relationships on resources.
- **Commands** show how the CLI can read or mutate those resources.
- **Flags** show how commands filter or set fields.
- **Summaries** expose analytics dimensions + metrics.
- **Neighbors** rank “next best” resources for exploration.
- **Metapaths** show similarity via shared features (shared commands/fields/etc.).
- **Filter paths** show multi-hop filters inferred from CLI flags.

### Typical exploration flow (what an agent should do first)

1) **Search** for a term (`xbe knowledge search job`).
2) **Open the resource** (`xbe knowledge resource jobs`).
3) **Inspect relationships** (`xbe knowledge relations --resource jobs`).
4) **Find commands** (`xbe knowledge commands --resource jobs`).
5) **Inspect flags / filter paths** (`xbe knowledge flags ...`, `xbe knowledge filters ...`).
6) **Check summaries** for analytics (`xbe knowledge summaries --details`).
7) **Expand neighbors** for adjacent exploration (`xbe knowledge neighbors jobs`).

### Knowledge command help (with output)

These are the exact `--help` outputs for every knowledge command.

#### `xbe knowledge --help`

```bash
$ xbe knowledge --help
Explore the local knowledge database produced by the Cartographer pipeline.

This toolkit is designed for AI agents and power users who need to quickly map
resources, commands, fields, summaries, and neighborhood relationships without
prior context.

USAGE:
  xbe knowledge [command]

AVAILABLE COMMANDS:
  commands           List or search CLI commands in the knowledge base
  fields             List fields and their resources
  filters            Show inferred filter paths for list commands
  flags              List flags and their field semantics
  metapath           Show similarity via shared features (metapaths)
  neighbors          Rank neighborhood resources for exploration
  relations          List relationships between resources
  resource           Show details about a resource
  resources          List resources in the knowledge base
  search             Search across resources, commands, fields, and summaries
  summaries          List summary resources and their features

FLAGS:
  Use 'xbe --help' for global flags (--json, --limit, --base-url, etc.)

EXAMPLES:
  # Search across resources, commands, fields, and summaries
  xbe knowledge search job

  # Show a resource with relationships, summaries, and commands
  xbe knowledge resource jobs

  # Rank neighbors for exploration
  xbe knowledge neighbors jobs --limit 20

  # Show multi-hop filter paths inferred from commands
  xbe knowledge filters --resource jobs
```

#### `xbe knowledge search --help`

```bash
$ xbe knowledge search --help
Search across resources, commands, fields, and summaries

USAGE:
  xbe knowledge search <query> [flags]

FLAGS:
      --kind string         Comma-separated kinds to search (resources,commands,fields,flags,relationships,summaries,dimensions,metrics)

EXAMPLES:
  # Search everything
  xbe knowledge search job

  # Limit to resources + commands
  xbe knowledge search job --kind resources,commands
```

#### `xbe knowledge resources --help`

```bash
$ xbe knowledge resources --help
List resources in the knowledge base

USAGE:
  xbe knowledge resources [flags]

FLAGS:
      --field string        Only resources that define a field (attribute or relationship)
      --query string        Substring filter for resource names
      --relationship string Only resources with a relationship name
      --target string       Only resources with relationships targeting this resource

EXAMPLES:
  # List all resources
  xbe knowledge resources

  # Filter resources that include a field
  xbe knowledge resources --field status

  # Filter resources that relate to brokers
  xbe knowledge resources --target brokers
```

#### `xbe knowledge resource --help`

```bash
$ xbe knowledge resource --help
Show details about a resource

USAGE:
  xbe knowledge resource <name> [flags]

FLAGS:
      --sections string     Comma-separated sections (fields,relationships,summaries,summary-features,commands)

EXAMPLES:
  # Show all details
  xbe knowledge resource jobs

  # Only show relationships and commands
  xbe knowledge resource jobs --sections relationships,commands
```

#### `xbe knowledge commands --help`

```bash
$ xbe knowledge commands --help
List or search CLI commands in the knowledge base

USAGE:
  xbe knowledge commands [flags]

FLAGS:
      --kind string         Filter by command kind (view, do, summarize)
      --query string        Substring filter for command path or description
      --resource string     Only commands that operate on a resource
      --verb string         Filter by verb (list, show, create, update, delete)

EXAMPLES:
  # Search commands by keyword
  xbe knowledge commands --query project

  # Commands tied to a resource
  xbe knowledge commands --resource jobs
```

#### `xbe knowledge fields --help`

```bash
$ xbe knowledge fields --help
List fields and their resources

USAGE:
  xbe knowledge fields [flags]

FLAGS:
      --kind string         Filter by kind (attribute, relationship)
      --query string        Substring filter for field or resource
      --resource string     Only fields for a resource

EXAMPLES:
  # Fields for a resource
  xbe knowledge fields --resource jobs

  # Search field names
  xbe knowledge fields --query status
```

#### `xbe knowledge flags --help`

```bash
$ xbe knowledge flags --help
List flags and their field semantics

USAGE:
  xbe knowledge flags [flags]

FLAGS:
      --command string      Filter by command path (substring match)
      --mapped string       Filter by mapping status (true/false)
      --query string        Substring filter for flag name or description
      --resource string     Filter by resource

EXAMPLES:
  # Flags for a specific command
  xbe knowledge flags --command "view jobs list"

  # Unmapped flags
  xbe knowledge flags --mapped=false
```

#### `xbe knowledge relations --help`

```bash
$ xbe knowledge relations --help
List relationships between resources

USAGE:
  xbe knowledge relations [flags]

FLAGS:
      --kind string         Filter by edge kind (relationship, summary)
      --resource string     Filter by source resource
      --target string       Filter by target resource

EXAMPLES:
  # Relationships from a resource
  xbe knowledge relations --resource jobs

  # Relationships targeting a resource
  xbe knowledge relations --target brokers
```

#### `xbe knowledge summaries --help`

```bash
$ xbe knowledge summaries --help
List summary resources and their features

USAGE:
  xbe knowledge summaries [flags]

FLAGS:
      --details             Include dimensions and metrics
      --summary string      Filter by summary resource

EXAMPLES:
  # List summary resources
  xbe knowledge summaries

  # Show details for a summary
  xbe knowledge summaries --summary transport-summaries --details
```

#### `xbe knowledge neighbors --help`

```bash
$ xbe knowledge neighbors --help
Rank neighborhood resources for exploration

USAGE:
  xbe knowledge neighbors <resource> [flags]

FLAGS:
      --explain             Include component-level evidence
      --min-score float     Minimum neighbor score

EXAMPLES:
  # Top neighbors
  xbe knowledge neighbors jobs --limit 20

  # Explain why neighbors are connected
  xbe knowledge neighbors jobs --explain
```

#### `xbe knowledge metapath --help`

```bash
$ xbe knowledge metapath --help
Show similarity via shared features (metapaths)

USAGE:
  xbe knowledge metapath <resource> [flags]

FLAGS:
      --kind string         Filter by feature kind (command_field, summary_dimension, summary_metric, filter_target)

EXAMPLES:
  # Shared command-field similarity
  xbe knowledge metapath jobs --kind command_field
```

#### `xbe knowledge filters --help`

```bash
$ xbe knowledge filters --help
Show inferred filter paths for list commands

USAGE:
  xbe knowledge filters [flags]

FLAGS:
      --command string      Filter by command path (substring)
      --flag string         Filter by flag name
      --resource string     Filter by resource

EXAMPLES:
  # Filter paths for a resource
  xbe knowledge filters --resource jobs

  # Filter paths for a command
  xbe knowledge filters --command "view jobs list"

  # Filter paths for a specific flag
  xbe knowledge filters --flag broker
```

### 20 Examples (with output)

Note: Output is abbreviated and will vary with your local knowledge database.

1) Global search across resources, commands, fields, and summaries

```bash
$ xbe knowledge search job
KIND    NAME                                   DETAIL
field   jobs.status                            attribute
resource jobs
command view jobs list                         List jobs
...
```

2) Find resources that define a field

```bash
$ xbe knowledge resources --field status
RESOURCE                 LABEL_FIELDS          SERVER_TYPES
jobs                     name,code             Job
projects                 name                  Project
...
```

3) Show a full resource profile

```bash
$ xbe knowledge resource jobs
Resource: jobs
Label fields: name, code

Fields:
NAME            KIND        LABEL
status          attribute
broker          relationship
...
```

4) Show relationships from a resource

```bash
$ xbe knowledge relations --resource jobs
SOURCE   RELATION   TARGET          KIND
jobs     broker     brokers         relationship
jobs     project    projects        relationship
...
```

5) Show relationships that target a resource

```bash
$ xbe knowledge relations --target brokers
SOURCE           RELATION   TARGET    KIND
jobs             broker     brokers   relationship
projects         broker     brokers   relationship
...
```

6) Commands that operate on a resource

```bash
$ xbe knowledge commands --resource jobs
COMMAND                 KIND   VERB   RESOURCE   DESCRIPTION
view jobs list           view   list   jobs       List jobs
view jobs show           view   show   jobs       Show job details
...
```

7) List attributes only for a resource

```bash
$ xbe knowledge fields --resource jobs --kind attribute
RESOURCE   FIELD       KIND       LABEL
jobs       status      attribute
jobs       created-at  attribute
...
```

8) Unmapped flags for a specific command

```bash
$ xbe knowledge flags --command "view jobs list" --mapped=false
FLAG         COMMAND            RESOURCE   RELATION   FIELD   MATCH   MODIFIER
q            view jobs list
...
```

9) Mapped flags for a resource

```bash
$ xbe knowledge flags --resource jobs --mapped=true
FLAG      COMMAND            RESOURCE   RELATION   FIELD        MATCH     MODIFIER
broker    view jobs list     jobs       broker     broker       exact
status    view jobs list     jobs       jobs       status       exact
...
```

10) List summary resources

```bash
$ xbe knowledge summaries
SUMMARY                     PRIMARIES                 DIMENSIONS   METRICS
transport-summaries         transport-orders          8            12
shift-summaries             job-schedule-shifts       6            9
...
```

11) Show summary dimensions + metrics

```bash
$ xbe knowledge summaries --summary transport-summaries --details
Summary: transport-summaries
  Primaries: transport-orders
  Dimensions:
    broker
    project
  Metrics:
    total_distance
    total_hours
```

12) Neighborhood ranking for exploration

```bash
$ xbe knowledge neighbors jobs --limit 5
NEIGHBOR   SCORE   REL   SUMMARY   FILTERS   SHARED_FIELDS   SHARED_DIMS   SHARED_METRICS   SHARED_FILTERS
projects   9.50    1     0         2         4              0             0                1
brokers    7.00    1     0         1         2              0             0                0
...
```

13) Explain why neighbors are connected

```bash
$ xbe knowledge neighbors jobs --limit 2 --explain
Neighbor: projects (score 9.50)
  relationship (1) - project
  filter_path (2) - project.customer
  shared_command_field (4)
```

14) Metapath similarity via shared command fields

```bash
$ xbe knowledge metapath jobs --kind command_field
TARGET     PATH_KIND       SHARED
projects   command_field   6
brokers    command_field   4
...
```

15) Multi-hop filter paths for a resource

```bash
$ xbe knowledge filters --resource jobs
COMMAND           RESOURCE   FLAG     PATH                TARGET     TARGET_FIELD   HOPS   MATCH      MODIFIER
view jobs list    jobs       broker   broker              brokers                  1      rel
view jobs list    jobs       customer project.customer    customers  customer       2      rel_attr
...
```

16) Filter paths for a specific command/flag

```bash
$ xbe knowledge filters --command "view projects list" --flag broker
COMMAND              RESOURCE   FLAG     PATH     TARGET    TARGET_FIELD   HOPS   MATCH   MODIFIER
view projects list   projects   broker   broker   brokers                 1      rel
```

17) Search summary-related terms

```bash
$ xbe knowledge search "transport summary"
KIND    NAME                           DETAIL
summary transport-summaries
command summarize transport-summary create
...
```

18) Find resources that relate to brokers

```bash
$ xbe knowledge resources --target brokers
RESOURCE           LABEL_FIELDS   SERVER_TYPES
jobs               name,code      Job
projects           name           Project
transport-orders   name           TransportOrder
...
```

19) List summarize commands

```bash
$ xbe knowledge commands --query summarize --kind summarize
COMMAND                                KIND       VERB    RESOURCE
summarize transport-summary create     summarize  create  transport-summaries
summarize shift-summary create         summarize  create  shift-summaries
...
```

20) Field search by name fragment

```bash
$ xbe knowledge fields --query created-at
RESOURCE   FIELD        KIND
jobs       created-at   attribute
projects   created-at   attribute
...
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
│   ├── model-filter-infos   Fetch filter options for resources
│   │   └── create           Fetch filter options for a resource
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
│   ├── pave-frame-actual-hours Manage pave frame actual hours
│   │   ├── create           Create a pave frame actual hour
│   │   ├── update           Update a pave frame actual hour
│   │   └── delete           Delete a pave frame actual hour
│   ├── device-location-events Record device location events
│   │   └── create           Create a device location event
│   ├── user-location-events Manage user location events
│   │   ├── create           Create a user location event
│   │   ├── update           Update a user location event
│   │   └── delete           Delete a user location event
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
│   ├── key-result-scrappages Manage key result scrappages
│   │   └── create           Create a key result scrappage
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
│   ├── projects-file-imports Import projects files
│   │   └── create           Import projects from a file
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
│   ├── predictions         Manage predictions
│   │   ├── create           Create a prediction
│   │   ├── update           Update a prediction
│   │   └── delete           Delete a prediction
│   ├── prediction-subject-bids Manage prediction subject bids
│   │   ├── create           Create a prediction subject bid
│   │   ├── update           Update a prediction subject bid
│   │   └── delete           Delete a prediction subject bid
│   ├── prediction-subject-gaps Manage prediction subject gaps
│   │   ├── create           Create a prediction subject gap
│   │   ├── update           Update a prediction subject gap
│   │   └── delete           Delete a prediction subject gap
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
│   ├── incident-unit-of-measure-quantities Manage incident unit of measure quantities
│   │   ├── create           Create an incident unit of measure quantity
│   │   ├── update           Update an incident unit of measure quantity
│   │   └── delete           Delete an incident unit of measure quantity
│   ├── root-causes          Manage root causes
│   │   ├── create           Create a root cause
│   │   ├── update           Update a root cause
│   │   └── delete           Delete a root cause
│   ├── liability-incidents  Manage liability incidents
│   │   ├── create           Create a liability incident
│   │   ├── update           Update a liability incident
│   │   └── delete           Delete a liability incident
│   ├── production-incidents Manage production incidents
│   │   ├── create           Create a production incident
│   │   ├── update           Update a production incident
│   │   └── delete           Delete a production incident
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
│   ├── material-transaction-inspections Manage material transaction inspections
│   │   ├── create           Create a material transaction inspection
│   │   ├── update           Update a material transaction inspection
│   │   └── delete           Delete a material transaction inspection
│   ├── raw-material-transactions Manage raw material transactions
│   │   └── update           Update a raw material transaction
│   ├── raw-transport-orders Manage raw transport orders
│   │   ├── create           Create a raw transport order
│   │   ├── update           Update a raw transport order
│   │   └── delete           Delete a raw transport order
│   ├── material-type-unavailabilities Manage material type unavailabilities
│   │   ├── create           Create a material type unavailability
│   │   ├── update           Update a material type unavailability
│   │   └── delete           Delete a material type unavailability
│   ├── material-supplier-memberships Manage material supplier memberships
│   │   ├── create           Create a material supplier membership
│   │   ├── update           Update a material supplier membership
│   │   └── delete           Delete a material supplier membership
│   ├── trucker-memberships  Manage trucker memberships
│   │   ├── create           Create a trucker membership
│   │   ├── update           Update a trucker membership
│   │   └── delete           Delete a trucker membership
│   ├── trucker-settings     Manage trucker settings
│   │   ├── create           Create trucker settings
│   │   └── update           Update trucker settings
│   ├── memberships          Manage user-organization memberships
│   │   ├── create           Create a membership
│   │   ├── update           Update a membership
│   │   └── delete           Delete a membership
│   ├── user-languages       Manage user languages
│   │   ├── create           Create a user language
│   │   ├── update           Update a user language
│   │   └── delete           Delete a user language
│   ├── meeting-attendees    Manage meeting attendees
│   │   ├── create           Create a meeting attendee
│   │   ├── update           Update a meeting attendee
│   │   └── delete           Delete a meeting attendee
│   ├── rate-agreements      Manage rate agreements
│   │   ├── create           Create a rate agreement
│   │   ├── update           Update a rate agreement
│   │   └── delete           Delete a rate agreement
│   ├── proffers             Manage proffers
│   │   ├── create           Create a proffer
│   │   ├── update           Update a proffer
│   │   └── delete           Delete a proffer
│   ├── public-praise-reactions Manage public praise reactions
│   │   ├── create           Create a public praise reaction
│   │   └── delete           Delete a public praise reaction
│   ├── developer-trucker-certification-multipliers Manage developer trucker certification multipliers
│   │   ├── create           Create a developer trucker certification multiplier
│   │   ├── update           Update a developer trucker certification multiplier
│   │   └── delete           Delete a developer trucker certification multiplier
│   ├── objective-stakeholder-classification-quotes Manage objective stakeholder classification quotes
│   │   ├── create           Create an objective stakeholder classification quote
│   │   ├── update           Update an objective stakeholder classification quote
│   │   └── delete           Delete an objective stakeholder classification quote
│   ├── organization-invoices-batch-invoice-batchings Batch organization invoices batch invoices
│   │   └── create           Batch an organization invoices batch invoice
│   ├── organization-invoices-batch-invoice-failures Fail organization invoices batch invoices
│   │   └── create           Fail an organization invoices batch invoice
│   ├── organization-invoices-batch-processes Process organization invoices batches
│   │   └── create           Process an organization invoices batch
│   ├── organization-invoices-batch-pdf-files Download organization invoices batch PDF files
│   │   └── download         Download an organization invoices batch PDF file
│   ├── open-door-issues     Manage open door issues
│   │   ├── create           Create an open door issue
│   │   ├── update           Update an open door issue
│   │   └── delete           Delete an open door issue
│   └── taggings             Manage taggings
│       ├── create           Create a tagging
│       └── delete           Delete a tagging
├── view                    Browse and view XBE content
│   ├── action-item-line-items Browse action item line items
│   │   ├── list            List action item line items
│   │   └── show <id>       Show action item line item details
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
│   ├── user-languages       Browse user language preferences
│   │   ├── list            List user languages with filtering
│   │   └── show <id>       Show user language details
│   ├── user-creator-feed-creators Browse user creator feed creators
│   │   ├── list            List user creator feed creators with filtering
│   │   └── show <id>       Show user creator feed creator details
│   ├── material-suppliers  Browse material suppliers
│   │   └── list            List suppliers with filtering
│   ├── material-supplier-memberships Browse material supplier memberships
│   │   ├── list            List material supplier memberships with filtering
│   │   └── show <id>       Show material supplier membership details
│   ├── trucker-memberships Browse trucker memberships
│   │   ├── list            List trucker memberships with filtering
│   │   └── show <id>       Show trucker membership details
│   ├── trucker-settings    Browse trucker settings
│   │   ├── list            List trucker settings
│   │   └── show <id>       Show trucker setting details
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
│   ├── raw-material-transactions Browse raw material transactions
│   │   ├── list            List raw material transactions with filtering
│   │   └── show <id>       Show raw material transaction details
│   ├── raw-transport-orders Browse raw transport orders
│   │   ├── list            List raw transport orders with filtering
│   │   └── show <id>       Show raw transport order details
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
│   ├── key-result-scrappages Browse key result scrappages
│   │   ├── list            List key result scrappages
│   │   └── show <id>       Show key result scrappage details
│   ├── action-item-trackers Browse action item trackers
│   │   ├── list            List action item trackers
│   │   └── show <id>       Show action item tracker details
│   ├── truckers            Browse trucking companies
│   │   └── list            List truckers with filtering
│   ├── trucker-invoice-payments Browse trucker invoice payments
│   │   ├── list            List trucker invoice payments
│   │   └── show <id>       Show trucker invoice payment details
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
│   ├── projects-file-imports Browse projects file imports
│   │   ├── list            List projects file imports
│   │   └── show <id>       Show projects file import details
│   ├── project-labor-classifications Browse project labor classifications
│   │   ├── list            List project labor classifications with filtering
│   │   └── show <id>       Show project labor classification details
│   ├── project-transport-location-event-types Browse project transport location event types
│   │   ├── list            List project transport location event types with filtering
│   │   └── show <id>       Show project transport location event type details
│   ├── project-transport-plan-event-location-predictions Browse project transport plan event location predictions
│   │   ├── list            List project transport plan event location predictions
│   │   └── show <id>       Show project transport plan event location prediction details
│   ├── predictions         Browse predictions
│   │   ├── list            List predictions
│   │   └── show <id>       Show prediction details
│   ├── prediction-subject-bids Browse prediction subject bids
│   │   ├── list            List prediction subject bids with filtering
│   │   └── show <id>       Show prediction subject bid details
│   ├── prediction-subject-gaps Browse prediction subject gaps
│   │   ├── list            List prediction subject gaps with filtering
│   │   └── show <id>       Show prediction subject gap details
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
│   ├── incident-unit-of-measure-quantities Browse incident unit of measure quantities
│   │   ├── list            List incident unit of measure quantities
│   │   └── show <id>       Show incident unit of measure quantity details
│   ├── root-causes          Browse root causes
│   │   ├── list            List root causes
│   │   └── show <id>       Show root cause details
│   ├── liability-incidents  Browse liability incidents
│   │   ├── list            List liability incidents
│   │   └── show <id>       Show liability incident details
│   ├── production-incidents Browse production incidents
│   │   ├── list            List production incidents
│   │   └── show <id>       Show production incident details
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
│   ├── user-location-events Browse user location events
│   │   ├── list            List user location events with filtering
│   │   └── show <id>       Show user location event details
│   ├── transport-routes    Browse transport routes
│   │   ├── list            List transport routes with filtering
│   │   └── show <id>       Show transport route details
│   ├── site-events         Browse site events
│   │   ├── list            List site events with filtering
│   │   └── show <id>       Show site event details
│   ├── memberships         Browse user-organization memberships
│   │   ├── list            List memberships with filtering
│   │   └── show <id>       Show membership details
│   ├── meeting-attendees   Browse meeting attendees
│   │   ├── list            List meeting attendees with filtering
│   │   └── show <id>       Show meeting attendee details
│   ├── search-catalog-entries Browse search catalog entries
│   │   ├── list            List search catalog entries with filtering
│   │   └── show <id>       Show search catalog entry details
│   ├── features            Browse product features
│   │   ├── list            List features with filtering
│   │   └── show <id>       Show feature details
│   ├── release-notes       Browse release notes
│   │   ├── list            List release notes with filtering
│   │   └── show <id>       Show release note details
│   ├── press-releases      Browse press releases
│   │   ├── list            List press releases
│   │   └── show <id>       Show press release details
│   ├── public-praise-reactions Browse public praise reactions
│   │   ├── list            List public praise reactions
│   │   └── show <id>       Show public praise reaction details
│   ├── platform-statuses   Browse platform status updates
│   │   ├── list            List platform statuses
│   │   └── show <id>       Show platform status details
│   ├── pave-frame-actual-hours Browse pave frame actual hours
│   │   ├── list            List pave frame actual hours
│   │   └── show <id>       Show pave frame actual hour details
│   ├── base-summary-templates Browse base summary templates
│   │   ├── list            List base summary templates with filtering
│   │   └── show <id>       Show base summary template details
│   ├── glossary-terms      Browse glossary terms
│   │   ├── list            List glossary terms with filtering
│   │   └── show <id>       Show glossary term details
│   ├── developer-trucker-certification-multipliers Browse developer trucker certification multipliers
│   │   ├── list            List developer trucker certification multipliers with filtering
│   │   └── show <id>       Show developer trucker certification multiplier details
│   ├── objective-stakeholder-classification-quotes Browse objective stakeholder classification quotes
│   │   ├── list            List objective stakeholder classification quotes with filtering
│   │   └── show <id>       Show objective stakeholder classification quote details
│   ├── organization-invoices-batch-invoice-batchings Browse organization invoices batch invoice batchings
│   │   ├── list            List organization invoices batch invoice batchings
│   │   └── show <id>       Show organization invoices batch invoice batching details
│   ├── organization-invoices-batch-invoice-failures Browse organization invoices batch invoice failures
│   │   ├── list            List organization invoices batch invoice failures
│   │   └── show <id>       Show organization invoices batch invoice failure details
│   ├── organization-invoices-batch-processes Browse organization invoices batch processes
│   │   ├── list            List organization invoices batch processes
│   │   └── show <id>       Show organization invoices batch process details
│   ├── organization-invoices-batch-status-changes Browse organization invoices batch status changes
│   │   ├── list            List organization invoices batch status changes
│   │   └── show <id>       Show organization invoices batch status change details
│   ├── organization-invoices-batch-pdf-files Browse organization invoices batch PDF files
│   │   ├── list            List organization invoices batch PDF files
│   │   └── show <id>       Show organization invoices batch PDF file details
│   ├── open-door-issues    Browse open door issues
│   │   ├── list            List open door issues with filtering
│   │   └── show <id>       Show open door issue details
│   └── taggings            Browse taggings
│       ├── list            List taggings with filtering
│       └── show <id>       Show tagging details
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

### Public Praise Reactions

```bash
# List reactions on public praises
xbe view public-praise-reactions list

# Create a reaction
xbe do public-praise-reactions create --public-praise 123 --reaction-classification 45

# Delete a reaction
xbe do public-praise-reactions delete 987 --confirm
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

### Open Door Issues

```bash
# List open door issues
xbe view open-door-issues list

# Filter by organization
xbe view open-door-issues list --organization "Broker|123"

# Show open door issue details
xbe view open-door-issues show 456

# Create an open door issue
xbe do open-door-issues create \
  --description "Safety concern reported by driver" \
  --status editing \
  --organization "Broker|123" \
  --reported-by 789
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

### Trading Partners

```bash
# List trading partner links
xbe view trading-partners list

# Filter by organization and partner
xbe view trading-partners list --organization "Broker|123"
xbe view trading-partners list --partner "Customer|456"

# Show a trading partner link
xbe view trading-partners show 789

# Create a trading partner link
xbe do trading-partners create --organization "Broker|123" --partner "Customer|456" \
  --trading-partner-type BrokerCustomer

# Delete a trading partner link
xbe do trading-partners delete 789 --confirm
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

### Meeting Attendees

```bash
# List meeting attendees
xbe view meeting-attendees list

# Filter by meeting
xbe view meeting-attendees list --meeting 123

# Filter by user
xbe view meeting-attendees list --user 456

# Filter by location kind
xbe view meeting-attendees list --location-kind on_site

# Filter by presence requirement or present status
xbe view meeting-attendees list --is-presence-required true
xbe view meeting-attendees list --is-present false

# Show a meeting attendee
xbe view meeting-attendees show 789

# Create a meeting attendee
xbe do meeting-attendees create --meeting 123 --user 456 --location-kind on_site --is-present true

# Update a meeting attendee
xbe do meeting-attendees update 789 --location-kind remote --is-present false

# Delete a meeting attendee
xbe do meeting-attendees delete 789 --confirm
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

### User Languages

User languages associate users with preferred languages and default settings.

```bash
# List user languages
xbe view user-languages list

# Filter by user
xbe view user-languages list --user 123

# Filter by language
xbe view user-languages list --language 456

# Filter by default status
xbe view user-languages list --is-default true

# Show a user language
xbe view user-languages show 789

# Create a user language
xbe do user-languages create --user 123 --language 456 --is-default true

# Update default status
xbe do user-languages update 789 --is-default false

# Delete a user language
xbe do user-languages delete 789 --confirm
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

### Objective Stakeholder Classification Quotes

```bash
# List quotes
xbe view objective-stakeholder-classification-quotes list

# Filter by classification
xbe view objective-stakeholder-classification-quotes list --objective-stakeholder-classification 123

# Filter by interest degree range
xbe view objective-stakeholder-classification-quotes list --interest-degree-min 2 --interest-degree-max 5

# Show quote details
xbe view objective-stakeholder-classification-quotes show 456

# Create a quote
xbe do objective-stakeholder-classification-quotes create \
  --objective-stakeholder-classification 123 \
  --content "Stakeholder values transparency"

# Create a generated quote
xbe do objective-stakeholder-classification-quotes create \
  --objective-stakeholder-classification 123 \
  --content "Generated summary" \
  --is-generated

# Update a quote
xbe do objective-stakeholder-classification-quotes update 456 --content "Updated content"

# Delete a quote (requires --confirm)
xbe do objective-stakeholder-classification-quotes delete 456 --confirm
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

### Trucker Memberships

Trucker memberships link users to truckers and control role settings and notifications.

```bash
# List trucker memberships
xbe view trucker-memberships list

# Filter by trucker
xbe view trucker-memberships list --trucker 123

# Show membership details
xbe view trucker-memberships show 456

# Create a membership
xbe do trucker-memberships create --user 123 --trucker 456

# Update a membership
xbe do trucker-memberships update 456 --kind manager --title "Operations Manager"

# Update trailer coassignment reset date
xbe do trucker-memberships update 456 --trailer-coassignments-reset-on 2025-01-15

# Delete a membership (requires --confirm)
xbe do trucker-memberships delete 456 --confirm
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

### Predictions

Predictions capture probability distributions for a prediction subject, along with status and scoring metadata.

```bash
# Create a prediction
xbe do predictions create \
  --prediction-subject 123 \
  --status draft \
  --probability-distribution '{"class_name":"TriangularDistribution","minimum":100,"mode":120,"maximum":140}'

# List predictions filtered by prediction subject
xbe view predictions list --prediction-subject 123

# Show prediction details
xbe view predictions show <id>

# Update prediction status
xbe do predictions update <id> --status submitted

# Delete a prediction (requires --confirm)
xbe do predictions delete <id> --confirm
```

### Prediction Subject Bids

Prediction subject bids capture bidder amounts tied to a prediction subject's lowest losing bid detail.

```bash
# Create a prediction subject bid
xbe do prediction-subject-bids create \
  --bidder 123 \
  --lowest-losing-bid-prediction-subject-detail 456 \
  --amount 120000

# List bids filtered by prediction subject
xbe view prediction-subject-bids list --prediction-subject 789

# Show bid details
xbe view prediction-subject-bids show <id>

# Update a bid amount
xbe do prediction-subject-bids update <id> --amount 125000

# Delete a bid (requires --confirm)
xbe do prediction-subject-bids delete <id> --confirm
```

### Prediction Subject Gaps

Prediction subject gaps capture differences between primary and secondary prediction amounts.

```bash
# Create a prediction subject gap
xbe do prediction-subject-gaps create \
  --prediction-subject 123 \
  --gap-type actual_vs_consensus

# List gaps filtered by prediction subject
xbe view prediction-subject-gaps list --prediction-subject 123

# Show gap details
xbe view prediction-subject-gaps show <id>

# Approve a gap
xbe do prediction-subject-gaps update <id> --status approved

# Delete a gap (requires --confirm)
xbe do prediction-subject-gaps delete <id> --confirm
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

### User Location Events

User location events capture user-reported latitude/longitude with a provenance (gps/map).

```bash
# List events
xbe view user-location-events list

# Filter by user and time
xbe view user-location-events list --user 123 --event-at-min 2025-01-01T00:00:00Z

# Show event details
xbe view user-location-events show <id>

# Create a user location event
xbe do user-location-events create \
  --user 123 \
  --provenance gps \
  --event-at 2025-01-01T12:00:00Z \
  --event-latitude 40.7128 \
  --event-longitude -74.0060

# Update a user location event
xbe do user-location-events update <id> --event-latitude 41.0 --event-longitude -87.0

# Delete a user location event
xbe do user-location-events delete <id> --confirm
```

### Root Causes

Root causes capture underlying issues for incidents and can be linked to parent root causes.

```bash
# List root causes
xbe view root-causes list

# Filter by incident
xbe view root-causes list --incident-type production-incidents --incident-id 123

# Create a root cause
xbe do root-causes create \
  --incident-type production-incidents \
  --incident-id 123 \
  --title "Mechanical failure" \
  --description "Hydraulic leak caused downtime" \
  --is-triaged

# Update a root cause
xbe do root-causes update 456 --title "Updated root cause"

# Delete a root cause
xbe do root-causes delete 456 --confirm
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

### Model Filter Infos

Model filter infos return available option values for resource filters (create-only).

```bash
# Fetch filter options for projects
xbe do model-filter-infos create --resource-type projects

# Limit to selected filter keys
xbe do model-filter-infos create --resource-type projects --filter-keys customer,project_manager

# Scope options to a broker
xbe do model-filter-infos create --resource-type projects --scope-filter broker=123
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
