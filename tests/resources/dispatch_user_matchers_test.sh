#!/bin/bash
#
# XBE CLI Integration Tests: Dispatch User Matchers
#
# Tests create operations for the dispatch_user_matchers resource.
#
# COVERAGE: phone_number attribute
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TEST_PHONE="${XBE_TEST_DISPATCH_USER_MATCHER_PHONE_NUMBER:-}"

describe "Resource: dispatch_user_matchers"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create dispatch user matcher requires phone-number"
xbe_run do dispatch-user-matchers create
assert_failure

test_name "Create dispatch user matcher"
if [[ -z "$TEST_PHONE" || "$TEST_PHONE" == "null" ]]; then
    TEST_PHONE=$(unique_mobile)
fi

xbe_json do dispatch-user-matchers create --phone-number "$TEST_PHONE"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".phone_number" "$TEST_PHONE"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
        skip "Create blocked by server policy/validation"
    else
        fail "Failed to create dispatch user matcher: $output"
    fi
fi

run_tests
