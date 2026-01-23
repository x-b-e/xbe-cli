#!/bin/bash
#
# XBE CLI Integration Tests: Job Site Times
#
# Tests CRUD operations for the job_site_times resource.
# Job site times capture user time at job sites for job production plans.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_JOB_SITE_TIME_ID=""
JOB_PRODUCTION_PLAN_ID=""
USER_ID=""
SAMPLE_JOB_SITE_TIME_ID=""
SAMPLE_START_AT=""
SAMPLE_END_AT=""
CANDIDATE_START_AT=""
CANDIDATE_END_AT=""
SKIP_CREATE=0

describe "Resource: job-site-times"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job site times"
xbe_json view job-site-times list --limit 5
assert_success

test_name "List job site times returns array"
xbe_json view job-site-times list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job site times"
fi

# ============================================================================
# Prerequisites - Locate sample job site time
# ============================================================================

test_name "Locate job site time for prerequisites"
xbe_json view job-site-times list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_JOB_SITE_TIME_ID=$(json_get ".[0].id")
        JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
        USER_ID=$(json_get ".[0].user_id")
        SAMPLE_START_AT=$(json_get ".[0].start_at")
        SAMPLE_END_AT=$(json_get ".[0].end_at")
        pass
    else
        if [[ -n "$XBE_TEST_JOB_SITE_TIME_ID" ]]; then
            xbe_json view job-site-times show "$XBE_TEST_JOB_SITE_TIME_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_JOB_SITE_TIME_ID=$(json_get ".id")
                JOB_PRODUCTION_PLAN_ID=$(json_get ".job_production_plan_id")
                USER_ID=$(json_get ".user_id")
                SAMPLE_START_AT=$(json_get ".start_at")
                SAMPLE_END_AT=$(json_get ".end_at")
                pass
            else
                skip "Failed to load XBE_TEST_JOB_SITE_TIME_ID"
                SKIP_CREATE=1
            fi
        else
            skip "No job site times found. Set XBE_TEST_JOB_SITE_TIME_ID to enable create/update tests."
            SKIP_CREATE=1
        fi
    fi
else
    fail "Failed to list job site times for prerequisites"
    SKIP_CREATE=1
fi

# ============================================================================
# Compute a non-overlapping time window for create/update
# ============================================================================

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Compute non-overlapping time window"
    xbe_json view job-site-times list --job-production-plan "$JOB_PRODUCTION_PLAN_ID" --user "$USER_ID" --limit 200
    if [[ $status -eq 0 ]]; then
        CANDIDATE_TIMES=$(python3 - <<'PY'
import json
import sys
from datetime import datetime, timedelta

raw = sys.stdin.read().strip()
if not raw:
    sys.exit(0)

data = json.loads(raw)
intervals = []
for item in data:
    s = item.get("start_at")
    if not s:
        continue
    e = item.get("end_at") or s
    try:
        s_dt = datetime.fromisoformat(s.replace("Z", "+00:00"))
        e_dt = datetime.fromisoformat(e.replace("Z", "+00:00"))
    except ValueError:
        continue
    if e_dt < s_dt:
        e_dt = s_dt
    intervals.append((s_dt, e_dt))

if not intervals:
    sys.exit(0)

intervals.sort(key=lambda x: x[0])
min_start = intervals[0][0]
max_end = max(e for _, e in intervals)

duration = timedelta(minutes=30)
step = timedelta(minutes=15)

candidate = min_start
while candidate + duration <= max_end:
    cand_end = candidate + duration
    overlaps = False
    for start, end in intervals:
        if candidate < end and cand_end > start:
            overlaps = True
            break
    if not overlaps:
        print(candidate.isoformat().replace("+00:00", "Z"))
        print(cand_end.isoformat().replace("+00:00", "Z"))
        sys.exit(0)
    candidate += step
PY
<<< "$output")

        CANDIDATE_START_AT=$(echo "$CANDIDATE_TIMES" | sed -n '1p')
        CANDIDATE_END_AT=$(echo "$CANDIDATE_TIMES" | sed -n '2p')

        if [[ -n "$CANDIDATE_START_AT" && -n "$CANDIDATE_END_AT" ]]; then
            pass
        else
            skip "No available time window found; skipping create/update tests"
            SKIP_CREATE=1
        fi
    else
        skip "Failed to load job site times for time window computation"
        SKIP_CREATE=1
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Create job site time with required fields"
    xbe_json do job-site-times create \
        --job-production-plan "$JOB_PRODUCTION_PLAN_ID" \
        --user "$USER_ID" \
        --start-at "$CANDIDATE_START_AT" \
        --end-at "$CANDIDATE_END_AT" \
        --description "Test job site time"

    if [[ $status -eq 0 ]]; then
        CREATED_JOB_SITE_TIME_ID=$(json_get ".id")
        if [[ -n "$CREATED_JOB_SITE_TIME_ID" && "$CREATED_JOB_SITE_TIME_ID" != "null" ]]; then
            register_cleanup "job-site-times" "$CREATED_JOB_SITE_TIME_ID"
            pass
        else
            fail "Created job site time but no ID returned"
        fi
    else
        fail "Failed to create job site time"
    fi
