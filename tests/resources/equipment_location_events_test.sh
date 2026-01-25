#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Location Events
#
# Tests CRUD operations for equipment location events.
#
# NOTE: This test requires creating prerequisite resources: broker and equipment classification
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EQUIPMENT_LOCATION_EVENT_ID=""
CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_EQUIPMENT_ID_2=""
UPDATED_BY_ID=""

describe "Resource: equipment-location-events"

# ============================================================================
# Prerequisites - Create broker and equipment classification
# ============================================================================

test_name "Create prerequisite broker for equipment location events"
BROKER_NAME=$(unique_name "EquipLocEventBroker")

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
EC_NAME=$(unique_name "EquipLocEventClass")
EC_ABBR="EL$(date +%s | tail -c 4)"

xbe_json do equipment-classifications create \
    --name "$EC_NAME" \
    --abbreviation "$EC_ABBR"

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

# ============================================================================
# Create equipment
# ============================================================================

test_name "Create equipment for location events"
EQUIPMENT_NICKNAME=$(unique_name "LocEventEquip")

xbe_json do equipment create \
    --nickname "$EQUIPMENT_NICKNAME" \
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
        echo "Cannot continue without equipment"
        run_tests
    fi
else
    fail "Failed to create equipment"
    echo "Cannot continue without equipment"
    run_tests
fi

test_name "Create second equipment for update tests"
EQUIPMENT_NICKNAME_2=$(unique_name "LocEventEquip2")

xbe_json do equipment create \
    --nickname "$EQUIPMENT_NICKNAME_2" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_ID_2" && "$CREATED_EQUIPMENT_ID_2" != "null" ]]; then
        register_cleanup "equipment" "$CREATED_EQUIPMENT_ID_2"
        pass
    else
        fail "Created equipment but no ID returned"
        echo "Cannot continue without equipment"
        run_tests
    fi
else
    fail "Failed to create equipment"
    echo "Cannot continue without equipment"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment location event with required fields"
EVENT_AT="2025-01-15T12:00:00Z"
EVENT_LAT="40.7128"
EVENT_LON="-74.0060"

xbe_json do equipment-location-events create \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --event-at "$EVENT_AT" \
    --event-latitude "$EVENT_LAT" \
    --event-longitude "$EVENT_LON" \
    --provenance "gps"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_LOCATION_EVENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_LOCATION_EVENT_ID" && "$CREATED_EQUIPMENT_LOCATION_EVENT_ID" != "null" ]]; then
        register_cleanup "equipment-location-events" "$CREATED_EQUIPMENT_LOCATION_EVENT_ID"
        pass
    else
        fail "Created equipment location event but no ID returned"
    fi
else
    fail "Failed to create equipment location event"
fi

if [[ -z "$CREATED_EQUIPMENT_LOCATION_EVENT_ID" || "$CREATED_EQUIPMENT_LOCATION_EVENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment location event ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update equipment location event"
UPDATED_EVENT_AT="2025-01-16T12:00:00Z"
UPDATED_LAT="41.0000"
UPDATED_LON="-73.9000"

xbe_json do equipment-location-events update "$CREATED_EQUIPMENT_LOCATION_EVENT_ID" \
    --equipment "$CREATED_EQUIPMENT_ID_2" \
    --event-at "$UPDATED_EVENT_AT" \
    --event-latitude "$UPDATED_LAT" \
    --event-longitude "$UPDATED_LON" \
    --provenance "map"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment location event"
xbe_json view equipment-location-events show "$CREATED_EQUIPMENT_LOCATION_EVENT_ID"
if [[ $status -eq 0 ]]; then
    UPDATED_BY_ID=$(json_get ".updated_by_id")
    pass
else
    fail "Show failed"
fi

# ============================================================================
# LIST Tests (Filters)
# ============================================================================

test_name "List equipment location events with --equipment filter"
xbe_json view equipment-location-events list --equipment "$CREATED_EQUIPMENT_ID_2" --limit 5
assert_success

test_name "List equipment location events with --provenance filter"
xbe_json view equipment-location-events list --provenance "map" --limit 5
assert_success

test_name "List equipment location events with --event-at-min filter"
xbe_json view equipment-location-events list --event-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List equipment location events with --event-at-max filter"
xbe_json view equipment-location-events list --event-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List equipment location events with --is-event-at filter"
xbe_json view equipment-location-events list --is-event-at true --limit 5
assert_success

test_name "List equipment location events with --created-at-min filter"
xbe_json view equipment-location-events list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List equipment location events with --created-at-max filter"
xbe_json view equipment-location-events list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List equipment location events with --is-created-at filter"
xbe_json view equipment-location-events list --is-created-at true --limit 5
assert_success

test_name "List equipment location events with --updated-at-min filter"
xbe_json view equipment-location-events list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List equipment location events with --updated-at-max filter"
xbe_json view equipment-location-events list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List equipment location events with --is-updated-at filter"
xbe_json view equipment-location-events list --is-updated-at true --limit 5
assert_success

if [[ -n "$UPDATED_BY_ID" && "$UPDATED_BY_ID" != "null" ]]; then
    test_name "List equipment location events with --updated-by filter"
    xbe_json view equipment-location-events list --updated-by "$UPDATED_BY_ID" --limit 5
    assert_success
else
    test_name "List equipment location events with --updated-by filter"
    skip "No updated-by ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create equipment location event without --equipment fails"
xbe_json do equipment-location-events create --provenance "gps"
assert_failure

test_name "Create equipment location event without --provenance fails"
xbe_json do equipment-location-events create --equipment "$CREATED_EQUIPMENT_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do equipment-location-events update "$CREATED_EQUIPMENT_LOCATION_EVENT_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment location event"
xbe_run do equipment-location-events delete "$CREATED_EQUIPMENT_LOCATION_EVENT_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
