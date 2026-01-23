#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Trailers
#
# Tests CRUD operations for the lineup_scenario_trailers resource.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_LST_ID=""
SAMPLE_LINEUP_SCENARIO_TRUCKER_ID=""
SAMPLE_LINEUP_SCENARIO_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_TRAILER_ID=""
SAMPLE_LAST_ASSIGNED_ON=""
CREATED_LST_ID=""

UPDATE_DATE="2024-01-02"

describe "Resource: lineup-scenario-trailers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup scenario trailers"
xbe_json view lineup-scenario-trailers list --limit 5
assert_success

test_name "List lineup scenario trailers returns array"
xbe_json view lineup-scenario-trailers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup scenario trailers"
fi

# ============================================================================
# Prerequisites - Locate sample lineup scenario trailer
# ============================================================================

test_name "Locate lineup scenario trailer for filters"
xbe_json view lineup-scenario-trailers list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_LST_ID=$(json_get ".[0].id")
        SAMPLE_LINEUP_SCENARIO_TRUCKER_ID=$(json_get ".[0].lineup_scenario_trucker_id")
        SAMPLE_TRAILER_ID=$(json_get ".[0].trailer_id")
        SAMPLE_LAST_ASSIGNED_ON=$(json_get ".[0].last_assigned_on")
        pass
    else
        if [[ -n "$XBE_TEST_LINEUP_SCENARIO_TRAILER_ID" ]]; then
            xbe_json view lineup-scenario-trailers show "$XBE_TEST_LINEUP_SCENARIO_TRAILER_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_LST_ID=$(json_get ".id")
                SAMPLE_LINEUP_SCENARIO_TRUCKER_ID=$(json_get ".lineup_scenario_trucker_id")
                SAMPLE_TRAILER_ID=$(json_get ".trailer_id")
                SAMPLE_LAST_ASSIGNED_ON=$(json_get ".last_assigned_on")
                pass
            else
                skip "Failed to load XBE_TEST_LINEUP_SCENARIO_TRAILER_ID"
            fi
        else
            skip "No lineup scenario trailers found. Set XBE_TEST_LINEUP_SCENARIO_TRAILER_ID to enable filter tests."
        fi
    fi
else
    fail "Failed to list lineup scenario trailers for prerequisites"
fi

