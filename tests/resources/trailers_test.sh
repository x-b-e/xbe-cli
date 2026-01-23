#!/bin/bash
#
# XBE CLI Integration Tests: Trailers
#
# Tests CRUD operations for the trailers resource.
# Trailers are the cargo-carrying units in a trucking fleet.
#
# NOTE: Trailers require a trailer-classification when creating.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TRAILER_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""

describe "Resource: trailers"

# ============================================================================
# Prerequisites - Create broker and trucker
# ============================================================================

test_name "Create prerequisite broker for trailers tests"
BROKER_NAME=$(unique_name "TrailerTestBroker")

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

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "TrailerTestTrucker")
TRUCKER_ADDRESS="350 N Orleans St, Chicago, IL 60654"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trailer with required fields"
TRAILER_NUMBER=$(unique_name "Trailer")

# Trailers require a trailer-classification. Use ID 1 (6 Wheeler / Tandem) which exists.
xbe_json do trailers create \
    --number "$TRAILER_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "1"

if [[ $status -eq 0 ]]; then
    CREATED_TRAILER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRAILER_ID" && "$CREATED_TRAILER_ID" != "null" ]]; then
        register_cleanup "trailers" "$CREATED_TRAILER_ID"
        pass
    else
        fail "Created trailer but no ID returned"
    fi
else
    fail "Failed to create trailer"
fi

# Only continue if we successfully created a trailer
if [[ -z "$CREATED_TRAILER_ID" || "$CREATED_TRAILER_ID" == "null" ]]; then
    echo "Cannot continue without a valid trailer ID"
    run_tests
fi

test_name "Create trailer with trailer details"
TRAILER_NUMBER2=$(unique_name "Trailer2")
xbe_json do trailers create \
    --number "$TRAILER_NUMBER2" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "1" \
    --kind "end_dump" \
    --composition "steel"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "trailers" "$id"
    pass
else
    fail "Failed to create trailer with trailer details"
fi

test_name "Create trailer with capacity"
TRAILER_NUMBER3=$(unique_name "Trailer3")
xbe_json do trailers create \
    --number "$TRAILER_NUMBER3" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "1" \
    --capacity-lbs 50000
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "trailers" "$id"
    pass
else
    fail "Failed to create trailer with capacity"
fi

test_name "Create trailer with in-service status"
TRAILER_NUMBER4=$(unique_name "Trailer4")
xbe_json do trailers create \
    --number "$TRAILER_NUMBER4" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "1" \
    --in-service true
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "trailers" "$id"
    pass
else
    fail "Failed to create trailer with in-service status"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update trailer number"
UPDATED_NUMBER=$(unique_name "UpdatedTrailer")
xbe_json do trailers update "$CREATED_TRAILER_ID" --number "$UPDATED_NUMBER"
assert_success

test_name "Update trailer kind"
xbe_json do trailers update "$CREATED_TRAILER_ID" --kind "belly_dump"
assert_success

test_name "Update trailer composition"
xbe_json do trailers update "$CREATED_TRAILER_ID" --composition "aluminum"
assert_success

test_name "Update trailer capacity-lbs"
xbe_json do trailers update "$CREATED_TRAILER_ID" --capacity-lbs 45000
assert_success

test_name "Update trailer in-service"
xbe_json do trailers update "$CREATED_TRAILER_ID" --in-service true
assert_success

test_name "Update trailer coal-chute"
xbe_json do trailers update "$CREATED_TRAILER_ID" --coal-chute true
assert_success

test_name "Update trailer vibrator"
xbe_json do trailers update "$CREATED_TRAILER_ID" --vibrator true
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trailers"
xbe_json view trailers list --limit 5
assert_success

test_name "List trailers returns array"
xbe_json view trailers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trailers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trailers with --trucker filter"
xbe_json view trailers list --trucker "$CREATED_TRUCKER_ID" --limit 10
assert_success

test_name "List trailers with --in-service filter"
xbe_json view trailers list --in-service true --limit 10
assert_success

test_name "List trailers with --number-like filter"
xbe_json view trailers list --number-like "Trailer" --limit 10
assert_success

test_name "List trailers with --trailer-classification filter"
xbe_json view trailers list --trailer-classification "1" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List trailers with --limit"
xbe_json view trailers list --limit 3
assert_success

test_name "List trailers with --offset"
xbe_json view trailers list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete trailer requires --confirm flag"
xbe_run do trailers delete "$CREATED_TRAILER_ID"
assert_failure

test_name "Delete trailer with --confirm"
# Create a trailer specifically for deletion
TRAILER_DEL_NUMBER=$(unique_name "DeleteTrailer")
xbe_json do trailers create \
    --number "$TRAILER_DEL_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "1"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do trailers delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create trailer for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create trailer without number fails"
xbe_json do trailers create --trucker "$CREATED_TRUCKER_ID" --trailer-classification "1"
assert_failure

test_name "Create trailer without trucker fails"
xbe_json do trailers create --number "NoTrucker" --trailer-classification "1"
assert_failure

test_name "Create trailer without trailer-classification fails"
xbe_json do trailers create --number "NoClass" --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do trailers update "$CREATED_TRAILER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
