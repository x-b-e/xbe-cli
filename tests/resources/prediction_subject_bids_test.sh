#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subject Bids
#
# Tests CRUD operations for the prediction-subject-bids resource.
# Requires lowest losing bid prediction subject detail and bidder IDs in staging.
#
# COVERAGE: All create/update attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

DETAIL_ID="${XBE_TEST_LOWEST_LOSING_BID_PREDICTION_SUBJECT_DETAIL_ID:-}"
BIDDER_ID="${XBE_TEST_BIDDER_ID:-}"
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"
EXISTING_BID_ID="${XBE_TEST_PREDICTION_SUBJECT_BID_ID:-}"

CREATED_ID=""

describe "Resource: prediction-subject-bids"

# ==========================================================================
# CREATE Tests
# ==========================================================================

if [[ -n "$DETAIL_ID" && -n "$BIDDER_ID" ]]; then
    test_name "Create prediction subject bid"
    xbe_json do prediction-subject-bids create \
        --bidder "$BIDDER_ID" \
        --lowest-losing-bid-prediction-subject-detail "$DETAIL_ID" \
        --amount 120000

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "prediction-subject-bids" "$CREATED_ID"
            pass
        else
            fail "Created prediction subject bid but no ID returned"
        fi
    else
        fail "Failed to create prediction subject bid"
    fi
else
    test_name "Skip create tests (missing prediction subject bid prerequisites)"
    skip "Set XBE_TEST_LOWEST_LOSING_BID_PREDICTION_SUBJECT_DETAIL_ID and XBE_TEST_BIDDER_ID to run create tests"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Show prediction subject bid by ID"
    xbe_json view prediction-subject-bids show "$CREATED_ID"
    assert_success
elif [[ -n "$EXISTING_BID_ID" ]]; then
    test_name "Show prediction subject bid by existing ID"
    xbe_json view prediction-subject-bids show "$EXISTING_BID_ID"
    assert_success
else
    test_name "Skip show tests (missing prediction subject bid ID)"
    skip "Set XBE_TEST_PREDICTION_SUBJECT_BID_ID or enable create tests"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List prediction subject bids"
xbe_json view prediction-subject-bids list --limit 5
assert_success

test_name "List prediction subject bids returns array"
xbe_json view prediction-subject-bids list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list prediction subject bids"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

if [[ -n "$BIDDER_ID" ]]; then
    test_name "List prediction subject bids with --bidder filter"
    xbe_json view prediction-subject-bids list --bidder "$BIDDER_ID" --limit 10
    assert_success
else
    test_name "Skip --bidder filter (missing bidder ID)"
    skip "Set XBE_TEST_BIDDER_ID to test bidder filter"
fi

if [[ -n "$DETAIL_ID" ]]; then
    test_name "List prediction subject bids with --lowest-losing-bid-prediction-subject-detail filter"
    xbe_json view prediction-subject-bids list --lowest-losing-bid-prediction-subject-detail "$DETAIL_ID" --limit 10
    assert_success
else
    test_name "Skip detail filter (missing detail ID)"
    skip "Set XBE_TEST_LOWEST_LOSING_BID_PREDICTION_SUBJECT_DETAIL_ID to test detail filter"
fi

if [[ -n "$BROKER_ID" ]]; then
    test_name "List prediction subject bids with --broker filter"
    xbe_json view prediction-subject-bids list --broker "$BROKER_ID" --limit 10
    assert_success
else
    test_name "Skip --broker filter (missing broker ID)"
    skip "Set XBE_TEST_BROKER_ID to test broker filter"
fi

if [[ -n "$PREDICTION_SUBJECT_ID" ]]; then
    test_name "List prediction subject bids with --prediction-subject filter"
    xbe_json view prediction-subject-bids list --prediction-subject "$PREDICTION_SUBJECT_ID" --limit 10
    assert_success
else
    test_name "Skip --prediction-subject filter (missing prediction subject ID)"
    skip "Set XBE_TEST_PREDICTION_SUBJECT_ID to test prediction subject filter"
fi

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List prediction subject bids with --limit"
xbe_json view prediction-subject-bids list --limit 3
assert_success

test_name "List prediction subject bids with --offset"
xbe_json view prediction-subject-bids list --limit 3 --offset 3
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update prediction subject bid amount"
    xbe_json do prediction-subject-bids update "$CREATED_ID" --amount 125000
    assert_success
else
    test_name "Skip update tests (missing created bid)"
    skip "Enable create tests to run update"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete prediction subject bid requires --confirm flag"
    xbe_run do prediction-subject-bids delete "$CREATED_ID"
    assert_failure

    test_name "Delete prediction subject bid with --confirm"
    xbe_run do prediction-subject-bids delete "$CREATED_ID" --confirm
    assert_success
else
    test_name "Skip delete tests (missing created bid)"
    skip "Enable create tests to run delete"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create prediction subject bid without bidder fails"
xbe_run do prediction-subject-bids create --lowest-losing-bid-prediction-subject-detail "1"
assert_failure

test_name "Create prediction subject bid without detail fails"
xbe_run do prediction-subject-bids create --bidder "1"
assert_failure

test_name "Update prediction subject bid without fields fails"
xbe_run do prediction-subject-bids update 1
assert_failure

test_name "Delete prediction subject bid without --confirm fails"
xbe_run do prediction-subject-bids delete 1
assert_failure

run_tests
