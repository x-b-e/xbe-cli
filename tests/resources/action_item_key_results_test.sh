#!/bin/bash
#
# XBE CLI Integration Tests: Action Item Key Results
#
# Tests list/show and create/delete behavior for action-item-key-results.
#
# COVERAGE: List filters + show + create attributes + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: action-item-key-results"

SAMPLE_ID=""
SAMPLE_ACTION_ITEM_ID=""
SAMPLE_KEY_RESULT_ID=""

CREATED_ID=""
ACTION_ITEM_ID="${XBE_TEST_ACTION_ITEM_ID:-}"
KEY_RESULT_ID="${XBE_TEST_KEY_RESULT_ID:-}"
SKIP_MUTATION=0

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping create/delete tests)"
    SKIP_MUTATION=1
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List action item key results"
xbe_json view action-item-key-results list --limit 5
assert_success

test_name "List action item key results returns array"
xbe_json view action-item-key-results list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list action item key results"
fi

test_name "Capture sample action item key result"
xbe_json view action-item-key-results list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ACTION_ITEM_ID=$(json_get ".[0].action_item_id")
    SAMPLE_KEY_RESULT_ID=$(json_get ".[0].key_result_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No action item key results available for follow-on tests"
    fi
else
    skip "Could not list action item key results to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List action item key results with --action-item filter"
if [[ -n "$SAMPLE_ACTION_ITEM_ID" && "$SAMPLE_ACTION_ITEM_ID" != "null" ]]; then
    xbe_json view action-item-key-results list --action-item "$SAMPLE_ACTION_ITEM_ID" --limit 5
    assert_success
else
    skip "No action item ID available"
fi

test_name "List action item key results with --key-result filter"
if [[ -n "$SAMPLE_KEY_RESULT_ID" && "$SAMPLE_KEY_RESULT_ID" != "null" ]]; then
    xbe_json view action-item-key-results list --key-result "$SAMPLE_KEY_RESULT_ID" --limit 5
    assert_success
else
    skip "No key result ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show action item key result"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view action-item-key-results show "$SAMPLE_ID"
    assert_success
else
    skip "No action item key result ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create action item key result without required fields fails"
xbe_json do action-item-key-results create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/delete tests without XBE_TOKEN"
fi

if [[ -z "$ACTION_ITEM_ID" || -z "$KEY_RESULT_ID" ]]; then
    skip "XBE_TEST_ACTION_ITEM_ID or XBE_TEST_KEY_RESULT_ID not set"
else
    test_name "Create action item key result"
    xbe_json do action-item-key-results create \
        --action-item "$ACTION_ITEM_ID" \
        --key-result "$KEY_RESULT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "action-item-key-results" "$CREATED_ID"
            pass
        else
            fail "Created action item key result but no ID returned"
        fi
    else
        fail "Failed to create action item key result"
    fi
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete action item key result"
    xbe_json do action-item-key-results delete "$CREATED_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
