#!/bin/bash
#
# XBE CLI Integration Tests: Device Location Events
#
# Tests create behavior for device location events.
#
# COVERAGE: Create (payload + explicit fields) + error cases
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: device-location-events"

# ==========================================================================
# CREATE Tests
# ==========================================================================

DEVICE_IDENTIFIER="cli-test-device-$(date +%s)-$RANDOM"
EVENT_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
EVENT_UUID="evt-$(date +%s)-$RANDOM"
PAYLOAD=$(printf '{"uuid":"%s","timestamp":"%s","activity":{"type":"walking"},"coords":{"latitude":40.0,"longitude":-74.0}}' "$EVENT_UUID" "$EVENT_TIME")

test_name "Create device location event with payload"
xbe_json do device-location-events create \
    --device-identifier "$DEVICE_IDENTIFIER" \
    --payload "$PAYLOAD"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to create device location event with payload"
fi

EVENT_ID="evt-fields-$(date +%s)-$RANDOM"
EVENT_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)

test_name "Create device location event with explicit fields"
xbe_json do device-location-events create \
    --device-identifier "$DEVICE_IDENTIFIER" \
    --event-id "$EVENT_ID" \
    --event-at "$EVENT_AT" \
    --event-description "moving" \
    --event-latitude 40.7128 \
    --event-longitude -74.0060
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to create device location event with explicit fields"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create device location event without device-identifier fails"
xbe_run do device-location-events create --payload "$PAYLOAD"
assert_failure

test_name "Create device location event with invalid payload JSON fails"
xbe_run do device-location-events create --device-identifier "$DEVICE_IDENTIFIER" --payload "{invalid"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
