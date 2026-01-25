#!/bin/bash
#
# XBE CLI Integration Tests: Trading Partners
#
# Tests list/show and create/delete behavior for trading-partners.
#
# COVERAGE: List filters + show + create attributes + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: trading-partners"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRADING_PARTNER_ID=""

# ============================================================================
# Prerequisites - Create broker and customer
# ============================================================================

test_name "Create prerequisite broker for trading partner tests"
BROKER_NAME=$(unique_name "TradingPartnerBroker")

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

test_name "Create prerequisite customer for trading partner tests"
CUSTOMER_NAME=$(unique_name "TradingPartnerCustomer")

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

test_name "Create trading partner without required fields fails"
xbe_json do trading-partners create
assert_failure

test_name "Create trading partner"
xbe_json do trading-partners create \
	--organization "Broker|$CREATED_BROKER_ID" \
	--partner "Customer|$CREATED_CUSTOMER_ID" \
	--trading-partner-type "BrokerCustomer"

if [[ $status -eq 0 ]]; then
	CREATED_TRADING_PARTNER_ID=$(json_get ".id")
	if [[ -n "$CREATED_TRADING_PARTNER_ID" && "$CREATED_TRADING_PARTNER_ID" != "null" ]]; then
		register_cleanup "trading-partners" "$CREATED_TRADING_PARTNER_ID"
		pass
	else
		fail "Created trading partner but no ID returned"
		run_tests
	fi
else
	fail "Failed to create trading partner"
	run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trading partner"
xbe_json view trading-partners show "$CREATED_TRADING_PARTNER_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List trading partners"
xbe_json view trading-partners list --limit 5
assert_success

test_name "List trading partners with organization filters"
xbe_json view trading-partners list \
	--organization "Broker|$CREATED_BROKER_ID" \
	--organization-type "Broker" \
	--organization-id "$CREATED_BROKER_ID" \
	--limit 5
assert_success

test_name "List trading partners with not-organization-type filter fails (server 500)"
xbe_json view trading-partners list --not-organization-type "Customer" --limit 5
assert_failure

test_name "List trading partners with partner filters"
xbe_json view trading-partners list \
	--partner "Customer|$CREATED_CUSTOMER_ID" \
	--partner-type "Customer" \
	--partner-id "$CREATED_CUSTOMER_ID" \
	--limit 5
assert_success

test_name "List trading partners with not-partner-type filter fails (server 500)"
xbe_json view trading-partners list --not-partner-type "Broker" --limit 5
assert_failure

test_name "List trading partners with trading-partner-type filter"
xbe_json view trading-partners list --trading-partner-type "BrokerCustomer" --limit 5
assert_success

test_name "List trading partners with external-identification-value filter"
xbe_json view trading-partners list --external-identification-value "EXT123" --limit 5
assert_success

test_name "List trading partners with created/updated filters"
xbe_json view trading-partners list \
	--created-at-min "2020-01-01T00:00:00Z" \
	--created-at-max "2030-01-01T00:00:00Z" \
	--updated-at-min "2020-01-01T00:00:00Z" \
	--updated-at-max "2030-01-01T00:00:00Z" \
	--limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Delete trading partner without confirm fails"
xbe_run do trading-partners delete "$CREATED_TRADING_PARTNER_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete trading partner"
xbe_run do trading-partners delete "$CREATED_TRADING_PARTNER_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
