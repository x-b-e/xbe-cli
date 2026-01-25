#!/bin/bash
#
# XBE CLI Integration Tests: Contractors
#
# Tests CRUD operations for the contractors resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CONTRACTOR_ID=""

describe "Resource: contractors"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for contractor tests"
BROKER_NAME=$(unique_name "ContractorBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create contractor with required fields"
CONTRACTOR_NAME=$(unique_name "Contractor")

xbe_json do contractors create --name "$CONTRACTOR_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
	CREATED_CONTRACTOR_ID=$(json_get ".id")
	if [[ -n "$CREATED_CONTRACTOR_ID" && "$CREATED_CONTRACTOR_ID" != "null" ]]; then
		register_cleanup "contractors" "$CREATED_CONTRACTOR_ID"
		pass
	else
		fail "Created contractor but no ID returned"
		run_tests
	fi
else
	fail "Failed to create contractor"
	run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show contractor"
xbe_json view contractors show "$CREATED_CONTRACTOR_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update contractor name"
UPDATED_NAME="${CONTRACTOR_NAME}-Updated"
xbe_json do contractors update "$CREATED_CONTRACTOR_ID" --name "$UPDATED_NAME"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List contractors"
xbe_json view contractors list --limit 5
assert_success

test_name "List contractors with broker filter"
xbe_json view contractors list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List contractors with name filter"
xbe_json view contractors list --name "$UPDATED_NAME" --limit 5
assert_success

test_name "List contractors with incidents filter"
xbe_json view contractors list --incidents 1 --limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update contractor without fields fails"
xbe_json do contractors update "$CREATED_CONTRACTOR_ID"
assert_failure

test_name "Delete contractor without confirm fails"
xbe_run do contractors delete "$CREATED_CONTRACTOR_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete contractor"
xbe_run do contractors delete "$CREATED_CONTRACTOR_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
