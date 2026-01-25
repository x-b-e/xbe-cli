#!/bin/bash
#
# XBE CLI Integration Tests: Rate Adjustments
#
# Tests CRUD operations for the rate_adjustments resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_COST_INDEX_ID=""
CREATED_COST_INDEX_ID_2=""
CREATED_RATE_ADJUSTMENT_ID=""
RATE_ID=""

TEST_ZERO_INTERCEPT_VALUE="100"
TEST_ZERO_INTERCEPT_RATIO="0.25"
UPDATED_ZERO_INTERCEPT_VALUE="110"
UPDATED_ZERO_INTERCEPT_RATIO="0.30"
TEST_ADJUSTMENT_MIN="1.00"
TEST_ADJUSTMENT_MAX="5.00"
UPDATED_ADJUSTMENT_MIN="2.00"
UPDATED_ADJUSTMENT_MAX="6.00"

describe "Resource: rate_adjustments"

# ============================================================================
# Prerequisites - Cost indexes
# ============================================================================

test_name "Create cost index for rate adjustment tests"
COST_INDEX_NAME=$(unique_name "RateAdjustmentIndex")

xbe_json do cost-indexes create --name "$COST_INDEX_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_COST_INDEX_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_INDEX_ID" && "$CREATED_COST_INDEX_ID" != "null" ]]; then
        register_cleanup "cost-indexes" "$CREATED_COST_INDEX_ID"
        pass
    else
        fail "Created cost index but no ID returned"
    fi
else
    test_name "Fallback to existing cost index"
    xbe_json view cost-indexes list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATED_COST_INDEX_ID=$(json_get ".[0].id")
        if [[ -n "$CREATED_COST_INDEX_ID" && "$CREATED_COST_INDEX_ID" != "null" ]]; then
            pass
        else
            fail "No cost index available for tests"
        fi
    else
        fail "Failed to list cost indexes"
    fi
fi

if [[ -z "$CREATED_COST_INDEX_ID" || "$CREATED_COST_INDEX_ID" == "null" ]]; then
    echo "Cannot continue without a cost index ID"
    run_tests
fi

test_name "Create second cost index for update tests"
COST_INDEX_NAME_2=$(unique_name "RateAdjustmentIndex2")

xbe_json do cost-indexes create --name "$COST_INDEX_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_COST_INDEX_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_COST_INDEX_ID_2" && "$CREATED_COST_INDEX_ID_2" != "null" ]]; then
        register_cleanup "cost-indexes" "$CREATED_COST_INDEX_ID_2"
        pass
    else
        fail "Created second cost index but no ID returned"
    fi
else
    skip "Failed to create second cost index; cost-index update test will be skipped"
fi

# ============================================================================
# Prerequisites - Rate
# ============================================================================

test_name "Capture sample rate for rate adjustment tests"
xbe_json view rates list --limit 1

if [[ $status -eq 0 ]]; then
    RATE_ID=$(json_get ".[0].id")
    if [[ -n "$RATE_ID" && "$RATE_ID" != "null" ]]; then
        pass
    else
        if [[ -n "$XBE_TEST_RATE_ID" ]]; then
            RATE_ID="$XBE_TEST_RATE_ID"
            echo "    Using XBE_TEST_RATE_ID: $RATE_ID"
            pass
        else
            skip "No rate available for create/update tests"
        fi
    fi
else
    if [[ -n "$XBE_TEST_RATE_ID" ]]; then
        RATE_ID="$XBE_TEST_RATE_ID"
        echo "    Using XBE_TEST_RATE_ID: $RATE_ID"
        pass
    else
        skip "Failed to list rates and XBE_TEST_RATE_ID not set"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create rate adjustment with required fields"
if [[ -n "$RATE_ID" && "$RATE_ID" != "null" ]]; then
    xbe_json do rate-adjustments create \
        --rate "$RATE_ID" \
        --cost-index "$CREATED_COST_INDEX_ID" \
        --zero-intercept-value "$TEST_ZERO_INTERCEPT_VALUE" \
        --zero-intercept-ratio "$TEST_ZERO_INTERCEPT_RATIO" \
        --adjustment-min "$TEST_ADJUSTMENT_MIN" \
        --adjustment-max "$TEST_ADJUSTMENT_MAX" \
        --prevent-rating-when-index-value-missing

    if [[ $status -eq 0 ]]; then
        CREATED_RATE_ADJUSTMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_RATE_ADJUSTMENT_ID" && "$CREATED_RATE_ADJUSTMENT_ID" != "null" ]]; then
            register_cleanup "rate-adjustments" "$CREATED_RATE_ADJUSTMENT_ID"
            pass
        else
            fail "Created rate adjustment but no ID returned"
        fi
    else
        fail "Failed to create rate adjustment"
    fi
