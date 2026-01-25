#!/bin/bash
#
# XBE CLI Integration Tests: User Device Location Tracking Requests
#
# Tests create operations for the user-device-location-tracking-requests resource.
#
# COVERAGE: Create attributes + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

USER_ID="${XBE_TEST_USER_DEVICE_LOCATION_TRACKING_REQUEST_USER_ID:-}"

describe "Resource: user-device-location-tracking-requests"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create location tracking request without required fields fails"
xbe_run do user-device-location-tracking-requests create
assert_failure

# ============================================================================
# Prerequisites
# ============================================================================

if [[ -z "$USER_ID" ]]; then
    test_name "Find user with push-enabled device (optional)"
    xbe_json view devices list --has-push-token true --limit 1
    if [[ $status -eq 0 ]]; then
        USER_ID=$(json_get ".[0].user_id")
        if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
            pass
        else
            skip "No device with push token found"
        fi
    else
        skip "Failed to list devices"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create location tracking request"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json do user-device-location-tracking-requests create \
        --user "$USER_ID" \
        --location-tracking-kind continuous \
        --location-tracking-action start

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".user_id" "$USER_ID"
        assert_json_equals ".location_tracking_kind" "continuous"
        assert_json_equals ".location_tracking_action" "start"
    else
        if [[ "$output" == *"no devices configured for push notifications"* ]] || [[ "$output" == *"no devices configured for push notifications have push tokens set"* ]]; then
            skip "User lacks push-enabled devices"
        elif [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to create request"
        else
            fail "Failed to create location tracking request"
        fi
    fi
else
    skip "No user with push-enabled device available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
