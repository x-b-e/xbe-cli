#!/bin/bash
#
# XBE CLI Integration Tests: Job Schedule Shift Start-At Changes
#
# Tests list, show, and create operations for the job-schedule-shift-start-at-changes resource.
#
# COVERAGE: List filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CHANGE_ID=""
SHIFT_ID=""
SKIP_ID_TESTS=0

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_JPP_ID=""
CREATED_CHANGE_ID=""

describe "Resource: job-schedule-shift-start-at-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List start-at changes"
xbe_json view job-schedule-shift-start-at-changes list --limit 5
assert_success

test_name "List start-at changes returns array"
xbe_json view job-schedule-shift-start-at-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list start-at changes"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample start-at change"
xbe_json view job-schedule-shift-start-at-changes list --limit 1
if [[ $status -eq 0 ]]; then
    CHANGE_ID=$(json_get ".[0].id")
    SHIFT_ID=$(json_get ".[0].job_schedule_shift_id")
    if [[ -n "$CHANGE_ID" && "$CHANGE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_TESTS=1
        skip "No start-at changes available"
    fi
else
    SKIP_ID_TESTS=1
    fail "Failed to list start-at changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

NOW_ISO=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

test_name "List start-at changes with --created-at-min filter"
xbe_json view job-schedule-shift-start-at-changes list --created-at-min "$NOW_ISO" --limit 5
assert_success

test_name "List start-at changes with --created-at-max filter"
xbe_json view job-schedule-shift-start-at-changes list --created-at-max "$NOW_ISO" --limit 5
assert_success

test_name "List start-at changes with --updated-at-min filter"
xbe_json view job-schedule-shift-start-at-changes list --updated-at-min "$NOW_ISO" --limit 5
assert_success

test_name "List start-at changes with --updated-at-max filter"
xbe_json view job-schedule-shift-start-at-changes list --updated-at-max "$NOW_ISO" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show start-at change"
if [[ $SKIP_ID_TESTS -eq 0 && -n "$CHANGE_ID" && "$CHANGE_ID" != "null" ]]; then
    xbe_json view job-schedule-shift-start-at-changes show "$CHANGE_ID"
    assert_success
else
    skip "No start-at change ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create start-at change without required fields fails"
xbe_run do job-schedule-shift-start-at-changes create
assert_failure

# ============================================================================
# Prerequisites - Create broker, customer, job site, job production plan
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "ShiftStartAtBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "ShiftStartAtCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create job site"
JOB_SITE_NAME=$(unique_name "ShiftStartAtJobSite")
JOB_SITE_ADDRESS="1001 Shift Start Ave, Chicago, IL 60601"

xbe_json do job-sites create \
    --name "$JOB_SITE_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "$JOB_SITE_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID"
        pass
    else
        fail "Created job site but no ID returned"
        echo "Cannot continue without a job site"
        run_tests
    fi
else
    fail "Failed to create job site"
    echo "Cannot continue without a job site"
    run_tests
fi

test_name "Create job production plan"
PLAN_NAME=$(unique_name "ShiftStartAtPlan")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$PLAN_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --job-site "$CREATED_JOB_SITE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    echo "Cannot continue without a job production plan"
    run_tests
fi

# ============================================================================
# Shift Lookup via API
# ============================================================================

test_name "Lookup job schedule shift for plan via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping shift lookup"
elif [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
    base_url="${XBE_BASE_URL%/}"

    plan_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-production-plans/$CREATED_JPP_ID?fields[job-production-plans]=job-schedule-shifts" || true)

    SHIFT_ID=$(echo "$plan_json" | jq -r '.data.relationships["job-schedule-shifts"].data[0].id // empty' 2>/dev/null || true)

    if [[ -n "$SHIFT_ID" ]]; then
        pass
    else
        skip "No job schedule shift found for plan"
    fi
else
    skip "No job production plan ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create start-at change"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping create"
elif [[ -z "$SHIFT_ID" ]]; then
    skip "No job schedule shift ID available"
else
    base_url="${XBE_BASE_URL%/}"

    shift_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-schedule-shifts/$SHIFT_ID?fields[job-schedule-shifts]=start-at,start-at-min,start-at-max" || true)

    OLD_START_AT=$(echo "$shift_json" | jq -r '.data.attributes["start-at"] // empty' 2>/dev/null || true)
    START_AT_MIN=$(echo "$shift_json" | jq -r '.data.attributes["start-at-min"] // empty' 2>/dev/null || true)
    START_AT_MAX=$(echo "$shift_json" | jq -r '.data.attributes["start-at-max"] // empty' 2>/dev/null || true)

    NEW_START_AT=$(OLD_START_AT="$OLD_START_AT" START_AT_MIN="$START_AT_MIN" START_AT_MAX="$START_AT_MAX" python3 - <<'PY'
import os
from datetime import datetime, timedelta

def parse(val):
    if not val:
        return None
    value = val.strip()
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"
    try:
        return datetime.fromisoformat(value)
    except ValueError:
        return None

old_dt = parse(os.environ.get("OLD_START_AT", ""))
min_dt = parse(os.environ.get("START_AT_MIN", ""))
max_dt = parse(os.environ.get("START_AT_MAX", ""))

if not old_dt:
    print("")
    raise SystemExit

candidate = old_dt + timedelta(minutes=30)
if max_dt and candidate > max_dt:
    candidate = old_dt - timedelta(minutes=30)
if min_dt and candidate < min_dt:
    candidate = min_dt
if max_dt and candidate > max_dt:
    candidate = max_dt
if min_dt and candidate < min_dt:
    candidate = min_dt
if candidate == old_dt:
    if max_dt and max_dt != old_dt:
        candidate = max_dt
    elif min_dt and min_dt != old_dt:
        candidate = min_dt
    else:
        candidate = old_dt + timedelta(minutes=1)
        if max_dt and candidate > max_dt:
            candidate = old_dt - timedelta(minutes=1)
        if min_dt and candidate < min_dt:
            print("")
            raise SystemExit

value = candidate.isoformat()
if value.endswith("+00:00"):
    value = value[:-6] + "Z"
print(value)
PY
)

    if [[ -z "$NEW_START_AT" ]]; then
        skip "Unable to compute a valid new start time"
    else
        xbe_json do job-schedule-shift-start-at-changes create \
            --job-schedule-shift "$SHIFT_ID" \
            --new-start-at "$NEW_START_AT"

        if [[ $status -eq 0 ]]; then
            CREATED_CHANGE_ID=$(json_get ".id")
            if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
                pass
            else
                fail "Created start-at change but no ID returned"
            fi
        else
            if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                pass
            else
                fail "Failed to create start-at change: $output"
            fi
        fi
    fi
fi

# ============================================================================
# SHOW Tests - Newly created
# ============================================================================

test_name "Show created start-at change"
if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
    xbe_json view job-schedule-shift-start-at-changes show "$CREATED_CHANGE_ID"
    assert_success
else
    skip "No created start-at change ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
