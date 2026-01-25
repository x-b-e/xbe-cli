#!/bin/bash
#
# XBE CLI Integration Tests: Liability Incidents
#
# Tests CRUD operations for the liability-incidents resource.
#
# COVERAGE: Writable attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
PARENT_INCIDENT_ID=""
CREATED_LIABILITY_INCIDENT_ID=""

describe "Resource: liability-incidents"

# ============================================================================
# Prerequisites - Create broker and customer
# ============================================================================

test_name "Create prerequisite broker for liability incident tests"
BROKER_NAME=$(unique_name "LiabilityIncidentBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        run_tests
    fi
fi

test_name "Create prerequisite customer for liability incident tests"
CUSTOMER_NAME=$(unique_name "LiabilityIncidentCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        run_tests
    fi
else
    fail "Failed to create customer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create liability incident (parent)"
xbe_json do liability-incidents create \
    --subject-type customers \
    --subject-id "$CREATED_CUSTOMER_ID" \
    --start-at "2026-01-01T08:00:00Z" \
    --status open \
    --kind damage

if [[ $status -eq 0 ]]; then
    PARENT_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$PARENT_INCIDENT_ID" && "$PARENT_INCIDENT_ID" != "null" ]]; then
        register_cleanup "liability-incidents" "$PARENT_INCIDENT_ID"
        pass
    else
        fail "Created parent liability incident but no ID returned"
        run_tests
    fi
else
    fail "Failed to create parent liability incident"
    run_tests
fi

test_name "Create liability incident with full attributes"
xbe_json do liability-incidents create \
    --subject-type customers \
    --subject-id "$CREATED_CUSTOMER_ID" \
    --start-at "2026-01-01T09:00:00Z" \
    --end-at "2026-01-01T11:00:00Z" \
    --status open \
    --kind damage \
    --description "Test liability incident description" \
    --natures "personal,property" \
    --severity high \
    --headline "Liability incident headline" \
    --net-impact-minutes 60 \
    --net-impact-dollars 1000 \
    --is-down-time \
    --parent-id "$PARENT_INCIDENT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_LIABILITY_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_LIABILITY_INCIDENT_ID" && "$CREATED_LIABILITY_INCIDENT_ID" != "null" ]]; then
        register_cleanup "liability-incidents" "$CREATED_LIABILITY_INCIDENT_ID"
        pass
    else
        fail "Created liability incident but no ID returned"
        run_tests
    fi
else
    fail "Failed to create liability incident"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show liability incident by ID"
xbe_json view liability-incidents show "$CREATED_LIABILITY_INCIDENT_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List liability incidents"
xbe_json view liability-incidents list --limit 5
assert_success

test_name "List liability incidents returns array"
xbe_json view liability-incidents list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list liability incidents"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

FILTER_TESTS=(
  "--status open"
  "--kind damage"
  "--severity high"
  "--did-stop-work false"
  "--natures personal"
  "--subject Customer|$CREATED_CUSTOMER_ID"
  "--subject-type Customer"
  "--parent $PARENT_INCIDENT_ID"
  "--has-parent true"
  "--broker $CREATED_BROKER_ID"
  "--customer $CREATED_CUSTOMER_ID"
  "--developer 1"
  "--trucker 1"
  "--contractor 1"
  "--material-supplier 1"
  "--material-site 1"
  "--job-production-plan 1"
  "--job-production-plan-project 1"
  "--equipment 1"
  "--assignee 1"
  "--created-by 1"
  "--tender-job-schedule-shift 1"
  "--tender-job-schedule-shift-driver 1"
  "--tender-job-schedule-shift-trucker 1"
  "--job-number JOB-123"
  "--has-equipment false"
  "--has-live-action-items false"
  "--incident-tag 1"
  "--incident-tag-slug test-tag"
  "--zero-incident-tags true"
  "--root-causes 1"
  "--action-items 1"
  "--notifiable-to 1"
  "--user-has-stake 1"
  "--responsible-person 1"
  "--start-on 2026-01-01"
  "--start-on-min 2026-01-01"
  "--start-on-max 2026-01-02"
  "--start-at-min 2026-01-01T08:00:00Z"
  "--start-at-max 2026-01-01T12:00:00Z"
  "--end-at-min 2026-01-01T10:00:00Z"
  "--end-at-max 2026-01-01T12:00:00Z"
  "--net-impact-minutes-min 30"
  "--net-impact-minutes-max 120"
  "--net-impact-dollars-min 500"
  "--net-impact-dollars-max 1500"
  "--q test"
)

for filter in "${FILTER_TESTS[@]}"; do
    test_name "List liability incidents with filter: $filter"
    xbe_json view liability-incidents list $filter --limit 10
    assert_success
done

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List liability incidents with --limit"
xbe_json view liability-incidents list --limit 3
assert_success

test_name "List liability incidents with --offset"
xbe_json view liability-incidents list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update liability incident attributes"
xbe_json do liability-incidents update "$CREATED_LIABILITY_INCIDENT_ID" \
    --status closed \
    --kind theft \
    --severity medium \
    --headline "Updated liability headline" \
    --description "Updated liability incident description" \
    --end-at "2026-01-01T12:00:00Z" \
    --net-impact-minutes 90 \
    --net-impact-dollars 1500 \
    --is-down-time=false \
    --natures personal
assert_success

test_name "Update liability incident type"
xbe_json do liability-incidents update "$CREATED_LIABILITY_INCIDENT_ID" \
    --new-type LiabilityIncident
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update liability incident with did-stop-work fails"
xbe_run do liability-incidents update "$CREATED_LIABILITY_INCIDENT_ID" \
    --did-stop-work true
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete liability incident requires --confirm flag"
xbe_run do liability-incidents delete "$CREATED_LIABILITY_INCIDENT_ID"
assert_failure

test_name "Delete liability incident with --confirm"
DELETE_INCIDENT_ID=""
xbe_json do liability-incidents create \
    --subject-type customers \
    --subject-id "$CREATED_CUSTOMER_ID" \
    --start-at "2026-01-02T08:00:00Z" \
    --status open

if [[ $status -eq 0 ]]; then
    DELETE_INCIDENT_ID=$(json_get ".id")
    xbe_run do liability-incidents delete "$DELETE_INCIDENT_ID" --confirm
    assert_success
else
    skip "Could not create incident for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
