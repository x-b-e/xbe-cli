#!/bin/bash
#
# XBE CLI Integration Tests: Login Code Requests
#
# Tests create operations for the login-code-requests resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CONTACT_METHOD="${XBE_TEST_LOGIN_CODE_CONTACT_METHOD:-}"
DEVICE_ID="${XBE_TEST_LOGIN_CODE_DEVICE_ID:-}"

describe "Resource: login-code-requests"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create login code request without required fields fails"
xbe_run do login-code-requests create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$CONTACT_METHOD" ]]; then
    CONTACT_METHOD=$(unique_email)
fi

if [[ -z "$DEVICE_ID" ]]; then
    DEVICE_ID="xbe-cli-login-code-$(unique_suffix)"
fi

test_name "Create login code request"
xbe_json do login-code-requests create \
    --contact-method "$CONTACT_METHOD" \
    --device-id "$DEVICE_ID"

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".contact_method" "$CONTACT_METHOD"
    assert_json_equals ".device_id" "$DEVICE_ID"
    assert_json_equals ".result" "sent"
else
    fail "Failed to create login code request"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
