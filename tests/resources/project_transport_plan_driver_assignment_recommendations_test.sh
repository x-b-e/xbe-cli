#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Driver Assignment Recommendations
#
# Tests view and create behavior for project_transport_plan_driver_assignment_recommendations.
#
# COVERAGE: List + list filter + create + show + required flag failure
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-transport-plan-driver-assignment-recommendations"

PLAN_DRIVER_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_DRIVER_ID:-}"
CREATED_RECOMMENDATION_ID=""

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List project transport plan driver assignment recommendations"
xbe_json view project-transport-plan-driver-assignment-recommendations list --limit 5
assert_success

test_name "List project transport plan driver assignment recommendations returns array"
xbe_json view project-transport-plan-driver-assignment-recommendations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan driver assignment recommendations"
fi

if [[ -n "$PLAN_DRIVER_ID" ]]; then
    test_name "List recommendations filtered by project transport plan driver"
    xbe_json view project-transport-plan-driver-assignment-recommendations list \
        --project-transport-plan-driver "$PLAN_DRIVER_ID" \
        --limit 10
    assert_success
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_DRIVER_ID to test filter"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create recommendation without project transport plan driver fails"
xbe_run do project-transport-plan-driver-assignment-recommendations create
assert_failure

if [[ -n "$PLAN_DRIVER_ID" ]]; then
    test_name "Create project transport plan driver assignment recommendation"
    xbe_json do project-transport-plan-driver-assignment-recommendations create \
        --project-transport-plan-driver "$PLAN_DRIVER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_RECOMMENDATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_RECOMMENDATION_ID" && "$CREATED_RECOMMENDATION_ID" != "null" ]]; then
            pass
        else
            fail "Created recommendation but no ID returned"
        fi
    else
        fail "Failed to create project transport plan driver assignment recommendation"
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_DRIVER_ID to run create/show tests"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_RECOMMENDATION_ID" && "$CREATED_RECOMMENDATION_ID" != "null" ]]; then
    test_name "Show project transport plan driver assignment recommendation"
    xbe_json view project-transport-plan-driver-assignment-recommendations show "$CREATED_RECOMMENDATION_ID"
    assert_success
else
    skip "No recommendation created; skipping show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
