#!/bin/bash
#
# XBE CLI Integration Tests: Developer Memberships
#
# Tests CRUD operations for the developer-memberships resource.
# Developer memberships define relationships between users and developers.
#
# NOTE: This test requires creating prerequisite resources: broker, developer, user, project office
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEMBERSHIP_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_USER_ID=""
CREATED_PROJECT_OFFICE_ID=""
EXTERNAL_EMPLOYEE_ID=""

describe "Resource: developer-memberships"

# ============================================================================
# Prerequisites - Create broker, developer, project office, and user
# ============================================================================

test_name "Create prerequisite broker for developer membership tests"
BROKER_NAME=$(unique_name "DeveloperMembershipBroker")

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

test_name "Create prerequisite developer"
DEVELOPER_NAME=$(unique_name "DeveloperMembershipDeveloper")

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
    if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
        CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
        echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
        pass
    else
        fail "Failed to create developer and XBE_TEST_DEVELOPER_ID not set"
        echo "Cannot continue without a developer"
        run_tests
    fi
fi

test_name "Create project office for developer membership tests"
PROJECT_OFFICE_NAME=$(unique_name "DeveloperMembershipOffice")
PROJECT_OFFICE_ABBR="DMO${RANDOM}"

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

test_name "Create prerequisite user for developer membership tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Developer Membership Test User" \
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

test_name "Create developer membership with required fields"

xbe_json do developer-memberships create \
    --user "$CREATED_USER_ID" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "developer-memberships" "$CREATED_MEMBERSHIP_ID"
        pass
    else
        fail "Created developer membership but no ID returned"
    fi
else
    fail "Failed to create developer membership"
fi

# Only continue if we successfully created a membership
if [[ -z "$CREATED_MEMBERSHIP_ID" || "$CREATED_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer membership ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

EXTERNAL_EMPLOYEE_ID="EXT-$(date +%s)-$RANDOM"

test_name "Update developer membership core fields"
xe_developer_memberships_core_update=(
    do developer-memberships update "$CREATED_MEMBERSHIP_ID"
    --kind "manager"
    --is-admin "true"
    --title "Developer Manager"
    --color-hex "#112233"
    --external-employee-id "$EXTERNAL_EMPLOYEE_ID"
    --explicit-sort-order "10"
    --start-at "2024-01-01T00:00:00Z"
    --end-at "2024-12-31T00:00:00Z"
    --drives-shift-type "day"
    --project-office "$CREATED_PROJECT_OFFICE_ID"
)

xbe_json "${xe_developer_memberships_core_update[@]}"
assert_success

test_name "Update developer membership permissions"
xbe_json do developer-memberships update "$CREATED_MEMBERSHIP_ID" \
    --can-see-rates-as-manager "false" \
    --can-validate-profit-improvements "false" \
    --is-geofence-violation-team-member "true"
assert_success

test_name "Update developer membership notifications"
xbe_json do developer-memberships update "$CREATED_MEMBERSHIP_ID" \
    --is-unapproved-time-card-subscriber "true" \
    --is-default-job-production-plan-subscriber "true" \
    --enable-inventory-capacity-notifications "true"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List developer memberships"
xbe_json view developer-memberships list --limit 5
assert_success

test_name "List developer memberships returns array"
xbe_json view developer-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list developer memberships"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List developer memberships with --broker filter"
xbe_json view developer-memberships list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List developer memberships with --organization filter"
xbe_json view developer-memberships list --organization "Developer|$CREATED_DEVELOPER_ID" --limit 10
assert_success

test_name "List developer memberships with --project-office filter"
xbe_json view developer-memberships list --project-office "$CREATED_PROJECT_OFFICE_ID" --limit 10
assert_success

test_name "List developer memberships with --user filter"
xbe_json view developer-memberships list --user "$CREATED_USER_ID" --limit 10
assert_success

test_name "List developer memberships with --kind filter"
xbe_json view developer-memberships list --kind "manager" --limit 10
assert_success

test_name "List developer memberships with --q filter"
xbe_json view developer-memberships list --q "Developer Membership Test User" --limit 10
assert_success

test_name "List developer memberships with --drives-shift-type filter"
xbe_json view developer-memberships list --drives-shift-type "day" --limit 10
assert_success

test_name "List developer memberships with --external-employee-id filter"
xbe_json view developer-memberships list --external-employee-id "$EXTERNAL_EMPLOYEE_ID" --limit 10
assert_success

test_name "List developer memberships with --is-rate-editor filter"
xbe_json view developer-memberships list --is-rate-editor "true" --limit 10
assert_success

test_name "List developer memberships with --is-time-card-auditor filter"
xbe_json view developer-memberships list --is-time-card-auditor "true" --limit 10
assert_success

test_name "List developer memberships with --is-equipment-rental-team-member filter"
xbe_json view developer-memberships list --is-equipment-rental-team-member "true" --limit 10
assert_success

test_name "List developer memberships with --is-geofence-violation-team-member filter"
xbe_json view developer-memberships list --is-geofence-violation-team-member "true" --limit 10
assert_success

test_name "List developer memberships with --is-unapproved-time-card-subscriber filter"
xbe_json view developer-memberships list --is-unapproved-time-card-subscriber "true" --limit 10
assert_success

test_name "List developer memberships with --is-default-job-production-plan-subscriber filter"
xbe_json view developer-memberships list --is-default-job-production-plan-subscriber "true" --limit 10
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show developer membership"
xbe_json view developer-memberships show "$CREATED_MEMBERSHIP_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer membership requires --confirm flag"
xbe_run do developer-memberships delete "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Delete developer membership with --confirm"
# Create a membership specifically for deletion
TEST_DEL_EMAIL=$(unique_email)
xe_developer_memberships_delete_user=(
    do users create
    --name "Developer Membership Delete User"
    --email "$TEST_DEL_EMAIL"
)

xbe_json "${xe_developer_memberships_delete_user[@]}"
if [[ $status -eq 0 ]]; then
    DEL_USER_ID=$(json_get ".id")
    xbe_json do developer-memberships create \
        --user "$DEL_USER_ID" \
        --developer "$CREATED_DEVELOPER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do developer-memberships delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create developer membership for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create developer membership without user fails"
xbe_json do developer-memberships create --developer "$CREATED_DEVELOPER_ID"
assert_failure

test_name "Create developer membership without developer fails"
xbe_json do developer-memberships create --user "$CREATED_USER_ID"
assert_failure

test_name "Update developer membership without any fields fails"
xbe_json do developer-memberships update "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Update developer membership with can-validate-profit-improvements true fails"
xbe_json do developer-memberships update "$CREATED_MEMBERSHIP_ID" --can-validate-profit-improvements "true"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
