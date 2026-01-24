# Resource Implementation Decisions

Source of truth for what the CLI implements and what remains on the server.
Commands are derived from Cobra registrations (view/do/summarize).

## Work Queue (manual)

Active work items live here. Keep this short and current. Remove items once implemented
so the lists below remain the authoritative snapshot. Update the “Last updated” line
whenever you change this section.

Last updated: 2026-01-23

| Resource / Command | Status | Notes |
|--------------------|--------|-------|
| _TBD_ | planned | — |

## Resource Implementation Spec

Use this checklist for every resource. Treat it as the definition of “done.”

1. Server review
   - Read the resource in `/Users/seandevine/Code/server/app/resources/v1/*`.
     - Capture all attributes, filters, relationships, and special behaviors.
   - Read the policy to determine allowed actions (list/show/create/update/delete).
   - Read the model for validations, required fields, and constraints that affect CLI usage.

2. CLI requirements (by action)
   - View: `list`
     - Show the core columns that define the item (keep output concise).
     - Support every filter exposed by the resource.
     - Support paging/sorting conventions in the CLI.
   - View: `show`
     - Include all fields and relationships relevant to the resource (full detail).
   - Do: `create` / `update` / `delete`
     - Support every writable attribute exposed by the resource.
     - Respect policy constraints (e.g., read-only by policy).

3. Tests
   - Add automated tests covering all supported filters and writable attributes.
   - Follow existing test conventions in `tests/`.
   - Exercise success and relevant failure modes where possible.

4. Documentation
   - Update command help text (short + long + examples) for new commands.
   - Keep it concise and accurate to the server behaviors.
   - Update `README.md` with the new resource commands.

## Status Summary

- Server resources (routes): 665
- CLI command resources: 102
- Server resources covered by commands: 102
- Remaining (after skips + pending + not yet reviewed): 556

## CLI Alias Notes

When a CLI command does not match a server resource name, use these mappings.

- `device-location-event-summary` -> `device-location-event-summaries`
- `driver-day-summary` -> `driver-day-summaries`
- `job-production-plan-summary` -> `job-production-plan-summaries`
- `lane-summary` -> `cycle-summaries`
- `material-transaction-summary` -> `material-transaction-summaries`
- `ptp-driver-summary` -> `project-transport-plan-driver-summaries`
- `ptp-event-summary` -> `project-transport-plan-event-summaries`
- `ptp-event-time-summary` -> `project-transport-plan-event-time-summaries`
- `ptp-summary` -> `project-transport-plan-summaries`
- `ptp-trailer-summary` -> `project-transport-plan-trailer-summaries`
- `public-praise-summary` -> `public-praise-summaries`
- `shift-summary` -> `shift-summaries`
- `transport-order-efficiency-summary` -> `transport-order-efficiency-summaries`
- `transport-summary` -> `transport-summaries`

## Implemented (CLI commands exist for these resources)

```
action-items
broker-commitments
broker-tenders
broker-memberships
broker-settings
brokers
business-units
certification-requirements
certification-types
certifications
comments
cost-codes
cost-index-entries
cost-indexes
craft-classes
crafts
culture-values
custom-work-order-statuses
customer-application-approvals
customer-settings
customer-tenders
customers
developer-memberships
developer-reference-types
developer-references
developer-trucker-certification-classifications
developers
device-location-event-summary
devices
driver-day-summary
driver-day-shortfall-allocations
equipment
equipment-classifications
equipment-rentals
external-identification-types
external-identifications
features
geofences
glossary-terms
hos-violations
incident-participants
incident-tags
incidents
invoice-addresses
invoices
job-production-plan-cancellation-reason-types
job-production-plan-completions
job-production-plan-job-site-changes
job-production-plan-segments
job-production-plan-summary
job-production-plan-supply-demand-balances
job-production-plans
job-sites
labor-classifications
laborers
lane-summary
languages
lineups
material-mix-designs
material-sites
material-site-readings
material-suppliers
material-transaction-summary
material-transaction-cost-code-allocations
material-transactions
material-type-conversions
material-types
memberships
newsletters
parking-sites
posts
press-releases
profit-improvement-categories
project-abandonments
project-categories
project-cost-classifications
project-cost-codes
project-divisions
project-offices
project-phases
project-phase-revenue-items
project-resource-classifications
project-revenue-classifications
project-status-changes
project-transport-event-types
project-transport-plan-event-times
project-transport-plan-segment-tractors
projects
ptp-driver-summary
ptp-event-summary
ptp-event-time-summary
ptp-summary
ptp-trailer-summary
public-praise-summary
public-praises
quality-control-classifications
rates
reaction-classifications
release-notes
service-types
service-type-unit-of-measure-cohorts
shift-feedback-reasons
shift-feedbacks
shift-scope-tenders
shift-summary
stakeholder-classifications
tag-categories
tags
tender-returns
time-card-invoices
time-sheet-line-items
time-sheet-line-item-classifications
tractor-credentials
tractor-fuel-consumption-readings
tractor-trailer-credential-classifications
tractors
trailer-classifications
trailer-credentials
trailers
transport-order-efficiency-summary
transport-orders
transport-summary
trips
truck-scopes
trucker-insurances
truckers
unit-of-measures
user-credential-classifications
user-credentials
users
work-order-assignments
work-orders
```

