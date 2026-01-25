#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Locations
#
# Tests CRUD operations for the project-transport-locations resource.
# Project transport locations represent pickup, delivery, and staging locations.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LOCATION_ID=""
CREATED_BROKER_ID=""
CREATED_PROJECT_OFFICE_ID=""
CREATED_EXT_ID_TYPE_ID=""
CREATED_EXT_ID=""

describe "Resource: project-transport-locations"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for project transport locations tests"
BROKER_NAME=$(unique_name "PTLTestBroker")

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

test_name "Create prerequisite project office for nearest project office filter"
PROJECT_OFFICE_NAME=$(unique_name "PTLOffice")

xbe_json do project-offices create --name "$PROJECT_OFFICE_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_OFFICE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
        register_cleanup "project-offices" "$CREATED_PROJECT_OFFICE_ID"
        pass
    else
        fail "Created project office but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_PROJECT_OFFICE_ID" ]]; then
        CREATED_PROJECT_OFFICE_ID="$XBE_TEST_PROJECT_OFFICE_ID"
        echo "    Using XBE_TEST_PROJECT_OFFICE_ID: $CREATED_PROJECT_OFFICE_ID"
        pass
    else
        skip "Failed to create project office and XBE_TEST_PROJECT_OFFICE_ID not set"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport location with required fields"
TEST_NAME=$(unique_name "TransportLocation")
EXT_TMS_ID=$(unique_name "TMS")

xbe_json do project-transport-locations create \
    --name "$TEST_NAME" \
    --geocoding-method explicit \
    --address-latitude 41.8781 \
    --address-longitude -87.6298 \
    --address-street-one "100 Test St" \
    --address-street-two "Suite 200" \
    --address-city "Chicago" \
    --address-state-code "IL" \
    --address-country-code "US" \
    --address-postal-code "60601" \
    --address-time-zone-id "America/Chicago" \
    --address-splc "123456789" \
    --external-tms-company-id "$EXT_TMS_ID" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_LOCATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
        register_cleanup "project-transport-locations" "$CREATED_LOCATION_ID"
        pass
    else
        fail "Created project transport location but no ID returned"
    fi
else
    fail "Failed to create project transport location"
fi

