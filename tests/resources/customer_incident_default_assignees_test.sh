#!/bin/bash
#
# XBE CLI Integration Tests: Customer Incident Default Assignees
#
# Tests CRUD operations for the customer_incident_default_assignees resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_DEFAULT_ASSIGNEE_ID=""
ACTIVE_DEFAULT_ASSIGNEE_ID=""

describe "Resource: customer_incident_default_assignees"

# ============================================================================
# Prerequisites - Create broker, customer, user, and membership
# ============================================================================

test_name "Create prerequisite broker for customer incident default assignee tests"
BROKER_NAME=$(unique_name "CIDATestBroker")

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

test_name "Create prerequisite customer for customer incident default assignee tests"
CUSTOMER_NAME=$(unique_name "CIDATestCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

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

test_name "Create prerequisite user for customer incident default assignee tests"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Customer Incident Assignee" \
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

test_name "Create broker membership for default assignee user"
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer incident default assignee"
xbe_json do customer-incident-default-assignees create \
    --customer "$CREATED_CUSTOMER_ID" \
    --default-assignee "$CREATED_USER_ID" \
    --kind safety

if [[ $status -eq 0 ]]; then
    CREATED_DEFAULT_ASSIGNEE_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEFAULT_ASSIGNEE_ID" && "$CREATED_DEFAULT_ASSIGNEE_ID" != "null" ]]; then
        ACTIVE_DEFAULT_ASSIGNEE_ID="$CREATED_USER_ID"
        register_cleanup "customer-incident-default-assignees" "$CREATED_DEFAULT_ASSIGNEE_ID"
        pass
    else
        fail "Created default assignee but no ID returned"
    fi
else
    fail "Failed to create customer incident default assignee"
fi

if [[ -z "$CREATED_DEFAULT_ASSIGNEE_ID" || "$CREATED_DEFAULT_ASSIGNEE_ID" == "null" ]]; then
    echo "Cannot continue without a valid default assignee ID"
    run_tests
fi

test_name "Create default assignee without kind fails"
xbe_json do customer-incident-default-assignees create \
    --customer "$CREATED_CUSTOMER_ID" \
    --default-assignee "$CREATED_USER_ID"
assert_failure

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update customer incident default assignee kind"
xbe_json do customer-incident-default-assignees update "$CREATED_DEFAULT_ASSIGNEE_ID" --kind quality
assert_success

test_name "Update customer incident default assignee user"
TEST_EMAIL2=$(unique_email)
xbe_json do users create \
    --name "Customer Incident Assignee Two" \
    --email "$TEST_EMAIL2"

if [[ $status -eq 0 ]]; then
    UPDATED_USER_ID=$(json_get ".id")
    xbe_json do memberships create \
        --user "$UPDATED_USER_ID" \
        --organization "Broker|$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        UPDATED_MEMBERSHIP_ID=$(json_get ".id")
        register_cleanup "memberships" "$UPDATED_MEMBERSHIP_ID"
        xbe_json do customer-incident-default-assignees update "$CREATED_DEFAULT_ASSIGNEE_ID" --default-assignee "$UPDATED_USER_ID"
        if [[ $status -eq 0 ]]; then
            ACTIVE_DEFAULT_ASSIGNEE_ID="$UPDATED_USER_ID"
            pass
        else
            fail "Failed to update default assignee user"
        fi
    else
        skip "Could not create membership for updated user"
    fi
else
    skip "Could not create user for default assignee update"
fi

test_name "Update without any fields fails"
xbe_json do customer-incident-default-assignees update "$CREATED_DEFAULT_ASSIGNEE_ID"
assert_failure

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List customer incident default assignees"
xbe_json view customer-incident-default-assignees list --limit 5
assert_success

test_name "List customer incident default assignees returns array"
xbe_json view customer-incident-default-assignees list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer incident default assignees"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List customer incident default assignees with --customer filter"
xbe_json view customer-incident-default-assignees list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List customer incident default assignees with --default-assignee filter"
xbe_json view customer-incident-default-assignees list --default-assignee "$ACTIVE_DEFAULT_ASSIGNEE_ID" --limit 10
assert_success

test_name "List customer incident default assignees with --kind filter"
xbe_json view customer-incident-default-assignees list --kind quality --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List customer incident default assignees with --limit"
xbe_json view customer-incident-default-assignees list --limit 3
assert_success

test_name "List customer incident default assignees with --offset"
xbe_json view customer-incident-default-assignees list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer incident default assignee requires --confirm flag"
xbe_run do customer-incident-default-assignees delete "$CREATED_DEFAULT_ASSIGNEE_ID"
assert_failure

test_name "Delete customer incident default assignee with --confirm"
xbe_json do customer-incident-default-assignees create \
    --customer "$CREATED_CUSTOMER_ID" \
    --default-assignee "$ACTIVE_DEFAULT_ASSIGNEE_ID" \
    --kind equipment

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do customer-incident-default-assignees delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create default assignee for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
