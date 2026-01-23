#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Supply/Demand Balances
#
# Tests list and show operations for the job-production-plan-supply-demand-balances resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BALANCE_ID=""
JPP_ID=""
USE_OBSERVED=""
SKIP_ID_FILTERS=0

describe "Resource: job-production-plan-supply-demand-balances"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List supply/demand balances"
xbe_json view job-production-plan-supply-demand-balances list --limit 5
assert_success

test_name "List supply/demand balances returns array"
xbe_json view job-production-plan-supply-demand-balances list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list supply/demand balances"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample supply/demand balance"
xbe_json view job-production-plan-supply-demand-balances list --limit 1
if [[ $status -eq 0 ]]; then
    BALANCE_ID=$(json_get ".[0].id")
    JPP_ID=$(json_get ".[0].job_production_plan_id")
    USE_OBSERVED=$(json_get ".[0].use_observed_supply_parameters")
    if [[ -n "$BALANCE_ID" && "$BALANCE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No supply/demand balances available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list supply/demand balances"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List supply/demand balances with --job-production-plan filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$JPP_ID" && "$JPP_ID" != "null" ]]; then
    xbe_json view job-production-plan-supply-demand-balances list --job-production-plan "$JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List supply/demand balances with --use-observed-supply-parameters filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$USE_OBSERVED" && "$USE_OBSERVED" != "null" ]]; then
    xbe_json view job-production-plan-supply-demand-balances list --use-observed-supply-parameters "$USE_OBSERVED" --limit 5
    assert_success
else
    skip "No observed supply parameter value available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show supply/demand balance"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$BALANCE_ID" && "$BALANCE_ID" != "null" ]]; then
    xbe_json view job-production-plan-supply-demand-balances show "$BALANCE_ID"
    assert_success
else
    skip "No supply/demand balance ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
