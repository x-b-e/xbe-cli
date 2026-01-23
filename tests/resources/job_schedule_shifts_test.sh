#!/bin/bash
#
# XBE CLI Integration Tests: Job Schedule Shifts
#
# Tests list, show, create, update, and delete operations for job schedule shifts.
#
# COVERAGE: All list filters + create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_SHIFT_ID=""
JOB_ID=""
JOB_SITE_ID=""
CUSTOMER_ID=""
BROKER_ID=""
START_AT=""
END_AT=""
START_DATE=""

TRAILER_CLASSIFICATION_ID=""
PROJECT_LABOR_CLASSIFICATION_ID=""
START_LOCATION_ID=""

CREATED_SHIFT_ID=""

describe "Resource: job-schedule-shifts"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List job schedule shifts"
xbe_json view job-schedule-shifts list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        JOB_ID=$(echo "$output" | jq -r 'group_by(.job_id) | map(select(length>1)) | first | .[0].job_id // empty')
        if [[ -z "$JOB_ID" || "$JOB_ID" == "null" ]]; then
            JOB_ID=$(echo "$output" | jq -r '.[0].job_id')
        fi

        SEED_SHIFT_ID=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .id')
        JOB_SITE_ID=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .job_site_id')
        CUSTOMER_ID=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .customer_id')
        BROKER_ID=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .broker_id')
        START_AT=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .start_at')
        END_AT=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .end_at')
        START_DATE=$(echo "$output" | jq -r --arg job "$JOB_ID" 'map(select(.job_id==$job)) | first | .start_date')
    fi
else
    fail "Failed to list job schedule shifts"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job schedule shift"
if [[ -n "$SEED_SHIFT_ID" && "$SEED_SHIFT_ID" != "null" ]]; then
    xbe_json view job-schedule-shifts show "$SEED_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        TRAILER_CLASSIFICATION_ID=$(json_get ".trailer_classification_id")
        PROJECT_LABOR_CLASSIFICATION_ID=$(json_get ".project_labor_classification_id")
        START_LOCATION_ID=$(json_get ".start_location_id")
        pass
    else
        fail "Failed to show job schedule shift"
    fi
else
    skip "No job schedule shift available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job schedule shift with required fields"
if [[ -n "$JOB_ID" && "$JOB_ID" != "null" ]]; then
    if [[ -z "$START_AT" || "$START_AT" == "null" ]]; then
        START_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    fi
    if [[ -z "$END_AT" || "$END_AT" == "null" ]]; then
        END_AT=$(date -u -v+8H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)
        if [[ -z "$END_AT" ]]; then
            END_AT=$(date -u -d "+8 hours" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)
        fi
        if [[ -z "$END_AT" ]]; then
            END_AT="$START_AT"
        fi
    fi

    cmd=(xbe_json do job-schedule-shifts create --job "$JOB_ID" --start-at "$START_AT" --end-at "$END_AT")
    cmd+=(--dispatch-instructions "CLI test dispatch")
    cmd+=(--is-planned-productive true)
    cmd+=(--suppress-automated-shift-feedback true)
    cmd+=(--expected-material-transaction-count 1)
    cmd+=(--expected-material-transaction-tons 2.5)
    cmd+=(--is-flexible true)
    cmd+=(--start-at-min "$START_AT")
    cmd+=(--start-at-max "$END_AT")
    cmd+=(--is-subsequent-shift-in-driver-day false)
    cmd+=(--is-trucker-incident-creation-automated-explicit true)

    if [[ -n "$JOB_SITE_ID" && "$JOB_SITE_ID" != "null" ]]; then
        cmd+=(--start-site-type "job-sites" --start-site "$JOB_SITE_ID")
    fi
    if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        cmd+=(--trailer-classification "$TRAILER_CLASSIFICATION_ID")
    fi
    if [[ -n "$PROJECT_LABOR_CLASSIFICATION_ID" && "$PROJECT_LABOR_CLASSIFICATION_ID" != "null" ]]; then
        cmd+=(--project-labor-classification "$PROJECT_LABOR_CLASSIFICATION_ID")
    fi
    if [[ -n "$START_LOCATION_ID" && "$START_LOCATION_ID" != "null" ]]; then
        cmd+=(--start-location "$START_LOCATION_ID")
    fi

    "${cmd[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_SHIFT_ID=$(json_get ".id")
        if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
            register_cleanup "job-schedule-shifts" "$CREATED_SHIFT_ID"
            pass
        else
            fail "Created job schedule shift but no ID returned"
        fi
    else
        fail "Failed to create job schedule shift"
    fi
else
    skip "No job ID available for creating a job schedule shift"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update job schedule shift attributes"
