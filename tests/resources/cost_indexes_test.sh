#!/bin/bash
#
# XBE CLI Integration Tests: Cost Indexes and Cost Index Entries
#
# Tests CRUD operations for the cost_indexes and cost_index_entries resources.
# Cost indexes define pricing indexes used for rate adjustments.
# Cost index entries are the actual values for a cost index over time.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_COST_INDEX_ID=""
CREATED_COST_INDEX_ENTRY_ID=""
CREATED_BROKER_ID=""

describe "Resource: cost_indexes and cost_index_entries"

# ============================================================================
# Prerequisites - Create broker for broker-specific indexes
# ============================================================================

test_name "Create prerequisite broker for cost index tests"
BROKER_NAME=$(unique_name "CITestBroker")

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
# COST INDEX CREATE Tests
# ============================================================================

test_name "Create cost index with required fields"
TEST_NAME=$(unique_name "CostIndex")

xbe_json do cost-indexes create \
    --name "$TEST_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_COST_INDEX_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_INDEX_ID" && "$CREATED_COST_INDEX_ID" != "null" ]]; then
        register_cleanup "cost-indexes" "$CREATED_COST_INDEX_ID"
        pass
    else
        fail "Created cost index but no ID returned"
    fi
else
    fail "Failed to create cost index"
fi

# Only continue if we successfully created a cost index
if [[ -z "$CREATED_COST_INDEX_ID" || "$CREATED_COST_INDEX_ID" == "null" ]]; then
    echo "Cannot continue without a valid cost index ID"
    run_tests
fi

test_name "Create cost index with description"
TEST_NAME2=$(unique_name "CostIndex2")
xbe_json do cost-indexes create \
    --name "$TEST_NAME2" \
    --description "A test cost index with description"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-indexes" "$id"
    pass
else
    fail "Failed to create cost index with description"
fi

test_name "Create cost index with url"
TEST_NAME3=$(unique_name "CostIndex3")
xbe_json do cost-indexes create \
    --name "$TEST_NAME3" \
    --url "https://example.com/cost-index"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-indexes" "$id"
    pass
else
    fail "Failed to create cost index with url"
fi

test_name "Create cost index with broker"
TEST_NAME4=$(unique_name "CostIndex4")
xbe_json do cost-indexes create \
    --name "$TEST_NAME4" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-indexes" "$id"
    pass
else
    fail "Failed to create cost index with broker"
fi

test_name "Create cost index with all optional fields"
TEST_NAME5=$(unique_name "CostIndex5")
xbe_json do cost-indexes create \
    --name "$TEST_NAME5" \
    --description "Full test cost index" \
    --url "https://example.com/full-index" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-indexes" "$id"
    pass
else
    fail "Failed to create cost index with all optional fields"
fi

# ============================================================================
# COST INDEX UPDATE Tests
# ============================================================================

test_name "Update cost index name"
UPDATED_NAME=$(unique_name "UpdatedCI")
xbe_json do cost-indexes update "$CREATED_COST_INDEX_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update cost index description"
xbe_json do cost-indexes update "$CREATED_COST_INDEX_ID" --description "Updated description"
assert_success

test_name "Update cost index url"
xbe_json do cost-indexes update "$CREATED_COST_INDEX_ID" --url "https://updated-example.com"
assert_success

# ============================================================================
# COST INDEX LIST Tests - Basic
# ============================================================================

test_name "List cost indexes"
xbe_json view cost-indexes list --limit 5
assert_success

test_name "List cost indexes returns array"
xbe_json view cost-indexes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list cost indexes"
fi

# ============================================================================
# COST INDEX LIST Tests - Filters
# ============================================================================

test_name "List cost indexes with --broker filter"
xbe_json view cost-indexes list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List cost indexes with --is-broker false filter"
xbe_json view cost-indexes list --is-broker "false" --limit 10
assert_success

test_name "List cost indexes with --is-expired false filter"
xbe_json view cost-indexes list --is-expired "false" --limit 10
assert_success

# ============================================================================
# COST INDEX LIST Tests - Pagination
# ============================================================================

test_name "List cost indexes with --limit"
xbe_json view cost-indexes list --limit 3
assert_success

