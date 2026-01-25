#!/bin/bash
#
# XBE CLI Integration Tests: Open Door Team Memberships
#
# Tests CRUD operations for the open-door-team-memberships resource.
# Open door team memberships link memberships to organizations for Open Door issue access.
#
# COVERAGE: create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_ODTM_ID=""

ORG_TYPE="brokers"

describe "Resource: open-door-team-memberships"

# ============================================================================
# Prerequisites - Create broker, user, membership
# ============================================================================

test_name "Create prerequisite broker for open door team membership tests"
BROKER_NAME=$(unique_name "OpenDoorTeamMembershipBroker")

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

test_name "Create prerequisite user for open door team membership tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Open Door Team Membership User" \
    --email "$TEST_EMAIL"

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

test_name "Create broker membership for open door team membership"
xbe_json do broker-memberships create \
    --user "$CREATED_USER_ID" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "broker-memberships" "$CREATED_MEMBERSHIP_ID"
        pass
    else
        fail "Created broker membership but no ID returned"
        echo "Cannot continue without a membership"
        run_tests
    fi
else
    fail "Failed to create broker membership"
    echo "Cannot continue without a membership"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create open door team membership"
xbe_json do open-door-team-memberships create \
    --membership "$CREATED_MEMBERSHIP_ID" \
    --organization-type "$ORG_TYPE" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ODTM_ID=$(json_get ".id")
    if [[ -n "$CREATED_ODTM_ID" && "$CREATED_ODTM_ID" != "null" ]]; then
        register_cleanup "open-door-team-memberships" "$CREATED_ODTM_ID"
        pass
    else
        fail "Created open door team membership but no ID returned"
    fi
else
    fail "Failed to create open door team membership"
fi

# Only continue if we successfully created an open door team membership
if [[ -z "$CREATED_ODTM_ID" || "$CREATED_ODTM_ID" == "null" ]]; then
    echo "Cannot continue without a valid open door team membership ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update open door team membership relationships"
xbe_json do open-door-team-memberships update "$CREATED_ODTM_ID" \
    --membership "$CREATED_MEMBERSHIP_ID" \
    --organization-type "$ORG_TYPE" \
    --organization-id "$CREATED_BROKER_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show open door team membership"
xbe_json view open-door-team-memberships show "$CREATED_ODTM_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List open door team memberships"
xbe_json view open-door-team-memberships list --limit 5
assert_success

test_name "List open door team memberships returns array"
xbe_json view open-door-team-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list open door team memberships"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List open door team memberships with --organization filter"
xbe_json view open-door-team-memberships list --organization "Broker|$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List open door team memberships with --membership filter"
xbe_json view open-door-team-memberships list --membership "$CREATED_MEMBERSHIP_ID" --limit 10
assert_success

test_name "List open door team memberships with --created-at-min filter"
xbe_json view open-door-team-memberships list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List open door team memberships with --created-at-max filter"
xbe_json view open-door-team-memberships list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List open door team memberships with --updated-at-min filter"
xbe_json view open-door-team-memberships list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List open door team memberships with --updated-at-max filter"
xbe_json view open-door-team-memberships list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List open door team memberships with --is-created-at filter"
xbe_json view open-door-team-memberships list --is-created-at true --limit 5
assert_success

test_name "List open door team memberships with --is-updated-at filter"
xbe_json view open-door-team-memberships list --is-updated-at true --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete open door team membership"
xbe_run do open-door-team-memberships delete "$CREATED_ODTM_ID" --confirm
assert_success

run_tests
