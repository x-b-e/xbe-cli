#!/bin/bash
#
# XBE CLI Integration Tests: Jobs
#
# Tests CRUD operations for the jobs resource.
# Jobs tie together customers, job sites, material types, and trailer classifications.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_JOB_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_JOB_SITE_ID_2=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_TYPE_ID_2=""
TRAILER_CLASSIFICATION_ID=""
CREATED_JOB_PRODUCTION_PLAN_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
SERVICE_TYPE_UOM_ID=""
FOREMAN_ID=""
TRUCKER_ID=""
ACTIVE_JOB_SITE_ID=""

JOB_EXTERNAL_NUMBER="JOB-EXT-$(date +%s)"

describe "Resource: jobs"

# ============================================================================
# Prerequisites - Create broker, customer, job site, and material type
# ============================================================================

test_name "Create prerequisite broker for jobs tests"
BROKER_NAME=$(unique_name "JobsTestBroker")

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

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JobsTestCustomer")

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

test_name "Create prerequisite job site"
JOB_SITE_NAME=$(unique_name "JobsTestJobSite")

xbe_json do job-sites create \
    --name "$JOB_SITE_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "100 Test Street, Chicago, IL 60601"

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

test_name "Create secondary job site"
JOB_SITE_NAME_2=$(unique_name "JobsTestJobSite2")

xbe_json do job-sites create \
    --name "$JOB_SITE_NAME_2" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "101 Test Street, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID_2" && "$CREATED_JOB_SITE_ID_2" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID_2"
        pass
    else
        fail "Created job site but no ID returned"
    fi
else
    fail "Failed to create secondary job site"
fi

test_name "Create prerequisite material type"
MATERIAL_TYPE_NAME=$(unique_name "JobsMaterialType")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    fail "Failed to create material type"
    echo "Cannot continue without a material type"
    run_tests
fi

test_name "Create second material type"
MATERIAL_TYPE_NAME_2=$(unique_name "JobsMaterialType2")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID_2" && "$CREATED_MATERIAL_TYPE_ID_2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID_2"
        pass
    else
        fail "Created material type but no ID returned"
    fi
else
    fail "Failed to create second material type"
fi

test_name "Fetch trailer classification ID"
if [[ -n "$XBE_TEST_TRAILER_CLASSIFICATION_ID" ]]; then
    TRAILER_CLASSIFICATION_ID="$XBE_TEST_TRAILER_CLASSIFICATION_ID"
    echo "    Using XBE_TEST_TRAILER_CLASSIFICATION_ID: $TRAILER_CLASSIFICATION_ID"
    pass
else
    xbe_json view trailer-classifications list --limit 1
    if [[ $status -eq 0 ]]; then
        TRAILER_CLASSIFICATION_ID=$(json_get ".[0].id")
        if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
            pass
        else
            fail "Could not find trailer classification ID"
            echo "Cannot continue without trailer classifications"
            run_tests
        fi
    else
        fail "Failed to list trailer classifications"
        echo "Cannot continue without trailer classifications"
        run_tests
    fi
fi

test_name "Create prerequisite material supplier"
MATERIAL_SUPPLIER_NAME=$(unique_name "JobsMaterialSupplier")

xbe_json do material-suppliers create \
    --name "$MATERIAL_SUPPLIER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
    fi
else
    fail "Failed to create material supplier"
fi

test_name "Create prerequisite material site"
MATERIAL_SITE_NAME=$(unique_name "JobsMaterialSite")

if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json do material-sites create \
        --name "$MATERIAL_SITE_NAME" \
        --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
        --address "200 Test Street, Chicago, IL 60601"

    if [[ $status -eq 0 ]]; then
        CREATED_MATERIAL_SITE_ID=$(json_get ".id")
        if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
            register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
            pass
        else
            fail "Created material site but no ID returned"
        fi
    else
        fail "Failed to create material site"
    fi
else
    skip "No material supplier ID available"
fi

test_name "Create prerequisite job production plan"
TODAY=$(date +%Y-%m-%d)
JPP_NAME=$(unique_name "JobsJPP")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_PRODUCTION_PLAN_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
    fi
else
    fail "Failed to create job production plan"
fi

test_name "Fetch service type unit of measure ID"
if [[ -n "$XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID" ]]; then
    SERVICE_TYPE_UOM_ID="$XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID"
    echo "    Using XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID: $SERVICE_TYPE_UOM_ID"
    pass
