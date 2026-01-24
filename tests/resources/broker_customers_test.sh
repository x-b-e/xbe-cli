#!/bin/bash
#
# XBE CLI Integration Tests: Broker Customers
#
# Tests CRUD operations for the broker-customers resource.
# Broker customers link brokers and customers with optional external accounting IDs.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_BROKER_CUSTOMER_ID=""

describe "Resource: broker-customers"

# ============================================================================
# Prerequisites - Create broker and customer
# ============================================================================

test_name "Create prerequisite broker for broker-customer tests"
BROKER_NAME=$(unique_name "BrokerCustomerBroker")

xbe_json do brokers create --name "$BROKER_NAME"
if [[ $status -eq 0 ]]; then
	CREATED_BROKER_ID=$(json_get ".id")
	if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
		register_cleanup "brokers" "$CREATED_BROKER_ID"
		pass
	else
		fail "Created broker but no ID returned"
		run_tests
	fi
else
	fail "Failed to create broker"
	run_tests
fi

test_name "Create prerequisite customer for broker-customer tests"
CUSTOMER_NAME=$(unique_name "BrokerCustomerCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
	CREATED_CUSTOMER_ID=$(json_get ".id")
	if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
		register_cleanup "customers" "$CREATED_CUSTOMER_ID"
		pass
	else
		fail "Created customer but no ID returned"
		run_tests
	fi
else
	fail "Failed to create customer"
	run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker-customer with required fields"
EXTERNAL_ID="BC-$(date +%s)"

xbe_json do broker-customers create \
	--broker "$CREATED_BROKER_ID" \
	--customer "$CREATED_CUSTOMER_ID" \
	--external-accounting-broker-customer-id "$EXTERNAL_ID"

if [[ $status -eq 0 ]]; then
	CREATED_BROKER_CUSTOMER_ID=$(json_get ".id")
	if [[ -n "$CREATED_BROKER_CUSTOMER_ID" && "$CREATED_BROKER_CUSTOMER_ID" != "null" ]]; then
		register_cleanup "broker-customers" "$CREATED_BROKER_CUSTOMER_ID"
		pass
	else
		fail "Created broker-customer but no ID returned"
		run_tests
	fi
else
	fail "Failed to create broker-customer"
	run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker customer"
xbe_json view broker-customers show "$CREATED_BROKER_CUSTOMER_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update broker customer external-accounting-broker-customer-id"
UPDATED_EXTERNAL_ID="${EXTERNAL_ID}-U"
xbe_json do broker-customers update "$CREATED_BROKER_CUSTOMER_ID" \
	--external-accounting-broker-customer-id "$UPDATED_EXTERNAL_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List broker customers"
xbe_json view broker-customers list --limit 5
assert_success

test_name "List broker customers with broker filter"
xbe_json view broker-customers list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List broker customers with organization filters"
xbe_json view broker-customers list \
	--organization "Broker|$CREATED_BROKER_ID" \
	--organization-type "Broker" \
	--organization-id "$CREATED_BROKER_ID" \
	--limit 5
assert_success

test_name "List broker customers with not-organization-type filter fails (server 500)"
xbe_json view broker-customers list --not-organization-type "Customer" --limit 5
assert_failure

test_name "List broker customers with partner filters"
xbe_json view broker-customers list \
	--partner "Customer|$CREATED_CUSTOMER_ID" \
	--partner-type "Customer" \
	--partner-id "$CREATED_CUSTOMER_ID" \
	--limit 5
assert_success

test_name "List broker customers with not-partner-type filter fails (server 500)"
xbe_json view broker-customers list --not-partner-type "Trucker" --limit 5
assert_failure

test_name "List broker customers with trading-partner-type filter"
xbe_json view broker-customers list --trading-partner-type "BrokerCustomer" --limit 5
assert_success

test_name "List broker customers with external accounting ID filter"
xbe_json view broker-customers list --external-accounting-broker-customer-id "$UPDATED_EXTERNAL_ID" --limit 5
assert_success

test_name "List broker customers with external-identification-value filter"
xbe_json view broker-customers list --external-identification-value "EXT123" --limit 5
assert_success

test_name "List broker customers with created/updated filters"
xbe_json view broker-customers list \
	--created-at-min "2020-01-01T00:00:00Z" \
	--created-at-max "2030-01-01T00:00:00Z" \
	--updated-at-min "2020-01-01T00:00:00Z" \
	--updated-at-max "2030-01-01T00:00:00Z" \
	--limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update broker customer without fields fails"
xbe_json do broker-customers update "$CREATED_BROKER_CUSTOMER_ID"
assert_failure

test_name "Delete broker customer without confirm fails"
xbe_run do broker-customers delete "$CREATED_BROKER_CUSTOMER_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker customer"
xbe_run do broker-customers delete "$CREATED_BROKER_CUSTOMER_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
