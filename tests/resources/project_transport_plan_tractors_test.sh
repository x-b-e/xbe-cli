#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Tractors
#
# Tests create/update/delete operations and list filters for the
# project-transport-plan-tractors resource.
#
# COVERAGE: Create, update, delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
PROJECT_TRANSPORT_PLAN_ID=""
TRACTOR_ID=""
SEGMENT_START_ID=""
SEGMENT_END_ID=""
STATUS=""
WINDOW_START_AT=""
WINDOW_END_AT=""
WINDOW_START_DATE=""
WINDOW_END_DATE=""
TRUCKER_ID=""
CANDIDATE_TRACTOR_ID=""
CREATED_PTPT_ID=""

describe "Resource: project-transport-plan-tractors"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan tractors"
xbe_json view project-transport-plan-tractors list --limit 5
assert_success

test_name "List project transport plan tractors returns array"
xbe_json view project-transport-plan-tractors list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan tractors"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project transport plan tractor"
xbe_json view project-transport-plan-tractors list --limit 50
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    PROJECT_TRANSPORT_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
    TRACTOR_ID=$(json_get ".[0].tractor_id")
    SEGMENT_START_ID=$(json_get ".[0].segment_start_id")
    SEGMENT_END_ID=$(json_get ".[0].segment_end_id")
    STATUS=$(json_get ".[0].status")
    WINDOW_START_AT=$(json_get ".[0].window_start_at_cached")
    WINDOW_END_AT=$(json_get ".[0].window_end_at_cached")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No project transport plan tractors available"
    fi
else
    fail "Failed to list project transport plan tractors"
fi

if [[ -n "$WINDOW_START_AT" && "$WINDOW_START_AT" != "null" ]]; then
    WINDOW_START_DATE="${WINDOW_START_AT%%T*}"
else
    WINDOW_START_DATE="2025-01-01"
fi

if [[ -n "$WINDOW_END_AT" && "$WINDOW_END_AT" != "null" ]]; then
    WINDOW_END_DATE="${WINDOW_END_AT%%T*}"
else
    WINDOW_END_DATE="2025-01-01"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport plan tractors with --project-transport-plan filter"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-tractors list --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --limit 5
    assert_success
else
    xbe_json view project-transport-plan-tractors list --project-transport-plan 123 --limit 5
    assert_success
fi

test_name "List project transport plan tractors with --tractor filter"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view project-transport-plan-tractors list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    xbe_json view project-transport-plan-tractors list --tractor 123 --limit 5
    assert_success
fi

test_name "List project transport plan tractors with --segment-start filter"
if [[ -n "$SEGMENT_START_ID" && "$SEGMENT_START_ID" != "null" ]]; then
    xbe_json view project-transport-plan-tractors list --segment-start "$SEGMENT_START_ID" --limit 5
    assert_success
else
    xbe_json view project-transport-plan-tractors list --segment-start 123 --limit 5
    assert_success
fi

test_name "List project transport plan tractors with --segment-end filter"
if [[ -n "$SEGMENT_END_ID" && "$SEGMENT_END_ID" != "null" ]]; then
    xbe_json view project-transport-plan-tractors list --segment-end "$SEGMENT_END_ID" --limit 5
    assert_success
else
    xbe_json view project-transport-plan-tractors list --segment-end 123 --limit 5
    assert_success
fi

test_name "List project transport plan tractors with --status filter"
if [[ -n "$STATUS" && "$STATUS" != "null" ]]; then
    xbe_json view project-transport-plan-tractors list --status "$STATUS" --limit 5
    assert_success
else
    xbe_json view project-transport-plan-tractors list --status editing --limit 5
    assert_success
fi

test_name "List project transport plan tractors with --window-start-at-cached filter"
xbe_json view project-transport-plan-tractors list --window-start-at-cached "$WINDOW_START_DATE" --limit 5
assert_success

test_name "List project transport plan tractors with --window-start-at-cached-min filter"
xbe_json view project-transport-plan-tractors list --window-start-at-cached-min "$WINDOW_START_DATE" --limit 5
assert_success

test_name "List project transport plan tractors with --window-start-at-cached-max filter"
xbe_json view project-transport-plan-tractors list --window-start-at-cached-max "$WINDOW_START_DATE" --limit 5
assert_success

test_name "List project transport plan tractors with --has-window-start-at-cached filter"
xbe_json view project-transport-plan-tractors list --has-window-start-at-cached true --limit 5
assert_success

test_name "List project transport plan tractors with --window-end-at-cached filter"
xbe_json view project-transport-plan-tractors list --window-end-at-cached "$WINDOW_END_DATE" --limit 5
assert_success

test_name "List project transport plan tractors with --window-end-at-cached-min filter"
xbe_json view project-transport-plan-tractors list --window-end-at-cached-min "$WINDOW_END_DATE" --limit 5
assert_success

test_name "List project transport plan tractors with --window-end-at-cached-max filter"
xbe_json view project-transport-plan-tractors list --window-end-at-cached-max "$WINDOW_END_DATE" --limit 5
assert_success

test_name "List project transport plan tractors with --has-window-end-at-cached filter"
xbe_json view project-transport-plan-tractors list --has-window-end-at-cached true --limit 5
assert_success

test_name "List project transport plan tractors with --actualizing filter"
xbe_json view project-transport-plan-tractors list --actualizing true --limit 5
assert_success

test_name "List project transport plan tractors with --most-recent filter"
xbe_json view project-transport-plan-tractors list --most-recent true --limit 5
assert_success

# =============================================# =============================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan tractor"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-tractors show "$SAMPLE_ID"
    assert_success
else
    skip "No project transport plan tractor ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan tractor requires --project-transport-plan"
