#!/bin/bash
#
# XBE CLI Integration Tests: Memberships
#
# Tests CRUD operations for the memberships resource.
# Memberships define relationships between users and organizations.
#
# NOTE: This test requires creating prerequisite resources: broker and user
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEMBERSHIP_ID=""
CREATED_BROKER_ID=""
CREATED_USER_ID=""

describe "Resource: memberships"

# ============================================================================
# Prerequisites - Create broker and user
# ============================================================================

test_name "Create prerequisite broker for membership tests"
BROKER_NAME=$(unique_name "MemberTestBroker")

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

test_name "Create prerequisite user for membership tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Membership Test User" \
    --email "$TEST_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        # Note: Users cannot be deleted via API, so we don't register cleanup
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create membership with required fields"

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
    fi
else
    fail "Failed to create membership"
fi

# Only continue if we successfully created a membership
if [[ -z "$CREATED_MEMBERSHIP_ID" || "$CREATED_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid membership ID"
    run_tests
fi

test_name "Create membership with kind=manager"
TEST_EMAIL2=$(unique_email)
xbe_json do users create --name "Manager Test User" --email "$TEST_EMAIL2"
if [[ $status -eq 0 ]]; then
    MANAGER_USER_ID=$(json_get ".id")
    xbe_json do memberships create \
        --user "$MANAGER_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID" \
        --kind "manager"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "memberships" "$id"
        pass
    else
        fail "Failed to create membership with kind=manager"
    fi
else
    skip "Could not create user for manager membership test"
fi

test_name "Create membership with title"
TEST_EMAIL3=$(unique_email)
xbe_json do users create --name "Title Test User" --email "$TEST_EMAIL3"
if [[ $status -eq 0 ]]; then
    TITLE_USER_ID=$(json_get ".id")
    xbe_json do memberships create \
        --user "$TITLE_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID" \
        --title "Test Driver"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "memberships" "$id"
        pass
    else
        fail "Failed to create membership with title"
    fi
else
    skip "Could not create user for title membership test"
fi

test_name "Create membership with is-admin"
TEST_EMAIL4=$(unique_email)
xbe_json do users create --name "Admin Test User" --email "$TEST_EMAIL4"
if [[ $status -eq 0 ]]; then
    ADMIN_USER_ID=$(json_get ".id")
    xbe_json do memberships create \
        --user "$ADMIN_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID" \
        --kind "manager" \
        --is-admin "true"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "memberships" "$id"
        pass
    else
        fail "Failed to create membership with is-admin"
    fi
else
    skip "Could not create user for admin membership test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update membership title"
xbe_json do memberships update "$CREATED_MEMBERSHIP_ID" --title "Updated Title"
assert_success

test_name "Update membership kind"
xbe_json do memberships update "$CREATED_MEMBERSHIP_ID" --kind "manager"
assert_success

test_name "Update membership is-admin"
xbe_json do memberships update "$CREATED_MEMBERSHIP_ID" --is-admin "true"
assert_success

test_name "Update membership color-hex"
xbe_json do memberships update "$CREATED_MEMBERSHIP_ID" --color-hex "#FF5500"
assert_success

test_name "Update membership drives-shift-type"
xbe_json do memberships update "$CREATED_MEMBERSHIP_ID" --drives-shift-type "day"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List memberships"
xbe_json view memberships list --limit 5
assert_success

test_name "List memberships returns array"
xbe_json view memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list memberships"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List memberships with --broker filter"
xbe_json view memberships list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List memberships with --user filter"
xbe_json view memberships list --user "$CREATED_USER_ID" --limit 10
assert_success

test_name "List memberships with --kind filter"
xbe_json view memberships list --kind "manager" --limit 10
assert_success

test_name "List memberships with --is-rate-editor filter"
xbe_json view memberships list --is-rate-editor "false" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List memberships with --limit"
xbe_json view memberships list --limit 3
assert_success

test_name "List memberships with --offset"
xbe_json view memberships list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete membership requires --confirm flag"
xbe_run do memberships delete "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Delete membership with --confirm"
# Create a membership specifically for deletion
TEST_DEL_EMAIL=$(unique_email)
xbe_json do users create --name "Delete Test User" --email "$TEST_DEL_EMAIL"
if [[ $status -eq 0 ]]; then
    DEL_USER_ID=$(json_get ".id")
    xbe_json do memberships create \
        --user "$DEL_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do memberships delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create membership for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create membership without user fails"
xbe_json do memberships create --organization "Broker|$CREATED_BROKER_ID"
assert_failure

test_name "Create membership without organization fails"
xbe_json do memberships create --user "$CREATED_USER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do memberships update "$CREATED_MEMBERSHIP_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
