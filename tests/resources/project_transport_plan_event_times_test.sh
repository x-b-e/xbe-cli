#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Event Times
#
# Tests CRUD operations and list filters for the project_transport_plan_event_times resource.
#
# COVERAGE: All create/update attributes + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

EVENT_TIME_ID=""
EVENT_ID=""
PLAN_ID=""
CHANGED_BY_ID=""
KIND=""
START_AT=""
END_AT=""
AT_VALUE=""
SKIP_FILTERS=0
CREATED_ID=""
AVAILABLE_KIND=""
EXISTING_KINDS=""

describe "Resource: project-transport-plan-event-times"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan event times"
xbe_json view project-transport-plan-event-times list --limit 5
assert_success

test_name "List project transport plan event times returns array"
xbe_json view project-transport-plan-event-times list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan event times"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project transport plan event time"
xbe_json view project-transport-plan-event-times list --limit 1
if [[ $status -eq 0 ]]; then
    EVENT_TIME_ID=$(json_get ".[0].id")
    if [[ -n "$EVENT_TIME_ID" && "$EVENT_TIME_ID" != "null" ]]; then
        pass
    else
        SKIP_FILTERS=1
        skip "No project transport plan event times available"
    fi
else
    SKIP_FILTERS=1
    fail "Failed to list project transport plan event times"
fi

if [[ $SKIP_FILTERS -eq 0 ]]; then
    xbe_json view project-transport-plan-event-times show "$EVENT_TIME_ID"
    if [[ $status -eq 0 ]]; then
        EVENT_ID=$(json_get ".project_transport_plan_event_id")
        PLAN_ID=$(json_get ".project_transport_plan_id")
        CHANGED_BY_ID=$(json_get ".changed_by_id")
        KIND=$(json_get ".kind")
        START_AT=$(json_get ".start_at")
        END_AT=$(json_get ".end_at")
        AT_VALUE=$(json_get ".at")
    else
        SKIP_FILTERS=1
        fail "Failed to fetch project transport plan event time details"
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport plan event times with --project-transport-plan-event filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --project-transport-plan-event "$EVENT_ID" --limit 5
    assert_success
else
    skip "No project transport plan event ID available"
fi

test_name "List project transport plan event times with --project-transport-plan filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --project-transport-plan "$PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List project transport plan event times with --project-transport-plan-id filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --project-transport-plan-id "$PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List project transport plan event times with --changed-by filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$CHANGED_BY_ID" && "$CHANGED_BY_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --changed-by "$CHANGED_BY_ID" --limit 5
    assert_success
else
    skip "No changed-by ID available"
fi

test_name "List project transport plan event times with --kind filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$KIND" && "$KIND" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --kind "$KIND" --limit 5
    assert_success
else
    skip "No kind available"
fi

test_name "List project transport plan event times with --start-at-min filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --start-at-min "$START_AT" --limit 5
    assert_success
else
    skip "No start-at timestamp available"
fi

test_name "List project transport plan event times with --start-at-max filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --start-at-max "$START_AT" --limit 5
    assert_success
else
    skip "No start-at timestamp available"
fi

test_name "List project transport plan event times with --end-at-min filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --end-at-min "$END_AT" --limit 5
    assert_success
else
    skip "No end-at timestamp available"
fi

test_name "List project transport plan event times with --end-at-max filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --end-at-max "$END_AT" --limit 5
    assert_success
else
    skip "No end-at timestamp available"
fi

if [[ -z "$AT_VALUE" || "$AT_VALUE" == "null" ]]; then
    AT_VALUE="$START_AT"
fi

test_name "List project transport plan event times with --at-min filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$AT_VALUE" && "$AT_VALUE" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --at-min "$AT_VALUE" --limit 5
    assert_success
else
    skip "No at timestamp available"
fi

test_name "List project transport plan event times with --at-max filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$AT_VALUE" && "$AT_VALUE" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --at-max "$AT_VALUE" --limit 5
    assert_success
else
    skip "No at timestamp available"
fi

test_name "List project transport plan event times with --is-at filter"
xbe_json view project-transport-plan-event-times list --is-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan event time"
if [[ $SKIP_FILTERS -eq 0 && -n "$EVENT_TIME_ID" && "$EVENT_TIME_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-times show "$EVENT_TIME_ID"
    assert_success
else
    skip "No project transport plan event time ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan event time requires --project-transport-plan-event"
xbe_json do project-transport-plan-event-times create --kind planned --start-at "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
assert_failure

test_name "Create project transport plan event time requires --kind"
if [[ -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
    xbe_json do project-transport-plan-event-times create --project-transport-plan-event "$EVENT_ID" --start-at "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    assert_failure
else
    skip "No project transport plan event ID available"
fi

test_name "Create project transport plan event time requires --start-at or --at"
if [[ -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
    xbe_json do project-transport-plan-event-times create --project-transport-plan-event "$EVENT_ID" --kind planned
    assert_failure
else
    skip "No project transport plan event ID available"
fi

test_name "Find available kind for create"
if [[ -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-times list --project-transport-plan-event "$EVENT_ID" --limit 25
    if [[ $status -eq 0 ]]; then
        EXISTING_KINDS=$(echo "$output" | jq -r '.[].kind' | sort -u)
        for candidate in modeled expected planned actual; do
            if ! echo "$EXISTING_KINDS" | grep -qx "$candidate"; then
                AVAILABLE_KIND="$candidate"
                break
            fi
        done
        if [[ -n "$AVAILABLE_KIND" ]]; then
            pass
        else
            skip "No available time kind for create"
        fi
    else
        fail "Failed to list event times for create"
    fi
else
    skip "No project transport plan event ID available"
fi

if [[ -n "$AVAILABLE_KIND" ]]; then
    test_name "Create project transport plan event time"
    START_AT_CREATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    END_AT_CREATE="$START_AT_CREATE"
    xbe_json do project-transport-plan-event-times create \
        --project-transport-plan-event "$EVENT_ID" \
        --kind "$AVAILABLE_KIND" \
        --start-at "$START_AT_CREATE" \
        --end-at "$END_AT_CREATE"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-event-times" "$CREATED_ID"
            pass
        else
            fail "Created project transport plan event time but no ID returned"
        fi
    else
        fail "Failed to create project transport plan event time"
    fi
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update project transport plan event time start-at"
    sleep 1
    UPDATED_START=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do project-transport-plan-event-times update "$CREATED_ID" --start-at "$UPDATED_START" --end-at "$UPDATED_START"
    assert_success

    test_name "Update project transport plan event time end-at"
    sleep 1
    UPDATED_END=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do project-transport-plan-event-times update "$CREATED_ID" --end-at "$UPDATED_END"
    assert_success

    test_name "Update project transport plan event time kind"
    xbe_json do project-transport-plan-event-times update "$CREATED_ID" --kind "$AVAILABLE_KIND"
    assert_success

    test_name "Update project transport plan event time with legacy --at"
    sleep 1
    UPDATED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do project-transport-plan-event-times update "$CREATED_ID" --at "$UPDATED_AT"
    assert_success
else
    skip "No project transport plan event time available for updates"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete project transport plan event time requires --confirm"
    xbe_run do project-transport-plan-event-times delete "$CREATED_ID"
    assert_failure

    test_name "Delete project transport plan event time with --confirm"
    xbe_run do project-transport-plan-event-times delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No project transport plan event time available for deletion"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
