#!/bin/bash
#
# XBE CLI Integration Tests: Rates
#
# Tests operations for the rates resource.
# Rates require complex polymorphic relationships (rated can be broker-tenders,
# customer-tenders, or rate-agreements) and service-type-unit-of-measure.
#
# Note: Full CRUD testing is limited because creating the prerequisite chain
# (tenders, rate-agreements, service-type-unit-of-measures) is complex.
# This test focuses on list operations and error cases.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""

describe "Resource: rates"

# ============================================================================
# Prerequisites - Create a broker for filtering tests
# ============================================================================

test_name "Create prerequisite broker for rate tests"
BROKER_NAME=$(unique_name "RateTestBroker")

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
        fail "Failed to create broker"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List rates"
xbe_json view rates list --limit 5
assert_success

test_name "List rates returns array"
xbe_json view rates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list rates"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List rates with --status filter"
xbe_json view rates list --status active --limit 10
assert_success

test_name "List rates with --name-like filter"
xbe_json view rates list --name-like "hourly" --limit 10
assert_success

# Note: --rated-type and --rated-id filters are skipped because they require
# specific valid polymorphic type values from the API

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List rates with --limit"
xbe_json view rates list --limit 3
assert_success

test_name "List rates with --offset"
xbe_json view rates list --limit 3 --offset 1
assert_success

test_name "List rates with pagination (limit + offset)"
xbe_json view rates list --limit 5 --offset 10
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create rate without rated-type fails"
xbe_json do rates create --rated-id "1" --service-type-unit-of-measure "1"
assert_failure

test_name "Create rate without rated-id fails"
xbe_json do rates create --rated-type broker-tenders --service-type-unit-of-measure "1"
assert_failure

test_name "Create rate without service-type-unit-of-measure fails"
xbe_json do rates create --rated-type broker-tenders --rated-id "1"
assert_failure

test_name "Update rate without any fields fails"
# Use a placeholder ID - it should fail for missing fields before hitting the API
xbe_json do rates update "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
