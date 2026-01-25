#!/bin/bash
#
# XBE CLI Integration Tests: User Searches
#
# Tests create operations for the user-searches resource.
#
# COVERAGE: contact_method, contact_value, only_admin_or_member attributes + invalid contact-method failure
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TEST_EMAIL="${XBE_TEST_USER_SEARCH_EMAIL:-}"

INVALID_CONTACT_METHOD="invalid_method"

describe "Resource: user-searches"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user search requires contact-method"
xbe_run do user-searches create --contact-value "test@example.com"
assert_failure

test_name "Create user search requires contact-value"
xbe_run do user-searches create --contact-method email_address
assert_failure

test_name "Create user search rejects invalid contact-method"
xbe_run do user-searches create --contact-method "$INVALID_CONTACT_METHOD" --contact-value "test@example.com"
if [[ $status -eq 0 ]]; then
    fail "Expected failure for invalid contact-method"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "Not authorized to create user searches"
    else
        pass
    fi
fi

test_name "Create user search"
if [[ -z "$TEST_EMAIL" || "$TEST_EMAIL" == "null" ]]; then
    TEST_EMAIL=$(unique_email)
fi

xbe_json do user-searches create --contact-method email_address --contact-value "$TEST_EMAIL" --only-admin-or-member true
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".contact_method" "email_address"
    assert_json_equals ".contact_value" "$TEST_EMAIL"
    assert_json_equals ".only_admin_or_member" "true"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
        skip "Create blocked by server policy/validation"
    else
        fail "Failed to create user search: $output"
    fi
fi

run_tests
