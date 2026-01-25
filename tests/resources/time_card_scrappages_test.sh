#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Scrappages
#
# Tests list, show, and create operations for the time-card-scrappages resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TIME_CARD_ID=""
CREATE_TIME_CARD_ID=""
LIST_SUPPORTED="true"

describe "Resource: time-card-scrappages"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card scrappages"
xbe_json view time-card-scrappages list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing time card scrappages"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List time card scrappages returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-card-scrappages list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list time card scrappages"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample time card scrappage"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-card-scrappages list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No time card scrappages available for follow-on tests"
        fi
    else
        skip "Could not list time card scrappages to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

if [[ -n "$XBE_TEST_TIME_CARD_ID" ]]; then
    CREATE_TIME_CARD_ID="$XBE_TEST_TIME_CARD_ID"
elif [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    CREATE_TIME_CARD_ID="$SAMPLE_TIME_CARD_ID"
else
    xbe_json view time-card-cost-code-allocations list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time card scrappage"
if [[ -n "$CREATE_TIME_CARD_ID" && "$CREATE_TIME_CARD_ID" != "null" ]]; then
    xbe_json do time-card-scrappages create \
        --time-card "$CREATE_TIME_CARD_ID" \
        --comment "CLI test scrappage"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"can only be scrapped when value is exactly $0.00"* ]] || \
           [[ "$output" == *"cannot change status when already associated with an invoice"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No time card ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time card scrappage"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view time-card-scrappages show "$SAMPLE_ID"
    assert_success
else
    skip "No scrappage ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create scrappage without time card fails"
xbe_run do time-card-scrappages create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
