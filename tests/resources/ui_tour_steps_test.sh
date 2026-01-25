#!/bin/bash
#
# XBE CLI Integration Tests: UI Tour Steps
#
# Tests CRUD operations for the ui-tour-steps resource.
# UI tour steps define ordered UI walkthrough content.
#
# NOTE: This test may create a UI tour via direct API call when XBE_TOKEN is set.
#       Alternatively, set XBE_TEST_UI_TOUR_ID to use an existing UI tour.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_UI_TOUR_ID=""
CREATED_STEP_ID=""
UI_TOUR_CREATED_VIA_API="0"

check_jq

describe "Resource: ui-tour-steps"

cleanup_ui_tour() {
    if [[ "$UI_TOUR_CREATED_VIA_API" == "1" && -n "$CREATED_UI_TOUR_ID" && -n "$XBE_TOKEN" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/ui-tours/$CREATED_UI_TOUR_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" > /dev/null || true
    fi
}

trap 'cleanup_ui_tour; run_cleanup' EXIT

# =========================================================================
# Prerequisites - UI tour
# =========================================================================

test_name "Resolve UI tour prerequisite"
if [[ -n "$XBE_TEST_UI_TOUR_ID" ]]; then
    CREATED_UI_TOUR_ID="$XBE_TEST_UI_TOUR_ID"
    echo "    Using XBE_TEST_UI_TOUR_ID: $CREATED_UI_TOUR_ID"
    pass
else
    if [[ -z "$XBE_TOKEN" ]]; then
        skip "XBE_TEST_UI_TOUR_ID or XBE_TOKEN required to create UI tour"
        run_tests
    fi

    UI_TOUR_NAME=$(unique_name "UITour")
    UI_TOUR_ABBR="ui-tour-$(date +%s)-${RANDOM}"
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
        fail "Failed to create UI tour"
        echo "Response: $create_response"
        run_tests
    fi
fi

# =========================================================================
# CREATE Tests
# =========================================================================

test_name "Create UI tour step with required fields"
STEP_NAME=$(unique_name "UITourStep")
STEP_ABBR="step-$(date +%s)-${RANDOM}"
STEP_CONTENT="Test UI tour step content $(date +%s)"

xbe_json do ui-tour-steps create \
    --name "$STEP_NAME" \
    --content "$STEP_CONTENT" \
    --abbreviation "$STEP_ABBR" \
    --ui-tour "$CREATED_UI_TOUR_ID" \
    --sequence 1

if [[ $status -eq 0 ]]; then
    CREATED_STEP_ID=$(json_get ".id")
    if [[ -n "$CREATED_STEP_ID" && "$CREATED_STEP_ID" != "null" ]]; then
        register_cleanup "ui-tour-steps" "$CREATED_STEP_ID"
        pass
    else
        fail "Created UI tour step but no ID returned"
    fi
else
    fail "Failed to create UI tour step"
fi

if [[ -z "$CREATED_STEP_ID" || "$CREATED_STEP_ID" == "null" ]]; then
    echo "Cannot continue without a valid UI tour step ID"
    run_tests
fi

# =========================================================================
# UPDATE Tests - Attributes
# =========================================================================

test_name "Update UI tour step --name"
xbe_json do ui-tour-steps update "$CREATED_STEP_ID" --name "Updated step name $(date +%s)"
assert_success

test_name "Update UI tour step --content"
xbe_json do ui-tour-steps update "$CREATED_STEP_ID" --content "Updated content $(date +%s)"
assert_success

test_name "Update UI tour step --abbreviation"
NEW_ABBR="step-updated-$(date +%s)-${RANDOM}"
xbe_json do ui-tour-steps update "$CREATED_STEP_ID" --abbreviation "$NEW_ABBR"
assert_success

test_name "Update UI tour step --sequence"
xbe_json do ui-tour-steps update "$CREATED_STEP_ID" --sequence 2
assert_success

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show UI tour step"
xbe_json view ui-tour-steps show "$CREATED_STEP_ID"
assert_success

# =========================================================================
# LIST Tests - Basic
# =========================================================================

test_name "List UI tour steps"
xbe_json view ui-tour-steps list --limit 5
assert_success

test_name "List UI tour steps returns array"
xbe_json view ui-tour-steps list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list UI tour steps"
fi

# =========================================================================
# LIST Tests - Filters
# =========================================================================

test_name "List UI tour steps with --created-at-min"
NOW_TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
xbe_json view ui-tour-steps list --created-at-min "$NOW_TS" --limit 5
assert_success

test_name "List UI tour steps with --created-at-max"
xbe_json view ui-tour-steps list --created-at-max "$NOW_TS" --limit 5
assert_success

test_name "List UI tour steps with --is-created-at=true"
xbe_json view ui-tour-steps list --is-created-at true --limit 5
assert_success

test_name "List UI tour steps with --is-created-at=false"
xbe_json view ui-tour-steps list --is-created-at false --limit 5
assert_success

test_name "List UI tour steps with --updated-at-min"
xbe_json view ui-tour-steps list --updated-at-min "$NOW_TS" --limit 5
assert_success

test_name "List UI tour steps with --updated-at-max"
xbe_json view ui-tour-steps list --updated-at-max "$NOW_TS" --limit 5
assert_success

test_name "List UI tour steps with --is-updated-at=true"
xbe_json view ui-tour-steps list --is-updated-at true --limit 5
assert_success

test_name "List UI tour steps with --is-updated-at=false"
xbe_json view ui-tour-steps list --is-updated-at false --limit 5
assert_success

test_name "List UI tour steps with --not-id"
xbe_json view ui-tour-steps list --not-id "$CREATED_STEP_ID" --limit 5
assert_success

# =========================================================================
# LIST Tests - Pagination
# =========================================================================

test_name "List UI tour steps with --limit"
xbe_json view ui-tour-steps list --limit 3
assert_success

test_name "List UI tour steps with --offset"
xbe_json view ui-tour-steps list --limit 3 --offset 3
assert_success

# =========================================================================
# DELETE Tests
# =========================================================================

test_name "Delete UI tour step requires --confirm flag"
xbe_run do ui-tour-steps delete "$CREATED_STEP_ID"
assert_failure

test_name "Delete UI tour step with --confirm"
DEL_NAME=$(unique_name "UITourStepDelete")
DEL_ABBR="step-delete-$(date +%s)-${RANDOM}"
DEL_CONTENT="Delete step content $(date +%s)"

xbe_json do ui-tour-steps create \
    --name "$DEL_NAME" \
    --content "$DEL_CONTENT" \
    --abbreviation "$DEL_ABBR" \
    --ui-tour "$CREATED_UI_TOUR_ID"

if [[ $status -eq 0 ]]; then
    DEL_STEP_ID=$(json_get ".id")
    if [[ -n "$DEL_STEP_ID" && "$DEL_STEP_ID" != "null" ]]; then
        xbe_run do ui-tour-steps delete "$DEL_STEP_ID" --confirm
        assert_success
    else
        skip "Could not create UI tour step for deletion test"
    fi
else
    skip "Could not create UI tour step for deletion test"
fi

# =========================================================================
# Error Cases
# =========================================================================

test_name "Create UI tour step without --name fails"
xbe_json do ui-tour-steps create --content "Missing name" --abbreviation "missing-name" --ui-tour "$CREATED_UI_TOUR_ID"
assert_failure

test_name "Create UI tour step without --content fails"
xbe_json do ui-tour-steps create --name "Missing content" --abbreviation "missing-content" --ui-tour "$CREATED_UI_TOUR_ID"
assert_failure

test_name "Create UI tour step without --abbreviation fails"
xbe_json do ui-tour-steps create --name "Missing abbrev" --content "Missing abbrev" --ui-tour "$CREATED_UI_TOUR_ID"
assert_failure

test_name "Create UI tour step without --ui-tour fails"
xbe_json do ui-tour-steps create --name "Missing tour" --content "Missing tour" --abbreviation "missing-tour"
assert_failure

test_name "Update UI tour step without any fields fails"
xbe_run do ui-tour-steps update "$CREATED_STEP_ID"
assert_failure

# =========================================================================
# Summary
# =========================================================================

run_tests
