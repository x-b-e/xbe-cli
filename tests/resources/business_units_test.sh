#!/bin/bash
#
# XBE CLI Integration Tests: Business Units
#
# Tests CRUD operations for the business-units resource.
# Business units require a broker relationship and can have parent business units.
#
# COMPLETE COVERAGE: All 4 create/update attributes + 5 list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BUSINESS_UNIT_ID=""
CREATED_PARENT_BU_ID=""
CREATED_BROKER_ID=""

describe "Resource: business-units"

# ============================================================================
# Prerequisites - Create a broker for business unit tests
# ============================================================================

test_name "Create prerequisite broker for business-unit tests"
BROKER_NAME=$(unique_name "BUTestBroker")

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
    # Try using environment variable
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

test_name "Create business-unit with required fields"
TEST_NAME=$(unique_name "BusinessUnit")

xbe_json do business-units create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business-unit but no ID returned"
    fi
else
    fail "Failed to create business-unit"
fi

# Only continue if we successfully created a business unit
if [[ -z "$CREATED_BUSINESS_UNIT_ID" || "$CREATED_BUSINESS_UNIT_ID" == "null" ]]; then
    echo "Cannot continue without a valid business-unit ID"
    run_tests
fi

test_name "Create business-unit with external-id"
TEST_NAME2=$(unique_name "BusinessUnit2")
TEST_EXTERNAL_ID="EXT-$(date +%s)"
xbe_json do business-units create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --external-id "$TEST_EXTERNAL_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "business-units" "$id"
    pass
else
    fail "Failed to create business-unit with external-id"
fi

test_name "Create parent business-unit for hierarchy test"
PARENT_NAME=$(unique_name "ParentBU")
xbe_json do business-units create \
    --name "$PARENT_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_PARENT_BU_ID=$(json_get ".id")
    if [[ -n "$CREATED_PARENT_BU_ID" && "$CREATED_PARENT_BU_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_PARENT_BU_ID"
        pass
    else
        fail "Created parent business-unit but no ID returned"
    fi
else
    fail "Failed to create parent business-unit"
fi

test_name "Create child business-unit with parent"
if [[ -n "$CREATED_PARENT_BU_ID" && "$CREATED_PARENT_BU_ID" != "null" ]]; then
    CHILD_NAME=$(unique_name "ChildBU")
    xbe_json do business-units create \
        --name "$CHILD_NAME" \
        --broker "$CREATED_BROKER_ID" \
        --parent "$CREATED_PARENT_BU_ID"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "business-units" "$id"
        pass
    else
        fail "Failed to create child business-unit"
    fi
else
    skip "No parent business-unit available for child test"
fi

# ============================================================================
# UPDATE Tests - String Attributes
# ============================================================================

test_name "Update business-unit name"
UPDATED_NAME=$(unique_name "UpdatedBU")
xbe_json do business-units update "$CREATED_BUSINESS_UNIT_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update business-unit external-id"
UPDATED_EXTERNAL_ID="EXT-UPD-$(date +%s)"
xbe_json do business-units update "$CREATED_BUSINESS_UNIT_ID" --external-id "$UPDATED_EXTERNAL_ID"
assert_success

# ============================================================================
# UPDATE Tests - Relationship
# ============================================================================

test_name "Update business-unit parent"
if [[ -n "$CREATED_PARENT_BU_ID" && "$CREATED_PARENT_BU_ID" != "null" ]]; then
    xbe_json do business-units update "$CREATED_BUSINESS_UNIT_ID" --parent "$CREATED_PARENT_BU_ID"
    assert_success
else
    skip "No parent business-unit available for parent update test"
fi

# Note: Business-units resource does not have a "show" command

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List business-units"
xbe_json view business-units list --limit 5
assert_success

test_name "List business-units returns array"
xbe_json view business-units list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list business-units"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List business-units with --name filter"
xbe_json view business-units list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List business-units with --broker filter"
xbe_json view business-units list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List business-units with --parent filter"
if [[ -n "$CREATED_PARENT_BU_ID" && "$CREATED_PARENT_BU_ID" != "null" ]]; then
    xbe_json view business-units list --parent "$CREATED_PARENT_BU_ID" --limit 10
    assert_success
else
    skip "No parent business-unit available for parent filter test"
fi

test_name "List business-units with --with-children filter (true)"
xbe_json view business-units list --with-children true --limit 10
assert_success

test_name "List business-units with --with-children filter (false)"
xbe_json view business-units list --with-children false --limit 10
assert_success

test_name "List business-units with --without-children filter (true)"
xbe_json view business-units list --without-children true --limit 10
assert_success

test_name "List business-units with --without-children filter (false)"
xbe_json view business-units list --without-children false --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List business-units with --limit"
xbe_json view business-units list --limit 3
assert_success

test_name "List business-units with --offset"
xbe_json view business-units list --limit 3 --offset 3
assert_success

test_name "List business-units with pagination (limit + offset)"
xbe_json view business-units list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete business-unit requires --confirm flag"
xbe_json do business-units delete "$CREATED_BUSINESS_UNIT_ID"
assert_failure

test_name "Delete business-unit with --confirm"
# Create a business-unit specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMe")
xbe_json do business-units create \
    --name "$TEST_DEL_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do business-units delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create business-unit for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create business-unit without name fails"
xbe_json do business-units create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create business-unit without broker fails"
xbe_json do business-units create --name "Test BU"
assert_failure

test_name "Update without any fields fails"
xbe_json do business-units update "$CREATED_BUSINESS_UNIT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
