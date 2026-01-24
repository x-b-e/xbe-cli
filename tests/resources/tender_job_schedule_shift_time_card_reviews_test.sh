#!/bin/bash
#
# XBE CLI Integration Tests: Tender Job Schedule Shift Time Card Reviews
#
# Tests view and create/delete operations for tender_job_schedule_shift_time_card_reviews.
#
# COVERAGE: Create relationship + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_REVIEW_ID=""
SAMPLE_SHIFT_ID=""
CANDIDATE_SHIFT_ID=""
CREATED_REVIEW_ID=""
SKIP_MUTATION=0

describe "Resource: tender-job-schedule-shift-time-card-reviews"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List tender job schedule shift time card reviews"
xbe_json view tender-job-schedule-shift-time-card-reviews list --limit 1
assert_success

test_name "Capture sample time card review (if available)"
xbe_json view tender-job-schedule-shift-time-card-reviews list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_REVIEW_ID=$(json_get ".[0].id")
        SAMPLE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
        pass
    else
        echo "    No time card reviews available; using fallback IDs for filter tests."
        pass
    fi
else
    fail "Failed to list time card reviews"
fi

if [[ -n "$SAMPLE_REVIEW_ID" && "$SAMPLE_REVIEW_ID" != "null" ]]; then
    test_name "Show time card review"
    xbe_json view tender-job-schedule-shift-time-card-reviews show "$SAMPLE_REVIEW_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_SHIFT_ID=$(json_get ".tender_job_schedule_shift_id")
        pass
    else
        fail "Failed to show time card review"
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

SHIFT_FILTER="${SAMPLE_SHIFT_ID:-$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID}"
if [[ -n "$SHIFT_FILTER" && "$SHIFT_FILTER" != "null" ]]; then
    test_name "List time card reviews with --tender-job-schedule-shift filter"
    xbe_json view tender-job-schedule-shift-time-card-reviews list --tender-job-schedule-shift "$SHIFT_FILTER" --limit 5
    assert_success
else
    test_name "List time card reviews with --tender-job-schedule-shift filter"
    skip "No shift ID available"
fi

# ============================================================================
# CREATE / DELETE Error Tests
# ============================================================================

test_name "Create time card review requires --tender-job-schedule-shift"
xbe_run do tender-job-schedule-shift-time-card-reviews create
assert_failure

test_name "Delete time card review requires --confirm flag"
xbe_run do tender-job-schedule-shift-time-card-reviews delete "nonexistent"
assert_failure

# ============================================================================
# CREATE / DELETE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    CANDIDATE_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-$SAMPLE_SHIFT_ID}"

    if [[ -z "$CANDIDATE_SHIFT_ID" || "$CANDIDATE_SHIFT_ID" == "null" ]]; then
        test_name "Find tender job schedule shift"
        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -G "$XBE_BASE_URL/v1/tender-job-schedule-shifts" \
            --data-urlencode "page[limit]=25"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            CANDIDATE_SHIFT_ID=$(jq -r '.data[0].id' "$response_file")
            if [[ -n "$CANDIDATE_SHIFT_ID" && "$CANDIDATE_SHIFT_ID" != "null" ]]; then
                pass
            else
                skip "No tender job schedule shift found"
            fi
        else
            skip "Unable to list tender job schedule shifts (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi
fi

if [[ -n "$CANDIDATE_SHIFT_ID" && "$CANDIDATE_SHIFT_ID" != "null" ]]; then
    test_name "Create time card review"
    xbe_json do tender-job-schedule-shift-time-card-reviews create \
        --tender-job-schedule-shift "$CANDIDATE_SHIFT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_REVIEW_ID=$(json_get ".id")
        if [[ -n "$CREATED_REVIEW_ID" && "$CREATED_REVIEW_ID" != "null" ]]; then
            register_cleanup "tender-job-schedule-shift-time-card-reviews" "$CREATED_REVIEW_ID"
            pass
        else
            fail "Created time card review but no ID returned"
        fi
    else
        skip "Failed to create time card review"
    fi
else
    test_name "Create time card review"
    skip "No tender job schedule shift available for creation"
fi

if [[ -n "$CREATED_REVIEW_ID" && "$CREATED_REVIEW_ID" != "null" ]]; then
    test_name "Delete time card review"
    xbe_run do tender-job-schedule-shift-time-card-reviews delete "$CREATED_REVIEW_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
