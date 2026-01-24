#!/bin/bash
#
# XBE CLI Integration Tests: Action Item Trackers
#
# Tests list/show and create/update/delete behavior for action-item-trackers.
#
# COVERAGE: List + show + create/update attributes + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: action-item-trackers"

SAMPLE_ID=""
SAMPLE_ACTION_ITEM_ID=""

CREATED_TRACKER_ID=""
TRACKER_ID=""
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
ACTION_ITEM_ID="${XBE_TEST_ACTION_ITEM_ID:-}"
USER_ID="${XBE_TEST_USER_ID:-}"
SKIP_MUTATION=0

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping create/update/delete tests)"
    SKIP_MUTATION=1
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List action item trackers"
xbe_json view action-item-trackers list --limit 5
assert_success

test_name "List action item trackers returns array"
xbe_json view action-item-trackers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list action item trackers"
fi

test_name "Capture sample action item tracker"
xbe_json view action-item-trackers list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ACTION_ITEM_ID=$(json_get ".[0].action_item_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No action item trackers available for follow-on tests"
    fi
else
    skip "Could not list action item trackers to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show action item tracker"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view action-item-trackers show "$SAMPLE_ID"
    assert_success
else
    skip "No action item tracker ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create action item tracker without required fields fails"
xbe_json do action-item-trackers create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete tests without XBE_TOKEN"
else
    if [[ -z "$USER_ID" ]]; then
        xbe_json view users list --limit 1
        if [[ $status -eq 0 ]]; then
            USER_ID=$(json_get ".[0].id")
        fi
    fi

    if [[ -z "$ACTION_ITEM_ID" ]]; then
        if [[ -z "$BROKER_ID" ]]; then
            test_name "Create prerequisite broker for action item tracker tests"
            BROKER_NAME=$(unique_name "AITTrackerBroker")

            xbe_json do brokers create --name "$BROKER_NAME"
            if [[ $status -eq 0 ]]; then
                BROKER_ID=$(json_get ".id")
                if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
                    register_cleanup "brokers" "$BROKER_ID"
                    pass
                else
                    fail "Created broker but no ID returned"
                fi
            else
                fail "Failed to create broker"
            fi
        fi

        if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
            test_name "Create action item for tracker tests"
            TEST_TITLE=$(unique_name "ActionItemTracker")
            xbe_json do action-items create \
                --title "$TEST_TITLE" \
                --responsible-organization "Broker|$BROKER_ID"

            if [[ $status -eq 0 ]]; then
                ACTION_ITEM_ID=$(json_get ".id")
                if [[ -n "$ACTION_ITEM_ID" && "$ACTION_ITEM_ID" != "null" ]]; then
                    register_cleanup "action-items" "$ACTION_ITEM_ID"
                    pass
                else
                    fail "Created action item but no ID returned"
                fi
            else
                fail "Failed to create action item"
            fi
        fi
    fi

    if [[ -n "$ACTION_ITEM_ID" && "$ACTION_ITEM_ID" != "null" ]]; then
        test_name "Create action item tracker"
        create_args=(--action-item "$ACTION_ITEM_ID" --status ready_for_work --dev-effort-size m --dev-effort-minutes 30 --has-due-date-agreement true --is-unplanned true --priority-position 1)
        if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
            create_args+=(--dev-assignee "$USER_ID" --customer-success-assignee "$USER_ID")
        fi

        xbe_json do action-item-trackers create "${create_args[@]}"

        if [[ $status -eq 0 ]]; then
            CREATED_TRACKER_ID=$(json_get ".id")
            if [[ -n "$CREATED_TRACKER_ID" && "$CREATED_TRACKER_ID" != "null" ]]; then
                TRACKER_ID="$CREATED_TRACKER_ID"
                register_cleanup "action-item-trackers" "$CREATED_TRACKER_ID"
                pass
            else
                fail "Created action item tracker but no ID returned"
            fi
        else
            xbe_json view action-items show "$ACTION_ITEM_ID"
            if [[ $status -eq 0 ]]; then
                TRACKER_ID=$(json_get ".tracker_id")
                if [[ -n "$TRACKER_ID" && "$TRACKER_ID" != "null" ]]; then
                    skip "Tracker already exists for action item"
                else
                    fail "Failed to create action item tracker"
                fi
            else
                fail "Failed to create action item tracker"
            fi
        fi
    else
        skip "No action item available for tracker creation"
    fi
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$TRACKER_ID" && "$TRACKER_ID" != "null" ]]; then
    test_name "Update action item tracker with no fields fails"
    xbe_json do action-item-trackers update "$TRACKER_ID"
    assert_failure

    test_name "Update action item tracker status"
    xbe_json do action-item-trackers update "$TRACKER_ID" --status in_development
    assert_success

    test_name "Update action item tracker dev effort size"
    xbe_json do action-item-trackers update "$TRACKER_ID" --dev-effort-size l
    assert_success

    test_name "Update action item tracker dev effort minutes"
    xbe_json do action-item-trackers update "$TRACKER_ID" --dev-effort-minutes 45
    assert_success

    test_name "Update action item tracker due date agreement"
    xbe_json do action-item-trackers update "$TRACKER_ID" --has-due-date-agreement false
    assert_success

    test_name "Update action item tracker unplanned status"
    xbe_json do action-item-trackers update "$TRACKER_ID" --is-unplanned false
    assert_success

    test_name "Update action item tracker priority position"
    xbe_json do action-item-trackers update "$TRACKER_ID" --priority-position 2
    assert_success

    test_name "Update action item tracker dev assignee"
    if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
        xbe_json do action-item-trackers update "$TRACKER_ID" --dev-assignee "$USER_ID"
        assert_success
    else
        skip "No user ID available"
    fi

    test_name "Update action item tracker customer success assignee"
    if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
        xbe_json do action-item-trackers update "$TRACKER_ID" --customer-success-assignee "$USER_ID"
        assert_success
    else
        skip "No user ID available"
    fi
else
    skip "No action item tracker available for update tests"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_TRACKER_ID" && "$CREATED_TRACKER_ID" != "null" ]]; then
    test_name "Delete action item tracker"
    xbe_json do action-item-trackers delete "$CREATED_TRACKER_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
