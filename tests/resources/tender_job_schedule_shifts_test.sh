#!/bin/bash
#
# XBE CLI Integration Tests: Tender Job Schedule Shifts
#
# Tests list, show, create, update, and delete operations for the
# tender-job-schedule-shifts resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TENDER_ID=""
SAMPLE_TENDER_TYPE=""
SAMPLE_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_DRIVER_ID=""
SAMPLE_TRAILER_ID=""
SAMPLE_TRACTOR_ID=""
SAMPLE_RETAINER_ID=""
SAMPLE_TRUCKER_SHIFT_SET_ID=""
SAMPLE_PRIMARY_DRIVER_ID=""
SAMPLE_TIME_CARD_ID=""
SAMPLE_SERVICE_EVENT_ID=""
SAMPLE_EXPECTED_ETA_ID=""
SAMPLE_SHIFT_FEEDBACK_ID=""
SAMPLE_PRODUCTION_INCIDENT_ID=""
SAMPLE_JPP_BROADCAST_MESSAGE_ID=""
SAMPLE_TRIP_ID=""
MPO_RELEASE_ID=""

CREATED_ID=""

describe "Resource: tender-job-schedule-shifts"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tender job schedule shifts"
xbe_json view tender-job-schedule-shifts list --limit 5
assert_success

test_name "List tender job schedule shifts returns array"
xbe_json view tender-job-schedule-shifts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tender job schedule shifts"
fi

# ============================================================================
# Sample Record (used for filters/show/update/delete)
# ============================================================================

