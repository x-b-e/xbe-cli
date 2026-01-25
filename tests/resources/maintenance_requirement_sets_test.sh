#!/bin/bash
#
# XBE CLI Integration Tests: Maintenance Requirement Sets
#
# Tests CRUD operations for the maintenance-requirement-sets resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BROKER_ID_2=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_WORK_ORDER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_MAINTENANCE_REQUIREMENT_SET_ID=""

WORK_ORDER_SET_ID=""

TEMPLATE_SET_ID=""

OPTIONAL_SET_ID=""

describe "Resource: maintenance-requirement-sets"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for maintenance requirement set tests"
BROKER_NAME=$(unique_name "MaintReqSetBroker")

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

test_name "Create secondary broker for broker update tests"
BROKER_NAME_2=$(unique_name "MaintReqSetBroker2")

xbe_json do brokers create --name "$BROKER_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID_2" && "$CREATED_BROKER_ID_2" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID_2"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    skip "Could not create secondary broker"
fi

test_name "Create prerequisite business unit for work order tests"
BUSINESS_UNIT_NAME=$(unique_name "MaintReqSetBU")

xbe_json do business-units create \
    --name "$BUSINESS_UNIT_NAME" \
    --broker "$CREATED_BROKER_ID"

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

test_name "Create prerequisite work order for maintenance requirement set tests"

xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_WORK_ORDER_ID=$(json_get ".id")
    if [[ -n "$CREATED_WORK_ORDER_ID" && "$CREATED_WORK_ORDER_ID" != "null" ]]; then
        register_cleanup "work-orders" "$CREATED_WORK_ORDER_ID"
        pass
    else
        fail "Created work order but no ID returned"
        echo "Cannot continue without a work order"
        run_tests
    fi
else
    fail "Failed to create work order"
    echo "Cannot continue without a work order"
    run_tests
fi

test_name "Create prerequisite equipment classification for maintenance requirement set tests"
EQUIPMENT_CLASS_NAME=$(unique_name "MaintReqSetEquipClass")

xbe_json do equipment-classifications create --name "$EQUIPMENT_CLASS_NAME"

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

test_name "Create prerequisite equipment for filter tests"
EQUIPMENT_NAME=$(unique_name "MaintReqSetEquip")

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
    fi
else
    skip "Failed to create equipment; equipment filter test will be skipped"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create maintenance requirement set with required fields"

xbe_json do maintenance-requirement-sets create \
    --maintenance-type maintenance \
    --broker "$CREATED_BROKER_ID" \
    --is-template=false

