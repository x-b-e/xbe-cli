#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Cancellation Reason Types
#
# Tests CRUD operations for the job_production_plan_cancellation_reason_types resource.
# These types define reasons why job production plans can be cancelled.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TYPE_ID=""

describe "Resource: job_production_plan_cancellation_reason_types"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cancellation reason type with required fields"
TEST_NAME=$(unique_name "CancellationType")
TEST_SLUG=$(echo "$TEST_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')

xbe_json do job-production-plan-cancellation-reason-types create \
    --name "$TEST_NAME" \
    --slug "$TEST_SLUG"

if [[ $status -eq 0 ]]; then
    CREATED_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TYPE_ID" && "$CREATED_TYPE_ID" != "null" ]]; then
        register_cleanup "job-production-plan-cancellation-reason-types" "$CREATED_TYPE_ID"
        pass
    else
        fail "Created cancellation reason type but no ID returned"
    fi
else
    fail "Failed to create cancellation reason type"
fi

# Only continue if we successfully created a type
if [[ -z "$CREATED_TYPE_ID" || "$CREATED_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid cancellation reason type ID"
    run_tests
fi

test_name "Create cancellation reason type with description"
TEST_NAME2=$(unique_name "CancellationType2")
TEST_SLUG2=$(echo "$TEST_NAME2" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do job-production-plan-cancellation-reason-types create \
    --name "$TEST_NAME2" \
    --slug "$TEST_SLUG2" \
    --description "Cancelled due to weather conditions"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-production-plan-cancellation-reason-types" "$id"
    pass
else
    fail "Failed to create cancellation reason type with description"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update cancellation reason type name"
UPDATED_NAME=$(unique_name "UpdatedCancel")
xbe_json do job-production-plan-cancellation-reason-types update "$CREATED_TYPE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update cancellation reason type description"
xbe_json do job-production-plan-cancellation-reason-types update "$CREATED_TYPE_ID" --description "Updated description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List cancellation reason types"
xbe_json view job-production-plan-cancellation-reason-types list --limit 5
assert_success

test_name "List cancellation reason types returns array"
xbe_json view job-production-plan-cancellation-reason-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list cancellation reason types"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List cancellation reason types with --limit"
xbe_json view job-production-plan-cancellation-reason-types list --limit 3
assert_success

test_name "List cancellation reason types with --offset"
xbe_json view job-production-plan-cancellation-reason-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete cancellation reason type requires --confirm flag"
xbe_run do job-production-plan-cancellation-reason-types delete "$CREATED_TYPE_ID"
assert_failure

test_name "Delete cancellation reason type with --confirm"
# Create a type specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteCancel")
TEST_DEL_SLUG=$(echo "$TEST_DEL_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do job-production-plan-cancellation-reason-types create \
    --name "$TEST_DEL_NAME" \
    --slug "$TEST_DEL_SLUG"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do job-production-plan-cancellation-reason-types delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create cancellation reason type for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cancellation reason type without name fails"
xbe_json do job-production-plan-cancellation-reason-types create --slug "no-name"
assert_failure

test_name "Create cancellation reason type without slug fails"
xbe_json do job-production-plan-cancellation-reason-types create --name "NoSlug"
assert_failure

test_name "Update without any fields fails"
xbe_json do job-production-plan-cancellation-reason-types update "$CREATED_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
