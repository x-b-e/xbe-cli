#!/bin/bash
#
# XBE CLI Integration Tests: Cost Code Trucking Cost Summaries
#
# Tests list/show/create/update/delete operations for the cost-code-trucking-cost-summaries resource.
#
# COVERAGE: All list filters + create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUMMARY_ID=""
CREATED_BROKER_ID=""
CREATED_BY_ID=""
SAMPLE_ID=""

START_ON="2025-01-01"
END_ON="2025-01-31"
UPDATE_END_ON="2025-02-28"
NOW_ISO=$(date -u +%Y-%m-%dT%H:%M:%SZ)


describe "Resource: cost-code-trucking-cost-summaries"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "CCTCSBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cost code trucking cost summary"
xbe_json do cost-code-trucking-cost-summaries create \
    --broker "$CREATED_BROKER_ID" \
    --start-on "$START_ON" \
    --end-on "$END_ON"

if [[ $status -eq 0 ]]; then
    CREATED_SUMMARY_ID=$(json_get ".id")
    if [[ -n "$CREATED_SUMMARY_ID" && "$CREATED_SUMMARY_ID" != "null" ]]; then
        register_cleanup "cost-code-trucking-cost-summaries" "$CREATED_SUMMARY_ID"
        pass
    else
        fail "Created summary but no ID returned"
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
        pass
    else
        fail "Failed to create cost code trucking cost summary"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List cost code trucking cost summaries"
xbe_json view cost-code-trucking-cost-summaries list --limit 5
assert_success


test_name "List cost code trucking cost summaries returns array"
xbe_json view cost-code-trucking-cost-summaries list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
else
    fail "Failed to list cost code trucking cost summaries"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show cost code trucking cost summary"
SHOW_ID="${CREATED_SUMMARY_ID:-$SAMPLE_ID}"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view cost-code-trucking-cost-summaries show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".created_by_id")
        pass
    else
        fail "Failed to show cost code trucking cost summary"
    fi
else
    skip "No summary ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List summaries with --broker filter"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json view cost-code-trucking-cost-summaries list --broker "$CREATED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter test"
fi


test_name "List summaries with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view cost-code-trucking-cost-summaries list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available for filter test"
fi


test_name "List summaries with --created-at-min filter"
xbe_json view cost-code-trucking-cost-summaries list --created-at-min "$NOW_ISO" --limit 5
assert_success


test_name "List summaries with --created-at-max filter"
xbe_json view cost-code-trucking-cost-summaries list --created-at-max "$NOW_ISO" --limit 5
assert_success


test_name "List summaries with --is-created-at filter"
xbe_json view cost-code-trucking-cost-summaries list --is-created-at true --limit 5
assert_success


test_name "List summaries with --updated-at-min filter"
xbe_json view cost-code-trucking-cost-summaries list --updated-at-min "$NOW_ISO" --limit 5
assert_success


test_name "List summaries with --updated-at-max filter"
xbe_json view cost-code-trucking-cost-summaries list --updated-at-max "$NOW_ISO" --limit 5
assert_success


test_name "List summaries with --is-updated-at filter"
xbe_json view cost-code-trucking-cost-summaries list --is-updated-at true --limit 5
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update cost code trucking cost summary fails (immutable)"
if [[ -n "$CREATED_SUMMARY_ID" && "$CREATED_SUMMARY_ID" != "null" ]]; then
    xbe_run do cost-code-trucking-cost-summaries update "$CREATED_SUMMARY_ID" --end-on "$UPDATE_END_ON"
    assert_failure
else
    skip "No created summary ID available for update test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete cost code trucking cost summary"
if [[ -n "$CREATED_SUMMARY_ID" && "$CREATED_SUMMARY_ID" != "null" ]]; then
    xbe_run do cost-code-trucking-cost-summaries delete "$CREATED_SUMMARY_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to delete cost code trucking cost summary: $output"
        fi
    fi
else
    skip "No created summary ID available for delete test"
fi

run_tests
