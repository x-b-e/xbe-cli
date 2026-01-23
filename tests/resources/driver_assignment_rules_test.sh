#!/bin/bash
#
# XBE CLI Integration Tests: Driver Assignment Rules
#
# Tests CRUD operations for the driver_assignment_rules resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DRIVER_ASSIGNMENT_RULE_ID=""

RULE_TEXT=""
UPDATED_RULE_TEXT="Updated driver assignment rule"

describe "Resource: driver_assignment_rules"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for driver assignment rule tests"
BROKER_NAME=$(unique_name "DriverAssignmentRuleBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create driver assignment rule with required fields"
RULE_TEXT=$(unique_name "Driver assignment rule")

xbe_json do driver-assignment-rules create \
    --rule "$RULE_TEXT" \
    --level-type Broker \
    --level-id "$CREATED_BROKER_ID" \
    --is-active

if [[ $status -eq 0 ]]; then
    CREATED_DRIVER_ASSIGNMENT_RULE_ID=$(json_get ".id")
    if [[ -n "$CREATED_DRIVER_ASSIGNMENT_RULE_ID" && "$CREATED_DRIVER_ASSIGNMENT_RULE_ID" != "null" ]]; then
        register_cleanup "driver-assignment-rules" "$CREATED_DRIVER_ASSIGNMENT_RULE_ID"
        pass
    else
        fail "Created driver assignment rule but no ID returned"
    fi
else
    fail "Failed to create driver assignment rule"
fi

if [[ -z "$CREATED_DRIVER_ASSIGNMENT_RULE_ID" || "$CREATED_DRIVER_ASSIGNMENT_RULE_ID" == "null" ]]; then
    echo "Cannot continue without a valid driver assignment rule ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update driver assignment rule attributes"

xbe_json do driver-assignment-rules update "$CREATED_DRIVER_ASSIGNMENT_RULE_ID" \
    --rule "$UPDATED_RULE_TEXT" \
    --is-active=false

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver assignment rule"

xbe_json view driver-assignment-rules show "$CREATED_DRIVER_ASSIGNMENT_RULE_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show driver assignment rule"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver assignment rules"

xbe_json view driver-assignment-rules list --limit 5
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List driver assignment rules with --broker filter"

xbe_json view driver-assignment-rules list --broker "$CREATED_BROKER_ID" --limit 5
assert_success


test_name "List driver assignment rules with --level-type/--level-id filter"

xbe_json view driver-assignment-rules list \
    --level-type Broker \
    --level-id "$CREATED_BROKER_ID" \
    --limit 5

assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update driver assignment rule without any fields fails"

xbe_json do driver-assignment-rules update "$CREATED_DRIVER_ASSIGNMENT_RULE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