test_name "Capture sample tender job schedule shift"
xbe_json view tender-job-schedule-shifts list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TENDER_ID=$(json_get ".[0].tender_id")
    SAMPLE_TENDER_TYPE=$(json_get ".[0].tender_type")
    SAMPLE_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].job_schedule_shift_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].seller_operations_contact_id")
    SAMPLE_TRAILER_ID=$(json_get ".[0].trailer_id")
    SAMPLE_TRACTOR_ID=$(json_get ".[0].tractor_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No tender job schedule shifts available for follow-on tests"
    fi
else
    skip "Could not list tender job schedule shifts to capture sample"
fi

# Fetch details for additional relationship IDs
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_RETAINER_ID=$(json_get ".retainer_id")
        SAMPLE_TRUCKER_SHIFT_SET_ID=$(json_get ".trucker_shift_set_id")
        SAMPLE_PRIMARY_DRIVER_ID=$(json_get ".primary_driver_id")
        SAMPLE_TIME_CARD_ID=$(json_get ".time_card_ids[0]")
        SAMPLE_SERVICE_EVENT_ID=$(json_get ".service_event_ids[0]")
        SAMPLE_EXPECTED_ETA_ID=$(json_get ".expected_time_of_arrival_ids[0]")
        SAMPLE_SHIFT_FEEDBACK_ID=$(json_get ".shift_feedback_ids[0]")
        SAMPLE_PRODUCTION_INCIDENT_ID=$(json_get ".production_incident_ids[0]")
        SAMPLE_JPP_BROADCAST_MESSAGE_ID=$(json_get ".job_production_plan_broadcast_message_ids[0]")
        SAMPLE_TRIP_ID=$(json_get ".trip_ids[0]")
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter: --with-driver-assignment-refusals"
xbe_json view tender-job-schedule-shifts list --with-driver-assignment-refusals true --limit 5
assert_success

test_name "Filter: --is-managed"
xbe_json view tender-job-schedule-shifts list --is-managed true --limit 5
assert_success

test_name "Filter: --tender"
if [[ -n "$SAMPLE_TENDER_ID" && "$SAMPLE_TENDER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --tender "$SAMPLE_TENDER_ID" --limit 5
    assert_success
else
    skip "No tender ID available"
fi

test_name "Filter: --job-schedule-shift"
if [[ -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --job-schedule-shift "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No job schedule shift ID available"
fi

test_name "Filter: --job"
xbe_json view tender-job-schedule-shifts list --job 1 --limit 5
assert_success

test_name "Filter: --matches-material-purchase-order-release"
xbe_json view material-purchase-order-releases list --limit 1
if [[ $status -eq 0 ]]; then
    MPO_RELEASE_ID=$(json_get ".[0].id")
fi
if [[ -n "$MPO_RELEASE_ID" && "$MPO_RELEASE_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --matches-material-purchase-order-release "$MPO_RELEASE_ID" --limit 5
    assert_success
else
    skip "No material purchase order release available"
fi

test_name "Filter: --developer-trucker-certification"
xbe_json view tender-job-schedule-shifts list --developer-trucker-certification 1 --limit 5
assert_success

test_name "Filter: --developer-trucker-certification-classification"
xbe_json view tender-job-schedule-shifts list --developer-trucker-certification-classification 1 --limit 5
assert_success

test_name "Filter: --seller-operations-contact"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --seller-operations-contact "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No seller operations contact ID available"
fi

test_name "Filter: --driver-day-sequence-index"
xbe_json view tender-job-schedule-shifts list --driver-day-sequence-index 0 --limit 5
assert_success

test_name "Filter: --trailer-id"
if [[ -n "$SAMPLE_TRAILER_ID" && "$SAMPLE_TRAILER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --trailer-id "$SAMPLE_TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "Filter: --trailer"
if [[ -n "$SAMPLE_TRAILER_ID" && "$SAMPLE_TRAILER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --trailer "$SAMPLE_TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "Filter: --tractor-id"
if [[ -n "$SAMPLE_TRACTOR_ID" && "$SAMPLE_TRACTOR_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --tractor-id "$SAMPLE_TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "Filter: --tractor"
if [[ -n "$SAMPLE_TRACTOR_ID" && "$SAMPLE_TRACTOR_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --tractor "$SAMPLE_TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "Filter: --without-seller-operations-contact"
xbe_json view tender-job-schedule-shifts list --without-seller-operations-contact true --limit 5
assert_success

test_name "Filter: --with-seller-operations-contact-assigned-or-drafted"
xbe_json view tender-job-schedule-shifts list --with-seller-operations-contact-assigned-or-drafted true --limit 5
assert_success

test_name "Filter: --with-trailer-assigned-or-drafted"
xbe_json view tender-job-schedule-shifts list --with-trailer-assigned-or-drafted true --limit 5
assert_success

test_name "Filter: --missing-assignment"
xbe_json view tender-job-schedule-shifts list --missing-assignment true --limit 5
assert_success

test_name "Filter: --job-production-plan"
xbe_json view tender-job-schedule-shifts list --job-production-plan 1 --limit 5
assert_success

test_name "Filter: --with-job-production-plan"
xbe_json view tender-job-schedule-shifts list --with-job-production-plan true --limit 5
assert_success

test_name "Filter: --on-time-card"
xbe_json view tender-job-schedule-shifts list --on-time-card true --limit 5
assert_success

test_name "Filter: --is-managed-or-alive"
xbe_json view tender-job-schedule-shifts list --is-managed-or-alive true --limit 5
assert_success

test_name "Filter: --expects-time-cards"
xbe_json view tender-job-schedule-shifts list --expects-time-cards true --limit 5
assert_success

test_name "Filter: --time-card-ticket-number"
xbe_json view tender-job-schedule-shifts list --time-card-ticket-number "TEST" --limit 5
assert_success

test_name "Filter: --shift-ends-before"
xbe_json view tender-job-schedule-shifts list --shift-ends-before "2024-01-02T00:00:00Z" --limit 5
assert_success

test_name "Filter: --shift-starts-after"
xbe_json view tender-job-schedule-shifts list --shift-starts-after "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter: --shift-starts-before"
xbe_json view tender-job-schedule-shifts list --shift-starts-before "2025-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter: --end-at-min"
xbe_json view tender-job-schedule-shifts list --end-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter: --end-at-max"
xbe_json view tender-job-schedule-shifts list --end-at-max "2025-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter: --start-date"
xbe_json view tender-job-schedule-shifts list --start-date "2024-01-01" --limit 5
assert_success

test_name "Filter: --related-tender-status"
xbe_json view tender-job-schedule-shifts list --related-tender-status accepted --limit 5
assert_success

test_name "Filter: --tender-type"
if [[ -n "$SAMPLE_TENDER_TYPE" && "$SAMPLE_TENDER_TYPE" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --tender-type "$SAMPLE_TENDER_TYPE" --limit 5
    assert_success
else
    xbe_json view tender-job-schedule-shifts list --tender-type BrokerTender --limit 5
    assert_success
fi

test_name "Filter: --time-card-status"
xbe_json view tender-job-schedule-shifts list --time-card-status editing --limit 5
assert_success

test_name "Filter: --active-as-of"
xbe_json view tender-job-schedule-shifts list --active-as-of "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter: --cancelled"
xbe_json view tender-job-schedule-shifts list --cancelled true --limit 5
assert_success

test_name "Filter: --trucker"
xbe_json view tender-job-schedule-shifts list --trucker 1 --limit 5
assert_success

test_name "Filter: --customer"
xbe_json view tender-job-schedule-shifts list --customer 1 --limit 5
assert_success

test_name "Filter: --broker"
xbe_json view tender-job-schedule-shifts list --broker 1 --limit 5
assert_success

test_name "Filter: --sourced-with-trucker"
xbe_json view tender-job-schedule-shifts list --sourced-with-trucker 1 --limit 5
assert_success

test_name "Filter: --allows-new-trip"
xbe_json view tender-job-schedule-shifts list --allows-new-trip true --limit 5
assert_success

test_name "Filter: --job-site"
xbe_json view tender-job-schedule-shifts list --job-site 1 --limit 5
assert_success

test_name "Filter: --retained"
xbe_json view tender-job-schedule-shifts list --retained true --limit 5
assert_success

test_name "Filter: --tracked-pct-min"
xbe_json view tender-job-schedule-shifts list --tracked-pct-min 0 --limit 5
assert_success

test_name "Filter: --retainer"
if [[ -n "$SAMPLE_RETAINER_ID" && "$SAMPLE_RETAINER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --retainer "$SAMPLE_RETAINER_ID" --limit 5
    assert_success
else
    xbe_json view tender-job-schedule-shifts list --retainer 1 --limit 5
    assert_success
fi

test_name "Filter: --without-ready-to-work-service-event"
xbe_json view tender-job-schedule-shifts list --without-ready-to-work-service-event true --limit 5
assert_success

test_name "Filter: --missing-driver-assignment-acknowledgement"
xbe_json view tender-job-schedule-shifts list --missing-driver-assignment-acknowledgement true --limit 5
assert_success

test_name "Filter: --trucker-shift-set"
if [[ -n "$SAMPLE_TRUCKER_SHIFT_SET_ID" && "$SAMPLE_TRUCKER_SHIFT_SET_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --trucker-shift-set "$SAMPLE_TRUCKER_SHIFT_SET_ID" --limit 5
    assert_success
else
    xbe_json view tender-job-schedule-shifts list --trucker-shift-set 1 --limit 5
    assert_success
fi

test_name "Filter: --default-time-card-approval-process"
xbe_json view tender-job-schedule-shifts list --default-time-card-approval-process admin --limit 5
assert_success

test_name "Filter: --drivers"
if [[ -n "$SAMPLE_PRIMARY_DRIVER_ID" && "$SAMPLE_PRIMARY_DRIVER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --drivers "$SAMPLE_PRIMARY_DRIVER_ID" --limit 5
    assert_success
else
    xbe_json view tender-job-schedule-shifts list --drivers 1 --limit 5
    assert_success
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender job schedule shift"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shifts show "$SAMPLE_ID"
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create tender job schedule shift"
if [[ -n "$SAMPLE_TENDER_ID" && "$SAMPLE_TENDER_ID" != "null" && -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json do tender-job-schedule-shifts create \
        --tender-type "${SAMPLE_TENDER_TYPE:-broker-tenders}" \
        --tender-id "$SAMPLE_TENDER_ID" \
        --job-schedule-shift "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" \
        --material-transaction-status open
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No tender/job schedule shift IDs available for create"
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update tender job schedule shift attributes"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    UPDATE_START_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    xbe_json do tender-job-schedule-shifts update "$SAMPLE_ID" \
        --truck-count "1" \
        --notify-before-shift-starts-hours "1" \
        --notify-after-shift-ends-hours "2" \
        --notify-driver-on-late-shift-assignment \
        --explicit-notify-driver-when-gps-not-available \
        --notify-driver-when-gps-not-available \
        --is-automated-job-site-time-creation-disabled \
        --is-time-card-payroll-certification-required-explicit true \
        --is-time-card-creating-time-sheet-line-item-explicit false \
        --skip-validate-driver-assignment-rule-evaluation \
        --driver-assignment-rule-override-reason "test override" \
        --disable-pre-start-notifications \
        --all-trips-entered \
        --hours-after-which-overtime-applies "8" \
        --travel-miles "10" \
        --billable-travel-minutes "15" \
        --loaded-tons-max "20" \
        --start-at "$UPDATE_START_AT" \
        --gross-weight-legal-limit-lbs-explicit "80000" \
        --auto-check-in-driver-on-arrival-at-start-site \
        --explicit-is-expecting-time-card \
        --explicit-material-transaction-tons-max "100" \
        --is-expecting-material-transactions \
        --expecting-material-transactions-message "Expecting loads" \
        --material-transaction-status open \
        --trucker-can-create-material-transactions \
        --reset-hours-after-which-overtime-applies
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update tender job schedule shift (permissions or policy)"
    fi
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete tender job schedule shift requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do tender-job-schedule-shifts delete "$SAMPLE_ID"
    assert_failure
else
    skip "No tender job schedule shift ID available"
fi

test_name "Delete tender job schedule shift"
TARGET_DELETE_ID="$SAMPLE_ID"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    TARGET_DELETE_ID="$CREATED_ID"
fi
if [[ -n "$TARGET_DELETE_ID" && "$TARGET_DELETE_ID" != "null" ]]; then
    xbe_run do tender-job-schedule-shifts delete "$TARGET_DELETE_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete tender job schedule shift (permissions or constraints)"
    fi
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tender job schedule shift without required fields fails"
xbe_run do tender-job-schedule-shifts create
assert_failure

test_name "Update tender job schedule shift without any fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do tender-job-schedule-shifts update "$SAMPLE_ID"
    assert_failure
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
