#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Duplication Works
#
# Tests view operations for the job-production-plan-duplication-works resource.
# These records track async job production plan duplication work.
#
# COVERAGE: List + filters + show (when available)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: job-production-plan-duplication-works (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan duplication works"
xbe_json view job-production-plan-duplication-works list
assert_success

test_name "List job production plan duplication works returns array"
xbe_json view job-production-plan-duplication-works list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan duplication works"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List duplication works with --processed-at filter"
xbe_json view job-production-plan-duplication-works list --processed-at "2025-01-01T00:00:00Z"
assert_success

test_name "List duplication works with --job-production-plan-template-id filter"
xbe_json view job-production-plan-duplication-works list --job-production-plan-template-id "1"
assert_success

# ============================================================================
# SHOW Test (if available)
# ============================================================================

test_name "Show job production plan duplication work when available"
xbe_json view job-production-plan-duplication-works list --limit 1
if [[ $status -eq 0 ]]; then
    WORK_ID=$(json_get '.[0].id')
    if [[ -n "$WORK_ID" && "$WORK_ID" != "null" ]]; then
        xbe_json view job-production-plan-duplication-works show "$WORK_ID"
        assert_success
    else
        skip "No duplication work records available"
    fi
else
    fail "Failed to list duplication works for show test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
