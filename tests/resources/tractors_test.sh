#!/bin/bash
#
# XBE CLI Integration Tests: Tractors
#
# Tests CRUD operations for the tractors resource.
# Tractors are the power units (trucks) in a trucking fleet.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TRACTOR_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""

describe "Resource: tractors"

# ============================================================================
# Prerequisites - Create broker and trucker
# ============================================================================

test_name "Create prerequisite broker for tractors tests"
BROKER_NAME=$(unique_name "TractorTestBroker")

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
TRUCKER_NAME=$(unique_name "TractorTestTrucker")
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

test_name "Create tractor with required fields"
TRACTOR_NUMBER=$(unique_name "Tractor")

xbe_json do tractors create \
    --number "$TRACTOR_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRACTOR_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRACTOR_ID" && "$CREATED_TRACTOR_ID" != "null" ]]; then
        register_cleanup "tractors" "$CREATED_TRACTOR_ID"
        pass
    else
        fail "Created tractor but no ID returned"
    fi
else
    fail "Failed to create tractor"
fi

# Only continue if we successfully created a tractor
if [[ -z "$CREATED_TRACTOR_ID" || "$CREATED_TRACTOR_ID" == "null" ]]; then
    echo "Cannot continue without a valid tractor ID"
    run_tests
fi

test_name "Create tractor with truck details"
TRACTOR_NUMBER2=$(unique_name "Tractor2")
xbe_json do tractors create \
    --number "$TRACTOR_NUMBER2" \
    --trucker "$CREATED_TRUCKER_ID" \
    --truck-manufacturer-name "Peterbilt" \
    --truck-model-name "579" \
    --truck-model-year 2023
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tractors" "$id"
    pass
else
    fail "Failed to create tractor with truck details"
fi

test_name "Create tractor with registration info"
TRACTOR_NUMBER3=$(unique_name "Tractor3")
xbe_json do tractors create \
    --number "$TRACTOR_NUMBER3" \
    --trucker "$CREATED_TRUCKER_ID" \
    --plate-number "ABC123" \
    --plate-jurisdiction-code "IL" \
    --vin "1HGCM82633A123456"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tractors" "$id"
    pass
else
    fail "Failed to create tractor with registration info"
fi

test_name "Create tractor with in-service status"
TRACTOR_NUMBER4=$(unique_name "Tractor4")
xbe_json do tractors create \
    --number "$TRACTOR_NUMBER4" \
    --trucker "$CREATED_TRUCKER_ID" \
    --in-service true
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tractors" "$id"
    pass
else
    fail "Failed to create tractor with in-service status"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update tractor number"
UPDATED_NUMBER=$(unique_name "UpdatedTractor")
xbe_json do tractors update "$CREATED_TRACTOR_ID" --number "$UPDATED_NUMBER"
assert_success

test_name "Update tractor truck-manufacturer-name"
xbe_json do tractors update "$CREATED_TRACTOR_ID" --truck-manufacturer-name "Freightliner"
assert_success

test_name "Update tractor truck-model-name"
xbe_json do tractors update "$CREATED_TRACTOR_ID" --truck-model-name "Cascadia"
assert_success

test_name "Update tractor truck-model-year"
xbe_json do tractors update "$CREATED_TRACTOR_ID" --truck-model-year 2024
assert_success

test_name "Update tractor color-name"
xbe_json do tractors update "$CREATED_TRACTOR_ID" --color-name "blue"
assert_success

test_name "Update tractor plate-number"
xbe_json do tractors update "$CREATED_TRACTOR_ID" --plate-number "XYZ789"
assert_success

test_name "Update tractor in-service"
xbe_json do tractors update "$CREATED_TRACTOR_ID" --in-service true
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tractors"
xbe_json view tractors list --limit 5
assert_success

test_name "List tractors returns array"
xbe_json view tractors list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tractors"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List tractors with --trucker filter"
xbe_json view tractors list --trucker "$CREATED_TRUCKER_ID" --limit 10
assert_success

test_name "List tractors with --in-service filter"
xbe_json view tractors list --in-service true --limit 10
assert_success

test_name "List tractors with --number-like filter"
xbe_json view tractors list --number-like "Tractor" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List tractors with --limit"
xbe_json view tractors list --limit 3
assert_success

test_name "List tractors with --offset"
xbe_json view tractors list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete tractor requires --confirm flag"
xbe_run do tractors delete "$CREATED_TRACTOR_ID"
assert_failure

test_name "Delete tractor with --confirm"
# Create a tractor specifically for deletion
TRACTOR_DEL_NUMBER=$(unique_name "DeleteTractor")
xbe_json do tractors create \
    --number "$TRACTOR_DEL_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do tractors delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create tractor for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tractor without number fails"
xbe_json do tractors create --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create tractor without trucker fails"
xbe_json do tractors create --number "NoTrucker"
assert_failure

test_name "Update without any fields fails"
xbe_json do tractors update "$CREATED_TRACTOR_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
