#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Pre-Approvals
#
# Tests view and create/update/delete operations for time_card_pre_approvals.
#
# COVERAGE: Create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PRE_APPROVAL_ID=""
CANDIDATE_SHIFT_ID=""
CANDIDATE_SHIFT_START_AT=""
EXPLICIT_START_AT=""
EXPLICIT_END_AT=""
SAMPLE_PRE_APPROVAL_ID=""
SAMPLE_SHIFT_ID=""
SKIP_MUTATION=0

describe "Resource: time-card-pre-approvals"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card pre-approvals"
xbe_json view time-card-pre-approvals list --limit 1
assert_success

test_name "Capture sample pre-approval (if available)"
xbe_json view time-card-pre-approvals list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_PRE_APPROVAL_ID=$(json_get ".[0].id")
        SAMPLE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
        pass
    else
        echo "    No time card pre-approvals available; using fallback IDs for filter tests."
        pass
    fi
else
    fail "Failed to list time card pre-approvals"
fi

if [[ -n "$SAMPLE_PRE_APPROVAL_ID" && "$SAMPLE_PRE_APPROVAL_ID" != "null" ]]; then
    test_name "Show time card pre-approval"
    xbe_json view time-card-pre-approvals show "$SAMPLE_PRE_APPROVAL_ID"
    assert_success
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

SHIFT_FILTER="${SAMPLE_SHIFT_ID:-$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID}"
if [[ -n "$SHIFT_FILTER" && "$SHIFT_FILTER" != "null" ]]; then
    test_name "List time card pre-approvals with --tender-job-schedule-shift filter"
    xbe_json view time-card-pre-approvals list --tender-job-schedule-shift "$SHIFT_FILTER" --limit 5
    assert_success
else
    test_name "List time card pre-approvals with --tender-job-schedule-shift filter"
    skip "No shift ID available"
fi

# ============================================================================
# CREATE/UPDATE/DELETE Error Tests
# ============================================================================

test_name "Create time card pre-approval requires --tender-job-schedule-shift"
xbe_run do time-card-pre-approvals create
assert_failure

test_name "Update time card pre-approval without fields fails"
xbe_run do time-card-pre-approvals update "nonexistent"
assert_failure

test_name "Delete time card pre-approval requires --confirm flag"
xbe_run do time-card-pre-approvals delete "nonexistent"
assert_failure

# ============================================================================
# CREATE/UPDATE/DELETE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find tender job schedule shift without pre-approval"
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/tender-job-schedule-shifts" \
        --data-urlencode "page[limit]=25" \
        --data-urlencode "fields[tender-job-schedule-shifts]=start-at,time-card-pre-approval"

    http_code="$output"
    if [[ $status -eq 0 && "$http_code" == 2* ]]; then
        CANDIDATE_SHIFT_ID=$(jq -r '.data[] | select(.relationships["time-card-pre-approval"].data == null) | .id' "$response_file" | head -n 1)
        if [[ -n "$CANDIDATE_SHIFT_ID" && "$CANDIDATE_SHIFT_ID" != "null" ]]; then
            CANDIDATE_SHIFT_START_AT=$(jq -r ".data[] | select(.id==\"$CANDIDATE_SHIFT_ID\") | .attributes[\"start-at\"]" "$response_file")
            pass
        else
            skip "No shift without pre-approval found"
        fi
    else
        skip "Unable to list tender job schedule shifts (HTTP ${http_code})"
    fi
    rm -f "$response_file"
fi

if [[ -n "$CANDIDATE_SHIFT_ID" && "$CANDIDATE_SHIFT_ID" != "null" ]]; then
    if [[ -n "$CANDIDATE_SHIFT_START_AT" && "$CANDIDATE_SHIFT_START_AT" != "null" ]]; then
        EXPLICIT_START_AT="$CANDIDATE_SHIFT_START_AT"
        EXPLICIT_END_AT=$(EXPLICIT_START_AT="$EXPLICIT_START_AT" python3 - <<'PY'
import os
from datetime import datetime, timedelta

start = os.environ.get("EXPLICIT_START_AT", "").strip()
if not start:
    raise SystemExit(0)
if start.endswith("Z"):
    start = start[:-1] + "+00:00"
try:
    parsed = datetime.fromisoformat(start)
except ValueError:
    raise SystemExit(0)
end = parsed + timedelta(hours=8)
print(end.isoformat().replace("+00:00", "Z"))
PY
)
    fi

    test_name "Create time card pre-approval"
    NOTE="$(unique_name "PreApproval")"

    create_args=(do time-card-pre-approvals create \
        --tender-job-schedule-shift "$CANDIDATE_SHIFT_ID" \
        --maximum-quantities-attributes "[]" \
        --explicit-down-minutes 0 \
        --should-automatically-create-and-submit \
        --automatic-submission-delay-minutes 30 \
        --delay-automatic-submission-after-hours \
        --note "$NOTE" \
        --skip-quantity-validation)

    if [[ -n "$EXPLICIT_START_AT" && -n "$EXPLICIT_END_AT" ]]; then
        create_args+=(--explicit-start-at "$EXPLICIT_START_AT" --explicit-end-at "$EXPLICIT_END_AT")
    fi

    xbe_json "${create_args[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_PRE_APPROVAL_ID=$(json_get ".id")
        if [[ -n "$CREATED_PRE_APPROVAL_ID" && "$CREATED_PRE_APPROVAL_ID" != "null" ]]; then
            register_cleanup "time-card-pre-approvals" "$CREATED_PRE_APPROVAL_ID"
            pass
        else
            fail "Created pre-approval but no ID returned"
        fi
    else
        fail "Failed to create time card pre-approval"
    fi
else
    test_name "Create time card pre-approval"
    skip "No shift available for creation"
fi

if [[ -n "$CREATED_PRE_APPROVAL_ID" && "$CREATED_PRE_APPROVAL_ID" != "null" ]]; then
    test_name "Update time card pre-approval"
    UPDATED_NOTE="$(unique_name "PreApprovalUpdate")"
    xbe_json do time-card-pre-approvals update "$CREATED_PRE_APPROVAL_ID" --note "$UPDATED_NOTE"
    assert_success

    test_name "Show time card pre-approval reflects update"
    xbe_json view time-card-pre-approvals show "$CREATED_PRE_APPROVAL_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".note" "$UPDATED_NOTE"
    else
        fail "Failed to show time card pre-approval"
    fi

    test_name "Delete time card pre-approval"
    xbe_run do time-card-pre-approvals delete "$CREATED_PRE_APPROVAL_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
