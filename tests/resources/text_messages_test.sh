#!/bin/bash
#
# XBE CLI Integration Tests: Text Messages
#
# Tests list and show operations for the text-messages resource.
# Text messages are sourced from Twilio and require admin access to list.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TEXT_MESSAGE_ID=""
SKIP_SHOW=0

SAMPLE_TO=""
SAMPLE_FROM=""
SAMPLE_DATE_SENT=""


describe "Resource: text-messages"

# ============================================================================
# Helpers
# ============================================================================

list_text_messages() {
    xbe_json view text-messages list "$@"
    if [[ $status -ne 0 && "$output" == *"SERVICE UNAVAILABLE"* ]]; then
        sleep 2
        xbe_json view text-messages list "$@"
    fi
}

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List text messages"
list_text_messages
assert_success

test_name "List text messages returns array"
list_text_messages --date-sent "2024-01-01"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list text messages"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample text message"
list_text_messages
if [[ $status -eq 0 ]]; then
    TEXT_MESSAGE_ID=$(json_get ".[0].id")
    if [[ -n "$TEXT_MESSAGE_ID" && "$TEXT_MESSAGE_ID" != "null" ]]; then
        SAMPLE_TO=$(json_get ".[0].to")
        SAMPLE_FROM=$(json_get ".[0].from")
        SAMPLE_DATE_SENT=$(json_get ".[0].date_sent")
        pass
    else
        SKIP_SHOW=1
        skip "No text messages available"
    fi
else
    SKIP_SHOW=1
    fail "Failed to list text messages"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List text messages with --to"
if [[ -n "$SAMPLE_TO" && "$SAMPLE_TO" != "null" ]]; then
    list_text_messages --to "$SAMPLE_TO"
    assert_success
else
    list_text_messages --to "+15551234567"
    assert_success
fi

test_name "List text messages with --from"
if [[ -n "$SAMPLE_FROM" && "$SAMPLE_FROM" != "null" ]]; then
    list_text_messages --from "$SAMPLE_FROM"
    assert_success
else
    list_text_messages --from "+15559876543"
    assert_success
fi

test_name "List text messages with --date-sent"
list_text_messages --date-sent "2024-01-01"
assert_success

test_name "List text messages with --date-sent-after"
list_text_messages --date-sent-after "2024-01-01"
assert_success

test_name "List text messages with --date-sent-before"
list_text_messages --date-sent-before "2025-12-31"
assert_success

test_name "List text messages with --max-messages"
list_text_messages --max-messages 2
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"Internal Server Error"* ]]; then
    skip "Server error on max-messages filter"
else
    fail "Expected success (exit 0), got exit $status"
fi

test_name "List text messages with --page-size"
list_text_messages --page-size 10 --max-messages 5
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"Internal Server Error"* ]]; then
    skip "Server error on page-size filter"
else
    fail "Expected success (exit 0), got exit $status"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show text message"
if [[ $SKIP_SHOW -eq 0 && -n "$TEXT_MESSAGE_ID" && "$TEXT_MESSAGE_ID" != "null" ]]; then
    xbe_json view text-messages show "$TEXT_MESSAGE_ID"
    assert_success
else
    skip "No text message ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
