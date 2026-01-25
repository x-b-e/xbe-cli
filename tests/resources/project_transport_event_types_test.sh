#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Event Types
#
# Tests CRUD operations for the project_transport_event_types resource.
# These types define events that occur during transport operations.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TYPE_ID=""
CREATED_BROKER_ID=""

describe "Resource: project_transport_event_types"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for project transport event types tests"
BROKER_NAME=$(unique_name "PTETTestBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport event type with required fields"
TEST_NAME=$(unique_name "TransportEvent")
TEST_CODE="TE$(date +%s | tail -c 4)"

xbe_json do project-transport-event-types create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --code "$TEST_CODE"

if [[ $status -eq 0 ]]; then
    CREATED_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TYPE_ID" && "$CREATED_TYPE_ID" != "null" ]]; then
        register_cleanup "project-transport-event-types" "$CREATED_TYPE_ID"
        pass
    else
        fail "Created project transport event type but no ID returned"
    fi
else
    fail "Failed to create project transport event type"
fi

# Only continue if we successfully created a type
if [[ -z "$CREATED_TYPE_ID" || "$CREATED_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid project transport event type ID"
    run_tests
fi

test_name "Create project transport event type with different code"
TEST_NAME2=$(unique_name "TransportEvent2")
TEST_CODE2="T2$(date +%s | tail -c 4)"
xbe_json do project-transport-event-types create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --code "$TEST_CODE2"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-transport-event-types" "$id"
    pass
else
    fail "Failed to create project transport event type with code"
fi

test_name "Create project transport event type with dwell-minutes-min-default"
TEST_NAME3=$(unique_name "TransportEvent3")
TEST_CODE3="T3$(date +%s | tail -c 4)"
xbe_json do project-transport-event-types create \
    --name "$TEST_NAME3" \
    --broker "$CREATED_BROKER_ID" \
    --code "$TEST_CODE3" \
    --dwell-minutes-min-default "15"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-transport-event-types" "$id"
    pass
else
    fail "Failed to create project transport event type with dwell-minutes-min-default"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport event type name"
UPDATED_NAME=$(unique_name "UpdatedPTET")
xbe_json do project-transport-event-types update "$CREATED_TYPE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update project transport event type code"
xbe_json do project-transport-event-types update "$CREATED_TYPE_ID" --code "UPTE"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport event types"
xbe_json view project-transport-event-types list --limit 5
assert_success

test_name "List project transport event types returns array"
xbe_json view project-transport-event-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport event types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport event types with --broker filter"
xbe_json view project-transport-event-types list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project transport event types with --limit"
xbe_json view project-transport-event-types list --limit 3
assert_success

test_name "List project transport event types with --offset"
xbe_json view project-transport-event-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport event type requires --confirm flag"
xbe_run do project-transport-event-types delete "$CREATED_TYPE_ID"
assert_failure

test_name "Delete project transport event type with --confirm"
# Create a type specifically for deletion
TEST_DEL_NAME=$(unique_name "DeletePTET")
TEST_DEL_CODE="TD$(date +%s | tail -c 4)"
xbe_json do project-transport-event-types create \
    --name "$TEST_DEL_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --code "$TEST_DEL_CODE"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-transport-event-types delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project transport event type for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project transport event type without name fails"
xbe_json do project-transport-event-types create --broker "$CREATED_BROKER_ID" --code "TST"
assert_failure

test_name "Create project transport event type without broker fails"
xbe_json do project-transport-event-types create --name "NoBroker" --code "TST"
assert_failure

test_name "Create project transport event type without code fails"
xbe_json do project-transport-event-types create --name "NoCode" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-transport-event-types update "$CREATED_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