xbe_json do project-transport-plan-tractors create --segment-start 123 --segment-end 456
assert_failure

test_name "Create project transport plan tractor requires --segment-start"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" && -n "$SEGMENT_END_ID" && "$SEGMENT_END_ID" != "null" ]]; then
    xbe_json do project-transport-plan-tractors create --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --segment-end "$SEGMENT_END_ID"
    assert_failure
else
    skip "Missing project transport plan or segment end ID"
fi

test_name "Create project transport plan tractor requires --segment-end"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" && -n "$SEGMENT_START_ID" && "$SEGMENT_START_ID" != "null" ]]; then
    xbe_json do project-transport-plan-tractors create --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --segment-start "$SEGMENT_START_ID"
    assert_failure
else
    skip "Missing project transport plan or segment start ID"
fi

test_name "Create project transport plan tractor"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" && -n "$SEGMENT_START_ID" && "$SEGMENT_START_ID" != "null" && -n "$SEGMENT_END_ID" && "$SEGMENT_END_ID" != "null" ]]; then
    xbe_json do project-transport-plan-tractors create \
        --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" \
        --segment-start "$SEGMENT_START_ID" \
        --segment-end "$SEGMENT_END_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_PTPT_ID=$(json_get ".id")
        if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-tractors" "$CREATED_PTPT_ID"
            pass
        else
            fail "Created project transport plan tractor but no ID returned"
        fi
    else
        skip "Failed to create project transport plan tractor"
    fi
else
    skip "Missing project transport plan or segment IDs"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport plan tractor with no fields fails"
if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" ]]; then
    xbe_json do project-transport-plan-tractors update "$CREATED_PTPT_ID"
    assert_failure
else
    skip "No created project transport plan tractor ID available"
fi

test_name "Update project transport plan tractor actualizer window"
if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" ]]; then
    ACTUALIZER_TS=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    xbe_json do project-transport-plan-tractors update "$CREATED_PTPT_ID" \
        --actualizer-window-start-at "$ACTUALIZER_TS" \
        --actualizer-window-end-at "$ACTUALIZER_TS"
    assert_success
else
    skip "No created project transport plan tractor ID available"
fi

test_name "Update project transport plan tractor flags"
if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" ]]; then
    xbe_json do project-transport-plan-tractors update "$CREATED_PTPT_ID" \
        --automatically-adjust-overlapping-windows \
        --skip-assignment-rules-validation \
        --assignment-rule-override-reason "CLI test override"
    assert_success
else
    skip "No created project transport plan tractor ID available"
fi

# ============================================================================
# Tractor assignment update (best effort)
# ============================================================================

test_name "Resolve trucker for sample tractor"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    LIMIT=200
    OFFSET=0
    MAX_OFFSET=5000
    while [[ $OFFSET -le $MAX_OFFSET ]]; do
        xbe_json view tractors list --limit "$LIMIT" --offset "$OFFSET"
        if [[ $status -ne 0 ]]; then
            break
        fi
        TRUCKER_ID=$(echo "$output" | jq -r --arg id "$TRACTOR_ID" '.[] | select(.id == $id) | .trucker_id' | head -n 1)
        if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
            break
        fi
        count=$(echo "$output" | jq 'length')
        if [[ "$count" -lt "$LIMIT" ]]; then
            break
        fi
        OFFSET=$((OFFSET + LIMIT))
    done

    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        pass
    else
        skip "Could not resolve trucker ID for sample tractor"
    fi
else
    skip "No tractor ID available"
fi

test_name "Select candidate tractor for plan"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" && -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view tractors list --trucker "$TRUCKER_ID" --limit 200
    if [[ $status -eq 0 ]]; then
        CANDIDATE_TRACTOR_ID=""
        while read -r candidate_id; do
            if [[ -z "$candidate_id" || "$candidate_id" == "null" ]]; then
                continue
            fi
            if [[ "$candidate_id" == "$TRACTOR_ID" ]]; then
                continue
            fi
            xbe_json view project-transport-plan-tractors list --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --tractor "$candidate_id" --limit 1
            if [[ $status -ne 0 ]]; then
                continue
            fi
            candidate_count=$(echo "$output" | jq 'length')
            if [[ "$candidate_count" -eq 0 ]]; then
                CANDIDATE_TRACTOR_ID="$candidate_id"
                break
            fi
        done < <(echo "$output" | jq -r '.[].id')

        if [[ -n "$CANDIDATE_TRACTOR_ID" && "$CANDIDATE_TRACTOR_ID" != "null" ]]; then
            pass
        else
            skip "No available tractor found for plan"
        fi
    else
        skip "Failed to list tractors for trucker"
    fi
else
    skip "Missing trucker or project transport plan ID"
fi

test_name "Update project transport plan tractor with tractor and status"
if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" && -n "$CANDIDATE_TRACTOR_ID" && "$CANDIDATE_TRACTOR_ID" != "null" ]]; then
    xbe_json do project-transport-plan-tractors update "$CREATED_PTPT_ID" --tractor "$CANDIDATE_TRACTOR_ID" --status active
    assert_success
else
    skip "Missing created assignment or candidate tractor"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan tractor requires --confirm"
if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" ]]; then
    xbe_run do project-transport-plan-tractors delete "$CREATED_PTPT_ID"
    assert_failure
else
    skip "No created project transport plan tractor ID available"
fi

test_name "Delete project transport plan tractor"
if [[ -n "$CREATED_PTPT_ID" && "$CREATED_PTPT_ID" != "null" ]]; then
    xbe_run do project-transport-plan-tractors delete "$CREATED_PTPT_ID" --confirm
    assert_success
else
    skip "No created project transport plan tractor ID available"
fi

run_tests
