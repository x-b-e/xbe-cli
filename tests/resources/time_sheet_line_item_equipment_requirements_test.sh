#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Line Item Equipment Requirements
#
# Tests CRUD operations for the time_sheet_line_item_equipment_requirements resource.
# These links connect equipment requirements to time sheet line items.
#
# COVERAGE: Create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LINK_ID=""
TIME_SHEET_LINE_ITEM_ID="${XBE_TEST_TIME_SHEET_LINE_ITEM_ID:-}"
EQUIPMENT_REQUIREMENT_ID="${XBE_TEST_EQUIPMENT_REQUIREMENT_ID:-}"

describe "Resource: time-sheet-line-item-equipment-requirements"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet line item equipment requirement without required fields fails"
xbe_json do time-sheet-line-item-equipment-requirements create
assert_failure

if [[ -z "$TIME_SHEET_LINE_ITEM_ID" || -z "$EQUIPMENT_REQUIREMENT_ID" ]]; then
    skip "Set XBE_TEST_TIME_SHEET_LINE_ITEM_ID and XBE_TEST_EQUIPMENT_REQUIREMENT_ID to run create/update/delete tests"
else
    test_name "Create time sheet line item equipment requirement"
    xbe_json do time-sheet-line-item-equipment-requirements create \
        --time-sheet-line-item "$TIME_SHEET_LINE_ITEM_ID" \
        --equipment-requirement "$EQUIPMENT_REQUIREMENT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "time-sheet-line-item-equipment-requirements" "$CREATED_LINK_ID"
            pass
        else
            fail "Created requirement link but no ID returned"
        fi
    else
        fail "Failed to create time sheet line item equipment requirement"
    fi
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Update time sheet line item equipment requirement is-primary"
    xbe_json do time-sheet-line-item-equipment-requirements update "$CREATED_LINK_ID" --is-primary true
    assert_success

    test_name "Update time sheet line item equipment requirement without fields fails"
    xbe_json do time-sheet-line-item-equipment-requirements update "$CREATED_LINK_ID"
    assert_failure
fi

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List time sheet line item equipment requirements"
xbe_json view time-sheet-line-item-equipment-requirements list --limit 1
assert_success

test_name "List time sheet line item equipment requirements returns array"
xbe_json view time-sheet-line-item-equipment-requirements list --limit 1
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time sheet line item equipment requirements"
fi

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Show time sheet line item equipment requirement"
    xbe_json view time-sheet-line-item-equipment-requirements show "$CREATED_LINK_ID"
    assert_success
fi

FILTER_TIME_SHEET_LINE_ITEM_ID="$TIME_SHEET_LINE_ITEM_ID"
FILTER_EQUIPMENT_REQUIREMENT_ID="$EQUIPMENT_REQUIREMENT_ID"

if [[ -z "$FILTER_TIME_SHEET_LINE_ITEM_ID" || -z "$FILTER_EQUIPMENT_REQUIREMENT_ID" ]]; then
    xbe_json view time-sheet-line-item-equipment-requirements list --limit 1
    if [[ $status -eq 0 ]]; then
        FILTER_TIME_SHEET_LINE_ITEM_ID=$(json_get ".[0].time_sheet_line_item_id")
        FILTER_EQUIPMENT_REQUIREMENT_ID=$(json_get ".[0].equipment_requirement_id")
    fi
fi

if [[ -n "$FILTER_TIME_SHEET_LINE_ITEM_ID" && "$FILTER_TIME_SHEET_LINE_ITEM_ID" != "null" ]]; then
    test_name "List time sheet line item equipment requirements with --time-sheet-line-item filter"
    xbe_json view time-sheet-line-item-equipment-requirements list --time-sheet-line-item "$FILTER_TIME_SHEET_LINE_ITEM_ID"
    assert_success
fi

if [[ -n "$FILTER_EQUIPMENT_REQUIREMENT_ID" && "$FILTER_EQUIPMENT_REQUIREMENT_ID" != "null" ]]; then
    test_name "List time sheet line item equipment requirements with --equipment-requirement filter"
    xbe_json view time-sheet-line-item-equipment-requirements list --equipment-requirement "$FILTER_EQUIPMENT_REQUIREMENT_ID"
    assert_success
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Delete time sheet line item equipment requirement requires --confirm flag"
    xbe_run do time-sheet-line-item-equipment-requirements delete "$CREATED_LINK_ID"
    assert_failure

    test_name "Delete time sheet line item equipment requirement with --confirm"
    xbe_json do time-sheet-line-item-equipment-requirements delete "$CREATED_LINK_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
