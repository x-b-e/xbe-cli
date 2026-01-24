#!/bin/bash
#
# XBE CLI Integration Tests: Customer Memberships
#
# Tests CRUD operations for the customer-memberships resource.
# Customer memberships define relationships between users and customers.
#
# NOTE: This test requires creating prerequisite resources: broker, customer, user
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEMBERSHIP_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_USER_ID=""
PROJECT_OFFICE_ID=""

describe "Resource: customer-memberships"

# ============================================================================
# Prerequisites - Create broker, customer, user, project office
# ============================================================================

test_name "Create prerequisite broker for customer membership tests"
BROKER_NAME=$(unique_name "CustomerMembershipBroker")

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

test_name "Create prerequisite customer for customer membership tests"
CUSTOMER_NAME=$(unique_name "CustomerMembershipCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create prerequisite user for customer membership tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create --name "Customer Membership User" --email "$TEST_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        # Users cannot be deleted via API, so we don't register cleanup
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

test_name "Create prerequisite project office for customer membership tests"
PROJECT_OFFICE_NAME=$(unique_name "CustomerMembershipOffice")
PROJECT_OFFICE_ABBREV="CM$((RANDOM % 1000))"

xbe_json do project-offices create --name "$PROJECT_OFFICE_NAME" --abbreviation "$PROJECT_OFFICE_ABBREV" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    PROJECT_OFFICE_ID=$(json_get ".id")
    if [[ -n "$PROJECT_OFFICE_ID" && "$PROJECT_OFFICE_ID" != "null" ]]; then
        register_cleanup "project-offices" "$PROJECT_OFFICE_ID"
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer membership with required fields"

xbe_json do customer-memberships create \
    --user "$CREATED_USER_ID" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "customer-memberships" "$CREATED_MEMBERSHIP_ID"
        pass
    else
        fail "Created customer membership but no ID returned"
    fi
else
    fail "Failed to create customer membership"
fi

if [[ -z "$CREATED_MEMBERSHIP_ID" || "$CREATED_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid customer membership ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests - Attribute Coverage
# ============================================================================

test_name "Update customer membership role fields"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID" \
    --kind "manager" \
    --title "Customer Manager" \
    --color-hex "#FF5500" \
    --external-employee-id "EXT-1234" \
    --explicit-sort-order "10" \
    --drives-shift-type "day"
assert_success

test_name "Update customer membership is-admin"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID" --is-admin "true"
assert_success

test_name "Update customer membership dates"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID" \
    --start-at "2025-01-01T00:00:00Z" \
    --end-at "2026-01-01T00:00:00Z"
assert_success

test_name "Update customer membership project office"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID" --project-office "$PROJECT_OFFICE_ID"
assert_success

test_name "Update customer membership permissions"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID" \
    --can-see-rates-as-driver "false" \
    --can-see-rates-as-manager "true" \
    --can-validate-profit-improvements "false" \
    --is-rate-editor "false" \
    --is-time-card-auditor "false" \
    --is-equipment-rental-team-member "false" \
    --is-geofence-violation-team-member "true"
assert_success

test_name "Update customer membership notifications"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID" \
    --is-unapproved-time-card-subscriber "true" \
    --is-default-job-production-plan-subscriber "true" \
    --enable-recap-notifications "true" \
    --enable-inventory-capacity-notifications "true"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List customer memberships"
xbe_json view customer-memberships list --limit 5
assert_success

test_name "List customer memberships returns array"
xbe_json view customer-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer memberships"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List customer memberships with --customer filter"
xbe_json view customer-memberships list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List customer memberships with --organization filter"
xbe_json view customer-memberships list --organization "Customer|$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List customer memberships with --broker filter"
xbe_json view customer-memberships list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List customer memberships with --user filter"
xbe_json view customer-memberships list --user "$CREATED_USER_ID" --limit 10
assert_success

test_name "List customer memberships with --project-office filter"
xbe_json view customer-memberships list --project-office "$PROJECT_OFFICE_ID" --limit 10
assert_success

test_name "List customer memberships with --kind filter"
xbe_json view customer-memberships list --kind "manager" --limit 10
assert_success

test_name "List customer memberships with --q filter"
xbe_json view customer-memberships list --q "Customer Membership User" --limit 10
assert_success

test_name "List customer memberships with --drives-shift-type filter"
xbe_json view customer-memberships list --drives-shift-type "day" --limit 10
assert_success

test_name "List customer memberships with --external-employee-id filter"
xbe_json view customer-memberships list --external-employee-id "EXT-1234" --limit 10
assert_success

test_name "List customer memberships with --is-rate-editor filter"
xbe_json view customer-memberships list --is-rate-editor "false" --limit 10
assert_success

test_name "List customer memberships with --is-time-card-auditor filter"
xbe_json view customer-memberships list --is-time-card-auditor "false" --limit 10
assert_success

test_name "List customer memberships with --is-equipment-rental-team-member filter"
xbe_json view customer-memberships list --is-equipment-rental-team-member "false" --limit 10
assert_success

test_name "List customer memberships with --is-geofence-violation-team-member filter"
xbe_json view customer-memberships list --is-geofence-violation-team-member "true" --limit 10
assert_success

test_name "List customer memberships with --is-unapproved-time-card-subscriber filter"
xbe_json view customer-memberships list --is-unapproved-time-card-subscriber "true" --limit 10
assert_success

test_name "List customer memberships with --is-default-job-production-plan-subscriber filter"
xbe_json view customer-memberships list --is-default-job-production-plan-subscriber "true" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List customer memberships with --limit"
xbe_json view customer-memberships list --limit 3
assert_success

test_name "List customer memberships with --offset"
xbe_json view customer-memberships list --limit 3 --offset 1
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer membership"
xbe_json view customer-memberships show "$CREATED_MEMBERSHIP_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer membership requires --confirm flag"
xbe_run do customer-memberships delete "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Delete customer membership with --confirm"
TEST_DEL_EMAIL=$(unique_email)
xbe_json do users create --name "Delete Customer Membership User" --email "$TEST_DEL_EMAIL"
if [[ $status -eq 0 ]]; then
    DEL_USER_ID=$(json_get ".id")
    xbe_json do customer-memberships create \
        --user "$DEL_USER_ID" \
        --customer "$CREATED_CUSTOMER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do customer-memberships delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create customer membership for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create customer membership without user fails"
xbe_json do customer-memberships create --customer "$CREATED_CUSTOMER_ID"
assert_failure

test_name "Create customer membership without customer fails"
xbe_json do customer-memberships create --user "$CREATED_USER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do customer-memberships update "$CREATED_MEMBERSHIP_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
