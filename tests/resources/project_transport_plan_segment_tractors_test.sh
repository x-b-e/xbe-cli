#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Segment Tractors
#
# Tests create/delete operations and list filters for the
# project-transport-plan-segment-tractors resource.
#
# COVERAGE: Create, delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SEGMENT_ID=""
TRACTOR_ID=""
TRUCKER_ID=""
CREATED_TRACTOR_ID=""
CREATED_SEGMENT_TRACTOR_ID=""
CANDIDATE_TRACTOR_ID=""

describe "Resource: project-transport-plan-segment-tractors"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan segment tractors"
xbe_json view project-transport-plan-segment-tractors list --limit 5
assert_success

test_name "List project transport plan segment tractors returns array"
xbe_json view project-transport-plan-segment-tractors list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan segment tractors"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project transport plan segment tractor"
xbe_json view project-transport-plan-segment-tractors list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SEGMENT_ID=$(json_get ".[0].project_transport_plan_segment_id")
    TRACTOR_ID=$(json_get ".[0].tractor_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No project transport plan segment tractors available"
    fi
else
    fail "Failed to list project transport plan segment tractors"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport plan segment tractors with --project-transport-plan-segment filter"
if [[ -n "$SEGMENT_ID" && "$SEGMENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-tractors list --project-transport-plan-segment "$SEGMENT_ID" --limit 5
    assert_success
else
    skip "No project transport plan segment ID available"
fi

test_name "List project transport plan segment tractors with --tractor filter"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-tractors list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan segment tractor"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-tractors show "$SAMPLE_ID"
    assert_success
else
    skip "No project transport plan segment tractor ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan segment tractor requires --project-transport-plan-segment"
xbe_json do project-transport-plan-segment-tractors create --tractor 123
assert_failure

test_name "Create project transport plan segment tractor requires --tractor"
if [[ -n "$SEGMENT_ID" && "$SEGMENT_ID" != "null" ]]; then
    xbe_json do project-transport-plan-segment-tractors create --project-transport-plan-segment "$SEGMENT_ID"
    assert_failure
else
    skip "No project transport plan segment ID available"
fi

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

test_name "Select candidate tractor for segment"
if [[ -n "$SEGMENT_ID" && "$SEGMENT_ID" != "null" && -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-tractors list --project-transport-plan-segment "$SEGMENT_ID" --limit 200
    if [[ $status -eq 0 ]]; then
        EXISTING_TRACTORS=$(echo "$output" | jq -r '.[].tractor_id')
    else
        EXISTING_TRACTORS=""
    fi

    xbe_json view tractors list --trucker "$TRUCKER_ID" --limit 200
    if [[ $status -eq 0 ]]; then
        CANDIDATE_TRACTOR_ID=""
        while read -r candidate_id; do
            if [[ -z "$candidate_id" || "$candidate_id" == "null" ]]; then
                continue
            fi
            if ! echo "$EXISTING_TRACTORS" | grep -qx "$candidate_id"; then
                CANDIDATE_TRACTOR_ID="$candidate_id"
                break
            fi
        done < <(echo "$output" | jq -r '.[].id')

        if [[ -n "$CANDIDATE_TRACTOR_ID" && "$CANDIDATE_TRACTOR_ID" != "null" ]]; then
            pass
        else
            skip "No available tractor found for segment"
        fi
    else
        skip "Failed to list tractors for trucker"
    fi
else
    skip "Missing segment or trucker ID"
fi

if [[ -z "$CANDIDATE_TRACTOR_ID" && -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    test_name "Create tractor for segment trucker"
    TRACTOR_NUMBER="SEGTRAC$(date +%s)$RANDOM"
    xbe_json do tractors create --number "$TRACTOR_NUMBER" --trucker "$TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_TRACTOR_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRACTOR_ID" && "$CREATED_TRACTOR_ID" != "null" ]]; then
            register_cleanup "tractors" "$CREATED_TRACTOR_ID"
            CANDIDATE_TRACTOR_ID="$CREATED_TRACTOR_ID"
            pass
        else
            fail "Created tractor but no ID returned"
        fi
    else
        skip "Failed to create tractor for trucker"
    fi
fi

test_name "Create project transport plan segment tractor"
if [[ -n "$SEGMENT_ID" && "$SEGMENT_ID" != "null" && -n "$CANDIDATE_TRACTOR_ID" && "$CANDIDATE_TRACTOR_ID" != "null" ]]; then
    xbe_json do project-transport-plan-segment-tractors create \
        --project-transport-plan-segment "$SEGMENT_ID" \
        --tractor "$CANDIDATE_TRACTOR_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_SEGMENT_TRACTOR_ID=$(json_get ".id")
        if [[ -n "$CREATED_SEGMENT_TRACTOR_ID" && "$CREATED_SEGMENT_TRACTOR_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-segment-tractors" "$CREATED_SEGMENT_TRACTOR_ID"
            pass
        else
            fail "Created segment tractor but no ID returned"
        fi
    else
        skip "Failed to create project transport plan segment tractor"
    fi
else
    skip "Missing segment or candidate tractor ID"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan segment tractor requires --confirm"
if [[ -n "$CREATED_SEGMENT_TRACTOR_ID" && "$CREATED_SEGMENT_TRACTOR_ID" != "null" ]]; then
    xbe_run do project-transport-plan-segment-tractors delete "$CREATED_SEGMENT_TRACTOR_ID"
    assert_failure
else
    skip "No created segment tractor ID available"
fi

test_name "Delete project transport plan segment tractor"
if [[ -n "$CREATED_SEGMENT_TRACTOR_ID" && "$CREATED_SEGMENT_TRACTOR_ID" != "null" ]]; then
    xbe_run do project-transport-plan-segment-tractors delete "$CREATED_SEGMENT_TRACTOR_ID" --confirm
    assert_success
else
    skip "No created segment tractor ID available"
fi

run_tests
