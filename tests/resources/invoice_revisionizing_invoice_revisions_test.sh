#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Revisionizing Invoice Revisions
#
# Tests list and show operations for the invoice-revisionizing-invoice-revisions resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_WORK_ID=""
SAMPLE_REVISION_ID=""
SAMPLE_INVOICE_ID=""
LIST_SUPPORTED="true"

describe "Resource: invoice-revisionizing-invoice-revisions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoice revisionizing invoice revisions"
xbe_json view invoice-revisionizing-invoice-revisions list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing invoice revisionizing invoice revisions"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List invoice revisionizing invoice revisions returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list invoice revisionizing invoice revisions"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample invoice revisionizing invoice revision"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_WORK_ID=$(json_get ".[0].invoice_revisionizing_work_id")
        SAMPLE_REVISION_ID=$(json_get ".[0].invoice_revision_id")
        SAMPLE_INVOICE_ID=$(json_get ".[0].invoice_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No invoice revisionizing invoice revisions available for follow-on tests"
        fi
    else
        skip "Could not list invoice revisionizing invoice revisions to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List invoice revisionizing invoice revisions with --invoice-revisionizing-work"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    if [[ -n "$SAMPLE_WORK_ID" && "$SAMPLE_WORK_ID" != "null" ]]; then
        xbe_json view invoice-revisionizing-invoice-revisions list --invoice-revisionizing-work "$SAMPLE_WORK_ID" --limit 5
        assert_success
    else
        skip "No invoice revisionizing work ID available"
    fi
else
    skip "List endpoint not supported"
fi

test_name "List invoice revisionizing invoice revisions with --invoice-revision"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    if [[ -n "$SAMPLE_REVISION_ID" && "$SAMPLE_REVISION_ID" != "null" ]]; then
        xbe_json view invoice-revisionizing-invoice-revisions list --invoice-revision "$SAMPLE_REVISION_ID" --limit 5
        assert_success
    else
        skip "No invoice revision ID available"
    fi
else
    skip "List endpoint not supported"
fi

test_name "List invoice revisionizing invoice revisions with --invoice"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    if [[ -n "$SAMPLE_INVOICE_ID" && "$SAMPLE_INVOICE_ID" != "null" ]]; then
        xbe_json view invoice-revisionizing-invoice-revisions list --invoice "$SAMPLE_INVOICE_ID" --limit 5
        assert_success
    else
        skip "No invoice ID available"
    fi
else
    skip "List endpoint not supported"
fi

test_name "List invoice revisionizing invoice revisions with --created-at-min"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions list --created-at-min "2020-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

test_name "List invoice revisionizing invoice revisions with --created-at-max"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions list --created-at-max "2030-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

test_name "List invoice revisionizing invoice revisions with --updated-at-min"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

test_name "List invoice revisionizing invoice revisions with --updated-at-max"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show invoice revisionizing invoice revision"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view invoice-revisionizing-invoice-revisions show "$SAMPLE_ID"
    assert_success
else
    skip "No invoice revisionizing invoice revision ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
