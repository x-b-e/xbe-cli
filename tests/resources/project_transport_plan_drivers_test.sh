#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Drivers
#
# Tests list, show, create, update, and delete operations for project_transport_plan_drivers.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_PTP_DRIVER_ID=""
PROJECT_TRANSPORT_PLAN_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_ID:-}"
PROJECT_TRANSPORT_PLAN_STATUS=""
ASSIGNMENT_STATUS=""
SEGMENT_START_ID="${XBE_TEST_PTP_SEGMENT_START_ID:-}"
SEGMENT_END_ID="${XBE_TEST_PTP_SEGMENT_END_ID:-}"
DRIVER_ID="${XBE_TEST_DRIVER_ID:-}"
INBOUND_PROJECT_OFFICE_ID="${XBE_TEST_PROJECT_OFFICE_ID:-}"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
WINDOW_START_AT_CACHED=""
WINDOW_END_AT_CACHED=""
WINDOW_START_DATE=""
WINDOW_END_DATE=""
CREATED_PTP_DRIVER_ID=""

describe "Resource: project-transport-plan-drivers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan drivers"
xbe_json view project-transport-plan-drivers list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_PTP_DRIVER_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$PROJECT_TRANSPORT_PLAN_ID" || "$PROJECT_TRANSPORT_PLAN_ID" == "null" ]]; then
            PROJECT_TRANSPORT_PLAN_ID=$(echo "$output" | jq -r '.[0].project_transport_plan_id')
        fi
        PROJECT_TRANSPORT_PLAN_STATUS=$(echo "$output" | jq -r '.[0].project_transport_plan_status')
        ASSIGNMENT_STATUS=$(echo "$output" | jq -r '.[0].status')
        if [[ -z "$SEGMENT_START_ID" || "$SEGMENT_START_ID" == "null" ]]; then
            SEGMENT_START_ID=$(echo "$output" | jq -r '.[0].segment_start_id')
        fi
        if [[ -z "$SEGMENT_END_ID" || "$SEGMENT_END_ID" == "null" ]]; then
            SEGMENT_END_ID=$(echo "$output" | jq -r '.[0].segment_end_id')
        fi
        if [[ -z "$DRIVER_ID" || "$DRIVER_ID" == "null" ]]; then
            DRIVER_ID=$(echo "$output" | jq -r '.[0].driver_id')
        fi
        if [[ -z "$INBOUND_PROJECT_OFFICE_ID" || "$INBOUND_PROJECT_OFFICE_ID" == "null" ]]; then
            INBOUND_PROJECT_OFFICE_ID=$(echo "$output" | jq -r '.[0].inbound_project_office_id')
        fi
        if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
            TRUCKER_ID=$(echo "$output" | jq -r '.[0].segment_start_trucker_id')
        fi
        if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
            BROKER_ID=$(echo "$output" | jq -r '.[0].broker_id')
        fi
        WINDOW_START_AT_CACHED=$(echo "$output" | jq -r '.[0].window_start_at_cached')
        WINDOW_END_AT_CACHED=$(echo "$output" | jq -r '.[0].window_end_at_cached')
        if [[ -n "$WINDOW_START_AT_CACHED" && "$WINDOW_START_AT_CACHED" != "null" ]]; then
            WINDOW_START_DATE="${WINDOW_START_AT_CACHED:0:10}"
        fi
        if [[ -n "$WINDOW_END_AT_CACHED" && "$WINDOW_END_AT_CACHED" != "null" ]]; then
            WINDOW_END_DATE="${WINDOW_END_AT_CACHED:0:10}"
        fi
    fi
else
    fail "Failed to list project transport plan drivers"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan driver"
