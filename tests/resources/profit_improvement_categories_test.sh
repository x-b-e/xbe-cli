#!/bin/bash
#
# XBE CLI Integration Tests: Profit Improvement Categories
#
# Tests CRUD operations for the profit_improvement_categories resource.
# Profit improvement categories classify profit enhancement initiatives.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PIC_ID=""

describe "Resource: profit-improvement-categories"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create profit-improvement-category with required fields"
TEST_NAME=$(unique_name "PIC")
xbe_json do profit-improvement-categories create --name "$TEST_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_PIC_ID=$(json_get ".id")
    if [[ -n "$CREATED_PIC_ID" && "$CREATED_PIC_ID" != "null" ]]; then
        register_cleanup "profit-improvement-categories" "$CREATED_PIC_ID"
        pass
    else
        fail "Created profit-improvement-category but no ID returned"
    fi
else
    fail "Failed to create profit-improvement-category"
fi

# Only continue if we successfully created a profit-improvement-category
if [[ -z "$CREATED_PIC_ID" || "$CREATED_PIC_ID" == "null" ]]; then
    echo "Cannot continue without a valid profit-improvement-category ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update profit-improvement-category name"
UPDATED_NAME=$(unique_name "UpdatedPIC")
xbe_json do profit-improvement-categories update "$CREATED_PIC_ID" --name "$UPDATED_NAME"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List profit improvement categories"
xbe_json view profit-improvement-categories list
assert_success

test_name "List profit improvement categories returns array"
xbe_json view profit-improvement-categories list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list profit improvement categories"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List profit improvement categories with --name filter"
xbe_json view profit-improvement-categories list --name "safety"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List profit improvement categories with --limit"
xbe_json view profit-improvement-categories list --limit 5
assert_success

test_name "List profit improvement categories with --offset"
xbe_json view profit-improvement-categories list --limit 5 --offset 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete profit-improvement-category requires --confirm flag"
xbe_json do profit-improvement-categories delete "$CREATED_PIC_ID"
assert_failure

test_name "Delete profit-improvement-category with --confirm"
# Create a profit-improvement-category specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMe")
xbe_json do profit-improvement-categories create --name "$TEST_DEL_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do profit-improvement-categories delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create profit-improvement-category for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create profit-improvement-category without name fails"
xbe_json do profit-improvement-categories create
assert_failure

test_name "Update without any fields fails"
xbe_json do profit-improvement-categories update "$CREATED_PIC_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
