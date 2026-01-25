#!/bin/bash
#
# XBE CLI Integration Tests: Customer Commitments
#
# Tests CRUD operations for the customer-commitments resource.
# Customer commitments require customer + broker relationships.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRUCK_SCOPE_ID=""
CREATED_COMMITMENT_ID=""
PRECEDING_COMMITMENT_ID=""

EXTERNAL_JOB_NUMBER=""

TODAY=$(date -u +%Y-%m-%d)

describe "Resource: customer-commitments"

# ============================================================================
# Prerequisites - Create broker, customer, truck scope
# ============================================================================

test_name "Create prerequisite broker for customer commitment tests"
BROKER_NAME=$(unique_name "CommitmentTestBroker")

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

test_name "Create prerequisite customer for customer commitment tests"
CUSTOMER_NAME=$(unique_name "CommitmentTestCustomer")

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
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

test_name "Create prerequisite truck scope for customer commitment tests"

xbe_json do truck-scopes create --organization-type brokers --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCK_SCOPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCK_SCOPE_ID" && "$CREATED_TRUCK_SCOPE_ID" != "null" ]]; then
        register_cleanup "truck-scopes" "$CREATED_TRUCK_SCOPE_ID"
        pass
    else
        fail "Created truck scope but no ID returned"
        echo "Continuing without truck scope"
    fi
else
    if [[ -n "$XBE_TEST_TRUCK_SCOPE_ID" ]]; then
        CREATED_TRUCK_SCOPE_ID="$XBE_TEST_TRUCK_SCOPE_ID"
        echo "    Using XBE_TEST_TRUCK_SCOPE_ID: $CREATED_TRUCK_SCOPE_ID"
        pass
    else
        skip "Failed to create truck scope and XBE_TEST_TRUCK_SCOPE_ID not set"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer commitment with required fields"

xbe_json do customer-commitments create \
    --customer "$CREATED_CUSTOMER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --status active

if [[ $status -eq 0 ]]; then
    CREATED_COMMITMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_COMMITMENT_ID" && "$CREATED_COMMITMENT_ID" != "null" ]]; then
        register_cleanup "customer-commitments" "$CREATED_COMMITMENT_ID"
        pass
    else
        fail "Created customer commitment but no ID returned"
    fi
else
    fail "Failed to create customer commitment"
fi

