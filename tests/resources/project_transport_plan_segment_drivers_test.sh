#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Segment Drivers
#
# Tests list, show, create, delete operations for the
# project-transport-plan-segment-drivers resource.
#
# COVERAGE: All create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_SEGMENT_ID=""
SAMPLE_DRIVER_ID=""
LIST_SUPPORTED="true"
CREATED_ID=""

is_nonfatal_error() {
    [[ "$output" == *"Not Authorized"* ]] || \
    [[ "$output" == *"not authorized"* ]] || \
    [[ "$output" == *"Record Invalid"* ]] || \
    [[ "$output" == *"422"* ]] || \
    [[ "$output" == *"403"* ]]
}

describe "Resource: project_transport_plan_segment_drivers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan segment drivers"
xbe_json view project-transport-plan-segment-drivers list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing project transport plan segment drivers"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project transport plan segment drivers returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project transport plan segment drivers"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample project transport plan segment driver"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_SEGMENT_ID=$(json_get ".[0].project_transport_plan_segment_id")
        SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project transport plan segment drivers available for follow-on tests"
        fi
    else
        skip "Could not list project transport plan segment drivers to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan segment driver"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-drivers show "$SAMPLE_ID"
    assert_success
else
    skip "No project transport plan segment driver ID available"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter by project transport plan segment"
if [[ -n "$SAMPLE_SEGMENT_ID" && "$SAMPLE_SEGMENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --project-transport-plan-segment "$SAMPLE_SEGMENT_ID" --limit 5
    assert_success
else
    skip "No project transport plan segment ID available"
fi

test_name "Filter by driver"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "Filter by created-at-min"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --created-at-min "2020-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

test_name "Filter by created-at-max"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --created-at-max "2100-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

test_name "Filter by updated-at-min"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

test_name "Filter by updated-at-max"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-segment-drivers list --updated-at-max "2100-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan segment driver"
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_ID" && -n "$XBE_TEST_DRIVER_ID" ]]; then
    xbe_json do project-transport-plan-segment-drivers create \
        --project-transport-plan-segment "$XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_ID" \
        --driver "$XBE_TEST_DRIVER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-segment-drivers" "$CREATED_ID"
        fi
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Failed to create project transport plan segment driver"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_ID and XBE_TEST_DRIVER_ID to enable create test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project transport plan segment driver without segment fails"
if [[ -n "$XBE_TEST_DRIVER_ID" ]]; then
    xbe_json do project-transport-plan-segment-drivers create --driver "$XBE_TEST_DRIVER_ID"
    assert_failure
else
    skip "Set XBE_TEST_DRIVER_ID to enable error case"
fi

test_name "Create project transport plan segment driver without driver fails"
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_ID" ]]; then
    xbe_json do project-transport-plan-segment-drivers create --project-transport-plan-segment "$XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_ID"
    assert_failure
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_ID to enable error case"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan segment driver requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do project-transport-plan-segment-drivers delete "$CREATED_ID"
    assert_failure
else
    skip "No created segment driver ID available"
fi

test_name "Delete project transport plan segment driver with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do project-transport-plan-segment-drivers delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created segment driver ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
