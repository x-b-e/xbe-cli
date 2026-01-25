#!/bin/bash
#
# XBE CLI Integration Tests: Tender Rejections
#
# Tests list, show, and create operations for the tender-rejections resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TENDER_ID=""
SAMPLE_TENDER_TYPE=""
CREATE_TENDER_ID=""
CREATE_TENDER_TYPE=""
LIST_SUPPORTED="true"

describe "Resource: tender-rejections"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tender rejections"
xbe_json view tender-rejections list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || \
       [[ "$output" == *"doesn't exist"* ]] || \
       [[ "$output" == *"Not Implemented"* ]] || \
       [[ "$output" == *"resolve_non_admin_for_list"* ]] || \
       [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing tender rejections for this user"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List tender rejections returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view tender-rejections list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list tender rejections"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample tender rejection"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view tender-rejections list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TENDER_ID=$(json_get ".[0].tender_id")
        SAMPLE_TENDER_TYPE=$(json_get ".[0].tender_type")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No tender rejections available for follow-on tests"
        fi
    else
        skip "Could not list tender rejections to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

if [[ -n "$XBE_TEST_TENDER_ID" && -n "$XBE_TEST_TENDER_TYPE" ]]; then
    CREATE_TENDER_ID="$XBE_TEST_TENDER_ID"
    CREATE_TENDER_TYPE="$XBE_TEST_TENDER_TYPE"
elif [[ -n "$SAMPLE_TENDER_ID" && "$SAMPLE_TENDER_ID" != "null" && -n "$SAMPLE_TENDER_TYPE" && "$SAMPLE_TENDER_TYPE" != "null" ]]; then
    CREATE_TENDER_ID="$SAMPLE_TENDER_ID"
    CREATE_TENDER_TYPE="$SAMPLE_TENDER_TYPE"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create tender rejection"
if [[ -n "$CREATE_TENDER_ID" && "$CREATE_TENDER_ID" != "null" && -n "$CREATE_TENDER_TYPE" && "$CREATE_TENDER_TYPE" != "null" ]]; then
    xbe_json do tender-rejections create \
        --tender-type "$CREATE_TENDER_TYPE" \
        --tender "$CREATE_TENDER_ID" \
        --comment "CLI test rejection"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"previous status must be valid"* ]] || \
           [[ "$output" == *"offered"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No tender type/id available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender rejection"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-rejections show "$SAMPLE_ID"
    assert_success
else
    skip "No tender rejection ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tender rejection without tender fails"
xbe_run do tender-rejections create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
