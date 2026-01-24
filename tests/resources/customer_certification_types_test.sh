#!/bin/bash
#
# XBE CLI Integration Tests: Customer Certification Types
#
# Tests CRUD operations for the customer_certification_types resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_CERTIFICATION_TYPE_ID=""
CREATED_CUSTOMER_CERT_TYPE_ID=""
SECOND_BROKER_ID=""
SECOND_CUSTOMER_ID=""
SECOND_CERTIFICATION_TYPE_ID=""

describe "Resource: customer_certification_types"

# ============================================================================
# Prerequisites - Create broker, customer, and certification type
# ============================================================================

test_name "Create prerequisite broker for customer certification type tests"
BROKER_NAME=$(unique_name "CCTBroker")

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

test_name "Create prerequisite customer for customer certification type tests"
CUSTOMER_NAME=$(unique_name "CCTCustomer")

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

test_name "Create prerequisite certification type for customer certification type tests"
CERT_NAME=$(unique_name "CCTCert")

xbe_json do certification-types create \
    --name "$CERT_NAME" \
    --can-apply-to "Trucker" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CERTIFICATION_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CERTIFICATION_TYPE_ID" && "$CREATED_CERTIFICATION_TYPE_ID" != "null" ]]; then
        register_cleanup "certification-types" "$CREATED_CERTIFICATION_TYPE_ID"
        pass
    else
        fail "Created certification type but no ID returned"
        echo "Cannot continue without a certification type"
        run_tests
    fi
else
    fail "Failed to create certification type"
    echo "Cannot continue without a certification type"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer certification type with required fields"
xbe_json do customer-certification-types create \
    --customer "$CREATED_CUSTOMER_ID" \
    --certification-type "$CREATED_CERTIFICATION_TYPE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_CERT_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_CERT_TYPE_ID" && "$CREATED_CUSTOMER_CERT_TYPE_ID" != "null" ]]; then
        register_cleanup "customer-certification-types" "$CREATED_CUSTOMER_CERT_TYPE_ID"
        pass
    else
        fail "Created customer certification type but no ID returned"
    fi
else
    fail "Failed to create customer certification type"
fi

if [[ -z "$CREATED_CUSTOMER_CERT_TYPE_ID" || "$CREATED_CUSTOMER_CERT_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid customer certification type ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer certification type"
xbe_json view customer-certification-types show "$CREATED_CUSTOMER_CERT_TYPE_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Create secondary broker for update"
SECOND_BROKER_NAME=$(unique_name "CCTBroker2")

xbe_json do brokers create --name "$SECOND_BROKER_NAME"

if [[ $status -eq 0 ]]; then
    SECOND_BROKER_ID=$(json_get ".id")
    if [[ -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$SECOND_BROKER_ID"
        pass
    else
        fail "Created second broker but no ID returned"
        echo "Cannot continue without a second broker"
        run_tests
    fi
else
        fail "Failed to create second broker"
        echo "Cannot continue without a second broker"
        run_tests
fi

test_name "Create second customer for update"
SECOND_CUSTOMER_NAME=$(unique_name "CCTCustomer2")

xbe_json do customers create \
    --name "$SECOND_CUSTOMER_NAME" \
    --broker "$SECOND_BROKER_ID"

if [[ $status -eq 0 ]]; then
    SECOND_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$SECOND_CUSTOMER_ID" && "$SECOND_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$SECOND_CUSTOMER_ID"
        pass
    else
        fail "Created second customer but no ID returned"
        echo "Cannot continue without a second customer"
        run_tests
    fi
else
    fail "Failed to create second customer"
    echo "Cannot continue without a second customer"
    run_tests
fi

test_name "Create certification type for second broker"
SECOND_CERT_NAME=$(unique_name "CCTCert2")

xbe_json do certification-types create \
    --name "$SECOND_CERT_NAME" \
    --can-apply-to "Trucker" \
    --broker "$SECOND_BROKER_ID"

if [[ $status -eq 0 ]]; then
    SECOND_CERTIFICATION_TYPE_ID=$(json_get ".id")
    if [[ -n "$SECOND_CERTIFICATION_TYPE_ID" && "$SECOND_CERTIFICATION_TYPE_ID" != "null" ]]; then
        register_cleanup "certification-types" "$SECOND_CERTIFICATION_TYPE_ID"
        pass
    else
        fail "Created second certification type but no ID returned"
        echo "Cannot continue without a second certification type"
        run_tests
    fi
else
    fail "Failed to create second certification type"
    echo "Cannot continue without a second certification type"
    run_tests
fi

test_name "Update customer certification type customer and certification type"
xbe_json do customer-certification-types update "$CREATED_CUSTOMER_CERT_TYPE_ID" \
    --customer "$SECOND_CUSTOMER_ID" \
    --certification-type "$SECOND_CERTIFICATION_TYPE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List customer certification types"
xbe_json view customer-certification-types list --limit 5
assert_success

test_name "List customer certification types returns array"
xbe_json view customer-certification-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer certification types"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List customer certification types with --created-at-min filter"
xbe_json view customer-certification-types list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List customer certification types with --created-at-max filter"
xbe_json view customer-certification-types list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List customer certification types with --is-created-at filter"
xbe_json view customer-certification-types list --is-created-at true --limit 5
assert_success

test_name "List customer certification types with --updated-at-min filter"
xbe_json view customer-certification-types list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List customer certification types with --updated-at-max filter"
xbe_json view customer-certification-types list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List customer certification types with --is-updated-at filter"
xbe_json view customer-certification-types list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List customer certification types with --limit"
xbe_json view customer-certification-types list --limit 3
assert_success

test_name "List customer certification types with --offset"
xbe_json view customer-certification-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer certification type requires --confirm flag"
xbe_run do customer-certification-types delete "$CREATED_CUSTOMER_CERT_TYPE_ID"
assert_failure

test_name "Delete customer certification type with --confirm"
xbe_run do customer-certification-types delete "$CREATED_CUSTOMER_CERT_TYPE_ID" --confirm
assert_success

run_tests