## Pending Decisions

These need a decision before implementation work proceeds.

- `maintenance-requirements`
- `shift-scopes`

## Remaining (by priority)

### Highest (Core operations & scheduling) (201)

```
crew-assignment-confirmations
crew-rates
crew-requirement-credential-classifications
crew-requirements
driver-assignment-acknowledgements
driver-assignment-refusals
driver-assignment-rules
driver-day-adjustment-plans
driver-day-adjustments
driver-day-constraints
driver-day-shortfall-calculations
driver-day-trips-adjustments
driver-managers
driver-movement-observations
driver-movement-segment-sets
driver-movement-segments
equipment-location-estimates
equipment-location-events
equipment-movement-requirement-locations
equipment-movement-requirements
equipment-movement-stop-completions
equipment-movement-stop-requirements
equipment-movement-stops
equipment-movement-trip-customer-cost-allocations
equipment-movement-trip-dispatch-fulfillment-clerks
equipment-movement-trip-dispatches
equipment-movement-trip-job-production-plans
equipment-movement-trips
equipment-requirements
equipment-suppliers
equipment-utilization-readings
geofence-restriction-violations
geofence-restrictions
hos-annotations
hos-availability-snapshots
hos-day-regulation-sets
hos-days
hos-events
hos-ruleset-assignments
inventory-capacities
inventory-changes
inventory-estimates
job-production-plan-abandonments
job-production-plan-alarm-subscribers
job-production-plan-alarms
job-production-plan-approvals
job-production-plan-broadcast-messages
job-production-plan-cancellations
job-production-plan-change-sets
job-production-plan-completions
job-production-plan-cost-codes
job-production-plan-display-unit-of-measures
job-production-plan-driver-movements
job-production-plan-duplication-works
job-production-plan-duplications
job-production-plan-inspectors
job-production-plan-job-site-location-estimates
job-production-plan-locations
job-production-plan-material-site-changes
job-production-plan-material-sites
job-production-plan-material-type-quality-control-requirements
job-production-plan-material-types
job-production-plan-project-phase-revenue-items
job-production-plan-recap-generations
job-production-plan-recaps
job-production-plan-rejections
job-production-plan-safety-risk-communication-suggestions
job-production-plan-safety-risks
job-production-plan-safety-risks-suggestions
job-production-plan-schedule-change-works
job-production-plan-schedule-changes
job-production-plan-scrappages
job-production-plan-segment-sets
job-production-plan-segments
job-production-plan-service-type-unit-of-measure-cohorts
job-production-plan-service-type-unit-of-measures
job-production-plan-status-changes
job-production-plan-submissions
job-production-plan-subscriptions
job-production-plan-supply-demand-balance-calculators
job-production-plan-time-card-approvers
job-production-plan-trailer-classifications
job-production-plan-trucking-incident-detectors
job-production-plan-unabandonments
job-production-plan-unapprovals
job-production-plan-uncancellations
job-production-plan-uncompletions
job-production-plan-unscrappages
job-schedule-shift-is-managed-toggles
job-schedule-shift-splits
job-schedule-shift-start-at-changes
job-schedule-shift-start-site-changes
job-schedule-shifts
job-site-times
labor-requirements
lineup-dispatch-fulfillment-clerks
lineup-dispatch-shifts
lineup-dispatch-statuses
lineup-dispatches
lineup-job-production-plans
lineup-job-schedule-shift-trucker-assignment-recommendations
lineup-job-schedule-shifts
lineup-scenario-generators
lineup-scenario-lineup-job-schedule-shifts
lineup-scenario-lineups
lineup-scenario-solutions
lineup-scenario-trailer-lineup-job-schedule-shifts
lineup-scenario-trailers
lineup-scenario-truckers
lineup-scenarios
lineup-summary-requests
maintenance-requirement-maintenance-requirement-parts
maintenance-requirement-parts
maintenance-requirement-rule-evaluation-clerks
maintenance-requirement-rule-maintenance-requirement-sets
maintenance-requirement-rules
maintenance-requirement-set-maintenance-requirements
maintenance-requirement-sets
material-mix-design-matches
material-purchase-order-release-redemptions
material-purchase-order-releases
material-purchase-orders
material-site-inventory-locations
material-site-measures
material-site-mergers
material-site-mixing-lots
material-site-reading-material-types
material-site-readings
material-site-subscriptions
material-site-unavailabilities
material-supplier-memberships
material-transaction-acceptances
material-transaction-denials
material-transaction-diversions
material-transaction-field-scopes
material-transaction-inspection-rejections
material-transaction-inspections
material-transaction-invalidations
material-transaction-preloads
material-transaction-rejections
material-transaction-shift-assignments
material-transaction-status-changes
material-transaction-submissions
material-transaction-ticket-generators
material-type-conversions
material-type-material-site-inventory-locations
material-type-unavailabilities
material-unit-of-measure-quantities
resource-unavailabilities
service-events
service-sites
service-type-unit-of-measure-cohorts
service-type-unit-of-measure-quantities
service-type-unit-of-measures
shift-counters
shift-scope-matches
shift-scope-tenders
shift-set-time-card-constraints
shift-time-card-requisitions
site-events
time-card-approval-audits
time-card-approvals
time-card-cost-code-allocations
time-card-invoices
time-card-payroll-certifications
time-card-pre-approvals
time-card-rejections
time-card-scrappages
time-card-status-changes
time-card-submissions
time-card-time-changes
time-card-unapprovals
time-card-unscrappages
time-card-unsubmissions
time-cards
time-sheet-approvals
time-sheet-cost-code-allocations
time-sheet-line-item-equipment-requirements
time-sheet-line-items
time-sheet-no-shows
time-sheet-rejections
time-sheet-status-changes
time-sheet-submissions
time-sheet-unapprovals
time-sheet-unsubmissions
time-sheets
tractor-fuel-consumption-readings
tractor-odometer-readings
transport-order-materials
transport-order-project-transport-plan-strategy-set-predictions
transport-order-stop-materials
transport-order-stops
transport-references
transport-routes
work-order-assignments
work-order-service-codes
```

