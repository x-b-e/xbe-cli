#!/bin/bash
#
# XBE CLI Integration Tests: Crafts and Craft Classes
#
# Tests CRUD operations for the crafts and craft_classes resources.
# Crafts define trade classifications for workers.
# Craft classes are subdivisions within a craft.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CRAFT_ID=""
CREATED_CRAFT_CLASS_ID=""
CREATED_BROKER_ID=""

describe "Resource: crafts and craft_classes"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for craft tests"
BROKER_NAME=$(unique_name "CraftTestBroker")

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
# CRAFT CREATE Tests
# ============================================================================

test_name "Create craft with required fields"
TEST_NAME=$(unique_name "Craft")

xbe_json do crafts create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_ID" && "$CREATED_CRAFT_ID" != "null" ]]; then
        register_cleanup "crafts" "$CREATED_CRAFT_ID"
        pass
    else
        fail "Created craft but no ID returned"
    fi
else
    fail "Failed to create craft"
fi

# Only continue if we successfully created a craft
if [[ -z "$CREATED_CRAFT_ID" || "$CREATED_CRAFT_ID" == "null" ]]; then
    echo "Cannot continue without a valid craft ID"
    run_tests
fi

test_name "Create craft with code"
TEST_NAME2=$(unique_name "Craft2")
xbe_json do crafts create \
    --name "$TEST_NAME2" \
    --code "CRF2" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "crafts" "$id"
    pass
else
    fail "Failed to create craft with code"
fi

# ============================================================================
# CRAFT UPDATE Tests
# ============================================================================

test_name "Update craft name"
UPDATED_NAME=$(unique_name "UpdatedCraft")
xbe_json do crafts update "$CREATED_CRAFT_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update craft code"
xbe_json do crafts update "$CREATED_CRAFT_ID" --code "UPDC"
assert_success

# ============================================================================
# CRAFT LIST Tests - Basic
# ============================================================================

test_name "List crafts"
xbe_json view crafts list --limit 5
assert_success

test_name "List crafts returns array"
xbe_json view crafts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list crafts"
fi

# ============================================================================
# CRAFT LIST Tests - Filters
# ============================================================================

test_name "List crafts with --broker filter"
xbe_json view crafts list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# CRAFT LIST Tests - Pagination
# ============================================================================

test_name "List crafts with --limit"
xbe_json view crafts list --limit 3
assert_success

test_name "List crafts with --offset"
xbe_json view crafts list --limit 3 --offset 3
assert_success

# ============================================================================
# CRAFT CLASS CREATE Tests
# ============================================================================

test_name "Create craft class with required fields"
CC_NAME=$(unique_name "CraftClass")

xbe_json do craft-classes create \
    --name "$CC_NAME" \
    --craft "$CREATED_CRAFT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
        register_cleanup "craft-classes" "$CREATED_CRAFT_CLASS_ID"
        pass
    else
        fail "Created craft class but no ID returned"
    fi
else
    fail "Failed to create craft class"
fi

test_name "Create craft class with code"
CC_NAME2=$(unique_name "CraftClass2")
xbe_json do craft-classes create \
    --name "$CC_NAME2" \
    --code "CC2" \
    --craft "$CREATED_CRAFT_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "craft-classes" "$id"
    pass
else
    fail "Failed to create craft class with code"
fi

test_name "Create craft class with is-valid-for-drivers"
CC_NAME3=$(unique_name "CraftClass3")
xbe_json do craft-classes create \
    --name "$CC_NAME3" \
    --craft "$CREATED_CRAFT_ID" \
    --is-valid-for-drivers
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "craft-classes" "$id"
    pass
else
    fail "Failed to create craft class with is-valid-for-drivers"
fi

# ============================================================================
# CRAFT CLASS UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
    test_name "Update craft class name"
    UPDATED_CC_NAME=$(unique_name "UpdatedCC")
    xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID" --name "$UPDATED_CC_NAME"
    assert_success

    test_name "Update craft class code"
    xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID" --code "UPDCC"
    assert_success

    test_name "Update craft class is-valid-for-drivers"
    xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID" --is-valid-for-drivers=true
    assert_success
fi

# ============================================================================
# CRAFT CLASS LIST Tests
# ============================================================================

test_name "List craft classes"
xbe_json view craft-classes list --limit 5
assert_success

test_name "List craft classes returns array"
xbe_json view craft-classes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list craft classes"
fi

test_name "List craft classes with --craft filter"
xbe_json view craft-classes list --craft "$CREATED_CRAFT_ID" --limit 10
assert_success

test_name "List craft classes with --broker filter"
xbe_json view craft-classes list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List craft classes with --limit"
xbe_json view craft-classes list --limit 3
assert_success

test_name "List craft classes with --offset"
xbe_json view craft-classes list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete craft class requires --confirm flag"
if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
    xbe_run do craft-classes delete "$CREATED_CRAFT_CLASS_ID"
    assert_failure
else
    skip "No craft class ID for delete test"
fi

test_name "Delete craft class with --confirm"
# Create a craft class specifically for deletion
DEL_CC_NAME=$(unique_name "DeleteCC")
xbe_json do craft-classes create \
    --name "$DEL_CC_NAME" \
    --craft "$CREATED_CRAFT_ID"
if [[ $status -eq 0 ]]; then
    DEL_CC_ID=$(json_get ".id")
    xbe_run do craft-classes delete "$DEL_CC_ID" --confirm
    assert_success
else
    skip "Could not create craft class for deletion test"
fi

test_name "Delete craft requires --confirm flag"
xbe_run do crafts delete "$CREATED_CRAFT_ID"
assert_failure

test_name "Delete craft with --confirm"
# Create a craft specifically for deletion
DEL_CRAFT_NAME=$(unique_name "DeleteCraft")
xbe_json do crafts create \
    --name "$DEL_CRAFT_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do crafts delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create craft for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create craft without name fails"
xbe_json do crafts create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create craft without broker fails"
xbe_json do crafts create --name "NoBroker"
assert_failure

test_name "Create craft class without name fails"
xbe_json do craft-classes create --craft "$CREATED_CRAFT_ID"
assert_failure

test_name "Create craft class without craft fails"
xbe_json do craft-classes create --name "NoCraft"
assert_failure

test_name "Update craft without any fields fails"
xbe_json do crafts update "$CREATED_CRAFT_ID"
assert_failure

test_name "Update craft class without any fields fails"
if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
    xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID"
    assert_failure
else
    skip "No craft class ID for update error test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
