#!/bin/bash
#
# XBE CLI Integration Tests: Proffers
#
# Tests CRUD operations for the proffers resource.
# Proffers are feature suggestions submitted by users.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROFFER_ID=""

describe "Resource: proffers"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create proffer with required fields"
TITLE="CLI Proffer $(date +%s)"
xbe_json do proffers create --title "$TITLE"

if [[ $status -eq 0 ]]; then
    CREATED_PROFFER_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROFFER_ID" && "$CREATED_PROFFER_ID" != "null" ]]; then
        register_cleanup "proffers" "$CREATED_PROFFER_ID"
        pass
    else
        fail "Created proffer but no ID returned"
    fi
else
    fail "Failed to create proffer"
fi

test_name "Create proffer with description and kind"
SECOND_TITLE="CLI Proffer Detail $(date +%s)"
xbe_json do proffers create --title "$SECOND_TITLE" --description "Proffer details" --kind hot_feed_post

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "proffers" "$id"
    pass
else
    fail "Failed to create proffer with description and kind"
fi

# Only continue if we successfully created a proffer
if [[ -z "$CREATED_PROFFER_ID" || "$CREATED_PROFFER_ID" == "null" ]]; then
    echo "Cannot continue without a valid proffer ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update proffer title"
xbe_json do proffers update "$CREATED_PROFFER_ID" --title "Updated Proffer $(date +%s)"
assert_success

test_name "Update proffer description"
xbe_json do proffers update "$CREATED_PROFFER_ID" --description "Updated description $(date +%s)"
assert_success

test_name "Update proffer kind"
xbe_json do proffers update "$CREATED_PROFFER_ID" --kind make_it_so_action
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List proffers"
xbe_json view proffers list --limit 5
assert_success

test_name "List proffers returns array"
xbe_json view proffers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list proffers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List proffers with --kind filter"
xbe_json view proffers list --kind hot_feed_post --limit 5
assert_success

test_name "List proffers with --created-by filter"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        xbe_json view proffers list --created-by "$CURRENT_USER_ID" --limit 5
        assert_success
    else
        fail "Auth whoami returned no ID"
    fi
else
    fail "Auth whoami failed"
fi

test_name "List proffers with --similar-to-text filter"
xbe_json view proffers list --similar-to-text "Updated description" --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    if [[ "$output" == *"OpenAI"* || "$output" == *"embedding"* || "$output" == *"Embeddings"* ]]; then
        skip "similar-to-text unavailable (embedding service not configured)"
    else
        fail "Failed to list proffers with similar-to-text filter"
    fi
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List proffers with --limit"
xbe_json view proffers list --limit 3
assert_success

test_name "List proffers with --offset"
xbe_json view proffers list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete proffer requires --confirm flag"
xbe_json do proffers delete "$CREATED_PROFFER_ID"
assert_failure

test_name "Delete proffer with --confirm"
# Create a proffer specifically for deletion
DEL_TITLE="CLI Proffer Delete $(date +%s)"
xbe_json do proffers create --title "$DEL_TITLE"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do proffers delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create proffer for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create proffer without title fails"
xbe_json do proffers create
assert_failure

test_name "Update without any fields fails"
xbe_json do proffers update "$CREATED_PROFFER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
