#!/bin/bash
#
# XBE CLI Integration Tests: Native App Releases
#
# Tests CRUD operations for the native-app-releases resource.
# Native app releases track mobile build metadata and release channels.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RELEASE_ID=""
CREATED_RELEASE_SHA=""
CREATED_BUILD_NUMBER=""
CREATED_GIT_TAG=""

RELEASE_DETAILS='[{"channel":"apple-app-store","status":"uploaded"}]'
UPDATED_RELEASE_DETAILS='[{"channel":"google-play-store","status":"released"}]'

describe "Resource: native-app-releases"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create native app release with required fields"
CREATED_RELEASE_SHA="git-sha-$(date +%s)-${RANDOM}"
CREATED_BUILD_NUMBER="build-$(date +%s)-${RANDOM}"

xbe_json do native-app-releases create \
    --git-sha "$CREATED_RELEASE_SHA" \
    --build-number "$CREATED_BUILD_NUMBER"

if [[ $status -eq 0 ]]; then
    CREATED_RELEASE_ID=$(json_get ".id")
    if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
        register_cleanup "native-app-releases" "$CREATED_RELEASE_ID"
        pass
    else
        fail "Created native app release but no ID returned"
    fi
else
    fail "Failed to create native app release"
fi

# Only continue if we successfully created a release
if [[ -z "$CREATED_RELEASE_ID" || "$CREATED_RELEASE_ID" == "null" ]]; then
    echo "Cannot continue without a valid native app release ID"
    run_tests
fi

test_name "Create native app release with all attributes"
CREATED_GIT_TAG="tag-$(date +%s)-${RANDOM}"
FULL_RELEASE_SHA="git-sha-$(date +%s)-${RANDOM}"
FULL_BUILD_NUMBER="build-$(date +%s)-${RANDOM}"

xbe_json do native-app-releases create \
    --git-sha "$FULL_RELEASE_SHA" \
    --build-number "$FULL_BUILD_NUMBER" \
    --git-tag "$CREATED_GIT_TAG" \
    --release-channel-details "$RELEASE_DETAILS" \
    --notes "Initial upload"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "native-app-releases" "$id"
    pass
else
    fail "Failed to create native app release with all attributes"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update native app release notes"
xbe_json do native-app-releases update "$CREATED_RELEASE_ID" --notes "Updated notes"
assert_success

test_name "Update native app release channel details"
xbe_json do native-app-releases update "$CREATED_RELEASE_ID" --release-channel-details "$UPDATED_RELEASE_DETAILS"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List native app releases"
xbe_json view native-app-releases list
assert_success

test_name "List native app releases returns array"
xbe_json view native-app-releases list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list native app releases"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List native app releases with --git-tag"
xbe_json view native-app-releases list --git-tag "$CREATED_GIT_TAG"
assert_success

test_name "List native app releases with --git-sha"
xbe_json view native-app-releases list --git-sha "$CREATED_RELEASE_SHA"
assert_success

test_name "List native app releases with --build-number"
xbe_json view native-app-releases list --build-number "$CREATED_BUILD_NUMBER"
assert_success

test_name "List native app releases with --release-status"
xbe_json view native-app-releases list --release-status "uploaded"
assert_success

test_name "List native app releases with --release-channel"
xbe_json view native-app-releases list --release-channel "apple-app-store"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List native app releases with --limit"
xbe_json view native-app-releases list --limit 5
assert_success

test_name "List native app releases with --offset"
xbe_json view native-app-releases list --limit 5 --offset 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete native app release requires --confirm flag"
xbe_json do native-app-releases delete "$CREATED_RELEASE_ID"
assert_failure

test_name "Delete native app release with --confirm"
DEL_SHA="git-sha-$(date +%s)-${RANDOM}"
DEL_BUILD_NUMBER="build-$(date +%s)-${RANDOM}"

xbe_json do native-app-releases create \
    --git-sha "$DEL_SHA" \
    --build-number "$DEL_BUILD_NUMBER"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    register_cleanup "native-app-releases" "$DEL_ID"
    xbe_json do native-app-releases delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create native app release for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create native app release without required fields fails"
xbe_json do native-app-releases create
assert_failure

test_name "Create native app release with invalid release-channel-details fails"
BAD_SHA="git-sha-$(date +%s)-${RANDOM}"
BAD_BUILD_NUMBER="build-$(date +%s)-${RANDOM}"

xbe_json do native-app-releases create \
    --git-sha "$BAD_SHA" \
    --build-number "$BAD_BUILD_NUMBER" \
    --release-channel-details "not-json"
assert_failure

test_name "Update native app release without fields fails"
xbe_json do native-app-releases update "$CREATED_RELEASE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
