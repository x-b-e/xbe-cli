#!/bin/bash
#
# XBE CLI Integration Tests: Customer Vendors
#
# Tests CRUD operations for the customer-vendors resource.
# Customer vendors link customers and vendors (truckers) with optional external accounting IDs.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRUCKER_ID=""
CREATED_CUSTOMER_VENDOR_ID=""

describe "Resource: customer-vendors"

# ============================================================================
# Prerequisites - Create broker, customer, and vendor (trucker)
# ============================================================================

test_name "Create prerequisite broker for customer-vendor tests"
BROKER_NAME=$(unique_name "CustomerVendorBroker")

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

test_name "Create prerequisite trucker for customer-vendor tests"
TRUCKER_NAME=$(unique_name "CustomerVendorTrucker")

xbe_json do truckers create --name "$TRUCKER_NAME" --broker "$CREATED_BROKER_ID" --company-address "123 Test St" --skip-company-address-geocoding true
if [[ $status -eq 0 ]]; then
	CREATED_TRUCKER_ID=$(json_get ".id")
	if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
		register_cleanup "truckers" "$CREATED_TRUCKER_ID"
		pass
	else
		fail "Created trucker but no ID returned"
		run_tests
	fi
else
	fail "Failed to create trucker"
	run_tests
fi

test_name "Create prerequisite customer for customer-vendor tests"
CUSTOMER_NAME=$(unique_name "CustomerVendorCustomer")

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

test_name "Create customer-vendor with required fields"
EXTERNAL_ID="CV-$(date +%s)"

xbe_json do customer-vendors create \
	--customer "$CREATED_CUSTOMER_ID" \
	--vendor "Trucker|$CREATED_TRUCKER_ID" \
	--external-accounting-customer-vendor-id "$EXTERNAL_ID"

if [[ $status -eq 0 ]]; then
	CREATED_CUSTOMER_VENDOR_ID=$(json_get ".id")
	if [[ -n "$CREATED_CUSTOMER_VENDOR_ID" && "$CREATED_CUSTOMER_VENDOR_ID" != "null" ]]; then
		register_cleanup "customer-vendors" "$CREATED_CUSTOMER_VENDOR_ID"
		pass
	else
		fail "Created customer-vendor but no ID returned"
		run_tests
	fi
else
	fail "Failed to create customer-vendor"
	run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer vendor"
xbe_json view customer-vendors show "$CREATED_CUSTOMER_VENDOR_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update customer vendor external-accounting-customer-vendor-id"
UPDATED_EXTERNAL_ID="${EXTERNAL_ID}-U"
xbe_json do customer-vendors update "$CREATED_CUSTOMER_VENDOR_ID" \
	--external-accounting-customer-vendor-id "$UPDATED_EXTERNAL_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List customer vendors"
xbe_json view customer-vendors list --limit 5
assert_success

test_name "List customer vendors with customer filter"
xbe_json view customer-vendors list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "List customer vendors with organization filters"
xbe_json view customer-vendors list \
	--organization "Customer|$CREATED_CUSTOMER_ID" \
	--organization-type "Customer" \
	--organization-id "$CREATED_CUSTOMER_ID" \
	--limit 5
assert_success

test_name "List customer vendors with not-organization-type filter fails (server 500)"
xbe_json view customer-vendors list --not-organization-type "Broker" --limit 5
assert_failure

test_name "List customer vendors with partner filters"
xbe_json view customer-vendors list \
	--partner "Trucker|$CREATED_TRUCKER_ID" \
	--partner-type "Trucker" \
	--partner-id "$CREATED_TRUCKER_ID" \
	--limit 5
assert_success

test_name "List customer vendors with not-partner-type filter fails (server 500)"
xbe_json view customer-vendors list --not-partner-type "Customer" --limit 5
assert_failure

test_name "List customer vendors with trading-partner-type filter"
xbe_json view customer-vendors list --trading-partner-type "CustomerVendor" --limit 5
assert_success

test_name "List customer vendors with external-identification-value filter"
xbe_json view customer-vendors list --external-identification-value "EXT123" --limit 5
assert_success

test_name "List customer vendors with created/updated filters"
xbe_json view customer-vendors list \
	--created-at-min "2020-01-01T00:00:00Z" \
	--created-at-max "2030-01-01T00:00:00Z" \
	--updated-at-min "2020-01-01T00:00:00Z" \
	--updated-at-max "2030-01-01T00:00:00Z" \
	--limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update customer vendor without fields fails"
xbe_json do customer-vendors update "$CREATED_CUSTOMER_VENDOR_ID"
assert_failure

test_name "Delete customer vendor without confirm fails"
xbe_run do customer-vendors delete "$CREATED_CUSTOMER_VENDOR_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer vendor"
xbe_run do customer-vendors delete "$CREATED_CUSTOMER_VENDOR_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
