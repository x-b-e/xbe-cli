#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Memberships
#
# Tests CRUD operations for the trucker-memberships resource.
# Trucker memberships link users to truckers.
#
# NOTE: This test requires creating prerequisite resources: broker, trucker, project office, user
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEMBERSHIP_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_PROJECT_OFFICE_ID=""
CREATED_USER_ID=""

describe "Resource: trucker-memberships"

# ==========================================================================
# Prerequisites - Create broker, trucker, project office, and user
# ==========================================================================

test_name "Create prerequisite broker for trucker membership tests"
BROKER_NAME=$(unique_name "TMBroker")

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

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "TruckerMember")
TRUCKER_ADDRESS="100 Haul Lane, Truckerville, TT 55555"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

test_name "Create prerequisite project office"
PROJECT_OFFICE_NAME=$(unique_name "TMOffice")

xbe_json do project-offices create --name "$PROJECT_OFFICE_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_OFFICE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
        register_cleanup "project-offices" "$CREATED_PROJECT_OFFICE_ID"
        pass
    else
        fail "Created project office but no ID returned"
    fi
else
    xbe_json view project-offices list --broker "$CREATED_BROKER_ID" --limit 1
    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_OFFICE_ID=$(json_get ".[0].id")
        if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
            echo "    Using existing project office: $CREATED_PROJECT_OFFICE_ID"
            pass
        else
            skip "Failed to create or locate project office"
        fi
    else
        skip "Failed to create or locate project office"
    fi
fi

if [[ "$CREATED_PROJECT_OFFICE_ID" == "null" ]]; then
    CREATED_PROJECT_OFFICE_ID=""
fi

test_name "Create prerequisite user"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Trucker Member" \
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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create trucker membership with required fields"

xbe_json do trucker-memberships create \
    --user "$CREATED_USER_ID" \
    --trucker "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "trucker-memberships" "$CREATED_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
    fi
else
    fail "Failed to create trucker membership"
fi