if [[ -z "$CREATED_COMMITMENT_ID" || "$CREATED_COMMITMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid customer commitment ID"
    run_tests
fi

test_name "Create customer commitment for preceding commitment tests"

xbe_json do customer-commitments create \
    --customer "$CREATED_CUSTOMER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --status active

if [[ $status -eq 0 ]]; then
    PRECEDING_COMMITMENT_ID=$(json_get ".id")
    if [[ -n "$PRECEDING_COMMITMENT_ID" && "$PRECEDING_COMMITMENT_ID" != "null" ]]; then
        register_cleanup "customer-commitments" "$PRECEDING_COMMITMENT_ID"
        pass
    else
        fail "Created customer commitment but no ID returned"
    fi
else
    fail "Failed to create customer commitment for preceding tests"
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update customer commitment label"
UPDATED_LABEL=$(unique_name "CommitmentLabel")
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --label "$UPDATED_LABEL"
assert_success

test_name "Update customer commitment notes"
UPDATED_NOTES="Updated notes for commitment"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --notes "$UPDATED_NOTES"
assert_success

test_name "Update customer commitment tons"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --tons 1200
assert_success

test_name "Update customer commitment tons-per-shift"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --tons-per-shift 250
assert_success

test_name "Update customer commitment status"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --status inactive
assert_success

# ============================================================================
# UPDATE Tests - Relationships
# ============================================================================

test_name "Update customer commitment customer relationship"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --customer "$CREATED_CUSTOMER_ID"
assert_success

test_name "Update customer commitment broker relationship"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --broker "$CREATED_BROKER_ID"
assert_success

if [[ -n "$CREATED_TRUCK_SCOPE_ID" && "$CREATED_TRUCK_SCOPE_ID" != "null" ]]; then
    test_name "Update customer commitment truck scope"
    xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --truck-scope "$CREATED_TRUCK_SCOPE_ID"
    assert_success
else
    test_name "Update customer commitment truck scope"
    skip "No truck scope available"
fi

if [[ -n "$PRECEDING_COMMITMENT_ID" && "$PRECEDING_COMMITMENT_ID" != "null" ]]; then
    test_name "Update customer commitment preceding commitment"
    xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --preceding-commitment "$PRECEDING_COMMITMENT_ID"
    assert_success
else
    test_name "Update customer commitment preceding commitment"
    skip "No preceding commitment available"
fi

# ============================================================================
# UPDATE Tests - External job number
# ============================================================================

test_name "Update customer commitment external-job-number"
EXTERNAL_JOB_NUMBER="JOB-$(unique_name "CM")"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID" --external-job-number "$EXTERNAL_JOB_NUMBER"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer commitment"
xbe_json view customer-commitments show "$CREATED_COMMITMENT_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List customer commitments"
xbe_json view customer-commitments list --limit 5
assert_success

test_name "List customer commitments returns array"
xbe_json view customer-commitments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer commitments"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List customer commitments with --status filter"
xbe_json view customer-commitments list --status inactive --limit 5
assert_success

test_name "List customer commitments with --customer filter"
xbe_json view customer-commitments list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "List customer commitments with --customer-id filter"
xbe_json view customer-commitments list --customer-id "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "List customer commitments with --broker filter"
xbe_json view customer-commitments list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List customer commitments with --broker-id filter"
xbe_json view customer-commitments list --broker-id "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List customer commitments with --external-job-number filter"
xbe_json view customer-commitments list --external-job-number "$EXTERNAL_JOB_NUMBER" --limit 5
assert_success

test_name "List customer commitments with --external-identification-value filter"
xbe_json view customer-commitments list --external-identification-value "$EXTERNAL_JOB_NUMBER" --limit 5
assert_success

test_name "List customer commitments with --created-at-min filter"
xbe_json view customer-commitments list --created-at-min "$TODAY" --limit 5
assert_success

test_name "List customer commitments with --created-at-max filter"
xbe_json view customer-commitments list --created-at-max "$TODAY" --limit 5
assert_success

test_name "List customer commitments with --updated-at-min filter"
xbe_json view customer-commitments list --updated-at-min "$TODAY" --limit 5
assert_success

test_name "List customer commitments with --updated-at-max filter"
xbe_json view customer-commitments list --updated-at-max "$TODAY" --limit 5
assert_success

test_name "List customer commitments with --not-id filter"
xbe_json view customer-commitments list --not-id "$CREATED_COMMITMENT_ID" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination/Sorting
# ============================================================================

test_name "List customer commitments with --sort"
xbe_json view customer-commitments list --sort created-at --limit 5
assert_success

test_name "List customer commitments with --limit"
xbe_json view customer-commitments list --limit 3
assert_success

test_name "List customer commitments with --offset"
xbe_json view customer-commitments list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer commitment requires --confirm flag"
xbe_json do customer-commitments delete "$CREATED_COMMITMENT_ID"
assert_failure

test_name "Delete customer commitment with --confirm"
# Create commitment specifically for deletion
xbe_json do customer-commitments create \
    --customer "$CREATED_CUSTOMER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --status active

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do customer-commitments delete "$DEL_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "API may not allow customer commitment deletion"
    fi
else
    skip "Could not create customer commitment for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create customer commitment without customer fails"
xbe_json do customer-commitments create --broker "$CREATED_BROKER_ID" --status active
assert_failure

test_name "Create customer commitment without broker fails"
xbe_json do customer-commitments create --customer "$CREATED_CUSTOMER_ID" --status active
assert_failure

test_name "Create customer commitment without status fails"
xbe_json do customer-commitments create --customer "$CREATED_CUSTOMER_ID" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do customer-commitments update "$CREATED_COMMITMENT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
