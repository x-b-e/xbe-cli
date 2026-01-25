#!/bin/bash
#
# XBE CLI Integration Tests: Custom Work Order Statuses
#
# Tests CRUD operations for the custom_work_order_statuses resource.
# Custom work order statuses allow organizations to define their own
# workflow states that map to primary statuses.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_STATUS_ID=""
CREATED_BROKER_ID=""

describe "Resource: custom_work_order_statuses"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for custom work order status tests"
BROKER_NAME=$(unique_name "CWOSTestBroker")

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

test_name "Create custom work order status with required fields"
TEST_LABEL=$(unique_name "CWOStatus")

xbe_json do custom-work-order-statuses create \
    --label "$TEST_LABEL" \
    --primary-status "editing" \
    --broker "$CREATED_BROKER_ID" \
    --color-hex "#FF0000"

if [[ $status -eq 0 ]]; then
    CREATED_STATUS_ID=$(json_get ".id")
    if [[ -n "$CREATED_STATUS_ID" && "$CREATED_STATUS_ID" != "null" ]]; then
        register_cleanup "custom-work-order-statuses" "$CREATED_STATUS_ID"
        pass
    else
        fail "Created custom work order status but no ID returned"
    fi
else
    fail "Failed to create custom work order status"
fi

# Only continue if we successfully created a custom work order status
if [[ -z "$CREATED_STATUS_ID" || "$CREATED_STATUS_ID" == "null" ]]; then
    echo "Cannot continue without a valid custom work order status ID"
    run_tests
fi

test_name "Create custom work order status with description"
TEST_LABEL2=$(unique_name "CWOStatus2")
xbe_json do custom-work-order-statuses create \
    --label "$TEST_LABEL2" \
    --primary-status "in_progress" \
    --broker "$CREATED_BROKER_ID" \
    --color-hex "#00FF00" \
    --description "A custom status with description"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "custom-work-order-statuses" "$id"
    pass
else
    fail "Failed to create custom work order status with description"
fi

test_name "Create custom work order status with ready_for_work status"
TEST_LABEL3=$(unique_name "CWOStatus3")
xbe_json do custom-work-order-statuses create \
    --label "$TEST_LABEL3" \
    --primary-status "ready_for_work" \
    --broker "$CREATED_BROKER_ID" \
    --color-hex "#FF5500"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "custom-work-order-statuses" "$id"
    pass
else
    fail "Failed to create custom work order status with color-hex"
fi

test_name "Create custom work order status with all fields"
TEST_LABEL4=$(unique_name "CWOStatus4")
xbe_json do custom-work-order-statuses create \
    --label "$TEST_LABEL4" \
    --primary-status "on_hold" \
    --broker "$CREATED_BROKER_ID" \
    --description "Full custom status" \
    --color-hex "#3366CC"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "custom-work-order-statuses" "$id"
    pass
else
    fail "Failed to create custom work order status with all optional fields"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update custom work order status label"
UPDATED_LABEL=$(unique_name "UpdatedCWOS")
xbe_json do custom-work-order-statuses update "$CREATED_STATUS_ID" --label "$UPDATED_LABEL"
assert_success

test_name "Update custom work order status description"
xbe_json do custom-work-order-statuses update "$CREATED_STATUS_ID" --description "Updated description"
assert_success

test_name "Update custom work order status color-hex"
xbe_json do custom-work-order-statuses update "$CREATED_STATUS_ID" --color-hex "#00FF00"
assert_success

test_name "Update custom work order status primary-status"
xbe_json do custom-work-order-statuses update "$CREATED_STATUS_ID" --primary-status "in_progress"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List custom work order statuses"
xbe_json view custom-work-order-statuses list --limit 5
assert_success

test_name "List custom work order statuses returns array"
xbe_json view custom-work-order-statuses list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list custom work order statuses"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List custom work order statuses with --broker filter"
xbe_json view custom-work-order-statuses list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List custom work order statuses with --limit"
xbe_json view custom-work-order-statuses list --limit 3
assert_success

test_name "List custom work order statuses with --offset"
xbe_json view custom-work-order-statuses list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete custom work order status requires --confirm flag"
xbe_run do custom-work-order-statuses delete "$CREATED_STATUS_ID"
assert_failure

test_name "Delete custom work order status with --confirm"
# Create a custom work order status specifically for deletion
TEST_DEL_LABEL=$(unique_name "DeleteCWOS")
xbe_json do custom-work-order-statuses create \
    --label "$TEST_DEL_LABEL" \
    --primary-status "completed" \
    --broker "$CREATED_BROKER_ID" \
    --color-hex "#AABBCC"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do custom-work-order-statuses delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create custom work order status for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create custom work order status without label fails"
xbe_json do custom-work-order-statuses create --primary-status "editing" --broker "$CREATED_BROKER_ID" --color-hex "#000000"
assert_failure

test_name "Create custom work order status without primary-status fails"
xbe_json do custom-work-order-statuses create --label "NoStatus" --broker "$CREATED_BROKER_ID" --color-hex "#000000"
assert_failure

test_name "Create custom work order status without broker fails"
xbe_json do custom-work-order-statuses create --label "NoBroker" --primary-status "editing" --color-hex "#000000"
assert_failure

test_name "Create custom work order status without color-hex fails"
xbe_json do custom-work-order-statuses create --label "NoColor" --primary-status "editing" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do custom-work-order-statuses update "$CREATED_STATUS_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
