#!/bin/bash
#
# XBE CLI Integration Tests: Communications
#
# Tests list and show operations for the communications resource.
# Communications record inbound/outbound messages with delivery status.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

COMMUNICATION_ID=""
SKIP_SHOW=0

SAMPLE_SUBJECT_TYPE=""
SAMPLE_SUBJECT_ID=""
SAMPLE_USER_ID=""
SAMPLE_MESSAGE_TYPE=""
SAMPLE_MESSAGE_ID=""
SAMPLE_DELIVERY_STATUS=""
SAMPLE_MESSAGE_SENT_AT=""


describe "Resource: communications"

# ============================================================================
# Helpers
# ============================================================================

list_communications() {
    xbe_json view communications list "$@"
    if [[ $status -ne 0 && "$output" == *"SERVICE UNAVAILABLE"* ]]; then
        sleep 2
        xbe_json view communications list "$@"
    fi
}

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List communications"
list_communications --limit 5
assert_success

test_name "List communications returns array"
list_communications --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list communications"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample communication"
list_communications --limit 1
if [[ $status -eq 0 ]]; then
    COMMUNICATION_ID=$(json_get ".[0].id")
    if [[ -n "$COMMUNICATION_ID" && "$COMMUNICATION_ID" != "null" ]]; then
        SAMPLE_SUBJECT_TYPE=$(json_get ".[0].subject_type")
        SAMPLE_SUBJECT_ID=$(json_get ".[0].subject_id")
        SAMPLE_USER_ID=$(json_get ".[0].user_id")
        SAMPLE_MESSAGE_TYPE=$(json_get ".[0].message_type")
        SAMPLE_MESSAGE_ID=$(json_get ".[0].message_id")
        SAMPLE_DELIVERY_STATUS=$(json_get ".[0].delivery_status")
        SAMPLE_MESSAGE_SENT_AT=$(json_get ".[0].message_sent_at")
        pass
    else
        SKIP_SHOW=1
        skip "No communications available"
    fi
else
    SKIP_SHOW=1
    fail "Failed to list communications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List communications with --user"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    list_communications --user "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No sample user ID available"
fi

test_name "List communications with --subject-type"
if [[ -n "$SAMPLE_SUBJECT_TYPE" && "$SAMPLE_SUBJECT_TYPE" != "null" ]]; then
    list_communications --subject-type "$SAMPLE_SUBJECT_TYPE" --limit 5
    assert_success
else
    list_communications --subject-type "Project" --limit 5
    assert_success
fi

test_name "List communications with --subject-type and --subject-id"
if [[ -n "$SAMPLE_SUBJECT_TYPE" && "$SAMPLE_SUBJECT_TYPE" != "null" && -n "$SAMPLE_SUBJECT_ID" && "$SAMPLE_SUBJECT_ID" != "null" ]]; then
    list_communications --subject-type "$SAMPLE_SUBJECT_TYPE" --subject-id "$SAMPLE_SUBJECT_ID" --limit 5
    assert_success
else
    skip "No sample subject reference available"
fi

test_name "List communications with --delivery-status"
if [[ -n "$SAMPLE_DELIVERY_STATUS" && "$SAMPLE_DELIVERY_STATUS" != "null" ]]; then
    list_communications --delivery-status "$SAMPLE_DELIVERY_STATUS" --limit 5
    assert_success
else
    list_communications --delivery-status "incoming_received" --limit 5
    assert_success
fi

test_name "List communications with --message-type"
if [[ -n "$SAMPLE_MESSAGE_TYPE" && "$SAMPLE_MESSAGE_TYPE" != "null" ]]; then
    list_communications --message-type "$SAMPLE_MESSAGE_TYPE" --limit 5
    assert_success
else
    list_communications --message-type "TextMessage" --limit 5
    assert_success
fi

test_name "List communications with --message-id"
if [[ -n "$SAMPLE_MESSAGE_ID" && "$SAMPLE_MESSAGE_ID" != "null" ]]; then
    list_communications --message-id "$SAMPLE_MESSAGE_ID" --limit 5
    assert_success
else
    list_communications --message-id "SM123" --limit 5
    assert_success
fi

test_name "List communications with --message-sent-at-min"
list_communications --message-sent-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List communications with --message-sent-at-max"
list_communications --message-sent-at-max "2025-12-31T23:59:59Z" --limit 5
assert_success

test_name "List communications with --is-message-sent-at"
list_communications --is-message-sent-at true --limit 5
assert_success

test_name "List communications with --is-addressed"
list_communications --is-addressed true --limit 5
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"undefined method"* && "$output" == *"addressed"* ]]; then
    skip "Server error on is-addressed filter"
else
    fail "Expected success (exit 0), got exit $status"
fi

test_name "List communications with --is-retried"
list_communications --is-retried false --limit 5
assert_success

test_name "List communications with --created-at-min"
list_communications --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List communications with --created-at-max"
list_communications --created-at-max "2025-12-31T23:59:59Z" --limit 5
assert_success

test_name "List communications with --updated-at-min"
list_communications --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List communications with --updated-at-max"
list_communications --updated-at-max "2025-12-31T23:59:59Z" --limit 5
assert_success

test_name "List communications with --is-created-at"
list_communications --is-created-at true --limit 5
assert_success

test_name "List communications with --is-updated-at"
list_communications --is-updated-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show communication"
if [[ $SKIP_SHOW -eq 0 && -n "$COMMUNICATION_ID" && "$COMMUNICATION_ID" != "null" ]]; then
    xbe_json view communications show "$COMMUNICATION_ID"
    assert_success
else
    skip "No communication ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
