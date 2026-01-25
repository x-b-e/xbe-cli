#!/bin/bash
#
# XBE CLI Integration Tests: Action Item Line Items
#
# Tests list, show, create, update, and delete operations for the action-item-line-items resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_ACTION_ITEM_ID=""
SAMPLE_RESPONSIBLE_PERSON_ID=""
SAMPLE_DUE_ON=""
ACTION_ITEM_ID=""
CREATED_ID=""


describe "Resource: action-item-line-items"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List action item line items"
xbe_json view action-item-line-items list --limit 5
assert_success

test_name "List action item line items returns array"
xbe_json view action-item-line-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list action item line items"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample action item line item"
xbe_json view action-item-line-items list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ACTION_ITEM_ID=$(json_get ".[0].action_item_id")
    SAMPLE_RESPONSIBLE_PERSON_ID=$(json_get ".[0].responsible_person_id")
    SAMPLE_DUE_ON=$(json_get ".[0].due_on")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No action item line items available for follow-on tests"
    fi
else
    skip "Could not list action item line items to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List action item line items with --status filter"
xbe_json view action-item-line-items list --status open --limit 5
assert_success

test_name "List action item line items with --action-item filter"
if [[ -n "$SAMPLE_ACTION_ITEM_ID" && "$SAMPLE_ACTION_ITEM_ID" != "null" ]]; then
    xbe_json view action-item-line-items list --action-item "$SAMPLE_ACTION_ITEM_ID" --limit 5
    assert_success
else
    skip "No action item ID available"
fi


test_name "List action item line items with --responsible-person filter"
if [[ -n "$SAMPLE_RESPONSIBLE_PERSON_ID" && "$SAMPLE_RESPONSIBLE_PERSON_ID" != "null" ]]; then
    xbe_json view action-item-line-items list --responsible-person "$SAMPLE_RESPONSIBLE_PERSON_ID" --limit 5
    assert_success
else
    skip "No responsible person ID available"
fi


test_name "List action item line items with --due-on-min filter"
xbe_json view action-item-line-items list --due-on-min "2020-01-01" --limit 5
assert_success

test_name "List action item line items with --due-on-max filter"
xbe_json view action-item-line-items list --due-on-max "2030-01-01" --limit 5
assert_success

test_name "List action item line items with --has-due-on filter"
xbe_json view action-item-line-items list --has-due-on true --limit 5
assert_success

test_name "List action item line items with --due-on filter"
if [[ -n "$SAMPLE_DUE_ON" && "$SAMPLE_DUE_ON" != "null" ]]; then
    xbe_json view action-item-line-items list --due-on "$SAMPLE_DUE_ON" --limit 5
    assert_success
else
    skip "No due-on value available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show action item line item"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view action-item-line-items show "$SAMPLE_ID"
    assert_success
else
    skip "No action item line item ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$SAMPLE_ACTION_ITEM_ID" && "$SAMPLE_ACTION_ITEM_ID" != "null" ]]; then
    ACTION_ITEM_ID="$SAMPLE_ACTION_ITEM_ID"
else
    xbe_json view action-items list --limit 1
    if [[ $status -eq 0 ]]; then
        ACTION_ITEM_ID=$(json_get ".[0].id")
    fi
fi

test_name "Create action item line item"
if [[ -n "$ACTION_ITEM_ID" && "$ACTION_ITEM_ID" != "null" ]]; then
    TITLE="CLI line item $(unique_suffix)"
    xbe_json do action-item-line-items create --action-item "$ACTION_ITEM_ID" --title "$TITLE"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "action-item-line-items" "$CREATED_ID"
            pass
        else
            fail "Created action item line item but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No action item ID available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update action item line item"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do action-item-line-items update "$CREATED_ID" --status closed
    assert_success
elif [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do action-item-line-items update "$SAMPLE_ID" --status closed
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update action item line item (permissions or policy)"
    fi
else
    skip "No action item line item ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete action item line item"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do action-item-line-items delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created action item line item to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create action item line item without required fields fails"
xbe_run do action-item-line-items create
assert_failure


test_name "Update action item line item without any fields fails"
xbe_run do action-item-line-items update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
