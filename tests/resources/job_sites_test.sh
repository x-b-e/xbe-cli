#!/bin/bash
#
# XBE CLI Integration Tests: Job Sites
#
# Tests CRUD operations for the job_sites resource.
# Job sites are locations where work is performed.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_JOB_SITE_ID=""
CREATED_CUSTOMER_ID=""
CREATED_BROKER_ID=""

describe "Resource: job_sites"

# ============================================================================
# Prerequisites - Create broker and customer for job site tests
# ============================================================================

test_name "Create prerequisite broker for job site tests"
BROKER_NAME=$(unique_name "JSTestBroker")

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

test_name "Create prerequisite customer for job site tests"
CUSTOMER_NAME=$(unique_name "JSTestCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job site with required fields"
TEST_NAME=$(unique_name "JobSite")

xbe_json do job-sites create \
    --name "$TEST_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "100 Test Street, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID"
        pass
    else
        fail "Created job site but no ID returned"
    fi
else
    fail "Failed to create job site"
fi

# Only continue if we successfully created a job site
if [[ -z "$CREATED_JOB_SITE_ID" || "$CREATED_JOB_SITE_ID" == "null" ]]; then
    echo "Cannot continue without a valid job site ID"
    run_tests
fi

test_name "Create job site with notes"
TEST_NAME2=$(unique_name "JobSite2")
xbe_json do job-sites create \
    --name "$TEST_NAME2" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "101 Test Street, Chicago, IL 60601" \
    --notes "Test notes for job site"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with notes"
fi

test_name "Create job site with phone-number"
TEST_NAME3=$(unique_name "JobSite3")
TEST_PHONE=$(unique_mobile)
xbe_json do job-sites create \
    --name "$TEST_NAME3" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "102 Test Street, Chicago, IL 60601" \
    --phone-number "$TEST_PHONE"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with phone-number"
fi

test_name "Create job site with contact-name"
TEST_NAME4=$(unique_name "JobSite4")
xbe_json do job-sites create \
    --name "$TEST_NAME4" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "103 Test Street, Chicago, IL 60601" \
    --contact-name "John Contact"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with contact-name"
fi

test_name "Create job site with active=false"
TEST_NAME5=$(unique_name "JobSite5")
xbe_json do job-sites create \
    --name "$TEST_NAME5" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "104 Test Street, Chicago, IL 60601" \
    --active=false
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with active=false"
fi

test_name "Create job site with default-time-card-approval-process"
TEST_NAME6=$(unique_name "JobSite6")
xbe_json do job-sites create \
    --name "$TEST_NAME6" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "105 Test Street, Chicago, IL 60601" \
    --default-time-card-approval-process "admin"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with default-time-card-approval-process"
fi

test_name "Create job site with address"
TEST_NAME7=$(unique_name "JobSite7")
xbe_json do job-sites create \
    --name "$TEST_NAME7" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "123 Test Street, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with address"
fi

test_name "Create job site with coordinates and skip-geocoding"
TEST_NAME8=$(unique_name "JobSite8")
xbe_json do job-sites create \
    --name "$TEST_NAME8" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "456 Manual Coord St" \
    --address-latitude "41.8781" \
    --address-longitude "-87.6298" \
    --skip-geocoding
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "job-sites" "$id"
    pass
else
    fail "Failed to create job site with coordinates"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update job site name"
UPDATED_NAME=$(unique_name "UpdatedJS")
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update job site notes"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --notes "Updated notes"
assert_success

test_name "Update job site phone-number"
UPDATED_PHONE=$(unique_mobile)
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --phone-number "$UPDATED_PHONE"
assert_success

test_name "Update job site contact-name"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --contact-name "Updated Contact"
assert_success

test_name "Update job site to inactive"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --active=false
assert_success

test_name "Update job site to active"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --active
assert_success

test_name "Update job site default-time-card-approval-process"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --default-time-card-approval-process "field"
assert_success

test_name "Update job site address"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" --address "789 Updated Ave, Chicago, IL 60602"
assert_success

test_name "Update job site coordinates with skip-geocoding"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID" \
    --address-latitude "41.8800" \
    --address-longitude "-87.6300" \
    --skip-geocoding
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job sites"
xbe_json view job-sites list --limit 5
assert_success

test_name "List job sites returns array"
xbe_json view job-sites list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job sites"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List job sites with --name filter"
xbe_json view job-sites list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List job sites with --name-like filter"
xbe_json view job-sites list --name-like "JobSite" --limit 10
assert_success

test_name "List job sites with --active filter"
xbe_json view job-sites list --active --limit 10
assert_success

test_name "List job sites with --broker filter"
xbe_json view job-sites list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List job sites with --customer filter"
xbe_json view job-sites list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List job sites with --q filter"
xbe_json view job-sites list --q "JobSite" --limit 10
assert_success

test_name "List job sites with --has-material-site filter"
xbe_json view job-sites list --has-material-site true --limit 10
assert_success

test_name "List job sites with --is-stockpiling filter"
xbe_json view job-sites list --is-stockpiling true --limit 10
assert_success

test_name "List job sites with --active-since filter"
xbe_json view job-sites list --active-since "2020-01-01" --limit 10
assert_success

test_name "List job sites with --broker-id filter"
xbe_json view job-sites list --broker-id "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List job sites with --external-identification-value filter"
xbe_json view job-sites list --external-identification-value "EXT123" --limit 10
assert_success

test_name "List job sites with --external-job-number filter"
xbe_json view job-sites list --external-job-number "JOB123" --limit 10
assert_success

# Note: --address-near requires lat,lng,radius format - skip if complex
# Note: --material-site requires a valid material site ID - skip without prerequisite

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List job sites with --limit"
xbe_json view job-sites list --limit 3
assert_success

test_name "List job sites with --offset"
xbe_json view job-sites list --limit 3 --offset 3
assert_success

test_name "List job sites with pagination (limit + offset)"
xbe_json view job-sites list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job site requires --confirm flag"
xbe_json do job-sites delete "$CREATED_JOB_SITE_ID"
assert_failure

test_name "Delete job site with --confirm"
# Create a job site specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMeJS")
xbe_json do job-sites create \
    --name "$TEST_DEL_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "999 Delete Street, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do job-sites delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create job site for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job site without name fails"
xbe_json do job-sites create --customer "$CREATED_CUSTOMER_ID" --address "123 Test St"
assert_failure

test_name "Create job site without customer fails"
xbe_json do job-sites create --name "Test Job Site" --address "123 Test St"
assert_failure

test_name "Create job site without address fails"
xbe_json do job-sites create --name "Test Job Site" --customer "$CREATED_CUSTOMER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do job-sites update "$CREATED_JOB_SITE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
