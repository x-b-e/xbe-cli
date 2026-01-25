#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Trailers
#
# Tests view/create/update/delete behavior for project-transport-plan-trailers.
#
# COVERAGE: List + list filters + show + create/update/delete + required flag failure
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-transport-plan-trailers"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""
SAMPLE_TRAILER_ID=""
SAMPLE_SEGMENT_START_ID=""
SAMPLE_SEGMENT_END_ID=""
SAMPLE_STATUS=""
SAMPLE_WINDOW_START_AT=""
SAMPLE_WINDOW_END_AT=""

CREATED_ID=""

PLAN_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_ID:-}"
SEGMENT_START_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_START_ID:-}"
SEGMENT_END_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_END_ID:-}"
TRAILER_ID="${XBE_TEST_TRAILER_ID:-}"
STATUS_FILTER="${XBE_TEST_PROJECT_TRANSPORT_PLAN_TRAILER_STATUS:-}"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan trailers"
xbe_json view project-transport-plan-trailers list --limit 5
assert_success

test_name "List project transport plan trailers returns array"
xbe_json view project-transport-plan-trailers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan trailers"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample trailer assignment"
xbe_json view project-transport-plan-trailers list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
    SAMPLE_TRAILER_ID=$(json_get ".[0].trailer_id")
    SAMPLE_SEGMENT_START_ID=$(json_get ".[0].segment_start_id")
    SAMPLE_SEGMENT_END_ID=$(json_get ".[0].segment_end_id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_WINDOW_START_AT=$(json_get ".[0].window_start_at_cached")
    SAMPLE_WINDOW_END_AT=$(json_get ".[0].window_end_at_cached")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No project transport plan trailers available for follow-on tests"
    fi
else
    skip "Could not list project transport plan trailers to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trailer assignments with --project-transport-plan filter"
FILTER_PLAN_ID="$SAMPLE_PLAN_ID"
if [[ -z "$FILTER_PLAN_ID" || "$FILTER_PLAN_ID" == "null" ]]; then
    FILTER_PLAN_ID="$PLAN_ID"
fi
if [[ -n "$FILTER_PLAN_ID" && "$FILTER_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-trailers list --project-transport-plan "$FILTER_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List trailer assignments with --trailer filter"
FILTER_TRAILER_ID="$SAMPLE_TRAILER_ID"
if [[ -z "$FILTER_TRAILER_ID" || "$FILTER_TRAILER_ID" == "null" ]]; then
    FILTER_TRAILER_ID="$TRAILER_ID"
fi
if [[ -n "$FILTER_TRAILER_ID" && "$FILTER_TRAILER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-trailers list --trailer "$FILTER_TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "List trailer assignments with --segment-start filter"
if [[ -n "$SAMPLE_SEGMENT_START_ID" && "$SAMPLE_SEGMENT_START_ID" != "null" ]]; then
    xbe_json view project-transport-plan-trailers list --segment-start "$SAMPLE_SEGMENT_START_ID" --limit 5
    assert_success
else
    skip "No segment start ID available"
fi

test_name "List trailer assignments with --segment-end filter"
if [[ -n "$SAMPLE_SEGMENT_END_ID" && "$SAMPLE_SEGMENT_END_ID" != "null" ]]; then
    xbe_json view project-transport-plan-trailers list --segment-end "$SAMPLE_SEGMENT_END_ID" --limit 5
    assert_success
else
    skip "No segment end ID available"
fi

test_name "List trailer assignments with --status filter"
FILTER_STATUS="$SAMPLE_STATUS"
if [[ -z "$FILTER_STATUS" || "$FILTER_STATUS" == "null" ]]; then
    FILTER_STATUS="$STATUS_FILTER"
fi
if [[ -n "$FILTER_STATUS" && "$FILTER_STATUS" != "null" ]]; then
    xbe_json view project-transport-plan-trailers list --status "$FILTER_STATUS" --limit 5
    assert_success
else
    skip "No status available for filter"
fi

test_name "List trailer assignments with --window-start-at-cached-min filter"
if [[ -n "$SAMPLE_WINDOW_START_AT" && "$SAMPLE_WINDOW_START_AT" != "null" ]]; then
    WINDOW_START_DATE="${SAMPLE_WINDOW_START_AT:0:10}"
    xbe_json view project-transport-plan-trailers list --window-start-at-cached-min "$WINDOW_START_DATE" --limit 5
    assert_success
else
    skip "No window start date available"
fi

test_name "List trailer assignments with --window-start-at-cached-max filter"
if [[ -n "$SAMPLE_WINDOW_START_AT" && "$SAMPLE_WINDOW_START_AT" != "null" ]]; then
    WINDOW_START_DATE="${SAMPLE_WINDOW_START_AT:0:10}"
    xbe_json view project-transport-plan-trailers list --window-start-at-cached-max "$WINDOW_START_DATE" --limit 5
    assert_success
else
    skip "No window start date available"
fi

test_name "List trailer assignments with --window-end-at-cached-min filter"
if [[ -n "$SAMPLE_WINDOW_END_AT" && "$SAMPLE_WINDOW_END_AT" != "null" ]]; then
    WINDOW_END_DATE="${SAMPLE_WINDOW_END_AT:0:10}"
    xbe_json view project-transport-plan-trailers list --window-end-at-cached-min "$WINDOW_END_DATE" --limit 5
    assert_success
else
    skip "No window end date available"
fi

test_name "List trailer assignments with --window-end-at-cached-max filter"
if [[ -n "$SAMPLE_WINDOW_END_AT" && "$SAMPLE_WINDOW_END_AT" != "null" ]]; then
    WINDOW_END_DATE="${SAMPLE_WINDOW_END_AT:0:10}"
    xbe_json view project-transport-plan-trailers list --window-end-at-cached-max "$WINDOW_END_DATE" --limit 5
    assert_success
else
    skip "No window end date available"
fi

test_name "List trailer assignments with --most-recent filter"
xbe_json view project-transport-plan-trailers list --most-recent true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trailer assignment details"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-trailers show "$SAMPLE_ID"
    assert_success
else
    skip "No trailer assignment ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trailer assignment without required flags fails"
xbe_run do project-transport-plan-trailers create
assert_failure

if [[ -n "$PLAN_ID" && -n "$SEGMENT_START_ID" && -n "$SEGMENT_END_ID" ]]; then
    test_name "Create project transport plan trailer"

    create_args=(do project-transport-plan-trailers create \
        --project-transport-plan "$PLAN_ID" \
        --segment-start "$SEGMENT_START_ID" \
        --segment-end "$SEGMENT_END_ID" \
        --status editing)

    if [[ -n "$TRAILER_ID" ]]; then
        create_args+=(--trailer "$TRAILER_ID")
    fi

    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-trailers" "$CREATED_ID"
            pass
        else
            fail "Created trailer assignment but no ID returned"
        fi
    else
        fail "Failed to create project transport plan trailer"
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_ID, XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_START_ID, and XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_END_ID to run create tests"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    if [[ -n "$TRAILER_ID" ]]; then
        test_name "Update trailer assignment to active"
        xbe_json do project-transport-plan-trailers update "$CREATED_ID" --status active --trailer "$TRAILER_ID"
        assert_success
    else
        test_name "Update trailer assignment status"
        xbe_json do project-transport-plan-trailers update "$CREATED_ID" --status editing
        assert_success
    fi
else
    skip "No trailer assignment created; skipping update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete trailer assignment requires --confirm flag"
    xbe_run do project-transport-plan-trailers delete "$CREATED_ID"
    assert_failure

    test_name "Delete trailer assignment with --confirm"
    xbe_run do project-transport-plan-trailers delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No trailer assignment created; skipping delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
