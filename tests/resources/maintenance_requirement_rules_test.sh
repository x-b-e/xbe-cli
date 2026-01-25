#!/bin/bash
#
# XBE CLI Integration Tests: Maintenance Requirement Rules
#
# Tests CRUD operations for the maintenance_requirement_rules resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_SECOND_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_2_ID=""
CREATED_EQUIPMENT_ID=""
CREATED_EQUIPMENT_2_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_BUSINESS_UNIT_2_ID=""
CREATED_RULE_CLASS_ID=""
CREATED_RULE_EQUIP_ID=""
CREATED_RULE_BU_ID=""

RULE_TEXT_CLASS=""
RULE_TEXT_EQUIP=""
RULE_TEXT_BU=""
UPDATED_RULE_TEXT="Updated maintenance rule"

describe "Resource: maintenance_requirement_rules"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for maintenance requirement rule tests"
BROKER_NAME=$(unique_name "MaintenanceRuleBroker")

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

test_name "Create secondary broker for maintenance requirement rule updates"
BROKER2_NAME=$(unique_name "MaintenanceRuleBroker2")

xbe_json do brokers create --name "$BROKER2_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_SECOND_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_SECOND_BROKER_ID" && "$CREATED_SECOND_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_SECOND_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        CREATED_SECOND_BROKER_ID="$CREATED_BROKER_ID"
    fi
else
    CREATED_SECOND_BROKER_ID="$CREATED_BROKER_ID"
    echo "    Using primary broker for updates: $CREATED_SECOND_BROKER_ID"
    pass
fi

test_name "Create prerequisite equipment classification"
EC_NAME=$(unique_name "MaintReqClass")
EC_ABBR="MR$(date +%s | tail -c 4)$(printf '%02d' $((RANDOM % 100)))"

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

test_name "Create secondary equipment classification"
EC2_NAME=$(unique_name "MaintReqClass2")
EC2_ABBR="MR2$(date +%s | tail -c 4)$(printf '%02d' $((RANDOM % 100)))"

xbe_json do equipment-classifications create \
    --name "$EC2_NAME" \
    --abbreviation "$EC2_ABBR"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_CLASSIFICATION_2_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_CLASSIFICATION_2_ID" && "$CREATED_EQUIPMENT_CLASSIFICATION_2_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIPMENT_CLASSIFICATION_2_ID"
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

test_name "Create prerequisite business unit"
BU_NAME=$(unique_name "MaintReqBU")

xbe_json do business-units create --name "$BU_NAME" --broker "$CREATED_BROKER_ID"

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

test_name "Create secondary business unit"
BU2_NAME=$(unique_name "MaintReqBU2")

xbe_json do business-units create --name "$BU2_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_2_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_2_ID" && "$CREATED_BUSINESS_UNIT_2_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_2_ID"
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

test_name "Create prerequisite equipment"
EQUIPMENT_NICKNAME=$(unique_name "MaintReqEquip")

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

test_name "Create secondary equipment"
EQUIPMENT2_NICKNAME=$(unique_name "MaintReqEquip2")

xbe_json do equipment create \
    --nickname "$EQUIPMENT2_NICKNAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_2_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_2_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_2_ID" && "$CREATED_EQUIPMENT_2_ID" != "null" ]]; then
        register_cleanup "equipment" "$CREATED_EQUIPMENT_2_ID"
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

test_name "Create maintenance requirement rule with equipment classification"
RULE_TEXT_CLASS=$(unique_name "MaintRuleClass")

xbe_json do maintenance-requirement-rules create \
    --rule "$RULE_TEXT_CLASS" \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --is-active

if [[ $status -eq 0 ]]; then
    CREATED_RULE_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_RULE_CLASS_ID" && "$CREATED_RULE_CLASS_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-rules" "$CREATED_RULE_CLASS_ID"
        pass
    else
        fail "Created maintenance requirement rule but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement rule"
fi

test_name "Create maintenance requirement rule with equipment"
RULE_TEXT_EQUIP=$(unique_name "MaintRuleEquip")

xbe_json do maintenance-requirement-rules create \
    --rule "$RULE_TEXT_EQUIP" \
    --broker "$CREATED_BROKER_ID" \
    --equipment "$CREATED_EQUIPMENT_ID" \
    --is-active

if [[ $status -eq 0 ]]; then
    CREATED_RULE_EQUIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_RULE_EQUIP_ID" && "$CREATED_RULE_EQUIP_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-rules" "$CREATED_RULE_EQUIP_ID"
        pass
    else
        fail "Created maintenance requirement rule but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement rule"
fi

test_name "Create maintenance requirement rule with business unit"
RULE_TEXT_BU=$(unique_name "MaintRuleBU")

xbe_json do maintenance-requirement-rules create \
    --rule "$RULE_TEXT_BU" \
    --broker "$CREATED_BROKER_ID" \
    --business-unit "$CREATED_BUSINESS_UNIT_ID" \
    --is-active

if [[ $status -eq 0 ]]; then
    CREATED_RULE_BU_ID=$(json_get ".id")
    if [[ -n "$CREATED_RULE_BU_ID" && "$CREATED_RULE_BU_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-rules" "$CREATED_RULE_BU_ID"
        pass
    else
        fail "Created maintenance requirement rule but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement rule"
fi

if [[ -z "$CREATED_RULE_CLASS_ID" || "$CREATED_RULE_CLASS_ID" == "null" ]]; then
    echo "Cannot continue without a valid maintenance requirement rule ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update maintenance requirement rule attributes and broker"

xbe_json do maintenance-requirement-rules update "$CREATED_RULE_CLASS_ID" \
    --rule "$UPDATED_RULE_TEXT" \
    --is-active=false \
    --broker "$CREATED_SECOND_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_2_ID"

assert_success

test_name "Update maintenance requirement rule equipment relationship"

xbe_json do maintenance-requirement-rules update "$CREATED_RULE_EQUIP_ID" \
    --equipment "$CREATED_EQUIPMENT_2_ID"

assert_success

test_name "Update maintenance requirement rule business unit relationship"

xbe_json do maintenance-requirement-rules update "$CREATED_RULE_BU_ID" \
    --business-unit "$CREATED_BUSINESS_UNIT_2_ID"

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show maintenance requirement rule"

xbe_json view maintenance-requirement-rules show "$CREATED_RULE_CLASS_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show maintenance requirement rule"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List maintenance requirement rules"

xbe_json view maintenance-requirement-rules list --limit 5
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List maintenance requirement rules with --broker filter"

xbe_json view maintenance-requirement-rules list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List maintenance requirement rules with --equipment filter"

xbe_json view maintenance-requirement-rules list --equipment "$CREATED_EQUIPMENT_2_ID" --limit 5
assert_success

test_name "List maintenance requirement rules with --equipment-classification filter"

xbe_json view maintenance-requirement-rules list --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_2_ID" --limit 5
assert_success

test_name "List maintenance requirement rules with --business-unit filter"

xbe_json view maintenance-requirement-rules list --business-unit "$CREATED_BUSINESS_UNIT_2_ID" --limit 5
assert_success

test_name "List maintenance requirement rules with --is-active filter"

xbe_json view maintenance-requirement-rules list --is-active false --limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update maintenance requirement rule without any fields fails"

xbe_json do maintenance-requirement-rules update "$CREATED_RULE_CLASS_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