# Only continue if we successfully created a location
if [[ -z "$CREATED_LOCATION_ID" || "$CREATED_LOCATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid project transport location ID"
    run_tests
fi

test_name "Create project transport location with different geocoding method"
TEST_NAME2=$(unique_name "TransportLocation2")
xbe_json do project-transport-locations create \
    --name "$TEST_NAME2" \
    --geocoding-method none \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-transport-locations" "$id"
    pass
else
    fail "Failed to create project transport location with geocoding-method none"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport location attributes"
UPDATED_NAME=$(unique_name "UpdatedPTL")
UPDATED_EXT_TMS_ID=$(unique_name "TMS-UPD")
xbe_json do project-transport-locations update "$CREATED_LOCATION_ID" \
    --name "$UPDATED_NAME" \
    --geocoding-method none \
    --address-full "200 Updated Ave, Chicago, IL 60602" \
    --is-active=false \
    --is-valid-for-stop=false \
    --skip-detection \
    --external-tms-company-id "$UPDATED_EXT_TMS_ID"
assert_success

if [[ -n "${XBE_TEST_PROJECT_TRANSPORT_ORGANIZATION_ID:-}" ]]; then
    test_name "Update project transport location project transport organization"
    xbe_json do project-transport-locations update "$CREATED_LOCATION_ID" \
        --project-transport-organization "$XBE_TEST_PROJECT_TRANSPORT_ORGANIZATION_ID"
    assert_success
else
    test_name "Update project transport location project transport organization skipped"
    skip "XBE_TEST_PROJECT_TRANSPORT_ORGANIZATION_ID not set"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport location"
xbe_json view project-transport-locations show "$CREATED_LOCATION_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport locations"
xbe_json view project-transport-locations list --limit 5
assert_success

test_name "List project transport locations returns array"
xbe_json view project-transport-locations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport locations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport locations with --broker filter"
xbe_json view project-transport-locations list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List project transport locations with --name filter"
SEARCH_TERM="${UPDATED_NAME:0:12}"
xbe_json view project-transport-locations list --name "$SEARCH_TERM" --limit 10
assert_success

test_name "List project transport locations with --q filter"
xbe_json view project-transport-locations list --q "$SEARCH_TERM" --limit 10
assert_success

test_name "List project transport locations with --near filter"
xbe_json view project-transport-locations list --near "41.8781|-87.6298|25" --limit 10
assert_success

test_name "List project transport locations with --external-tms-company-id filter"
xbe_json view project-transport-locations list --external-tms-company-id "$UPDATED_EXT_TMS_ID" --limit 10
assert_success

if [[ -n "$CREATED_PROJECT_OFFICE_ID" ]]; then
    test_name "List project transport locations with --nearest-project-office-cached filter"
    xbe_json view project-transport-locations list --nearest-project-office-cached "$CREATED_PROJECT_OFFICE_ID" --limit 10
    assert_success
fi

if [[ -n "${XBE_TEST_PROJECT_TRANSPORT_ORGANIZATION_ID:-}" ]]; then
    test_name "List project transport locations with --project-transport-organization filter"
    xbe_json view project-transport-locations list --project-transport-organization "$XBE_TEST_PROJECT_TRANSPORT_ORGANIZATION_ID" --limit 10
    assert_success
fi

# ============================================================================
# LIST Tests - External Identification Filter
# ============================================================================

test_name "Create external identification type for project transport location"
EXT_ID_TYPE_NAME=$(unique_name "PTL-ExtIdType")

xbe_json do external-identification-types create \
    --name "$EXT_ID_TYPE_NAME" \
    --can-apply-to "ProjectTransportLocation"

if [[ $status -eq 0 ]]; then
    CREATED_EXT_ID_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_EXT_ID_TYPE_ID" && "$CREATED_EXT_ID_TYPE_ID" != "null" ]]; then
        register_cleanup "external-identification-types" "$CREATED_EXT_ID_TYPE_ID"
        pass
    else
        fail "Created external identification type but no ID returned"
    fi
else
    skip "Failed to create external identification type"
fi

if [[ -n "$CREATED_EXT_ID_TYPE_ID" && "$CREATED_EXT_ID_TYPE_ID" != "null" ]]; then
    test_name "Create external identification for project transport location"
    EXT_ID_VALUE=$(unique_name "PTL-EXT")
    xbe_json do external-identifications create \
        --external-identification-type "$CREATED_EXT_ID_TYPE_ID" \
        --identifies-type "project-transport-locations" \
        --identifies-id "$CREATED_LOCATION_ID" \
        --value "$EXT_ID_VALUE"

    if [[ $status -eq 0 ]]; then
        CREATED_EXT_ID=$(json_get ".id")
        if [[ -n "$CREATED_EXT_ID" && "$CREATED_EXT_ID" != "null" ]]; then
            register_cleanup "external-identifications" "$CREATED_EXT_ID"
            pass
        else
            fail "Created external identification but no ID returned"
        fi
    else
        fail "Failed to create external identification"
    fi

    if [[ -n "$CREATED_EXT_ID" && "$CREATED_EXT_ID" != "null" ]]; then
        test_name "List project transport locations with --external-identification-value filter"
        xbe_json view project-transport-locations list --external-identification-value "$EXT_ID_VALUE" --limit 10
        assert_success
    fi
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project transport locations with --limit"
xbe_json view project-transport-locations list --limit 3
assert_success

test_name "List project transport locations with --offset"
xbe_json view project-transport-locations list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport location requires --confirm flag"
xbe_run do project-transport-locations delete "$CREATED_LOCATION_ID"
assert_failure

test_name "Delete project transport location with --confirm"
TEST_DEL_NAME=$(unique_name "DeletePTL")
xbe_json do project-transport-locations create \
    --name "$TEST_DEL_NAME" \
    --geocoding-method none \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-transport-locations delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project transport location for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project transport location without geocoding method fails"
xbe_json do project-transport-locations create --name "NoGeo" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project transport location without broker fails"
xbe_json do project-transport-locations create --name "NoBroker" --geocoding-method explicit
assert_failure
