#!/bin/bash
#
# XBE CLI Integration Tests: Platform Statuses
#
# Tests CRUD operations for the platform-statuses resource.
#
# COVERAGE: Create, update, delete + list pagination
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_STATUS_ID=""

describe "Resource: platform-statuses"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create platform status with required fields"
TEST_TITLE=$(unique_name "PlatformStatus")
xbe_json do platform-statuses create \
    --title "$TEST_TITLE" \
    --description "Test platform status description"

if [[ $status -eq 0 ]]; then
    CREATED_STATUS_ID=$(json_get ".id")
    if [[ -n "$CREATED_STATUS_ID" && "$CREATED_STATUS_ID" != "null" ]]; then
        register_cleanup "platform-statuses" "$CREATED_STATUS_ID"
        pass
    else
        fail "Created platform status but no ID returned"
    fi
else
    fail "Failed to create platform status"
fi

# Only continue if we successfully created a platform status
if [[ -z "$CREATED_STATUS_ID" || "$CREATED_STATUS_ID" == "null" ]]; then
    echo "Cannot continue without a valid platform status ID"
    run_tests
fi

test_name "Create platform status with all attributes"
TEST_TITLE2=$(unique_name "PlatformStatusFull")
xbe_json do platform-statuses create \
    --title "$TEST_TITLE2" \
    --description "Full platform status details" \
    --published-at "2024-05-01T00:00:00Z" \
    --start-at "2024-05-01T01:00:00Z" \
    --end-at "2024-05-01T03:00:00Z"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "platform-statuses" "$id"
    pass
else
    fail "Failed to create platform status with all attributes"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update platform status title"
xbe_json do platform-statuses update "$CREATED_STATUS_ID" --title "Updated $TEST_TITLE"
assert_success

test_name "Update platform status description"
xbe_json do platform-statuses update "$CREATED_STATUS_ID" --description "Updated description"
assert_success

test_name "Update platform status published-at"
xbe_json do platform-statuses update "$CREATED_STATUS_ID" --published-at "2024-06-01T00:00:00Z"
assert_success

test_name "Update platform status start-at"
xbe_json do platform-statuses update "$CREATED_STATUS_ID" --start-at "2024-06-01T01:00:00Z"
assert_success

test_name "Update platform status end-at"
xbe_json do platform-statuses update "$CREATED_STATUS_ID" --end-at "2024-06-01T02:00:00Z"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show platform status"
xbe_json view platform-statuses show "$CREATED_STATUS_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List platform statuses"
xbe_json view platform-statuses list --limit 5
assert_success

test_name "List platform statuses returns array"
xbe_json view platform-statuses list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list platform statuses"
fi

test_name "List platform statuses with --offset"
xbe_json view platform-statuses list --limit 5 --offset 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete platform status requires --confirm flag"
xbe_json do platform-statuses delete "$CREATED_STATUS_ID"
assert_failure

test_name "Delete platform status with --confirm"
TEST_DEL_TITLE=$(unique_name "PlatformStatusDelete")
xbe_json do platform-statuses create \
    --title "$TEST_DEL_TITLE" \
    --description "Delete test status"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do platform-statuses delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create platform status for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create platform status without title fails"
xbe_json do platform-statuses create --description "Missing title"
assert_failure

test_name "Create platform status without description fails"
xbe_json do platform-statuses create --title "Missing description"
assert_failure

test_name "Update without any fields fails"
xbe_json do platform-statuses update "$CREATED_STATUS_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
