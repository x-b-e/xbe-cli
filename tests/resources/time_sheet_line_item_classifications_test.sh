#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Line Item Classifications
#
# Tests CRUD operations for the time_sheet_line_item_classifications resource.
# Time sheet line item classifications define types of time sheet entries.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CLASSIFICATION_ID=""

describe "Resource: time_sheet_line_item_classifications"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet line item classification with required fields"
TEST_NAME=$(unique_name "TSLIClass")

xbe_json do time-sheet-line-item-classifications create \
    --name "$TEST_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "time-sheet-line-item-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created time sheet line item classification but no ID returned"
    fi
else
    fail "Failed to create time sheet line item classification"
fi

# Only continue if we successfully created a classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid time sheet line item classification ID"
    run_tests
fi

test_name "Create time sheet line item classification with description"
TEST_NAME2=$(unique_name "TSLIClass2")
xbe_json do time-sheet-line-item-classifications create \
    --name "$TEST_NAME2" \
    --description "Hours worked beyond standard schedule"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "time-sheet-line-item-classifications" "$id"
    pass
else
    fail "Failed to create time sheet line item classification with description"
fi

test_name "Create time sheet line item classification with subject-types"
TEST_NAME3=$(unique_name "TSLIClass3")
xbe_json do time-sheet-line-item-classifications create \
    --name "$TEST_NAME3" \
    --subject-types "DriverDay"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "time-sheet-line-item-classifications" "$id"
    pass
else
    fail "Failed to create time sheet line item classification with subject-types"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update time sheet line item classification name"
UPDATED_NAME=$(unique_name "UpdatedTSLIC")
xbe_json do time-sheet-line-item-classifications update "$CREATED_CLASSIFICATION_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update time sheet line item classification description"
xbe_json do time-sheet-line-item-classifications update "$CREATED_CLASSIFICATION_ID" --description "Updated description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet line item classifications"
xbe_json view time-sheet-line-item-classifications list --limit 5
assert_success

test_name "List time sheet line item classifications returns array"
xbe_json view time-sheet-line-item-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time sheet line item classifications"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List time sheet line item classifications with --limit"
xbe_json view time-sheet-line-item-classifications list --limit 3
assert_success

test_name "List time sheet line item classifications with --offset"
xbe_json view time-sheet-line-item-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete time sheet line item classification requires --confirm flag"
xbe_run do time-sheet-line-item-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Delete time sheet line item classification with --confirm"
# Create a classification specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteTSLIC")
xbe_json do time-sheet-line-item-classifications create \
    --name "$TEST_DEL_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do time-sheet-line-item-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create time sheet line item classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create time sheet line item classification without name fails"
xbe_json do time-sheet-line-item-classifications create --description "No name"
assert_failure

test_name "Update without any fields fails"
xbe_json do time-sheet-line-item-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
