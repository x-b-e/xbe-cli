#!/bin/bash
#
# XBE CLI Integration Tests: Craft Classes
#
# Tests CRUD operations for the craft_classes resource.
# Craft classes are sub-classifications within a craft.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CRAFT_CLASS_ID=""
CREATED_BROKER_ID=""
CREATED_CRAFT_ID=""

describe "Resource: craft_classes"

# ============================================================================
# Prerequisites - Create broker and craft
# ============================================================================

test_name "Create prerequisite broker for craft classes tests"
BROKER_NAME=$(unique_name "CCTestBroker")

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

test_name "Create prerequisite craft"
CRAFT_NAME=$(unique_name "CCTestCraft")

xbe_json do crafts create \
    --name "$CRAFT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_ID" && "$CREATED_CRAFT_ID" != "null" ]]; then
        register_cleanup "crafts" "$CREATED_CRAFT_ID"
        pass
    else
        fail "Created craft but no ID returned"
        echo "Cannot continue without a craft"
        run_tests
    fi
else
    fail "Failed to create craft"
    echo "Cannot continue without a craft"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create craft class with required fields"
CLASS_NAME=$(unique_name "CraftClass")

xbe_json do craft-classes create \
    --name "$CLASS_NAME" \
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

# Only continue if we successfully created a craft class
if [[ -z "$CREATED_CRAFT_CLASS_ID" || "$CREATED_CRAFT_CLASS_ID" == "null" ]]; then
    echo "Cannot continue without a valid craft class ID"
    run_tests
fi

test_name "Create craft class with code"
CLASS_NAME2=$(unique_name "CraftClass2")
xbe_json do craft-classes create \
    --name "$CLASS_NAME2" \
    --craft "$CREATED_CRAFT_ID" \
    --code "CC2"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "craft-classes" "$id"
    pass
else
    fail "Failed to create craft class with code"
fi

test_name "Create craft class with is-valid-for-drivers"
CLASS_NAME3=$(unique_name "CraftClass3")
xbe_json do craft-classes create \
    --name "$CLASS_NAME3" \
    --craft "$CREATED_CRAFT_ID" \
    --is-valid-for-drivers
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "craft-classes" "$id"
    pass
else
    fail "Failed to create craft class with is-valid-for-drivers"
fi

test_name "Create craft class with all attributes"
CLASS_NAME4=$(unique_name "CraftClass4")
xbe_json do craft-classes create \
    --name "$CLASS_NAME4" \
    --craft "$CREATED_CRAFT_ID" \
    --code "CC4" \
    --is-valid-for-drivers
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "craft-classes" "$id"
    pass
else
    fail "Failed to create craft class with all attributes"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update craft class name"
UPDATED_NAME=$(unique_name "UpdatedCraftClass")
xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update craft class code"
xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID" --code "UPD"
assert_success

test_name "Update craft class is-valid-for-drivers"
xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID" --is-valid-for-drivers
assert_success

# ============================================================================
# LIST Tests - Basic
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

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List craft classes with --craft filter"
xbe_json view craft-classes list --craft "$CREATED_CRAFT_ID" --limit 10
assert_success

test_name "List craft classes with --broker filter"
xbe_json view craft-classes list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List craft classes with --is-valid-for-drivers filter"
xbe_json view craft-classes list --is-valid-for-drivers true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

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
xbe_run do craft-classes delete "$CREATED_CRAFT_CLASS_ID"
assert_failure

test_name "Delete craft class with --confirm"
# Create a craft class specifically for deletion
DEL_CLASS_NAME=$(unique_name "DeleteCraftClass")
xbe_json do craft-classes create \
    --name "$DEL_CLASS_NAME" \
    --craft "$CREATED_CRAFT_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do craft-classes delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create craft class for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create craft class without name fails"
xbe_json do craft-classes create --craft "$CREATED_CRAFT_ID"
assert_failure

test_name "Create craft class without craft fails"
xbe_json do craft-classes create --name "NoParentCraft"
assert_failure

test_name "Update without any fields fails"
xbe_json do craft-classes update "$CREATED_CRAFT_CLASS_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
