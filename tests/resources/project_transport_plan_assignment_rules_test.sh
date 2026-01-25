#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Assignment Rules
#
# Tests CRUD operations for the project_transport_plan_assignment_rules resource.
# These rules define how drivers/tractors/trailers are assigned in transport plans.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RULE_ID=""
CREATED_BROKER_ID=""

describe "Resource: project_transport_plan_assignment_rules"

# ==========================================================================
# Prerequisites - Create broker
# ==========================================================================

test_name "Create prerequisite broker for assignment rule tests"
BROKER_NAME=$(unique_name "PTPARTestBroker")

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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create assignment rule with required fields"
RULE_TEXT=$(unique_name "AssignRule")

xbe_json do project-transport-plan-assignment-rules create \
    --rule "$RULE_TEXT" \
    --asset-type driver \
    --level-type brokers \
    --level-id "$CREATED_BROKER_ID" \
    --is-active

if [[ $status -eq 0 ]]; then
    CREATED_RULE_ID=$(json_get ".id")
    if [[ -n "$CREATED_RULE_ID" && "$CREATED_RULE_ID" != "null" ]]; then
        register_cleanup "project-transport-plan-assignment-rules" "$CREATED_RULE_ID"
        pass
    else
        fail "Created assignment rule but no ID returned"
    fi
else
    fail "Failed to create assignment rule"
fi

# Only continue if we successfully created a rule
if [[ -z "$CREATED_RULE_ID" || "$CREATED_RULE_ID" == "null" ]]; then
    echo "Cannot continue without a valid assignment rule ID"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show assignment rule"
xbe_json view project-transport-plan-assignment-rules show "$CREATED_RULE_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update assignment rule text"
UPDATED_RULE=$(unique_name "UpdatedRule")
xbe_json do project-transport-plan-assignment-rules update "$CREATED_RULE_ID" --rule "$UPDATED_RULE"
assert_success

test_name "Update assignment rule asset type"
xbe_json do project-transport-plan-assignment-rules update "$CREATED_RULE_ID" --asset-type tractor
assert_success

test_name "Update assignment rule is-active"
xbe_json do project-transport-plan-assignment-rules update "$CREATED_RULE_ID" --is-active=false
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List assignment rules"
xbe_json view project-transport-plan-assignment-rules list --limit 5
assert_success

test_name "List assignment rules returns array"
xbe_json view project-transport-plan-assignment-rules list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list assignment rules"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List assignment rules with --asset-type filter"
xbe_json view project-transport-plan-assignment-rules list --asset-type tractor --limit 10
assert_success

test_name "List assignment rules with --level filter"
xbe_json view project-transport-plan-assignment-rules list --level "brokers|$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List assignment rules with --level-type filter"
xbe_json view project-transport-plan-assignment-rules list --level-type brokers --limit 10
assert_success

test_name "List assignment rules with --level-id filter"
xbe_json view project-transport-plan-assignment-rules list --level-id "$CREATED_BROKER_ID" --level-type brokers --limit 10
assert_success

test_name "List assignment rules with --not-level-type filter"
xbe_json view project-transport-plan-assignment-rules list --not-level-type broker --limit 10
assert_success

test_name "List assignment rules with --broker filter"
xbe_json view project-transport-plan-assignment-rules list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List assignment rules with --is-active filter"
xbe_json view project-transport-plan-assignment-rules list --is-active false --limit 10
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete assignment rule requires --confirm flag"
xbe_run do project-transport-plan-assignment-rules delete "$CREATED_RULE_ID"
assert_failure

# Create rule for deletion
RULE_DEL_TEXT=$(unique_name "DeleteRule")
xbe_json do project-transport-plan-assignment-rules create \
    --rule "$RULE_DEL_TEXT" \
    --asset-type trailer \
    --level-type brokers \
    --level-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-transport-plan-assignment-rules delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create assignment rule for deletion test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create assignment rule without rule fails"
xbe_json do project-transport-plan-assignment-rules create \
    --asset-type driver \
    --level-type brokers \
    --level-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create assignment rule without asset-type fails"
xbe_json do project-transport-plan-assignment-rules create \
    --rule "Missing asset type" \
    --level-type brokers \
    --level-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create assignment rule without level-id fails"
xbe_json do project-transport-plan-assignment-rules create \
    --rule "Missing level" \
    --asset-type driver \
    --level-type brokers
assert_failure

test_name "Update without any fields fails"
xbe_json do project-transport-plan-assignment-rules update "$CREATED_RULE_ID"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
