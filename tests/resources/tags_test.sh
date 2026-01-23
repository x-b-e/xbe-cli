#!/bin/bash
#
# XBE CLI Integration Tests: Tags and Tag Categories
#
# Tests CRUD operations for both tags and tag_categories resources.
# Tags depend on tag_categories, so both are tested together.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CATEGORY_ID=""
CREATED_TAG_ID=""

describe "Resource: tags and tag_categories"

# ============================================================================
# TAG CATEGORY CREATE Tests
# ============================================================================

test_name "Create tag category with required fields"
TEST_CAT_NAME=$(unique_name "TagCat")
TEST_CAT_SLUG="cat-$(date +%s | tail -c 6)"

xbe_json do tag-categories create \
    --name "$TEST_CAT_NAME" \
    --slug "$TEST_CAT_SLUG" \
    --can-apply-to "Comment"

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

# Only continue if we successfully created a tag category
if [[ -z "$CREATED_CATEGORY_ID" || "$CREATED_CATEGORY_ID" == "null" ]]; then
    echo "Cannot continue without a valid tag category ID"
    run_tests
fi

test_name "Create tag category with description"
TEST_CAT_NAME2=$(unique_name "TagCat2")
TEST_CAT_SLUG2="cat2-$(date +%s | tail -c 6)"
xbe_json do tag-categories create \
    --name "$TEST_CAT_NAME2" \
    --slug "$TEST_CAT_SLUG2" \
    --can-apply-to "Comment" \
    --description "Test category with description"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tag-categories" "$id"
    pass
else
    fail "Failed to create tag category with description"
fi

test_name "Create tag category with multiple can-apply-to"
TEST_CAT_NAME3=$(unique_name "TagCat3")
TEST_CAT_SLUG3="cat3-$(date +%s | tail -c 6)"
xbe_json do tag-categories create \
    --name "$TEST_CAT_NAME3" \
    --slug "$TEST_CAT_SLUG3" \
    --can-apply-to "Comment,FileAttachment"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tag-categories" "$id"
    pass
else
    fail "Failed to create tag category with multiple can-apply-to"
fi

# ============================================================================
# TAG CATEGORY UPDATE Tests
# ============================================================================

test_name "Update tag category name"
UPDATED_CAT_NAME=$(unique_name "UpdatedTagCat")
xbe_json do tag-categories update "$CREATED_CATEGORY_ID" --name "$UPDATED_CAT_NAME"
assert_success

test_name "Update tag category description"
xbe_json do tag-categories update "$CREATED_CATEGORY_ID" --description "Updated description"
assert_success

# ============================================================================
# TAG CATEGORY LIST Tests
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

test_name "List tag categories with --limit"
xbe_json view tag-categories list --limit 3
assert_success

test_name "List tag categories with --offset"
xbe_json view tag-categories list --limit 3 --offset 3
assert_success

# ============================================================================
# TAG CREATE Tests
# ============================================================================

test_name "Create tag with required fields"
TEST_TAG_NAME=$(unique_name "Tag")

xbe_json do tags create \
    --name "$TEST_TAG_NAME" \
    --tag-category "$CREATED_CATEGORY_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TAG_ID=$(json_get ".id")
    if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" ]]; then
        register_cleanup "tags" "$CREATED_TAG_ID"
        pass
    else
        fail "Created tag but no ID returned"
    fi
else
    fail "Failed to create tag"
fi

test_name "Create another tag in same category"
TEST_TAG_NAME2=$(unique_name "Tag2")
xbe_json do tags create \
    --name "$TEST_TAG_NAME2" \
    --tag-category "$CREATED_CATEGORY_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tags" "$id"
    pass
else
    fail "Failed to create another tag"
fi

# ============================================================================
# TAG UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" ]]; then
    test_name "Update tag name"
    UPDATED_TAG_NAME=$(unique_name "UpdatedTag")
    xbe_json do tags update "$CREATED_TAG_ID" --name "$UPDATED_TAG_NAME"
    assert_success
fi

# ============================================================================
# TAG LIST Tests
# ============================================================================

test_name "List tags"
xbe_json view tags list --limit 5
assert_success

test_name "List tags returns array"
xbe_json view tags list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tags"
fi

test_name "List tags with --tag-category-id filter"
xbe_json view tags list --tag-category-id "$CREATED_CATEGORY_ID" --limit 10
assert_success

test_name "List tags with --limit"
xbe_json view tags list --limit 3
assert_success

test_name "List tags with --offset"
xbe_json view tags list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete tag requires --confirm flag"
if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" ]]; then
    xbe_run do tags delete "$CREATED_TAG_ID"
    assert_failure
else
    skip "No tag ID for delete test"
fi

test_name "Delete tag with --confirm"
# Create a tag specifically for deletion
TEST_DEL_TAG=$(unique_name "DeleteTag")
xbe_json do tags create \
    --name "$TEST_DEL_TAG" \
    --tag-category "$CREATED_CATEGORY_ID"
if [[ $status -eq 0 ]]; then
    DEL_TAG_ID=$(json_get ".id")
    xbe_run do tags delete "$DEL_TAG_ID" --confirm
    assert_success
else
    skip "Could not create tag for deletion test"
fi

test_name "Delete tag category requires --confirm flag"
xbe_run do tag-categories delete "$CREATED_CATEGORY_ID"
assert_failure

test_name "Delete tag category with --confirm"
# Create a tag category specifically for deletion
TEST_DEL_CAT=$(unique_name "DeleteCat")
TEST_DEL_CAT_SLUG="del-$(date +%s | tail -c 6)"
xbe_json do tag-categories create \
    --name "$TEST_DEL_CAT" \
    --slug "$TEST_DEL_CAT_SLUG" \
    --can-apply-to "Comment"
if [[ $status -eq 0 ]]; then
    DEL_CAT_ID=$(json_get ".id")
    xbe_run do tag-categories delete "$DEL_CAT_ID" --confirm
    assert_success
else
    skip "Could not create tag category for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tag category without name fails"
xbe_json do tag-categories create --slug "test" --can-apply-to "Comment"
assert_failure

test_name "Create tag category without slug fails"
xbe_json do tag-categories create --name "Test" --can-apply-to "Comment"
assert_failure

test_name "Create tag category without can-apply-to fails"
xbe_json do tag-categories create --name "Test" --slug "test"
assert_failure

test_name "Create tag without name fails"
xbe_json do tags create --tag-category "$CREATED_CATEGORY_ID"
assert_failure

test_name "Create tag without tag-category fails"
xbe_json do tags create --name "Test"
assert_failure

test_name "Update tag category without any fields fails"
xbe_json do tag-categories update "$CREATED_CATEGORY_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
