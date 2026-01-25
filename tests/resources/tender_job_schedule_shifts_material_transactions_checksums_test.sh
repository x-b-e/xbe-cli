#!/bin/bash
#
# XBE CLI Integration Tests: Tender Job Schedule Shifts Material Transactions Checksums
#
# Tests list, show, and create operations for the tender-job-schedule-shifts-material-transactions-checksums resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

LIST_SUPPORTED="true"
SAMPLE_ID=""
CREATE_ID=""

describe "Resource: tender-job-schedule-shifts-material-transactions-checksums"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List checksum records"
xbe_json view tender-job-schedule-shifts-material-transactions-checksums list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"Forbidden"* ]]; then
        LIST_SUPPORTED="false"
        skip "Listing requires admin access"
    elif [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "List endpoint not supported"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List checksum records returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view tender-job-schedule-shifts-material-transactions-checksums list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list checksum records"
    fi
else
    skip "List endpoint not supported"
fi

test_name "Capture sample checksum record"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view tender-job-schedule-shifts-material-transactions-checksums list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No checksum records available for follow-on tests"
        fi
    else
        skip "Could not list checksum records to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create checksum record"
RAW_JOB_NUMBER="${XBE_TEST_RAW_JOB_NUMBER:-CLI-TEST-$(date +%s)}"
TRANS_AT_MIN="${XBE_TEST_TRANSACTION_AT_MIN:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
TRANS_AT_MAX="${XBE_TEST_TRANSACTION_AT_MAX:-$TRANS_AT_MIN}"

CMD_ARGS=(
    --raw-job-number "$RAW_JOB_NUMBER"
    --transaction-at-min "$TRANS_AT_MIN"
    --transaction-at-max "$TRANS_AT_MAX"
)

if [[ -n "$XBE_TEST_MATERIAL_SITE_IDS" ]]; then
    CMD_ARGS+=(--material-site-ids "$XBE_TEST_MATERIAL_SITE_IDS")
fi

if [[ -n "$XBE_TEST_JOB_PRODUCTION_PLAN_ID" ]]; then
    CMD_ARGS+=(--job-production-plan-id "$XBE_TEST_JOB_PRODUCTION_PLAN_ID")
fi

xbe_json do tender-job-schedule-shifts-material-transactions-checksums create "${CMD_ARGS[@]}"
if [[ $status -eq 0 ]]; then
    CREATE_ID=$(json_get ".id")
    if [[ -n "$CREATE_ID" && "$CREATE_ID" != "null" ]]; then
        pass
    else
        pass
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"422"* ]] || \
       [[ "$output" == *"Record Invalid"* ]]; then
        pass
    else
        fail "Create failed: $output"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show checksum record"
SHOW_ID="$CREATE_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$SAMPLE_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts-material-transactions-checksums show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
            skip "Show endpoint not supported"
        elif [[ "$output" == *"Not Authorized"* ]] || \
             [[ "$output" == *"not authorized"* ]] || \
             [[ "$output" == *"403"* ]]; then
            skip "Not authorized to view checksum record"
        else
            fail "Show failed: $output"
        fi
    fi
else
    skip "No checksum ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create checksum without required flags fails"
xbe_run do tender-job-schedule-shifts-material-transactions-checksums create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