if [[ -n "$SEED_PTP_DRIVER_ID" && "$SEED_PTP_DRIVER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers show "$SEED_PTP_DRIVER_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        pass
    else
        fail "Failed to show project transport plan driver"
    fi
else
    skip "No project transport plan driver available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan driver"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" && -n "$SEGMENT_START_ID" && "$SEGMENT_START_ID" != "null" && -n "$SEGMENT_END_ID" && "$SEGMENT_END_ID" != "null" ]]; then
    CONFIRM_AT_MAX=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do project-transport-plan-drivers create \
        --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" \
        --segment-start "$SEGMENT_START_ID" \
        --segment-end "$SEGMENT_END_ID" \
        --status editing \
        --confirm-note "CLI test confirm" \
        --confirm-at-max "$CONFIRM_AT_MAX" \
        --skip-assignment-rules-validation \
        --assignment-rule-override-reason "CLI test override"
    if [[ $status -eq 0 ]]; then
        CREATED_PTP_DRIVER_ID=$(json_get ".id")
        if [[ -n "$CREATED_PTP_DRIVER_ID" && "$CREATED_PTP_DRIVER_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-drivers" "$CREATED_PTP_DRIVER_ID"
            pass
        else
            fail "Created project transport plan driver but no ID returned"
        fi
    else
        fail "Failed to create project transport plan driver"
    fi
else
    skip "Missing project transport plan or segment IDs (set XBE_TEST_PROJECT_TRANSPORT_PLAN_ID, XBE_TEST_PTP_SEGMENT_START_ID, XBE_TEST_PTP_SEGMENT_END_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport plan driver attributes"
if [[ -n "$CREATED_PTP_DRIVER_ID" && "$CREATED_PTP_DRIVER_ID" != "null" ]]; then
    UPDATED_CONFIRM_AT_MAX=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do project-transport-plan-drivers update "$CREATED_PTP_DRIVER_ID" \
        --status editing \
        --confirm-note "Updated confirm note" \
        --confirm-at-max "$UPDATED_CONFIRM_AT_MAX" \
        --assignment-rule-override-reason "Updated override reason" \
        --skip-assignment-rules-validation=false \
        --segment-start "$SEGMENT_START_ID" \
        --segment-end "$SEGMENT_END_ID" \
        --driver "" \
        --inbound-project-office-explicit ""
    assert_success
else
    skip "No created project transport plan driver to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan driver requires --confirm flag"
if [[ -n "$CREATED_PTP_DRIVER_ID" && "$CREATED_PTP_DRIVER_ID" != "null" ]]; then
    xbe_run do project-transport-plan-drivers delete "$CREATED_PTP_DRIVER_ID"
    assert_failure
else
    skip "No created project transport plan driver to delete"
fi

test_name "Delete project transport plan driver"
if [[ -n "$CREATED_PTP_DRIVER_ID" && "$CREATED_PTP_DRIVER_ID" != "null" ]]; then
    xbe_run do project-transport-plan-drivers delete "$CREATED_PTP_DRIVER_ID" --confirm
    assert_success
else
    skip "No created project transport plan driver to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by project transport plan"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available for filter"
fi

test_name "Filter by project transport plan status"
if [[ -n "$PROJECT_TRANSPORT_PLAN_STATUS" && "$PROJECT_TRANSPORT_PLAN_STATUS" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --project-transport-plan-status "$PROJECT_TRANSPORT_PLAN_STATUS" --limit 5
    assert_success
else
    skip "No project transport plan status available for filter"
fi

test_name "Filter by driver"
if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --driver "$DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available for filter"
fi

test_name "Filter by has-driver"
xbe_json view project-transport-plan-drivers list --has-driver true --limit 5
assert_success

test_name "Filter by segment start"
if [[ -n "$SEGMENT_START_ID" && "$SEGMENT_START_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --segment-start "$SEGMENT_START_ID" --limit 5
    assert_success
else
    skip "No segment start ID available for filter"
fi

test_name "Filter by segment end"
if [[ -n "$SEGMENT_END_ID" && "$SEGMENT_END_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --segment-end "$SEGMENT_END_ID" --limit 5
    assert_success
else
    skip "No segment end ID available for filter"
fi

test_name "Filter by status"
if [[ -n "$ASSIGNMENT_STATUS" && "$ASSIGNMENT_STATUS" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --status "$ASSIGNMENT_STATUS" --limit 5
    assert_success
else
    skip "No status available for filter"
fi

test_name "Filter by window start date"
if [[ -n "$WINDOW_START_DATE" && "$WINDOW_START_DATE" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --window-start-at-cached "$WINDOW_START_DATE" --limit 5
    assert_success
else
    skip "No window start date available for filter"
fi

test_name "Filter by window start min/max"
if [[ -n "$WINDOW_START_DATE" && "$WINDOW_START_DATE" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --window-start-at-cached-min "$WINDOW_START_DATE" --window-start-at-cached-max "$WINDOW_START_DATE" --limit 5
    assert_success
else
    skip "No window start date available for filter"
fi

test_name "Filter by has-window-start-at-cached"
if [[ -n "$WINDOW_START_DATE" && "$WINDOW_START_DATE" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --has-window-start-at-cached true --limit 5
    assert_success
else
    skip "No window start date available for filter"
fi

test_name "Filter by window end date"
if [[ -n "$WINDOW_END_DATE" && "$WINDOW_END_DATE" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --window-end-at-cached "$WINDOW_END_DATE" --limit 5
    assert_success
else
    skip "No window end date available for filter"
fi

test_name "Filter by window end min/max"
if [[ -n "$WINDOW_END_DATE" && "$WINDOW_END_DATE" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --window-end-at-cached-min "$WINDOW_END_DATE" --window-end-at-cached-max "$WINDOW_END_DATE" --limit 5
    assert_success
else
    skip "No window end date available for filter"
fi

test_name "Filter by has-window-end-at-cached"
if [[ -n "$WINDOW_END_DATE" && "$WINDOW_END_DATE" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --has-window-end-at-cached true --limit 5
    assert_success
else
    skip "No window end date available for filter"
fi

test_name "Filter by trucker"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available for filter"
fi

test_name "Filter by inbound project office"
if [[ -n "$INBOUND_PROJECT_OFFICE_ID" && "$INBOUND_PROJECT_OFFICE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --inbound-project-office "$INBOUND_PROJECT_OFFICE_ID" --limit 5
    assert_success
else
    skip "No inbound project office ID available for filter"
fi

test_name "Filter by most recent"
xbe_json view project-transport-plan-drivers list --most-recent true --limit 5
assert_success

test_name "Filter by managed or none transport order"
xbe_json view project-transport-plan-drivers list --has-managed-transport-order-or-no-transport-order true --limit 5
assert_success

test_name "Filter by broker"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-drivers list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

run_tests
