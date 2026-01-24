#!/bin/bash
#
# XBE CLI Integration Tests: Broker Equipment Classifications
#
# Tests CRUD operations for the broker_equipment_classifications resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
PARENT_EQUIP_CLASS_ID=""
CHILD_EQUIP_CLASS_ID=""
CREATED_BROKER_EQUIP_CLASS_ID=""
SECOND_BROKER_ID=""
SECOND_CHILD_EQUIP_CLASS_ID=""

describe "Resource: broker_equipment_classifications"

# ============================================================================
# Prerequisites - Create broker and equipment classifications
# ============================================================================

test_name "Create prerequisite broker for broker equipment classification tests"
BROKER_NAME=$(unique_name "BECBroker")

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

test_name "Create parent equipment classification"
PARENT_CLASS_NAME=$(unique_name "BECParent")

xbe_json do equipment-classifications create --name "$PARENT_CLASS_NAME"

if [[ $status -eq 0 ]]; then
    PARENT_EQUIP_CLASS_ID=$(json_get ".id")
    if [[ -n "$PARENT_EQUIP_CLASS_ID" && "$PARENT_EQUIP_CLASS_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$PARENT_EQUIP_CLASS_ID"
        pass
    else
        fail "Created parent equipment classification but no ID returned"
        echo "Cannot continue without a parent equipment classification"
        run_tests
    fi
else
    fail "Failed to create parent equipment classification"
    echo "Cannot continue without a parent equipment classification"
    run_tests
fi

test_name "Create child equipment classification"
CHILD_CLASS_NAME=$(unique_name "BECChild")

xbe_json do equipment-classifications create --name "$CHILD_CLASS_NAME" --parent "$PARENT_EQUIP_CLASS_ID"

if [[ $status -eq 0 ]]; then
    CHILD_EQUIP_CLASS_ID=$(json_get ".id")
    if [[ -n "$CHILD_EQUIP_CLASS_ID" && "$CHILD_EQUIP_CLASS_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CHILD_EQUIP_CLASS_ID"
        pass
    else
        fail "Created child equipment classification but no ID returned"
        echo "Cannot continue without a child equipment classification"
        run_tests
    fi
else
    fail "Failed to create child equipment classification"
    echo "Cannot continue without a child equipment classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker equipment classification with required fields"
xbe_json do broker-equipment-classifications create \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CHILD_EQUIP_CLASS_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_EQUIP_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_EQUIP_CLASS_ID" && "$CREATED_BROKER_EQUIP_CLASS_ID" != "null" ]]; then
        register_cleanup "broker-equipment-classifications" "$CREATED_BROKER_EQUIP_CLASS_ID"
        pass
    else
        fail "Created broker equipment classification but no ID returned"
    fi
else
    fail "Failed to create broker equipment classification"
fi

if [[ -z "$CREATED_BROKER_EQUIP_CLASS_ID" || "$CREATED_BROKER_EQUIP_CLASS_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker equipment classification ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker equipment classification"
xbe_json view broker-equipment-classifications show "$CREATED_BROKER_EQUIP_CLASS_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Create secondary broker for update"
SECOND_BROKER_NAME=$(unique_name "BECBroker2")

xbe_json do brokers create --name "$SECOND_BROKER_NAME"

if [[ $status -eq 0 ]]; then
    SECOND_BROKER_ID=$(json_get ".id")
    if [[ -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$SECOND_BROKER_ID"
        pass
    else
        fail "Created second broker but no ID returned"
        echo "Cannot continue without a second broker"
        run_tests
    fi
else
    fail "Failed to create second broker"
    echo "Cannot continue without a second broker"
    run_tests
fi

test_name "Create second child equipment classification"
SECOND_CHILD_CLASS_NAME=$(unique_name "BECChild2")

xbe_json do equipment-classifications create --name "$SECOND_CHILD_CLASS_NAME" --parent "$PARENT_EQUIP_CLASS_ID"

if [[ $status -eq 0 ]]; then
    SECOND_CHILD_EQUIP_CLASS_ID=$(json_get ".id")
    if [[ -n "$SECOND_CHILD_EQUIP_CLASS_ID" && "$SECOND_CHILD_EQUIP_CLASS_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$SECOND_CHILD_EQUIP_CLASS_ID"
        pass
    else
        fail "Created second child equipment classification but no ID returned"
        echo "Cannot continue without a second child equipment classification"
        run_tests
    fi
else
    fail "Failed to create second child equipment classification"
    echo "Cannot continue without a second child equipment classification"
    run_tests
fi

test_name "Update broker equipment classification broker and equipment classification"
xbe_json do broker-equipment-classifications update "$CREATED_BROKER_EQUIP_CLASS_ID" \
    --broker "$SECOND_BROKER_ID" \
    --equipment-classification "$SECOND_CHILD_EQUIP_CLASS_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker equipment classifications"
xbe_json view broker-equipment-classifications list --limit 5
assert_success

test_name "List broker equipment classifications returns array"
xbe_json view broker-equipment-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker equipment classifications"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List broker equipment classifications with --broker filter"
xbe_json view broker-equipment-classifications list --broker "$SECOND_BROKER_ID" --limit 5
assert_success

test_name "List broker equipment classifications with --equipment-classification filter"
xbe_json view broker-equipment-classifications list --equipment-classification "$SECOND_CHILD_EQUIP_CLASS_ID" --limit 5
assert_success

test_name "List broker equipment classifications with --created-at-min filter"
xbe_json view broker-equipment-classifications list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker equipment classifications with --created-at-max filter"
xbe_json view broker-equipment-classifications list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker equipment classifications with --is-created-at filter"
xbe_json view broker-equipment-classifications list --is-created-at true --limit 5
assert_success

test_name "List broker equipment classifications with --updated-at-min filter"
xbe_json view broker-equipment-classifications list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker equipment classifications with --updated-at-max filter"
xbe_json view broker-equipment-classifications list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker equipment classifications with --is-updated-at filter"
xbe_json view broker-equipment-classifications list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broker equipment classifications with --limit"
xbe_json view broker-equipment-classifications list --limit 3
assert_success

test_name "List broker equipment classifications with --offset"
xbe_json view broker-equipment-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create broker equipment classification without broker fails"
xbe_json do broker-equipment-classifications create --equipment-classification "$CHILD_EQUIP_CLASS_ID"
assert_failure

test_name "Create broker equipment classification without equipment classification fails"
xbe_json do broker-equipment-classifications create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update broker equipment classification without any fields fails"
xbe_json do broker-equipment-classifications update "$CREATED_BROKER_EQUIP_CLASS_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker equipment classification requires --confirm flag"
xbe_run do broker-equipment-classifications delete "$CREATED_BROKER_EQUIP_CLASS_ID"
assert_failure

test_name "Delete broker equipment classification with --confirm"
xbe_run do broker-equipment-classifications delete "$CREATED_BROKER_EQUIP_CLASS_ID" --confirm
assert_success

run_tests