### High (Project & commercial workflows) (110)

```
bidders
broker-commitments
commitment-items
commitment-material-sites
commitment-simulation-periods
commitment-simulation-sets
commitment-simulations
commitments
customer-commitments
invoice-addresses
invoice-approvals
invoice-generations
invoice-pdf-emails
invoice-rejections
invoice-revisionables
invoice-revisionizing-invoice-revisions
invoice-revisionizing-works
invoice-revisionizings
invoice-revisions
invoice-sends
invoice-status-changes
proffer-likes
proffers
profit-improvement-subscriptions
profit-improvements
project-abandonments
project-approvals
project-bid-location-material-types
project-bid-locations
project-cancellations
project-completions
project-customers
project-duplications
project-estimate-file-imports
project-estimate-sets
project-import-file-verifications
project-labor-classifications
project-margin-matrices
project-material-type-quality-control-requirements
project-material-types
project-phase-cost-item-actuals
project-phase-cost-item-price-estimates
project-phase-cost-item-quantity-estimates
project-phase-cost-items
project-phase-dates-estimates
project-phase-revenue-item-actuals
project-phase-revenue-item-quantity-estimates
project-phase-revenue-items
project-project-cost-classifications
project-rejections
project-revenue-item-price-estimates
project-revenue-item-quantity-estimates
project-revenue-items
project-status-changes
project-submissions
project-subscriptions
project-trailer-classifications
project-transport-location-event-types
project-transport-locations
project-transport-organizations
project-transport-plan-assignment-rules
project-transport-plan-driver-assignment-recommendations
project-transport-plan-driver-confirmations
project-transport-plan-drivers
project-transport-plan-event-location-prediction-autopsies
project-transport-plan-event-location-predictions
project-transport-plan-event-times
project-transport-plan-events
project-transport-plan-planned-event-time-schedules
project-transport-plan-segment-drivers
project-transport-plan-segment-sets
project-transport-plan-segment-tractors
project-transport-plan-segment-trailers
project-transport-plan-segments
project-transport-plan-stop-insertions
project-transport-plan-stop-order-stops
project-transport-plan-stops
project-transport-plan-strategies
project-transport-plan-strategy-sets
project-transport-plan-strategy-steps
project-transport-plan-tractors
project-transport-plan-trailer-assignment-recommendations
project-transport-plan-trailers
project-transport-plans
project-truckers
project-unabandonments
rate-adjustments
rate-agreement-copier-works
rate-agreement-copiers
rate-agreements
rate-agreements-copiers
retainer-deductions
retainer-earning-statuses
retainer-payment-deductions
retainer-payments
retainer-periods
retainers
tender-acceptances
tender-cancellations
tender-job-schedule-shift-cancellations
tender-job-schedule-shift-drivers
tender-job-schedule-shift-time-card-reviews
tender-job-schedule-shifts
tender-job-schedule-shifts-material-transactions-checksums
tender-offers
tender-raters
tender-re-rates
tender-rejections
tender-status-changes
```

