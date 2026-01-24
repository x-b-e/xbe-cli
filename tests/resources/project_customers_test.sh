#!/bin/bash
#
# XBE CLI Integration Tests: Project Customers
#
# Tests CRUD operations for the project_customers resource.
#
# COVERAGE: All create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_CUSTOMER_ID=""
CREATED_PROJECT_CUSTOMER_ID=""

describe "Resource: project_customers"

# ============================================================================
# Prerequisites - Create broker, developer, project, customer
# ============================================================================

test_name "Create prerequisite broker for project customer tests"
BROKER_NAME=$(unique_name "ProjectCustomerBroker")

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

test_name "Create developer for project customer tests"
if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
    CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
    echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
    pass
else
    DEV_NAME=$(unique_name "ProjectCustomerDev")
    xbe_json do developers create \
        --name "$DEV_NAME" \
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
fi

test_name "Create project for project customer tests"
if [[ -n "$XBE_TEST_PROJECT_ID" ]]; then
    CREATED_PROJECT_ID="$XBE_TEST_PROJECT_ID"
    echo "    Using XBE_TEST_PROJECT_ID: $CREATED_PROJECT_ID"
    pass
else
    PROJECT_NAME=$(unique_name "ProjectCustomerProject")
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
fi

test_name "Create customer for project customer tests"
if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
    CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
    echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
    pass
else
    CUSTOMER_NAME=$(unique_name "ProjectCustomerCustomer")
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
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project customer with required fields"
xbe_json do project-customers create \
    --project "$CREATED_PROJECT_ID" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_CUSTOMER_ID" && "$CREATED_PROJECT_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "project-customers" "$CREATED_PROJECT_CUSTOMER_ID"
        pass
    else
        fail "Created project customer but no ID returned"
    fi
else
    fail "Failed to create project customer"
fi

if [[ -z "$CREATED_PROJECT_CUSTOMER_ID" || "$CREATED_PROJECT_CUSTOMER_ID" == "null" ]]; then
    echo "Cannot continue without a valid project customer ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project customer"
xbe_json view project-customers show "$CREATED_PROJECT_CUSTOMER_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project customers"
xbe_json view project-customers list --limit 5
assert_success

test_name "List project customers returns array"
xbe_json view project-customers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project customers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project customers with --project filter"
xbe_json view project-customers list --project "$CREATED_PROJECT_ID" --limit 5
assert_success

test_name "List project customers with --customer filter"
xbe_json view project-customers list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "List project customers with --created-at-min filter"
xbe_json view project-customers list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project customers with --created-at-max filter"
xbe_json view project-customers list --created-at-max "2100-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project customers with --updated-at-min filter"
xbe_json view project-customers list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project customers with --updated-at-max filter"
xbe_json view project-customers list --updated-at-max "2100-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project customer without project fails"
xbe_json do project-customers create --customer "$CREATED_CUSTOMER_ID"
assert_failure

test_name "Create project customer without customer fails"
xbe_json do project-customers create --project "$CREATED_PROJECT_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project customer requires --confirm flag"
xbe_run do project-customers delete "$CREATED_PROJECT_CUSTOMER_ID"
assert_failure

test_name "Delete project customer with --confirm"
xbe_run do project-customers delete "$CREATED_PROJECT_CUSTOMER_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
