#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Location Event Types
#
# Tests view and create/delete operations for the project_transport_location_event_types resource.
# These links attach transport event types to transport locations.
#
# COVERAGE: Create attributes + list filters + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LINK_ID=""
CREATED_BROKER_ID=""
CREATED_EVENT_TYPE_ID=""

LOCATION_ID="${XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID:-}"
EVENT_TYPE_ID="${XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID:-}"

describe "Resource: project-transport-location-event-types"

# ============================================================================
# Prerequisites - Create broker and event type if needed
# ============================================================================

if [[ -z "$EVENT_TYPE_ID" ]]; then
    test_name "Create prerequisite broker for project transport event type"
    BROKER_NAME=$(unique_name "PTLETBroker")

    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        fi
    fi

    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        test_name "Create prerequisite project transport event type"
        EVENT_NAME=$(unique_name "PTLETEvent")
        EVENT_CODE="PTL$(date +%s | tail -c 4)"

        xbe_json do project-transport-event-types create \
            --name "$EVENT_NAME" \
            --code "$EVENT_CODE" \
            --broker "$CREATED_BROKER_ID"

        if [[ $status -eq 0 ]]; then
            CREATED_EVENT_TYPE_ID=$(json_get ".id")
            if [[ -n "$CREATED_EVENT_TYPE_ID" && "$CREATED_EVENT_TYPE_ID" != "null" ]]; then
                register_cleanup "project-transport-event-types" "$CREATED_EVENT_TYPE_ID"
                EVENT_TYPE_ID="$CREATED_EVENT_TYPE_ID"
                pass
            else
                fail "Created project transport event type but no ID returned"
            fi
        else
            fail "Failed to create project transport event type"
        fi
    fi
else
    test_name "Using existing project transport event type"
    echo "    Using XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID: $EVENT_TYPE_ID"
    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport location event type without required fields fails"
xbe_json do project-transport-location-event-types create
assert_failure

if [[ -z "$LOCATION_ID" ]]; then
    skip "Set XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID to run create/show/delete tests"
elif [[ -z "$EVENT_TYPE_ID" ]]; then
    skip "Missing project transport event type ID; skipping create/show/delete tests"
else
    test_name "Create project transport location event type with required fields"
    xbe_json do project-transport-location-event-types create \
        --project-transport-location "$LOCATION_ID" \
        --project-transport-event-type "$EVENT_TYPE_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "project-transport-location-event-types" "$CREATED_LINK_ID"
            pass
        else
            fail "Created project transport location event type but no ID returned"
        fi
    else
        fail "Failed to create project transport location event type"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Show project transport location event type"
    xbe_json view project-transport-location-event-types show "$CREATED_LINK_ID"
    assert_success
else
    skip "No project transport location event type created; skipping show"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport location event types"
xbe_json view project-transport-location-event-types list --limit 5
assert_success

test_name "List project transport location event types returns array"
xbe_json view project-transport-location-event-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport location event types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

if [[ -n "$LOCATION_ID" ]]; then
    test_name "List project transport location event types with --project-transport-location filter"
    xbe_json view project-transport-location-event-types list --project-transport-location "$LOCATION_ID" --limit 10
    assert_success
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID to test location filter"
fi

if [[ -n "$EVENT_TYPE_ID" ]]; then
    test_name "List project transport location event types with --project-transport-event-type filter"
    xbe_json view project-transport-location-event-types list --project-transport-event-type "$EVENT_TYPE_ID" --limit 10
    assert_success
else
    skip "Missing project transport event type ID; skipping event type filter"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Delete project transport location event type requires --confirm flag"
    xbe_run do project-transport-location-event-types delete "$CREATED_LINK_ID"
    assert_failure

    test_name "Delete project transport location event type with --confirm"
    xbe_run do project-transport-location-event-types delete "$CREATED_LINK_ID" --confirm
    assert_success
else
    skip "No project transport location event type created; skipping delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
