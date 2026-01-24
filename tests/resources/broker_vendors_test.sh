#!/bin/bash
#
# XBE CLI Integration Tests: Broker Vendors
#
# Tests CRUD operations for the broker-vendors resource.
# Broker vendors link brokers and vendors (truckers or material sites) with optional external accounting IDs.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_BROKER_VENDOR_ID=""

describe "Resource: broker-vendors"

# ============================================================================
# Prerequisites - Create broker and vendor (trucker)
# ============================================================================

test_name "Create prerequisite broker for broker-vendor tests"
BROKER_NAME=$(unique_name "BrokerVendorBroker")

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

test_name "Create prerequisite trucker for broker-vendor tests"
TRUCKER_NAME=$(unique_name "BrokerVendorTrucker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker-vendor with required fields"
EXTERNAL_ID="BV-$(date +%s)"

xbe_json do broker-vendors create \
	--broker "$CREATED_BROKER_ID" \
	--vendor "Trucker|$CREATED_TRUCKER_ID" \
	--external-accounting-broker-vendor-id "$EXTERNAL_ID"

if [[ $status -eq 0 ]]; then
	CREATED_BROKER_VENDOR_ID=$(json_get ".id")
	if [[ -n "$CREATED_BROKER_VENDOR_ID" && "$CREATED_BROKER_VENDOR_ID" != "null" ]]; then
		register_cleanup "broker-vendors" "$CREATED_BROKER_VENDOR_ID"
		pass
	else
		fail "Created broker-vendor but no ID returned"
		run_tests
	fi
else
	fail "Failed to create broker-vendor"
	run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker vendor"
xbe_json view broker-vendors show "$CREATED_BROKER_VENDOR_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update broker vendor external-accounting-broker-vendor-id"
UPDATED_EXTERNAL_ID="${EXTERNAL_ID}-U"
xbe_json do broker-vendors update "$CREATED_BROKER_VENDOR_ID" \
	--external-accounting-broker-vendor-id "$UPDATED_EXTERNAL_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List broker vendors"
xbe_json view broker-vendors list --limit 5
assert_success

test_name "List broker vendors with broker filter"
xbe_json view broker-vendors list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List broker vendors with organization filters"
xbe_json view broker-vendors list \
	--organization "Broker|$CREATED_BROKER_ID" \
	--organization-type "Broker" \
	--organization-id "$CREATED_BROKER_ID" \
	--limit 5
assert_success

test_name "List broker vendors with not-organization-type filter fails (server 500)"
xbe_json view broker-vendors list --not-organization-type "Customer" --limit 5
assert_failure

test_name "List broker vendors with partner filters"
xbe_json view broker-vendors list \
	--partner "Trucker|$CREATED_TRUCKER_ID" \
	--partner-type "Trucker" \
	--partner-id "$CREATED_TRUCKER_ID" \
	--limit 5
assert_success

test_name "List broker vendors with not-partner-type filter fails (server 500)"
xbe_json view broker-vendors list --not-partner-type "Customer" --limit 5
assert_failure

test_name "List broker vendors with trading-partner-type filter"
xbe_json view broker-vendors list --trading-partner-type "BrokerVendor" --limit 5
assert_success

test_name "List broker vendors with external accounting ID filter"
xbe_json view broker-vendors list --external-accounting-broker-vendor-id "$UPDATED_EXTERNAL_ID" --limit 5
assert_success

test_name "List broker vendors with external-identification-value filter"
xbe_json view broker-vendors list --external-identification-value "EXT123" --limit 5
assert_success

test_name "List broker vendors with created/updated filters"
xbe_json view broker-vendors list \
	--created-at-min "2020-01-01T00:00:00Z" \
	--created-at-max "2030-01-01T00:00:00Z" \
	--updated-at-min "2020-01-01T00:00:00Z" \
	--updated-at-max "2030-01-01T00:00:00Z" \
	--limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update broker vendor without fields fails"
xbe_json do broker-vendors update "$CREATED_BROKER_VENDOR_ID"
assert_failure

test_name "Delete broker vendor without confirm fails"
xbe_run do broker-vendors delete "$CREATED_BROKER_VENDOR_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker vendor"
xbe_run do broker-vendors delete "$CREATED_BROKER_VENDOR_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
