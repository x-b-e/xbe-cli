#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Time Changes
#
# Tests list, show, create, update, and delete operations for the time-card-time-changes resource.
#
# COVERAGE: List filters + show + create attributes + update/delete (when allowed)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TIME_CARD_TIME_CHANGE_ID=""
TIME_CARD_ID=""
CREATED_BY_ID=""
IS_PROCESSED=""
CREATED_CHANGE_ID=""
CREATED_IS_PROCESSED=""
TIME_CHANGES_ATTRIBUTES=""
UPDATED_TIME_CHANGES=""

SKIP_SAMPLE_TESTS=0


describe "Resource: time-card-time-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card time changes"
xbe_json view time-card-time-changes list --limit 5
assert_success

test_name "List time card time changes returns array"
xbe_json view time-card-time-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time card time changes"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample time card time change"
xbe_json view time-card-time-changes list --limit 1
if [[ $status -eq 0 ]]; then
    TIME_CARD_TIME_CHANGE_ID=$(json_get ".[0].id")
    TIME_CARD_ID=$(json_get ".[0].time_card_id")
    CREATED_BY_ID=$(json_get ".[0].created_by_id")
    IS_PROCESSED=$(json_get ".[0].is_processed")
    if [[ -n "$TIME_CARD_TIME_CHANGE_ID" && "$TIME_CARD_TIME_CHANGE_ID" != "null" ]]; then
        pass
    else
        SKIP_SAMPLE_TESTS=1
        skip "No time card time changes available"
    fi
else
    SKIP_SAMPLE_TESTS=1
    fail "Failed to list time card time changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List time card time changes with --time-card filter"
if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    xbe_json view time-card-time-changes list --time-card "$TIME_CARD_ID" --limit 5
    assert_success
else
    skip "No time card ID available"
fi

test_name "List time card time changes with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view time-card-time-changes list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List time card time changes with --is-processed filter"
if [[ -n "$IS_PROCESSED" && "$IS_PROCESSED" != "null" ]]; then
    xbe_json view time-card-time-changes list --is-processed "$IS_PROCESSED" --limit 5
    assert_success
else
    skip "No processed flag available"
fi

test_name "List time card time changes with --broker filter"
if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    xbe_json view time-card-time-changes list --broker "$XBE_TEST_BROKER_ID" --limit 5
    assert_success
else
    skip "XBE_TEST_BROKER_ID not set"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time card time change"
if [[ $SKIP_SAMPLE_TESTS -eq 0 && -n "$TIME_CARD_TIME_CHANGE_ID" && "$TIME_CARD_TIME_CHANGE_ID" != "null" ]]; then
    xbe_json view time-card-time-changes show "$TIME_CARD_TIME_CHANGE_ID"
    assert_success
else
    skip "No time card time change ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create time card time change without required fields fails"
xbe_run do time-card-time-changes create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Prepare time changes attributes"
if [[ $SKIP_SAMPLE_TESTS -eq 0 && -n "$TIME_CARD_TIME_CHANGE_ID" && "$TIME_CARD_TIME_CHANGE_ID" != "null" ]]; then
    xbe_json view time-card-time-changes show "$TIME_CARD_TIME_CHANGE_ID"
    if [[ $status -eq 0 ]]; then
        NEW_DOWN_MINUTES=$(echo "$output" | jq -r '((.time_changes_details.after.down_minutes // .time_changes_details.before.down_minutes // 0) | tonumber) + 5' 2>/dev/null)
        if [[ -z "$NEW_DOWN_MINUTES" || "$NEW_DOWN_MINUTES" == "null" ]]; then
            NEW_DOWN_MINUTES=5
        fi
        TIME_CHANGES_ATTRIBUTES="{\"down_minutes\": $NEW_DOWN_MINUTES}"
        UPDATED_DOWN_MINUTES=$((NEW_DOWN_MINUTES + 5))
        UPDATED_TIME_CHANGES="{\"down_minutes\": $UPDATED_DOWN_MINUTES}"
        pass
    else
        skip "Failed to load time change details"
    fi
else
    skip "No sample time change available"
fi


test_name "Create time card time change"
if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" && -n "$TIME_CHANGES_ATTRIBUTES" ]]; then
    COMMENT_TEXT="CLI test $(unique_suffix)"
    xbe_json do time-card-time-changes create \
        --time-card "$TIME_CARD_ID" \
        --time-changes-attributes "$TIME_CHANGES_ATTRIBUTES" \
        --comment "$COMMENT_TEXT" \
        --skip-time-card-not-editable true \
        --skip-quantity-validation true
    if [[ $status -eq 0 ]]; then
        CREATED_CHANGE_ID=$(json_get ".id")
        CREATED_IS_PROCESSED=$(json_get ".is_processed")
        if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
            register_cleanup "time-card-time-changes" "$CREATED_CHANGE_ID"
            pass
        else
            fail "Created time card time change but no ID returned"
        fi
    else
        skip "Create failed (time card may not allow changes)"
    fi
else
    skip "No time card ID or time changes attributes available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update time card time change"
if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
    if [[ "$CREATED_IS_PROCESSED" == "false" ]]; then
        UPDATE_COMMENT="Updated comment $(unique_suffix)"
        xbe_json do time-card-time-changes update "$CREATED_CHANGE_ID" \
            --comment "$UPDATE_COMMENT" \
            --time-changes-attributes "$UPDATED_TIME_CHANGES" \
            --skip-quantity-validation true
        assert_success
    else
        xbe_run do time-card-time-changes update "$CREATED_CHANGE_ID" --comment "Should fail"
        assert_failure
    fi
else
    skip "No created time card time change ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete time card time change"
if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
    if [[ "$CREATED_IS_PROCESSED" == "false" ]]; then
        xbe_run do time-card-time-changes delete "$CREATED_CHANGE_ID" --confirm
        assert_success
    else
        xbe_run do time-card-time-changes delete "$CREATED_CHANGE_ID" --confirm
        assert_failure
    fi
else
    skip "No created time card time change ID available"
fi

run_tests
