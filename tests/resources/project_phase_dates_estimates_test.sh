#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Dates Estimates
#
# Tests CRUD operations and list filters for the project-phase-dates-estimates resource.
#
# COVERAGE: create/update/delete + list filters + date attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PROJECT_PHASE_ID="${XBE_TEST_PROJECT_PHASE_ID:-}"
PROJECT_ESTIMATE_SET_ID="${XBE_TEST_PROJECT_ESTIMATE_SET_ID:-}"
PROJECT_ID="${XBE_TEST_PROJECT_ID:-}"
CURRENT_USER_ID=""
CREATED_ID=""

describe "Resource: project-phase-dates-estimates"

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

test_name "List project phase dates estimates"
xbe_json view project-phase-dates-estimates list --limit 5
assert_success

test_name "List project phase dates estimates returns array"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project phase dates estimates"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$PROJECT_PHASE_ID" && "$PROJECT_PHASE_ID" != "null" && -n "$PROJECT_ESTIMATE_SET_ID" && "$PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
    test_name "Create project phase dates estimate"
    START_DATE="2025-01-01"
    END_DATE="2025-01-15"

    xbe_json do project-phase-dates-estimates create \
        --project-phase "$PROJECT_PHASE_ID" \
        --project-estimate-set "$PROJECT_ESTIMATE_SET_ID" \
        --start-date "$START_DATE" \
        --end-date "$END_DATE" \
        --created-by "$CURRENT_USER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-phase-dates-estimates" "$CREATED_ID"
            pass
        else
            fail "Created dates estimate but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create dates estimate"
        fi
    fi
else
    skip "Missing prerequisite IDs. Set XBE_TEST_PROJECT_PHASE_ID and XBE_TEST_PROJECT_ESTIMATE_SET_ID to enable create/update/delete tests."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update dates estimate end date"
    xbe_json do project-phase-dates-estimates update "$CREATED_ID" --end-date 2025-02-01
    assert_success

    test_name "Update dates estimate start date"
    xbe_json do project-phase-dates-estimates update "$CREATED_ID" --start-date 2025-01-05
    assert_success
else
    skip "Skipping update tests (no created ID)"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Show project phase dates estimate"
    xbe_json view project-phase-dates-estimates show "$CREATED_ID"
    assert_success
else
    test_name "Show project phase dates estimate from list"
    xbe_json view project-phase-dates-estimates list --limit 1
    if [[ $status -eq 0 ]]; then
        SHOW_ID=$(json_get ".[0].id")
        if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
            xbe_json view project-phase-dates-estimates show "$SHOW_ID"
            assert_success
        else
            skip "No dates estimate available for show test"
        fi
    else
        fail "Failed to list dates estimates for show test"
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

if [[ -n "$PROJECT_PHASE_ID" && "$PROJECT_PHASE_ID" != "null" ]]; then
    test_name "List dates estimates with --project-phase"
    xbe_json view project-phase-dates-estimates list --project-phase "$PROJECT_PHASE_ID" --limit 5
    assert_success
else
    skip "Skipping --project-phase filter test (missing ID)"
fi

if [[ -n "$PROJECT_ESTIMATE_SET_ID" && "$PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
    test_name "List dates estimates with --project-estimate-set"
    xbe_json view project-phase-dates-estimates list --project-estimate-set "$PROJECT_ESTIMATE_SET_ID" --limit 5
    assert_success
else
    skip "Skipping --project-estimate-set filter test (missing ID)"
fi

if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    test_name "List dates estimates with --project"
    xbe_json view project-phase-dates-estimates list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "Skipping --project filter test (missing ID)"
fi

test_name "List dates estimates with --created-by"
xbe_json view project-phase-dates-estimates list --created-by "$CURRENT_USER_ID" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete project phase dates estimate"
    xbe_json do project-phase-dates-estimates delete "$CREATED_ID" --confirm
    assert_success
else
    skip "Skipping delete test (no created ID)"
fi

run_tests
