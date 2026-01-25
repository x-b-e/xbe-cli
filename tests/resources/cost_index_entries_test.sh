#!/bin/bash
#
# XBE CLI Integration Tests: Cost Index Entries
#
# Tests CRUD operations for the cost_index_entries resource.
# Cost index entries are time-period values within a cost index.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ENTRY_ID=""
CREATED_BROKER_ID=""
CREATED_COST_INDEX_ID=""

describe "Resource: cost_index_entries"

# ============================================================================
# Prerequisites - Create broker and cost index
# ============================================================================

test_name "Create prerequisite broker for cost index entries tests"
BROKER_NAME=$(unique_name "CIETestBroker")

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

test_name "Create prerequisite cost index"
INDEX_NAME=$(unique_name "CIETestIndex")

xbe_json do cost-indexes create \
    --name "$INDEX_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_COST_INDEX_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_INDEX_ID" && "$CREATED_COST_INDEX_ID" != "null" ]]; then
        register_cleanup "cost-indexes" "$CREATED_COST_INDEX_ID"
        pass
    else
        fail "Created cost index but no ID returned"
        echo "Cannot continue without a cost index"
        run_tests
    fi
else
    fail "Failed to create cost index"
    echo "Cannot continue without a cost index"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cost index entry with required fields"

xbe_json do cost-index-entries create \
    --cost-index "$CREATED_COST_INDEX_ID" \
    --start-on "2024-01-01" \
    --end-on "2024-03-31" \
    --value "1.05"

if [[ $status -eq 0 ]]; then
    CREATED_ENTRY_ID=$(json_get ".id")
    if [[ -n "$CREATED_ENTRY_ID" && "$CREATED_ENTRY_ID" != "null" ]]; then
        register_cleanup "cost-index-entries" "$CREATED_ENTRY_ID"
        pass
    else
        fail "Created cost index entry but no ID returned"
    fi
else
    fail "Failed to create cost index entry"
fi

# Only continue if we successfully created an entry
if [[ -z "$CREATED_ENTRY_ID" || "$CREATED_ENTRY_ID" == "null" ]]; then
    echo "Cannot continue without a valid cost index entry ID"
    run_tests
fi

test_name "Create cost index entry with different value"
xbe_json do cost-index-entries create \
    --cost-index "$CREATED_COST_INDEX_ID" \
    --start-on "2024-07-01" \
    --end-on "2024-09-30" \
    --value "1.15"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-index-entries" "$id"
    pass
else
    fail "Failed to create cost index entry with different value"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update cost index entry start-on"
xbe_json do cost-index-entries update "$CREATED_ENTRY_ID" --start-on "2024-01-15"
assert_success

test_name "Update cost index entry end-on"
xbe_json do cost-index-entries update "$CREATED_ENTRY_ID" --end-on "2024-03-31"
assert_success

test_name "Update cost index entry value"
xbe_json do cost-index-entries update "$CREATED_ENTRY_ID" --value "1.08"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List cost index entries"
xbe_json view cost-index-entries list --limit 5
assert_success

test_name "List cost index entries returns array"
xbe_json view cost-index-entries list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list cost index entries"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List cost index entries with --cost-index filter"
xbe_json view cost-index-entries list --cost-index "$CREATED_COST_INDEX_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List cost index entries with --limit"
xbe_json view cost-index-entries list --limit 3
assert_success

test_name "List cost index entries with --offset"
xbe_json view cost-index-entries list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete cost index entry requires --confirm flag"
xbe_run do cost-index-entries delete "$CREATED_ENTRY_ID"
assert_failure

test_name "Delete cost index entry with --confirm"
# Create an entry specifically for deletion
xbe_json do cost-index-entries create \
    --cost-index "$CREATED_COST_INDEX_ID" \
    --start-on "2025-01-01" \
    --end-on "2025-03-31" \
    --value "1.20"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do cost-index-entries delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create cost index entry for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cost index entry without cost-index fails"
xbe_json do cost-index-entries create --start-on "2024-01-01" --end-on "2024-03-31" --value "1.0"
assert_failure

test_name "Create cost index entry without start-on fails"
xbe_json do cost-index-entries create --cost-index "$CREATED_COST_INDEX_ID" --end-on "2024-03-31" --value "1.0"
assert_failure

test_name "Create cost index entry without end-on fails"
xbe_json do cost-index-entries create --cost-index "$CREATED_COST_INDEX_ID" --start-on "2024-01-01" --value "1.0"
assert_failure

test_name "Create cost index entry without value fails"
xbe_json do cost-index-entries create --cost-index "$CREATED_COST_INDEX_ID" --start-on "2024-01-01" --end-on "2024-03-31"
assert_failure

test_name "Update without any fields fails"
xbe_json do cost-index-entries update "$CREATED_ENTRY_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
