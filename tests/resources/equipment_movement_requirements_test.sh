#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Requirements
#
# Tests list, show, create, update operations for the equipment-movement-requirements resource.
#
# COVERAGE: All list filters + create/update attributes (with conditional relationship coverage)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REQUIREMENT_ID=""
CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_CUSTOMER_ID=""

SAMPLE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_EQUIPMENT_ID=""
SAMPLE_INBOUND_REQUIREMENT_ID=""
SAMPLE_OUTBOUND_REQUIREMENT_ID=""
SAMPLE_ORIGIN_ID=""
SAMPLE_DESTINATION_ID=""

RELATIONSHIP_REQUIREMENT_ID=""

describe "Resource: equipment-movement-requirements"

# ==========================================================================
# Sample data (used for relationship updates/filters when available)
# ==========================================================================

test_name "Capture sample requirement (if available)"
xbe_json view equipment-movement-requirements list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_EQUIPMENT_ID=$(json_get ".[0].equipment_id")
    SAMPLE_INBOUND_REQUIREMENT_ID=$(json_get ".[0].inbound_requirement_id")
    SAMPLE_OUTBOUND_REQUIREMENT_ID=$(json_get ".[0].outbound_requirement_id")
    SAMPLE_ORIGIN_ID=$(json_get ".[0].origin_id")
    SAMPLE_DESTINATION_ID=$(json_get ".[0].destination_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No sample requirement available"
    fi
else
    skip "Could not list requirements for sample"
fi

# ==========================================================================
# Prerequisites - Create broker, equipment classification, equipment, customer
# ==========================================================================

test_name "Create prerequisite broker for equipment movement requirement tests"
BROKER_NAME=$(unique_name "EquipMoveBroker")

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
CLASS_NAME=$(unique_name "EquipMoveClass")
CLASS_ABBR="EM$(date +%s | tail -c 4)"

xbe_json do equipment-classifications create --name "$CLASS_NAME" --abbreviation "$CLASS_ABBR"

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

test_name "Create prerequisite equipment"
EQUIPMENT_NICKNAME=$(unique_name "EquipMove")

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

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "EquipMoveCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without customer"
    run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create requirement with required fields"
xbe_json do equipment-movement-requirements create \
    --broker "$CREATED_BROKER_ID" \
    --equipment "$CREATED_EQUIPMENT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_REQUIREMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_REQUIREMENT_ID" && "$CREATED_REQUIREMENT_ID" != "null" ]]; then
        register_cleanup "equipment-movement-requirements" "$CREATED_REQUIREMENT_ID"
        pass
    else
        fail "Created requirement but no ID returned"
    fi
else
    fail "Failed to create requirement"
fi

