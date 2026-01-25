#!/bin/bash
#
# XBE CLI Integration Tests: Geofence Restrictions
#
# Tests CRUD operations for the geofence-restrictions resource.
# Geofence restrictions assign truckers to geofences and configure notification pacing.
#
# NOTE: This test requires creating a prerequisite broker, truckers, and geofences.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
SECOND_TRUCKER_ID=""
CREATED_GEOFENCE_ID=""
SECOND_GEOFENCE_ID=""
CREATED_RESTRICTION_ID=""

describe "Resource: geofence-restrictions"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for geofence restriction tests"
BROKER_NAME=$(unique_name "GeofenceRestrictionBroker")

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

# ============================================================================
# Prerequisites - Create truckers
# ============================================================================

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "GeofenceRestrictionTrucker")
xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "100 Restriction Lane"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
    fi
else
    fail "Failed to create trucker"
fi

if [[ -z "$CREATED_TRUCKER_ID" || "$CREATED_TRUCKER_ID" == "null" ]]; then
    echo "Cannot continue without a valid trucker ID"
    run_tests
fi

test_name "Create second prerequisite trucker"
TRUCKER_NAME2=$(unique_name "GeofenceRestrictionTrucker2")
xbe_json do truckers create \
    --name "$TRUCKER_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "200 Restriction Lane"

if [[ $status -eq 0 ]]; then
    SECOND_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$SECOND_TRUCKER_ID" && "$SECOND_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$SECOND_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
    fi
else
    fail "Failed to create second trucker"
fi

# ============================================================================
# Prerequisites - Create geofences
# ============================================================================

test_name "Create prerequisite geofence"
GEOFENCE_NAME=$(unique_name "RestrictionGeofence")
POLYGON='[[-122.4,37.8],[-122.4,37.7],[-122.3,37.7],[-122.3,37.8],[-122.4,37.8]]'

xbe_json do geofences create \
    --name "$GEOFENCE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --restriction-mode "custom_truckers" \
    --polygon-coordinates "$POLYGON"

if [[ $status -eq 0 ]]; then
    CREATED_GEOFENCE_ID=$(json_get ".id")
    if [[ -n "$CREATED_GEOFENCE_ID" && "$CREATED_GEOFENCE_ID" != "null" ]]; then
        register_cleanup "geofences" "$CREATED_GEOFENCE_ID"
        pass
    else
        fail "Created geofence but no ID returned"
    fi
else
    fail "Failed to create geofence"
fi

if [[ -z "$CREATED_GEOFENCE_ID" || "$CREATED_GEOFENCE_ID" == "null" ]]; then
    echo "Cannot continue without a valid geofence ID"
    run_tests
fi

test_name "Create second prerequisite geofence"
GEOFENCE_NAME2=$(unique_name "RestrictionGeofence2")
POLYGON2='[[-122.5,37.9],[-122.5,37.85],[-122.45,37.85],[-122.45,37.9],[-122.5,37.9]]'

xbe_json do geofences create \
    --name "$GEOFENCE_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --restriction-mode "custom_truckers" \
    --polygon-coordinates "$POLYGON2"

if [[ $status -eq 0 ]]; then
    SECOND_GEOFENCE_ID=$(json_get ".id")
    if [[ -n "$SECOND_GEOFENCE_ID" && "$SECOND_GEOFENCE_ID" != "null" ]]; then
        register_cleanup "geofences" "$SECOND_GEOFENCE_ID"
        pass
    else
        fail "Created geofence but no ID returned"
    fi
else
    fail "Failed to create second geofence"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create geofence restriction with required fields"
xbe_json do geofence-restrictions create \
    --geofence "$CREATED_GEOFENCE_ID" \
    --trucker "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_RESTRICTION_ID=$(json_get ".id")
    if [[ -n "$CREATED_RESTRICTION_ID" && "$CREATED_RESTRICTION_ID" != "null" ]]; then
        register_cleanup "geofence-restrictions" "$CREATED_RESTRICTION_ID"
        pass
    else
        fail "Created geofence restriction but no ID returned"
    fi
else
    fail "Failed to create geofence restriction"
fi

if [[ -z "$CREATED_RESTRICTION_ID" || "$CREATED_RESTRICTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid geofence restriction ID"
    run_tests
fi

test_name "Create geofence restriction with status and notification pacing"
xbe_json do geofence-restrictions create \
    --geofence "$SECOND_GEOFENCE_ID" \
    --trucker "$SECOND_TRUCKER_ID" \
    --status active \
    --max-seconds-between-notification 300
if [[ $status -eq 0 ]]; then
    SECOND_RESTRICTION_ID=$(json_get ".id")
    register_cleanup "geofence-restrictions" "$SECOND_RESTRICTION_ID"
    pass
else
    fail "Failed to create geofence restriction with optional fields"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update geofence restriction --status"
xbe_json do geofence-restrictions update "$CREATED_RESTRICTION_ID" --status inactive
assert_success

test_name "Update geofence restriction --max-seconds-between-notification"
xbe_json do geofence-restrictions update "$CREATED_RESTRICTION_ID" --max-seconds-between-notification 600
assert_success

test_name "Update geofence restriction --geofence"
xbe_json do geofence-restrictions update "$CREATED_RESTRICTION_ID" --geofence "$SECOND_GEOFENCE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List geofence restrictions"
xbe_json view geofence-restrictions list --limit 5
assert_success

test_name "List geofence restrictions returns array"
xbe_json view geofence-restrictions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list geofence restrictions"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List geofence restrictions with --geofence filter"
xbe_json view geofence-restrictions list --geofence "$SECOND_GEOFENCE_ID" --limit 10
assert_success

test_name "List geofence restrictions with --trucker filter"
xbe_json view geofence-restrictions list --trucker "$CREATED_TRUCKER_ID" --limit 10
assert_success

test_name "List geofence restrictions with --status filter"
xbe_json view geofence-restrictions list --status inactive --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete geofence restriction requires --confirm flag"
xbe_run do geofence-restrictions delete "$CREATED_RESTRICTION_ID"
assert_failure

test_name "Delete geofence restriction with --confirm"
DEL_RESTRICTION_GEOFENCE="$CREATED_GEOFENCE_ID"
DEL_RESTRICTION_TRUCKER="$SECOND_TRUCKER_ID"
xbe_json do geofence-restrictions create \
    --geofence "$DEL_RESTRICTION_GEOFENCE" \
    --trucker "$DEL_RESTRICTION_TRUCKER"

if [[ $status -eq 0 ]]; then
    DEL_RESTRICTION_ID=$(json_get ".id")
    if [[ -n "$DEL_RESTRICTION_ID" && "$DEL_RESTRICTION_ID" != "null" ]]; then
        xbe_run do geofence-restrictions delete "$DEL_RESTRICTION_ID" --confirm
        assert_success
    else
        skip "Could not create geofence restriction for deletion test"
    fi
else
    skip "Could not create geofence restriction for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create geofence restriction without --geofence fails"
xbe_json do geofence-restrictions create \
    --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create geofence restriction without --trucker fails"
xbe_json do geofence-restrictions create \
    --geofence "$CREATED_GEOFENCE_ID"
assert_failure

test_name "Update geofence restriction without any fields fails"
xbe_run do geofence-restrictions update "$CREATED_RESTRICTION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
