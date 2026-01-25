#!/bin/bash
#
# XBE CLI Integration Tests: Equipment
#
# Tests CRUD operations for the equipment resource.
# Equipment represents tracked assets like tools, machines, and other items.
#
# NOTE: This test requires creating prerequisite resources: broker and equipment classification
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EQUIPMENT_ID=""
CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""

describe "Resource: equipment"

# ============================================================================
# Prerequisites - Create broker and equipment classification
# ============================================================================

test_name "Create prerequisite broker for equipment tests"
BROKER_NAME=$(unique_name "EquipTestBroker")

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
EC_NAME=$(unique_name "EquipClass")
EC_ABBR="EC$(date +%s | tail -c 4)"

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
# CREATE Tests
# ============================================================================

test_name "Create equipment with required fields"

EQUIPMENT_NICKNAME=$(unique_name "Excavator")

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

# Only continue if we successfully created equipment
if [[ -z "$CREATED_EQUIPMENT_ID" || "$CREATED_EQUIPMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment ID"
    run_tests
fi

test_name "Create equipment with --serial-number"
EQUIPMENT2_NICKNAME=$(unique_name "Loader")
xbe_json do equipment create \
    --nickname "$EQUIPMENT2_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --serial-number "SN-12345"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment" "$id"
    pass
else
    fail "Failed to create equipment with serial-number"
fi

test_name "Create equipment with --manufacturer-name"
EQUIPMENT3_NICKNAME=$(unique_name "Dozer")
xbe_json do equipment create \
    --nickname "$EQUIPMENT3_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --manufacturer-name "Caterpillar"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment" "$id"
    pass
else
    fail "Failed to create equipment with manufacturer-name"
fi

test_name "Create equipment with --model-description"
EQUIPMENT4_NICKNAME=$(unique_name "Crane")
xbe_json do equipment create \
    --nickname "$EQUIPMENT4_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --model-description "Model 320D"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment" "$id"
    pass
else
    fail "Failed to create equipment with model-description"
fi

test_name "Create equipment with --year"
EQUIPMENT5_NICKNAME=$(unique_name "Grader")
xbe_json do equipment create \
    --nickname "$EQUIPMENT5_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --year "2020"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment" "$id"
    pass
else
    fail "Failed to create equipment with year"
fi

test_name "Create equipment with --description"
EQUIPMENT6_NICKNAME=$(unique_name "Roller")
xbe_json do equipment create \
    --nickname "$EQUIPMENT6_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --description "Heavy duty roller"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment" "$id"
    pass
else
    fail "Failed to create equipment with description"
fi

test_name "Create equipment with --is-active false"
EQUIPMENT7_NICKNAME=$(unique_name "Paver")
xbe_json do equipment create \
    --nickname "$EQUIPMENT7_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --is-active=false
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment" "$id"
    pass
else
    fail "Failed to create equipment with is-active false"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update equipment --nickname"
NEW_NICKNAME=$(unique_name "UpdatedExcavator")
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --nickname "$NEW_NICKNAME"
assert_success

test_name "Update equipment --serial-number"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --serial-number "SN-99999"
assert_success

test_name "Update equipment --manufacturer-name"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --manufacturer-name "Komatsu"
assert_success

test_name "Update equipment --model-description"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --model-description "Model PC200"
assert_success

test_name "Update equipment --year"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --year "2021"
assert_success

test_name "Update equipment --description"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --description "Updated description"
assert_success

test_name "Update equipment --is-active"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --is-active=false
assert_success

test_name "Update equipment --is-active back to true"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID" --is-active=true
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment"
xbe_json view equipment list --limit 5
assert_success

test_name "List equipment returns array"
xbe_json view equipment list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List equipment with --equipment-classification filter"
xbe_json view equipment list --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" --limit 10
assert_success

test_name "List equipment with --broker filter"
xbe_json view equipment list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List equipment with --is-active filter"
xbe_json view equipment list --is-active true --limit 10
assert_success

test_name "List equipment with --nickname-like filter"
xbe_json view equipment list --nickname-like "$NEW_NICKNAME" --limit 10
assert_success

test_name "List equipment with --search filter"
xbe_json view equipment list --search "$NEW_NICKNAME" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List equipment with --limit"
xbe_json view equipment list --limit 3
assert_success

test_name "List equipment with --offset"
xbe_json view equipment list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment requires --confirm flag"
xbe_run do equipment delete "$CREATED_EQUIPMENT_ID"
assert_failure

test_name "Delete equipment with --confirm"
# Create equipment specifically for deletion
DEL_NICKNAME=$(unique_name "DelEquip")
xbe_json do equipment create \
    --nickname "$DEL_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_EQUIPMENT_ID=$(json_get ".id")
    xbe_run do equipment delete "$DEL_EQUIPMENT_ID" --confirm
    assert_success
else
    skip "Could not create equipment for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create equipment without --nickname fails"
xbe_json do equipment create \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create equipment without --equipment-classification fails"
xbe_json do equipment create \
    --nickname "Test" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create equipment without --organization-type fails"
xbe_json do equipment create \
    --nickname "Test" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create equipment without --organization-id fails"
xbe_json do equipment create \
    --nickname "Test" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers"
assert_failure

test_name "Update without any fields fails"
xbe_json do equipment update "$CREATED_EQUIPMENT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
