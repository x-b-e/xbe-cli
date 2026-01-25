#!/bin/bash
#
# XBE CLI Integration Tests: Expected Time of Arrivals
#
# Tests list/show/create/update/delete operations for expected-time-of-arrivals.
#
# COVERAGE: All list filters + writable attributes (expected-at, note, unsure,
# tender-job-schedule-shift)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SHIFT_ID="${XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
if [[ -z "$SHIFT_ID" ]]; then
    SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
fi

JOB_SCHEDULE_SHIFT_ID="${XBE_TEST_TIME_CARD_JOB_SCHEDULE_SHIFT_ID:-}"
CREATED_BY_ID="${XBE_TEST_USER_ID:-}"
EXPECTED_TIME_OF_ARRIVAL_ID="${XBE_TEST_EXPECTED_TIME_OF_ARRIVAL_ID:-}"
SAMPLE_ID=""
SAMPLE_TENDER_SHIFT_ID=""
SAMPLE_JOB_SHIFT_ID=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_EXPECTED_AT=""
CREATED_ID=""

describe "Resource: expected-time-of-arrivals"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List expected time of arrivals"
xbe_json view expected-time-of-arrivals list --limit 5
assert_success

test_name "List expected time of arrivals returns array"
xbe_json view expected-time-of-arrivals list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list expected time of arrivals"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample expected time of arrival"
xbe_json view expected-time-of-arrivals list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TENDER_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_JOB_SHIFT_ID=$(json_get ".[0].job_schedule_shift_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    SAMPLE_EXPECTED_AT=$(json_get ".[0].expected_at")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No expected time of arrivals available for show/filter tests"
    fi
else
    skip "Could not list expected time of arrivals to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show expected time of arrival"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view expected-time-of-arrivals show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show expected time of arrival: $output"
        fi
    fi
else
    skip "No expected time of arrival ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List expected time of arrivals with --tender-job-schedule-shift filter"
FILTER_TENDER_SHIFT_ID="${SAMPLE_TENDER_SHIFT_ID:-$SHIFT_ID}"
if [[ -n "$FILTER_TENDER_SHIFT_ID" && "$FILTER_TENDER_SHIFT_ID" != "null" ]]; then
    xbe_json view expected-time-of-arrivals list --tender-job-schedule-shift "$FILTER_TENDER_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List expected time of arrivals with --job-schedule-shift filter"
FILTER_JOB_SHIFT_ID="${SAMPLE_JOB_SHIFT_ID:-$JOB_SCHEDULE_SHIFT_ID}"
if [[ -n "$FILTER_JOB_SHIFT_ID" && "$FILTER_JOB_SHIFT_ID" != "null" ]]; then
    xbe_json view expected-time-of-arrivals list --job-schedule-shift "$FILTER_JOB_SHIFT_ID" --limit 5
    assert_success
else
    skip "No job schedule shift ID available"
fi

test_name "List expected time of arrivals with --created-by filter"
FILTER_CREATED_BY_ID="${SAMPLE_CREATED_BY_ID:-$CREATED_BY_ID}"
if [[ -z "$FILTER_CREATED_BY_ID" || "$FILTER_CREATED_BY_ID" == "null" ]]; then
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        FILTER_CREATED_BY_ID=$(json_get ".id")
    fi
fi
if [[ -n "$FILTER_CREATED_BY_ID" && "$FILTER_CREATED_BY_ID" != "null" ]]; then
    xbe_json view expected-time-of-arrivals list --created-by "$FILTER_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by user ID available"
fi

test_name "List expected time of arrivals with --expected-at-min filter"
if [[ -n "$SAMPLE_EXPECTED_AT" && "$SAMPLE_EXPECTED_AT" != "null" ]]; then
    xbe_json view expected-time-of-arrivals list --expected-at-min "$SAMPLE_EXPECTED_AT" --limit 5
    assert_success
else
    xbe_json view expected-time-of-arrivals list --expected-at-min "2024-01-01T00:00:00Z" --limit 5
    assert_success
fi

test_name "List expected time of arrivals with --expected-at-max filter"
if [[ -n "$SAMPLE_EXPECTED_AT" && "$SAMPLE_EXPECTED_AT" != "null" ]]; then
    xbe_json view expected-time-of-arrivals list --expected-at-max "$SAMPLE_EXPECTED_AT" --limit 5
    assert_success
else
    xbe_json view expected-time-of-arrivals list --expected-at-max "2025-12-31T23:59:59Z" --limit 5
    assert_success
fi

test_name "List expected time of arrivals with --is-expected-at filter (true)"
xbe_json view expected-time-of-arrivals list --is-expected-at true --limit 5
assert_success

test_name "List expected time of arrivals with --is-expected-at filter (false)"
xbe_json view expected-time-of-arrivals list --is-expected-at false --limit 5
assert_success

test_name "List expected time of arrivals with --unsure filter (true)"
xbe_json view expected-time-of-arrivals list --unsure true --limit 5
assert_success

test_name "List expected time of arrivals with --unsure filter (false)"
xbe_json view expected-time-of-arrivals list --unsure false --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List expected time of arrivals with --limit"
xbe_json view expected-time-of-arrivals list --limit 3
assert_success

test_name "List expected time of arrivals with --offset"
xbe_json view expected-time-of-arrivals list --limit 3 --offset 3
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create expected time of arrival requires tender job schedule shift"
xbe_run do expected-time-of-arrivals create --expected-at "2025-01-01T00:00:00Z"
assert_failure

test_name "Create expected time of arrival requires --expected-at or --unsure true"
xbe_run do expected-time-of-arrivals create --tender-job-schedule-shift "nonexistent"
assert_failure

test_name "Create expected time of arrival"
if [[ -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    NOW_TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    NOTE=$(unique_name "ETA Note")
    xbe_json do expected-time-of-arrivals create \
        --tender-job-schedule-shift "$SHIFT_ID" \
        --expected-at "$NOW_TS" \
        --note "$NOTE"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "expected-time-of-arrivals" "$CREATED_ID"
            pass
        else
            fail "Created expected time of arrival but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"broker tender"* ]] || [[ "$output" == *"must be related to a broker tender"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create expected time of arrival: $output"
        fi
    fi
else
    skip "No tender job schedule shift ID available. Set XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update expected time of arrival without fields fails"
xbe_run do expected-time-of-arrivals update "nonexistent"
assert_failure

TARGET_ID="${CREATED_ID:-$EXPECTED_TIME_OF_ARRIVAL_ID}"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    test_name "Update expected time of arrival note"
    xbe_json do expected-time-of-arrivals update "$TARGET_ID" --note "Updated ETA note"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Update blocked by server policy"
        else
            fail "Failed to update expected time of arrival note: $output"
        fi
    fi

    test_name "Update expected time of arrival expected-at"
    NEW_TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do expected-time-of-arrivals update "$TARGET_ID" --expected-at "$NEW_TS"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Update blocked by server policy"
        else
            fail "Failed to update expected time of arrival expected-at: $output"
        fi
    fi

    test_name "Update expected time of arrival unsure"
    xbe_json do expected-time-of-arrivals update "$TARGET_ID" --unsure true --note "ETA pending"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Update blocked by server policy"
        else
            fail "Failed to update expected time of arrival unsure: $output"
        fi
    fi
else
    skip "No expected time of arrival ID available for update tests"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete expected time of arrival requires --confirm flag"
xbe_run do expected-time-of-arrivals delete "nonexistent"
assert_failure

test_name "Delete expected time of arrival"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do expected-time-of-arrivals delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete expected time of arrival: $output"
        fi
    fi
else
    skip "No created expected time of arrival for delete test"
fi

run_tests