test_name "List cost indexes with --offset"
xbe_json view cost-indexes list --limit 3 --offset 3
assert_success

# ============================================================================
# COST INDEX ENTRY CREATE Tests
# ============================================================================

test_name "Create cost index entry with required fields"

xbe_json do cost-index-entries create \
    --cost-index "$CREATED_COST_INDEX_ID" \
    --start-on "2024-01-01" \
    --end-on "2024-03-31" \
    --value "1.05"

if [[ $status -eq 0 ]]; then
    CREATED_COST_INDEX_ENTRY_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_INDEX_ENTRY_ID" && "$CREATED_COST_INDEX_ENTRY_ID" != "null" ]]; then
        register_cleanup "cost-index-entries" "$CREATED_COST_INDEX_ENTRY_ID"
        pass
    else
        fail "Created cost index entry but no ID returned"
    fi
else
    fail "Failed to create cost index entry"
fi

test_name "Create cost index entry with end-on"
xbe_json do cost-index-entries create \
    --cost-index "$CREATED_COST_INDEX_ID" \
    --start-on "2024-04-01" \
    --end-on "2024-06-30" \
    --value "1.08"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-index-entries" "$id"
    pass
else
    fail "Failed to create cost index entry with end-on"
fi

# ============================================================================
# COST INDEX ENTRY UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_COST_INDEX_ENTRY_ID" && "$CREATED_COST_INDEX_ENTRY_ID" != "null" ]]; then
    test_name "Update cost index entry value"
    xbe_json do cost-index-entries update "$CREATED_COST_INDEX_ENTRY_ID" --value "1.10"
    assert_success

    test_name "Update cost index entry end-on"
    xbe_json do cost-index-entries update "$CREATED_COST_INDEX_ENTRY_ID" --end-on "2024-03-31"
    assert_success
fi

# ============================================================================
# COST INDEX ENTRY LIST Tests
# ============================================================================

test_name "List cost index entries"
xbe_json view cost-index-entries list --limit 5
assert_success

test_name "List cost index entries with --cost-index filter"
xbe_json view cost-index-entries list --cost-index "$CREATED_COST_INDEX_ID" --limit 10
assert_success

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
if [[ -n "$CREATED_COST_INDEX_ENTRY_ID" && "$CREATED_COST_INDEX_ENTRY_ID" != "null" ]]; then
    xbe_run do cost-index-entries delete "$CREATED_COST_INDEX_ENTRY_ID"
    assert_failure
else
    skip "No cost index entry ID for delete test"
fi

test_name "Delete cost index entry with --confirm"
# Create an entry specifically for deletion
xbe_json do cost-index-entries create \
    --cost-index "$CREATED_COST_INDEX_ID" \
    --start-on "2025-01-01" \
    --value "1.15"
if [[ $status -eq 0 ]]; then
    DEL_ENTRY_ID=$(json_get ".id")
    xbe_run do cost-index-entries delete "$DEL_ENTRY_ID" --confirm
    assert_success
else
    skip "Could not create cost index entry for deletion test"
fi

test_name "Delete cost index requires --confirm flag"
xbe_run do cost-indexes delete "$CREATED_COST_INDEX_ID"
assert_failure

test_name "Delete cost index with --confirm"
# Create a cost index specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteCI")
xbe_json do cost-indexes create \
    --name "$TEST_DEL_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do cost-indexes delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create cost index for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cost index without name fails"
xbe_json do cost-indexes create --description "No name"
assert_failure

test_name "Create cost index entry without cost-index fails"
xbe_json do cost-index-entries create --start-on "2024-01-01" --value "1.00"
assert_failure

test_name "Create cost index entry without start-on fails"
xbe_json do cost-index-entries create --cost-index "$CREATED_COST_INDEX_ID" --value "1.00"
assert_failure

test_name "Create cost index entry without value fails"
xbe_json do cost-index-entries create --cost-index "$CREATED_COST_INDEX_ID" --start-on "2024-01-01"
assert_failure

test_name "Update cost index without any fields fails"
xbe_json do cost-indexes update "$CREATED_COST_INDEX_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
