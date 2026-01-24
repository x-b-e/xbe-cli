#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Event Location Prediction Autopsies
#
# Tests list, show, and filter operations for the project-transport-plan-event-location-prediction-autopsies resource.
#
# COVERAGE: List + show + filter + error case
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_EVENT_ID=""
LIST_SUPPORTED="true"

describe "Resource: project-transport-plan-event-location-prediction-autopsies"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan event location prediction autopsies"
xbe_json view project-transport-plan-event-location-prediction-autopsies list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing project transport plan event location prediction autopsies"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project transport plan event location prediction autopsies returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-event-location-prediction-autopsies list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project transport plan event location prediction autopsies"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample project transport plan event location prediction autopsy"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-event-location-prediction-autopsies list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_EVENT_ID=$(json_get ".[0].project_transport_plan_event_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No autopsies available for follow-on tests"
        fi
    else
        skip "Could not list autopsies to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter autopsies by project transport plan event"
if [[ -n "$SAMPLE_EVENT_ID" && "$SAMPLE_EVENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-location-prediction-autopsies list --project-transport-plan-event "$SAMPLE_EVENT_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by project transport plan event failed"
    fi
else
    skip "No project transport plan event ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan event location prediction autopsy"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-location-prediction-autopsies show "$SAMPLE_ID"
    assert_success
else
    skip "No autopsy ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Show autopsy without ID fails"
xbe_run view project-transport-plan-event-location-prediction-autopsies show
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
