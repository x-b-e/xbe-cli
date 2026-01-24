#!/bin/bash
#
# XBE CLI Integration Tests: Action Item Tracker Update Requests
#
# Tests CRUD operations for the action_item_tracker_update_requests resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REQUEST_ID=""
CREATED_BROKER_ID=""
CREATED_ACTION_ITEM_ID=""
ACTION_ITEM_TRACKER_ID=""
WHOAMI_USER_ID=""

REQUESTED_BY_ID=""
REQUESTED_FROM_ID=""

REQUEST_NOTE="Please provide an update on this action item."
UPDATE_NOTE="Initial update: work started."

DESIRED_DUE_ON="2026-02-01"
UPDATED_DUE_ON="2026-03-01"

DESCRIBE_RESOURCE="action-item-tracker-update-requests"

describe "Resource: ${DESCRIBE_RESOURCE}"

# ==========================================================================
# Resolve current user
# ==========================================================================

test_name "Get current user for update request tests"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
        pass
    else
        fail "Whoami returned no user ID"
    fi
else
    if [[ -n "$XBE_TEST_USER_ID" ]]; then
        WHOAMI_USER_ID="$XBE_TEST_USER_ID"
        pass
    else
        skip "Unable to resolve current user"
    fi
fi

REQUESTED_BY_ID="${XBE_TEST_USER_ID:-$WHOAMI_USER_ID}"
REQUESTED_FROM_ID="${XBE_TEST_USER_ID:-$WHOAMI_USER_ID}"

# ==========================================================================
# Prerequisites - Create broker and action item
# ==========================================================================

test_name "Create prerequisite broker for update request tests"
BROKER_NAME=$(unique_name "AITURBroker")

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
        skip "Failed to create broker and XBE_TEST_BROKER_ID not set"
    fi
fi

# Create action item (used to get tracker ID)
test_name "Create action item for update request tests"
ACTION_ITEM_TITLE=$(unique_name "AITURActionItem")

if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json do action-items create \
        --title "$ACTION_ITEM_TITLE" \
        --responsible-organization "Broker|$CREATED_BROKER_ID"
else
    xbe_json do action-items create --title "$ACTION_ITEM_TITLE"
fi

if [[ $status -eq 0 ]]; then
    CREATED_ACTION_ITEM_ID=$(json_get ".id")
    if [[ -n "$CREATED_ACTION_ITEM_ID" && "$CREATED_ACTION_ITEM_ID" != "null" ]]; then
        register_cleanup "action-items" "$CREATED_ACTION_ITEM_ID"
        pass
    else
        fail "Created action item but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_ACTION_ITEM_ID" ]]; then
        CREATED_ACTION_ITEM_ID="$XBE_TEST_ACTION_ITEM_ID"
        echo "    Using XBE_TEST_ACTION_ITEM_ID: $CREATED_ACTION_ITEM_ID"
        pass
    else
        fail "Failed to create action item"
    fi
fi

if [[ -z "$CREATED_ACTION_ITEM_ID" || "$CREATED_ACTION_ITEM_ID" == "null" ]]; then
    echo "Cannot continue without a valid action item ID"
    run_tests
fi

# Resolve action item tracker ID (async creation)
test_name "Resolve action item tracker ID"
ACTION_ITEM_TRACKER_ID="${XBE_TEST_ACTION_ITEM_TRACKER_ID:-}"
ACTION_ITEM_ID_FOR_TRACKER="${CREATED_ACTION_ITEM_ID:-$XBE_TEST_ACTION_ITEM_ID}"

if [[ -z "$ACTION_ITEM_TRACKER_ID" && -n "$ACTION_ITEM_ID_FOR_TRACKER" && "$ACTION_ITEM_ID_FOR_TRACKER" != "null" ]]; then
    for attempt in {1..6}; do
        xbe_json view action-items show "$ACTION_ITEM_ID_FOR_TRACKER"
        if [[ $status -eq 0 ]]; then
            ACTION_ITEM_TRACKER_ID=$(json_get ".tracker_id")
            if [[ -n "$ACTION_ITEM_TRACKER_ID" && "$ACTION_ITEM_TRACKER_ID" != "null" ]]; then
                break
            fi
        fi
        sleep 2
    done
fi

if [[ -n "$ACTION_ITEM_TRACKER_ID" && "$ACTION_ITEM_TRACKER_ID" != "null" ]]; then
    pass
else
    fail "Unable to resolve action item tracker ID"
    run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create update request with attributes"
