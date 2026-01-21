# Resource Implementation Decisions

This file tracks decisions about which server resources to implement in the CLI.

## Implemented Resources

### Reference Resources (View + Do)

| Resource | View | Do | Notes |
|----------|------|-----|-------|
| `labor-classifications` | list | create, update, delete | Worker types with capabilities |
| `equipment-classifications` | list | create, update, delete | Equipment hierarchy with parent relationships |
| `certification-types` | list | create, update, delete | Certifications with can-apply-to and requirements |
| `tag-categories` | list | create, update, delete | Tag organization by entity type |
| `external-identification-types` | list | create, update, delete | External ID formats (license numbers, tax IDs, etc.) |
| `crafts` | list | create, update, delete | Trade classifications, broker-scoped |
| `tags` | list | create, update, delete | Labels for entities, organized by tag-category |
| `custom-work-order-statuses` | list | create, update, delete | Custom status labels per broker |
| `quality-control-classifications` | list | create, update, delete | QC inspection types, broker-scoped |
| `stakeholder-classifications` | list | create, update, delete | Stakeholder roles with leverage factor (admin-only write) |
| `project-cost-classifications` | list | create, update, delete | Hierarchical cost categories, broker-scoped with parent/child relationships |
| `project-revenue-classifications` | list | create, update, delete | Hierarchical revenue categories, broker-scoped with parent/child relationships |
| `project-resource-classifications` | list | create, update, delete | Hierarchical resource categories, broker-scoped with parent/child relationships |
| `shift-feedback-reasons` | list | create, update, delete | Feedback types for shifts (positive/negative), filters: name, kind, slug, default-rating, has-bot |
| `cost-indexes` | list | create, update, delete | Pricing indexes for rate adjustments, optional broker relationship, filters: broker, is-broker, is-expired |
| `time-sheet-line-item-classifications` | list | create, update, delete | Categories for time sheet line items |
| `job-production-plan-cancellation-reason-types` | list | create, update, delete | Reasons for job production plan cancellation, slug is create-only |
| `project-offices` | list | create, update, delete | Broker-scoped, filters: broker, is_active, name, abbreviation, q |
| `craft-classes` | list | create, update, delete | Child of craft, filters: craft, broker, is_valid_for_drivers |
| `project-transport-event-types` | list | create, update, delete | Broker-scoped |
| `cost-index-entries` | list | create, update, delete | Child of cost-index, time-series entries |
| `developer-reference-types` | list | create, update, delete | Developer-scoped, subject_types array |
| `tractor-trailer-credential-classifications` | list | create, update, delete | Polymorphic org-scoped |
| `user-credential-classifications` | list | create, update, delete | Polymorphic org-scoped |
| `developer-trucker-certification-classifications` | list | create, update, delete | Developer-scoped |
| `truck-scopes` | list | create, update, delete | Polymorphic org-scoped, geocoded address |

### Credential Resources (View + Do)

| Resource | View | Do | Notes |
|----------|------|-----|-------|
| `user-credentials` | list | create, update, delete | Credentials for users, filters: user, user-credential-classification, issued-on min/max, expires-on min/max, active-on |
| `tractor-credentials` | list | create, update, delete | Credentials for tractors, filters: tractor, tractor-trailer-credential-classification, issued-on min/max, expires-on min/max, active-on |
| `trailer-credentials` | list | create, update, delete | Credentials for trailers, filters: trailer, tractor-trailer-credential-classification, issued-on min/max, expires-on min/max, active-on |
| `certifications` | list | create, update, delete | Certifications for polymorphic certifiable entities, filters: certification-type, status, expires-within-days, expires-before, broker |
| `certification-requirements` | list | create, update, delete | Requirements defining certifications needed by entities, filters: certification-type, required-by (Type\|ID format) |

### Reference Resources (View Only)

| Resource | View | Do | Notes |
|----------|------|-----|-------|
| `profit-improvement-categories` | list | — | **Read-only by policy** - categories cannot be created/updated/deleted via API |
| `trailer-classifications` | list | — | |
| `service-types` | list | — | |
| `unit-of-measures` | list | — | |
| `material-types` | list | — | |
| `incident-tags` | list | — | |
| `project-categories` | list | — | |
| `project-divisions` | list | — | |
| `cost-codes` | list | — | |
| `culture-values` | list | create, update, delete | |
| `glossary-terms` | list, show | create, update, delete | |
| `languages` | list | — | Static lookup table (ISO language codes) |
| `reaction-classifications` | list | — | Read-only by policy (emoji reactions) |

## Skipped Resources

| Resource | Reason | Decided |
|----------|--------|---------|
| `job-type` | Not used | 2025-01-21 |
| `level` | Abstract resource - not a real endpoint | 2025-01-21 |
| `credential-classification` | Abstract resource - not a real endpoint | 2025-01-21 |
| `resource-classification` | Abstract resource - not a real endpoint | 2025-01-21 |

## Pending Decisions

| Resource | Attributes | Notes |
|----------|------------|-------|
| `shift-scope` | Many attributes | Complex resource with many relationships - handle later |
| `maintenance-requirement` | Many attributes | Complex resource with has_many relationships - handle later |

## Not Yet Reviewed

Simple reference resources that need evaluation:

| Resource | Attributes | Notes |
|----------|------------|-------|
| `ticket-report-type` | name (read-only), truckers_config | Uses STI, complex config - may skip |

## Join Tables (Not Standalone Resources)

These are relationship tables, not standalone reference resources:

| Resource | Relationship | Notes |
|----------|--------------|-------|
| `broker-certification-type` | broker <-> certification_type | Pure join |
| `broker-equipment-classification` | broker <-> equipment_classification | Pure join |
| `customer-certification-type` | customer <-> certification_type | Pure join |
| `project-labor-classification` | project <-> labor_classification | Join with rate attributes |
| `project-material-type` | project <-> material_type | Join table |
| `project-cost-code` | project <-> cost_code | Join table |
| `project-trailer-classification` | project <-> trailer_classification | Join table |
| `project-project-cost-classification` | project <-> project_cost_classification | Join table |
| `resource-classification-project-cost-classification` | resource_classification <-> project_cost_classification | Join table |
| `service-type-unit-of-measure` | service_type <-> unit_of_measure | Join table |
| `tagging` | polymorphic tag assignments | Join table |
| `user-language` | user <-> language | Join table |
| `public-praise-culture-value` | public_praise <-> culture_value | Join table |
| `objective-stakeholder-classification` | objective <-> stakeholder_classification | Join table |

## Vendor-Specific Integrations (Skip)

| Resource | Notes |
|----------|-------|
| `curran-cost-code` | Curran vendor integration |
| `superior-bowen-cost-code` | Superior Bowen vendor integration |
| `superior-bowen-*-ticket-report-type` | Multiple Superior Bowen ticket report types |

## Notes

- **Abstract resources** in the server (marked with `abstract`) are not real API endpoints and should be skipped
- **Read-only by policy** means the server policy explicitly returns `false` for create/update/destroy actions
- Resources marked as broker-scoped require a broker relationship and may need special handling
