#!/bin/bash
#
# XBE CLI Integration Tests: Bidders
#
# Tests CRUD operations for the bidders resource.
# Bidders represent entities that submit bids within a broker's workflows.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BIDDER_ID=""
CREATED_BIDDER_FALSE_ID=""
BIDDER_NAME=""
BIDDER_NAME_FALSE=""


describe "Resource: bidders"

# ==========================================================================
# Prerequisites - Create broker
# ==========================================================================

test_name "Create prerequisite broker for bidder tests"
BROKER_NAME=$(unique_name "BidderTestBroker")

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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create bidder with required fields"
BIDDER_NAME=$(unique_name "Bidder")

xbe_json do bidders create \
    --name "$BIDDER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --is-self-for-broker=true

if [[ $status -eq 0 ]]; then
    CREATED_BIDDER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BIDDER_ID" && "$CREATED_BIDDER_ID" != "null" ]]; then
        register_cleanup "bidders" "$CREATED_BIDDER_ID"
        pass
    else
        fail "Created bidder but no ID returned"
    fi
else
    fail "Failed to create bidder"
fi

# Only continue if we successfully created a bidder
if [[ -z "$CREATED_BIDDER_ID" || "$CREATED_BIDDER_ID" == "null" ]]; then
    echo "Cannot continue without a valid bidder ID"
    run_tests
fi

test_name "Create bidder with is-self-for-broker false"
BIDDER_NAME_FALSE=$(unique_name "BidderFalse")

xbe_json do bidders create \
    --name "$BIDDER_NAME_FALSE" \
    --broker "$CREATED_BROKER_ID" \
    --is-self-for-broker=false

if [[ $status -eq 0 ]]; then
    CREATED_BIDDER_FALSE_ID=$(json_get ".id")
    if [[ -n "$CREATED_BIDDER_FALSE_ID" && "$CREATED_BIDDER_FALSE_ID" != "null" ]]; then
        register_cleanup "bidders" "$CREATED_BIDDER_FALSE_ID"
        pass
    else
        fail "Created bidder but no ID returned"
    fi
else
    fail "Failed to create bidder with is-self-for-broker false"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show bidder by ID"
xbe_json view bidders show "$CREATED_BIDDER_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List bidders"
xbe_json view bidders list --limit 5
assert_success

test_name "List bidders returns array"
xbe_json view bidders list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list bidders"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List bidders with --broker filter"
xbe_json view bidders list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List bidders with --name filter"
xbe_json view bidders list --name "$BIDDER_NAME" --limit 10
assert_success

test_name "List bidders with --name-like filter"
NAME_LIKE=${BIDDER_NAME:0:12}
xbe_json view bidders list --name-like "$NAME_LIKE" --limit 10
assert_success

test_name "List bidders with --is-self-for-broker true"
xbe_json view bidders list --is-self-for-broker true --limit 10
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List bidders with --limit"
xbe_json view bidders list --limit 3
assert_success

test_name "List bidders with --offset"
xbe_json view bidders list --limit 3 --offset 3
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update bidder name"
UPDATED_NAME=$(unique_name "UpdatedBidder")
xbe_json do bidders update "$CREATED_BIDDER_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update bidder is-self-for-broker"
xbe_json do bidders update "$CREATED_BIDDER_ID" --is-self-for-broker=false
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete bidder requires --confirm flag"
xbe_run do bidders delete "$CREATED_BIDDER_ID"
assert_failure

test_name "Delete bidder with --confirm"
BIDDER_DELETE_NAME=$(unique_name "BidderDelete")
xbe_json do bidders create \
    --name "$BIDDER_DELETE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --is-self-for-broker=false
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do bidders delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create bidder for deletion test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create bidder without name fails"
xbe_json do bidders create --broker "$CREATED_BROKER_ID" --is-self-for-broker=true
assert_failure

test_name "Create bidder without broker fails"
xbe_json do bidders create --name "NoBroker" --is-self-for-broker=false
assert_failure

test_name "Create bidder without is-self-for-broker fails"
xbe_json do bidders create --name "NoSelf" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do bidders update "$CREATED_BIDDER_ID"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
