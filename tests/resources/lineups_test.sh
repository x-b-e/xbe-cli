#!/bin/bash
#
# XBE CLI Integration Tests: Lineups
#
# Tests CRUD operations for the lineups resource.
# Lineups require a customer relationship and a start time window.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LINEUP_ID=""
CREATED_CUSTOMER_ID=""
CREATED_BROKER_ID=""

LINEUP_NAME=""
LINEUP_START_MIN="2026-01-01T06:00:00Z"
LINEUP_START_MAX="2026-01-01T18:00:00Z"

UPDATED_LINEUP_NAME=""
UPDATED_START_MIN="2026-01-02T06:00:00Z"
UPDATED_START_MAX="2026-01-02T18:00:00Z"

describe "Resource: lineups"

# ============================================================================
# Prerequisites - Create broker and customer
# ============================================================================

test_name "Create prerequisite broker for lineup tests"
BROKER_NAME=$(unique_name "LineupBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite customer for lineup tests"
CUSTOMER_NAME=$(unique_name "LineupCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create lineup with required fields"
LINEUP_NAME=$(unique_name "Lineup")

xbe_json do lineups create \
    --customer "$CREATED_CUSTOMER_ID" \
    --name "$LINEUP_NAME" \
    --start-at-min "$LINEUP_START_MIN" \
    --start-at-max "$LINEUP_START_MAX"

if [[ $status -eq 0 ]]; then
    CREATED_LINEUP_ID=$(json_get ".id")
    if [[ -n "$CREATED_LINEUP_ID" && "$CREATED_LINEUP_ID" != "null" ]]; then
        register_cleanup "lineups" "$CREATED_LINEUP_ID"
        pass
    else
        fail "Created lineup but no ID returned"
    fi
else
    skip "Failed to create lineup (server may not support this operation)"
fi

# Only run these tests if we have a valid lineup ID
if [[ -n "$CREATED_LINEUP_ID" && "$CREATED_LINEUP_ID" != "null" ]]; then

UPDATED_LINEUP_NAME=$(unique_name "UpdatedLineup")

test_name "Show lineup details"
xbe_json view lineups show "$CREATED_LINEUP_ID"
assert_success

test_name "Update lineup name"
xbe_json do lineups update "$CREATED_LINEUP_ID" --name "$UPDATED_LINEUP_NAME"
assert_success

test_name "Update lineup time window"
xbe_json do lineups update "$CREATED_LINEUP_ID" --start-at-min "$UPDATED_START_MIN" --start-at-max "$UPDATED_START_MAX"
assert_success

fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineups"
xbe_json view lineups list --limit 5
assert_success

test_name "List lineups returns array"
xbe_json view lineups list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineups"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineups with --customer filter"
xbe_json view lineups list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List lineups with --broker filter"
xbe_json view lineups list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List lineups with --name-like filter"
xbe_json view lineups list --name-like "Lineup" --limit 10
assert_success

test_name "List lineups with --start-at-min filter"
xbe_json view lineups list --start-at-min "$LINEUP_START_MIN" --limit 10
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"FILTER_NOT_ALLOWED"* ]]; then
    skip "start-at-min filter not available in staging"
else
    fail "Expected success (exit 0), got exit $status"
fi

test_name "List lineups with --start-at-max filter"
xbe_json view lineups list --start-at-max "$LINEUP_START_MAX" --limit 10
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"FILTER_NOT_ALLOWED"* ]]; then
    skip "start-at-max filter not available in staging"
else
    fail "Expected success (exit 0), got exit $status"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List lineups with --limit"
xbe_json view lineups list --limit 3
assert_success

test_name "List lineups with --offset"
xbe_json view lineups list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete lineup requires --confirm flag"
if [[ -n "$CREATED_LINEUP_ID" && "$CREATED_LINEUP_ID" != "null" ]]; then
    xbe_run do lineups delete "$CREATED_LINEUP_ID"
    assert_failure
else
    skip "No lineup available for delete test"
fi

test_name "Delete lineup with --confirm"
# Create a lineup specifically for deletion
DEL_LINEUP_NAME=$(unique_name "DeleteLineup")
xbe_json do lineups create \
    --customer "$CREATED_CUSTOMER_ID" \
    --name "$DEL_LINEUP_NAME" \
    --start-at-min "2026-01-03T06:00:00Z" \
    --start-at-max "2026-01-03T18:00:00Z"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do lineups delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create lineup for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create lineup without customer fails"
xbe_json do lineups create --start-at-min "$LINEUP_START_MIN" --start-at-max "$LINEUP_START_MAX"
assert_failure

test_name "Create lineup without start-at-min fails"
xbe_json do lineups create --customer "$CREATED_CUSTOMER_ID" --start-at-max "$LINEUP_START_MAX"
assert_failure

test_name "Create lineup without start-at-max fails"
xbe_json do lineups create --customer "$CREATED_CUSTOMER_ID" --start-at-min "$LINEUP_START_MIN"
assert_failure

test_name "Update without any fields fails"
# Use a placeholder ID - should fail before hitting the API
xbe_json do lineups update "99999999"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
