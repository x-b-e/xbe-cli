#!/bin/bash
#
# XBE CLI Integration Tests: User Location Events
#
# Tests list/show operations, filters, and create/update/delete behavior.
#
# COVERAGE: List + filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CURRENT_USER_ID=""
SAMPLE_EVENT_ID=""
SAMPLE_EVENT_AT=""
SAMPLE_PROVENANCE=""
SAMPLE_USER_ID=""
SAMPLE_UPDATED_BY_ID=""

CREATED_EVENT_ID=""


describe "Resource: user-location-events"

# ============================================================================
# AUTH / CONTEXT
# ============================================================================

test_name "Resolve current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        skip "No user ID returned"
    fi
else
    skip "Failed to resolve current user"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user location events"
xbe_json view user-location-events list --limit 5
assert_success

test_name "List user location events returns array"
xbe_json view user-location-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user location events"
fi

test_name "Capture sample user location event (if available)"
xbe_json view user-location-events list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_EVENT_ID=$(json_get ".[0].id")
        SAMPLE_EVENT_AT=$(json_get ".[0].event_at")
        SAMPLE_PROVENANCE=$(json_get ".[0].provenance")
        SAMPLE_USER_ID=$(json_get ".[0].user_id")
        SAMPLE_UPDATED_BY_ID=$(json_get ".[0].updated_by_id")
        pass
    else
        echo "    No user location events available; using fallback IDs for filter tests."
        pass
    fi
else
    fail "Failed to list user location events"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

USER_FILTER="${SAMPLE_USER_ID:-${CURRENT_USER_ID:-1}}"
UPDATED_BY_FILTER="${SAMPLE_UPDATED_BY_ID:-${CURRENT_USER_ID:-1}}"
PROVENANCE_FILTER="${SAMPLE_PROVENANCE:-gps}"
EVENT_AT_FILTER="${SAMPLE_EVENT_AT:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"


test_name "List user location events with --user filter"
xbe_json view user-location-events list --user "$USER_FILTER" --limit 5
assert_success

test_name "List user location events with --updated-by filter"
xbe_json view user-location-events list --updated-by "$UPDATED_BY_FILTER" --limit 5
assert_success

test_name "List user location events with --event-at-min filter"
xbe_json view user-location-events list --event-at-min "$EVENT_AT_FILTER" --limit 5
assert_success

test_name "List user location events with --event-at-max filter"
xbe_json view user-location-events list --event-at-max "$EVENT_AT_FILTER" --limit 5
assert_success

test_name "List user location events with --provenance filter"
xbe_json view user-location-events list --provenance "$PROVENANCE_FILTER" --limit 5
assert_success

test_name "List user location events with --is-event-at filter"
xbe_json view user-location-events list --is-event-at true --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    test_name "Create user location event"
    EVENT_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    xbe_json do user-location-events create \
        --user "$CURRENT_USER_ID" \
        --provenance gps \
        --event-at "$EVENT_AT" \
        --event-latitude 40.0 \
        --event-longitude -74.0
    if [[ $status -eq 0 ]]; then
        CREATED_EVENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_EVENT_ID" && "$CREATED_EVENT_ID" != "null" ]]; then
            register_cleanup "user-location-events" "$CREATED_EVENT_ID"
            pass
        else
            fail "Created user location event but no ID returned"
        fi
    else
        fail "Failed to create user location event: $output"
    fi
else
    test_name "Create user location event"
    skip "No current user available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_EVENT_ID" && "$CREATED_EVENT_ID" != "null" ]]; then
    test_name "Update user location event"
    UPDATED_EVENT_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    xbe_json do user-location-events update "$CREATED_EVENT_ID" \
        --event-at "$UPDATED_EVENT_AT" \
        --event-latitude 41.0 \
        --event-longitude -87.0 \
        --provenance map
    assert_success

    test_name "Show user location event reflects updates"
    xbe_json view user-location-events show "$CREATED_EVENT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".provenance" "map"
    else
        fail "Failed to show user location event"
    fi
else
    test_name "Update user location event"
    skip "No user location event available to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_EVENT_ID" && "$CREATED_EVENT_ID" != "null" ]]; then
    test_name "Delete user location event"
    xbe_json do user-location-events delete "$CREATED_EVENT_ID" --confirm
    assert_success
else
    test_name "Delete user location event"
    skip "No user location event available to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create user location event without user fails"
xbe_run do user-location-events create --provenance gps --event-latitude 40.0 --event-longitude -74.0
assert_failure

if [[ -n "$CREATED_EVENT_ID" && "$CREATED_EVENT_ID" != "null" ]]; then
    test_name "Update user location event without changes fails"
    xbe_run do user-location-events update "$CREATED_EVENT_ID"
    assert_failure
else
    test_name "Update user location event without changes fails"
    skip "No user location event available to update"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
