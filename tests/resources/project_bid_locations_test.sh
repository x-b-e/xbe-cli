#!/bin/bash
#
# XBE CLI Integration Tests: Project Bid Locations
#
# Tests CRUD operations for the project-bid-locations resource.
# Project bid locations require a project relationship and geometry.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_BID_LOCATION_ID=""

GEOMETRY_POINT="POINT(-77.0365 38.8977)"
GEOMETRY_LINE="LINESTRING(-77.0365 38.8977,-77.0400 38.9000)"
NEAR_FILTER="38.8977|-77.0365|5"


describe "Resource: project-bid-locations"

# ============================================================================
# Prerequisites - Create broker, developer, project
# ============================================================================

test_name "Create prerequisite broker for project bid location tests"
BROKER_NAME=$(unique_name "BidLocationBroker")

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

test_name "Create prerequisite developer for project bid location tests"
DEV_NAME=$(unique_name "BidLocationDeveloper")

xbe_json do developers create --name "$DEV_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

test_name "Create prerequisite project for project bid location tests"
PROJECT_NAME=$(unique_name "BidLocationProject")

xbe_json do projects create \
    --name "$PROJECT_NAME" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        echo "Cannot continue without a project"
        run_tests
    fi
else
    fail "Failed to create project"
    echo "Cannot continue without a project"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project bid location with required fields"
LOCATION_NAME=$(unique_name "BidLocation")

xbe_json do project-bid-locations create \
    --project "$CREATED_PROJECT_ID" \
    --geometry "$GEOMETRY_POINT" \
    --name "$LOCATION_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_BID_LOCATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_BID_LOCATION_ID" && "$CREATED_PROJECT_BID_LOCATION_ID" != "null" ]]; then
        register_cleanup "project-bid-locations" "$CREATED_PROJECT_BID_LOCATION_ID"
        pass
    else
        fail "Created project bid location but no ID returned"
    fi
else
    fail "Failed to create project bid location"
fi

if [[ -z "$CREATED_PROJECT_BID_LOCATION_ID" || "$CREATED_PROJECT_BID_LOCATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid project bid location ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project bid location by ID"
xbe_json view project-bid-locations show "$CREATED_PROJECT_BID_LOCATION_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project bid locations"
xbe_json view project-bid-locations list --limit 5
assert_success

test_name "List project bid locations returns array"
xbe_json view project-bid-locations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project bid locations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project bid locations with --project filter"
xbe_json view project-bid-locations list --project "$CREATED_PROJECT_ID" --limit 10
assert_success

test_name "List project bid locations with --near filter"
xbe_json view project-bid-locations list --near "$NEAR_FILTER" --limit 10
assert_success

test_name "List project bid locations with --state-code filter"
xbe_json view project-bid-locations list --state-code "DC" --limit 10
assert_success

test_name "List project bid locations with --county filter"
xbe_json view project-bid-locations list --county "District of Columbia" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project bid locations with --limit"
xbe_json view project-bid-locations list --limit 3
assert_success

test_name "List project bid locations with --offset"
xbe_json view project-bid-locations list --limit 3 --offset 1
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project bid location name"
UPDATED_NAME=$(unique_name "BidLocationUpdated")
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update project bid location notes"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --notes "Updated notes"
assert_success

test_name "Update project bid location geometry"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --geometry "$GEOMETRY_LINE"
assert_success

test_name "Update project bid location address"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --address "1600 Pennsylvania Ave NW, Washington, DC"
assert_success

test_name "Update project bid location address latitude"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --address-latitude "38.8977"
assert_success

test_name "Update project bid location address longitude"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --address-longitude "-77.0365"
assert_success

test_name "Update project bid location address place ID"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --address-place-id "test-place-id"
assert_success

test_name "Update project bid location address plus code"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --address-plus-code "87C4VXX7+F5"
assert_success

test_name "Update project bid location skip address geocoding"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID" --skip-address-geocoding
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project bid location requires --confirm flag"
xbe_run do project-bid-locations delete "$CREATED_PROJECT_BID_LOCATION_ID"
assert_failure

test_name "Delete project bid location with --confirm"
DELETE_NAME=$(unique_name "BidLocationDelete")
xbe_json do project-bid-locations create \
    --project "$CREATED_PROJECT_ID" \
    --geometry "$GEOMETRY_POINT" \
    --name "$DELETE_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-bid-locations delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project bid location for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project bid location without project fails"
xbe_json do project-bid-locations create --geometry "$GEOMETRY_POINT"
assert_failure

test_name "Create project bid location without geometry fails"
xbe_json do project-bid-locations create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-bid-locations update "$CREATED_PROJECT_BID_LOCATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
