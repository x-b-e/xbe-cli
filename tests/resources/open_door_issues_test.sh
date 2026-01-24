#!/bin/bash
#
# XBE CLI Integration Tests: Open Door Issues
#
# Tests CRUD operations for the open-door-issues resource.
# Open door issues capture concerns reported for organizations.
#
# NOTE: This test requires creating prerequisite resources: brokers and users
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ISSUE_ID=""
CREATED_BROKER_ID=""
CREATED_BROKER_ID_2=""
REPORTER_USER_ID=""
ALT_REPORTER_USER_ID=""

describe "Resource: open-door-issues"

# ============================================================================
# Prerequisites - Brokers and users
# ============================================================================

test_name "Create broker for open door issue tests"
BROKER_NAME=$(unique_name "OpenDoorBroker")

xbe_json do brokers create --name "$BROKER_NAME" --is-accepting-open-door-issues true

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
        xbe_run do brokers update "$CREATED_BROKER_ID" --is-accepting-open-door-issues true
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create secondary broker for organization update tests"
BROKER_NAME_2=$(unique_name "OpenDoorBrokerAlt")

xbe_json do brokers create --name "$BROKER_NAME_2" --is-accepting-open-door-issues true

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID_2" && "$CREATED_BROKER_ID_2" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID_2"
        pass
    else
        fail "Created secondary broker but no ID returned"
    fi
else
    skip "Could not create secondary broker"
fi

test_name "Create reporter user for open door issue tests"
REPORTER_EMAIL=$(unique_email)
REPORTER_NAME=$(unique_name "OpenDoorReporter")

xbe_json do users create \
    --email "$REPORTER_EMAIL" \
    --name "$REPORTER_NAME"

