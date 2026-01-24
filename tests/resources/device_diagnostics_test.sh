#!/bin/bash
#
# XBE CLI Integration Tests: Device Diagnostics
#
# Tests create operations and list filters for device diagnostics.
#
# COVERAGE: Create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_DIAGNOSTIC_ID=""
SECOND_DIAGNOSTIC_ID=""
CURRENT_USER_ID=""
DEVICE_ID=""

CHANGED_AT="2025-01-01T12:00:00Z"
CHANGESET='{"battery_level":85,"network":"wifi"}'

DEVICE_IDENTIFIER=$(unique_name "DeviceDiag")


describe "Resource: device-diagnostics"

# ==========================================================================
# Prerequisites
# ==========================================================================

test_name "Fetch current user (optional)"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        skip "No user ID returned"
    fi
else
    skip "Failed to fetch current user"
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create device diagnostic with full attributes"

create_args=(do device-diagnostics create
    --device-identifier "$DEVICE_IDENTIFIER"
    --is-trackable=true
    --is-tracking=true
    --stop-on-terminate=false
    --is-in-power-saver-mode=true
    --is-ignoring-battery-optimizations=true
    --permission-status authorized
    --motion-permission-status granted
    --location-accuracy-authorization-status full
    --are-location-services-enabled=true
    --is-gps-location-provider-enabled=true
    --is-network-location-provider-enabled=false
    --is-not-tracking-because-of-stationary-mode=false
    --changeset "$CHANGESET"
    --changed-at "$CHANGED_AT"
    --change-trigger-source "app"
    --change-trigger-context "diagnostics")

if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    create_args+=(--user "$CURRENT_USER_ID")
fi

xbe_json "${create_args[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_DIAGNOSTIC_ID=$(json_get ".id")
    DEVICE_ID=$(json_get ".device_id")
    if [[ -n "$CREATED_DIAGNOSTIC_ID" && "$CREATED_DIAGNOSTIC_ID" != "null" ]]; then
        pass
    else
        fail "Created device diagnostic but no ID returned"
    fi
else
    fail "Failed to create device diagnostic"
fi

# Only continue if we successfully created a diagnostic
if [[ -z "$CREATED_DIAGNOSTIC_ID" || "$CREATED_DIAGNOSTIC_ID" == "null" ]]; then
    echo "Cannot continue without a valid device diagnostic ID"
    run_tests
fi

if [[ -z "$DEVICE_ID" || "$DEVICE_ID" == "null" ]]; then
    xbe_json view devices list --identifier "$DEVICE_IDENTIFIER" --limit 1
    if [[ $status -eq 0 ]]; then
        DEVICE_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$DEVICE_ID" && "$DEVICE_ID" != "null" ]]; then
    test_name "Create device diagnostic using --device"
    xbe_json do device-diagnostics create --device "$DEVICE_ID" --is-tracking=false
    if [[ $status -eq 0 ]]; then
        SECOND_DIAGNOSTIC_ID=$(json_get ".id")
        if [[ -n "$SECOND_DIAGNOSTIC_ID" && "$SECOND_DIAGNOSTIC_ID" != "null" ]]; then
            pass
        else
            fail "Created device diagnostic with --device but no ID returned"
        fi
    else
        fail "Failed to create device diagnostic with --device"
    fi
else
    test_name "Create device diagnostic using --device"
    skip "No device ID returned from first create"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show device diagnostic"
xbe_json view device-diagnostics show "$CREATED_DIAGNOSTIC_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List device diagnostics"
xbe_json view device-diagnostics list --limit 5
assert_success

test_name "List device diagnostics returns array"
xbe_json view device-diagnostics list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list device diagnostics"
fi

# ==========================================================================
# LIST Tests - All Filters
# ==========================================================================

test_name "List device diagnostics with --user filter"
USER_FILTER_ID="$CURRENT_USER_ID"
if [[ -z "$USER_FILTER_ID" || "$USER_FILTER_ID" == "null" ]]; then
    USER_FILTER_ID="1"
fi
xbe_json view device-diagnostics list --user "$USER_FILTER_ID" --limit 5
assert_success


test_name "List device diagnostics with --device filter"
DEVICE_FILTER_ID="$DEVICE_ID"
if [[ -z "$DEVICE_FILTER_ID" || "$DEVICE_FILTER_ID" == "null" ]]; then
    DEVICE_FILTER_ID="1"
fi
xbe_json view device-diagnostics list --device "$DEVICE_FILTER_ID" --limit 5
assert_success


test_name "List device diagnostics with --device-identifier filter"
DEVICE_IDENTIFIER_FILTER="$DEVICE_IDENTIFIER"
if [[ -z "$DEVICE_IDENTIFIER_FILTER" ]]; then
    DEVICE_IDENTIFIER_FILTER="test-device"
fi
xbe_json view device-diagnostics list --device-identifier "$DEVICE_IDENTIFIER_FILTER" --limit 5
assert_success


test_name "List device diagnostics with --permission-status filter"
xbe_json view device-diagnostics list --permission-status "authorized" --limit 5
assert_success


test_name "List device diagnostics with --not-permission-status filter"
xbe_json view device-diagnostics list --not-permission-status "denied" --limit 5
assert_success


test_name "List device diagnostics with --has-permission-status filter"
xbe_json view device-diagnostics list --has-permission-status true --limit 5
assert_success


test_name "List device diagnostics with --motion-permission-status filter"
xbe_json view device-diagnostics list --motion-permission-status "granted" --limit 5
assert_success


test_name "List device diagnostics with --not-motion-permission-status filter"
xbe_json view device-diagnostics list --not-motion-permission-status "denied" --limit 5
assert_success


test_name "List device diagnostics with --has-motion-permission-status filter"
xbe_json view device-diagnostics list --has-motion-permission-status true --limit 5
assert_success


test_name "List device diagnostics with --location-accuracy-authorization-status filter"
xbe_json view device-diagnostics list --location-accuracy-authorization-status "full" --limit 5
assert_success


test_name "List device diagnostics with --has-location-accuracy-authorization-status filter"
xbe_json view device-diagnostics list --has-location-accuracy-authorization-status true --limit 5
assert_success


test_name "List device diagnostics with --are-location-services-enabled filter"
xbe_json view device-diagnostics list --are-location-services-enabled true --limit 5
assert_success


test_name "List device diagnostics with --is-gps-location-provider-enabled filter"
xbe_json view device-diagnostics list --is-gps-location-provider-enabled true --limit 5
assert_success


test_name "List device diagnostics with --is-network-location-provider-enabled filter"
xbe_json view device-diagnostics list --is-network-location-provider-enabled false --limit 5
assert_success


test_name "List device diagnostics with --is-not-tracking-because-of-stationary-mode filter"
xbe_json view device-diagnostics list --is-not-tracking-because-of-stationary-mode false --limit 5
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
