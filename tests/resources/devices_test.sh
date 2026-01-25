#!/bin/bash
#
# XBE CLI Integration Tests: Devices
#
# Tests list and update operations for the devices resource.
# Devices represent mobile app instances.
#
# NOTE: Devices cannot be created or deleted via the API.
# They are created automatically when users log in to the mobile app.
# This test focuses on list operations with filters.
#
# COVERAGE: List filters + update attributes (if device exists)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: devices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List devices"
xbe_json view devices list --limit 5
assert_success

test_name "List devices returns array"
xbe_json view devices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list devices"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List devices with --user filter"
# Use a likely non-existent user ID to test filter works without errors
xbe_json view devices list --user 1 --limit 10
assert_success

test_name "List devices with --identifier filter"
# Use a dummy identifier to test filter works without errors
xbe_json view devices list --identifier "test-device-id" --limit 10
assert_success

test_name "List devices with --is-pushable filter"
xbe_json view devices list --is-pushable true --limit 10
assert_success

test_name "List devices with --has-push-token filter"
xbe_json view devices list --has-push-token true --limit 10
assert_success

test_name "List devices with --first-device-location-event-at-min filter"
xbe_json view devices list --first-device-location-event-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List devices with --first-device-location-event-at-max filter"
xbe_json view devices list --first-device-location-event-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List devices with --last-device-location-event-at-min filter"
xbe_json view devices list --last-device-location-event-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List devices with --last-device-location-event-at-max filter"
xbe_json view devices list --last-device-location-event-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List devices with --limit"
xbe_json view devices list --limit 3
assert_success

test_name "List devices with --offset"
xbe_json view devices list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

# NOTE: Update tests require an existing device that the user can access.
# We'll try to find one and update it if available.

test_name "Attempt to update device (may skip if no devices available)"
xbe_json view devices list --limit 1
if [[ $status -eq 0 ]]; then
    DEVICE_ID=$(json_get ".[0].id")
    if [[ -n "$DEVICE_ID" && "$DEVICE_ID" != "null" ]]; then
        # Try to update the device nickname
        xbe_json do devices update "$DEVICE_ID" --nickname "Test CLI Device"
        if [[ $status -eq 0 ]]; then
            pass
        else
            # May fail due to permissions - that's acceptable
            skip "Could not update device - may not have permission"
        fi
    else
        skip "No devices available to test update"
    fi
else
    skip "Could not list devices to find one for update test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update device without any fields fails"
xbe_json view devices list --limit 1
if [[ $status -eq 0 ]]; then
    DEVICE_ID=$(json_get ".[0].id")
    if [[ -n "$DEVICE_ID" && "$DEVICE_ID" != "null" ]]; then
        xbe_run do devices update "$DEVICE_ID"
        assert_failure
    else
        skip "No devices available to test error case"
    fi
else
    skip "Could not list devices for error case test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
