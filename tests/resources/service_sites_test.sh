#!/bin/bash
#
# XBE CLI Integration Tests: Service Sites
#
# Tests CRUD operations for the service-sites resource.
# Service sites are locations used for service work orders.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SERVICE_SITE_ID=""
CREATED_BROKER_ID=""

describe "Resource: service-sites"

# ============================================================================
# Prerequisites - Create broker for service site tests
# ============================================================================

test_name "Create prerequisite broker for service site tests"
BROKER_NAME=$(unique_name "ServiceSiteBroker")

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
# CREATE Tests
# ============================================================================

test_name "Create service site with required fields"
SITE_NAME=$(unique_name "ServiceSite")

xbe_json do service-sites create \
    --name "$SITE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --address "100 Service Site Ave, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_SERVICE_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_SERVICE_SITE_ID" && "$CREATED_SERVICE_SITE_ID" != "null" ]]; then
        register_cleanup "service-sites" "$CREATED_SERVICE_SITE_ID"
        pass
    else
        fail "Created service site but no ID returned"
    fi
else
    fail "Failed to create service site"
fi

if [[ -z "$CREATED_SERVICE_SITE_ID" || "$CREATED_SERVICE_SITE_ID" == "null" ]]; then
    echo "Cannot continue without a valid service site ID"
    run_tests
fi

test_name "Create service site with coordinates and skip-geocoding"
SITE_NAME2=$(unique_name "ServiceSiteCoord")

xbe_json do service-sites create \
    --name "$SITE_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --address "200 Service Site Ave, Chicago, IL 60601" \
    --address-latitude "41.8781" \
    --address-longitude "-87.6298" \
    --skip-geocoding

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "service-sites" "$id"
    pass
else
    fail "Failed to create service site with coordinates"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update service site name"
UPDATED_NAME=$(unique_name "UpdatedServiceSite")
xbe_json do service-sites update "$CREATED_SERVICE_SITE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update service site address"
xbe_json do service-sites update "$CREATED_SERVICE_SITE_ID" --address "300 Updated St, Chicago, IL 60602"
assert_success

test_name "Update service site coordinates with skip-geocoding"
xbe_json do service-sites update "$CREATED_SERVICE_SITE_ID" \
    --address-latitude "41.8800" \
    --address-longitude "-87.6300" \
    --skip-geocoding
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show service site"
xbe_json view service-sites show "$CREATED_SERVICE_SITE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List service sites"
xbe_json view service-sites list --limit 5
assert_success

test_name "List service sites returns array"
xbe_json view service-sites list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list service sites"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List service sites with --name filter"
xbe_json view service-sites list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List service sites with --broker filter"
xbe_json view service-sites list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List service sites with --limit"
xbe_json view service-sites list --limit 3
assert_success

test_name "List service sites with --offset"
xbe_json view service-sites list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete service site requires --confirm flag"
xbe_json do service-sites delete "$CREATED_SERVICE_SITE_ID"
assert_failure

test_name "Delete service site with --confirm"
SITE_DELETE_NAME=$(unique_name "DeleteServiceSite")
xbe_json do service-sites create \
    --name "$SITE_DELETE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --address "400 Delete St, Chicago, IL 60603"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    register_cleanup "service-sites" "$DEL_ID"
    xbe_run do service-sites delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create service site for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create service site without name fails"
xbe_json do service-sites create --broker "$CREATED_BROKER_ID" --address "123 Test St"
assert_failure

test_name "Create service site without broker fails"
xbe_json do service-sites create --name "Test Service Site" --address "123 Test St"
assert_failure

test_name "Create service site without address fails"
xbe_json do service-sites create --name "Test Service Site" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do service-sites update "$CREATED_SERVICE_SITE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
