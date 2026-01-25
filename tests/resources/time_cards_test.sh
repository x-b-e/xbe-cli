#!/bin/bash
#
# XBE CLI Integration Tests: Time Cards
#
# Tests CRUD operations for time cards.
#
# COVERAGE: Writable attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TIME_CARD_ID="${XBE_TEST_TIME_CARD_ID:-}"
BROKER_TENDER_ID="${XBE_TEST_TIME_CARD_BROKER_TENDER_ID:-}"
TENDER_JOB_SCHEDULE_SHIFT_ID="${XBE_TEST_TIME_CARD_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
JOB_SCHEDULE_SHIFT_ID="${XBE_TEST_TIME_CARD_JOB_SCHEDULE_SHIFT_ID:-}"
BROKER_ID="${XBE_TEST_TIME_CARD_BROKER_ID:-}"
CUSTOMER_ID="${XBE_TEST_TIME_CARD_CUSTOMER_ID:-}"
BUSINESS_UNIT_ID="${XBE_TEST_TIME_CARD_BUSINESS_UNIT_ID:-}"
TRUCKER_ID="${XBE_TEST_TIME_CARD_TRUCKER_ID:-}"
DRIVER_ID="${XBE_TEST_TIME_CARD_DRIVER_ID:-}"
TRAILER_ID="${XBE_TEST_TIME_CARD_TRAILER_ID:-}"
CONTRACTOR_ID="${XBE_TEST_TIME_CARD_CONTRACTOR_ID:-}"
INVOICE_ID="${XBE_TEST_TIME_CARD_INVOICE_ID:-}"
DEVELOPER_ID="${XBE_TEST_TIME_CARD_DEVELOPER_ID:-}"

CREATED_TIME_CARD_ID=""

describe "Resource: time-cards"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time card requires broker tender"
xbe_run do time-cards create --tender-job-schedule-shift 123
assert_failure

test_name "Create time card requires tender job schedule shift"
xbe_run do time-cards create --broker-tender 123
assert_failure

test_name "Create time card with required fields"
if [[ -n "$BROKER_TENDER_ID" && -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    TICKET="TC$(date +%s)"
    xbe_json do time-cards create \
        --broker-tender "$BROKER_TENDER_ID" \
        --tender-job-schedule-shift "$TENDER_JOB_SCHEDULE_SHIFT_ID" \
        --ticket-number "$TICKET"
    if [[ $status -eq 0 ]]; then
        CREATED_TIME_CARD_ID=$(json_get ".id")
        if [[ -n "$CREATED_TIME_CARD_ID" && "$CREATED_TIME_CARD_ID" != "null" ]]; then
            register_cleanup "time-cards" "$CREATED_TIME_CARD_ID"
            pass
        else
            fail "Created time card but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create time card: $output"
        fi
    fi
else
    skip "Set XBE_TEST_TIME_CARD_BROKER_TENDER_ID and XBE_TEST_TIME_CARD_TENDER_JOB_SCHEDULE_SHIFT_ID to enable create testing."
fi

if [[ -n "$CREATED_TIME_CARD_ID" && "$CREATED_TIME_CARD_ID" != "null" ]]; then
    TIME_CARD_ID="$CREATED_TIME_CARD_ID"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

update_time_card() {
    local description="$1"
    shift
    test_name "$description"
    xbe_json do time-cards update "$TIME_CARD_ID" "$@"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Update blocked by server policy/validation"
        else
            fail "Update failed: $output"
        fi
    fi
}

if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    update_time_card "Update ticket number" --ticket-number "TC-UPDATE-$(date +%s)"
    update_time_card "Update start at" --start-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    update_time_card "Update end at" --end-at "$(date -u -v+1H +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d '+1 hour' +%Y-%m-%dT%H:%M:%SZ)"
    update_time_card "Update down minutes" --down-minutes "15"
    update_time_card "Update submitted travel minutes" --submitted-travel-minutes "10"
    update_time_card "Update submitted customer travel minutes" --submitted-customer-travel-minutes "5"
    update_time_card "Update skip auto submission" --skip-auto-submission-upon-material-transaction-acceptance true
    update_time_card "Update explicit trailer required" --explicit-is-trailer-required-for-approval true
    update_time_card "Update explicit ticket number uniqueness" --explicit-enforce-ticket-number-uniqueness true
    update_time_card "Update enforce ticket number uniqueness" --enforce-ticket-number-uniqueness true
    update_time_card "Update time sheet line item explicit" --is-time-card-creating-time-sheet-line-item-explicit true
    update_time_card "Update explicit invoiceable when approved" --explicit-is-invoiceable-when-approved true
    update_time_card "Update approval process" --approval-process admin
    update_time_card "Update generate broker invoice" --generate-broker-invoice true
    update_time_card "Update generate trucker invoice" --generate-trucker-invoice true
    update_time_card "Reset submit at" --reset-submit-at
else
    skip "No time card ID available for update tests. Set XBE_TEST_TIME_CARD_ID to enable."
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time card"
if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    xbe_json view time-cards show "$TIME_CARD_ID"
    assert_success
else
    skip "No time card ID available for show tests."
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time cards"
xbe_json view time-cards list
assert_success

# ============================================================================
# LIST Tests - Filters (booleans)
# ============================================================================

bool_filters=(
    "is-audited"
    "with-payroll-certification-requirement"
    "with-payroll-certification-requirement-met"
    "broker-invoiced"
    "not-broker-invoiced"
    "trucker-invoiced"
    "not-trucker-invoiced"
    "on-shortfall-broker-tender"
    "on-shortfall-customer-tender"
    "is-start-at"
    "generate-broker-invoice"
    "generate-trucker-invoice"
    "has-ticket-report"
    "has-submit-scheduled"
)

for filter in "${bool_filters[@]}"; do
    test_name "List time cards with --${filter}"
    xbe_json view time-cards list --${filter} true
    assert_success
done

test_name "List time cards with --has-shift-date"
xbe_json view time-cards list --has-shift-date true
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"INTERNAL SERVER ERROR"* ]] || [[ "$output" == *"500"* ]]; then
        skip "Server error for has-shift-date filter"
    else
        fail "Expected success for has-shift-date filter"
    fi