# Try to enrich scenario/trucker IDs when available
if [[ -n "$SAMPLE_LST_ID" && "$SAMPLE_LST_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailers show "$SAMPLE_LST_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_LINEUP_SCENARIO_ID=$(json_get ".lineup_scenario_id")
        SAMPLE_TRUCKER_ID=$(json_get ".trucker_id")
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_LST_ID" && "$SAMPLE_LST_ID" != "null" ]]; then
    test_name "Show lineup scenario trailer"
    xbe_json view lineup-scenario-trailers show "$SAMPLE_LST_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show lineup scenario trailer"
    fi
else
    test_name "Show lineup scenario trailer"
    skip "No lineup scenario trailer available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter lineup scenario trailers by lineup scenario trucker"
if [[ -n "$SAMPLE_LINEUP_SCENARIO_TRUCKER_ID" && "$SAMPLE_LINEUP_SCENARIO_TRUCKER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --lineup-scenario-trucker "$SAMPLE_LINEUP_SCENARIO_TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by lineup scenario trucker"
    fi
else
    skip "No lineup scenario trucker ID available for filter test"
fi

test_name "Filter lineup scenario trailers by trailer"
if [[ -n "$SAMPLE_TRAILER_ID" && "$SAMPLE_TRAILER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --trailer "$SAMPLE_TRAILER_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by trailer"
    fi
else
    skip "No trailer ID available for filter test"
fi

test_name "Filter lineup scenario trailers by lineup scenario"
if [[ -n "$SAMPLE_LINEUP_SCENARIO_ID" && "$SAMPLE_LINEUP_SCENARIO_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --lineup-scenario "$SAMPLE_LINEUP_SCENARIO_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by lineup scenario"
    fi
else
    skip "No lineup scenario ID available for filter test"
fi

test_name "Filter lineup scenario trailers by trucker"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --trucker "$SAMPLE_TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by trucker"
    fi
else
    skip "No trucker ID available for filter test"
fi

test_name "Filter lineup scenario trailers by last assigned date"
if [[ -n "$SAMPLE_LAST_ASSIGNED_ON" && "$SAMPLE_LAST_ASSIGNED_ON" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --last-assigned-on "$SAMPLE_LAST_ASSIGNED_ON"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by last assigned date"
    fi
else
    skip "No last assigned date available for filter test"
fi

test_name "Filter lineup scenario trailers by last assigned date min"
if [[ -n "$SAMPLE_LAST_ASSIGNED_ON" && "$SAMPLE_LAST_ASSIGNED_ON" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --last-assigned-on-min "$SAMPLE_LAST_ASSIGNED_ON"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by last assigned date min"
    fi
else
    skip "No last assigned date available for filter test"
fi

test_name "Filter lineup scenario trailers by last assigned date max"
if [[ -n "$SAMPLE_LAST_ASSIGNED_ON" && "$SAMPLE_LAST_ASSIGNED_ON" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --last-assigned-on-max "$SAMPLE_LAST_ASSIGNED_ON"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LST_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario trailer"
        fi
    else
        fail "Failed to filter by last assigned date max"
    fi
else
    skip "No last assigned date available for filter test"
fi

test_name "Filter lineup scenario trailers by has last assigned date"
if [[ -n "$SAMPLE_LAST_ASSIGNED_ON" && "$SAMPLE_LAST_ASSIGNED_ON" != "null" ]]; then
    xbe_json view lineup-scenario-trailers list --has-last-assigned-on true
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Failed to filter by has last assigned date"
    fi
else
    skip "No last assigned date available for filter test"
fi

# ============================================================================
# CREATE Tests - Best Effort
# ============================================================================

test_name "Create lineup scenario trailer"
if [[ -n "$XBE_TEST_LINEUP_SCENARIO_TRUCKER_ID" && -n "$XBE_TEST_TRAILER_ID" ]]; then
    xbe_json do lineup-scenario-trailers create \
        --lineup-scenario-trucker "$XBE_TEST_LINEUP_SCENARIO_TRUCKER_ID" \
        --trailer "$XBE_TEST_TRAILER_ID" \
        --last-assigned-on "$UPDATE_DATE"

    if [[ $status -eq 0 ]]; then
        CREATED_LST_ID=$(json_get ".id")
        if [[ -n "$CREATED_LST_ID" && "$CREATED_LST_ID" != "null" ]]; then
            register_cleanup "lineup-scenario-trailers" "$CREATED_LST_ID"
            pass
        else
            fail "Created lineup scenario trailer but no ID returned"
        fi
    else
        fail "Failed to create lineup scenario trailer"
    fi
else
    skip "Set XBE_TEST_LINEUP_SCENARIO_TRUCKER_ID and XBE_TEST_TRAILER_ID to enable create test"
fi

# ============================================================================
# UPDATE Tests - Best Effort
# ============================================================================

test_name "Update lineup scenario trailer"
if [[ -n "$CREATED_LST_ID" && "$CREATED_LST_ID" != "null" ]]; then
    xbe_json do lineup-scenario-trailers update "$CREATED_LST_ID" --last-assigned-on "$UPDATE_DATE"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".last_assigned_on" "$UPDATE_DATE"
    else
        fail "Failed to update lineup scenario trailer"
    fi
else
    skip "No created lineup scenario trailer to update"
fi

# ============================================================================
# DELETE Tests - Best Effort
# ============================================================================

test_name "Delete lineup scenario trailer"
if [[ -n "$CREATED_LST_ID" && "$CREATED_LST_ID" != "null" ]]; then
    xbe_run do lineup-scenario-trailers delete "$CREATED_LST_ID" --confirm
    assert_success
else
    skip "No created lineup scenario trailer to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required fields fails"
xbe_run do lineup-scenario-trailers create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
