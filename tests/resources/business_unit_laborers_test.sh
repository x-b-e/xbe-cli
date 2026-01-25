#!/bin/bash
#
# XBE CLI Integration Tests: Business Unit Laborers
#
# Tests list/show and create/delete behavior for business-unit-laborers.
#
# COVERAGE: List filters + show + create attributes + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: business-unit-laborers"

CREATED_BROKER_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_CUSTOMER_ID=""
CREATED_LABOR_CLASSIFICATION_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_LABORER_ID=""
CREATED_LINK_ID=""

# ============================================================================
# Prerequisites - Create broker, business unit, customer, labor classification, user, and laborer
# ============================================================================

test_name "Create prerequisite broker for business unit laborer tests"
BROKER_NAME=$(unique_name "BusinessUnitLaborerBroker")

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
	if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
		CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
		echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
		pass
	else
		fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
		run_tests
	fi
fi

test_name "Create prerequisite business unit for business unit laborer tests"
BUSINESS_UNIT_NAME=$(unique_name "BusinessUnitLaborerUnit")

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

test_name "Create prerequisite customer for business unit laborer tests"
CUSTOMER_NAME=$(unique_name "BusinessUnitLaborerCustomer")

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

test_name "Create prerequisite labor classification"
LC_NAME=$(unique_name "BusinessUnitLaborerClass")
LC_ABBR="LC$(date +%s | tail -c 4)"

xbe_json do labor-classifications create \
	--name "$LC_NAME" \
	--abbreviation "$LC_ABBR"

if [[ $status -eq 0 ]]; then
	CREATED_LABOR_CLASSIFICATION_ID=$(json_get ".id")
	if [[ -n "$CREATED_LABOR_CLASSIFICATION_ID" && "$CREATED_LABOR_CLASSIFICATION_ID" != "null" ]]; then
		register_cleanup "labor-classifications" "$CREATED_LABOR_CLASSIFICATION_ID"
		pass
	else
		fail "Created labor classification but no ID returned"
		run_tests
	fi
else
	fail "Failed to create labor classification"
	run_tests
fi

test_name "Create prerequisite user for laborer"
USER_EMAIL=$(unique_email)
USER_NAME=$(unique_name "BusinessUnitLaborerUser")

xbe_json do users create \
	--email "$USER_EMAIL" \
	--name "$USER_NAME"

if [[ $status -eq 0 ]]; then
	CREATED_USER_ID=$(json_get ".id")
	if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
		pass
	else
		fail "Created user but no ID returned"
		run_tests
	fi
else
	fail "Failed to create user"
	run_tests
fi

test_name "Create membership for user to customer"
xbe_json do memberships create \
	--user "$CREATED_USER_ID" \
	--organization "Customer|$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
	CREATED_MEMBERSHIP_ID=$(json_get ".id")
	if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
		register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
		pass
	else
		fail "Created membership but no ID returned"
		run_tests
	fi
else
	fail "Failed to create membership"
	run_tests
fi

test_name "Create prerequisite laborer"

xbe_json do laborers create \
	--labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
	--user "$CREATED_USER_ID" \
	--organization-type "customers" \
	--organization-id "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
	CREATED_LABORER_ID=$(json_get ".id")
	if [[ -n "$CREATED_LABORER_ID" && "$CREATED_LABORER_ID" != "null" ]]; then
		register_cleanup "laborers" "$CREATED_LABORER_ID"
		pass
	else
		fail "Created laborer but no ID returned"
		run_tests
	fi
else
	fail "Failed to create laborer"
	run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create business unit laborer without required fields fails"
xbe_json do business-unit-laborers create
assert_failure

test_name "Create business unit laborer"
xbe_json do business-unit-laborers create \
	--business-unit "$CREATED_BUSINESS_UNIT_ID" \
	--laborer "$CREATED_LABORER_ID"

if [[ $status -eq 0 ]]; then
	CREATED_LINK_ID=$(json_get ".id")
	if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
		register_cleanup "business-unit-laborers" "$CREATED_LINK_ID"
		pass
	else
		fail "Created business unit laborer but no ID returned"
		run_tests
	fi
else
	fail "Failed to create business unit laborer"
	run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show business unit laborer"
xbe_json view business-unit-laborers show "$CREATED_LINK_ID"
assert_success

# ==========================================================================
# LIST Tests
# ==========================================================================

test_name "List business unit laborers"
xbe_json view business-unit-laborers list --limit 5
assert_success

test_name "List business unit laborers with business unit filter"
xbe_json view business-unit-laborers list --business-unit "$CREATED_BUSINESS_UNIT_ID" --limit 5
assert_success

test_name "List business unit laborers with laborer filter"
xbe_json view business-unit-laborers list --laborer "$CREATED_LABORER_ID" --limit 5
assert_success

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Delete business unit laborer without confirm fails"
xbe_run do business-unit-laborers delete "$CREATED_LINK_ID"
assert_failure

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete business unit laborer"
xbe_run do business-unit-laborers delete "$CREATED_LINK_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