if [[ $status -eq 0 ]]; then
    REPORTER_USER_ID=$(json_get ".id")
    if [[ -n "$REPORTER_USER_ID" && "$REPORTER_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a reporter user"
        run_tests
    fi
else
    fail "Failed to create reporter user"
    echo "Cannot continue without a reporter user"
    run_tests
fi

test_name "Create alternate reporter user for updates"
ALT_REPORTER_EMAIL=$(unique_email)
ALT_REPORTER_NAME=$(unique_name "OpenDoorReporterAlt")

xbe_json do users create \
    --email "$ALT_REPORTER_EMAIL" \
    --name "$ALT_REPORTER_NAME"

if [[ $status -eq 0 ]]; then
    ALT_REPORTER_USER_ID=$(json_get ".id")
    if [[ -n "$ALT_REPORTER_USER_ID" && "$ALT_REPORTER_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without alternate reporter user"
        run_tests
    fi
else
    fail "Failed to create alternate reporter user"
    echo "Cannot continue without alternate reporter user"
    run_tests
fi

test_name "Create membership for reporter user to broker"
xbe_json do memberships create \
    --user "$REPORTER_USER_ID" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    REPORTER_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$REPORTER_MEMBERSHIP_ID" && "$REPORTER_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "memberships" "$REPORTER_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
        echo "Cannot continue without a reporter membership"
        run_tests
    fi
else
    fail "Failed to create reporter membership"
    echo "Cannot continue without a reporter membership"
    run_tests
fi

test_name "Create membership for alternate reporter to broker"
xbe_json do memberships create \
    --user "$ALT_REPORTER_USER_ID" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    ALT_REPORTER_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$ALT_REPORTER_MEMBERSHIP_ID" && "$ALT_REPORTER_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "memberships" "$ALT_REPORTER_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
        echo "Cannot continue without an alternate reporter membership"
        run_tests
    fi
else
    fail "Failed to create alternate reporter membership"
    echo "Cannot continue without an alternate reporter membership"
    run_tests
fi

if [[ -n "$CREATED_BROKER_ID_2" && "$CREATED_BROKER_ID_2" != "null" ]]; then
    test_name "Create membership for alternate reporter to secondary broker"
    xbe_json do memberships create \
        --user "$ALT_REPORTER_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID_2"

    if [[ $status -eq 0 ]]; then
        ALT_REPORTER_MEMBERSHIP_ID_2=$(json_get ".id")
        if [[ -n "$ALT_REPORTER_MEMBERSHIP_ID_2" && "$ALT_REPORTER_MEMBERSHIP_ID_2" != "null" ]]; then
            register_cleanup "memberships" "$ALT_REPORTER_MEMBERSHIP_ID_2"
            pass
        else
            fail "Created membership but no ID returned"
        fi
    else
        fail "Failed to create alternate reporter membership for secondary broker"
    fi
else
    skip "Skipping alternate reporter membership for secondary broker"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create open door issue with required fields"

xbe_json do open-door-issues create \
    --description "Safety concern at job site" \
    --status editing \
    --organization "Broker|$CREATED_BROKER_ID" \
    --reported-by "$REPORTER_USER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ISSUE_ID=$(json_get ".id")
    if [[ -n "$CREATED_ISSUE_ID" && "$CREATED_ISSUE_ID" != "null" ]]; then
        register_cleanup "open-door-issues" "$CREATED_ISSUE_ID"
        pass
    else
        fail "Created open door issue but no ID returned"
    fi
else
    fail "Failed to create open door issue"
fi

if [[ -z "$CREATED_ISSUE_ID" || "$CREATED_ISSUE_ID" == "null" ]]; then
    echo "Cannot continue without a valid open door issue ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update open door issue --description"
xbe_json do open-door-issues update "$CREATED_ISSUE_ID" --description "Updated issue description"
assert_success

test_name "Update open door issue --status"
xbe_json do open-door-issues update "$CREATED_ISSUE_ID" --status resolved
assert_success

test_name "Update open door issue --reported-by"
xbe_json do open-door-issues update "$CREATED_ISSUE_ID" --reported-by "$ALT_REPORTER_USER_ID"
assert_success

if [[ -n "$CREATED_BROKER_ID_2" && "$CREATED_BROKER_ID_2" != "null" ]]; then
    test_name "Update open door issue --organization"
    xbe_json do open-door-issues update "$CREATED_ISSUE_ID" --organization "Broker|$CREATED_BROKER_ID_2"
    assert_success
else
    skip "Skipping organization update test (no secondary broker)"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List open door issues"
xbe_json view open-door-issues list --limit 5
assert_success

test_name "List open door issues returns array"
xbe_json view open-door-issues list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list open door issues"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List open door issues with --organization filter"
xbe_json view open-door-issues list --organization "Broker|$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List open door issues with --organization-type and --organization-id filters"
xbe_json view open-door-issues list --organization-type Broker --organization-id "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List open door issues with --not-organization-type filter"
xbe_json view open-door-issues list --not-organization-type Customer --limit 10
assert_success

test_name "List open door issues with --created-at-min filter"
xbe_json view open-door-issues list --created-at-min "2000-01-01T00:00:00Z" --limit 10
assert_success

test_name "List open door issues with --created-at-max filter"
xbe_json view open-door-issues list --created-at-max "2099-01-01T00:00:00Z" --limit 10
assert_success

test_name "List open door issues with --updated-at-min filter"
xbe_json view open-door-issues list --updated-at-min "2000-01-01T00:00:00Z" --limit 10
assert_success

test_name "List open door issues with --updated-at-max filter"
xbe_json view open-door-issues list --updated-at-max "2099-01-01T00:00:00Z" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination and Sorting
# ============================================================================

test_name "List open door issues with --limit"
xbe_json view open-door-issues list --limit 3
assert_success

test_name "List open door issues with --offset"
xbe_json view open-door-issues list --limit 3 --offset 3
assert_success

test_name "List open door issues with --sort"
xbe_json view open-door-issues list --sort created-at --limit 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete open door issue requires --confirm flag"
xbe_run do open-door-issues delete "$CREATED_ISSUE_ID"
assert_failure

test_name "Delete open door issue with --confirm"
# Create an issue specifically for deletion
xbe_json do open-door-issues create \
    --description "Issue to delete" \
    --status editing \
    --organization "Broker|$CREATED_BROKER_ID" \
    --reported-by "$REPORTER_USER_ID"

if [[ $status -eq 0 ]]; then
    DEL_ISSUE_ID=$(json_get ".id")
    if [[ -n "$DEL_ISSUE_ID" && "$DEL_ISSUE_ID" != "null" ]]; then
        xbe_run do open-door-issues delete "$DEL_ISSUE_ID" --confirm
        assert_success
    else
        skip "Could not create open door issue for deletion test"
    fi
else
    skip "Could not create open door issue for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create open door issue without --description fails"
xbe_json do open-door-issues create \
    --status editing \
    --organization "Broker|$CREATED_BROKER_ID" \
    --reported-by "$REPORTER_USER_ID"
assert_failure

test_name "Create open door issue without --status fails"
xbe_json do open-door-issues create \
    --description "Missing status" \
    --organization "Broker|$CREATED_BROKER_ID" \
    --reported-by "$REPORTER_USER_ID"
assert_failure

test_name "Create open door issue without --organization fails"
xbe_json do open-door-issues create \
    --description "Missing organization" \
    --status editing \
    --reported-by "$REPORTER_USER_ID"
assert_failure

test_name "Create open door issue without --reported-by fails"
xbe_json do open-door-issues create \
    --description "Missing reporter" \
    --status editing \
    --organization "Broker|$CREATED_BROKER_ID"
assert_failure

test_name "Update open door issue without any fields fails"
xbe_run do open-door-issues update "$CREATED_ISSUE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
