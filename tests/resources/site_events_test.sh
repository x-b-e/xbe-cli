#!/bin/bash
#
# XBE CLI Integration Tests: Site Events
#
# Tests create, update, delete operations and list filters for the site_events resource.
#
# COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SITE_EVENT_ID=""
CREATED_EVENT_TYPE=""
CREATED_EVENT_AT=""
CREATED_TJSS_ID=""
CREATED_MT_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
EVENT_SITE_TYPE=""
EVENT_SITE_ID=""

SAMPLE_EVENT_ID=""
SAMPLE_EVENT_TYPE=""
SAMPLE_EVENT_AT=""
SAMPLE_TJSS_ID=""
SAMPLE_MT_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_TRUCKER_ID=""

MATERIAL_SITE_ID=""
DEST_ID=""
DEST_TYPE=""
ORIGIN_ID=""
ORIGIN_TYPE=""


describe "Resource: site-events"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List site events"
xbe_json view site-events list --limit 5
assert_success

test_name "List site events returns array"
xbe_json view site-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list site events"
fi

test_name "Capture sample site event (if available)"
xbe_json view site-events list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_EVENT_ID=$(json_get ".[0].id")
        SAMPLE_EVENT_TYPE=$(json_get ".[0].event_type")
        SAMPLE_EVENT_AT=$(json_get ".[0].event_at")
        SAMPLE_TJSS_ID=$(json_get ".[0].tender_job_schedule_shift_id")
        SAMPLE_MT_ID=$(json_get ".[0].material_transaction_id")
        SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
        SAMPLE_TRUCKER_ID=$(json_get ".[0].trucker_id")
        pass
    else
        echo "    No site events available; using fallback IDs for filter tests."
        pass
    fi
else
    fail "Failed to list site events"
fi

# ============================================================================
# CREATE Tests - Locate dependencies
# ============================================================================

test_name "Find material transaction with shift for site event"
xbe_json view material-transactions list --has-shift true --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        CREATED_MT_ID=$(json_get ".[0].id")
        if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
            pass
        else
            skip "Material transaction ID missing"
        fi
    else
        skip "No material transactions with shifts available"
    fi
else
    skip "Failed to list material transactions"
fi

if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
    test_name "Lookup event site for material transaction"
    xbe_json view material-transactions show "$CREATED_MT_ID"
    if [[ $status -eq 0 ]]; then
        MATERIAL_SITE_ID=$(json_get ".material_site_id")
        DEST_ID=$(json_get ".destination_id")
        DEST_TYPE=$(json_get ".destination_type")
        ORIGIN_ID=$(json_get ".origin_id")
        ORIGIN_TYPE=$(json_get ".origin_type")

        if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
            EVENT_SITE_TYPE="material-sites"
            EVENT_SITE_ID="$MATERIAL_SITE_ID"
        elif [[ -n "$DEST_ID" && "$DEST_ID" != "null" ]]; then
            if [[ "$DEST_TYPE" == "job-sites" || "$DEST_TYPE" == "material-sites" || "$DEST_TYPE" == "parking-sites" ]]; then
                EVENT_SITE_TYPE="$DEST_TYPE"
                EVENT_SITE_ID="$DEST_ID"
            fi
        elif [[ -n "$ORIGIN_ID" && "$ORIGIN_ID" != "null" ]]; then
            if [[ "$ORIGIN_TYPE" == "job-sites" || "$ORIGIN_TYPE" == "material-sites" || "$ORIGIN_TYPE" == "parking-sites" ]]; then
                EVENT_SITE_TYPE="$ORIGIN_TYPE"
                EVENT_SITE_ID="$ORIGIN_ID"
            fi
        fi

        if [[ -n "$EVENT_SITE_ID" && "$EVENT_SITE_ID" != "null" ]]; then
            pass
        else
            skip "No valid event site found for material transaction"
        fi
    else
        skip "Failed to show material transaction"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" && -n "$EVENT_SITE_ID" && "$EVENT_SITE_ID" != "null" ]]; then
    test_name "Create site event"
    EVENT_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    EVENT_DETAILS="Site event $(date +%s)"

    xbe_json do site-events create \
        --event-type start_work \
        --event-kind load \
        --event-details "$EVENT_DETAILS" \
        --event-at "$EVENT_AT" \
        --event-latitude 41.8781 \
        --event-longitude -87.6298 \
        --material-transaction "$CREATED_MT_ID" \
        --event-site-type "$EVENT_SITE_TYPE" \
        --event-site-id "$EVENT_SITE_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_SITE_EVENT_ID=$(json_get ".id")
        CREATED_EVENT_TYPE=$(json_get ".event_type")
        CREATED_EVENT_AT=$(json_get ".event_at")
        CREATED_TJSS_ID=$(json_get ".tender_job_schedule_shift_id")
        CREATED_BROKER_ID=$(json_get ".broker_id")
        CREATED_TRUCKER_ID=$(json_get ".trucker_id")
        if [[ -n "$CREATED_SITE_EVENT_ID" && "$CREATED_SITE_EVENT_ID" != "null" ]]; then
            register_cleanup "site-events" "$CREATED_SITE_EVENT_ID"
            pass
        else
            fail "Created site event but no ID returned"
        fi
    else
        fail "Failed to create site event: $output"
    fi
