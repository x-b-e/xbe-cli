#!/bin/bash
#
# XBE CLI Integration Tests: Project Status Changes
#
# Tests list and show operations for the project-status-changes resource.
# Project status changes capture status history for projects.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PROJECT_STATUS_CHANGE_ID=""
PROJECT_ID=""
STATUS=""
SKIP_ID_FILTERS=0

describe "Resource: project-status-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project status changes"
xbe_json view project-status-changes list --limit 5
assert_success

test_name "List project status changes returns array"
xbe_json view project-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project status changes"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project status change"
xbe_json view project-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    PROJECT_STATUS_CHANGE_ID=$(json_get ".[0].id")
    PROJECT_ID=$(json_get ".[0].project_id")
    STATUS=$(json_get ".[0].status")
    if [[ -n "$PROJECT_STATUS_CHANGE_ID" && "$PROJECT_STATUS_CHANGE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No project status changes available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list project status changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project status changes with --project filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view project-status-changes list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List project status changes with --status filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$STATUS" && "$STATUS" != "null" ]]; then
    xbe_json view project-status-changes list --status "$STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project status change"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$PROJECT_STATUS_CHANGE_ID" && "$PROJECT_STATUS_CHANGE_ID" != "null" ]]; then
    xbe_json view project-status-changes show "$PROJECT_STATUS_CHANGE_ID"
    assert_success
else
    skip "No project status change ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
