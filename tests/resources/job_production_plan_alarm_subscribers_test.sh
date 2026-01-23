#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Alarm Subscribers
#
# Tests CRUD operations for the job_production_plan_alarm_subscribers resource.
# Alarm subscribers link users to job production plan alarms.
#
# COVERAGE: All create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUBSCRIBER_ID=""
CREATED_USER_ID=""
SAMPLE_ID=""
SAMPLE_ALARM_ID=""
SAMPLE_SUBSCRIBER_ID=""

describe "Resource: job-production-plan-alarm-subscribers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan alarm subscribers"
xbe_json view job-production-plan-alarm-subscribers list --limit 5
assert_success

test_name "List job production plan alarm subscribers returns array"
xbe_json view job-production-plan-alarm-subscribers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_ALARM_ID=$(echo "$output" | jq -r '.[0].job_production_plan_alarm_id // empty')
    SAMPLE_SUBSCRIBER_ID=$(echo "$output" | jq -r '.[0].subscriber_id // empty')
else
    fail "Failed to list job production plan alarm subscribers"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan alarm subscriber"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-alarm-subscribers show "$SAMPLE_ID"
    assert_success
else
    skip "No alarm subscriber ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan alarm subscriber"
ALARM_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_ALARM_ID:-$SAMPLE_ALARM_ID}"
if [[ -n "$ALARM_ID" && "$ALARM_ID" != "null" ]]; then
    TEST_EMAIL=$(unique_email)
    xbe_json do users create --name "JPP Alarm Subscriber" --email "$TEST_EMAIL"
    if [[ $status -eq 0 ]]; then
        CREATED_USER_ID=$(json_get ".id")
        xbe_json do job-production-plan-alarm-subscribers create \
            --job-production-plan-alarm "$ALARM_ID" \
            --subscriber "$CREATED_USER_ID"
        if [[ $status -eq 0 ]]; then
            CREATED_SUBSCRIBER_ID=$(json_get ".id")
            if [[ -n "$CREATED_SUBSCRIBER_ID" && "$CREATED_SUBSCRIBER_ID" != "null" ]]; then
                register_cleanup "job-production-plan-alarm-subscribers" "$CREATED_SUBSCRIBER_ID"
                pass
            else
                fail "Created alarm subscriber but no ID returned"
            fi
        else
            if [[ "$output" == *"already"* ]] || [[ "$output" == *"exists"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                pass
            else
                fail "Failed to create alarm subscriber: $output"
            fi
        fi
    else
        fail "Failed to create user for alarm subscriber"
    fi
else
    skip "No alarm ID available (set XBE_TEST_JOB_PRODUCTION_PLAN_ALARM_ID)"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List alarm subscribers with --job-production-plan-alarm filter"
if [[ -n "$ALARM_ID" && "$ALARM_ID" != "null" ]]; then
    xbe_json view job-production-plan-alarm-subscribers list --job-production-plan-alarm "$ALARM_ID" --limit 5
    assert_success
else
    skip "No alarm ID available for filter test"
fi

test_name "List alarm subscribers with --subscriber filter"
SUBSCRIBER_FILTER_ID="${CREATED_USER_ID:-$SAMPLE_SUBSCRIBER_ID}"
if [[ -n "$SUBSCRIBER_FILTER_ID" && "$SUBSCRIBER_FILTER_ID" != "null" ]]; then
    xbe_json view job-production-plan-alarm-subscribers list --subscriber "$SUBSCRIBER_FILTER_ID" --limit 5
    assert_success
else
    skip "No subscriber ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete alarm subscriber requires --confirm flag"
if [[ -n "$CREATED_SUBSCRIBER_ID" && "$CREATED_SUBSCRIBER_ID" != "null" ]]; then
    xbe_run do job-production-plan-alarm-subscribers delete "$CREATED_SUBSCRIBER_ID"
    assert_failure
else
    skip "No created alarm subscriber for delete confirmation test"
fi

test_name "Delete alarm subscriber with --confirm"
if [[ -n "$CREATED_SUBSCRIBER_ID" && "$CREATED_SUBSCRIBER_ID" != "null" ]]; then
    xbe_run do job-production-plan-alarm-subscribers delete "$CREATED_SUBSCRIBER_ID" --confirm
    assert_success
else
    skip "No created alarm subscriber to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create alarm subscriber without alarm fails"
if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
    xbe_run do job-production-plan-alarm-subscribers create --subscriber "$CREATED_USER_ID"
    assert_failure
else
    skip "No user available for missing alarm test"
fi

test_name "Create alarm subscriber without subscriber fails"
if [[ -n "$ALARM_ID" && "$ALARM_ID" != "null" ]]; then
    xbe_run do job-production-plan-alarm-subscribers create --job-production-plan-alarm "$ALARM_ID"
    assert_failure
else
    skip "No alarm ID available for missing subscriber test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
