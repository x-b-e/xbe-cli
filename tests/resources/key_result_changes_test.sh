#!/bin/bash
#
# XBE CLI Integration Tests: Key Result Changes
#
# Tests list and show operations for the key-result-changes resource.
# Key result changes track updates to key result schedules.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

KEY_RESULT_CHANGE_ID=""
KEY_RESULT_ID=""
OBJECTIVE_ID=""
BROKER_ID=""
ORGANIZATION_ID=""
ORGANIZATION_TYPE=""
ORGANIZATION_TYPE_FILTER=""
CHANGED_BY_ID=""
SKIP_ID_FILTERS=0

describe "Resource: key-result-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List key result changes"
xbe_json view key-result-changes list --limit 5
assert_success

test_name "List key result changes returns array"
xbe_json view key-result-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list key result changes"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample key result change"
xbe_json view key-result-changes list --limit 1
if [[ $status -eq 0 ]]; then
    KEY_RESULT_CHANGE_ID=$(json_get ".[0].id")
    KEY_RESULT_ID=$(json_get ".[0].key_result_id")
    OBJECTIVE_ID=$(json_get ".[0].objective_id")
    BROKER_ID=$(json_get ".[0].broker_id")
    ORGANIZATION_ID=$(json_get ".[0].organization_id")
    ORGANIZATION_TYPE=$(json_get ".[0].organization_type")
    CHANGED_BY_ID=$(json_get ".[0].changed_by_id")
    case "$ORGANIZATION_TYPE" in
        brokers)
            ORGANIZATION_TYPE_FILTER="Broker"
            ;;
        customers)
            ORGANIZATION_TYPE_FILTER="Customer"
            ;;
        truckers)
            ORGANIZATION_TYPE_FILTER="Trucker"
            ;;
        developers)
            ORGANIZATION_TYPE_FILTER="Developer"
            ;;
        material-suppliers)
            ORGANIZATION_TYPE_FILTER="MaterialSupplier"
            ;;
        *)
            ORGANIZATION_TYPE_FILTER=""
            ;;
    esac
    if [[ -n "$KEY_RESULT_CHANGE_ID" && "$KEY_RESULT_CHANGE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No key result changes available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list key result changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

NOW_ISO=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

test_name "List key result changes with --created-at-min filter"
xbe_json view key-result-changes list --created-at-min "$NOW_ISO" --limit 5
assert_success

test_name "List key result changes with --created-at-max filter"
xbe_json view key-result-changes list --created-at-max "$NOW_ISO" --limit 5
assert_success

test_name "List key result changes with --updated-at-min filter"
xbe_json view key-result-changes list --updated-at-min "$NOW_ISO" --limit 5
assert_success

test_name "List key result changes with --updated-at-max filter"
xbe_json view key-result-changes list --updated-at-max "$NOW_ISO" --limit 5
assert_success

test_name "List key result changes with --key-result filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$KEY_RESULT_ID" && "$KEY_RESULT_ID" != "null" ]]; then
    xbe_json view key-result-changes list --key-result "$KEY_RESULT_ID" --limit 5
    assert_success
else
    skip "No key result ID available"
fi

test_name "List key result changes with --objective filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$OBJECTIVE_ID" && "$OBJECTIVE_ID" != "null" ]]; then
    xbe_json view key-result-changes list --objective "$OBJECTIVE_ID" --limit 5
    assert_success
else
    skip "No objective ID available"
fi

test_name "List key result changes with --broker filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view key-result-changes list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List key result changes with --organization filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" && -n "$ORGANIZATION_TYPE_FILTER" ]]; then
    xbe_json view key-result-changes list --organization "${ORGANIZATION_TYPE_FILTER}|${ORGANIZATION_ID}" --limit 5
    assert_success
else
    skip "No organization available"
fi

test_name "List key result changes with --changed-by filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$CHANGED_BY_ID" && "$CHANGED_BY_ID" != "null" ]]; then
    xbe_json view key-result-changes list --changed-by "$CHANGED_BY_ID" --limit 5
    assert_success
else
    skip "No changed-by ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show key result change"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$KEY_RESULT_CHANGE_ID" && "$KEY_RESULT_CHANGE_ID" != "null" ]]; then
    xbe_json view key-result-changes show "$KEY_RESULT_CHANGE_ID"
    assert_success
else
    skip "No key result change ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
