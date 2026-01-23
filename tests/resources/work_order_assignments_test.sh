#!/bin/bash
#
# XBE CLI Integration Tests: Work Order Assignments
#
# Tests create/update/delete operations and list filters for the
# work-order-assignments resource.
#
# COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_WORK_ORDER_ID=""
CREATED_WORK_ORDER_ID_2=""
CREATED_USER_ID=""
UPDATED_USER_ID=""
CREATED_ASSIGNMENT_ID=""

describe "Resource: work-order-assignments"

# ============================================================================
# Prerequisites - Create broker, business unit, work orders, users
# ============================================================================

test_name "Create prerequisite broker for work order assignment tests"
BROKER_NAME=$(unique_name "WOAssignBroker")

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

test_name "Create prerequisite business unit for work order assignment tests"
BUSINESS_UNIT_NAME=$(unique_name "WOAssignBU")

xbe_json do business-units create \
    --name "$BUSINESS_UNIT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business unit but no ID returned"
        echo "Cannot continue without a business unit"
        run_tests
    fi
else
    fail "Failed to create business unit"
    echo "Cannot continue without a business unit"
    run_tests
fi

test_name "Create prerequisite work order"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_WORK_ORDER_ID=$(json_get ".id")
    if [[ -n "$CREATED_WORK_ORDER_ID" && "$CREATED_WORK_ORDER_ID" != "null" ]]; then
        register_cleanup "work-orders" "$CREATED_WORK_ORDER_ID"
        pass
    else
        fail "Created work order but no ID returned"
        echo "Cannot continue without a work order"
        run_tests
    fi
else
    fail "Failed to create work order"
    echo "Cannot continue without a work order"
    run_tests
fi

test_name "Create prerequisite user for work order assignment tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Work Order Assignment User" \
    --email "$TEST_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create work order assignment"
xbe_json do work-order-assignments create \
    --work-order "$CREATED_WORK_ORDER_ID" \
    --user "$CREATED_USER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ASSIGNMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_ASSIGNMENT_ID" && "$CREATED_ASSIGNMENT_ID" != "null" ]]; then
        register_cleanup "work-order-assignments" "$CREATED_ASSIGNMENT_ID"
        pass
    else
        fail "Created assignment but no ID returned"
        echo "Cannot continue without a work order assignment"
        run_tests
    fi
else
    fail "Failed to create work order assignment"
    echo "Cannot continue without a work order assignment"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Create second work order for assignment update"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_WORK_ORDER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_WORK_ORDER_ID_2" && "$CREATED_WORK_ORDER_ID_2" != "null" ]]; then
        register_cleanup "work-orders" "$CREATED_WORK_ORDER_ID_2"
        pass
    else
        fail "Created work order but no ID returned"
        echo "Cannot continue without a second work order"
        run_tests
    fi
else
    fail "Failed to create second work order"
    echo "Cannot continue without a second work order"
    run_tests
fi

test_name "Create second user for assignment update"
TEST_EMAIL_2=$(unique_email)

xbe_json do users create \
    --name "Work Order Assignment User 2" \
    --email "$TEST_EMAIL_2"

if [[ $status -eq 0 ]]; then
    UPDATED_USER_ID=$(json_get ".id")
    if [[ -n "$UPDATED_USER_ID" && "$UPDATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a second user"
        run_tests
    fi
else
    fail "Failed to create second user"
    echo "Cannot continue without a second user"
    run_tests
fi

test_name "Update work order assignment relationships"
xbe_json do work-order-assignments update "$CREATED_ASSIGNMENT_ID" \
    --work-order "$CREATED_WORK_ORDER_ID_2" \
    --user "$UPDATED_USER_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show work order assignment"
xbe_json view work-order-assignments show "$CREATED_ASSIGNMENT_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List work order assignments"
xbe_json view work-order-assignments list --limit 10
assert_success

test_name "List work order assignments returns array"
xbe_json view work-order-assignments list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list work order assignments"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List work order assignments with --work-order filter"
xbe_json view work-order-assignments list --work-order "$CREATED_WORK_ORDER_ID_2" --limit 10
assert_success

test_name "List work order assignments with --user filter"
xbe_json view work-order-assignments list --user "$UPDATED_USER_ID" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete work order assignment"
xbe_json do work-order-assignments delete "$CREATED_ASSIGNMENT_ID" --confirm
assert_success

run_tests
