#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Stop Completions
#
# Tests CRUD operations for the equipment_movement_stop_completions resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_LOCATION_ID=""
CREATED_TRIP_ID=""
CREATED_STOP_ID=""
CREATED_REQUIREMENT_ID=""
CREATED_STOP_REQUIREMENT_ID=""
CREATED_COMPLETION_ID=""

COMPLETED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
UPDATED_COMPLETED_AT="$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")"
LATITUDE="34.0500"
LONGITUDE="-118.2500"
UPDATED_LATITUDE="34.0600"
UPDATED_LONGITUDE="-118.2600"

DIRECT_API_AVAILABLE=0
if [[ -n "$XBE_TOKEN" ]]; then
    DIRECT_API_AVAILABLE=1
fi

cleanup_api_resources() {
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        return
    fi

    if [[ -n "$CREATED_STOP_REQUIREMENT_ID" && "$CREATED_STOP_REQUIREMENT_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/equipment-movement-stop-requirements/$CREATED_STOP_REQUIREMENT_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_REQUIREMENT_ID" && "$CREATED_REQUIREMENT_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/equipment-movement-requirements/$CREATED_REQUIREMENT_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/equipment-movement-stops/$CREATED_STOP_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_TRIP_ID" && "$CREATED_TRIP_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/equipment-movement-trips/$CREATED_TRIP_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/equipment-movement-requirement-locations/$CREATED_LOCATION_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi
}

trap 'cleanup_api_resources; run_cleanup' EXIT

api_post() {
    local path="$1"
    local body="$2"
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        output="Missing XBE_TOKEN for direct API calls"
        status=1
        return
    fi
    run curl -sS -X POST "$XBE_BASE_URL$path" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json" \
        -d "$body"
}

describe "Resource: equipment_movement_stop_completions"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for stop completion tests"
BROKER_NAME=$(unique_name "EMStopBroker")

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
        fail "Failed to create broker"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create equipment classification for stop completion tests"
EQUIPMENT_CLASS_NAME=$(unique_name "EMStopEquipmentClass")

xbe_json do equipment-classifications create --name "$EQUIPMENT_CLASS_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_CLASSIFICATION_ID" && "$CREATED_EQUIPMENT_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIPMENT_CLASSIFICATION_ID"
        pass
    else
        fail "Created equipment classification but no ID returned"
        echo "Cannot continue without equipment classification"
        run_tests
    fi
else
    fail "Failed to create equipment classification"
    echo "Cannot continue without equipment classification"
    run_tests
fi

test_name "Create equipment for stop completion tests"
EQUIPMENT_NAME=$(unique_name "EMStopEquipment")

xbe_json do equipment create \
    --nickname "$EQUIPMENT_NAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type brokers \
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

if [[ -n "$XBE_TEST_EQUIPMENT_MOVEMENT_STOP_ID" && -n "$XBE_TEST_EQUIPMENT_ID" ]]; then
    CREATED_STOP_ID="$XBE_TEST_EQUIPMENT_MOVEMENT_STOP_ID"
    CREATED_EQUIPMENT_ID="$XBE_TEST_EQUIPMENT_ID"
else
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        skip "Set XBE_TOKEN or XBE_TEST_EQUIPMENT_MOVEMENT_STOP_ID/XBE_TEST_EQUIPMENT_ID to run stop completion tests"
        run_tests
    fi

    test_name "Create equipment movement requirement location"
    LOCATION_NAME=$(unique_name "EMStopLocation")
    api_post "/v1/equipment-movement-requirement-locations" "{\"data\":{\"type\":\"equipment-movement-requirement-locations\",\"attributes\":{\"name\":\"$LOCATION_NAME\",\"latitude\":$LATITUDE,\"longitude\":$LONGITUDE},\"relationships\":{\"broker\":{\"data\":{\"type\":\"brokers\",\"id\":\"$CREATED_BROKER_ID\"}}}}}"
    if [[ $status -eq 0 ]]; then
        CREATED_LOCATION_ID=$(json_get ".data.id")
        if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
            pass
        else
            fail "Created location but no ID returned"
            echo "Cannot continue without location"
            run_tests
        fi
    else
        fail "Failed to create equipment movement requirement location"
        run_tests
    fi

    test_name "Create equipment movement trip"
    api_post "/v1/equipment-movement-trips" "{\"data\":{\"type\":\"equipment-movement-trips\",\"relationships\":{\"broker\":{\"data\":{\"type\":\"brokers\",\"id\":\"$CREATED_BROKER_ID\"}}}}}"
    if [[ $status -eq 0 ]]; then
        CREATED_TRIP_ID=$(json_get ".data.id")
        if [[ -n "$CREATED_TRIP_ID" && "$CREATED_TRIP_ID" != "null" ]]; then
            pass
        else
            fail "Created trip but no ID returned"
            echo "Cannot continue without trip"
            run_tests
        fi
    else
        fail "Failed to create equipment movement trip"
        run_tests
    fi

    test_name "Create equipment movement stop"
    api_post "/v1/equipment-movement-stops" "{\"data\":{\"type\":\"equipment-movement-stops\",\"attributes\":{\"sequence-position\":1},\"relationships\":{\"trip\":{\"data\":{\"type\":\"equipment-movement-trips\",\"id\":\"$CREATED_TRIP_ID\"}},\"location\":{\"data\":{\"type\":\"equipment-movement-requirement-locations\",\"id\":\"$CREATED_LOCATION_ID\"}}}}}"
    if [[ $status -eq 0 ]]; then
        CREATED_STOP_ID=$(json_get ".data.id")
        if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
            pass
        else
            fail "Created stop but no ID returned"
            echo "Cannot continue without stop"
            run_tests
        fi
    else
        fail "Failed to create equipment movement stop"
        run_tests
    fi

    test_name "Create equipment movement requirement"
    api_post "/v1/equipment-movement-requirements" "{\"data\":{\"type\":\"equipment-movement-requirements\",\"attributes\":{\"note\":\"Test requirement\"},\"relationships\":{\"broker\":{\"data\":{\"type\":\"brokers\",\"id\":\"$CREATED_BROKER_ID\"}},\"equipment\":{\"data\":{\"type\":\"equipment\",\"id\":\"$CREATED_EQUIPMENT_ID\"}},\"origin\":{\"data\":{\"type\":\"equipment-movement-requirement-locations\",\"id\":\"$CREATED_LOCATION_ID\"}}}}}"
    if [[ $status -eq 0 ]]; then
        CREATED_REQUIREMENT_ID=$(json_get ".data.id")
        if [[ -n "$CREATED_REQUIREMENT_ID" && "$CREATED_REQUIREMENT_ID" != "null" ]]; then
            pass
        else
            fail "Created requirement but no ID returned"
            echo "Cannot continue without requirement"
            run_tests
        fi
    else
        fail "Failed to create equipment movement requirement"
        run_tests
    fi

    test_name "Create equipment movement stop requirement"
    api_post "/v1/equipment-movement-stop-requirements" "{\"data\":{\"type\":\"equipment-movement-stop-requirements\",\"attributes\":{\"kind\":\"origin\"},\"relationships\":{\"stop\":{\"data\":{\"type\":\"equipment-movement-stops\",\"id\":\"$CREATED_STOP_ID\"}},\"requirement\":{\"data\":{\"type\":\"equipment-movement-requirements\",\"id\":\"$CREATED_REQUIREMENT_ID\"}}}}}"
    if [[ $status -eq 0 ]]; then
        CREATED_STOP_REQUIREMENT_ID=$(json_get ".data.id")
        if [[ -n "$CREATED_STOP_REQUIREMENT_ID" && "$CREATED_STOP_REQUIREMENT_ID" != "null" ]]; then
            pass
        else
            fail "Created stop requirement but no ID returned"
            echo "Cannot continue without stop requirement"
            run_tests
        fi
    else
        fail "Failed to create equipment movement stop requirement"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create stop completion with required fields"

xbe_json do equipment-movement-stop-completions create \
    --stop "$CREATED_STOP_ID" \
    --completed-at "$COMPLETED_AT" \
    --latitude "$LATITUDE" \
    --longitude "$LONGITUDE" \
    --note "Completed stop"

if [[ $status -eq 0 ]]; then
    CREATED_COMPLETION_ID=$(json_get ".id")
    if [[ -n "$CREATED_COMPLETION_ID" && "$CREATED_COMPLETION_ID" != "null" ]]; then
        register_cleanup "equipment-movement-stop-completions" "$CREATED_COMPLETION_ID"
        pass
    else
        fail "Created completion but no ID returned"
    fi
else
    fail "Failed to create stop completion"
fi

if [[ -z "$CREATED_COMPLETION_ID" || "$CREATED_COMPLETION_ID" == "null" ]]; then
    echo "Cannot continue without a valid stop completion ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update stop completion attributes"

xbe_json do equipment-movement-stop-completions update "$CREATED_COMPLETION_ID" \
    --completed-at "$UPDATED_COMPLETED_AT" \
    --latitude "$UPDATED_LATITUDE" \
    --longitude "$UPDATED_LONGITUDE" \
    --note "Updated completion"

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show stop completion"

xbe_json view equipment-movement-stop-completions show "$CREATED_COMPLETION_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show stop completion"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List stop completions"

xbe_json view equipment-movement-stop-completions list --limit 5
assert_success

test_name "List stop completions returns array"

xbe_json view equipment-movement-stop-completions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list stop completions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List stop completions with --stop filter"

xbe_json view equipment-movement-stop-completions list --stop "$CREATED_STOP_ID" --limit 5
assert_success

test_name "List stop completions with --equipment filter"

xbe_json view equipment-movement-stop-completions list --equipment "$CREATED_EQUIPMENT_ID" --limit 5
assert_success

test_name "List stop completions with --completed-at-min filter"

xbe_json view equipment-movement-stop-completions list --completed-at-min "$COMPLETED_AT" --limit 5
assert_success

test_name "List stop completions with --completed-at-max filter"

xbe_json view equipment-movement-stop-completions list --completed-at-max "$UPDATED_COMPLETED_AT" --limit 5
assert_success

test_name "List stop completions with --is-completed-at filter"

xbe_json view equipment-movement-stop-completions list --is-completed-at true --limit 5
assert_success

test_name "List stop completions with --created-at-min filter"

xbe_json view equipment-movement-stop-completions list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List stop completions with --created-at-max filter"

xbe_json view equipment-movement-stop-completions list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List stop completions with --is-created-at filter"

xbe_json view equipment-movement-stop-completions list --is-created-at true --limit 5
assert_success

test_name "List stop completions with --updated-at-min filter"

xbe_json view equipment-movement-stop-completions list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List stop completions with --updated-at-max filter"

xbe_json view equipment-movement-stop-completions list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List stop completions with --is-updated-at filter"

xbe_json view equipment-movement-stop-completions list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List stop completions with --limit"

xbe_json view equipment-movement-stop-completions list --limit 3
assert_success

test_name "List stop completions with --offset"

xbe_json view equipment-movement-stop-completions list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create stop completion without stop fails"

xbe_json do equipment-movement-stop-completions create --completed-at "$COMPLETED_AT"
assert_failure

test_name "Create stop completion without completed-at fails"

xbe_json do equipment-movement-stop-completions create --stop "$CREATED_STOP_ID"
assert_failure

test_name "Create stop completion with only latitude fails"

xbe_json do equipment-movement-stop-completions create --stop "$CREATED_STOP_ID" --completed-at "$COMPLETED_AT" --latitude "$LATITUDE"
assert_failure

test_name "Update stop completion without fields fails"

xbe_json do equipment-movement-stop-completions update "$CREATED_COMPLETION_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete stop completion requires --confirm flag"

xbe_run do equipment-movement-stop-completions delete "$CREATED_COMPLETION_ID"
assert_failure

test_name "Delete stop completion with --confirm"

xbe_run do equipment-movement-stop-completions delete "$CREATED_COMPLETION_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
