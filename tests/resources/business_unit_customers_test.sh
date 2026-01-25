#!/bin/bash
#
# XBE CLI Integration Tests: Business Unit Customers
#
# Tests list/show and create/delete behavior for business-unit-customers.
#
# COVERAGE: List filters + show + create attributes + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: business-unit-customers"

CREATED_BROKER_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_CUSTOMER_ID=""
CREATED_LINK_ID=""

# ============================================================================
# Prerequisites - Create broker, business unit, and customer
# ============================================================================

test_name "Create prerequisite broker for business unit customer tests"
BROKER_NAME=$(unique_name "BusinessUnitCustomerBroker")

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

test_name "Create prerequisite business unit for business unit customer tests"
BUSINESS_UNIT_NAME=$(unique_name "BusinessUnitCustomerUnit")

xbe_json do business-units create --name "$BUSINESS_UNIT_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
	CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
	if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
		register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
		pass
	else
		fail "Created business unit but no ID returned"
		run_tests
	fi
else
	fail "Failed to create business unit"
	run_tests
fi

test_name "Create prerequisite customer for business unit customer tests"
CUSTOMER_NAME=$(unique_name "BusinessUnitCustomerCustomer")

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

test_name "Create business unit customer without required fields fails"
xbe_json do business-unit-customers create
assert_failure

test_name "Create business unit customer"
xbe_json do business-unit-customers create \
	--business-unit "$CREATED_BUSINESS_UNIT_ID" \
	--customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
	CREATED_LINK_ID=$(json_get ".id")
	if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
		register_cleanup "business-unit-customers" "$CREATED_LINK_ID"
		pass
	else
		fail "Created business unit customer but no ID returned"
		run_tests
	fi
else
	fail "Failed to create business unit customer"
	run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show business unit customer"
xbe_json view business-unit-customers show "$CREATED_LINK_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List business unit customers"
xbe_json view business-unit-customers list --limit 5
assert_success

test_name "List business unit customers with business unit filter"
xbe_json view business-unit-customers list --business-unit "$CREATED_BUSINESS_UNIT_ID" --limit 5
assert_success

test_name "List business unit customers with customer filter"
xbe_json view business-unit-customers list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Delete business unit customer without confirm fails"
xbe_run do business-unit-customers delete "$CREATED_LINK_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete business unit customer"
xbe_run do business-unit-customers delete "$CREATED_LINK_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
