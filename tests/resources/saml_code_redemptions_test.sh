#!/bin/bash
#
# XBE CLI Integration Tests: SAML Code Redemptions
#
# Tests create operations for the saml-code-redemptions resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAML_CODE="${XBE_TEST_SAML_CODE:-}"

describe "Resource: saml-code-redemptions"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create SAML code redemption without required fields fails"
xbe_run do saml-code-redemptions create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create SAML code redemption"
if [[ -n "$SAML_CODE" ]]; then
    xbe_json do saml-code-redemptions create --code "$SAML_CODE"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_has ".auth_token"
    else
        fail "Failed to redeem SAML code"
    fi
else
    skip "XBE_TEST_SAML_CODE not set"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