if [[ -z "$CREATED_REQUIREMENT_ID" || "$CREATED_REQUIREMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid requirement ID"
    run_tests
fi

test_name "Create requirement with note"
xbe_json do equipment-movement-requirements create \
    --broker "$CREATED_BROKER_ID" \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --note "Test movement note"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-movement-requirements" "$id"
    pass
else
    fail "Failed to create requirement with note"
fi

test_name "Create requirement with timing"
xbe_json do equipment-movement-requirements create \
    --broker "$CREATED_BROKER_ID" \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --origin-at-min "2025-01-01T08:00:00Z" \
    --destination-at-max "2025-01-01T17:00:00Z"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-movement-requirements" "$id"
    pass
else
    fail "Failed to create requirement with timing"
fi

test_name "Create requirement with customer-explicit"
xbe_json do equipment-movement-requirements create \
    --broker "$CREATED_BROKER_ID" \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --customer-explicit "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-movement-requirements" "$id"
    pass
else
    fail "Failed to create requirement with customer-explicit"
fi

# ==========================================================================
# Relationship-focused create/update (optional if sample data available)
# ==========================================================================

test_name "Create requirement for relationship update tests"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" && -n "$SAMPLE_EQUIPMENT_ID" && "$SAMPLE_EQUIPMENT_ID" != "null" ]]; then
    xbe_json do equipment-movement-requirements create \
        --broker "$SAMPLE_BROKER_ID" \
        --equipment "$SAMPLE_EQUIPMENT_ID"
    if [[ $status -eq 0 ]]; then
        RELATIONSHIP_REQUIREMENT_ID=$(json_get ".id")
        if [[ -n "$RELATIONSHIP_REQUIREMENT_ID" && "$RELATIONSHIP_REQUIREMENT_ID" != "null" ]]; then
            register_cleanup "equipment-movement-requirements" "$RELATIONSHIP_REQUIREMENT_ID"
            pass
        else
            skip "No requirement ID returned for relationship tests"
        fi
    else
        skip "Could not create requirement for relationship tests"
    fi
else
    skip "No sample broker/equipment available for relationship tests"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update requirement note"
xbe_json do equipment-movement-requirements update "$CREATED_REQUIREMENT_ID" --note "Updated note"
assert_success

test_name "Update requirement timing"
xbe_json do equipment-movement-requirements update "$CREATED_REQUIREMENT_ID" \
    --origin-at-min "2025-01-02T08:00:00Z" \
    --destination-at-max "2025-01-02T17:00:00Z"
assert_success

test_name "Update requirement customer-explicit"
xbe_json do equipment-movement-requirements update "$CREATED_REQUIREMENT_ID" --customer-explicit "$CREATED_CUSTOMER_ID"
assert_success

test_name "Update requirement origin/destination"
if [[ -n "$RELATIONSHIP_REQUIREMENT_ID" && "$RELATIONSHIP_REQUIREMENT_ID" != "null" && -n "$SAMPLE_ORIGIN_ID" && "$SAMPLE_ORIGIN_ID" != "null" && -n "$SAMPLE_DESTINATION_ID" && "$SAMPLE_DESTINATION_ID" != "null" && "$SAMPLE_ORIGIN_ID" != "$SAMPLE_DESTINATION_ID" ]]; then
    xbe_json do equipment-movement-requirements update "$RELATIONSHIP_REQUIREMENT_ID" \
        --origin "$SAMPLE_ORIGIN_ID" \
        --destination "$SAMPLE_DESTINATION_ID"
    assert_success
else
    skip "No suitable origin/destination IDs available"
fi

test_name "Update requirement inbound/outbound"
if [[ -n "$RELATIONSHIP_REQUIREMENT_ID" && "$RELATIONSHIP_REQUIREMENT_ID" != "null" && -n "$SAMPLE_INBOUND_REQUIREMENT_ID" && "$SAMPLE_INBOUND_REQUIREMENT_ID" != "null" && -n "$SAMPLE_OUTBOUND_REQUIREMENT_ID" && "$SAMPLE_OUTBOUND_REQUIREMENT_ID" != "null" ]]; then
    xbe_json do equipment-movement-requirements update "$RELATIONSHIP_REQUIREMENT_ID" \
        --inbound-requirement "$SAMPLE_INBOUND_REQUIREMENT_ID" \
        --outbound-requirement "$SAMPLE_OUTBOUND_REQUIREMENT_ID"
    assert_success
else
    skip "No inbound/outbound requirement IDs available"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List equipment movement requirements"
xbe_json view equipment-movement-requirements list --limit 5
assert_success

test_name "List requirements returns array"
xbe_json view equipment-movement-requirements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list requirements"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List requirements with --broker filter"
xbe_json view equipment-movement-requirements list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List requirements with --equipment filter"
xbe_json view equipment-movement-requirements list --equipment "$CREATED_EQUIPMENT_ID" --limit 5
assert_success

test_name "List requirements with --inbound-requirement filter"
if [[ -n "$SAMPLE_INBOUND_REQUIREMENT_ID" && "$SAMPLE_INBOUND_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view equipment-movement-requirements list --inbound-requirement "$SAMPLE_INBOUND_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No inbound requirement ID available"
fi

test_name "List requirements with --outbound-requirement filter"
if [[ -n "$SAMPLE_OUTBOUND_REQUIREMENT_ID" && "$SAMPLE_OUTBOUND_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view equipment-movement-requirements list --outbound-requirement "$SAMPLE_OUTBOUND_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No outbound requirement ID available"
fi

test_name "List requirements with origin-at-min min/max filters"
xbe_json view equipment-movement-requirements list --origin-at-min-min "2020-01-01T00:00:00Z" --origin-at-min-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List requirements with destination-at-max min/max filters"
xbe_json view equipment-movement-requirements list --destination-at-max-min "2020-01-01T00:00:00Z" --destination-at-max-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List requirements with created-at-min filter"
xbe_json view equipment-movement-requirements list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List requirements with created-at-max filter"
xbe_json view equipment-movement-requirements list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List requirements with updated-at-min filter"
xbe_json view equipment-movement-requirements list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List requirements with updated-at-max filter"
xbe_json view equipment-movement-requirements list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show equipment movement requirement"
xbe_json view equipment-movement-requirements show "$CREATED_REQUIREMENT_ID"
assert_success

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create requirement without required fields fails"
xbe_run do equipment-movement-requirements create
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
