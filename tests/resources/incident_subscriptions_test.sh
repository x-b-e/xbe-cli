#!/bin/bash
#
# XBE CLI Integration Tests: Incident Subscriptions
#
# Tests CRUD operations and list filters for the incident-subscriptions resource.
#
# COVERAGE: Create + update + delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_CUSTOMER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_SUBSCRIPTION_ID=""

describe "Resource: incident-subscriptions"

# ============================================================================
# Prerequisites - Create broker, user, membership
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "IncidentSubBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
    fi
fi

if [[ -z "$CREATED_BROKER_ID" || "$CREATED_BROKER_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker ID"
    run_tests
fi

test_name "Create prerequisite user"
USER_NAME=$(unique_name "IncidentSubUser")
USER_EMAIL=$(unique_email)
USER_MOBILE=$(unique_mobile)

xbe_json do users create --name "$USER_NAME" --email "$USER_EMAIL" --mobile "$USER_MOBILE"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        register_cleanup "users" "$CREATED_USER_ID"
        pass
    else
        fail "Created user but no ID returned"
    fi
else
    fail "Failed to create user"
fi

if [[ -z "$CREATED_USER_ID" || "$CREATED_USER_ID" == "null" ]]; then
    echo "Cannot continue without a valid user ID"
    run_tests
fi

test_name "Create membership for user and broker"
xbe_json do memberships create --user "$CREATED_USER_ID" --organization "Broker|$CREATED_BROKER_ID"

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

# ============================================================================
# Additional resources for filter coverage
# ============================================================================

test_name "Create customer for filter tests"
CUSTOMER_NAME=$(unique_name "IncidentSubCustomer")
xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
    fi
else
    fail "Failed to create customer"
fi

test_name "Create material supplier for filter tests"
MATERIAL_SUPPLIER_NAME=$(unique_name "IncidentSubSupplier")
xbe_json do material-suppliers create --name "$MATERIAL_SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
    fi
else
    fail "Failed to create material supplier"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create incident subscription with broker scope"
xbe_json do incident-subscriptions create \
    --user "$CREATED_USER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --kind "equipment" \
    --contact-method "email_address"

if [[ $status -eq 0 ]]; then
    CREATED_SUBSCRIPTION_ID=$(json_get ".id")
    if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
        register_cleanup "incident-subscriptions" "$CREATED_SUBSCRIPTION_ID"
        pass
    else
        fail "Created subscription but no ID returned"
    fi
else
    fail "Failed to create incident subscription"
fi

if [[ -z "$CREATED_SUBSCRIPTION_ID" || "$CREATED_SUBSCRIPTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid incident subscription ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show incident subscription"
xbe_json view incident-subscriptions show "$CREATED_SUBSCRIPTION_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update incident subscription contact method"
xbe_json do incident-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "mobile_number"
assert_success

test_name "Update incident subscription kind"
xbe_json do incident-subscriptions update "$CREATED_SUBSCRIPTION_ID" --kind "safety"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List incident subscriptions"
xbe_json view incident-subscriptions list --limit 5
assert_success

test_name "List incident subscriptions returns array"
xbe_json view incident-subscriptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list incident subscriptions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List incident subscriptions with --user filter"
xbe_json view incident-subscriptions list --user "$CREATED_USER_ID" --limit 10
assert_success

test_name "List incident subscriptions with --broker filter"
xbe_json view incident-subscriptions list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List incident subscriptions with --customer filter"
xbe_json view incident-subscriptions list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List incident subscriptions with --material-supplier filter"
xbe_json view incident-subscriptions list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --limit 10
assert_success

test_name "List incident subscriptions with --kind filter"
xbe_json view incident-subscriptions list --kind "equipment" --limit 10
assert_success

test_name "List incident subscriptions with --contact-method filter"
xbe_json view incident-subscriptions list --contact-method "email_address" --limit 10
assert_success

test_name "List incident subscriptions with --incident filter"
xbe_json view incident-subscriptions list --incident "1" --limit 10
assert_success

test_name "List incident subscriptions with --incident-start-on filter"
xbe_json view incident-subscriptions list --incident-start-on "2024-01-01" --limit 10
assert_success

test_name "List incident subscriptions with --incident-start-on-min filter"
xbe_json view incident-subscriptions list --incident-start-on-min "2024-01-01" --limit 10
assert_success

test_name "List incident subscriptions with --incident-start-on-max filter"
xbe_json view incident-subscriptions list --incident-start-on-max "2025-12-31" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List incident subscriptions with --limit"
xbe_json view incident-subscriptions list --limit 3
assert_success

test_name "List incident subscriptions with --offset"
xbe_json view incident-subscriptions list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete incident subscription"
xbe_json do incident-subscriptions delete "$CREATED_SUBSCRIPTION_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create incident subscription without user fails"
xbe_json do incident-subscriptions create --broker "$CREATED_BROKER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
