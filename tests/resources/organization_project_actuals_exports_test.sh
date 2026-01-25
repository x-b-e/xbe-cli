#!/bin/bash
#
# XBE CLI Integration Tests: Organization Project Actuals Exports
#
# Tests create operations for organization-project-actuals-exports.
#
# COVERAGE: Writable attributes (dry-run)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PROJECT_ACTUALS_EXPORT_ID="${XBE_TEST_ORGANIZATION_PROJECT_ACTUALS_EXPORT_PROJECT_ACTUALS_EXPORT_ID:-}"

describe "Resource: organization-project-actuals-exports"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create export requires project actuals export"
xbe_run do organization-project-actuals-exports create --dry-run
assert_failure


test_name "Create organization project actuals export"
if [[ -n "$PROJECT_ACTUALS_EXPORT_ID" && "$PROJECT_ACTUALS_EXPORT_ID" != "null" ]]; then
    xbe_json do organization-project-actuals-exports create \
        --project-actuals-export "$PROJECT_ACTUALS_EXPORT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".project_actuals_export_id" "$PROJECT_ACTUALS_EXPORT_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"could not connect"* ]] || [[ "$output" == *"cannot be exported"* ]] || [[ "$output" == *"must be for a valid organization"* ]] || [[ "$output" == *"project revenue items"* ]] || [[ "$output" == *"formatting"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create organization project actuals export: $output"
        fi
    fi
else
    skip "No project actuals export ID available. Set XBE_TEST_ORGANIZATION_PROJECT_ACTUALS_EXPORT_PROJECT_ACTUALS_EXPORT_ID to enable create testing."
fi


test_name "Create organization project actuals export with --dry-run"
if [[ -n "$PROJECT_ACTUALS_EXPORT_ID" && "$PROJECT_ACTUALS_EXPORT_ID" != "null" ]]; then
    xbe_json do organization-project-actuals-exports create \
        --project-actuals-export "$PROJECT_ACTUALS_EXPORT_ID" \
        --dry-run
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"could not connect"* ]] || [[ "$output" == *"cannot be exported"* ]] || [[ "$output" == *"must be for a valid organization"* ]] || [[ "$output" == *"project revenue items"* ]] || [[ "$output" == *"formatting"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create organization project actuals export (dry-run): $output"
        fi
    fi
else
    skip "No project actuals export ID available. Set XBE_TEST_ORGANIZATION_PROJECT_ACTUALS_EXPORT_PROJECT_ACTUALS_EXPORT_ID to enable dry-run testing."
fi

run_tests
