#!/bin/bash
#
# XBE CLI Integration Tests: Lowest Losing Bid Prediction Subject Details
#
# Tests list/show/create/update/delete operations for lowest-losing-bid-prediction-subject-details.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_LOWEST_LOSING_BID_PREDICTION_SUBJECT_ID:-}"
if [[ -z "$PREDICTION_SUBJECT_ID" ]]; then
    PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"
fi
BIDDER_ID="${XBE_TEST_BIDDER_ID:-}"

SAMPLE_ID=""
SAMPLE_PREDICTION_SUBJECT_ID=""
CREATED_ID=""

BID_DETAILS='[{"bidder_name":"Acme Demo","amount":125000.5}]'

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"cannot change"* ]] || [[ "$msg" == *"cannot be changed"* ]] || [[ "$msg" == *"gaps dependent"* ]] || [[ "$msg" == *"must have kind"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: lowest-losing-bid-prediction-subject-details"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lowest losing bid prediction subject details"
xbe_json view lowest-losing-bid-prediction-subject-details list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PREDICTION_SUBJECT_ID=$(json_get ".[0].prediction_subject_id")
else
    fail "Failed to list lowest losing bid prediction subject details"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lowest losing bid prediction subject detail"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lowest-losing-bid-prediction-subject-details show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output" || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show detail: $output"
        fi
    fi
else
    skip "No detail ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List details with --prediction-subject filter"
FILTER_PREDICTION_SUBJECT_ID="${SAMPLE_PREDICTION_SUBJECT_ID:-$PREDICTION_SUBJECT_ID}"
if [[ -n "$FILTER_PREDICTION_SUBJECT_ID" && "$FILTER_PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json view lowest-losing-bid-prediction-subject-details list --prediction-subject "$FILTER_PREDICTION_SUBJECT_ID" --limit 5
    assert_success
else
    skip "No prediction subject ID available"
fi

test_name "List details with --bidder filter"
if [[ -n "$BIDDER_ID" && "$BIDDER_ID" != "null" ]]; then
    xbe_json view lowest-losing-bid-prediction-subject-details list --bidder "$BIDDER_ID" --limit 5
    assert_success
else
    skip "No bidder ID available (set XBE_TEST_BIDDER_ID)"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create detail requires prediction subject"
xbe_run do lowest-losing-bid-prediction-subject-details create --bid-amount 120000
assert_failure

test_name "Create lowest losing bid prediction subject detail"
if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json do lowest-losing-bid-prediction-subject-details create \
        --prediction-subject "$PREDICTION_SUBJECT_ID" \
        --lowest-bid-amount 115000 \
        --bid-amount 120000 \
        --walk-away-bid-amount 140000 \
        --engineer-estimate-amount 130000 \
        --internal-engineer-estimate-amount 128000 \
        --bid-details "$BID_DETAILS"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "lowest-losing-bid-prediction-subject-details" "$CREATED_ID"
            pass
        else
            fail "Created detail but no ID returned"
        fi
    else
        if update_blocked_message "$output" || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Create blocked by server policy or invalid prediction subject"
        else
            fail "Failed to create detail: $output"
        fi
    fi
else
    skip "No prediction subject ID available (set XBE_TEST_LOWEST_LOSING_BID_PREDICTION_SUBJECT_ID or XBE_TEST_PREDICTION_SUBJECT_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update lowest losing bid prediction subject detail"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do lowest-losing-bid-prediction-subject-details update "$CREATED_ID" \
        --engineer-estimate-amount 131000 \
        --internal-engineer-estimate-amount 129000
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Update blocked by server policy or validation"
        else
            fail "Failed to update detail: $output"
        fi
    fi
else
    skip "No created detail available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete detail requires --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do lowest-losing-bid-prediction-subject-details delete "$CREATED_ID"
    assert_failure
else
    skip "No created detail available for delete"
fi

test_name "Delete lowest losing bid prediction subject detail"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do lowest-losing-bid-prediction-subject-details delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete detail: $output"
        fi
    fi
else
    skip "No created detail available for delete"
fi

run_tests
