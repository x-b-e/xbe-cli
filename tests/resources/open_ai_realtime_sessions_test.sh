#!/bin/bash
#
# XBE CLI Integration Tests: OpenAI Realtime Sessions
#
# Tests create, list filters, and show operations for the open_ai_realtime_sessions resource.
#
# COVERAGE: All create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SESSION_ID=""
CREATED_USER_ID=""

CLIENT_FEATURE_REQUIRED="development"
CLIENT_FEATURE_ALT="slack_realtime_chat"

SESSION_EXPIRES_AT=""

# Use a wide date range for time filters
RANGE_MIN="2000-01-01T00:00:00Z"
RANGE_MAX="2100-01-01T00:00:00Z"

# Create a test user for user filter
create_test_user() {
    local name
    local email
    name=$(unique_name "RealtimeSessionUser")
    email=$(unique_email)

    xbe_json do users create --name "$name" --email "$email"
    if [[ $status -eq 0 ]]; then
        CREATED_USER_ID=$(json_get ".id")
        if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
            register_cleanup "users" "$CREATED_USER_ID"
            return 0
        fi
    fi
    return 1
}

describe "Resource: open-ai-realtime-sessions"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite user"
if create_test_user; then
    pass
else
    fail "Failed to create user for open-ai-realtime-sessions"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create OpenAI realtime session with required fields"
xbe_json do open-ai-realtime-sessions create --client-feature "$CLIENT_FEATURE_REQUIRED"

if [[ $status -eq 0 ]]; then
    CREATED_SESSION_ID=$(json_get ".id")
    SESSION_EXPIRES_AT=$(json_get ".client_secret_expires_at")
    if [[ -n "$CREATED_SESSION_ID" && "$CREATED_SESSION_ID" != "null" ]]; then
        pass
    else
        fail "Created session but no ID returned"
    fi
else
    fail "Failed to create OpenAI realtime session: $output"
fi

# Only continue if we successfully created a session
if [[ -z "$CREATED_SESSION_ID" || "$CREATED_SESSION_ID" == "null" ]]; then
    echo "Cannot continue without a valid session ID"
    run_tests
fi

test_name "Create OpenAI realtime session with model"
xbe_json do open-ai-realtime-sessions create \
    --client-feature "$CLIENT_FEATURE_REQUIRED" \
    --model "gpt-4o-realtime-preview"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create session with model"
fi

test_name "Create OpenAI realtime session with user"
xbe_json do open-ai-realtime-sessions create \
    --client-feature "$CLIENT_FEATURE_ALT" \
    --user "$CREATED_USER_ID"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create session with user"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List OpenAI realtime sessions"
xbe_json view open-ai-realtime-sessions list
assert_success

test_name "List OpenAI realtime sessions returns array"
xbe_json view open-ai-realtime-sessions list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list OpenAI realtime sessions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List sessions with --client-feature filter"
xbe_json view open-ai-realtime-sessions list --client-feature "$CLIENT_FEATURE_REQUIRED"
assert_success

test_name "List sessions with --user filter"
xbe_json view open-ai-realtime-sessions list --user "$CREATED_USER_ID"
assert_success

test_name "List sessions with --client-secret-expires-at-min filter"
xbe_json view open-ai-realtime-sessions list --client-secret-expires-at-min "$RANGE_MIN"
assert_success

test_name "List sessions with --client-secret-expires-at-max filter"
xbe_json view open-ai-realtime-sessions list --client-secret-expires-at-max "$RANGE_MAX"
assert_success

test_name "List sessions with --is-client-secret-expires-at filter"
xbe_json view open-ai-realtime-sessions list --is-client-secret-expires-at true
assert_success

test_name "List sessions with --created-at-min filter"
xbe_json view open-ai-realtime-sessions list --created-at-min "$RANGE_MIN"
assert_success

test_name "List sessions with --created-at-max filter"
xbe_json view open-ai-realtime-sessions list --created-at-max "$RANGE_MAX"
assert_success

test_name "List sessions with --is-created-at filter"
xbe_json view open-ai-realtime-sessions list --is-created-at true
assert_success

test_name "List sessions with --updated-at-min filter"
xbe_json view open-ai-realtime-sessions list --updated-at-min "$RANGE_MIN"
assert_success

test_name "List sessions with --updated-at-max filter"
xbe_json view open-ai-realtime-sessions list --updated-at-max "$RANGE_MAX"
assert_success

test_name "List sessions with --is-updated-at filter"
xbe_json view open-ai-realtime-sessions list --is-updated-at true
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show OpenAI realtime session details"
xbe_json view open-ai-realtime-sessions show "$CREATED_SESSION_ID"

if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$CREATED_SESSION_ID"
else
    fail "Failed to show OpenAI realtime session"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create OpenAI realtime session without client feature fails"
xbe_json do open-ai-realtime-sessions create
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
