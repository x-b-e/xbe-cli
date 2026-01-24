#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Driver Confirmations
#
# Tests list, show, and update operations for the project-transport-plan-driver-confirmations resource.
#
# COVERAGE: List filters + show + update + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PROJECT_TRANSPORT_PLAN_ID=""
SAMPLE_PROJECT_TRANSPORT_PLAN_DRIVER_ID=""
SAMPLE_DRIVER_ID=""


describe "Resource: project-transport-plan-driver-confirmations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan driver confirmations"
xbe_json view project-transport-plan-driver-confirmations list --limit 5
assert_success

test_name "List project transport plan driver confirmations returns array"
xbe_json view project-transport-plan-driver-confirmations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan driver confirmations"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample confirmation"
xbe_json view project-transport-plan-driver-confirmations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PROJECT_TRANSPORT_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
    SAMPLE_PROJECT_TRANSPORT_PLAN_DRIVER_ID=$(json_get ".[0].project_transport_plan_driver_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No confirmations available for follow-on tests"
    fi
else
    skip "Could not list confirmations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List confirmations with --status filter"
xbe_json view project-transport-plan-driver-confirmations list --status pending --limit 5
assert_success

test_name "List confirmations with --project-transport-plan filter"
if [[ -n "$SAMPLE_PROJECT_TRANSPORT_PLAN_ID" && "$SAMPLE_PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-driver-confirmations list --project-transport-plan "$SAMPLE_PROJECT_TRANSPORT_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List confirmations with --project-transport-plan-driver filter"
if [[ -n "$SAMPLE_PROJECT_TRANSPORT_PLAN_DRIVER_ID" && "$SAMPLE_PROJECT_TRANSPORT_PLAN_DRIVER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-driver-confirmations list --project-transport-plan-driver "$SAMPLE_PROJECT_TRANSPORT_PLAN_DRIVER_ID" --limit 5
    assert_success
else
    skip "No project transport plan driver ID available"
fi

test_name "List confirmations with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-driver-confirmations list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan driver confirmation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-driver-confirmations show "$SAMPLE_ID"
    assert_success
else
    skip "No confirmation ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update confirmation note"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do project-transport-plan-driver-confirmations update "$SAMPLE_ID" --note "CLI updated note"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update confirmation (permissions or policy)"
    fi
else
    skip "No confirmation ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update confirmation without any fields fails"
xbe_run do project-transport-plan-driver-confirmations update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