if [[ -n "$REQUESTED_BY_ID" && "$REQUESTED_BY_ID" != "null" && -n "$REQUESTED_FROM_ID" && "$REQUESTED_FROM_ID" != "null" ]]; then
    xbe_json do action-item-tracker-update-requests create \
        --action-item-tracker "$ACTION_ITEM_TRACKER_ID" \
        --requested-by "$REQUESTED_BY_ID" \
        --requested-from "$REQUESTED_FROM_ID" \
        --request-note "$REQUEST_NOTE" \
        --due-on "$DESIRED_DUE_ON" \
        --update-note "$UPDATE_NOTE"

    if [[ $status -eq 0 ]]; then
        CREATED_REQUEST_ID=$(json_get ".id")
        if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
            register_cleanup "action-item-tracker-update-requests" "$CREATED_REQUEST_ID"
            pass
        else
            fail "Created update request but no ID returned"
        fi
    else
        fail "Failed to create update request"
    fi
else
    skip "No user ID available for requested-by/from (set XBE_TEST_USER_ID)"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show update request"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json view action-item-tracker-update-requests show "$CREATED_REQUEST_ID"
    assert_success
else
    skip "No update request ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update request note"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do action-item-tracker-update-requests update "$CREATED_REQUEST_ID" --request-note "Updated request note"
    assert_success
else
    skip "No update request ID available"
fi

test_name "Update due-on"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do action-item-tracker-update-requests update "$CREATED_REQUEST_ID" --due-on "$UPDATED_DUE_ON"
    assert_success
else
    skip "No update request ID available"
fi

test_name "Update update-note"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do action-item-tracker-update-requests update "$CREATED_REQUEST_ID" --update-note "Follow-up update note"
    assert_success
else
    skip "No update request ID available"
fi

test_name "Update without attributes fails"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_run do action-item-tracker-update-requests update "$CREATED_REQUEST_ID"
    assert_failure
else
    skip "No update request ID available"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List update requests"
xbe_json view action-item-tracker-update-requests list --limit 5
assert_success

test_name "List update requests returns array"
xbe_json view action-item-tracker-update-requests list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list update requests"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List update requests with --created-at-min"
xbe_json view action-item-tracker-update-requests list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List update requests with --created-at-max"
xbe_json view action-item-tracker-update-requests list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List update requests with --is-created-at"
xbe_json view action-item-tracker-update-requests list --is-created-at true --limit 5
assert_success

test_name "List update requests with --updated-at-min"
xbe_json view action-item-tracker-update-requests list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List update requests with --updated-at-max"
xbe_json view action-item-tracker-update-requests list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List update requests with --is-updated-at"
xbe_json view action-item-tracker-update-requests list --is-updated-at true --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete update request requires --confirm flag"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_run do action-item-tracker-update-requests delete "$CREATED_REQUEST_ID"
    assert_failure
else
    skip "No update request ID available"
fi

test_name "Delete update request with --confirm"
if [[ -n "$REQUESTED_BY_ID" && "$REQUESTED_BY_ID" != "null" && -n "$REQUESTED_FROM_ID" && "$REQUESTED_FROM_ID" != "null" ]]; then
    xbe_json do action-item-tracker-update-requests create \
        --action-item-tracker "$ACTION_ITEM_TRACKER_ID" \
        --requested-by "$REQUESTED_BY_ID" \
        --requested-from "$REQUESTED_FROM_ID" \
        --request-note "Delete test request"

    if [[ $status -eq 0 ]]; then
        DELETE_REQUEST_ID=$(json_get ".id")
        if [[ -n "$DELETE_REQUEST_ID" && "$DELETE_REQUEST_ID" != "null" ]]; then
            xbe_run do action-item-tracker-update-requests delete "$DELETE_REQUEST_ID" --confirm
            assert_success
        else
            skip "Could not create update request for deletion"
        fi
    else
        skip "Could not create update request for deletion"
    fi
else
    skip "No user ID available for delete test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create update request without --action-item-tracker fails"
if [[ -n "$REQUESTED_BY_ID" && "$REQUESTED_BY_ID" != "null" && -n "$REQUESTED_FROM_ID" && "$REQUESTED_FROM_ID" != "null" ]]; then
    xbe_run do action-item-tracker-update-requests create \
        --requested-by "$REQUESTED_BY_ID" \
        --requested-from "$REQUESTED_FROM_ID"
    assert_failure
else
    skip "No user ID available"
fi

test_name "Create update request without --requested-by fails"
if [[ -n "$REQUESTED_FROM_ID" && "$REQUESTED_FROM_ID" != "null" ]]; then
    xbe_run do action-item-tracker-update-requests create \
        --action-item-tracker "$ACTION_ITEM_TRACKER_ID" \
        --requested-from "$REQUESTED_FROM_ID"
    assert_failure
else
    skip "No user ID available"
fi

test_name "Create update request without --requested-from fails"
if [[ -n "$REQUESTED_BY_ID" && "$REQUESTED_BY_ID" != "null" ]]; then
    xbe_run do action-item-tracker-update-requests create \
        --action-item-tracker "$ACTION_ITEM_TRACKER_ID" \
        --requested-by "$REQUESTED_BY_ID"
    assert_failure
else
    skip "No user ID available"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
