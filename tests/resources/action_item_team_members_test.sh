#!/bin/bash
#
# XBE CLI Integration Tests: Action Item Team Members
#
# Tests CRUD operations and list filters for the action_item_team_members resource.
#
# COVERAGE: Create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_ACTION_ITEM_ID=""
CREATED_TEAM_MEMBER_ID=""

SAMPLE_ID=""


describe "Resource: action-item-team-members"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List action item team members"
xbe_json view action-item-team-members list --limit 5
assert_success

test_name "List action item team members returns array"
xbe_json view action-item-team-members list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list action item team members"
fi

# ============================================================================
# Prerequisites - Broker, user, membership, action item
# ============================================================================

test_name "Create prerequisite broker for action item team member tests"
BROKER_NAME=$(unique_name "AITeamBroker")

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

test_name "Create prerequisite user for action item team member tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create --name "Action Item Team Member Test User" --email "$TEST_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

test_name "Create prerequisite membership for team member user"
if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" && -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json do memberships create \
        --user "$CREATED_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_MEMBERSHIP_ID=$(json_get ".id")
        if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
            register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
            pass
        else
            fail "Created membership but no ID returned"
            echo "Cannot continue without a membership"
            run_tests
        fi
    else
        fail "Failed to create membership"
        echo "Cannot continue without a membership"
        run_tests
    fi
else
    skip "Missing broker or user ID"
    run_tests
fi

test_name "Create prerequisite action item"
TEST_TITLE=$(unique_name "ActionItemTeam")

xbe_json do action-items create \
    --title "$TEST_TITLE" \
    --responsible-organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ACTION_ITEM_ID=$(json_get ".id")
    if [[ -n "$CREATED_ACTION_ITEM_ID" && "$CREATED_ACTION_ITEM_ID" != "null" ]]; then
        register_cleanup "action-items" "$CREATED_ACTION_ITEM_ID"
        pass
    else
        fail "Created action item but no ID returned"
        echo "Cannot continue without an action item"
        run_tests
    fi
else
    fail "Failed to create action item"
    echo "Cannot continue without an action item"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create action item team member requires --action-item"
xbe_json do action-item-team-members create --user 123
assert_failure

test_name "Create action item team member requires --user"
if [[ -n "$CREATED_ACTION_ITEM_ID" && "$CREATED_ACTION_ITEM_ID" != "null" ]]; then
    xbe_json do action-item-team-members create --action-item "$CREATED_ACTION_ITEM_ID"
    assert_failure
else
    skip "No action item ID available"
fi

test_name "Create action item team member"
if [[ -n "$CREATED_ACTION_ITEM_ID" && "$CREATED_ACTION_ITEM_ID" != "null" && -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
    xbe_json do action-item-team-members create \
        --action-item "$CREATED_ACTION_ITEM_ID" \
        --user "$CREATED_USER_ID" \
        --is-responsible-person

    if [[ $status -eq 0 ]]; then
        CREATED_TEAM_MEMBER_ID=$(json_get ".id")
        if [[ -n "$CREATED_TEAM_MEMBER_ID" && "$CREATED_TEAM_MEMBER_ID" != "null" ]]; then
            register_cleanup "action-item-team-members" "$CREATED_TEAM_MEMBER_ID"
            pass
        else
            fail "Created team member but no ID returned"
        fi
    else
        fail "Failed to create action item team member"
    fi
else
    skip "Missing action item or user ID"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List action item team members with --action-item filter"
if [[ -n "$CREATED_ACTION_ITEM_ID" && "$CREATED_ACTION_ITEM_ID" != "null" ]]; then
    xbe_json view action-item-team-members list --action-item "$CREATED_ACTION_ITEM_ID" --limit 5
    assert_success
else
    skip "No action item ID available"
fi

test_name "List action item team members with --user filter"
if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
    xbe_json view action-item-team-members list --user "$CREATED_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List action item team members with --is-responsible-person filter"
if [[ -n "$CREATED_TEAM_MEMBER_ID" && "$CREATED_TEAM_MEMBER_ID" != "null" ]]; then
    xbe_json view action-item-team-members list --is-responsible-person true --limit 5
    assert_success
else
    skip "No team member ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show action item team member"
if [[ -n "$CREATED_TEAM_MEMBER_ID" && "$CREATED_TEAM_MEMBER_ID" != "null" ]]; then
    xbe_json view action-item-team-members show "$CREATED_TEAM_MEMBER_ID"
    assert_success
else
    skip "No action item team member ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update action item team member responsible flag"
if [[ -n "$CREATED_TEAM_MEMBER_ID" && "$CREATED_TEAM_MEMBER_ID" != "null" ]]; then
    xbe_json do action-item-team-members update "$CREATED_TEAM_MEMBER_ID" --is-responsible-person=false
    assert_success
else
    skip "No action item team member ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete action item team member requires --confirm"
if [[ -n "$CREATED_TEAM_MEMBER_ID" && "$CREATED_TEAM_MEMBER_ID" != "null" ]]; then
    xbe_run do action-item-team-members delete "$CREATED_TEAM_MEMBER_ID"
    assert_failure
else
    skip "No action item team member ID available"
fi

test_name "Delete action item team member"
if [[ -n "$CREATED_TEAM_MEMBER_ID" && "$CREATED_TEAM_MEMBER_ID" != "null" ]]; then
    xbe_run do action-item-team-members delete "$CREATED_TEAM_MEMBER_ID" --confirm
    assert_success
else
    skip "No action item team member ID available"
fi

run_tests