else
    skip "No rate ID available for create tests"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update zero intercept values"
if [[ -n "$CREATED_RATE_ADJUSTMENT_ID" && "$CREATED_RATE_ADJUSTMENT_ID" != "null" ]]; then
    xbe_json do rate-adjustments update "$CREATED_RATE_ADJUSTMENT_ID" \
        --zero-intercept-value "$UPDATED_ZERO_INTERCEPT_VALUE" \
        --zero-intercept-ratio "$UPDATED_ZERO_INTERCEPT_RATIO"
    assert_success
else
    skip "No rate adjustment available for update tests"
fi

test_name "Update adjustment bounds"
if [[ -n "$CREATED_RATE_ADJUSTMENT_ID" && "$CREATED_RATE_ADJUSTMENT_ID" != "null" ]]; then
    xbe_json do rate-adjustments update "$CREATED_RATE_ADJUSTMENT_ID" \
        --adjustment-min "$UPDATED_ADJUSTMENT_MIN" \
        --adjustment-max "$UPDATED_ADJUSTMENT_MAX"
    assert_success
else
    skip "No rate adjustment available for update tests"
fi

test_name "Update prevent rating flag"
if [[ -n "$CREATED_RATE_ADJUSTMENT_ID" && "$CREATED_RATE_ADJUSTMENT_ID" != "null" ]]; then
    xbe_json do rate-adjustments update "$CREATED_RATE_ADJUSTMENT_ID" \
        --prevent-rating-when-index-value-missing=false
    assert_success
else
    skip "No rate adjustment available for update tests"
fi

test_name "Update cost index relationship"
if [[ -n "$CREATED_RATE_ADJUSTMENT_ID" && "$CREATED_RATE_ADJUSTMENT_ID" != "null" && -n "$CREATED_COST_INDEX_ID_2" && "$CREATED_COST_INDEX_ID_2" != "null" ]]; then
    xbe_json do rate-adjustments update "$CREATED_RATE_ADJUSTMENT_ID" \
        --cost-index "$CREATED_COST_INDEX_ID_2"
    assert_success
else
    skip "Missing rate adjustment or second cost index for relationship update"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show rate adjustment"
if [[ -n "$CREATED_RATE_ADJUSTMENT_ID" && "$CREATED_RATE_ADJUSTMENT_ID" != "null" ]]; then
    xbe_json view rate-adjustments show "$CREATED_RATE_ADJUSTMENT_ID"
    assert_success
else
    skip "No rate adjustment available for show test"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List rate adjustments"
xbe_json view rate-adjustments list --limit 5
assert_success

test_name "List rate adjustments returns array"
xbe_json view rate-adjustments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list rate adjustments"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List rate adjustments with --rate filter"
if [[ -n "$RATE_ID" && "$RATE_ID" != "null" ]]; then
    xbe_json view rate-adjustments list --rate "$RATE_ID" --limit 5
    assert_success
else
    skip "No rate ID available for filter test"
fi

test_name "List rate adjustments with --cost-index filter"
if [[ -n "$CREATED_COST_INDEX_ID" && "$CREATED_COST_INDEX_ID" != "null" ]]; then
    xbe_json view rate-adjustments list --cost-index "$CREATED_COST_INDEX_ID" --limit 5
    assert_success
else
    skip "No cost index ID available for filter test"
fi

test_name "List rate adjustments with --parent-rate-adjustment filter"
PARENT_RATE_ADJUSTMENT_ID=""
xbe_json view rate-adjustments list --limit 5
if [[ $status -eq 0 ]]; then
    PARENT_RATE_ADJUSTMENT_ID=$(json_get ".[0].parent_rate_adjustment_id")
    if [[ -n "$PARENT_RATE_ADJUSTMENT_ID" && "$PARENT_RATE_ADJUSTMENT_ID" != "null" ]]; then
        xbe_json view rate-adjustments list --parent-rate-adjustment "$PARENT_RATE_ADJUSTMENT_ID" --limit 5
        assert_success
    else
        skip "No parent rate adjustment found for filter test"
    fi
else
    skip "Failed to list rate adjustments for parent filter"
fi

test_name "List rate adjustments with --is-parent-rate-adjustment true"
xbe_json view rate-adjustments list --is-parent-rate-adjustment true --limit 5
assert_success

test_name "List rate adjustments with --is-parent-rate-adjustment false"
xbe_json view rate-adjustments list --is-parent-rate-adjustment false --limit 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
