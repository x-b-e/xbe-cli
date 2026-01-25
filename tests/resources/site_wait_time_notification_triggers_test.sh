#!/bin/bash
#
# XBE CLI Integration Tests: Site Wait Time Notification Triggers
#
# Tests list and show operations for the site_wait_time_notification_triggers resource.
# Site wait time notification triggers capture excessive wait time events at sites.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TRIGGER_ID=""
SHIFT_ID=""
JOB_PLAN_ID=""
SITE_TYPE=""
SITE_ID=""
EVENT_AT=""
SKIP_ID_FILTERS=0


describe "Resource: site-wait-time-notification-triggers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List site wait time notification triggers"
xbe_json view site-wait-time-notification-triggers list --limit 5
assert_success

test_name "List site wait time notification triggers returns array"
xbe_json view site-wait-time-notification-triggers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list site wait time notification triggers"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample site wait time notification trigger"
xbe_json view site-wait-time-notification-triggers list --limit 1
if [[ $status -eq 0 ]]; then
    TRIGGER_ID=$(json_get ".[0].id")
    SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    JOB_PLAN_ID=$(json_get ".[0].job_production_plan_id")
    SITE_TYPE=$(json_get ".[0].site_type")
    SITE_ID=$(json_get ".[0].site_id")
    EVENT_AT=$(json_get ".[0].event_at")
    if [[ -n "$TRIGGER_ID" && "$TRIGGER_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No site wait time notification triggers available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list site wait time notification triggers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List site wait time notification triggers with --tender-job-schedule-shift filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers list --tender-job-schedule-shift "$SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List site wait time notification triggers with --job-production-plan filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$JOB_PLAN_ID" && "$JOB_PLAN_ID" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers list --job-production-plan "$JOB_PLAN_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List site wait time notification triggers with --site-type filter"
if [[ -n "$SITE_TYPE" && "$SITE_TYPE" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers list --site-type "$SITE_TYPE" --limit 5
    assert_success
else
    skip "No site type available"
fi

test_name "List site wait time notification triggers with --site-id filter"
if [[ -n "$SITE_ID" && "$SITE_ID" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers list --site-id "$SITE_ID" --limit 5
    assert_success
else
    skip "No site ID available"
fi

test_name "List site wait time notification triggers with --event-at-min filter"
if [[ -n "$EVENT_AT" && "$EVENT_AT" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers list --event-at-min "$EVENT_AT" --limit 5
    assert_success
else
    skip "No event-at value available"
fi

test_name "List site wait time notification triggers with --event-at-max filter"
if [[ -n "$EVENT_AT" && "$EVENT_AT" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers list --event-at-max "$EVENT_AT" --limit 5
    assert_success
else
    skip "No event-at value available"
fi

test_name "List site wait time notification triggers with --is-event-at filter"
xbe_json view site-wait-time-notification-triggers list --is-event-at true --limit 5
assert_success

test_name "List site wait time notification triggers with --created-at-min filter"
xbe_json view site-wait-time-notification-triggers list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List site wait time notification triggers with --created-at-max filter"
xbe_json view site-wait-time-notification-triggers list --created-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List site wait time notification triggers with --updated-at-min filter"
xbe_json view site-wait-time-notification-triggers list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List site wait time notification triggers with --updated-at-max filter"
xbe_json view site-wait-time-notification-triggers list --updated-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List site wait time notification triggers with --is-created-at filter"
xbe_json view site-wait-time-notification-triggers list --is-created-at true --limit 5
assert_success

test_name "List site wait time notification triggers with --is-updated-at filter"
xbe_json view site-wait-time-notification-triggers list --is-updated-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show site wait time notification trigger"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TRIGGER_ID" && "$TRIGGER_ID" != "null" ]]; then
    xbe_json view site-wait-time-notification-triggers show "$TRIGGER_ID"
    assert_success
else
    skip "No site wait time notification trigger ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
