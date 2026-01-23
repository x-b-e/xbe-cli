#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Utilization Readings
#
# Tests CRUD operations for the equipment-utilization-readings resource.
# Equipment utilization readings require an equipment relationship and at least
# one reading value (odometer or hourmeter).
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_BUSINESS_UNIT_ID_2=""
CREATED_READING_ID_1=""
CREATED_READING_ID_2=""
CURRENT_USER_ID=""
HAS_USER="false"

REPORTED_AT_1="2025-01-05T08:00:00Z"
REPORTED_AT_2="2025-01-06T08:00:00Z"
REPORTED_AT_2_UPDATED="2025-01-06T09:00:00Z"

SOURCE_TELEMATICS='{"source":"telematics"}'
SOURCE_MANUAL='{"source":"manual"}'

describe "Resource: equipment-utilization-readings"

# ============================================================================
# Prerequisites - Create broker, equipment classification, equipment, business unit
# ============================================================================

test_name "Create prerequisite broker for equipment utilization readings tests"
BROKER_NAME=$(unique_name "UtilizationBroker")

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
EC_NAME=$(unique_name "UtilizationClass")
EC_ABBR="EU$(date +%s | tail -c 4)"

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

test_name "Create prerequisite equipment"
EQUIPMENT_NICKNAME=$(unique_name "UtilizationEquip")

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
    fi
else
    fail "Failed to create equipment"
fi

if [[ -z "$CREATED_EQUIPMENT_ID" || "$CREATED_EQUIPMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment ID"
    run_tests
fi

test_name "Create prerequisite business unit"
BUSINESS_UNIT_NAME=$(unique_name "UtilizationBU")

xbe_json do business-units create --name "$BUSINESS_UNIT_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business unit but no ID returned"
        echo "Cannot continue without a business unit"
        run_tests
    fi
else
    fail "Failed to create business unit"
    echo "Cannot continue without a business unit"
    run_tests
fi

test_name "Fetch current user id"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        HAS_USER="true"
        pass
    else
        skip "No user ID returned"
    fi
else
    skip "Failed to fetch current user"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment utilization reading with odometer"
CREATE_ARGS=(
    --equipment "$CREATED_EQUIPMENT_ID"
    --reported-at "$REPORTED_AT_1"
    --odometer "100"
    --business-unit "$CREATED_BUSINESS_UNIT_ID"
)
if [[ "$HAS_USER" == "true" ]]; then
    CREATE_ARGS+=(--user "$CURRENT_USER_ID")
fi

xbe_json do equipment-utilization-readings create "${CREATE_ARGS[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_READING_ID_1=$(json_get ".id")
    if [[ -n "$CREATED_READING_ID_1" && "$CREATED_READING_ID_1" != "null" ]]; then
        register_cleanup "equipment-utilization-readings" "$CREATED_READING_ID_1"
        pass
    else
        fail "Created reading but no ID returned"
    fi
else
    fail "Failed to create equipment utilization reading"
fi

if [[ -z "$CREATED_READING_ID_1" || "$CREATED_READING_ID_1" == "null" ]]; then
    echo "Cannot continue without a valid reading ID"
    run_tests
fi

test_name "Create equipment utilization reading with hourmeter and source"
CREATE_ARGS=(
    --equipment "$CREATED_EQUIPMENT_ID"
    --reported-at "$REPORTED_AT_2"
    --hourmeter "12"
    --other-readings "$SOURCE_TELEMATICS"
)
if [[ "$HAS_USER" == "true" ]]; then
    CREATE_ARGS+=(--user "$CURRENT_USER_ID")
fi

xbe_json do equipment-utilization-readings create "${CREATE_ARGS[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_READING_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_READING_ID_2" && "$CREATED_READING_ID_2" != "null" ]]; then
        register_cleanup "equipment-utilization-readings" "$CREATED_READING_ID_2"
        pass
    else
        fail "Created reading but no ID returned"
    fi
else
    fail "Failed to create equipment utilization reading with hourmeter"
fi

if [[ -z "$CREATED_READING_ID_2" || "$CREATED_READING_ID_2" == "null" ]]; then
    echo "Cannot continue without a valid reading ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment utilization reading"
