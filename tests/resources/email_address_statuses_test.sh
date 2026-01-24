#!/bin/bash
#
# XBE CLI Integration Tests: Email Address Statuses
#
# Tests create behavior for email address statuses.
#
# COVERAGE: Create + required flag
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: email-address-statuses"

test_name "Create email address status requires --email-address"
xbe_run do email-address-statuses create
assert_failure

test_name "Create email address status"
xbe_json do email-address-statuses create --email-address "test@example.com"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_has ".email_address"
else
    fail "Failed to create email address status"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