else
    test_name "Create job site time with required fields"
    skip "Prerequisites not available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_JOB_SITE_TIME_ID" && "$CREATED_JOB_SITE_TIME_ID" != "null" ]]; then
    test_name "Update job site time description"
    xbe_json do job-site-times update "$CREATED_JOB_SITE_TIME_ID" --description "Updated job site time"
    assert_success

    test_name "Update job site time without fields fails"
    xbe_json do job-site-times update "$CREATED_JOB_SITE_TIME_ID"
    assert_failure
else
    test_name "Update job site time description"
    skip "No job site time created"
    test_name "Update job site time without fields fails"
    skip "No job site time created"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_JOB_SITE_TIME_ID" && "$CREATED_JOB_SITE_TIME_ID" != "null" ]]; then
    test_name "Show job site time"
    xbe_json view job-site-times show "$CREATED_JOB_SITE_TIME_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show job site time"
    fi
elif [[ -n "$SAMPLE_JOB_SITE_TIME_ID" ]]; then
    test_name "Show job site time (sample)"
    xbe_json view job-site-times show "$SAMPLE_JOB_SITE_TIME_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show job site time"
    fi
else
    test_name "Show job site time"
    skip "No job site time ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job site times with --job-production-plan filter"
if [[ -n "$JOB_PRODUCTION_PLAN_ID" ]]; then
    xbe_json view job-site-times list --job-production-plan "$JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List job site times with --user filter"
if [[ -n "$USER_ID" ]]; then
    xbe_json view job-site-times list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List job site times with --time-card filter"
xbe_json view job-site-times list --time-card 123 --limit 5
assert_success

test_name "List job site times with --start-at-min filter"
xbe_json view job-site-times list --start-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List job site times with --start-at-max filter"
xbe_json view job-site-times list --start-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List job site times with --is-start-at filter"
xbe_json view job-site-times list --is-start-at true --limit 5
assert_success

test_name "List job site times with --end-at-min filter"
xbe_json view job-site-times list --end-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List job site times with --end-at-max filter"
xbe_json view job-site-times list --end-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List job site times with --is-end-at filter"
xbe_json view job-site-times list --is-end-at true --limit 5
assert_success

test_name "List job site times with --created-at-min filter"
xbe_json view job-site-times list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List job site times with --created-at-max filter"
xbe_json view job-site-times list --created-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List job site times with --is-created-at filter"
xbe_json view job-site-times list --is-created-at true --limit 5
assert_success

test_name "List job site times with --updated-at-min filter"
xbe_json view job-site-times list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List job site times with --updated-at-max filter"
xbe_json view job-site-times list --updated-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List job site times with --is-updated-at filter"
xbe_json view job-site-times list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination / Sorting
# ============================================================================

test_name "List job site times with --limit"
xbe_json view job-site-times list --limit 3
assert_success

test_name "List job site times with --offset"
xbe_json view job-site-times list --limit 3 --offset 1
assert_success

test_name "List job site times with --sort"
xbe_json view job-site-times list --sort start-at --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_JOB_SITE_TIME_ID" && "$CREATED_JOB_SITE_TIME_ID" != "null" ]]; then
    test_name "Delete job site time requires --confirm flag"
    xbe_run do job-site-times delete "$CREATED_JOB_SITE_TIME_ID"
    assert_failure

    test_name "Delete job site time with --confirm"
    xbe_run do job-site-times delete "$CREATED_JOB_SITE_TIME_ID" --confirm
    assert_success
else
    test_name "Delete job site time requires --confirm flag"
    skip "No job site time created"
    test_name "Delete job site time with --confirm"
    skip "No job site time created"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job site time without job production plan fails"
xbe_json do job-site-times create --user 123 --start-at "2026-01-23T08:00:00Z"
assert_failure

test_name "Create job site time without user fails"
xbe_json do job-site-times create --job-production-plan 123 --start-at "2026-01-23T08:00:00Z"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
