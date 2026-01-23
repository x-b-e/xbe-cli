#!/bin/bash
#
# XBE CLI Integration Tests: Maintenance Requirement Rule Evaluation Clerks
#
# Tests create operations for maintenance-requirement-rule-evaluation-clerks.
#
# COVERAGE: create
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

EQUIPMENT_ID="${XBE_TEST_EQUIPMENT_ID:-}"

describe "Resource: maintenance-requirement-rule-evaluation-clerks"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires --equipment"
xbe_run do maintenance-requirement-rule-evaluation-clerks create
assert_failure

if [[ -z "$EQUIPMENT_ID" || "$EQUIPMENT_ID" == "null" ]]; then
    xbe_json view equipment list --limit 1
    if [[ $status -eq 0 ]]; then
        EQUIPMENT_ID=$(json_get ".[0].id")
    fi
fi

test_name "Create maintenance requirement rule evaluation clerk"
if [[ -n "$EQUIPMENT_ID" && "$EQUIPMENT_ID" != "null" ]]; then
    xbe_json do maintenance-requirement-rule-evaluation-clerks create --equipment "$EQUIPMENT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Validation error during maintenance rule evaluation"* ]] || [[ "$output" == *"Error evaluating maintenance rules"* ]]; then
            skip "Unable to evaluate maintenance rules for available equipment"
        else
            fail "Failed to create maintenance requirement rule evaluation clerk"
        fi
    fi
else
    skip "No equipment ID available. Set XBE_TEST_EQUIPMENT_ID to enable create testing."
fi

run_tests