xbe_json view equipment-utilization-readings show "$CREATED_READING_ID_1"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment utilization readings"
xbe_json view equipment-utilization-readings list --limit 5
assert_success

test_name "List equipment utilization readings returns array"
xbe_json view equipment-utilization-readings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment utilization readings"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List equipment utilization readings with --equipment filter"
xbe_json view equipment-utilization-readings list --equipment "$CREATED_EQUIPMENT_ID" --limit 10
assert_success

test_name "List equipment utilization readings with --business-unit filter"
xbe_json view equipment-utilization-readings list --business-unit "$CREATED_BUSINESS_UNIT_ID" --limit 10
assert_success

if [[ "$HAS_USER" == "true" ]]; then
    test_name "List equipment utilization readings with --user filter"
    xbe_json view equipment-utilization-readings list --user "$CURRENT_USER_ID" --limit 10
    assert_success
else
    test_name "List equipment utilization readings with --user filter"
    skip "No user ID available"
fi

test_name "List equipment utilization readings with --reported-at-min filter"
xbe_json view equipment-utilization-readings list --reported-at-min "2025-01-01T00:00:00Z" --limit 10
assert_success

test_name "List equipment utilization readings with --reported-at-max filter"
xbe_json view equipment-utilization-readings list --reported-at-max "2025-12-31T23:59:59Z" --limit 10
assert_success

test_name "List equipment utilization readings with --source manual filter"
xbe_json view equipment-utilization-readings list --source manual --limit 10
assert_success

test_name "List equipment utilization readings with --source telematics filter"
xbe_json view equipment-utilization-readings list --source telematics --limit 10
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Create second business unit for update tests"
BUSINESS_UNIT_NAME_2=$(unique_name "UtilizationBU2")

xbe_json do business-units create --name "$BUSINESS_UNIT_NAME_2" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID_2" && "$CREATED_BUSINESS_UNIT_ID_2" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID_2"
        pass
    else
        fail "Created business unit but no ID returned"
        echo "Cannot continue without a business unit"
        run_tests
    fi
else
    fail "Failed to create business unit"
    echo "Cannot continue without a business unit"
    run_tests
fi

test_name "Update equipment utilization reading odometer"
xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2" --odometer "110"
assert_success

test_name "Update equipment utilization reading hourmeter"
xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2" --hourmeter "15"
assert_success

test_name "Update equipment utilization reading reported-at"
xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2" --reported-at "$REPORTED_AT_2_UPDATED"
assert_success

test_name "Update equipment utilization reading business unit"
xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2" --business-unit "$CREATED_BUSINESS_UNIT_ID_2"
assert_success

if [[ "$HAS_USER" == "true" ]]; then
    test_name "Update equipment utilization reading user"
    xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2" --user "$CURRENT_USER_ID"
    assert_success
else
    test_name "Update equipment utilization reading user"
    skip "No user ID available"
fi

test_name "Update equipment utilization reading other readings"
xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2" --other-readings "$SOURCE_MANUAL"
assert_success

test_name "Update without any fields fails"
xbe_json do equipment-utilization-readings update "$CREATED_READING_ID_2"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment utilization reading requires --confirm flag"
xbe_json do equipment-utilization-readings delete "$CREATED_READING_ID_1"
assert_failure

test_name "Delete equipment utilization reading with --confirm"
# Create a reading specifically for deletion
xbe_json do equipment-utilization-readings create \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --reported-at "2025-01-07T08:00:00Z" \
    --odometer "130"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do equipment-utilization-readings delete "$DEL_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        register_cleanup "equipment-utilization-readings" "$DEL_ID"
        skip "API may not allow equipment utilization reading deletion"
    fi
else
    skip "Could not create reading for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create equipment utilization reading without equipment fails"
xbe_json do equipment-utilization-readings create --reported-at "$REPORTED_AT_1" --odometer "50"
assert_failure

test_name "Create equipment utilization reading without reported-at fails"
xbe_json do equipment-utilization-readings create --equipment "$CREATED_EQUIPMENT_ID" --odometer "50"
assert_failure

test_name "Create equipment utilization reading without odometer or hourmeter fails"
xbe_json do equipment-utilization-readings create --equipment "$CREATED_EQUIPMENT_ID" --reported-at "$REPORTED_AT_1"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