else
    test_name "Create site event"
    skip "Missing material transaction or event site for creation"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_SITE_EVENT_ID" && "$CREATED_SITE_EVENT_ID" != "null" ]]; then
    test_name "Update site event"
    UPDATED_DETAILS="Updated site event $(date +%s)"
    UPDATED_EVENT_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    xbe_json do site-events update "$CREATED_SITE_EVENT_ID" \
        --event-details "$UPDATED_DETAILS" \
        --event-at "$UPDATED_EVENT_AT" \
        --event-latitude 40.0 \
        --event-longitude -80.0
    assert_success

    test_name "Show site event reflects updates"
    xbe_json view site-events show "$CREATED_SITE_EVENT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".event_details" "$UPDATED_DETAILS"
    else
        fail "Failed to show site event"
    fi
else
    test_name "Update site event"
    skip "No site event available to update"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

EVENT_TYPE_FILTER="${CREATED_EVENT_TYPE:-$SAMPLE_EVENT_TYPE}"
EVENT_AT_FILTER="${CREATED_EVENT_AT:-$SAMPLE_EVENT_AT}"
TENDER_SHIFT_FILTER="${CREATED_TJSS_ID:-$SAMPLE_TJSS_ID}"
MATERIAL_TRANSACTION_FILTER="${CREATED_MT_ID:-$SAMPLE_MT_ID}"
BROKER_FILTER="${CREATED_BROKER_ID:-$SAMPLE_BROKER_ID}"
TRUCKER_FILTER="${CREATED_TRUCKER_ID:-$SAMPLE_TRUCKER_ID}"
DRIVER_DAY_FILTER="${XBE_TEST_DRIVER_DAY_ID:-1}"

if [[ -z "$EVENT_TYPE_FILTER" || "$EVENT_TYPE_FILTER" == "null" ]]; then
    EVENT_TYPE_FILTER="start_work"
fi
if [[ -z "$EVENT_AT_FILTER" || "$EVENT_AT_FILTER" == "null" ]]; then
    EVENT_AT_FILTER=$(date -u +%Y-%m-%dT%H:%M:%SZ)
fi
if [[ -z "$TENDER_SHIFT_FILTER" || "$TENDER_SHIFT_FILTER" == "null" ]]; then
    TENDER_SHIFT_FILTER=1
fi
if [[ -z "$MATERIAL_TRANSACTION_FILTER" || "$MATERIAL_TRANSACTION_FILTER" == "null" ]]; then
    MATERIAL_TRANSACTION_FILTER=1
fi
if [[ -z "$BROKER_FILTER" || "$BROKER_FILTER" == "null" ]]; then
    BROKER_FILTER=1
fi
if [[ -z "$TRUCKER_FILTER" || "$TRUCKER_FILTER" == "null" ]]; then
    TRUCKER_FILTER=1
fi



test_name "List site events with --tender-job-schedule-shift filter"
xbe_json view site-events list --tender-job-schedule-shift "$TENDER_SHIFT_FILTER" --limit 5
assert_success

test_name "List site events with --material-transaction filter"
xbe_json view site-events list --material-transaction "$MATERIAL_TRANSACTION_FILTER" --limit 5
assert_success

test_name "List site events with --broker filter"
xbe_json view site-events list --broker "$BROKER_FILTER" --limit 5
assert_success

test_name "List site events with --trucker filter"
xbe_json view site-events list --trucker "$TRUCKER_FILTER" --limit 5
assert_success

test_name "List site events with --event-type filter"
xbe_json view site-events list --event-type "$EVENT_TYPE_FILTER" --limit 5
assert_success

test_name "List site events with --event-at-min filter"
xbe_json view site-events list --event-at-min "$EVENT_AT_FILTER" --limit 5
assert_success

test_name "List site events with --event-at-max filter"
xbe_json view site-events list --event-at-max "$EVENT_AT_FILTER" --limit 5
assert_success

test_name "List site events with --driver-day filter"
xbe_json view site-events list --driver-day "$DRIVER_DAY_FILTER" --limit 5
assert_success

test_name "List site events with --has-shift filter"
xbe_json view site-events list --has-shift true --limit 5
assert_success

test_name "List site events with --most-recent-by-shift filter"
xbe_json view site-events list --most-recent-by-shift true --limit 5
assert_success

test_name "List site events with --most-recent-by-driver-day filter"
xbe_json view site-events list --most-recent-by-driver-day true --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_SITE_EVENT_ID" && "$CREATED_SITE_EVENT_ID" != "null" ]]; then
    test_name "Delete site event requires --confirm flag"
    xbe_json do site-events delete "$CREATED_SITE_EVENT_ID"
    assert_failure

    test_name "Delete site event with --confirm"
    xbe_json do site-events delete "$CREATED_SITE_EVENT_ID" --confirm
    assert_success
else
    test_name "Delete site event"
    skip "No site event available to delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
