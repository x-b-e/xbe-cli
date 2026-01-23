#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Location Estimates
#
# Tests view operations for the equipment-location-estimates resource.
# Equipment location estimates return the most recent known location for equipment.
#
# NOTE: This test creates prerequisite broker, equipment classification, and equipment.
#
# COVERAGE: List filters + required filter failure
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""

describe "Resource: equipment-location-estimates (view-only)"

# ============================================================================
# Prerequisites - Create broker, equipment classification, equipment
# ============================================================================

test_name "Create prerequisite broker for equipment location estimate tests"
BROKER_NAME=$(unique_name "EquipLocBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite equipment classification"
CLASSIFICATION_NAME=$(unique_name "EquipLocClass")

xbe_json do equipment-classifications create --name "$CLASSIFICATION_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_CLASSIFICATION_ID" && "$CREATED_EQUIPMENT_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIPMENT_CLASSIFICATION_ID"
        pass
    else
        fail "Created equipment classification but no ID returned"
        echo "Cannot continue without an equipment classification"
        run_tests
    fi
else
    fail "Failed to create equipment classification"
    echo "Cannot continue without an equipment classification"
    run_tests
fi

test_name "Create equipment for location estimate tests"
EQUIPMENT_NAME=$(unique_name "EquipLoc")

xbe_json do equipment create \
    --nickname "$EQUIPMENT_NAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_ID" && "$CREATED_EQUIPMENT_ID" != "null" ]]; then
        register_cleanup "equipment" "$CREATED_EQUIPMENT_ID"
        pass
    else
        fail "Created equipment but no ID returned"
    fi
else
    fail "Failed to create equipment"
fi

if [[ -z "$CREATED_EQUIPMENT_ID" || "$CREATED_EQUIPMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment ID"
    run_tests
fi

# ============================================================================
# LIST Tests - Required Filter
# ============================================================================

test_name "List equipment location estimates requires --equipment"
xbe_run view equipment-location-estimates list
assert_failure

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment location estimates returns array"
xbe_json view equipment-location-estimates list --equipment "$CREATED_EQUIPMENT_ID"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment location estimates"
fi

test_name "List equipment location estimates returns matching equipment ID"
xbe_json view equipment-location-estimates list --equipment "$CREATED_EQUIPMENT_ID"
if [[ $status -eq 0 ]]; then
    equipment_id=$(json_get ".[0].equipment_id")
    if [[ "$equipment_id" == "$CREATED_EQUIPMENT_ID" ]]; then
        pass
    else
        fail "Expected equipment_id to be $CREATED_EQUIPMENT_ID, got $equipment_id"
    fi
else
    fail "Failed to list equipment location estimates"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List equipment location estimates with --as-of"
xbe_json view equipment-location-estimates list \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --as-of "2026-01-23T12:00:00Z"
assert_success

test_name "List equipment location estimates with --earliest-event-at"
xbe_json view equipment-location-estimates list \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --earliest-event-at "2026-01-22T00:00:00Z"
assert_success

test_name "List equipment location estimates with --latest-event-at"
xbe_json view equipment-location-estimates list \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --latest-event-at "2026-01-23T23:59:59Z"
assert_success

test_name "List equipment location estimates with --max-abs-latency-seconds"
xbe_json view equipment-location-estimates list \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --max-abs-latency-seconds 3600
assert_success

test_name "List equipment location estimates with --max-latest-seconds (server error expected)"
xbe_json view equipment-location-estimates list \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --max-latest-seconds 86400
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
