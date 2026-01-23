#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Job Production Plans
#
# Tests CRUD operations for the lineup_job_production_plans resource.
#
# COVERAGE: All filters + all create attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_LJPP_ID=""
SAMPLE_LINEUP_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""
CREATED_LJPP_ID=""
CREATED_JPP_ID=""
CUSTOMER_NAME=""
CUSTOMER_ID=""
START_ON=""
START_TIME=""
SKIP_CREATE=0

describe "Resource: lineup-job-production-plans"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup job production plans"
xbe_json view lineup-job-production-plans list --limit 5
assert_success

test_name "List lineup job production plans returns array"
xbe_json view lineup-job-production-plans list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup job production plans"
fi

# ============================================================================
# Prerequisites - Locate sample lineup job production plan
# ============================================================================

test_name "Locate lineup job production plan for filters"
xbe_json view lineup-job-production-plans list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_LJPP_ID=$(json_get ".[0].id")
        SAMPLE_LINEUP_ID=$(json_get ".[0].lineup_id")
        SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
        pass
    else
        if [[ -n "$XBE_TEST_LINEUP_JOB_PRODUCTION_PLAN_ID" ]]; then
            xbe_json view lineup-job-production-plans show "$XBE_TEST_LINEUP_JOB_PRODUCTION_PLAN_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_LJPP_ID=$(json_get ".id")
                SAMPLE_LINEUP_ID=$(json_get ".lineup_id")
                SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".job_production_plan_id")
                pass
            else
                skip "Failed to load XBE_TEST_LINEUP_JOB_PRODUCTION_PLAN_ID"
                SKIP_CREATE=1
            fi
        else
            skip "No lineup job production plans found. Set XBE_TEST_LINEUP_JOB_PRODUCTION_PLAN_ID to enable filter/create tests."
            SKIP_CREATE=1
        fi
    fi
else
    fail "Failed to list lineup job production plans for prerequisites"
    SKIP_CREATE=1
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_LJPP_ID" && "$SAMPLE_LJPP_ID" != "null" ]]; then
    test_name "Show lineup job production plan"
    xbe_json view lineup-job-production-plans show "$SAMPLE_LJPP_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show lineup job production plan"
    fi
else
    test_name "Show lineup job production plan"
    skip "No lineup job production plan available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter lineup job production plans by lineup"
if [[ -n "$SAMPLE_LINEUP_ID" && "$SAMPLE_LINEUP_ID" != "null" ]]; then
    xbe_json view lineup-job-production-plans list --lineup "$SAMPLE_LINEUP_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LJPP_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup job production plan"
        fi
    else
        fail "Failed to filter by lineup"
    fi
else
    skip "No lineup ID available for filter test"
fi

test_name "Filter lineup job production plans by job production plan"
if [[ -n "$SAMPLE_JOB_PRODUCTION_PLAN_ID" && "$SAMPLE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view lineup-job-production-plans list --job-production-plan "$SAMPLE_JOB_PRODUCTION_PLAN_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LJPP_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup job production plan"
        fi
    else
        fail "Failed to filter by job production plan"
    fi
else
    skip "No job production plan ID available for filter test"
fi

# ============================================================================
# CREATE Tests - Best Effort
# ============================================================================

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Resolve lineup customer and start time"
    xbe_json view job-production-plans show "$SAMPLE_JOB_PRODUCTION_PLAN_ID"
    if [[ $status -eq 0 ]]; then
        CUSTOMER_NAME=$(json_get ".customer")
        START_ON=$(json_get ".start_on")
        START_TIME=$(json_get ".start_time")
        if [[ -n "$CUSTOMER_NAME" && "$CUSTOMER_NAME" != "null" && -n "$START_ON" && "$START_ON" != "null" && -n "$START_TIME" && "$START_TIME" != "null" ]]; then
            pass
        else
            skip "Missing customer or start time from sample job production plan"
            SKIP_CREATE=1
        fi
    else
        skip "Failed to load sample job production plan"
        SKIP_CREATE=1
    fi
fi

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Resolve customer ID for lineup creation"
    xbe_json view customers list --name "$CUSTOMER_NAME" --limit 20
    if [[ $status -eq 0 ]]; then
        CANDIDATE_IDS=$(echo "$output" | jq -r --arg name "$CUSTOMER_NAME" '.[] | select(.name == $name) | .id')
        if [[ -n "$CANDIDATE_IDS" ]]; then
            for candidate in $CANDIDATE_IDS; do
                xbe_json view job-production-plans list --start-on "$START_ON" --customer "$candidate" --limit 50
                if [[ $status -eq 0 ]]; then
                    if echo "$output" | jq -e --arg id "$SAMPLE_JOB_PRODUCTION_PLAN_ID" '.[] | select(.id == $id)' > /dev/null; then
                        CUSTOMER_ID="$candidate"
                        break
                    fi
                fi
            done
        fi

        if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
            pass
        else
            skip "Failed to match customer ID for lineup"
            SKIP_CREATE=1
        fi
    else
        skip "Failed to list customers for customer ID lookup"
        SKIP_CREATE=1
    fi
fi

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Create job production plan for lineup"
    JPP_NAME=$(unique_name "LineupPlan")
    xbe_json do job-production-plans create \
        --job-name "$JPP_NAME" \
        --start-on "$START_ON" \
        --start-time "$START_TIME" \
        --customer "$CUSTOMER_ID" \
        --explicit-requires-inspector false

    if [[ $status -eq 0 ]]; then
        CREATED_JPP_ID=$(json_get ".id")
        if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
            register_cleanup "job-production-plans" "$CREATED_JPP_ID"
            pass
        else
            fail "Created job production plan but no ID returned"
            SKIP_CREATE=1
        fi
    else
        fail "Failed to create job production plan"
        SKIP_CREATE=1
    fi
else
    test_name "Create job production plan for lineup"
    skip "Prerequisites not available"
fi

if [[ $SKIP_CREATE -eq 0 && -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
    test_name "Create lineup job production plan"
    xbe_json do lineup-job-production-plans create \
        --lineup "$SAMPLE_LINEUP_ID" \
        --job-production-plan "$CREATED_JPP_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LJPP_ID=$(json_get ".id")
        if [[ -n "$CREATED_LJPP_ID" && "$CREATED_LJPP_ID" != "null" ]]; then
            register_cleanup "lineup-job-production-plans" "$CREATED_LJPP_ID"
            pass
        else
            fail "Created lineup job production plan but no ID returned"
        fi
    else
        fail "Failed to create lineup job production plan"
    fi
else
    test_name "Create lineup job production plan"
    skip "Prerequisites not available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LJPP_ID" && "$CREATED_LJPP_ID" != "null" ]]; then
    test_name "Delete lineup job production plan"
    xbe_run do lineup-job-production-plans delete "$CREATED_LJPP_ID" --confirm
    assert_success
else
    test_name "Delete lineup job production plan"
    skip "No lineup job production plan created"
fi

run_tests
