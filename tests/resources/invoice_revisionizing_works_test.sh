#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Revisionizing Works
#
# Tests view operations for the invoice-revisionizing-works resource.
# These records track bulk invoice revisionizing requests.
#
# COVERAGE: List + filters + show (when available)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: invoice-revisionizing-works (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoice revisionizing works"
xbe_json view invoice-revisionizing-works list
assert_success

test_name "List invoice revisionizing works returns array"
xbe_json view invoice-revisionizing-works list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list invoice revisionizing works"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List invoice revisionizing works with --broker filter"
xbe_json view invoice-revisionizing-works list --broker "1"
assert_success

test_name "List invoice revisionizing works with --created-by filter"
xbe_json view invoice-revisionizing-works list --created-by "1"
assert_success

test_name "List invoice revisionizing works with --organization filter"
xbe_json view invoice-revisionizing-works list --organization-type "Broker" --organization-id "1"
assert_success

test_name "List invoice revisionizing works with --jid filter"
xbe_json view invoice-revisionizing-works list --jid "12345"
assert_success

# ============================================================================
# SHOW Test (if available)
# ============================================================================

test_name "Show invoice revisionizing work when available"
xbe_json view invoice-revisionizing-works list --limit 1
if [[ $status -eq 0 ]]; then
    WORK_ID=$(json_get '.[0].id')
    if [[ -n "$WORK_ID" && "$WORK_ID" != "null" ]]; then
        xbe_json view invoice-revisionizing-works show "$WORK_ID"
        assert_success
    else
        skip "No invoice revisionizing work records available"
    fi
else
    fail "Failed to list invoice revisionizing works for show test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
