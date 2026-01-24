#!/bin/bash
#
# XBE CLI Integration Tests: User Auth Token Resets
#
# Tests create operations for the user-auth-token-resets resource.
#
# COVERAGE: user_id attribute + missing user ID + invalid user ID failure
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

RESET_USER_ID="${XBE_TEST_USER_AUTH_TOKEN_RESET_USER_ID:-}"
WHOAMI_USER_ID=""
TARGET_USER_ID=""
CREATED_USER="false"

INVALID_USER_ID="invalid-user-$(date +%s)"

describe "Resource: user-auth-token-resets"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user auth token reset requires --user-id"
xbe_run do user-auth-token-resets create
assert_failure

test_name "Create user auth token reset rejects invalid user ID"
xbe_run do user-auth-token-resets create --user-id "$INVALID_USER_ID"
if [[ $status -eq 0 ]]; then
    fail "Expected failure for invalid user ID"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "Not authorized to reset auth tokens"
    else
        pass
    fi
fi

test_name "Resolve current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    pass
else
    skip "Unable to resolve current user"
fi

test_name "Create user for auth token reset"
if [[ -n "$RESET_USER_ID" && "$RESET_USER_ID" != "null" ]]; then
    TARGET_USER_ID="$RESET_USER_ID"
    skip "Using XBE_TEST_USER_AUTH_TOKEN_RESET_USER_ID"
else
    TEST_USER_NAME=$(unique_name "AuthResetUser")
    TEST_USER_EMAIL=$(unique_email)
    xbe_json do users create --name "$TEST_USER_NAME" --email "$TEST_USER_EMAIL"
    if [[ $status -eq 0 ]]; then
        TARGET_USER_ID=$(json_get ".id")
        if [[ -n "$TARGET_USER_ID" && "$TARGET_USER_ID" != "null" ]]; then
            CREATED_USER="true"
            pass
        else
            fail "Created user but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
            skip "Not authorized to create users; set XBE_TEST_USER_AUTH_TOKEN_RESET_USER_ID"
        else
            fail "Failed to create user for auth token reset: $output"
        fi
    fi
fi

test_name "Reset user auth token"
if [[ -n "$TARGET_USER_ID" && "$TARGET_USER_ID" != "null" ]]; then
    if [[ -n "$WHOAMI_USER_ID" && "$TARGET_USER_ID" == "$WHOAMI_USER_ID" ]]; then
        skip "Target user matches current user; set XBE_TEST_USER_AUTH_TOKEN_RESET_USER_ID"
    else
        xbe_json do user-auth-token-resets create --user-id "$TARGET_USER_ID"
        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
            assert_json_bool ".is_reset" "true"
        else
            if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
                skip "Not authorized to reset auth tokens"
            elif [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]] || [[ "$output" == *"404"* ]]; then
                if [[ "$CREATED_USER" == "true" ]]; then
                    fail "Created user not found during reset"
                else
                    skip "User not found for reset"
                fi
            else
                fail "Failed to reset auth token: $output"
            fi
        fi
    fi
else
    skip "No user ID available. Set XBE_TEST_USER_AUTH_TOKEN_RESET_USER_ID to enable reset testing."
fi

run_tests
