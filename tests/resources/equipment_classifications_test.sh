#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Classifications
#
# Tests CRUD operations for the equipment_classifications resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""

describe "Resource: equipment_classifications"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment classification with required fields"
TEST_NAME=$(unique_name "EquipClass")
TEST_ABBR="EC$(date +%s | tail -c 4)"

xbe_json do equipment-classifications create \
    --name "$TEST_NAME" \
    --abbreviation "$TEST_ABBR"

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_ID"
        pass
    else
        fail "Created but no ID returned"
    fi
else
    fail "Failed to create equipment classification"
fi

# Only continue if we successfully created
if [[ -z "$CREATED_ID" || "$CREATED_ID" == "null" ]]; then
    echo "Cannot continue without a valid ID"
    run_tests
fi

test_name "Create equipment classification with mobilization-method"
TEST_NAME2=$(unique_name "EquipClass2")
TEST_ABBR2="EC2$(date +%s | tail -c 4)"
xbe_json do equipment-classifications create \
    --name "$TEST_NAME2" \
    --abbreviation "$TEST_ABBR2" \
    --mobilization-method "itself"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-classifications" "$id"
    pass
else
    fail "Failed to create with mobilization-method"
fi

test_name "Create equipment classification with parent"
TEST_NAME3=$(unique_name "EquipClass3")
TEST_ABBR3="EC3$(date +%s | tail -c 4)"
xbe_json do equipment-classifications create \
    --name "$TEST_NAME3" \
    --abbreviation "$TEST_ABBR3" \
    --parent "$CREATED_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-classifications" "$id"
    pass
else
    fail "Failed to create with parent"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update equipment classification name"
UPDATED_NAME=$(unique_name "UpdatedEquip")
xbe_json do equipment-classifications update "$CREATED_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update equipment classification abbreviation"
UPDATED_ABBR="UE$(date +%s | tail -c 4)"
xbe_json do equipment-classifications update "$CREATED_ID" --abbreviation "$UPDATED_ABBR"
assert_success

test_name "Update equipment classification mobilization-method"
xbe_json do equipment-classifications update "$CREATED_ID" --mobilization-method "trailer"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List equipment classifications"
xbe_json view equipment-classifications list --limit 5
assert_success

test_name "List equipment classifications returns array"
xbe_json view equipment-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list"
fi

test_name "List equipment classifications with --name filter"
xbe_json view equipment-classifications list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List equipment classifications with --limit"
xbe_json view equipment-classifications list --limit 3
assert_success

test_name "List equipment classifications with --offset"
xbe_json view equipment-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment classification requires --confirm flag"
xbe_run do equipment-classifications delete "$CREATED_ID"
assert_failure

test_name "Delete equipment classification with --confirm"
TEST_DEL_NAME=$(unique_name "DeleteEquip")
TEST_DEL_ABBR="DE$(date +%s | tail -c 4)"
xbe_json do equipment-classifications create \
    --name "$TEST_DEL_NAME" \
    --abbreviation "$TEST_DEL_ABBR"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do equipment-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create equipment classification without name fails"
xbe_json do equipment-classifications create --abbreviation "TEST"
assert_failure

test_name "Update without any fields fails"
xbe_json do equipment-classifications update "$CREATED_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
