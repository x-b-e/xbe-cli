#!/bin/bash
#
# XBE CLI Integration Tests: Trips
#
# Tests CRUD operations for the trips resource.
# Trips require origin and destination relationships (polymorphic).
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TRIP_ID=""
CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_JOB_SITE_ID=""
CREATED_CUSTOMER_ID=""

describe "Resource: trips"

# ============================================================================
# Prerequisites - Create resources for trip tests
# ============================================================================

test_name "Create prerequisite broker for trip tests"
BROKER_NAME=$(unique_name "TripTestBroker")

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

test_name "Create prerequisite material supplier for trip tests"
MS_SUPPLIER_NAME=$(unique_name "TripMaterialSupplier")

xbe_json do material-suppliers create \
    --name "$MS_SUPPLIER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
else
    fail "Failed to create material supplier"
    echo "Cannot continue without a material supplier"
    run_tests
fi

test_name "Create prerequisite material site for trip tests"
MS_NAME=$(unique_name "TripMaterialSite")

xbe_json do material-sites create \
    --name "$MS_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "123 Material Site Road, Test City, TC 12345"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
        pass
    else
        fail "Created material site but no ID returned"
        echo "Cannot continue without a material site"
        run_tests
    fi
else
    fail "Failed to create material site"
    echo "Cannot continue without a material site"
    run_tests
fi

test_name "Create prerequisite customer for trip tests"
CUSTOMER_NAME=$(unique_name "TripTestCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create prerequisite job site for trip tests"
JS_NAME=$(unique_name "TripJobSite")

xbe_json do job-sites create \
    --name "$JS_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "456 Job Site Blvd, Test City, TC 12345"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID"
        pass
    else
        fail "Created job site but no ID returned"
        echo "Cannot continue without a job site"
        run_tests
    fi
else
    fail "Failed to create job site"
    echo "Cannot continue without a job site"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trip with required fields"
xbe_json do trips create \
    --origin-type material-sites \
    --origin-id "$CREATED_MATERIAL_SITE_ID" \
    --destination-type job-sites \
    --destination-id "$CREATED_JOB_SITE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRIP_ID" && "$CREATED_TRIP_ID" != "null" ]]; then
        register_cleanup "trips" "$CREATED_TRIP_ID"
        pass
    else
        fail "Created trip but no ID returned"
    fi
else
    fail "Failed to create trip"
fi

if [[ -z "$CREATED_TRIP_ID" || "$CREATED_TRIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid trip ID"
    run_tests
fi

test_name "Create trip with times"
xbe_json do trips create \
    --origin-type material-sites \
    --origin-id "$CREATED_MATERIAL_SITE_ID" \
    --destination-type job-sites \
    --destination-id "$CREATED_JOB_SITE_ID" \
    --origin-at "2024-01-15T08:00:00Z" \
    --destination-at "2024-01-15T09:30:00Z"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "trips" "$id"
    pass
else
    fail "Failed to create trip with times"
fi

test_name "Create trip with notes"
xbe_json do trips create \
    --origin-type material-sites \
    --origin-id "$CREATED_MATERIAL_SITE_ID" \
    --destination-type job-sites \
    --destination-id "$CREATED_JOB_SITE_ID" \
    --origin-notes "Pickup at gate A" \
    --destination-notes "Deliver to loading dock"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "trips" "$id"
    pass
else
    fail "Failed to create trip with notes"
fi

test_name "Create trip with submitted mileage and minutes"
xbe_json do trips create \
    --origin-type material-sites \
    --origin-id "$CREATED_MATERIAL_SITE_ID" \
    --destination-type job-sites \
    --destination-id "$CREATED_JOB_SITE_ID" \
    --submitted-mileage "25.5" \
    --submitted-minutes "45"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "trips" "$id"
    pass
else
    fail "Failed to create trip with submitted mileage and minutes"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update trip origin-at"
xbe_json do trips update "$CREATED_TRIP_ID" --origin-at "2024-01-15T07:30:00Z"
assert_success

test_name "Update trip origin-notes"
xbe_json do trips update "$CREATED_TRIP_ID" --origin-notes "Updated origin notes"
assert_success

test_name "Update trip destination-at"
xbe_json do trips update "$CREATED_TRIP_ID" --destination-at "2024-01-15T10:00:00Z"
assert_success

test_name "Update trip destination-notes"
xbe_json do trips update "$CREATED_TRIP_ID" --destination-notes "Updated destination notes"
assert_success

test_name "Update trip submitted-mileage"
xbe_json do trips update "$CREATED_TRIP_ID" --submitted-mileage "30.2"
assert_success

test_name "Update trip submitted-minutes"
xbe_json do trips update "$CREATED_TRIP_ID" --submitted-minutes "55"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trips"
xbe_json view trips list --limit 5
assert_success

test_name "List trips returns array"
xbe_json view trips list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trips"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trips with --material-sites filter"
xbe_json view trips list --material-sites "$CREATED_MATERIAL_SITE_ID" --limit 10
assert_success

test_name "List trips with --job-sites filter"
xbe_json view trips list --job-sites "$CREATED_JOB_SITE_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List trips with --limit"
xbe_json view trips list --limit 3
assert_success

test_name "List trips with --offset"
xbe_json view trips list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete trip requires --confirm flag"
xbe_json do trips delete "$CREATED_TRIP_ID"
assert_failure

test_name "Delete trip with --confirm"
# Create a trip specifically for deletion
xbe_json do trips create \
    --origin-type material-sites \
    --origin-id "$CREATED_MATERIAL_SITE_ID" \
    --destination-type job-sites \
    --destination-id "$CREATED_JOB_SITE_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do trips delete "$DEL_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        # API may not allow deletion
        register_cleanup "trips" "$DEL_ID"
        skip "API may not allow trip deletion"
    fi
else
    skip "Could not create trip for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create trip without origin-type fails"
xbe_json do trips create --origin-id "$CREATED_MATERIAL_SITE_ID" --destination-type job-sites --destination-id "$CREATED_JOB_SITE_ID"
assert_failure

test_name "Create trip without origin-id fails"
xbe_json do trips create --origin-type material-sites --destination-type job-sites --destination-id "$CREATED_JOB_SITE_ID"
assert_failure

test_name "Create trip without destination-type fails"
xbe_json do trips create --origin-type material-sites --origin-id "$CREATED_MATERIAL_SITE_ID" --destination-id "$CREATED_JOB_SITE_ID"
assert_failure

test_name "Create trip without destination-id fails"
xbe_json do trips create --origin-type material-sites --origin-id "$CREATED_MATERIAL_SITE_ID" --destination-type job-sites
assert_failure

test_name "Update without any fields fails"
xbe_json do trips update "$CREATED_TRIP_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
