#!/bin/bash
#
# XBE CLI Integration Tests: UI Tours
#
# Tests CRUD operations for the ui_tours resource.
# UI tours define guided walkthroughs for users.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_UI_TOUR_ID=""

describe "Resource: ui-tours"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create UI tour with required fields"
TEST_NAME=$(unique_name "UiTour")
TEST_ABBREV=$(echo "$TEST_NAME" | tr '[:upper:]' '[:lower:]' | cut -c1-32)

xbe_json do ui-tours create --name "$TEST_NAME" --abbreviation "$TEST_ABBREV"

if [[ $status -eq 0 ]]; then
    CREATED_UI_TOUR_ID=$(json_get ".id")
    if [[ -n "$CREATED_UI_TOUR_ID" && "$CREATED_UI_TOUR_ID" != "null" ]]; then
        register_cleanup "ui-tours" "$CREATED_UI_TOUR_ID"
        pass
    else
        fail "Created UI tour but no ID returned"
    fi
else
    fail "Failed to create UI tour"
fi

# Only continue if we successfully created a UI tour
if [[ -z "$CREATED_UI_TOUR_ID" || "$CREATED_UI_TOUR_ID" == "null" ]]; then
    echo "Cannot continue without a valid UI tour ID"
    run_tests
fi

test_name "Create UI tour with description"
TEST_NAME2=$(unique_name "UiTour2")
TEST_ABBREV2=$(echo "$TEST_NAME2" | tr '[:upper:]' '[:lower:]' | cut -c1-32)
xbe_json do ui-tours create \
    --name "$TEST_NAME2" \
    --abbreviation "$TEST_ABBREV2" \
    --description "Tour for testing"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "ui-tours" "$id"
    pass
else
    fail "Failed to create UI tour with description"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update UI tour name"
UPDATED_NAME=$(unique_name "UpdatedUiTour")
xbe_json do ui-tours update "$CREATED_UI_TOUR_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update UI tour abbreviation"
UPDATED_ABBREV=$(echo "$UPDATED_NAME" | tr '[:upper:]' '[:lower:]' | cut -c1-28)
xbe_json do ui-tours update "$CREATED_UI_TOUR_ID" --abbreviation "$UPDATED_ABBREV"
assert_success

test_name "Update UI tour description"
xbe_json do ui-tours update "$CREATED_UI_TOUR_ID" --description "Updated description"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show UI tour details"
xbe_json view ui-tours show "$CREATED_UI_TOUR_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List UI tours"
xbe_json view ui-tours list --limit 5
assert_success

test_name "List UI tours returns array"
xbe_json view ui-tours list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list UI tours"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List UI tours with --abbreviation filter"
xbe_json view ui-tours list --abbreviation "$UPDATED_ABBREV"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List UI tours with --limit"
xbe_json view ui-tours list --limit 3
assert_success

test_name "List UI tours with --offset"
xbe_json view ui-tours list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete UI tour requires --confirm flag"
xbe_run do ui-tours delete "$CREATED_UI_TOUR_ID"
assert_failure

test_name "Delete UI tour with --confirm"
# Create a UI tour specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteUiTour")
TEST_DEL_ABBREV=$(echo "$TEST_DEL_NAME" | tr '[:upper:]' '[:lower:]' | cut -c1-32)
xbe_json do ui-tours create --name "$TEST_DEL_NAME" --abbreviation "$TEST_DEL_ABBREV"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do ui-tours delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create UI tour for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create UI tour without name fails"
xbe_json do ui-tours create --abbreviation "no-name"
assert_failure

test_name "Create UI tour without abbreviation fails"
xbe_json do ui-tours create --name "No Abbrev"
assert_failure

test_name "Update without any fields fails"
xbe_json do ui-tours update "$CREATED_UI_TOUR_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
