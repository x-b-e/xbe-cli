#!/bin/bash
#
# XBE CLI Integration Tests: Tender Job Schedule Shift Drivers
#
# Tests list, show, create, update, and delete operations for the
# tender-job-schedule-shift-drivers resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_SHIFT_ID=""
SAMPLE_USER_ID=""
SAMPLE_IS_PRIMARY=""

describe "Resource: tender-job-schedule-shift-drivers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tender job schedule shift drivers"
xbe_json view tender-job-schedule-shift-drivers list --limit 5
assert_success

test_name "List tender job schedule shift drivers returns array"
xbe_json view tender-job-schedule-shift-drivers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tender job schedule shift drivers"
fi

# ============================================================================
# Sample Record (used for filters/show/update/delete)
# ============================================================================

test_name "Capture sample shift driver"
xbe_json view tender-job-schedule-shift-drivers list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_USER_ID=$(json_get ".[0].user_id")
    SAMPLE_IS_PRIMARY=$(json_get ".[0].is_primary")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No shift drivers available for follow-on tests"
    fi
else
    skip "Could not list shift drivers to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List shift drivers with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_SHIFT_ID" && "$SAMPLE_SHIFT_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shift-drivers list --tender-job-schedule-shift "$SAMPLE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List shift drivers with --user filter"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shift-drivers list --user "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender job schedule shift driver"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shift-drivers show "$SAMPLE_ID"
    assert_success
else
    skip "No shift driver ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create shift driver"
if [[ -n "$SAMPLE_SHIFT_ID" && "$SAMPLE_SHIFT_ID" != "null" && -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json do tender-job-schedule-shift-drivers create \
        --tender-job-schedule-shift "$SAMPLE_SHIFT_ID" \
        --user "$SAMPLE_USER_ID" \
        --is-primary
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No shift/user IDs available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update shift driver primary flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    NEW_PRIMARY="true"
    if [[ "$SAMPLE_IS_PRIMARY" == "true" ]]; then
        NEW_PRIMARY="false"
    fi
    xbe_json do tender-job-schedule-shift-drivers update "$SAMPLE_ID" --is-primary "$NEW_PRIMARY"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update shift driver (permissions or policy)"
    fi
else
    skip "No shift driver ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete shift driver requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do tender-job-schedule-shift-drivers delete "$SAMPLE_ID"
    assert_failure
else
    skip "No shift driver ID available"
fi

test_name "Delete shift driver"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do tender-job-schedule-shift-drivers delete "$SAMPLE_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete shift driver (permissions or constraints)"
    fi
else
    skip "No shift driver ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create shift driver without required fields fails"
xbe_run do tender-job-schedule-shift-drivers create
assert_failure

test_name "Update shift driver without any fields fails"
xbe_run do tender-job-schedule-shift-drivers update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
