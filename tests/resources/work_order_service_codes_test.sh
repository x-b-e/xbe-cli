#!/bin/bash
#
# XBE CLI Integration Tests: Work Order Service Codes
#
# Tests CRUD operations for the work_order_service_codes resource.
# Work order service codes describe service categories used on work orders.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SERVICE_CODE_ID=""
CREATED_BROKER_ID=""

describe "Resource: work_order_service_codes"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for work order service code tests"
BROKER_NAME=$(unique_name "WOSCTestBroker")

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

test_name "Create work order service code with required fields"
TEST_CODE="WOSC-$(date +%s)-$RANDOM"

xbe_json do work-order-service-codes create \
    --code "$TEST_CODE" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_SERVICE_CODE_ID=$(json_get ".id")
    if [[ -n "$CREATED_SERVICE_CODE_ID" && "$CREATED_SERVICE_CODE_ID" != "null" ]]; then
        register_cleanup "work-order-service-codes" "$CREATED_SERVICE_CODE_ID"
        pass
    else
        fail "Created work order service code but no ID returned"
    fi
else
    fail "Failed to create work order service code"
fi

# Only continue if we successfully created a work order service code
if [[ -z "$CREATED_SERVICE_CODE_ID" || "$CREATED_SERVICE_CODE_ID" == "null" ]]; then
    echo "Cannot continue without a valid work order service code ID"
    run_tests
fi

test_name "Create work order service code with description"
TEST_CODE2="WOSC2-$(date +%s)-$RANDOM"
xbe_json do work-order-service-codes create \
    --code "$TEST_CODE2" \
    --broker "$CREATED_BROKER_ID" \
    --description "Service code with description"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "work-order-service-codes" "$id"
    pass
else
    fail "Failed to create work order service code with description"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update work order service code value"
UPDATED_CODE="WOSC-UPDATED-$(date +%s)-$RANDOM"
xbe_json do work-order-service-codes update "$CREATED_SERVICE_CODE_ID" --code "$UPDATED_CODE"
assert_success

test_name "Update work order service code description"
xbe_json do work-order-service-codes update "$CREATED_SERVICE_CODE_ID" --description "Updated description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List work order service codes"
xbe_json view work-order-service-codes list --limit 5
assert_success

test_name "List work order service codes returns array"
xbe_json view work-order-service-codes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list work order service codes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List work order service codes with --broker filter"
xbe_json view work-order-service-codes list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List work order service codes with --limit"
xbe_json view work-order-service-codes list --limit 3
assert_success

test_name "List work order service codes with --offset"
xbe_json view work-order-service-codes list --limit 3 --offset 3
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show work order service code"
xbe_json view work-order-service-codes show "$CREATED_SERVICE_CODE_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete work order service code requires --confirm flag"
xbe_run do work-order-service-codes delete "$CREATED_SERVICE_CODE_ID"
assert_failure

test_name "Delete work order service code with --confirm"
TEST_DEL_CODE="WOSC-DEL-$(date +%s)-$RANDOM"
xbe_json do work-order-service-codes create \
    --code "$TEST_DEL_CODE" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do work-order-service-codes delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create work order service code for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create work order service code without code fails"
xbe_json do work-order-service-codes create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create work order service code without broker fails"
xbe_json do work-order-service-codes create --code "NO-BROKER"
assert_failure

test_name "Update without any fields fails"
xbe_json do work-order-service-codes update "$CREATED_SERVICE_CODE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
