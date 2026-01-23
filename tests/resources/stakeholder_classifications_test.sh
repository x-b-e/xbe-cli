#!/bin/bash
#
# XBE CLI Integration Tests: Stakeholder Classifications
#
# Tests CRUD operations for the stakeholder_classifications resource.
# Stakeholder classifications define types of stakeholders.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CLASSIFICATION_ID=""

describe "Resource: stakeholder_classifications"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create stakeholder classification with required fields"
TEST_TITLE=$(unique_name "StakeholderClass")

xbe_json do stakeholder-classifications create \
    --title "$TEST_TITLE"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "stakeholder-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created stakeholder classification but no ID returned"
    fi
else
    fail "Failed to create stakeholder classification"
fi

# Only continue if we successfully created a classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid stakeholder classification ID"
    run_tests
fi

test_name "Create stakeholder classification with leverage-factor"
TEST_TITLE2=$(unique_name "StakeholderClass2")
xbe_json do stakeholder-classifications create \
    --title "$TEST_TITLE2" \
    --leverage-factor 5
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "stakeholder-classifications" "$id"
    pass
else
    fail "Failed to create stakeholder classification with leverage-factor"
fi

test_name "Create stakeholder classification with objectives-narrative-explicit"
TEST_TITLE3=$(unique_name "StakeholderClass3")
xbe_json do stakeholder-classifications create \
    --title "$TEST_TITLE3" \
    --objectives-narrative-explicit "Key decision maker for project funding"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "stakeholder-classifications" "$id"
    pass
else
    fail "Failed to create stakeholder classification with objectives-narrative-explicit"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update stakeholder classification title"
UPDATED_TITLE=$(unique_name "UpdatedSC")
xbe_json do stakeholder-classifications update "$CREATED_CLASSIFICATION_ID" --title "$UPDATED_TITLE"
assert_success

test_name "Update stakeholder classification leverage-factor"
xbe_json do stakeholder-classifications update "$CREATED_CLASSIFICATION_ID" --leverage-factor 8
assert_success

test_name "Update stakeholder classification objectives-narrative-explicit"
xbe_json do stakeholder-classifications update "$CREATED_CLASSIFICATION_ID" --objectives-narrative-explicit "Updated objectives narrative"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List stakeholder classifications"
xbe_json view stakeholder-classifications list --limit 5
assert_success

test_name "List stakeholder classifications returns array"
xbe_json view stakeholder-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list stakeholder classifications"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List stakeholder classifications with --limit"
xbe_json view stakeholder-classifications list --limit 3
assert_success

test_name "List stakeholder classifications with --offset"
xbe_json view stakeholder-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete stakeholder classification requires --confirm flag"
xbe_run do stakeholder-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Delete stakeholder classification with --confirm"
# Create a classification specifically for deletion
TEST_DEL_TITLE=$(unique_name "DeleteSC")
xbe_json do stakeholder-classifications create \
    --title "$TEST_DEL_TITLE"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do stakeholder-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create stakeholder classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create stakeholder classification without title fails"
xbe_json do stakeholder-classifications create --leverage-factor 3
assert_failure

test_name "Update without any fields fails"
xbe_json do stakeholder-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
