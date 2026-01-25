#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Settings
#
# Tests operations for the trucker-settings resource.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TRUCKER_SETTING_ID=""
TRUCKER_ID=""
TIME_SHEET_CLASSIFICATION_ID=""

describe "Resource: trucker-settings"

# ============================================================================
# LIST Tests - Get a trucker setting ID for update tests
# ============================================================================

test_name "List trucker settings"
xbe_json view trucker-settings list --limit 5
assert_success

test_name "List trucker settings returns array"
xbe_json view trucker-settings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker settings"
fi

test_name "Get a trucker setting ID for update tests"
xbe_json view trucker-settings list --limit 1
if [[ $status -eq 0 ]]; then
    TRUCKER_SETTING_ID=$(json_get ".[0].id")
    TRUCKER_ID=$(json_get ".[0].trucker_id")
    if [[ -n "$TRUCKER_SETTING_ID" && "$TRUCKER_SETTING_ID" != "null" ]]; then
        pass
        if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
            xbe_json view trucker-settings show "$TRUCKER_SETTING_ID"
            if [[ $status -eq 0 ]]; then
                TRUCKER_ID=$(json_get ".trucker_id")
            fi
        fi
    else
        skip "No trucker settings found in the system"
        run_tests
    fi
else
    fail "Failed to list trucker settings"
    run_tests
fi

# ============================================================================
# FILTER Tests
# ============================================================================

if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    test_name "List trucker settings filtered by trucker"
    xbe_json view trucker-settings list --trucker "$TRUCKER_ID"
    assert_success

    test_name "List trucker settings filtered by trucker-id"
    xbe_json view trucker-settings list --trucker-id "$TRUCKER_ID"
    assert_success
else
    test_name "Skip trucker filter tests (missing trucker ID)"
    skip "No trucker ID available from list"
fi

# ============================================================================
# UPDATE Tests - Gather required related IDs
# ============================================================================

test_name "Get time sheet line item classification ID"
xbe_json view time-sheet-line-item-classifications list --limit 1
if [[ $status -eq 0 ]]; then
    TIME_SHEET_CLASSIFICATION_ID=$(json_get ".[0].id")
    if [[ -n "$TIME_SHEET_CLASSIFICATION_ID" && "$TIME_SHEET_CLASSIFICATION_ID" != "null" ]]; then
        pass
    else
        skip "No time sheet line item classifications found"
    fi
else
    skip "Failed to list time sheet line item classifications"
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update trucker settings billing period fields"
xbe_json do trucker-settings update "$TRUCKER_SETTING_ID" \
    --billing-period-start-on 2025-01-01 \
    --billing-period-day-count 30 \
    --billing-period-end-invoice-offset-day-count 0
assert_success

test_name "Update trucker settings invoice flags"
xbe_json do trucker-settings update "$TRUCKER_SETTING_ID" \
    --split-billing-periods-spanning-months true \
    --group-daily-invoice-by-job-number true \
    --generate-daily-invoice true \
    --deliver-new-invoices true \
    --invoices-batch-processing-start-on 2025-01-02 \
    --invoices-grouped-by-time-card-start-date true \
    --invoice-date-calculation average
assert_success

test_name "Update trucker settings general fields"
xbe_json do trucker-settings update "$TRUCKER_SETTING_ID" \
    --create-detected-production-incidents true \
    --primary-color "#112233" \
    --logo-svg '<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"></svg>' \
    --default-tender-dispatch-instructions "Dispatch instructions" \
    --default-payment-terms-and-conditions "Net 30" \
    --default-hours-after-which-overtime-applies 8 \
    --sets-shift-material-transaction-expectations true \
    --show-planner-info-to-drivers true \
    --is-trucker-shift-rejection-permitted true
assert_success

test_name "Update trucker settings shift fields"
xbe_json do trucker-settings update "$TRUCKER_SETTING_ID" \
    --notify-driver-when-gps-not-available true \
    --day-shift-assignment-reminder-time 08:00 \
    --night-shift-assignment-reminder-time 18:00 \
    --minimum-driver-tracking-minutes 15 \
    --auto-generate-time-sheet-line-items-per-job true
assert_success

if [[ -n "$TIME_SHEET_CLASSIFICATION_ID" && "$TIME_SHEET_CLASSIFICATION_ID" != "null" ]]; then
    test_name "Update trucker settings time sheet fields"
    xbe_json do trucker-settings update "$TRUCKER_SETTING_ID" \
        --restrict-line-item-classification-edit-to-time-sheet-approver true \
        --default-time-sheet-line-item-classification-id "$TIME_SHEET_CLASSIFICATION_ID" \
        --auto-combine-overlapping-driver-days true
    assert_success
else
    test_name "Skip time sheet line item classification update"
    skip "Missing time sheet line item classification ID"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List trucker settings with --limit"
xbe_json view trucker-settings list --limit 3
assert_success

test_name "List trucker settings with --offset"
xbe_json view trucker-settings list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do trucker-settings update "$TRUCKER_SETTING_ID"
assert_failure

test_name "Update non-existent trucker settings fails"
xbe_json do trucker-settings update "99999999" --notify-driver-when-gps-not-available true
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