if [[ $status -eq 0 ]]; then
    CREATED_MAINTENANCE_REQUIREMENT_SET_ID=$(json_get ".id")
    if [[ -n "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" && "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-sets" "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID"
        pass
    else
        fail "Created maintenance requirement set but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement set"
fi

# Only continue if we successfully created a maintenance requirement set
if [[ -z "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" || "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" == "null" ]]; then
    echo "Cannot continue without a valid maintenance requirement set ID"
    run_tests
fi

test_name "Create maintenance requirement set with optional attributes and relationships"

xbe_json do maintenance-requirement-sets create \
    --maintenance-type maintenance \
    --broker "$CREATED_BROKER_ID" \
    --status ready_for_work \
    --is-archived \
    --is-template=false \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --work-order "$CREATED_WORK_ORDER_ID"

if [[ $status -eq 0 ]]; then
    OPTIONAL_SET_ID=$(json_get ".id")
    if [[ -n "$OPTIONAL_SET_ID" && "$OPTIONAL_SET_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-sets" "$OPTIONAL_SET_ID"
        pass
    else
        fail "Created maintenance requirement set but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement set with optional attributes"
fi

test_name "Create template maintenance requirement set"
TEMPLATE_NAME=$(unique_name "MaintReqTemplate")

xbe_json do maintenance-requirement-sets create \
    --maintenance-type inspection \
    --broker "$CREATED_BROKER_ID" \
    --is-template \
    --template-name "$TEMPLATE_NAME"

if [[ $status -eq 0 ]]; then
    TEMPLATE_SET_ID=$(json_get ".id")
    if [[ -n "$TEMPLATE_SET_ID" && "$TEMPLATE_SET_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-sets" "$TEMPLATE_SET_ID"
        pass
    else
        fail "Created template maintenance requirement set but no ID returned"
    fi
else
    fail "Failed to create template maintenance requirement set"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update maintenance requirement set --maintenance-type"
xbe_json do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" --maintenance-type inspection
assert_success

test_name "Update maintenance requirement set --status"
xbe_json do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" --status on_hold
assert_success

test_name "Update maintenance requirement set --is-archived"
xbe_json do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" --is-archived
assert_success

test_name "Update maintenance requirement set --equipment-classification"
xbe_json do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID"
assert_success

if [[ -n "$CREATED_BROKER_ID_2" ]]; then
    test_name "Update maintenance requirement set --broker"
    xbe_json do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" --broker "$CREATED_BROKER_ID_2"
    assert_success
else
    skip "Skipping broker update (no secondary broker)"
fi

test_name "Update maintenance requirement set --is-template and --template-name"
UPDATED_TEMPLATE_NAME=$(unique_name "MaintReqTemplateUpdated")

xbe_json do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID" \
    --is-template \
    --template-name "$UPDATED_TEMPLATE_NAME"
assert_success

# Work order relationship update on separate set (non-template)

test_name "Create maintenance requirement set for work order update"

xbe_json do maintenance-requirement-sets create \
    --maintenance-type maintenance \
    --broker "$CREATED_BROKER_ID" \
    --is-template=false

if [[ $status -eq 0 ]]; then
    WORK_ORDER_SET_ID=$(json_get ".id")
    if [[ -n "$WORK_ORDER_SET_ID" && "$WORK_ORDER_SET_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-sets" "$WORK_ORDER_SET_ID"
        pass
    else
        fail "Created maintenance requirement set but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement set for work order update"
fi

if [[ -n "$WORK_ORDER_SET_ID" && "$WORK_ORDER_SET_ID" != "null" ]]; then
    test_name "Update maintenance requirement set --work-order"
    xbe_json do maintenance-requirement-sets update "$WORK_ORDER_SET_ID" --work-order "$CREATED_WORK_ORDER_ID"
    assert_success
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show maintenance requirement set"
xbe_json view maintenance-requirement-sets show "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List maintenance requirement sets"
xbe_json view maintenance-requirement-sets list --limit 5
assert_success

test_name "List maintenance requirement sets returns array"
xbe_json view maintenance-requirement-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list maintenance requirement sets"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List maintenance requirement sets with --equipment-classification filter"
xbe_json view maintenance-requirement-sets list --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" --limit 5
assert_success

test_name "List maintenance requirement sets with --broker filter"
xbe_json view maintenance-requirement-sets list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List maintenance requirement sets with --equipment-business-unit filter"
xbe_json view maintenance-requirement-sets list --equipment-business-unit "$CREATED_BUSINESS_UNIT_ID" --limit 5
assert_success

if [[ -n "$CREATED_EQUIPMENT_ID" && "$CREATED_EQUIPMENT_ID" != "null" ]]; then
    test_name "List maintenance requirement sets with --equipment filter"
    xbe_json view maintenance-requirement-sets list --equipment "$CREATED_EQUIPMENT_ID" --limit 5
    assert_success
else
    skip "Skipping equipment filter test (no equipment ID)"
fi

test_name "List maintenance requirement sets with --maintenance-type filter"
xbe_json view maintenance-requirement-sets list --maintenance-type maintenance --limit 5
assert_success

test_name "List maintenance requirement sets with --status filter"
xbe_json view maintenance-requirement-sets list --status ready_for_work --limit 5
assert_success

test_name "List maintenance requirement sets with --is-template filter"
xbe_json view maintenance-requirement-sets list --is-template true --limit 5
assert_success

test_name "List maintenance requirement sets with --is-archived filter"
xbe_json view maintenance-requirement-sets list --is-archived true --limit 5
assert_success

test_name "List maintenance requirement sets with --completed-at-min filter"
xbe_json view maintenance-requirement-sets list --completed-at-min "2026-01-01T00:00:00Z" --limit 5
assert_success

test_name "List maintenance requirement sets with --completed-at-max filter"
xbe_json view maintenance-requirement-sets list --completed-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List maintenance requirement sets with --q filter"
if [[ -n "$TEMPLATE_NAME" ]]; then
    xbe_json view maintenance-requirement-sets list --q "$TEMPLATE_NAME" --limit 5
    assert_success
else
    skip "Skipping q filter test (no template name)"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List maintenance requirement sets with --limit"
xbe_json view maintenance-requirement-sets list --limit 2
assert_success

test_name "List maintenance requirement sets with --offset"
xbe_json view maintenance-requirement-sets list --limit 2 --offset 2
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete maintenance requirement set requires --confirm flag"
xbe_run do maintenance-requirement-sets delete "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID"
assert_failure

test_name "Delete maintenance requirement set with --confirm"
# Create a set specifically for deletion
xbe_json do maintenance-requirement-sets create \
    --maintenance-type maintenance \
    --broker "$CREATED_BROKER_ID" \
    --is-template=false

if [[ $status -eq 0 ]]; then
    DEL_SET_ID=$(json_get ".id")
    if [[ -n "$DEL_SET_ID" && "$DEL_SET_ID" != "null" ]]; then
        xbe_run do maintenance-requirement-sets delete "$DEL_SET_ID" --confirm
        assert_success
    else
        skip "Could not create maintenance requirement set for deletion test"
    fi
else
    skip "Could not create maintenance requirement set for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create maintenance requirement set without --maintenance-type fails"
xbe_json do maintenance-requirement-sets create \
    --broker "$CREATED_BROKER_ID" \
    --is-template=false
assert_failure

test_name "Create maintenance requirement set without --broker fails"
xbe_json do maintenance-requirement-sets create \
    --maintenance-type maintenance \
    --is-template=false
assert_failure

test_name "Update maintenance requirement set without any fields fails"
xbe_run do maintenance-requirement-sets update "$CREATED_MAINTENANCE_REQUIREMENT_SET_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