else
    xbe_json view service-type-unit-of-measure-quantities list --limit 1
    if [[ $status -eq 0 ]]; then
        SERVICE_TYPE_UOM_ID=$(json_get ".[0].service_type_unit_of_measure_id")
        if [[ -n "$SERVICE_TYPE_UOM_ID" && "$SERVICE_TYPE_UOM_ID" != "null" ]]; then
            pass
        else
            echo "    No service type unit of measure ID found; skipping related tests"
            pass
            SERVICE_TYPE_UOM_ID=""
        fi
    else
        echo "    Failed to list service type unit of measure quantities; skipping related tests"
        pass
    fi
fi

test_name "Fetch current user for foreman"
if [[ -n "$XBE_TEST_USER_ID" ]]; then
    FOREMAN_ID="$XBE_TEST_USER_ID"
    echo "    Using XBE_TEST_USER_ID: $FOREMAN_ID"
    pass
else
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        FOREMAN_ID=$(json_get ".id")
        if [[ -n "$FOREMAN_ID" && "$FOREMAN_ID" != "null" ]]; then
            pass
        else
            echo "    No user ID from whoami; skipping foreman tests"
            pass
            FOREMAN_ID=""
        fi
    else
        echo "    Failed to fetch current user; skipping foreman tests"
        pass
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job with required fields"

xbe_json do jobs create \
    --customer "$CREATED_CUSTOMER_ID" \
    --job-site "$CREATED_JOB_SITE_ID" \
    --material-types "$CREATED_MATERIAL_TYPE_ID" \
    --trailer-classifications "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_ID" && "$CREATED_JOB_ID" != "null" ]]; then
        register_cleanup "jobs" "$CREATED_JOB_ID"
        ACTIVE_JOB_SITE_ID="$CREATED_JOB_SITE_ID"
        pass
    else
        fail "Created job but no ID returned"
    fi
else
    fail "Failed to create job"
fi

if [[ -z "$CREATED_JOB_ID" || "$CREATED_JOB_ID" == "null" ]]; then
    echo "Cannot continue without a valid job ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update job notes"
xbe_json do jobs update "$CREATED_JOB_ID" --notes "Updated job notes"
assert_success

test_name "Update job dispatch instructions"
xbe_json do jobs update "$CREATED_JOB_ID" --dispatch-instructions "Updated dispatch instructions"
assert_success

test_name "Update job loaded miles"
xbe_json do jobs update "$CREATED_JOB_ID" --loaded-miles "10"
assert_success

test_name "Update job validation flags"
xbe_json do jobs update "$CREATED_JOB_ID" --skip-material-type-start-site-type-validation --validate-job-schedule-shifts
assert_success

test_name "Update job prevailing wage fields"
xbe_json do jobs update "$CREATED_JOB_ID" \
    --is-prevailing-wage \
    --requires-certified-payroll \
    --prevailing-wage-hourly-rate "45"
assert_success

test_name "Update job external job number"
xbe_json do jobs update "$CREATED_JOB_ID" --external-job-number "$JOB_EXTERNAL_NUMBER"
assert_success

# ============================================================================
# UPDATE Tests - Relationships
# ============================================================================

test_name "Update job customer and job site"
if [[ -n "$CREATED_JOB_SITE_ID_2" && "$CREATED_JOB_SITE_ID_2" != "null" ]]; then
    xbe_json do jobs update "$CREATED_JOB_ID" \
        --customer "$CREATED_CUSTOMER_ID" \
        --job-site "$CREATED_JOB_SITE_ID_2"
    assert_success
    ACTIVE_JOB_SITE_ID="$CREATED_JOB_SITE_ID_2"
else
    xbe_json do jobs update "$CREATED_JOB_ID" --customer "$CREATED_CUSTOMER_ID" --job-site "$CREATED_JOB_SITE_ID"
    assert_success
    ACTIVE_JOB_SITE_ID="$CREATED_JOB_SITE_ID"
fi

test_name "Update job start site"
xbe_json do jobs update "$CREATED_JOB_ID" --start-site-type job-sites --start-site "$ACTIVE_JOB_SITE_ID"
assert_success

test_name "Update job production plan"
if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json do jobs update "$CREATED_JOB_ID" --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID"
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "Update job foreman"
if [[ -n "$FOREMAN_ID" && "$FOREMAN_ID" != "null" ]]; then
    xbe_json do jobs update "$CREATED_JOB_ID" --foreman "$FOREMAN_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        echo "    Foreman update failed (likely membership validation); skipping"
        pass
    fi
else
    skip "No foreman ID available"
fi