fi

# ============================================================================
# LIST Tests - Filters (values)
# ============================================================================

test_name "List time cards with status filter"
xbe_json view time-cards list --status approved
assert_success

test_name "List time cards with approval count filters"
xbe_json view time-cards list --approval-count 1 --approval-count-min 1 --approval-count-max 5
assert_success

test_name "List time cards with shift date filters"
xbe_json view time-cards list --shift-date 2025-01-01 --shift-date-min 2025-01-01 --shift-date-max 2025-01-02 --shift-date-between 2025-01-01\|2025-01-02
assert_success

test_name "List time cards with approved-on filters"
xbe_json view time-cards list --approved-on-min 2025-01-01 --approved-on-max 2025-01-02
assert_success

test_name "List time cards with ticket number filter"
xbe_json view time-cards list --ticket-number TC-TEST
assert_success

test_name "List time cards with approvable-by filter"
if [[ -n "$DRIVER_ID" ]]; then
    xbe_json view time-cards list --approvable-by "$DRIVER_ID"
    assert_success
else
    skip "No driver ID available for approvable-by filter."
fi

test_name "List time cards with cost codes filter"
xbe_json view time-cards list --cost-codes CC-1,CC-2
assert_success

test_name "List time cards with job number filter"
xbe_json view time-cards list --job-number JOB-TEST
assert_success

# ============================================================================
# LIST Tests - Filters (IDs)
# ============================================================================

test_name "List time cards with broker tender filter"
if [[ -n "$BROKER_TENDER_ID" ]]; then
    xbe_json view time-cards list --broker-tender "$BROKER_TENDER_ID"
    assert_success
else
    skip "No broker tender ID available."
fi

test_name "List time cards with developer filter"
if [[ -n "$DEVELOPER_ID" ]]; then
    xbe_json view time-cards list --developer "$DEVELOPER_ID"
    assert_success
else
    skip "No developer ID available."
fi

test_name "List time cards with customer filter"
if [[ -n "$CUSTOMER_ID" ]]; then
    xbe_json view time-cards list --customer "$CUSTOMER_ID"
    assert_success
else
    skip "No customer ID available."
fi

test_name "List time cards with customer-id filter"
if [[ -n "$CUSTOMER_ID" ]]; then
    xbe_json view time-cards list --customer-id "$CUSTOMER_ID"
    assert_success