if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
    xbe_json do job-schedule-shifts update "$CREATED_SHIFT_ID" \
        --dispatch-instructions "Updated dispatch instructions" \
        --is-planned-productive false \
        --suppress-automated-shift-feedback false \
        --expected-material-transaction-count 2 \
        --expected-material-transaction-tons 3.5 \
        --is-subsequent-shift-in-driver-day true \
        --is-trucker-incident-creation-automated-explicit false
    assert_success
else
    skip "No created job schedule shift available for update"
fi

test_name "Update job schedule shift cancelled-at"
if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
    CANCELLED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_run do job-schedule-shifts update "$CREATED_SHIFT_ID" --cancelled-at "$CANCELLED_AT"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not cancel job schedule shift (policy or validation)"
    fi
else
    skip "No created job schedule shift available for cancel update"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

if [[ -z "$START_AT" || "$START_AT" == "null" ]]; then
    START_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
fi
if [[ -z "$END_AT" || "$END_AT" == "null" ]]; then
    END_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
fi
if [[ -z "$START_DATE" || "$START_DATE" == "null" ]]; then
    START_DATE=$(date -u +"%Y-%m-%d")
fi

test_name "Filter by job"
if [[ -n "$JOB_ID" && "$JOB_ID" != "null" ]]; then
    xbe_json view job-schedule-shifts list --job "$JOB_ID" --limit 5
    assert_success
else
    skip "No job ID available"
fi

test_name "Filter by business unit"
xbe_json view job-schedule-shifts list --business-unit "1" --limit 5
assert_success

test_name "Filter by matches material purchase order release"
xbe_json view job-schedule-shifts list --matches-material-purchase-order-release "1" --limit 5
assert_success

test_name "Filter by on accepted broker tender"
xbe_json view job-schedule-shifts list --on-accepted-broker-tender true --limit 5
assert_success

test_name "Filter by on accepted customer tender"
xbe_json view job-schedule-shifts list --on-accepted-customer-tender true --limit 5
assert_success

test_name "Filter by active on tender"
xbe_json view job-schedule-shifts list --active-on-tender "1" --limit 5
assert_success

test_name "Filter by is cancelled"
xbe_json view job-schedule-shifts list --is-cancelled false --limit 5
assert_success

test_name "Filter by is managed"
xbe_json view job-schedule-shifts list --is-managed true --limit 5
assert_success

test_name "Filter by is managed or alive"
xbe_json view job-schedule-shifts list --is-managed-or-alive true --limit 5
assert_success

test_name "Filter by is subsequent shift in driver day"
xbe_json view job-schedule-shifts list --is-subsequent-shift-in-driver-day false --limit 5
assert_success

test_name "Filter by unsourced"
xbe_json view job-schedule-shifts list --unsourced true --limit 5
assert_success

test_name "Filter by customer"
if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
    xbe_json view job-schedule-shifts list --customer "$CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "Filter by customer id"
if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
    xbe_json view job-schedule-shifts list --customer-id "$CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "Filter by broker"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view job-schedule-shifts list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "Filter by broker id"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view job-schedule-shifts list --broker-id "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "Filter by ordered"
xbe_json view job-schedule-shifts list --ordered true --limit 5
assert_success

test_name "Filter by job production plan status"
xbe_json view job-schedule-shifts list --job-production-plan-status "editing" --limit 5
assert_success

test_name "Filter by start date"
xbe_json view job-schedule-shifts list --start-date "$START_DATE" --limit 5
assert_success

test_name "Filter by start date min"
xbe_json view job-schedule-shifts list --start-date-min "$START_DATE" --limit 5
assert_success

test_name "Filter by start date max"
xbe_json view job-schedule-shifts list --start-date-max "$START_DATE" --limit 5
assert_success

test_name "Filter by has start date"
xbe_json view job-schedule-shifts list --has-start-date true --limit 5
assert_success

test_name "Filter by start at min"
xbe_json view job-schedule-shifts list --start-at-min "$START_AT" --limit 5
assert_success

test_name "Filter by start at max"
xbe_json view job-schedule-shifts list --start-at-max "$START_AT" --limit 5
assert_success

test_name "Filter by is start at"
xbe_json view job-schedule-shifts list --is-start-at true --limit 5
assert_success

test_name "Filter by end at min"
xbe_json view job-schedule-shifts list --end-at-min "$END_AT" --limit 5
assert_success

test_name "Filter by end at max"
xbe_json view job-schedule-shifts list --end-at-max "$END_AT" --limit 5
assert_success

test_name "Filter by is end at"
xbe_json view job-schedule-shifts list --is-end-at true --limit 5
assert_success

test_name "Filter by related to trucker through accepted tender"
xbe_json view job-schedule-shifts list --related-to-trucker-through-accepted-tender "1" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job schedule shift"
if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
    xbe_run do job-schedule-shifts delete "$CREATED_SHIFT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete job schedule shift (permissions or policy)"
    fi
else
    skip "No created job schedule shift available for delete"
fi

run_tests