### Medium (Org/admin & reference) (161)

```
action-item-key-results
action-item-line-items
action-item-team-members
action-item-tracker-update-requests
action-item-trackers
administrative-incidents
ai-work-order-generations
api-tokens
application-settings
base-summary-templates
broker-certification-types
broker-customers
broker-equipment-classifications
broker-invoices
broker-project-transport-event-types
broker-retainer-payment-forecasts
broker-retainers
broker-tenders
broker-trucker-ratings
broker-vendors
built-time-cards
business-unit-customers
business-unit-equipments
business-unit-laborers
business-unit-memberships
change-requests
contractors
customer-application-approvals
customer-applications
customer-certification-types
customer-incident-default-assignees
customer-memberships
customer-retainers
customer-tenders
customer-truckers
customer-vendors
developer-certified-weighers
developer-memberships
developer-trucker-certification-multipliers
developer-trucker-certifications
device-diagnostics
device-location-events
dispatch-user-matchers
down-minutes-estimates
efficiency-incidents
email-address-statuses
expected-time-of-arrivals
file-attachment-signed-urls
file-attachments
file-imports
incident-headline-suggestions
incident-request-approvals
incident-request-cancellations
incident-request-rejections
incident-requests
incident-subscriptions
incident-tag-incidents
incident-unit-of-measure-quantities
invoices
jobs
key-result-changes
key-result-scrappages
key-result-status-changes
key-result-unscrappages
key-results
liability-incidents
login-code-redemptions
login-code-requests
lowest-losing-bid-prediction-subject-details
mechanic-user-associations
meeting-attendees
meetings
missing-rates
model-filter-infos
objective-changes
objective-stakeholder-classification-quotes
objective-stakeholder-classifications
objectives
open-door-issues
open-door-team-memberships
organization-formatters
organization-invoices-batch-files
organization-invoices-batch-invoice-batchings
organization-invoices-batch-invoice-failures
organization-invoices-batch-invoice-status-changes
organization-invoices-batch-invoice-unbatchings
organization-invoices-batch-invoices
organization-invoices-batch-pdf-files
organization-invoices-batch-pdf-generations
organization-invoices-batch-pdf-templates
organization-invoices-batch-processes
organization-invoices-batch-status-changes
organization-invoices-batches
pave-frame-actual-hours
place-predictions
places
platform-statuses
prediction-agents
prediction-knowledge-bases
prediction-subject-bids
prediction-subject-gap-portions
prediction-subject-gaps
prediction-subject-memberships
prediction-subject-recap-generations
prediction-subject-recaps
prediction-subjects
predictions
process-non-processed-time-card-time-changes
production-incident-detectors
production-incidents
production-measurements
projects-file-imports
prompt-prescriptions
public-praise-culture-values
public-praise-reactions
raw-material-transaction-import-results
raw-material-transaction-sales-customers
raw-material-transactions
raw-records
raw-transport-drivers
raw-transport-orders
raw-transport-projects
raw-transport-tractors
raw-transport-trailers
resource-classification-project-cost-classifications
rmt-adjustments
root-causes
safety-incidents
saml-code-redemptions
search-catalog-entries
sourcing-searches
taggings
tenders
ticket-report-dispatches
ticket-report-imports
trading-partners
trucker-application-approvals
trucker-applications
trucker-brokerages
trucker-invoice-payments
trucker-invoices
trucker-memberships
trucker-referral-codes
trucker-referrals
trucker-settings
trucker-shift-sets
user-auth-token-resets
user-creator-feed-creators
user-creator-feeds
user-device-location-tracking-requests
user-languages
user-location-estimates
user-location-events
user-location-requests
user-searches
user-sourced-material-transaction-image-attribute-extractions
vehicle-location-events
version-events
```