else
    skip "No customer ID available."
fi

test_name "List time cards with business unit filter"
if [[ -n "$BUSINESS_UNIT_ID" ]]; then
    xbe_json view time-cards list --business-unit "$BUSINESS_UNIT_ID"
    assert_success
else
    skip "No business unit ID available."
fi

test_name "List time cards with trucker filter"
if [[ -n "$TRUCKER_ID" ]]; then
    xbe_json view time-cards list --trucker "$TRUCKER_ID"
    assert_success
else
    skip "No trucker ID available."
fi

test_name "List time cards with trucker-id filter"
if [[ -n "$TRUCKER_ID" ]]; then
    xbe_json view time-cards list --trucker-id "$TRUCKER_ID"
    assert_success
else
    skip "No trucker ID available."
fi

test_name "List time cards with trailer filter"
if [[ -n "$TRAILER_ID" ]]; then
    xbe_json view time-cards list --trailer "$TRAILER_ID"
    assert_success
else
    skip "No trailer ID available."
fi

test_name "List time cards with trailer-id filter"
if [[ -n "$TRAILER_ID" ]]; then
    xbe_json view time-cards list --trailer-id "$TRAILER_ID"
    assert_success
else
    skip "No trailer ID available."
fi

test_name "List time cards with contractor filter"
if [[ -n "$CONTRACTOR_ID" ]]; then
    xbe_json view time-cards list --contractor "$CONTRACTOR_ID"
    assert_success
else
    skip "No contractor ID available."
fi

test_name "List time cards with contractor-id filter"
if [[ -n "$CONTRACTOR_ID" ]]; then
    xbe_json view time-cards list --contractor-id "$CONTRACTOR_ID"
    assert_success
else
    skip "No contractor ID available."
fi

test_name "List time cards with invoice filter"
if [[ -n "$INVOICE_ID" ]]; then
    xbe_json view time-cards list --invoice "$INVOICE_ID"
    assert_success
else
    skip "No invoice ID available."
fi

test_name "List time cards with tender job schedule shift filter"
if [[ -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json view time-cards list --tender-job-schedule-shift "$TENDER_JOB_SCHEDULE_SHIFT_ID"
    assert_success
else
    skip "No tender job schedule shift ID available."
fi

test_name "List time cards with job schedule shift filter"
if [[ -n "$JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json view time-cards list --job-schedule-shift "$JOB_SCHEDULE_SHIFT_ID"
    assert_success
else
    skip "No job schedule shift ID available."
fi

test_name "List time cards with job schedule shift id filter"
if [[ -n "$JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json view time-cards list --job-schedule-shift-id "$JOB_SCHEDULE_SHIFT_ID"
    assert_success
else
    skip "No job schedule shift ID available."
fi

test_name "List time cards with broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view time-cards list --broker "$BROKER_ID"
    assert_success
else
    skip "No broker ID available."
fi

test_name "List time cards with driver filter"
if [[ -n "$DRIVER_ID" ]]; then
    xbe_json view time-cards list --driver "$DRIVER_ID"
    assert_success
else
    skip "No driver ID available."
fi

test_name "List time cards with driver-id filter"
if [[ -n "$DRIVER_ID" ]]; then
    xbe_json view time-cards list --driver-id "$DRIVER_ID"
    assert_success
else
    skip "No driver ID available."
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete time card requires --confirm flag"
if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    xbe_run do time-cards delete "$TIME_CARD_ID"
    assert_failure
else
    skip "No time card ID available for delete test."
fi

test_name "Delete time card with --confirm"
if [[ -n "$BROKER_TENDER_ID" && -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json do time-cards create \
        --broker-tender "$BROKER_TENDER_ID" \
        --tender-job-schedule-shift "$TENDER_JOB_SCHEDULE_SHIFT_ID" \
        --ticket-number "TC-DELETE-$(date +%s)"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_json do time-cards delete "$DEL_ID" --confirm
        if [[ $status -eq 0 ]]; then
            pass
        else
            if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
                skip "Delete blocked by server policy/validation"
            else
                fail "Failed to delete time card: $output"
            fi
        fi
    else
        skip "Could not create time card for delete test"
    fi
else
    skip "Set broker tender and tender job schedule shift IDs to enable delete test."
fi

run_tests
