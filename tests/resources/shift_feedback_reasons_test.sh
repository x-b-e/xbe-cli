#!/bin/bash
#
# XBE CLI Integration Tests: Shift Feedback Reasons
#
# Tests CRUD operations for the shift_feedback_reasons resource.
# Shift feedback reasons define categories of feedback for shift performance.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REASON_ID=""

describe "Resource: shift_feedback_reasons"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create shift feedback reason with required fields"
TEST_NAME=$(unique_name "ShiftFeedback")
TEST_SLUG=$(echo "$TEST_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')

xbe_json do shift-feedback-reasons create \
    --name "$TEST_NAME" \
    --kind "driver" \
    --slug "$TEST_SLUG"

if [[ $status -eq 0 ]]; then
    CREATED_REASON_ID=$(json_get ".id")
    if [[ -n "$CREATED_REASON_ID" && "$CREATED_REASON_ID" != "null" ]]; then
        register_cleanup "shift-feedback-reasons" "$CREATED_REASON_ID"
        pass
    else
        fail "Created shift feedback reason but no ID returned"
    fi
else
    fail "Failed to create shift feedback reason"
fi

# Only continue if we successfully created a reason
if [[ -z "$CREATED_REASON_ID" || "$CREATED_REASON_ID" == "null" ]]; then
    echo "Cannot continue without a valid shift feedback reason ID"
    run_tests
fi

test_name "Create shift feedback reason with equipment kind"
TEST_NAME2=$(unique_name "ShiftFeedback2")
TEST_SLUG2=$(echo "$TEST_NAME2" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do shift-feedback-reasons create \
    --name "$TEST_NAME2" \
    --kind "equipment" \
    --slug "$TEST_SLUG2"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "shift-feedback-reasons" "$id"
    pass
else
    fail "Failed to create shift feedback reason with equipment kind"
fi

test_name "Create shift feedback reason with corrective-action"
TEST_NAME3=$(unique_name "ShiftFeedback3")
TEST_SLUG3=$(echo "$TEST_NAME3" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do shift-feedback-reasons create \
    --name "$TEST_NAME3" \
    --kind "equipment" \
    --slug "$TEST_SLUG3" \
    --corrective-action "Please improve this behavior"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "shift-feedback-reasons" "$id"
    pass
else
    fail "Failed to create shift feedback reason with corrective-action"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update shift feedback reason name"
UPDATED_NAME=$(unique_name "UpdatedSFR")
xbe_json do shift-feedback-reasons update "$CREATED_REASON_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update shift feedback reason corrective-action"
xbe_json do shift-feedback-reasons update "$CREATED_REASON_ID" --corrective-action "Updated action"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List shift feedback reasons"
xbe_json view shift-feedback-reasons list --limit 5
assert_success

test_name "List shift feedback reasons returns array"
xbe_json view shift-feedback-reasons list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list shift feedback reasons"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List shift feedback reasons with --kind filter"
xbe_json view shift-feedback-reasons list --kind "driver" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List shift feedback reasons with --limit"
xbe_json view shift-feedback-reasons list --limit 3
assert_success

test_name "List shift feedback reasons with --offset"
xbe_json view shift-feedback-reasons list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete shift feedback reason requires --confirm flag"
xbe_run do shift-feedback-reasons delete "$CREATED_REASON_ID"
assert_failure

test_name "Delete shift feedback reason with --confirm"
# Create a reason specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteSFR")
TEST_DEL_SLUG=$(echo "$TEST_DEL_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do shift-feedback-reasons create \
    --name "$TEST_DEL_NAME" \
    --kind "driver" \
    --slug "$TEST_DEL_SLUG"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do shift-feedback-reasons delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create shift feedback reason for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create shift feedback reason without name fails"
xbe_json do shift-feedback-reasons create --kind "driver" --slug "no-name"
assert_failure

test_name "Create shift feedback reason without kind fails"
xbe_json do shift-feedback-reasons create --name "NoKind" --slug "no-kind"
assert_failure

test_name "Create shift feedback reason without slug fails"
xbe_json do shift-feedback-reasons create --name "NoSlug" --kind "driver"
assert_failure

test_name "Update without any fields fails"
xbe_json do shift-feedback-reasons update "$CREATED_REASON_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
