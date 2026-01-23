#!/bin/bash
#
# XBE CLI Integration Tests: Service Type Unit of Measure Cohorts
#
# Tests CRUD operations for the service-type-unit-of-measure-cohorts resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_CUSTOMER_ALT_ID=""
CREATED_COHORT_ID=""
STUOM_TRIGGER_ID=""
STUOM_LIST_ID=""
STUOM_ALT_ID=""

describe "Resource: service-type-unit-of-measure-cohorts"

# ============================================================================
# Pre-checks
# ============================================================================

test_name "Check API token for service type unit of measure lookup"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping service type unit of measure cohort tests"
    run_tests
else
    pass
fi

# ============================================================================
# Prerequisites - Create broker and customers for tests
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "STUOMCohortBroker")

xbe_json do brokers create --name "$BROKER_NAME"
if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        run_tests
    fi
else
    fail "Failed to create broker"
    run_tests
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "STUOMCohortCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

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

test_name "Create secondary customer"
CUSTOMER_ALT_NAME=$(unique_name "STUOMCohortCustomerAlt")

xbe_json do customers create \
    --name "$CUSTOMER_ALT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ALT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ALT_ID" && "$CREATED_CUSTOMER_ALT_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ALT_ID"
        pass
    else
        fail "Created customer but no ID returned"
        run_tests
    fi
else
    fail "Failed to create secondary customer"
    run_tests
fi

# ============================================================================
# Fetch service type unit of measure IDs
# ============================================================================

test_name "Fetch service type unit of measure IDs"
base_url="${XBE_BASE_URL%/}"
stuom_json=$(curl -s \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    "$base_url/v1/service-type-unit-of-measures?page[limit]=5" || true)

STUOM_TRIGGER_ID=$(echo "$stuom_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
STUOM_LIST_ID=$(echo "$stuom_json" | jq -r '.data[1].id // empty' 2>/dev/null || true)
STUOM_ALT_ID=$(echo "$stuom_json" | jq -r '.data[2].id // empty' 2>/dev/null || true)

if [[ -z "$STUOM_TRIGGER_ID" || -z "$STUOM_LIST_ID" ]]; then
    skip "Not enough service type unit of measure IDs found"
    run_tests
else
    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cohort with required fields"
COHORT_NAME=$(unique_name "STUOMCohort")

xbe_json do service-type-unit-of-measure-cohorts create \
    --customer "$CREATED_CUSTOMER_ID" \
    --trigger "$STUOM_TRIGGER_ID" \
    --service-type-unit-of-measure-ids "$STUOM_LIST_ID" \
    --name "$COHORT_NAME" \
    --active=false

if [[ $status -eq 0 ]]; then
    CREATED_COHORT_ID=$(json_get ".id")
    if [[ -n "$CREATED_COHORT_ID" && "$CREATED_COHORT_ID" != "null" ]]; then
        register_cleanup "service-type-unit-of-measure-cohorts" "$CREATED_COHORT_ID"
        pass
    else
        fail "Created cohort but no ID returned"
        run_tests
    fi
else
    fail "Failed to create cohort"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update cohort name"
xbe_json do service-type-unit-of-measure-cohorts update "$CREATED_COHORT_ID" --name "Updated cohort name"
assert_success

test_name "Update cohort to active"
xbe_json do service-type-unit-of-measure-cohorts update "$CREATED_COHORT_ID" --active
assert_success

test_name "Update cohort customer"
xbe_json do service-type-unit-of-measure-cohorts update "$CREATED_COHORT_ID" --customer "$CREATED_CUSTOMER_ALT_ID"
assert_success

test_name "Update cohort trigger and service type unit of measure list"
if [[ -n "$STUOM_ALT_ID" ]]; then
    xbe_json do service-type-unit-of-measure-cohorts update "$CREATED_COHORT_ID" \
        --trigger "$STUOM_LIST_ID" \
        --service-type-unit-of-measure-ids "$STUOM_ALT_ID"
    assert_success
else
    skip "Need at least three service type unit of measure IDs for trigger update"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List cohorts"
xbe_json view service-type-unit-of-measure-cohorts list --limit 5
assert_success

test_name "List cohorts with --customer filter"
xbe_json view service-type-unit-of-measure-cohorts list --customer "$CREATED_CUSTOMER_ALT_ID" --limit 5
assert_success

test_name "List cohorts with --service-type-unit-of-measure-id filter"
filter_id="$STUOM_LIST_ID"
if [[ -n "$STUOM_ALT_ID" ]]; then
    filter_id="$STUOM_ALT_ID"
fi
xbe_json view service-type-unit-of-measure-cohorts list --service-type-unit-of-measure-id "$filter_id" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show cohort"
xbe_json view service-type-unit-of-measure-cohorts show "$CREATED_COHORT_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete cohort requires --confirm flag"
xbe_run do service-type-unit-of-measure-cohorts delete "$CREATED_COHORT_ID"
assert_failure

test_name "Delete cohort with --confirm"
xbe_run do service-type-unit-of-measure-cohorts delete "$CREATED_COHORT_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cohort without required fields fails"
xbe_json do service-type-unit-of-measure-cohorts create
assert_failure

test_name "Update without any fields fails"
xbe_json do service-type-unit-of-measure-cohorts update "$CREATED_COHORT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
