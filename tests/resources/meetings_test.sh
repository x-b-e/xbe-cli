#!/bin/bash
#
# XBE CLI Integration Tests: Meetings
#
# Tests CRUD operations for the meetings resource.
# Meetings track scheduled discussions tied to an organization.
#
# COVERAGE: All create/update attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEETING_ID=""
CREATED_BROKER_ID=""
CURRENT_USER_ID=""
CREATED_MEMBERSHIP_ID=""

describe "Resource: meetings"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Resolve current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "No user ID returned from whoami"
        run_tests
    fi
else
    fail "Failed to resolve current user"
    run_tests
fi

test_name "Create prerequisite broker for meetings tests"
BROKER_NAME=$(unique_name "MeetingsBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        run_tests
    fi
fi

test_name "Ensure current user membership for broker"
xbe_json do memberships create \
    --user "$CURRENT_USER_ID" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
    fi
    pass
else
    skip "Membership already exists or cannot be created"
fi

date_utc_now() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

date_utc_add_hours() {
    local hours="$1"
    if date -u -d "now +${hours} hour" +"%Y-%m-%dT%H:%M:%SZ" >/dev/null 2>&1; then
        date -u -d "now +${hours} hour" +"%Y-%m-%dT%H:%M:%SZ"
    else
        date -u -v+${hours}H +"%Y-%m-%dT%H:%M:%SZ"
    fi
}

START_AT=$(date_utc_now)
END_AT=$(date_utc_add_hours 1)
UPDATED_START_AT=$(date_utc_add_hours 2)
UPDATED_END_AT=$(date_utc_add_hours 3)

# ============================================================================
# Create
# ============================================================================

test_name "Create meeting with full attributes"
MEETING_SUBJECT=$(unique_name "Meeting")
MEETING_DESCRIPTION="Initial meeting description"

xbe_json do meetings create \
    --organization-type brokers \
    --organization-id "$CREATED_BROKER_ID" \
    --organizer "$CURRENT_USER_ID" \
    --subject "$MEETING_SUBJECT" \
    --description "$MEETING_DESCRIPTION" \
    --start-at "$START_AT" \
    --end-at "$END_AT" \
    --explicit-time-zone-id "America/Chicago" \
    --address "123 Test St, Chicago, IL" \
    --skip-geocoding \
    --address-latitude "41.88" \
    --address-longitude "-87.63" \
    --address-place-id "ChIJD7fiBh9u5kcRYJSMaMOCCwQ" \
    --address-plus-code "86HJWJ8F+7C"

if [[ $status -eq 0 ]]; then
    CREATED_MEETING_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEETING_ID" && "$CREATED_MEETING_ID" != "null" ]]; then
        register_cleanup "meetings" "$CREATED_MEETING_ID"
        pass
    else
        fail "Created meeting but no ID returned"
        run_tests
    fi
else
    fail "Failed to create meeting"
    run_tests
fi

# ============================================================================
# Update
# ============================================================================

test_name "Update meeting attributes"
xbe_json do meetings update "$CREATED_MEETING_ID" \
    --subject "Updated Meeting Subject" \
    --description "Updated meeting description" \
    --transcript "Discussed safety protocols and next steps." \
    --summary "Summary of updated meeting discussion." \
    --start-at "$UPDATED_START_AT" \
    --end-at "$UPDATED_END_AT" \
    --explicit-time-zone-id "America/New_York" \
    --address "456 Oak Ave, Springfield, IL" \
    --skip-geocoding \
    --address-latitude "39.78" \
    --address-longitude "-89.64" \
    --address-place-id "ChIJ7cmZVwkZ4ogRslZsYwJES7s" \
    --address-plus-code "86HGMH4X+5G"

assert_success

# ============================================================================
# View
# ============================================================================

test_name "Show meeting details"
xbe_json view meetings show "$CREATED_MEETING_ID"
assert_success
assert_json_has ".id"

# ============================================================================
# List + Filters
# ============================================================================

test_name "List meetings"
xbe_json view meetings list --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --organization filter"
xbe_json view meetings list --organization "Broker|$CREATED_BROKER_ID" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --organizer filter"
xbe_json view meetings list --organizer "$CURRENT_USER_ID" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --broker filter"
xbe_json view meetings list --broker "$CREATED_BROKER_ID" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --attendee filter"
xbe_json view meetings list --attendee "$CURRENT_USER_ID" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with start/end time filters"
xbe_json view meetings list --start-at-min "$START_AT" --start-at-max "$END_AT" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with end-at filters"
xbe_json view meetings list --end-at-min "$START_AT" --end-at-max "$END_AT" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --is-start-at filter"
xbe_json view meetings list --is-start-at true --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --is-end-at filter"
xbe_json view meetings list --is-end-at true --limit 5
assert_success
assert_json_is_array

test_name "List meetings with --is-safety-meeting filter"
xbe_json view meetings list --is-safety-meeting false --limit 5
assert_success
assert_json_is_array

test_name "List meetings with created/updated filters"
xbe_json view meetings list --created-at-min "$START_AT" --updated-at-max "$UPDATED_END_AT" --limit 5
assert_success
assert_json_is_array

test_name "List meetings with is-created-at/is-updated-at filters"
xbe_json view meetings list --is-created-at true --is-updated-at true --limit 5
assert_success
assert_json_is_array

test_name "List meetings with pagination"
xbe_json view meetings list --limit 3 --offset 0
assert_success
assert_json_is_array

# ============================================================================
# Delete
# ============================================================================

test_name "Delete meeting requires --confirm"
xbe_run do meetings delete "$CREATED_MEETING_ID"
assert_failure

test_name "Delete meeting with --confirm"
xbe_run do meetings delete "$CREATED_MEETING_ID" --confirm
assert_success

run_tests
