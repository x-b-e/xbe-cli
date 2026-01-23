#!/bin/bash
#
# XBE CLI Integration Tests: Crew Rates
#
# Tests CRUD operations for the crew_rates resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_CRAFT_ID=""
CREATED_CRAFT_CLASS_ID=""
CREATED_CREW_RATE_ID=""

TEST_START_ON="2025-01-01"
TEST_END_ON="2025-12-31"
UPDATED_END_ON="2026-01-31"


describe "Resource: crew_rates"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for crew rate tests"
BROKER_NAME=$(unique_name "CrewRateBroker")

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

test_name "Create equipment classification for crew rate tests"
EQUIPMENT_CLASS_NAME=$(unique_name "CrewRateEquipmentClass")

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

test_name "Create equipment for crew rate tests"
EQUIPMENT_NAME=$(unique_name "CrewRateEquipment")

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

test_name "Create craft for crew rate tests"
CRAFT_NAME=$(unique_name "CrewRateCraft")

xbe_json do crafts create --name "$CRAFT_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_ID" && "$CREATED_CRAFT_ID" != "null" ]]; then
        register_cleanup "crafts" "$CREATED_CRAFT_ID"
        pass
    else
        fail "Created craft but no ID returned"
        echo "Cannot continue without craft"
        run_tests
    fi
else
    fail "Failed to create craft"
    echo "Cannot continue without craft"
    run_tests
fi

test_name "Create craft class for crew rate tests"
CRAFT_CLASS_NAME=$(unique_name "CrewRateCraftClass")

xbe_json do craft-classes create --name "$CRAFT_CLASS_NAME" --craft "$CREATED_CRAFT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
        register_cleanup "craft-classes" "$CREATED_CRAFT_CLASS_ID"
        pass
    else
        fail "Created craft class but no ID returned"
        echo "Cannot continue without craft class"
        run_tests
    fi
else
    fail "Failed to create craft class"
    echo "Cannot continue without craft class"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create crew rate with required fields"

xbe_json do crew-rates create \
    --price-per-unit "75.00" \
    --start-on "$TEST_START_ON" \
    --end-on "$TEST_END_ON" \
    --is-active true \
    --broker "$CREATED_BROKER_ID" \
    --resource-classification-type EquipmentClassification \
    --resource-classification-id "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --description "Test crew rate"

if [[ $status -eq 0 ]]; then
    CREATED_CREW_RATE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CREW_RATE_ID" && "$CREATED_CREW_RATE_ID" != "null" ]]; then
        register_cleanup "crew-rates" "$CREATED_CREW_RATE_ID"
        pass
    else
        fail "Created crew rate but no ID returned"
    fi
else
    fail "Failed to create crew rate"
fi

if [[ -z "$CREATED_CREW_RATE_ID" || "$CREATED_CREW_RATE_ID" == "null" ]]; then
    echo "Cannot continue without a valid crew rate ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update crew rate attributes"

xbe_json do crew-rates update "$CREATED_CREW_RATE_ID" \
    --description "Updated crew rate" \
    --price-per-unit "80.00" \
    --end-on "$UPDATED_END_ON" \
    --is-active false

assert_success

test_name "Update crew rate relationships"

xbe_json do crew-rates update "$CREATED_CREW_RATE_ID" \
    --resource-type Equipment \
    --resource-id "$CREATED_EQUIPMENT_ID" \
    --craft-class "$CREATED_CRAFT_CLASS_ID"

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show crew rate"

xbe_json view crew-rates show "$CREATED_CREW_RATE_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show crew rate"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List crew rates"

xbe_json view crew-rates list --limit 5
assert_success

test_name "List crew rates returns array"

xbe_json view crew-rates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list crew rates"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List crew rates with --broker filter"

xbe_json view crew-rates list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List crew rates with resource classification filter"

xbe_json view crew-rates list \
    --resource-classification-type EquipmentClassification \
    --resource-classification-id "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --limit 5
assert_success

test_name "List crew rates with resource filter"

xbe_json view crew-rates list \
    --resource-type Equipment \
    --resource-id "$CREATED_EQUIPMENT_ID" \
    --limit 5
assert_success

test_name "List crew rates with craft class filter"

xbe_json view crew-rates list --craft-class "$CREATED_CRAFT_CLASS_ID" --limit 5
assert_success

test_name "List crew rates with --is-active filter"

xbe_json view crew-rates list --is-active false --limit 5
assert_success

test_name "List crew rates with --start-on filter"

xbe_json view crew-rates list --start-on "$TEST_START_ON" --limit 5
assert_success

test_name "List crew rates with --start-on-min filter"

xbe_json view crew-rates list --start-on-min "$TEST_START_ON" --limit 5
assert_success

test_name "List crew rates with --start-on-max filter"

xbe_json view crew-rates list --start-on-max "$TEST_START_ON" --limit 5
assert_success

test_name "List crew rates with --end-on filter"

xbe_json view crew-rates list --end-on "$UPDATED_END_ON" --limit 5
assert_success

test_name "List crew rates with --end-on-min filter"

xbe_json view crew-rates list --end-on-min "$TEST_END_ON" --limit 5
assert_success

test_name "List crew rates with --end-on-max filter"

xbe_json view crew-rates list --end-on-max "$UPDATED_END_ON" --limit 5
assert_success

test_name "List crew rates with --search filter"

xbe_json view crew-rates list --search "Updated crew rate" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List crew rates with --limit"

xbe_json view crew-rates list --limit 3
assert_success

test_name "List crew rates with --offset"

xbe_json view crew-rates list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create crew rate without price-per-unit fails"

xbe_json do crew-rates create \
    --start-on "$TEST_START_ON" \
    --is-active true \
    --broker "$CREATED_BROKER_ID" \
    --resource-classification-type EquipmentClassification \
    --resource-classification-id "$CREATED_EQUIPMENT_CLASSIFICATION_ID"

assert_failure

test_name "Update crew rate without any fields fails"

xbe_json do crew-rates update "$CREATED_CREW_RATE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