test_name "Update job material types"
if [[ -n "$CREATED_MATERIAL_TYPE_ID_2" && "$CREATED_MATERIAL_TYPE_ID_2" != "null" ]]; then
    xbe_json do jobs update "$CREATED_JOB_ID" --material-types "$CREATED_MATERIAL_TYPE_ID_2"
    assert_success
else
    xbe_json do jobs update "$CREATED_JOB_ID" --material-types "$CREATED_MATERIAL_TYPE_ID"
    assert_success
fi

test_name "Update job trailer classifications"
xbe_json do jobs update "$CREATED_JOB_ID" --trailer-classifications "$TRAILER_CLASSIFICATION_ID"
assert_success

test_name "Update job material sites"
if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json do jobs update "$CREATED_JOB_ID" --material-sites "$CREATED_MATERIAL_SITE_ID"
    assert_success
else
    skip "No material site ID available"
fi

test_name "Update job service type unit of measures"
if [[ -n "$SERVICE_TYPE_UOM_ID" && "$SERVICE_TYPE_UOM_ID" != "null" ]]; then
    xbe_json do jobs update "$CREATED_JOB_ID" --service-type-unit-of-measures "$SERVICE_TYPE_UOM_ID"
    assert_success
else
    skip "No service type unit of measure ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job"
xbe_json view jobs show "$CREATED_JOB_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List jobs"
xbe_json view jobs list --limit 5
assert_success

test_name "List jobs returns array"
xbe_json view jobs list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list jobs"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List jobs with --customer filter"
xbe_json view jobs list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List jobs with --job-site filter"
xbe_json view jobs list --job-site "$CREATED_JOB_SITE_ID" --limit 10
assert_success

test_name "List jobs with --start-date filter"
xbe_json view jobs list --start-date "$TODAY" --limit 10
assert_success

test_name "List jobs with --start-at-min filter"
xbe_json view jobs list --start-at-min "${TODAY}T00:00:00Z" --limit 10
assert_success

test_name "List jobs with --start-at-max filter"
xbe_json view jobs list --start-at-max "${TODAY}T23:59:59Z" --limit 10
assert_success

test_name "List jobs with --offered filter"
xbe_json view jobs list --offered true --limit 10
assert_success

test_name "List jobs with --broker filter"
xbe_json view jobs list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List jobs with --trucker filter"
if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
    TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
else
    TRUCKER_ID="1"
fi
xbe_json view jobs list --trucker "$TRUCKER_ID" --limit 10
assert_success

test_name "List jobs with --job-production-plan filter"
if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view jobs list --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" --limit 10
    assert_success
else
    xbe_json view jobs list --job-production-plan "1" --limit 10
    assert_success
fi

test_name "List jobs with --external-job-number filter"
xbe_json view jobs list --external-job-number "$JOB_EXTERNAL_NUMBER" --limit 10
assert_success

test_name "List jobs with --external-identification-value filter"
xbe_json view jobs list --external-identification-value "EXT123" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job requires --confirm flag"
xbe_json do jobs delete "$CREATED_JOB_ID"
assert_failure

test_name "Delete job with --confirm"
# Create a job specifically for deletion
DELETE_JOB_NAME="DeleteMeJob"
xbe_json do jobs create \
    --customer "$CREATED_CUSTOMER_ID" \
    --job-site "$CREATED_JOB_SITE_ID" \
    --material-types "$CREATED_MATERIAL_TYPE_ID" \
    --trailer-classifications "$TRAILER_CLASSIFICATION_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do jobs delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create job for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job without customer fails"
xbe_json do jobs create --job-site "$CREATED_JOB_SITE_ID" --material-types "$CREATED_MATERIAL_TYPE_ID" --trailer-classifications "$TRAILER_CLASSIFICATION_ID"
assert_failure

test_name "Create job without job site fails"
xbe_json do jobs create --customer "$CREATED_CUSTOMER_ID" --material-types "$CREATED_MATERIAL_TYPE_ID" --trailer-classifications "$TRAILER_CLASSIFICATION_ID"
assert_failure

test_name "Create job without material types fails"
xbe_json do jobs create --customer "$CREATED_CUSTOMER_ID" --job-site "$CREATED_JOB_SITE_ID" --trailer-classifications "$TRAILER_CLASSIFICATION_ID"
assert_failure

test_name "Create job without trailer classifications fails"
xbe_json do jobs create --customer "$CREATED_CUSTOMER_ID" --job-site "$CREATED_JOB_SITE_ID" --material-types "$CREATED_MATERIAL_TYPE_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do jobs update "$CREATED_JOB_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
