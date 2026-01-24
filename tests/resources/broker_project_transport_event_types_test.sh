#!/bin/bash
#
# XBE CLI Integration Tests: Broker Project Transport Event Types
#
# Tests CRUD operations for the broker_project_transport_event_types resource.
# These map broker-specific event type codes to project transport event types.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_EVENT_TYPE_ID=""
CREATED_BROKER_EVENT_TYPE_ID=""

describe "Resource: broker_project_transport_event_types"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for broker project transport event types tests"
BROKER_NAME=$(unique_name "BPTEBroker")

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
# Prerequisites - Create project transport event type
# ============================================================================

test_name "Create prerequisite project transport event type"
EVENT_TYPE_NAME=$(unique_name "BPTEType")
EVENT_TYPE_CODE="BPT$(date +%s | tail -c 4)"

xbe_json do project-transport-event-types create \
    --name "$EVENT_TYPE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --code "$EVENT_TYPE_CODE"

if [[ $status -eq 0 ]]; then
    CREATED_EVENT_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_EVENT_TYPE_ID" && "$CREATED_EVENT_TYPE_ID" != "null" ]]; then
        register_cleanup "project-transport-event-types" "$CREATED_EVENT_TYPE_ID"
        pass
    else
        fail "Created project transport event type but no ID returned"
        echo "Cannot continue without an event type"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID" ]]; then
        CREATED_EVENT_TYPE_ID="$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID"
        echo "    Using XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID: $CREATED_EVENT_TYPE_ID"
        pass
    else
        fail "Failed to create project transport event type and XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID not set"
        echo "Cannot continue without an event type"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker project transport event type with required fields"
BROKER_CODE="BP$(date +%s | tail -c 4)"

xbe_json do broker-project-transport-event-types create \
    --broker "$CREATED_BROKER_ID" \
    --project-transport-event-type "$CREATED_EVENT_TYPE_ID" \
    --code "$BROKER_CODE"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_EVENT_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_EVENT_TYPE_ID" && "$CREATED_BROKER_EVENT_TYPE_ID" != "null" ]]; then
        register_cleanup "broker-project-transport-event-types" "$CREATED_BROKER_EVENT_TYPE_ID"
        pass
    else
        fail "Created broker project transport event type but no ID returned"
    fi
else
    fail "Failed to create broker project transport event type"
fi

if [[ -z "$CREATED_BROKER_EVENT_TYPE_ID" || "$CREATED_BROKER_EVENT_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker project transport event type ID"
    run_tests
fi

test_name "Create broker project transport event type with different code"
BROKER_CODE2="BP2$(date +%s | tail -c 4)"
xbe_json do broker-project-transport-event-types create \
    --broker "$CREATED_BROKER_ID" \
    --project-transport-event-type "$CREATED_EVENT_TYPE_ID" \
    --code "$BROKER_CODE2"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "broker-project-transport-event-types" "$id"
    pass
else
    fail "Failed to create broker project transport event type with code"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update broker project transport event type code"
xbe_json do broker-project-transport-event-types update "$CREATED_BROKER_EVENT_TYPE_ID" --code "BPU"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker project transport event types"
xbe_json view broker-project-transport-event-types list --limit 5
assert_success

test_name "List broker project transport event types returns array"
xbe_json view broker-project-transport-event-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker project transport event types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List broker project transport event types with --broker filter"
xbe_json view broker-project-transport-event-types list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List broker project transport event types with --project-transport-event-type filter"
xbe_json view broker-project-transport-event-types list --project-transport-event-type "$CREATED_EVENT_TYPE_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broker project transport event types with --limit"
xbe_json view broker-project-transport-event-types list --limit 3
assert_success

test_name "List broker project transport event types with --offset"
xbe_json view broker-project-transport-event-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker project transport event type requires --confirm flag"
xbe_run do broker-project-transport-event-types delete "$CREATED_BROKER_EVENT_TYPE_ID"
assert_failure

test_name "Delete broker project transport event type with --confirm"
BROKER_CODE_DEL="BPD$(date +%s | tail -c 4)"
xbe_json do broker-project-transport-event-types create \
    --broker "$CREATED_BROKER_ID" \
    --project-transport-event-type "$CREATED_EVENT_TYPE_ID" \
    --code "$BROKER_CODE_DEL"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do broker-project-transport-event-types delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create broker project transport event type for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create broker project transport event type without code fails"
xbe_json do broker-project-transport-event-types create \
    --broker "$CREATED_BROKER_ID" \
    --project-transport-event-type "$CREATED_EVENT_TYPE_ID"
assert_failure

test_name "Create broker project transport event type without broker fails"
xbe_json do broker-project-transport-event-types create \
    --project-transport-event-type "$CREATED_EVENT_TYPE_ID" \
    --code "BPF"
assert_failure

test_name "Create broker project transport event type without project transport event type fails"
xbe_json do broker-project-transport-event-types create \
    --broker "$CREATED_BROKER_ID" \
    --code "BPF"
assert_failure

run_tests
