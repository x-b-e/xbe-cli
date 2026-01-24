#!/bin/bash
#
# XBE CLI Integration Tests: Tender Returns
#
# Tests list, show, and create operations for the tender-returns resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
LIST_SUPPORTED="false"

TENDER_ID="${XBE_TEST_TENDER_RETURN_TENDER_ID:-}"
TENDER_TYPE="${XBE_TEST_TENDER_RETURN_TENDER_TYPE:-}"

describe "Resource: tender-returns"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tender returns"
xbe_json view tender-returns list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Tender returns list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List tender returns returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view tender-returns list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list tender returns"
    fi
else
    skip "Tender returns list endpoint not available"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample tender return"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view tender-returns list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No tender returns available for show"
        fi
    else
        skip "Could not list tender returns to capture sample"
    fi
else
    skip "Tender returns list endpoint not available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender return"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-returns show "$SAMPLE_ID"
    assert_success
else
    skip "No tender return ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tender return without required flags fails"
xbe_run do tender-returns create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$TENDER_ID" || -z "$TENDER_TYPE" ]]; then
    xbe_json view tender-job-schedule-shifts list --limit 1
    if [[ $status -eq 0 ]]; then
        TENDER_ID=$(json_get ".[0].tender_id")
        TENDER_TYPE=$(json_get ".[0].tender_type")
    fi
fi

test_name "Create tender return"
if [[ -n "$TENDER_ID" && -n "$TENDER_TYPE" && "$TENDER_ID" != "null" && "$TENDER_TYPE" != "null" ]]; then
    COMMENT="Returned via CLI test"

    xbe_json do tender-returns create \
        --tender-type "$TENDER_TYPE" \
        --tender-id "$TENDER_ID" \
        --comment "$COMMENT"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".tender_id" "$TENDER_ID"
        assert_json_equals ".tender_type" "$TENDER_TYPE"
        assert_json_equals ".comment" "$COMMENT"
    else
        if [[ -n "$XBE_TEST_TENDER_RETURN_TENDER_ID" || -n "$XBE_TEST_TENDER_RETURN_TENDER_TYPE" ]]; then
            fail "Failed to create tender return"
        else
            skip "Tender return create failed; set XBE_TEST_TENDER_RETURN_TENDER_ID and XBE_TEST_TENDER_RETURN_TENDER_TYPE"
        fi
    fi
else
    skip "Set XBE_TEST_TENDER_RETURN_TENDER_ID and XBE_TEST_TENDER_RETURN_TENDER_TYPE"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