# Only continue if we successfully created a membership
if [[ -z "$CREATED_MEMBERSHIP_ID" || "$CREATED_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid membership ID"
    run_tests
fi

test_name "Create trucker membership with role and admin"
TEST_EMAIL2=$(unique_email)
xbe_json do users create --name "TM Manager User" --email "$TEST_EMAIL2"
if [[ $status -eq 0 ]]; then
    MANAGER_USER_ID=$(json_get ".id")
    xbe_json do trucker-memberships create \
        --user "$MANAGER_USER_ID" \
        --trucker "$CREATED_TRUCKER_ID" \
        --kind "manager" \
        --title "Operations Manager" \
        --is-admin "true"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "trucker-memberships" "$id"
        pass
    else
        fail "Failed to create trucker membership with role and admin"
    fi
else
    skip "Could not create user for manager membership test"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show trucker membership"
xbe_json view trucker-memberships show "$CREATED_MEMBERSHIP_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update trucker membership core fields"
xbe_json do trucker-memberships update "$CREATED_MEMBERSHIP_ID" \
    --kind "manager" \
    --is-admin "true" \
    --title "Fleet Manager" \
    --color-hex "#3366FF" \
    --external-employee-id "TM-EXT-123" \
    --explicit-sort-order "12"
assert_success

test_name "Update trucker membership timing"
xbe_json do trucker-memberships update "$CREATED_MEMBERSHIP_ID" \
    --start-at "2025-01-01T08:00:00Z" \
    --end-at "2025-02-01T17:00:00Z" \
    --drives-shift-type "day" \
    --trailer-coassignments-reset-on "2025-01-15"
assert_success

if [[ -n "$CREATED_PROJECT_OFFICE_ID" ]]; then
    test_name "Update trucker membership project office"
    xbe_json do trucker-memberships update "$CREATED_MEMBERSHIP_ID" \
        --project-office "$CREATED_PROJECT_OFFICE_ID"
    assert_success
else
    test_name "Update trucker membership project office"
    skip "Project office not available"
fi

test_name "Update trucker membership permissions and notifications"
xbe_json do trucker-memberships update "$CREATED_MEMBERSHIP_ID" \
    --can-see-rates-as-driver "true" \
    --can-see-rates-as-manager "true" \
    --can-validate-profit-improvements "false" \
    --is-geofence-violation-team-member "true" \
    --is-unapproved-time-card-subscriber "false" \
    --is-default-job-production-plan-subscriber "false" \
    --enable-recap-notifications "true" \
    --enable-inventory-capacity-notifications "true"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List trucker memberships"
xbe_json view trucker-memberships list --limit 5
assert_success

test_name "List trucker memberships returns array"
xbe_json view trucker-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker memberships"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List trucker memberships with --trucker filter"
xbe_json view trucker-memberships list --trucker "$CREATED_TRUCKER_ID" --limit 10
assert_success

test_name "List trucker memberships with --broker filter"
xbe_json view trucker-memberships list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List trucker memberships with --user filter"
xbe_json view trucker-memberships list --user "$CREATED_USER_ID" --limit 10
assert_success

if [[ -n "$CREATED_PROJECT_OFFICE_ID" ]]; then
    test_name "List trucker memberships with --project-office filter"
    xbe_json view trucker-memberships list --project-office "$CREATED_PROJECT_OFFICE_ID" --limit 10
    assert_success
else
    test_name "List trucker memberships with --project-office filter"
    skip "Project office not available"
fi

test_name "List trucker memberships with --kind filter"
xbe_json view trucker-memberships list --kind "manager" --limit 10
assert_success

test_name "List trucker memberships with --q filter"
xbe_json view trucker-memberships list --q "Trucker" --limit 10
assert_success

test_name "List trucker memberships with --drives-shift-type filter"
xbe_json view trucker-memberships list --drives-shift-type "day" --limit 10
assert_success

test_name "List trucker memberships with --external-employee-id filter"
xbe_json view trucker-memberships list --external-employee-id "TM-EXT-123" --limit 10
assert_success

test_name "List trucker memberships with --is-rate-editor filter"
xbe_json view trucker-memberships list --is-rate-editor "false" --limit 10
assert_success

test_name "List trucker memberships with --is-time-card-auditor filter"
xbe_json view trucker-memberships list --is-time-card-auditor "false" --limit 10
assert_success

test_name "List trucker memberships with --is-equipment-rental-team-member filter"
xbe_json view trucker-memberships list --is-equipment-rental-team-member "false" --limit 10
assert_success

test_name "List trucker memberships with --is-geofence-violation-team-member filter"
xbe_json view trucker-memberships list --is-geofence-violation-team-member "true" --limit 10
assert_success

test_name "List trucker memberships with --is-unapproved-time-card-subscriber filter"
xbe_json view trucker-memberships list --is-unapproved-time-card-subscriber "false" --limit 10
assert_success

test_name "List trucker memberships with --is-default-job-production-plan-subscriber filter"
xbe_json view trucker-memberships list --is-default-job-production-plan-subscriber "false" --limit 10
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List trucker memberships with --limit"
xbe_json view trucker-memberships list --limit 3
assert_success

test_name "List trucker memberships with --offset"
xbe_json view trucker-memberships list --limit 3 --offset 3
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete trucker membership requires --confirm flag"
xbe_run do trucker-memberships delete "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Delete trucker membership with --confirm"
TEST_DEL_EMAIL=$(unique_email)
xbe_json do users create --name "TM Delete User" --email "$TEST_DEL_EMAIL"
if [[ $status -eq 0 ]]; then
    DEL_USER_ID=$(json_get ".id")
    xbe_json do trucker-memberships create \
        --user "$DEL_USER_ID" \
        --trucker "$CREATED_TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do trucker-memberships delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create membership for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create trucker membership without user fails"
xbe_json do trucker-memberships create --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create trucker membership without trucker fails"
xbe_json do trucker-memberships create --user "$CREATED_USER_ID"
assert_failure

test_name "Create trucker membership with unapproved time card subscriber true fails"
xbe_json do trucker-memberships create --user "$CREATED_USER_ID" --trucker "$CREATED_TRUCKER_ID" --is-unapproved-time-card-subscriber "true"
assert_failure

test_name "Update without any fields fails"
xbe_json do trucker-memberships update "$CREATED_MEMBERSHIP_ID"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
