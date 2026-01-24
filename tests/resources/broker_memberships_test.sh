#!/bin/bash
#
# XBE CLI Integration Tests: Broker Memberships
#
# Tests CRUD operations for the broker-memberships resource.
# Broker memberships define relationships between users and brokers.
#
# NOTE: This test requires creating prerequisite resources: broker, user, project office
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEMBERSHIP_ID=""
CREATED_BROKER_ID=""
CREATED_USER_ID=""
CREATED_PROJECT_OFFICE_ID=""
EXTERNAL_EMPLOYEE_ID=""


describe "Resource: broker-memberships"

# ============================================================================
# Prerequisites - Create broker, project office, and user
# ============================================================================

test_name "Create prerequisite broker for broker membership tests"
BROKER_NAME=$(unique_name "BrokerMembershipBroker")

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

test_name "Create project office for broker membership tests"
PROJECT_OFFICE_NAME=$(unique_name "BrokerMembershipOffice")
PROJECT_OFFICE_ABBR="BMO${RANDOM}"

xbe_json do project-offices create --name "$PROJECT_OFFICE_NAME" --abbreviation "$PROJECT_OFFICE_ABBR" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_OFFICE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
        register_cleanup "project-offices" "$CREATED_PROJECT_OFFICE_ID"
        pass
    else
        fail "Created project office but no ID returned"
        echo "Cannot continue without a project office"
        run_tests
    fi
else
    fail "Failed to create project office"
    echo "Cannot continue without a project office"
    run_tests
fi

test_name "Create prerequisite user for broker membership tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Broker Membership Test User" \
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker membership with required fields"

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
    fi
else
    fail "Failed to create broker membership"
fi

# Only continue if we successfully created a membership
if [[ -z "$CREATED_MEMBERSHIP_ID" || "$CREATED_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker membership ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

EXTERNAL_EMPLOYEE_ID="EXT-$(date +%s)-$RANDOM"

test_name "Update broker membership core fields"
xe_broker_memberships_core_update=(
    do broker-memberships update "$CREATED_MEMBERSHIP_ID"
    --kind "manager"
    --is-admin "true"
    --title "Broker Manager"
    --color-hex "#112233"
    --external-employee-id "$EXTERNAL_EMPLOYEE_ID"
    --explicit-sort-order "10"
    --start-at "2024-01-01T00:00:00Z"
    --end-at "2024-12-31T00:00:00Z"
    --drives-shift-type "day"
    --project-office "$CREATED_PROJECT_OFFICE_ID"
)

xbe_json "${xe_broker_memberships_core_update[@]}"
assert_success

test_name "Update broker membership permissions"
xbe_json do broker-memberships update "$CREATED_MEMBERSHIP_ID" \
    --can-see-rates-as-manager "false" \
    --can-validate-profit-improvements "true" \
    --is-rate-editor "true" \
    --is-time-card-auditor "true" \
    --is-equipment-rental-team-member "true" \
    --is-geofence-violation-team-member "true"
assert_success

test_name "Update broker membership notifications"
xbe_json do broker-memberships update "$CREATED_MEMBERSHIP_ID" \
    --is-unapproved-time-card-subscriber "false" \
    --is-default-job-production-plan-subscriber "true" \
    --enable-recap-notifications "true" \
    --enable-inventory-capacity-notifications "true"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker memberships"
xbe_json view broker-memberships list --limit 5
assert_success

test_name "List broker memberships returns array"
xbe_json view broker-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker memberships"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List broker memberships with --broker filter"
xbe_json view broker-memberships list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List broker memberships with --organization filter"
xbe_json view broker-memberships list --organization "Broker|$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List broker memberships with --project-office filter"
xbe_json view broker-memberships list --project-office "$CREATED_PROJECT_OFFICE_ID" --limit 10
assert_success

test_name "List broker memberships with --user filter"
xbe_json view broker-memberships list --user "$CREATED_USER_ID" --limit 10
assert_success

test_name "List broker memberships with --kind filter"
xbe_json view broker-memberships list --kind "manager" --limit 10
assert_success

test_name "List broker memberships with --q filter"
xbe_json view broker-memberships list --q "Broker Membership Test User" --limit 10
assert_success

test_name "List broker memberships with --drives-shift-type filter"
xbe_json view broker-memberships list --drives-shift-type "day" --limit 10
assert_success

test_name "List broker memberships with --external-employee-id filter"
xbe_json view broker-memberships list --external-employee-id "$EXTERNAL_EMPLOYEE_ID" --limit 10
assert_success

test_name "List broker memberships with --is-rate-editor filter"
xbe_json view broker-memberships list --is-rate-editor "true" --limit 10
assert_success

test_name "List broker memberships with --is-time-card-auditor filter"
xbe_json view broker-memberships list --is-time-card-auditor "true" --limit 10
assert_success

test_name "List broker memberships with --is-equipment-rental-team-member filter"
xbe_json view broker-memberships list --is-equipment-rental-team-member "true" --limit 10
assert_success

test_name "List broker memberships with --is-geofence-violation-team-member filter"
xbe_json view broker-memberships list --is-geofence-violation-team-member "true" --limit 10
assert_success

test_name "List broker memberships with --is-unapproved-time-card-subscriber filter"
xbe_json view broker-memberships list --is-unapproved-time-card-subscriber "false" --limit 10
assert_success

test_name "List broker memberships with --is-default-job-production-plan-subscriber filter"
xbe_json view broker-memberships list --is-default-job-production-plan-subscriber "true" --limit 10
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker membership"
xbe_json view broker-memberships show "$CREATED_MEMBERSHIP_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker membership requires --confirm flag"
xbe_run do broker-memberships delete "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Delete broker membership with --confirm"
# Create a membership specifically for deletion
TEST_DEL_EMAIL=$(unique_email)
xe_broker_memberships_delete_user=(
    do users create
    --name "Broker Membership Delete User"
    --email "$TEST_DEL_EMAIL"
)

xbe_json "${xe_broker_memberships_delete_user[@]}"
if [[ $status -eq 0 ]]; then
    DEL_USER_ID=$(json_get ".id")
    xbe_json do broker-memberships create \
        --user "$DEL_USER_ID" \
        --broker "$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do broker-memberships delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create broker membership for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create broker membership without user fails"
xbe_json do broker-memberships create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create broker membership without broker fails"
xbe_json do broker-memberships create --user "$CREATED_USER_ID"
assert_failure

test_name "Update broker membership without any fields fails"
xbe_json do broker-memberships update "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Update broker membership with can-see-rates-as-driver fails"
xbe_json do broker-memberships update "$CREATED_MEMBERSHIP_ID" --can-see-rates-as-driver "true"
assert_failure

test_name "Update broker membership with unapproved time card subscriber true fails"
xbe_json do broker-memberships update "$CREATED_MEMBERSHIP_ID" --is-unapproved-time-card-subscriber "true"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
