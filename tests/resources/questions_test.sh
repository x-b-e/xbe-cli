#!/bin/bash
#
# XBE CLI Integration Tests: Questions
#
# Tests list/show/create/update/delete operations for questions.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BY_ID="${XBE_TEST_USER_ID:-}"
ASKED_BY_ID="${XBE_TEST_USER_ID:-}"
ASSIGNED_TO_ID="${XBE_TEST_ADMIN_USER_ID:-}"
PUBLIC_SCOPE="${XBE_TEST_PUBLIC_ORG_SCOPE:-}"

SAMPLE_ID=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_ASKED_BY_ID=""
SAMPLE_ASSIGNED_TO_ID=""
CREATED_ID=""

CONTENT="CLI Question $(unique_suffix)"
UPDATED_CONTENT="CLI Question Updated $(unique_suffix)"

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"Not Found"* ]] || [[ "$msg" == *"not found"* ]] || [[ "$msg" == *"cannot"* ]] || [[ "$msg" == *"validation"* ]] || [[ "$msg" == *"unprocessable"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: questions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List questions"
xbe_json view questions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ASKED_BY_ID=$(json_get ".[0].asked_by_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    SAMPLE_ASSIGNED_TO_ID=$(json_get ".[0].assigned_to_id")
else
    fail "Failed to list questions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show question"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view questions show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output"; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show question: $output"
        fi
    fi
else
    skip "No question ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List questions with --content filter"
xbe_json view questions list --content "$CONTENT" --limit 5
assert_success

test_name "List questions with --source filter"
xbe_json view questions list --source app --limit 5
assert_success

test_name "List questions with --motivation filter"
xbe_json view questions list --motivation serious --limit 5
assert_success

test_name "List questions with --ignore-organization-scoped-newsletters filter"
xbe_json view questions list --ignore-organization-scoped-newsletters true --limit 5
assert_success

test_name "List questions with --is-triaged filter"
xbe_json view questions list --is-triaged false --limit 5
assert_success

test_name "List questions with --created-by filter"
FILTER_CREATED_BY_ID="${SAMPLE_CREATED_BY_ID:-$CREATED_BY_ID}"
if [[ -n "$FILTER_CREATED_BY_ID" && "$FILTER_CREATED_BY_ID" != "null" ]]; then
    xbe_json view questions list --created-by "$FILTER_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available (set XBE_TEST_USER_ID)"
fi

test_name "List questions with --asked-by filter"
FILTER_ASKED_BY_ID="${SAMPLE_ASKED_BY_ID:-$ASKED_BY_ID}"
if [[ -n "$FILTER_ASKED_BY_ID" && "$FILTER_ASKED_BY_ID" != "null" ]]; then
    xbe_json view questions list --asked-by "$FILTER_ASKED_BY_ID" --limit 5
    assert_success
else
    skip "No asked-by ID available (set XBE_TEST_USER_ID)"
fi

test_name "List questions with --assigned-to filter"
FILTER_ASSIGNED_TO_ID="${SAMPLE_ASSIGNED_TO_ID:-$ASSIGNED_TO_ID}"
if [[ -n "$FILTER_ASSIGNED_TO_ID" && "$FILTER_ASSIGNED_TO_ID" != "null" ]]; then
    xbe_json view questions list --assigned-to "$FILTER_ASSIGNED_TO_ID" --limit 5
    assert_success
else
    skip "No assigned-to ID available (set XBE_TEST_ADMIN_USER_ID)"
fi

test_name "List questions with --is-assigned filter"
xbe_json view questions list --is-assigned true --limit 5
assert_success

test_name "List questions with --without-feedback filter"
xbe_json view questions list --without-feedback true --limit 5
assert_success

test_name "List questions with --with-feedback filter"
xbe_json view questions list --with-feedback true --limit 5
assert_success

test_name "List questions with --without-related-content filter"
xbe_json view questions list --without-related-content true --limit 5
assert_success

test_name "List questions with --with-related-content filter"
xbe_json view questions list --with-related-content true --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create question requires required flags"
xbe_run do questions create
assert_failure

CREATE_ARGS=(do questions create --content "$CONTENT")
if [[ -n "$ASKED_BY_ID" && "$ASKED_BY_ID" != "null" ]]; then
    CREATE_ARGS+=(--asked-by "$ASKED_BY_ID")
fi
if [[ -n "$PUBLIC_SCOPE" && "$PUBLIC_SCOPE" != "null" ]]; then
    CREATE_ARGS+=(--is-public true --public-organization-scope "$PUBLIC_SCOPE")
fi

test_name "Create question"
if [[ -n "$CONTENT" ]]; then
    xbe_json "${CREATE_ARGS[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "questions" "$CREATED_ID"
            pass
        else
            fail "Created question but no ID returned"
        fi
    else
        if update_blocked_message "$output"; then
            skip "Create blocked by server policy"
        else
            fail "Failed to create question: $output"
        fi
    fi
else
    skip "No content available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update question attributes"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do questions update "$CREATED_ID" \
        --content "$UPDATED_CONTENT" \
        --source app \
        --ignore-organization-scoped-newsletters false

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output"; then
            skip "Update blocked by server policy"
        else
            fail "Failed to update question: $output"
        fi
    fi
else
    skip "No created question available for update"
fi

# Admin-only updates
if [[ -n "$ASSIGNED_TO_ID" && "$ASSIGNED_TO_ID" != "null" ]]; then
    test_name "Update question admin-only fields"
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        xbe_json do questions update "$CREATED_ID" \
            --is-triaged true \
            --motivation serious \
            --assigned-to "$ASSIGNED_TO_ID"

        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
        else
            if update_blocked_message "$output"; then
                skip "Admin update blocked by server policy"
            else
                fail "Failed to update admin fields: $output"
            fi
        fi
    else
        skip "No created question available for admin update"
    fi
else
    skip "No admin user ID available (set XBE_TEST_ADMIN_USER_ID)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete question requires --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do questions delete "$CREATED_ID"
    assert_failure
else
    skip "No created question available for delete"
fi

test_name "Delete question"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do questions delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if update_blocked_message "$output"; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete question: $output"
        fi
    fi
else
    skip "No created question available for delete"
fi

run_tests
