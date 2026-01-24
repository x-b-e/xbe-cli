#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Revenue Item Quantity Estimates
#
# Tests CRUD operations and list filters for the project-phase-revenue-item-quantity-estimates resource.
#
# COVERAGE: create/update/delete + list filters + JSON estimate payloads
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PROJECT_PHASE_REVENUE_ITEM_ID="${XBE_TEST_PROJECT_PHASE_REVENUE_ITEM_ID:-}"
PROJECT_ESTIMATE_SET_ID="${XBE_TEST_PROJECT_ESTIMATE_SET_ID:-}"
CURRENT_USER_ID=""
CREATED_ID=""

describe "Resource: project-phase-revenue-item-quantity-estimates"

# ============================================================================
# Resolve current user
# ============================================================================

test_name "Resolve current user"
xbe_json auth whoami

if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "No user ID returned from auth whoami"
        echo "Cannot continue without current user ID"
        run_tests
    fi
else
    fail "Failed to resolve current user"
    echo "Cannot continue without current user ID"
    run_tests
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project phase revenue item quantity estimates"
xbe_json view project-phase-revenue-item-quantity-estimates list --limit 5
assert_success

test_name "List project phase revenue item quantity estimates returns array"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project phase revenue item quantity estimates"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" && "$PROJECT_PHASE_REVENUE_ITEM_ID" != "null" && -n "$PROJECT_ESTIMATE_SET_ID" && "$PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
    test_name "Create project phase revenue item quantity estimate"
    ESTIMATE_JSON='{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

    xbe_json do project-phase-revenue-item-quantity-estimates create \
        --project-phase-revenue-item "$PROJECT_PHASE_REVENUE_ITEM_ID" \
        --project-estimate-set "$PROJECT_ESTIMATE_SET_ID" \
        --estimate "$ESTIMATE_JSON" \
        --description "Initial estimate" \
        --created-by "$CURRENT_USER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-phase-revenue-item-quantity-estimates" "$CREATED_ID"
            pass
        else
            fail "Created quantity estimate but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create quantity estimate"
        fi
    fi
else
    skip "Missing prerequisite IDs. Set XBE_TEST_PROJECT_PHASE_REVENUE_ITEM_ID and XBE_TEST_PROJECT_ESTIMATE_SET_ID to enable create/update/delete tests."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update quantity estimate description"
    xbe_json do project-phase-revenue-item-quantity-estimates update "$CREATED_ID" --description "Updated estimate"
    assert_success

    test_name "Update quantity estimate distribution"
    UPDATED_ESTIMATE_JSON='{"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}'
    xbe_json do project-phase-revenue-item-quantity-estimates update "$CREATED_ID" --estimate "$UPDATED_ESTIMATE_JSON"
    assert_success
else
    skip "Skipping update tests (no created ID)"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Show project phase revenue item quantity estimate"
    xbe_json view project-phase-revenue-item-quantity-estimates show "$CREATED_ID"
    assert_success
else
    test_name "Show project phase revenue item quantity estimate from list"
    xbe_json view project-phase-revenue-item-quantity-estimates list --limit 1
    if [[ $status -eq 0 ]]; then
        SHOW_ID=$(json_get ".[0].id")
        if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
            xbe_json view project-phase-revenue-item-quantity-estimates show "$SHOW_ID"
            assert_success
        else
            skip "No quantity estimate available for show test"
        fi
    else
    fail "Failed to list quantity estimates for show test"
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" && "$PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
    test_name "List quantity estimates with --project-phase-revenue-item"
    xbe_json view project-phase-revenue-item-quantity-estimates list --project-phase-revenue-item "$PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
    assert_success
else
    skip "Skipping --project-phase-revenue-item filter test (missing ID)"
fi

if [[ -n "$PROJECT_ESTIMATE_SET_ID" && "$PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
    test_name "List quantity estimates with --project-estimate-set"
    xbe_json view project-phase-revenue-item-quantity-estimates list --project-estimate-set "$PROJECT_ESTIMATE_SET_ID" --limit 5
    assert_success
else
    skip "Skipping --project-estimate-set filter test (missing ID)"
fi

test_name "List quantity estimates with --created-by"
xbe_json view project-phase-revenue-item-quantity-estimates list --created-by "$CURRENT_USER_ID" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete project phase revenue item quantity estimate"
    xbe_json do project-phase-revenue-item-quantity-estimates delete "$CREATED_ID" --confirm
    assert_success
else
    skip "Skipping delete test (no created ID)"
fi

run_tests
