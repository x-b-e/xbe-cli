#!/bin/bash
#
# XBE CLI Integration Tests: Login Code Redemptions
#
# Tests create operations for the login-code-redemptions resource.
#
# COVERAGE: code, device_id attributes + invalid code failure
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

LOGIN_CODE="${XBE_TEST_LOGIN_CODE_REDEMPTION_CODE:-}"
DEVICE_ID="${XBE_TEST_LOGIN_CODE_DEVICE_ID:-}"

INVALID_CODE="invalid-code-$(date +%s)"

describe "Resource: login-code-redemptions"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create login code redemption requires --code"
xbe_run do login-code-redemptions create
assert_failure

test_name "Create login code redemption rejects invalid code"
xbe_run do login-code-redemptions create --code "$INVALID_CODE"
if [[ $status -eq 0 ]]; then
    fail "Expected failure for invalid login code"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "Not authorized to redeem login codes"
    else
        pass
    fi
fi

test_name "Redeem login code"
if [[ -n "$LOGIN_CODE" && "$LOGIN_CODE" != "null" ]]; then
    if [[ -n "$DEVICE_ID" && "$DEVICE_ID" != "null" ]]; then
        xbe_json do login-code-redemptions create --code "$LOGIN_CODE" --device-id "$DEVICE_ID"
    else
        xbe_json do login-code-redemptions create --code "$LOGIN_CODE"
    fi

    if [[ $status -eq 0 ]]; then
        assert_json_has ".auth_token"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not found"* ]] || [[ "$output" == *"not found"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Login code redemption not available (code invalid or expired)"
        else
            fail "Failed to redeem login code: $output"
        fi
    fi
else
    skip "No login code available. Set XBE_TEST_LOGIN_CODE_REDEMPTION_CODE to enable create testing."
fi

run_tests
