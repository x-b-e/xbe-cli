#!/bin/bash
#
# XBE CLI Integration Tests: Labor Classifications
#
# Tests CRUD operations for the labor_classifications resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""

describe "Resource: labor_classifications"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create labor classification with required fields"
TEST_NAME=$(unique_name "LaborClass")
TEST_ABBR="LC$(date +%s | tail -c 4)"

xbe_json do labor-classifications create \
    --name "$TEST_NAME" \
    --abbreviation "$TEST_ABBR"

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_ID"
        pass
    else
        fail "Created but no ID returned"
    fi
else
    fail "Failed to create labor classification"
fi

# Only continue if we successfully created
if [[ -z "$CREATED_ID" || "$CREATED_ID" == "null" ]]; then
    echo "Cannot continue without a valid ID"
    run_tests
fi

test_name "Create labor classification with is-manager"
TEST_NAME2=$(unique_name "LaborClass2")
TEST_ABBR2="LC2$(date +%s | tail -c 4)"
xbe_json do labor-classifications create \
    --name "$TEST_NAME2" \
    --abbreviation "$TEST_ABBR2" \
    --is-manager
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "labor-classifications" "$id"
    pass
else
    fail "Failed to create with is-manager"
fi

test_name "Create labor classification with is-time-card-approver"
TEST_NAME3=$(unique_name "LaborClass3")
TEST_ABBR3="LC3$(date +%s | tail -c 4)"
xbe_json do labor-classifications create \
    --name "$TEST_NAME3" \
    --abbreviation "$TEST_ABBR3" \
    --is-time-card-approver
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "labor-classifications" "$id"
    pass
else
    fail "Failed to create with is-time-card-approver"
fi

test_name "Create labor classification with can-manage-projects"
TEST_NAME4=$(unique_name "LaborClass4")
TEST_ABBR4="LC4$(date +%s | tail -c 4)"
xbe_json do labor-classifications create \
    --name "$TEST_NAME4" \
    --abbreviation "$TEST_ABBR4" \
    --can-manage-projects
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "labor-classifications" "$id"
    pass
else
    fail "Failed to create with can-manage-projects"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update labor classification name"
UPDATED_NAME=$(unique_name "UpdatedLabor")
xbe_json do labor-classifications update "$CREATED_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update labor classification abbreviation"
UPDATED_ABBR="UL$(date +%s | tail -c 4)"
xbe_json do labor-classifications update "$CREATED_ID" --abbreviation "$UPDATED_ABBR"
assert_success

test_name "Update labor classification is-manager to true"
xbe_json do labor-classifications update "$CREATED_ID" --is-manager
assert_success

test_name "Update labor classification is-time-card-approver to true"
xbe_json do labor-classifications update "$CREATED_ID" --is-time-card-approver
assert_success

test_name "Update labor classification can-manage-projects to true"
xbe_json do labor-classifications update "$CREATED_ID" --can-manage-projects
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List labor classifications"
xbe_json view labor-classifications list --limit 5
assert_success

test_name "List labor classifications returns array"
xbe_json view labor-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list"
fi

test_name "List labor classifications with --name filter"
xbe_json view labor-classifications list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List labor classifications with --limit"
xbe_json view labor-classifications list --limit 3
assert_success

test_name "List labor classifications with --offset"
xbe_json view labor-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete labor classification requires --confirm flag"
xbe_run do labor-classifications delete "$CREATED_ID"
assert_failure

test_name "Delete labor classification with --confirm"
TEST_DEL_NAME=$(unique_name "DeleteLabor")
TEST_DEL_ABBR="DL$(date +%s | tail -c 4)"
xbe_json do labor-classifications create \
    --name "$TEST_DEL_NAME" \
    --abbreviation "$TEST_DEL_ABBR"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do labor-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create labor classification without name fails"
xbe_json do labor-classifications create --abbreviation "TEST"
assert_failure

test_name "Update without any fields fails"
xbe_json do labor-classifications update "$CREATED_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
