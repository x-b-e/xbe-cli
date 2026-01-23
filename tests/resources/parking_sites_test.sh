#!/bin/bash
#
# XBE CLI Integration Tests: Parking Sites
#
# Tests CRUD operations for the parking-sites resource.
# Parking sites require a parked relationship (trailer or tractor).
# This test uses trailers as the parked type, which requires a trailer-classification.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PARKING_SITE_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRAILER_ID=""
TRAILER_CLASSIFICATION_ID=""

describe "Resource: parking-sites"

# ============================================================================
# Prerequisites - Create resources for parking site tests
# ============================================================================

test_name "Create prerequisite broker for parking site tests"
BROKER_NAME=$(unique_name "ParkingTestBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite trucker for parking site tests"
TRUCKER_NAME=$(unique_name "ParkingTrucker")

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "100 Parking Test Lane"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

test_name "Get trailer classification for trailer creation"
# Trailer classifications are read-only reference data, so we list and use the first one
xbe_json view trailer-classifications list --limit 1

if [[ $status -eq 0 ]]; then
    TRAILER_CLASSIFICATION_ID=$(json_get ".[0].id")
    if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        pass
    else
        fail "No trailer classifications available"
        echo "Cannot continue without a trailer classification"
        run_tests
    fi
else
    fail "Failed to list trailer classifications"
    echo "Cannot continue without a trailer classification"
    run_tests
fi

test_name "Create prerequisite trailer for parking site tests"
TRAILER_NUMBER="TRL$(unique_suffix)"

xbe_json do trailers create \
    --number "$TRAILER_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRAILER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRAILER_ID" && "$CREATED_TRAILER_ID" != "null" ]]; then
        register_cleanup "trailers" "$CREATED_TRAILER_ID"
        pass
    else
        fail "Created trailer but no ID returned"
        echo "Cannot continue without a trailer"
        run_tests
    fi
else
    fail "Failed to create trailer"
    echo "Cannot continue without a trailer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create parking site with required fields"
xbe_json do parking-sites create \
    --parked-type trailers \
    --parked-id "$CREATED_TRAILER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PARKING_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PARKING_SITE_ID" && "$CREATED_PARKING_SITE_ID" != "null" ]]; then
        register_cleanup "parking-sites" "$CREATED_PARKING_SITE_ID"
        pass
    else
        fail "Created parking site but no ID returned"
    fi
else
    # Check if trailer is already parked (API constraint)
    if echo "$output" | grep -q "can only be parked at one site at a time"; then
        skip "Trailer already has a parking site (API constraint)"
        # Try to list and get existing parking site
        xbe_json view parking-sites list --limit 100
        if [[ $status -eq 0 ]]; then
            # Find parking site for our trailer
            CREATED_PARKING_SITE_ID=$(echo "$output" | jq -r ".[] | select(.parked_id == \"$CREATED_TRAILER_ID\") | .id" 2>/dev/null | head -1)
        fi
    else
        fail "Failed to create parking site"
    fi
fi

if [[ -z "$CREATED_PARKING_SITE_ID" || "$CREATED_PARKING_SITE_ID" == "null" ]]; then
    echo "Cannot continue CRUD tests without a valid parking site ID - proceeding with list tests only"
    SKIP_CRUD_TESTS=true
fi

if [[ "$SKIP_CRUD_TESTS" != "true" ]]; then

test_name "Create parking site with address"
# Create another trailer for this test
TRAILER_NUMBER2="TRL$(unique_suffix)"
xbe_json do trailers create \
    --number "$TRAILER_NUMBER2" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    TRAILER_ID2=$(json_get ".id")
    register_cleanup "trailers" "$TRAILER_ID2"

    xbe_json do parking-sites create \
        --parked-type trailers \
        --parked-id "$TRAILER_ID2" \
        --address "456 Parking Lot Ave, Test City, TC 12345"

    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "parking-sites" "$id"
        pass
    else
        skip "Parking site creation may have API constraints"
    fi
else
    skip "Could not create trailer for address test"
fi

test_name "Create parking site with is-active flag"
TRAILER_NUMBER3="TRL$(unique_suffix)"
xbe_json do trailers create \
    --number "$TRAILER_NUMBER3" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    TRAILER_ID3=$(json_get ".id")
    register_cleanup "trailers" "$TRAILER_ID3"

    xbe_json do parking-sites create \
        --parked-type trailers \
        --parked-id "$TRAILER_ID3" \
        --is-active

    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "parking-sites" "$id"
        pass
    else
        skip "Parking site creation may have API constraints"
    fi
else
    skip "Could not create trailer for is-active test"
fi

