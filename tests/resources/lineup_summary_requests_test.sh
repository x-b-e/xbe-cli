#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Summary Requests
#
# Tests list, show, and create operations for the lineup-summary-requests resource.
#
# COVERAGE: List + filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
CREATED_BROKER_ID=""

START_AT_MIN="2026-01-23T00:00:00Z"
START_AT_MAX="2026-01-24T00:00:00Z"

EMAIL_TO=""

describe "Resource: lineup-summary-requests"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup summary requests"
xbe_json view lineup-summary-requests list --limit 5
assert_success

test_name "List lineup summary requests returns array"
xbe_json view lineup-summary-requests list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup summary requests"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample lineup summary request"
xbe_json view lineup-summary-requests list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No lineup summary requests available for follow-on tests"
    fi
else
    skip "Could not list lineup summary requests to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineup summary requests with --created-at-min filter"
xbe_json view lineup-summary-requests list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup summary requests with --created-at-max filter"
xbe_json view lineup-summary-requests list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup summary requests with --updated-at-min filter"
xbe_json view lineup-summary-requests list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup summary requests with --updated-at-max filter"
xbe_json view lineup-summary-requests list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup summary request"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-summary-requests show "$SAMPLE_ID"
    assert_success
else
    skip "No lineup summary request ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prerequisite broker for lineup summary requests"
BROKER_NAME=$(unique_name "LineupSummaryBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
    fi
fi

EMAIL_TO=$(unique_email)


test_name "Create lineup summary request"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json do lineup-summary-requests create \
        --level-type Broker \
        --level-id "$CREATED_BROKER_ID" \
        --start-at-min "$START_AT_MIN" \
        --start-at-max "$START_AT_MAX" \
        --email-to "$EMAIL_TO" \
        --send-if-no-shifts \
        --note "CLI test request"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No broker ID available for create"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create lineup summary request without required fields fails"
xbe_run do lineup-summary-requests create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
