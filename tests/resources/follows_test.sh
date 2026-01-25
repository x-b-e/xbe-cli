#!/bin/bash
#
# XBE CLI Integration Tests: Follows
#
# Tests list/show/create/update/delete operations for the follows resource.
#
# COVERAGE: List filters + create/update relationships + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FOLLOW_ID=""
CREATED_FOLLOW_NEW=0
FOLLOWER_USER_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_ID_2=""

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: follows"

# ============================================================================
# Resolve current user
# ============================================================================

test_name "Get current user for follow tests"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    FOLLOWER_USER_ID=$(json_get ".id")
    if [[ -n "$FOLLOWER_USER_ID" && "$FOLLOWER_USER_ID" != "null" ]]; then
        pass
    else
        fail "Whoami returned no user ID"
    fi
else
    if [[ -n "$XBE_TEST_USER_ID" ]]; then
        FOLLOWER_USER_ID="$XBE_TEST_USER_ID"
        pass
    else
        skip "Unable to resolve current user"
    fi
fi

# ============================================================================
# Prerequisites - Create broker, developer, and projects
# ============================================================================

test_name "Create prerequisite broker for follow tests"
BROKER_NAME=$(unique_name "FollowTestBroker")

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

test_name "Create prerequisite developer for follow tests"
DEVELOPER_NAME=$(unique_name "FollowTestDev")

xbe_json do developers create \
    --name "$DEVELOPER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

test_name "Create prerequisite project for follow tests"
PROJECT_NAME=$(unique_name "FollowTestProject")

xbe_json do projects create \
    --name "$PROJECT_NAME" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        echo "Cannot continue without a project"
        run_tests
    fi
else
    fail "Failed to create project"
    echo "Cannot continue without a project"
    run_tests
fi

test_name "Create second project for follow update tests"
PROJECT_NAME_2=$(unique_name "FollowTestProjectB")

xbe_json do projects create \
    --name "$PROJECT_NAME_2" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID_2" && "$CREATED_PROJECT_ID_2" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID_2"
        pass
    else
        fail "Created project but no ID returned"
        echo "Update tests may be skipped"
    fi
else
    fail "Failed to create second project"
    echo "Update tests may be skipped"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List follows"
xbe_json view follows list --limit 5
assert_success

test_name "List follows returns array"
xbe_json view follows list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list follows"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List follows with --follower"
if [[ -n "$FOLLOWER_USER_ID" && "$FOLLOWER_USER_ID" != "null" ]]; then
    xbe_json view follows list --follower "$FOLLOWER_USER_ID" --limit 5
    assert_success
else
    skip "No follower user ID available"
fi

test_name "List follows with --creator-type"
xbe_json view follows list --creator-type projects --limit 5
assert_success

test_name "List follows with --creator"
if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
    xbe_json view follows list --creator "Project|$CREATED_PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List follows with --creator-type and --creator-id"
if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
    xbe_json view follows list --creator-type projects --creator-id "$CREATED_PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List follows with --created-at-min"
xbe_json view follows list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List follows with --created-at-max"
xbe_json view follows list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List follows with --is-created-at=true"
xbe_json view follows list --is-created-at true --limit 5
assert_success

test_name "List follows with --is-created-at=false"
xbe_json view follows list --is-created-at false --limit 5
assert_success

test_name "List follows with --updated-at-min"
xbe_json view follows list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List follows with --updated-at-max"
xbe_json view follows list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List follows with --is-updated-at=true"
xbe_json view follows list --is-updated-at true --limit 5
assert_success

test_name "List follows with --is-updated-at=false"
xbe_json view follows list --is-updated-at false --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create follow with required fields"
if [[ -n "$FOLLOWER_USER_ID" && "$FOLLOWER_USER_ID" != "null" && -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
    xbe_json do follows create \
        --follower "$FOLLOWER_USER_ID" \
        --creator-type projects \
        --creator-id "$CREATED_PROJECT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_FOLLOW_ID=$(json_get ".id")
        if [[ -n "$CREATED_FOLLOW_ID" && "$CREATED_FOLLOW_ID" != "null" ]]; then
            CREATED_FOLLOW_NEW=1
            register_cleanup "follows" "$CREATED_FOLLOW_ID"
            pass
        else
            fail "Created follow but no ID returned"
        fi
    else
        fail "Failed to create follow"
    fi
else
    skip "Missing follower user ID or project ID"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show follow"
if [[ -n "$CREATED_FOLLOW_ID" && "$CREATED_FOLLOW_ID" != "null" ]]; then
    xbe_json view follows show "$CREATED_FOLLOW_ID"
    assert_success
else
    skip "No follow ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update follow creator"
if [[ -n "$CREATED_FOLLOW_ID" && "$CREATED_FOLLOW_ID" != "null" && -n "$CREATED_PROJECT_ID_2" && "$CREATED_PROJECT_ID_2" != "null" ]]; then
    xbe_json do follows update "$CREATED_FOLLOW_ID" --creator-type projects --creator-id "$CREATED_PROJECT_ID_2"
    assert_success
else
    skip "No follow ID or second project ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete follow"
if [[ "$CREATED_FOLLOW_NEW" -eq 1 && -n "$CREATED_FOLLOW_ID" && "$CREATED_FOLLOW_ID" != "null" ]]; then
    xbe_run do follows delete "$CREATED_FOLLOW_ID" --confirm
    assert_success
else
    skip "Follow was not created by test run"
fi

run_tests
