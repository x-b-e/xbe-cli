#!/bin/bash
#
# XBE CLI Integration Tests: Commitments
#
# Tests list and show operations for the commitments resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STATUS=""
SAMPLE_BUYER_TYPE=""
SAMPLE_BUYER_ID=""
SAMPLE_SELLER_TYPE=""
SAMPLE_SELLER_ID=""

BROKER_ID=""

describe "Resource: commitments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List commitments"
xbe_json view commitments list --limit 5
assert_success

test_name "List commitments returns array"
xbe_json view commitments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list commitments"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample commitment"
xbe_json view commitments list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_BUYER_TYPE=$(json_get ".[0].buyer_type")
    SAMPLE_BUYER_ID=$(json_get ".[0].buyer_id")
    SAMPLE_SELLER_TYPE=$(json_get ".[0].seller_type")
    SAMPLE_SELLER_ID=$(json_get ".[0].seller_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No commitments available for follow-on tests"
    fi
else
    skip "Could not list commitments to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List commitments with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view commitments list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_BROKER_ID"
elif [[ -n "$SAMPLE_BUYER_TYPE" && "$SAMPLE_BUYER_TYPE" == *"broker"* && -n "$SAMPLE_BUYER_ID" && "$SAMPLE_BUYER_ID" != "null" ]]; then
    BROKER_ID="$SAMPLE_BUYER_ID"
elif [[ -n "$SAMPLE_SELLER_TYPE" && "$SAMPLE_SELLER_TYPE" == *"broker"* && -n "$SAMPLE_SELLER_ID" && "$SAMPLE_SELLER_ID" != "null" ]]; then
    BROKER_ID="$SAMPLE_SELLER_ID"
fi

test_name "List commitments with --broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view commitments list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List commitments with --broker-id filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view commitments list --broker-id "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show commitment"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view commitments show "$SAMPLE_ID"
    assert_success
else
    skip "No commitment ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
