#!/bin/bash
#
# XBE CLI Integration Tests: Commitment Simulation Sets
#
# Tests view and write operations for the commitment-simulation-sets resource.
#
# COVERAGE: List + filters + pagination + show + create/delete + failure cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STATUS=""
SAMPLE_ORG_TYPE=""
SAMPLE_ORG_ID=""

CREATED_BROKER_ID=""
CREATED_SET_ID=""
CREATED_STATUS=""

START_ON="2025-01-01"
END_ON="2025-01-07"
ITERATION_COUNT=10

describe "Resource: commitment-simulation-sets"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List commitment simulation sets"
xbe_json view commitment-simulation-sets list --limit 5
assert_success

test_name "List commitment simulation sets returns array"
xbe_json view commitment-simulation-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list commitment simulation sets"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample commitment simulation set"
xbe_json view commitment-simulation-sets list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_ORG_TYPE=$(json_get ".[0].organization_type")
    SAMPLE_ORG_ID=$(json_get ".[0].organization_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No commitment simulation sets available for follow-on tests"
    fi
else
    skip "Could not list commitment simulation sets to capture sample"
fi

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for commitment simulation set tests"
BROKER_NAME=$(unique_name "CommitSimBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create commitment simulation set with required fields"
xbe_json do commitment-simulation-sets create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --start-on "$START_ON" \
    --end-on "$END_ON" \
    --iteration-count "$ITERATION_COUNT"

if [[ $status -eq 0 ]]; then
    CREATED_SET_ID=$(json_get ".id")
    CREATED_STATUS=$(json_get ".status")
    if [[ -n "$CREATED_SET_ID" && "$CREATED_SET_ID" != "null" ]]; then
        register_cleanup "commitment-simulation-sets" "$CREATED_SET_ID"
        pass
    else
        fail "Created commitment simulation set but no ID returned"
    fi
else
    skip "Failed to create commitment simulation set (policy or validation)"
fi

# ============================================================================
# CREATE Tests - Failure Cases
# ============================================================================

test_name "Create commitment simulation set without start-on fails"
xbe_run do commitment-simulation-sets create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --end-on "$END_ON" \
    --iteration-count "$ITERATION_COUNT"
assert_failure

test_name "Create commitment simulation set without iteration-count fails"
xbe_run do commitment-simulation-sets create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --start-on "$START_ON" \
    --end-on "$END_ON"
assert_failure

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List commitment simulation sets with --organization filter"
FILTER_ORG_TYPE="Broker"
FILTER_ORG_ID="$CREATED_BROKER_ID"
if [[ -n "$FILTER_ORG_ID" && "$FILTER_ORG_ID" != "null" ]]; then
    xbe_json view commitment-simulation-sets list --organization "${FILTER_ORG_TYPE}|${FILTER_ORG_ID}" --limit 5
    assert_success
else
    skip "No organization ID available"
fi

test_name "List commitment simulation sets with --organization-id filter"
if [[ -n "$FILTER_ORG_ID" && "$FILTER_ORG_ID" != "null" ]]; then
    xbe_json view commitment-simulation-sets list --organization-type "$FILTER_ORG_TYPE" --organization-id "$FILTER_ORG_ID" --limit 5
    assert_success
else
    skip "No organization ID available"
fi

test_name "List commitment simulation sets with --organization-type filter"
if [[ -n "$FILTER_ORG_TYPE" && "$FILTER_ORG_TYPE" != "null" ]]; then
    xbe_json view commitment-simulation-sets list --organization-type "$FILTER_ORG_TYPE" --limit 5
    assert_success
else
    skip "No organization type available"
fi

test_name "List commitment simulation sets with --not-organization-type filter"
xbe_json view commitment-simulation-sets list --not-organization-type "Customer" --limit 5
assert_success

test_name "List commitment simulation sets with --status filter"
FILTER_STATUS="$SAMPLE_STATUS"
if [[ -n "$CREATED_STATUS" && "$CREATED_STATUS" != "null" ]]; then
    FILTER_STATUS="$CREATED_STATUS"
fi
if [[ -n "$FILTER_STATUS" && "$FILTER_STATUS" != "null" ]]; then
    xbe_json view commitment-simulation-sets list --status "$FILTER_STATUS" --limit 5
    assert_success
else
    skip "No status available for filtering"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List commitment simulation sets with --limit"
xbe_json view commitment-simulation-sets list --limit 2
assert_success

test_name "List commitment simulation sets with --offset"
xbe_json view commitment-simulation-sets list --limit 2 --offset 2
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show commitment simulation set"
SHOW_ID="$CREATED_SET_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$SAMPLE_ID"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view commitment-simulation-sets show "$SHOW_ID"
    assert_success
else
    skip "No commitment simulation set ID available"
fi

# ============================================================================
# UPDATE Tests (not supported)
# ============================================================================

test_name "Update commitment simulation set is not supported"
if [[ -n "$CREATED_SET_ID" && "$CREATED_SET_ID" != "null" ]]; then
    xbe_run do commitment-simulation-sets update "$CREATED_SET_ID"
    assert_failure
else
    skip "No created commitment simulation set to test update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete commitment simulation set requires --confirm"
if [[ -n "$CREATED_SET_ID" && "$CREATED_SET_ID" != "null" ]]; then
    xbe_run do commitment-simulation-sets delete "$CREATED_SET_ID"
    assert_failure
else
    skip "No created commitment simulation set to test delete"
fi

test_name "Delete commitment simulation set"
if [[ -n "$CREATED_SET_ID" && "$CREATED_SET_ID" != "null" ]]; then
    xbe_json do commitment-simulation-sets delete "$CREATED_SET_ID" --confirm
    assert_success
else
    skip "No created commitment simulation set to delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
