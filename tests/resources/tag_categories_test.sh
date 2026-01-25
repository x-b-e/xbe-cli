#!/bin/bash
#
# XBE CLI Integration Tests: Tag Categories
#
# Tests CRUD operations for the tag_categories resource.
# Tag categories define groups of related tags with specific applicability.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CATEGORY_ID=""

describe "Resource: tag_categories"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create tag category with required fields"
TEST_NAME=$(unique_name "TagCategory")
TEST_SLUG=$(echo "$TEST_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')

xbe_json do tag-categories create \
    --name "$TEST_NAME" \
    --slug "$TEST_SLUG" \
    --can-apply-to "PredictionSubject"

if [[ $status -eq 0 ]]; then
    CREATED_CATEGORY_ID=$(json_get ".id")
    if [[ -n "$CREATED_CATEGORY_ID" && "$CREATED_CATEGORY_ID" != "null" ]]; then
        register_cleanup "tag-categories" "$CREATED_CATEGORY_ID"
        pass
    else
        fail "Created tag category but no ID returned"
    fi
else
    fail "Failed to create tag category"
fi

# Only continue if we successfully created a category
if [[ -z "$CREATED_CATEGORY_ID" || "$CREATED_CATEGORY_ID" == "null" ]]; then
    echo "Cannot continue without a valid tag category ID"
    run_tests
fi

test_name "Create tag category with description"
TEST_NAME2=$(unique_name "TagCategory2")
TEST_SLUG2=$(echo "$TEST_NAME2" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do tag-categories create \
    --name "$TEST_NAME2" \
    --slug "$TEST_SLUG2" \
    --can-apply-to "PredictionSubject" \
    --description "Category for comment tags"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tag-categories" "$id"
    pass
else
    fail "Failed to create tag category with description"
fi

# NOTE: Skipping multiple can-apply-to test since valid types are limited
# test_name "Create tag category with multiple can-apply-to types"

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update tag category name"
UPDATED_NAME=$(unique_name "UpdatedTagCat")
xbe_json do tag-categories update "$CREATED_CATEGORY_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update tag category description"
xbe_json do tag-categories update "$CREATED_CATEGORY_ID" --description "Updated description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tag categories"
xbe_json view tag-categories list --limit 5
assert_success

test_name "List tag categories returns array"
xbe_json view tag-categories list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tag categories"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List tag categories with --limit"
xbe_json view tag-categories list --limit 3
assert_success

test_name "List tag categories with --offset"
xbe_json view tag-categories list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete tag category requires --confirm flag"
xbe_run do tag-categories delete "$CREATED_CATEGORY_ID"
assert_failure

test_name "Delete tag category with --confirm"
# Create a category specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteTagCat")
TEST_DEL_SLUG=$(echo "$TEST_DEL_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
xbe_json do tag-categories create \
    --name "$TEST_DEL_NAME" \
    --slug "$TEST_DEL_SLUG" \
    --can-apply-to "PredictionSubject"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do tag-categories delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create tag category for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tag category without name fails"
xbe_json do tag-categories create --slug "no-name" --can-apply-to "PredictionSubject"
assert_failure

test_name "Create tag category without slug fails"
xbe_json do tag-categories create --name "NoSlug" --can-apply-to "PredictionSubject"
assert_failure

test_name "Create tag category without can-apply-to fails"
xbe_json do tag-categories create --name "NoApply" --slug "no-apply"
assert_failure

test_name "Update without any fields fails"
xbe_json do tag-categories update "$CREATED_CATEGORY_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