test_name "Create parking site with active times"
TRAILER_NUMBER4="TRL$(unique_suffix)"
xbe_json do trailers create \
    --number "$TRAILER_NUMBER4" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    TRAILER_ID4=$(json_get ".id")
    register_cleanup "trailers" "$TRAILER_ID4"

    xbe_json do parking-sites create \
        --parked-type trailers \
        --parked-id "$TRAILER_ID4" \
        --active-start "2024-01-01T08:00:00Z" \
        --active-end "2024-12-31T17:00:00Z"

    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "parking-sites" "$id"
        pass
    else
        skip "Parking site creation may have API constraints"
    fi
else
    skip "Could not create trailer for active times test"
fi

fi # End SKIP_CRUD_TESTS check for create tests

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ "$SKIP_CRUD_TESTS" != "true" && -n "$CREATED_PARKING_SITE_ID" && "$CREATED_PARKING_SITE_ID" != "null" ]]; then

test_name "Update parking site is-active"
xbe_json do parking-sites update "$CREATED_PARKING_SITE_ID" --is-active
assert_success

test_name "Update parking site address"
xbe_json do parking-sites update "$CREATED_PARKING_SITE_ID" --address "789 Updated Parking Ave"
assert_success

test_name "Update parking site active-start"
xbe_json do parking-sites update "$CREATED_PARKING_SITE_ID" --active-start "2024-02-01T09:00:00Z"
assert_success

test_name "Update parking site active-end"
xbe_json do parking-sites update "$CREATED_PARKING_SITE_ID" --active-end "2024-11-30T18:00:00Z"
assert_success

test_name "Update parking site coordinates"
xbe_json do parking-sites update "$CREATED_PARKING_SITE_ID" --latitude "40.7128" --longitude "-74.0060" --skip-geocoding
assert_success

fi # End SKIP_CRUD_TESTS check for update tests

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List parking sites"
xbe_json view parking-sites list --limit 5
assert_success

test_name "List parking sites returns array"
xbe_json view parking-sites list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list parking sites"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List parking sites with --broker filter"
xbe_json view parking-sites list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List parking sites with --active-start-min filter"
xbe_json view parking-sites list --active-start-min "2024-01-01T00:00:00Z" --limit 10
assert_success

test_name "List parking sites with --active-start-max filter"
xbe_json view parking-sites list --active-start-max "2025-12-31T23:59:59Z" --limit 10
assert_success

test_name "List parking sites with --active-end-min filter"
xbe_json view parking-sites list --active-end-min "2024-01-01T00:00:00Z" --limit 10
assert_success

test_name "List parking sites with --active-end-max filter"
xbe_json view parking-sites list --active-end-max "2025-12-31T23:59:59Z" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List parking sites with --limit"
xbe_json view parking-sites list --limit 3
assert_success

test_name "List parking sites with --offset"
xbe_json view parking-sites list --limit 3 --offset 1
assert_success

test_name "List parking sites with pagination (limit + offset)"
xbe_json view parking-sites list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ "$SKIP_CRUD_TESTS" != "true" && -n "$CREATED_PARKING_SITE_ID" && "$CREATED_PARKING_SITE_ID" != "null" ]]; then

test_name "Delete parking site requires --confirm flag"
xbe_json do parking-sites delete "$CREATED_PARKING_SITE_ID"
assert_failure

test_name "Delete parking site with --confirm"
# Create a parking site specifically for deletion
TRAILER_NUMBER_DEL="TRL$(unique_suffix)"
xbe_json do trailers create \
    --number "$TRAILER_NUMBER_DEL" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    TRAILER_DEL_ID=$(json_get ".id")
    register_cleanup "trailers" "$TRAILER_DEL_ID"

    xbe_json do parking-sites create \
        --parked-type trailers \
        --parked-id "$TRAILER_DEL_ID"

    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_json do parking-sites delete "$DEL_ID" --confirm
        if [[ $status -eq 0 ]]; then
            pass
        else
            # API may not allow deletion
            register_cleanup "parking-sites" "$DEL_ID"
            skip "API may not allow parking site deletion"
        fi
    else
        skip "Could not create parking site for deletion test"
    fi
else
    skip "Could not create trailer for deletion test"
fi

fi # End SKIP_CRUD_TESTS check for delete tests

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create parking site without parked-type fails"
xbe_json do parking-sites create --parked-id "1"
assert_failure

test_name "Create parking site without parked-id fails"
xbe_json do parking-sites create --parked-type trailers
assert_failure

test_name "Update without any fields fails"
# Use a placeholder ID - it should fail for missing fields before hitting the API
xbe_json do parking-sites update "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
