#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subject Gap Portions
#
# Tests list/show/create/update/delete operations for prediction-subject-gap-portions.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_GAP_ID="${XBE_TEST_PREDICTION_SUBJECT_GAP_ID:-}"

SAMPLE_ID=""
SAMPLE_GAP_ID=""
SAMPLE_CREATED_BY_ID=""
CREATED_ID=""

NAME="Gap Portion $(unique_suffix)"
UPDATED_NAME="Gap Portion Updated $(unique_suffix)"
DESCRIPTION="Portion description $(unique_suffix)"
UPDATED_DESCRIPTION="Updated description $(unique_suffix)"
AMOUNT="42.5"
UPDATED_AMOUNT="55.25"

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"cannot"* ]] || [[ "$msg" == *"insufficient"* ]] || [[ "$msg" == *"not found"* ]] || [[ "$msg" == *"Not Found"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: prediction-subject-gap-portions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction subject gap portions"
xbe_json view prediction-subject-gap-portions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_GAP_ID=$(json_get ".[0].prediction_subject_gap_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
else
    fail "Failed to list prediction subject gap portions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction subject gap portion"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-subject-gap-portions show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output"; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show prediction subject gap portion: $output"
        fi
    fi
else
    skip "No prediction subject gap portion ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction subject gap portions with --prediction-subject-gap filter"
FILTER_GAP_ID="${SAMPLE_GAP_ID:-$PREDICTION_SUBJECT_GAP_ID}"
if [[ -n "$FILTER_GAP_ID" && "$FILTER_GAP_ID" != "null" ]]; then
    xbe_json view prediction-subject-gap-portions list --prediction-subject-gap "$FILTER_GAP_ID" --limit 5
    assert_success
else
    skip "No prediction subject gap ID available (set XBE_TEST_PREDICTION_SUBJECT_GAP_ID)"
fi

test_name "List prediction subject gap portions with --status filter"
xbe_json view prediction-subject-gap-portions list --status draft --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction subject gap portion requires required flags"
xbe_run do prediction-subject-gap-portions create
assert_failure

test_name "Create prediction subject gap portion"
if [[ -n "$PREDICTION_SUBJECT_GAP_ID" && "$PREDICTION_SUBJECT_GAP_ID" != "null" ]]; then
    xbe_json do prediction-subject-gap-portions create \
        --prediction-subject-gap "$PREDICTION_SUBJECT_GAP_ID" \
        --name "$NAME" \
        --amount "$AMOUNT" \
        --status draft \
        --description "$DESCRIPTION"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "prediction-subject-gap-portions" "$CREATED_ID"
            pass
        else
            fail "Created prediction subject gap portion but no ID returned"
        fi
    else
        if update_blocked_message "$output"; then
            skip "Create blocked by server policy or invalid prediction subject gap"
        else
            fail "Failed to create prediction subject gap portion: $output"
        fi
    fi
else
    skip "No prediction subject gap ID available (set XBE_TEST_PREDICTION_SUBJECT_GAP_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update prediction subject gap portion attributes"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do prediction-subject-gap-portions update "$CREATED_ID" \
        --name "$UPDATED_NAME" \
        --amount "$UPDATED_AMOUNT" \
        --description "$UPDATED_DESCRIPTION"
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Update blocked by server policy or validation"
        else
            fail "Failed to update prediction subject gap portion: $output"
        fi
    fi
else
    skip "No created prediction subject gap portion available for update"
fi

test_name "Update prediction subject gap portion status"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do prediction-subject-gap-portions update "$CREATED_ID" --status approved
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Status update blocked by server policy or validation"
        else
            fail "Failed to update prediction subject gap portion status: $output"
        fi
    fi
else
    skip "No created prediction subject gap portion available for status update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete prediction subject gap portion requires --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-subject-gap-portions delete "$CREATED_ID"
    assert_failure
else
    skip "No created prediction subject gap portion available for delete"
fi

test_name "Delete prediction subject gap portion"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-subject-gap-portions delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete prediction subject gap portion: $output"
        fi
    fi
else
    skip "No created prediction subject gap portion available for delete"
fi

run_tests
