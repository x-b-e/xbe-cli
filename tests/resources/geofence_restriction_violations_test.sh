#!/bin/bash
#
# XBE CLI Integration Tests: Geofence Restriction Violations
#
# Tests list and show operations for the geofence-restriction-violations resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_GEOFENCE_ID=""
SAMPLE_TRAILER_ID=""
SAMPLE_TRACTOR_ID=""
SAMPLE_DRIVER_ID=""
SAMPLE_SHIFT_ID=""
SAMPLE_EVENT_AT=""
SAMPLE_NOTIFICATION_AT=""
SAMPLE_SHOULD_NOTIFY=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""

describe "Resource: geofence-restriction-violations"

# =========================================================================
# LIST Tests - Basic
# =========================================================================

test_name "List geofence restriction violations"
xbe_json view geofence-restriction-violations list --limit 5
assert_success

test_name "List geofence restriction violations returns array"
xbe_json view geofence-restriction-violations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list geofence restriction violations"
fi

# =========================================================================
# Sample Record (used for filters/show)
# =========================================================================

test_name "Capture sample geofence restriction violation"
xbe_json view geofence-restriction-violations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_GEOFENCE_ID=$(json_get ".[0].geofence_id")
    SAMPLE_TRAILER_ID=$(json_get ".[0].trailer_id")
    SAMPLE_TRACTOR_ID=$(json_get ".[0].tractor_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    SAMPLE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_EVENT_AT=$(json_get ".[0].event_at")
    SAMPLE_NOTIFICATION_AT=$(json_get ".[0].notification_sent_at")
    SAMPLE_SHOULD_NOTIFY=$(json_get ".[0].should_notify")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No geofence restriction violations available for follow-on tests"
    fi
else
    skip "Could not list geofence restriction violations to capture sample"
fi

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show geofence restriction violation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view geofence-restriction-violations show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_CREATED_AT=$(json_get ".created_at")
        SAMPLE_UPDATED_AT=$(json_get ".updated_at")
        pass
    else
        fail "Failed to show geofence restriction violation"
    fi
else
    skip "No violation ID available"
fi

# =========================================================================
# LIST Tests - Filters
# =========================================================================

test_name "List violations with --geofence filter"
if [[ -n "$SAMPLE_GEOFENCE_ID" && "$SAMPLE_GEOFENCE_ID" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --geofence "$SAMPLE_GEOFENCE_ID" --limit 5
    assert_success
else
    skip "No geofence ID available"
fi

test_name "List violations with --trailer filter"
if [[ -n "$SAMPLE_TRAILER_ID" && "$SAMPLE_TRAILER_ID" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --trailer "$SAMPLE_TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "List violations with --tractor filter"
if [[ -n "$SAMPLE_TRACTOR_ID" && "$SAMPLE_TRACTOR_ID" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --tractor "$SAMPLE_TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "List violations with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List violations with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_SHIFT_ID" && "$SAMPLE_SHIFT_ID" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --tender-job-schedule-shift "$SAMPLE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List violations with --event-at-min filter"
if [[ -n "$SAMPLE_EVENT_AT" && "$SAMPLE_EVENT_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --event-at-min "$SAMPLE_EVENT_AT" --limit 5
    assert_success
else
    skip "No event_at available"
fi

test_name "List violations with --event-at-max filter"
if [[ -n "$SAMPLE_EVENT_AT" && "$SAMPLE_EVENT_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --event-at-max "$SAMPLE_EVENT_AT" --limit 5
    assert_success
else
    skip "No event_at available"
fi

test_name "List violations with --notification-sent-at-min filter"
if [[ -n "$SAMPLE_NOTIFICATION_AT" && "$SAMPLE_NOTIFICATION_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --notification-sent-at-min "$SAMPLE_NOTIFICATION_AT" --limit 5
    assert_success
else
    skip "No notification_sent_at available"
fi

test_name "List violations with --notification-sent-at-max filter"
if [[ -n "$SAMPLE_NOTIFICATION_AT" && "$SAMPLE_NOTIFICATION_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --notification-sent-at-max "$SAMPLE_NOTIFICATION_AT" --limit 5
    assert_success
else
    skip "No notification_sent_at available"
fi

test_name "List violations with --should-notify filter"
if [[ -n "$SAMPLE_SHOULD_NOTIFY" && "$SAMPLE_SHOULD_NOTIFY" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --should-notify "$SAMPLE_SHOULD_NOTIFY" --limit 5
    assert_success
else
    skip "No should_notify value available"
fi

test_name "List violations with --created-at-min filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created_at available"
fi

test_name "List violations with --created-at-max filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created_at available"
fi

test_name "List violations with --updated-at-min filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --updated-at-min "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated_at available"
fi

test_name "List violations with --updated-at-max filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view geofence-restriction-violations list --updated-at-max "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated_at available"
fi

# =========================================================================
# LIST Tests - Pagination
# =========================================================================

test_name "List violations with --limit"
xbe_json view geofence-restriction-violations list --limit 3
assert_success

test_name "List violations with --offset"
xbe_json view geofence-restriction-violations list --limit 3 --offset 3
assert_success

# =========================================================================
# Summary
# =========================================================================

run_tests
