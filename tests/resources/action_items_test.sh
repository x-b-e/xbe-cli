#!/bin/bash
#
# XBE CLI Integration Tests: Action Items
#
# Tests CRUD operations for the action_items resource.
# Action items are trackable work such as tasks, bugs, features, and integrations.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ACTION_ITEM_ID=""
CREATED_BROKER_ID=""

describe "Resource: action_items"

# ============================================================================
# Prerequisites - Create broker for responsible-organization tests
# ============================================================================

test_name "Create prerequisite broker for action item tests"
BROKER_NAME=$(unique_name "AITestBroker")

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
# CREATE Tests
# ============================================================================

test_name "Create action item with required fields"
TEST_TITLE=$(unique_name "ActionItem")

xbe_json do action-items create \
    --title "$TEST_TITLE" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ACTION_ITEM_ID=$(json_get ".id")
    if [[ -n "$CREATED_ACTION_ITEM_ID" && "$CREATED_ACTION_ITEM_ID" != "null" ]]; then
        register_cleanup "action-items" "$CREATED_ACTION_ITEM_ID"
        pass
    else
        fail "Created action item but no ID returned"
    fi
else
    fail "Failed to create action item"
fi

# Only continue if we successfully created an action item
if [[ -z "$CREATED_ACTION_ITEM_ID" || "$CREATED_ACTION_ITEM_ID" == "null" ]]; then
    echo "Cannot continue without a valid action item ID"
    run_tests
fi

test_name "Create action item with description"
TEST_TITLE2=$(unique_name "ActionItem2")
xbe_json do action-items create \
    --title "$TEST_TITLE2" \
    --description "Test action item with description" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with description"
fi

test_name "Create action item with due-on"
TEST_TITLE3=$(unique_name "ActionItem3")
xbe_json do action-items create \
    --title "$TEST_TITLE3" \
    --due-on "2025-12-31" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with due-on"
fi

test_name "Create action item with status"
TEST_TITLE4=$(unique_name "ActionItem4")
xbe_json do action-items create \
    --title "$TEST_TITLE4" \
    --status "ready_for_work" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with status"
fi

test_name "Create action item with kind"
TEST_TITLE5=$(unique_name "ActionItem5")
xbe_json do action-items create \
    --title "$TEST_TITLE5" \
    --kind "feature" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with kind"
fi

test_name "Create action item with expected-cost-amount"
TEST_TITLE6=$(unique_name "ActionItem6")
xbe_json do action-items create \
    --title "$TEST_TITLE6" \
    --expected-cost-amount "1000.00" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with expected-cost-amount"
fi

test_name "Create action item with expected-benefit-amount"
TEST_TITLE7=$(unique_name "ActionItem7")
xbe_json do action-items create \
    --title "$TEST_TITLE7" \
    --expected-benefit-amount "5000.00" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with expected-benefit-amount"
fi

test_name "Create action item with requires-xbe-feature"
TEST_TITLE8=$(unique_name "ActionItem8")
xbe_json do action-items create \
    --title "$TEST_TITLE8" \
    --requires-xbe-feature "true" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with requires-xbe-feature"
fi

test_name "Create action item with all optional fields"
TEST_TITLE10=$(unique_name "ActionItem10")
xbe_json do action-items create \
    --title "$TEST_TITLE10" \
    --description "Full test action item" \
    --due-on "2025-06-30" \
    --status "in_progress" \
    --kind "bug_fix" \
    --expected-cost-amount "500.00" \
    --expected-benefit-amount "2000.00" \
    --requires-xbe-feature "false" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "action-items" "$id"
    pass
else
    fail "Failed to create action item with all optional fields"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update action item title"
UPDATED_TITLE=$(unique_name "UpdatedAI")
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --title "$UPDATED_TITLE"
assert_success

test_name "Update action item description"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --description "Updated description"
assert_success

test_name "Update action item due-on"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --due-on "2025-11-30"
assert_success

test_name "Update action item status to in_progress"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --status "in_progress"
assert_success

test_name "Update action item status to in_verification"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --status "in_verification"
assert_success

test_name "Update action item kind"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --kind "integration"
assert_success

test_name "Update action item expected-cost-amount"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --expected-cost-amount "1500.00"
assert_success

test_name "Update action item expected-benefit-amount"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --expected-benefit-amount "7500.00"
assert_success

test_name "Update action item requires-xbe-feature to true"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --requires-xbe-feature "true"
assert_success

test_name "Update action item requires-xbe-feature to false"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --requires-xbe-feature "false"
assert_success

test_name "Update action item completed-on"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID" --completed-on "2025-01-15"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List action items"
xbe_json view action-items list --limit 5
assert_success

test_name "List action items returns array"
xbe_json view action-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list action items"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List action items with --status filter"
xbe_json view action-items list --status "in_progress" --limit 10
assert_success

test_name "List action items with --status multiple values"
xbe_json view action-items list --status "in_progress,ready_for_work" --limit 10
assert_success

test_name "List action items with --kind filter"
xbe_json view action-items list --kind "feature" --limit 10
assert_success

test_name "List action items with --kind multiple values"
xbe_json view action-items list --kind "feature,bug_fix" --limit 10
assert_success

test_name "List action items with --q filter"
xbe_json view action-items list --q "ActionItem" --limit 10
assert_success

test_name "List action items with --broker filter"
xbe_json view action-items list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List action items with --is-completed filter"
xbe_json view action-items list --is-completed "false" --limit 10
assert_success

test_name "List action items with --requires-xbe-feature filter"
xbe_json view action-items list --requires-xbe-feature "true" --limit 10
assert_success

test_name "List action items with --due-on-min filter"
xbe_json view action-items list --due-on-min "2025-01-01" --limit 10
assert_success

test_name "List action items with --due-on-max filter"
xbe_json view action-items list --due-on-max "2025-12-31" --limit 10
assert_success

test_name "List action items with combined status and kind"
xbe_json view action-items list --status "in_progress" --kind "feature" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List action items with --limit"
xbe_json view action-items list --limit 3
assert_success

test_name "List action items with --offset"
xbe_json view action-items list --limit 3 --offset 3
assert_success

test_name "List action items with --sort"
xbe_json view action-items list --sort "-created-at" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete action item requires --confirm flag"
xbe_run do action-items delete "$CREATED_ACTION_ITEM_ID"
assert_failure

test_name "Delete action item with --confirm"
# Create an action item specifically for deletion
TEST_DEL_TITLE=$(unique_name "DeleteAI")
xbe_json do action-items create \
    --title "$TEST_DEL_TITLE" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do action-items delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create action item for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create action item without title fails"
xbe_json do action-items create --description "No title" --responsible-organization "Broker|$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do action-items update "$CREATED_ACTION_ITEM_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
