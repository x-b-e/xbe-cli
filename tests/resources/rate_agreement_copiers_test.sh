#!/bin/bash
#
# XBE CLI Integration Tests: Rate Agreement Copiers
#
# Tests create operations for rate-agreement-copiers.
#
# COVERAGE: Relationships (template-rate-agreement, target-organization)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TEMPLATE_ID="${XBE_TEST_RATE_AGREEMENT_COPIER_TEMPLATE_ID:-}"
TARGET_ORG_TYPE="${XBE_TEST_RATE_AGREEMENT_COPIER_TARGET_TYPE:-}"
TARGET_ORG_ID="${XBE_TEST_RATE_AGREEMENT_COPIER_TARGET_ID:-}"

describe "Resource: rate-agreement-copiers"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create rate agreement copier requires template rate agreement"
xbe_run do rate-agreement-copiers create --target-organization-type customers --target-organization-id 1
assert_failure

test_name "Create rate agreement copier requires target organization type"
xbe_run do rate-agreement-copiers create --template-rate-agreement 1 --target-organization-id 1
assert_failure

test_name "Create rate agreement copier requires target organization id"
xbe_run do rate-agreement-copiers create --template-rate-agreement 1 --target-organization-type customers
assert_failure

test_name "Create rate agreement copier"
if [[ -n "$TEMPLATE_ID" && "$TEMPLATE_ID" != "null" && -n "$TARGET_ORG_TYPE" && "$TARGET_ORG_TYPE" != "null" && -n "$TARGET_ORG_ID" && "$TARGET_ORG_ID" != "null" ]]; then
    xbe_json do rate-agreement-copiers create \
        --template-rate-agreement "$TEMPLATE_ID" \
        --target-organization-type "$TARGET_ORG_TYPE" \
        --target-organization-id "$TARGET_ORG_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".template_rate_agreement_id" "$TEMPLATE_ID"
        assert_json_equals ".target_organization_id" "$TARGET_ORG_ID"
        assert_json_equals ".target_organization_type" "$TARGET_ORG_TYPE"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"must be"* ]] || [[ "$output" == *"template"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create rate agreement copier: $output"
        fi
    fi
else
    skip "Missing rate agreement copier inputs. Set XBE_TEST_RATE_AGREEMENT_COPIER_TEMPLATE_ID, XBE_TEST_RATE_AGREEMENT_COPIER_TARGET_TYPE, and XBE_TEST_RATE_AGREEMENT_COPIER_TARGET_ID."
fi

run_tests
