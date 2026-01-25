#!/bin/bash
#
# XBE CLI Integration Tests: User UI Tours
#
# Tests list/show/create/update/delete operations for the user-ui-tours resource.
#
# COVERAGE: List filters + create/update attributes + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_USER_UI_TOUR_ID=""
CREATED_UI_TOUR_ID=""
UI_TOUR_CREATED_VIA_API="0"
USER_ID=""

check_jq

describe "Resource: user-ui-tours"

cleanup_ui_tour() {
    if [[ "$UI_TOUR_CREATED_VIA_API" == "1" && -n "$CREATED_UI_TOUR_ID" && -n "$XBE_TOKEN" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/ui-tours/$CREATED_UI_TOUR_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" > /dev/null || true
    fi
}

trap 'cleanup_ui_tour; run_cleanup' EXIT

# =========================================================================
# Resolve current user
# =========================================================================

test_name "Resolve current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    USER_ID=$(json_get ".id")
    if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
        pass
    else
        fail "Whoami returned no user ID"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_USER_ID" ]]; then
        USER_ID="$XBE_TEST_USER_ID"
        pass
    else
        skip "Unable to resolve current user (set XBE_TEST_USER_ID)"
        run_tests
    fi
fi

# =========================================================================
# Prerequisites - UI tour
# =========================================================================

test_name "Resolve UI tour prerequisite"
if [[ -n "$XBE_TOKEN" ]]; then
    UI_TOUR_NAME=$(unique_name "UITour")
    UI_TOUR_ABBR="user-ui-tour-$(date +%s)-${RANDOM}"
    UI_TOUR_DESC="CLI test UI tour"

    create_response=$(curl -sS -X POST "$XBE_BASE_URL/v1/ui-tours" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"data\":{\"type\":\"ui-tours\",\"attributes\":{\"name\":\"$UI_TOUR_NAME\",\"abbreviation\":\"$UI_TOUR_ABBR\",\"description\":\"$UI_TOUR_DESC\"}}}")

    CREATED_UI_TOUR_ID=$(echo "$create_response" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_UI_TOUR_ID" ]]; then
        UI_TOUR_CREATED_VIA_API="1"
        pass
    else
        if [[ -n "$XBE_TEST_UI_TOUR_ID" ]]; then
            CREATED_UI_TOUR_ID="$XBE_TEST_UI_TOUR_ID"
            pass
        else
            fail "Failed to create UI tour"
            echo "Response: $create_response"
            run_tests
        fi
    fi
else
    if [[ -n "$XBE_TEST_UI_TOUR_ID" ]]; then
        CREATED_UI_TOUR_ID="$XBE_TEST_UI_TOUR_ID"
        pass
    else
        skip "XBE_TOKEN or XBE_TEST_UI_TOUR_ID required to create UI tour"
        run_tests
    fi
fi

# =========================================================================
# CREATE Tests
# =========================================================================

test_name "Create user UI tour with --completed-at"
COMPLETED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

xbe_json do user-ui-tours create \
    --user "$USER_ID" \
    --ui-tour "$CREATED_UI_TOUR_ID" \
    --completed-at "$COMPLETED_AT"

if [[ $status -eq 0 ]]; then
    CREATED_USER_UI_TOUR_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_UI_TOUR_ID" && "$CREATED_USER_UI_TOUR_ID" != "null" ]]; then
        register_cleanup "user-ui-tours" "$CREATED_USER_UI_TOUR_ID"
        pass
    else
        fail "Created user UI tour but no ID returned"
    fi
else
    fail "Failed to create user UI tour"
fi

if [[ -z "$CREATED_USER_UI_TOUR_ID" || "$CREATED_USER_UI_TOUR_ID" == "null" ]]; then
    echo "Cannot continue without a valid user UI tour ID"
    run_tests
fi

# =========================================================================
# UPDATE Tests - Attributes
# =========================================================================

test_name "Update user UI tour to skipped"
SKIPPED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
xbe_json do user-ui-tours update "$CREATED_USER_UI_TOUR_ID" --completed-at "" --skipped-at "$SKIPPED_AT"
assert_success

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show user UI tour"
xbe_json view user-ui-tours show "$CREATED_USER_UI_TOUR_ID"
assert_success

# =========================================================================
# LIST Tests - Basic
# =========================================================================

test_name "List user UI tours"
xbe_json view user-ui-tours list --limit 5
assert_success

test_name "List user UI tours returns array"
xbe_json view user-ui-tours list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user UI tours"
fi

# =========================================================================
# LIST Tests - Filters
# =========================================================================

test_name "List user UI tours with --created-at-min"
NOW_TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
xbe_json view user-ui-tours list --created-at-min "$NOW_TS" --limit 5
assert_success

test_name "List user UI tours with --created-at-max"
xbe_json view user-ui-tours list --created-at-max "$NOW_TS" --limit 5
assert_success

test_name "List user UI tours with --is-created-at=true"
xbe_json view user-ui-tours list --is-created-at true --limit 5
assert_success

test_name "List user UI tours with --is-created-at=false"
xbe_json view user-ui-tours list --is-created-at false --limit 5
assert_success

test_name "List user UI tours with --updated-at-min"
xbe_json view user-ui-tours list --updated-at-min "$NOW_TS" --limit 5
assert_success

test_name "List user UI tours with --updated-at-max"
xbe_json view user-ui-tours list --updated-at-max "$NOW_TS" --limit 5
assert_success

test_name "List user UI tours with --is-updated-at=true"
xbe_json view user-ui-tours list --is-updated-at true --limit 5
assert_success

test_name "List user UI tours with --is-updated-at=false"
xbe_json view user-ui-tours list --is-updated-at false --limit 5
assert_success

test_name "List user UI tours with --not-id"
xbe_json view user-ui-tours list --not-id "$CREATED_USER_UI_TOUR_ID" --limit 5
assert_success

# =========================================================================
# Error Cases
# =========================================================================

test_name "Create user UI tour without --user fails"
xbe_json do user-ui-tours create --ui-tour "$CREATED_UI_TOUR_ID" --completed-at "$COMPLETED_AT"
assert_failure

test_name "Create user UI tour without --ui-tour fails"
xbe_json do user-ui-tours create --user "$USER_ID" --completed-at "$COMPLETED_AT"
assert_failure

test_name "Create user UI tour without completion or skip fails"
xbe_json do user-ui-tours create --user "$USER_ID" --ui-tour "$CREATED_UI_TOUR_ID"
assert_failure

test_name "Create user UI tour with both completion and skip fails"
xbe_json do user-ui-tours create --user "$USER_ID" --ui-tour "$CREATED_UI_TOUR_ID" --completed-at "$COMPLETED_AT" --skipped-at "$SKIPPED_AT"
assert_failure

test_name "Update user UI tour without any fields fails"
xbe_run do user-ui-tours update "$CREATED_USER_UI_TOUR_ID"
assert_failure

# =========================================================================
# DELETE Tests
# =========================================================================

test_name "Delete user UI tour requires --confirm flag"
xbe_run do user-ui-tours delete "$CREATED_USER_UI_TOUR_ID"
assert_failure

test_name "Delete user UI tour with --confirm"
xbe_run do user-ui-tours delete "$CREATED_USER_UI_TOUR_ID" --confirm
assert_success

# =========================================================================
# Summary
# =========================================================================

run_tests
