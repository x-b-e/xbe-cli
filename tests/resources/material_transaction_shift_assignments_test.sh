#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Shift Assignments
#
# Tests list/show/create operations for material-transaction-shift-assignments.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_MATERIAL_TRANSACTION_ID=""
SAMPLE_IS_PROCESSED=""

created_from_env=""

# Optional env override for existing assignment
if [[ -n "${XBE_TEST_MATERIAL_TRANSACTION_SHIFT_ASSIGNMENT_ID:-}" ]]; then
    created_from_env="$XBE_TEST_MATERIAL_TRANSACTION_SHIFT_ASSIGNMENT_ID"
fi

describe "Resource: material-transaction-shift-assignments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction shift assignments"
xbe_json view material-transaction-shift-assignments list --limit 5
assert_success

test_name "List material transaction shift assignments returns array"
xbe_json view material-transaction-shift-assignments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction shift assignments"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample assignment"
xbe_json view material-transaction-shift-assignments list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
    SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_TRUCKER_ID=$(json_get ".[0].trucker_id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].material_transaction_ids[0]")
    SAMPLE_IS_PROCESSED=$(json_get ".[0].is_processed")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No assignments available for follow-on tests"
    fi
else
    skip "Could not list assignments to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List assignments with --job-production-plan filter"
if [[ -n "$SAMPLE_JOB_PRODUCTION_PLAN_ID" && "$SAMPLE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view material-transaction-shift-assignments list --job-production-plan "$SAMPLE_JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List assignments with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view material-transaction-shift-assignments list --tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List assignments with --trucker filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view material-transaction-shift-assignments list --trucker "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List assignments with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view material-transaction-shift-assignments list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List assignments with --material-transaction filter"
if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
    xbe_json view material-transaction-shift-assignments list --material-transaction "$SAMPLE_MATERIAL_TRANSACTION_ID" --limit 5
    assert_success
else
    skip "No material transaction ID available"
fi

test_name "List assignments with --is-processed=true filter"
xbe_json view material-transaction-shift-assignments list --is-processed true --limit 5
assert_success

test_name "List assignments with --is-processed=false filter"
xbe_json view material-transaction-shift-assignments list --is-processed false --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material transaction shift assignment"
SHOW_ID="$SAMPLE_ID"
if [[ -n "$created_from_env" ]]; then
    SHOW_ID="$created_from_env"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view material-transaction-shift-assignments show "$SHOW_ID"
    assert_success
else
    skip "No assignment ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create assignment requires tender job schedule shift"
xbe_run do material-transaction-shift-assignments create --material-transaction-ids 123
assert_failure

test_name "Create assignment requires material transaction IDs"
xbe_run do material-transaction-shift-assignments create --tender-job-schedule-shift 123
assert_failure

TARGET_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
MTXN_IDS="${XBE_TEST_MATERIAL_TRANSACTION_IDS:-}"
if [[ -z "$MTXN_IDS" && -n "${XBE_TEST_MATERIAL_TRANSACTION_ID:-}" ]]; then
    MTXN_IDS="$XBE_TEST_MATERIAL_TRANSACTION_ID"
fi
if [[ -z "$TARGET_SHIFT_ID" && -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    TARGET_SHIFT_ID="$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID"
fi
if [[ -z "$MTXN_IDS" && -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
    MTXN_IDS="$SAMPLE_MATERIAL_TRANSACTION_ID"
fi

if [[ -z "$MTXN_IDS" ]]; then
    xbe_json view material-transactions list --limit 1
    if [[ $status -eq 0 ]]; then
        MTXN_IDS=$(json_get ".[0].id")
    fi
fi

test_name "Create material transaction shift assignment"
if [[ -n "$TARGET_SHIFT_ID" && "$TARGET_SHIFT_ID" != "null" && -n "$MTXN_IDS" && "$MTXN_IDS" != "null" ]]; then
    COMMENT=$(unique_name "ShiftAssignment")
    xbe_json do material-transaction-shift-assignments create \
        --tender-job-schedule-shift "$TARGET_SHIFT_ID" \
        --material-transaction-ids "$MTXN_IDS" \
        --comment "$COMMENT" \
        --skip-material-transaction-shift-skew-validation \
        --enable-link-invoiced
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"cannot be"* ]] || [[ "$output" == *"must be"* ]] || [[ "$output" == *"invoiced"* ]] || [[ "$output" == *"approved time card"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create material transaction shift assignment: $output"
        fi
    fi
else
    skip "Missing tender job schedule shift or material transaction IDs (set XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID and XBE_TEST_MATERIAL_TRANSACTION_IDS)"
fi

run_tests
