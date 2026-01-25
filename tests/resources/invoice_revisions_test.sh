#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Revisions
#
# Tests list/show operations for invoice-revisions.
#
# COVERAGE: List filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_REVISION=""
SAMPLE_INVOICE_TYPE=""
SAMPLE_INVOICE_ID=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""

describe "Resource: invoice-revisions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoice revisions"
xbe_json view invoice-revisions list --limit 5
assert_success

test_name "List invoice revisions returns array"
xbe_json view invoice-revisions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list invoice revisions"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample invoice revision"
xbe_json view invoice-revisions list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_REVISION=$(json_get ".[0].revision")
    SAMPLE_INVOICE_TYPE=$(json_get ".[0].invoice_type")
    SAMPLE_INVOICE_ID=$(json_get ".[0].invoice_id")
    SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
    SAMPLE_UPDATED_AT=$(json_get ".[0].updated_at")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No invoice revisions available for follow-on tests"
    fi
else
    skip "Could not list invoice revisions to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List invoice revisions with --revision filter"
if [[ -n "$SAMPLE_REVISION" && "$SAMPLE_REVISION" != "null" ]]; then
    xbe_json view invoice-revisions list --revision "$SAMPLE_REVISION" --limit 5
    assert_success
else
    skip "No revision available"
fi

test_name "List invoice revisions with --invoice filter"
if [[ -n "$SAMPLE_INVOICE_ID" && "$SAMPLE_INVOICE_ID" != "null" ]]; then
    if [[ -n "$SAMPLE_INVOICE_TYPE" && "$SAMPLE_INVOICE_TYPE" != "null" ]]; then
        xbe_json view invoice-revisions list --invoice "${SAMPLE_INVOICE_TYPE}|${SAMPLE_INVOICE_ID}" --limit 5
        assert_success
    else
        skip "No invoice type available"
    fi
else
    skip "No invoice ID available"
fi

test_name "List invoice revisions with --invoice-id filter"
if [[ -n "$SAMPLE_INVOICE_ID" && "$SAMPLE_INVOICE_ID" != "null" ]]; then
    xbe_json view invoice-revisions list --invoice-id "$SAMPLE_INVOICE_ID" --limit 5
    assert_success
else
    skip "No invoice ID available"
fi

CREATED_AT_FILTER="$SAMPLE_CREATED_AT"
if [[ -z "$CREATED_AT_FILTER" || "$CREATED_AT_FILTER" == "null" ]]; then
    CREATED_AT_FILTER=$(date -u +%Y-%m-%dT%H:%M:%SZ)
fi

test_name "List invoice revisions with --created-at-min filter"
xbe_json view invoice-revisions list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice revisions with --created-at-max filter"
xbe_json view invoice-revisions list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice revisions with --is-created-at=true filter"
xbe_json view invoice-revisions list --is-created-at true --limit 5
assert_success

test_name "List invoice revisions with --is-created-at=false filter"
xbe_json view invoice-revisions list --is-created-at false --limit 5
assert_success

UPDATED_AT_FILTER="$SAMPLE_UPDATED_AT"
if [[ -z "$UPDATED_AT_FILTER" || "$UPDATED_AT_FILTER" == "null" ]]; then
    UPDATED_AT_FILTER=$(date -u +%Y-%m-%dT%H:%M:%SZ)
fi

test_name "List invoice revisions with --updated-at-min filter"
xbe_json view invoice-revisions list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice revisions with --updated-at-max filter"
xbe_json view invoice-revisions list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice revisions with --is-updated-at=true filter"
xbe_json view invoice-revisions list --is-updated-at true --limit 5
assert_success

test_name "List invoice revisions with --is-updated-at=false filter"
xbe_json view invoice-revisions list --is-updated-at false --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show invoice revision"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view invoice-revisions show "$SAMPLE_ID"
    assert_success
else
    skip "No invoice revision available to show"
fi

run_tests
