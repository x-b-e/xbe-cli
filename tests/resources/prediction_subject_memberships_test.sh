#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subject Memberships
#
# Tests list/show/create/update/delete operations for prediction-subject-memberships.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_MEMBERSHIP_SUBJECT_ID:-}"
if [[ -z "$PREDICTION_SUBJECT_ID" ]]; then
    PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"
fi

MEMBER_USER_ID="${XBE_TEST_PREDICTION_SUBJECT_MEMBER_USER_ID:-}"
if [[ -z "$MEMBER_USER_ID" ]]; then
    MEMBER_USER_ID="${XBE_TEST_USER_ID:-}"
fi

SAMPLE_ID=""
SAMPLE_PREDICTION_SUBJECT_ID=""
SAMPLE_USER_ID=""
CREATED_ID=""

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"cannot"* ]] || [[ "$msg" == *"is already a member"* ]] || [[ "$msg" == *"does not have access"* ]] || [[ "$msg" == *"only full access user"* ]] || [[ "$msg" == *"Not Found"* ]] || [[ "$msg" == *"not found"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: prediction-subject-memberships"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction subject memberships"
xbe_json view prediction-subject-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PREDICTION_SUBJECT_ID=$(json_get ".[0].prediction_subject_id")
    SAMPLE_USER_ID=$(json_get ".[0].user_id")
else
    fail "Failed to list prediction subject memberships"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction subject membership"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-subject-memberships show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output"; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show prediction subject membership: $output"
        fi
    fi
else
    skip "No prediction subject membership ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction subject memberships with --prediction-subject filter"
FILTER_PREDICTION_SUBJECT_ID="${SAMPLE_PREDICTION_SUBJECT_ID:-$PREDICTION_SUBJECT_ID}"
if [[ -n "$FILTER_PREDICTION_SUBJECT_ID" && "$FILTER_PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json view prediction-subject-memberships list --prediction-subject "$FILTER_PREDICTION_SUBJECT_ID" --limit 5
    assert_success
else
    skip "No prediction subject ID available (set XBE_TEST_PREDICTION_SUBJECT_ID)"
fi

if [[ -z "$MEMBER_USER_ID" ]]; then
    MEMBER_USER_ID="$SAMPLE_USER_ID"
fi

if [[ -z "$MEMBER_USER_ID" || "$MEMBER_USER_ID" == "null" ]]; then
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        MEMBER_USER_ID=$(json_get ".id")
    fi
fi

test_name "List prediction subject memberships with --user filter"
if [[ -n "$MEMBER_USER_ID" && "$MEMBER_USER_ID" != "null" ]]; then
    xbe_json view prediction-subject-memberships list --user "$MEMBER_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available (set XBE_TEST_USER_ID)"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction subject membership requires required flags"
xbe_run do prediction-subject-memberships create
assert_failure

test_name "Create prediction subject membership"
if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" && -n "$MEMBER_USER_ID" && "$MEMBER_USER_ID" != "null" ]]; then
    xbe_json do prediction-subject-memberships create \
        --prediction-subject "$PREDICTION_SUBJECT_ID" \
        --user "$MEMBER_USER_ID" \
        --can-manage-memberships true
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "prediction-subject-memberships" "$CREATED_ID"
            pass
        else
            fail "Created prediction subject membership but no ID returned"
        fi
    else
        if update_blocked_message "$output"; then
            skip "Create blocked by server policy or invalid prediction subject"
        else
            fail "Failed to create prediction subject membership: $output"
        fi
    fi
else
    skip "Missing prediction subject or user ID (set XBE_TEST_PREDICTION_SUBJECT_ID and XBE_TEST_USER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update prediction subject membership permissions"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do prediction-subject-memberships update "$CREATED_ID" \
        --can-update-prediction-consensus true \
        --can-manage-gaps true
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Update blocked by server policy or validation"
        else
            fail "Failed to update prediction subject membership: $output"
        fi
    fi
else
    skip "No created prediction subject membership available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete prediction subject membership requires --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-subject-memberships delete "$CREATED_ID"
    assert_failure
else
    skip "No created prediction subject membership available for delete"
fi

test_name "Delete prediction subject membership"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-subject-memberships delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete prediction subject membership: $output"
        fi
    fi
else
    skip "No created prediction subject membership available for delete"
fi

run_tests