### Low (Analytics, exports, summaries) (24)

```
cost-code-trucking-cost-summaries
cycle-time-comparisons
haskell-lemon-inbound-material-transaction-exports
haskell-lemon-outbound-material-transaction-exports
integration-exports
integration-invoices-batch-exports
invoice-exports
job-production-plan-inspectable-summaries
job-production-plan-material-transaction-summaries
lehman-roberts-apex-viewpoint-ticket-exports
marketing-metrics
material-site-reading-summaries
material-transaction-rate-summaries
material-transactions-exports
organization-invoices-batch-file-exports
organization-project-actuals-exports
ozinga-tk-batch-file-exports
pave-frame-actual-statistics
project-actuals-exports
project-phase-revenue-item-actual-exports
raw-transport-exports
superior-bowen-apex-viewpoint-ticket-exports
ticket-reports
time-sheets-exports
```

### Lowest (Notifications, content, integrations & vendor-specific) (61)

```
answer-feedbacks
answer-related-contents
answers
broker-tender-cancelled-seller-notifications
broker-tender-offered-seller-notification-subscriptions
broker-tender-offered-seller-notifications
broker-tender-returned-buyer-notifications
comment-reactions
communications
customer-tender-offered-buyer-notification-subscriptions
customer-tender-offered-buyer-notifications
deere-equipments
deere-integrations
digital-fleet-ticket-events
digital-fleet-trucks
exporter-configurations
follows
gauge-vehicles
geotab-vehicles
go-motive-integrations
gps-insight-vehicles
importer-configurations
integration-configs
keep-truckin-users
keep-truckin-vehicles
native-app-releases
notification-delivery-decisions
notification-subscriptions
notifications
one-step-gps-vehicles
open-ai-realtime-sessions
open-ai-vector-stores
post-actions
post-children
post-router-jobs
post-routers
post-views
prediction-knowledge-base-answers
prediction-knowledge-base-questions
prompters
questions
samsara-integrations
samsara-vehicles
shift-acknowledgement-reminder-notification-subscriptions
site-wait-time-notification-triggers
superior-bowen-crew-ledgers
t3-equipmentshare-vehicles
teletrac-navman-vehicles
temeda-vehicles
tender-job-schedule-shift-cancelled-trucker-contact-notifications
tender-job-schedule-shift-fill-out-time-card-seller-notifications
tender-job-schedule-shift-starting-seller-notifications
tenna-vehicles
text-messages
textractions
ui-tour-steps
ui-tours
user-post-feed-posts
user-post-feeds
user-ui-tours
verizon-reveal-vehicles
```

## Not Yet Reviewed

- `ticket-report-types`

## Skipped (intentional)

These are intentionally excluded from remaining until the decision changes.

- `curran-cost-codes`
- `job-types`
- `superior-bowen-cost-codes`
- `superior-bowen-martin-marietta-ticket-report-types`

## Notes

- Abstract resources marked `abstract` in the server are not real endpoints and are not tracked here.
- Some commands may call supporting endpoints; those do not create a new command resource entry.
