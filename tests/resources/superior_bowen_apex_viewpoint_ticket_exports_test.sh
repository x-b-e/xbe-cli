#!/bin/bash
#
# XBE CLI Integration Tests: Superior Bowen Apex Viewpoint Ticket Exports
#
# Tests list and create operations for the superior-bowen-apex-viewpoint-ticket-exports resource.
#
# COVERAGE: List + create attributes + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

LOCATION_ID=""

if [[ -z "$XBE_TEST_SB_APEX_VIEWPOINT_LOCATION_ID" && -n "$XBE_TOKEN" ]]; then
    base_url="${XBE_BASE_URL%/}"
    meta_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/superior-bowen-apex-viewpoint-ticket-exports?page[limit]=1" || true)

    LOCATION_ID=$(echo "$meta_json" | jq -r '.meta.locations[0]["location-id"] // .meta.locations[0].location_id // empty' 2>/dev/null || true)
fi

if [[ -n "$XBE_TEST_SB_APEX_VIEWPOINT_LOCATION_ID" ]]; then
    LOCATION_ID="$XBE_TEST_SB_APEX_VIEWPOINT_LOCATION_ID"
fi

describe "Resource: superior-bowen-apex-viewpoint-ticket-exports"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List Superior Bowen Apex Viewpoint ticket exports"
xbe_json view superior-bowen-apex-viewpoint-ticket-exports list --limit 5
assert_success

test_name "List ticket exports returns array"
xbe_json view superior-bowen-apex-viewpoint-ticket-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list exports"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create export without required fields fails"
xbe_run do superior-bowen-apex-viewpoint-ticket-exports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

SALE_DATE_MIN="${XBE_TEST_SB_APEX_VIEWPOINT_SALE_DATE_MIN:-2025-01-01}"
SALE_DATE_MAX="${XBE_TEST_SB_APEX_VIEWPOINT_SALE_DATE_MAX:-2025-01-31}"

create_args=(
    do superior-bowen-apex-viewpoint-ticket-exports create
    --sale-date-min "$SALE_DATE_MIN"
    --sale-date-max "$SALE_DATE_MAX"
)

if [[ -n "$LOCATION_ID" ]]; then
    create_args+=(--location-ids "$LOCATION_ID")
fi

test_name "Create Superior Bowen Apex Viewpoint ticket export"
xbe_json "${create_args[@]}"

if [[ $status -eq 0 ]]; then
    assert_json_equals ".sale_date_min" "$SALE_DATE_MIN"
    assert_json_equals ".sale_date_max" "$SALE_DATE_MAX"
    if [[ -n "$LOCATION_ID" ]]; then
        assert_json_equals ".location_ids[0]" "$LOCATION_ID"
    fi
else
    if [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
        skip "Not authorized to create export"
    elif [[ "$output" == *"Couldn't find Broker"* ]] || [[ "$output" == *"broker doesn't exist"* ]]; then
        skip "Superior Bowen broker not configured"
    elif [[ "$output" == *"Connection to Apex/JWS"* ]] || [[ "$output" == *"Unexpected error querying Apex/JWS"* ]]; then
        skip "Apex/JWS unavailable for export"
    else
        fail "Failed to create export"
    fi
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
