#!/bin/bash
#
# XBE CLI Integration Tests: Incidents
#
# Tests list operations for the incidents resource.
# Note: Creating incidents requires using specific subclass types
# (safety-incidents, production-incidents, etc.) and is not supported
# through the base incidents resource.
#
# COVERAGE: List with common filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""

describe "Resource: incidents"

# ============================================================================
# Prerequisites - Create broker for filter tests
# ============================================================================

test_name "Create prerequisite broker for incident tests"
BROKER_NAME=$(unique_name "IncidentTestBroker")

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
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List incidents"
xbe_json view incidents list --limit 5
assert_success

test_name "List incidents returns array"
xbe_json view incidents list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list incidents"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List incidents with --broker filter"
xbe_json view incidents list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List incidents with --status filter"
xbe_json view incidents list --status "open" --limit 10
assert_success

test_name "List incidents with --severity filter"
xbe_json view incidents list --severity "high" --limit 10
assert_success

test_name "List incidents with --start-on-min filter"
xbe_json view incidents list --start-on-min "2024-01-01" --limit 10
assert_success

test_name "List incidents with --start-on-max filter"
xbe_json view incidents list --start-on-max "2025-12-31" --limit 10
assert_success

test_name "List incidents with --start-on-min and --start-on-max filter"
xbe_json view incidents list --start-on-min "2024-01-01" --start-on-max "2025-12-31" --limit 10
assert_success

test_name "List incidents with --has-parent false filter"
xbe_json view incidents list --has-parent "false" --limit 10
assert_success

test_name "List incidents with --q search filter"
xbe_json view incidents list --q "test" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List incidents with --limit"
xbe_json view incidents list --limit 3
assert_success

test_name "List incidents with --offset"
xbe_json view incidents list --limit 3 --offset 3
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
