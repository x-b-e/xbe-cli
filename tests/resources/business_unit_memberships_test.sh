#!/bin/bash
#
# XBE CLI Integration Tests: Business Unit Memberships
#
# Tests CRUD operations for the business-unit-memberships resource.
# Business unit memberships associate broker memberships with business units.
#
# COVERAGE: create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_BU_MEMBERSHIP_ID=""

KIND_CREATE="technician"
KIND_UPDATE="general"

describe "Resource: business-unit-memberships"

# ============================================================================
# Prerequisites - Create broker, business unit, user, membership
# ============================================================================

test_name "Create prerequisite broker for business unit membership tests"
BROKER_NAME=$(unique_name "BusinessUnitMembershipBroker")

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

test_name "Create prerequisite business unit for membership tests"
BUSINESS_UNIT_NAME=$(unique_name "BusinessUnitMembershipBU")

xbe_json do business-units create \
    --name "$BUSINESS_UNIT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business unit but no ID returned"
        echo "Cannot continue without a business unit"
        run_tests
    fi
else
    fail "Failed to create business unit"
    echo "Cannot continue without a business unit"
    run_tests
fi

test_name "Create prerequisite user for membership tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Business Unit Membership User" \
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

test_name "Create broker membership for business unit membership"
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

test_name "Create business unit membership"
xbe_json do business-unit-memberships create \
    --business-unit "$CREATED_BUSINESS_UNIT_ID" \
    --membership "$CREATED_MEMBERSHIP_ID" \
    --kind "$KIND_CREATE"

if [[ $status -eq 0 ]]; then
    CREATED_BU_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_BU_MEMBERSHIP_ID" && "$CREATED_BU_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "business-unit-memberships" "$CREATED_BU_MEMBERSHIP_ID"
        pass
    else
        fail "Created business unit membership but no ID returned"
    fi
else
    fail "Failed to create business unit membership"
fi

# Only continue if we successfully created a business unit membership
if [[ -z "$CREATED_BU_MEMBERSHIP_ID" || "$CREATED_BU_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid business unit membership ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update business unit membership kind"
xbe_json do business-unit-memberships update "$CREATED_BU_MEMBERSHIP_ID" \
    --kind "$KIND_UPDATE"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show business unit membership"
xbe_json view business-unit-memberships show "$CREATED_BU_MEMBERSHIP_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List business unit memberships"
xbe_json view business-unit-memberships list --limit 5
assert_success

test_name "List business unit memberships returns array"
xbe_json view business-unit-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list business unit memberships"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List business unit memberships with --business-unit filter"
xbe_json view business-unit-memberships list --business-unit "$CREATED_BUSINESS_UNIT_ID" --limit 10
assert_success

test_name "List business unit memberships with --membership filter"
xbe_json view business-unit-memberships list --membership "$CREATED_MEMBERSHIP_ID" --limit 10
assert_success

test_name "List business unit memberships with --kind filter"
xbe_json view business-unit-memberships list --kind "$KIND_UPDATE" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete business unit membership"
xbe_run do business-unit-memberships delete "$CREATED_BU_MEMBERSHIP_ID" --confirm
assert_success

run_tests
