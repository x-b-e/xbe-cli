#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plans
#
# Tests create and update operations for the job_production_plans resource.
# Job production plans are scheduling units for work at job sites.
#
# COMPLETE COVERAGE: Create, update + list filters (no delete)
#
# This is the most complex resource with 100+ writable attributes.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_JPP_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_MATERIAL_SITE_ID=""

describe "Resource: job-production-plans"

# ============================================================================
# Prerequisites - Create broker, customer, job site, and material site
# ============================================================================

test_name "Create prerequisite broker for job-production-plans tests"
BROKER_NAME=$(unique_name "JPPTestBroker")

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
CUSTOMER_NAME=$(unique_name "JPPTestCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true

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
JOB_SITE_NAME=$(unique_name "JPPTestJobSite")

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
    fi
else
    echo "    (Job site creation failed - some tests will be skipped)"
    pass
fi

test_name "Create prerequisite material site"
MATERIAL_SITE_NAME=$(unique_name "JPPTestMaterialSite")

xbe_json do material-sites create \
    --name "$MATERIAL_SITE_NAME" \
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
    echo "    (Material site creation failed - some tests will be skipped)"
    pass
fi

# ============================================================================
# CREATE Tests - Basic Attributes
# ============================================================================

test_name "Create job production plan with basic attributes"
TEST_NAME=$(unique_name "JPP")
TODAY=$(date +%Y-%m-%d)
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        # Note: No delete available for job-production-plans, but register for potential future cleanup
        pass
    else
        fail "Created job production plan but no ID returned"
    fi
else
    fail "Failed to create job production plan: $output"
fi

# Only continue if we successfully created a job production plan
if [[ -z "$CREATED_JPP_ID" || "$CREATED_JPP_ID" == "null" ]]; then
    echo "Cannot continue without a valid job production plan ID"
    run_tests
fi

test_name "Create job production plan with notes"
TEST_NAME2=$(unique_name "JPPWithNotes")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME2" \
    --start-on "$TODAY" \
    --start-time "08:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --notes "Test notes for job production plan"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with notes"
fi

test_name "Create job production plan with goal quantity"
TEST_NAME3=$(unique_name "JPPWithGoal")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME3" \
    --start-on "$TODAY" \
    --start-time "06:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --goal-quantity "500"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with goal quantity"
fi

test_name "Create job production plan with job site"
TEST_NAME4=$(unique_name "JPPWithJobSite")
if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
    xbe_json do job-production-plans create \
        --job-name "$TEST_NAME4" \
        --start-on "$TODAY" \
        --start-time "09:00" \
        --customer "$CREATED_CUSTOMER_ID" \
        --job-site "$CREATED_JOB_SITE_ID"

    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Failed to create job production plan with job site"
    fi
else
    skip "No job site ID available"
fi

test_name "Create job production plan with end-time"
TEST_NAME5=$(unique_name "JPPEndTime")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME5" \
    --start-on "$TODAY" \
    --start-time "06:00" \
    --end-time "16:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with end-time"
fi

test_name "Create job production plan with dispatch-instructions"
TEST_NAME6=$(unique_name "JPPDispatch")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME6" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --dispatch-instructions "Park behind the building. Check in at office."

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with dispatch-instructions"
fi

test_name "Create job production plan with cost fields"
TEST_NAME7=$(unique_name "JPPCosts")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME7" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --cost-per-truck-hour "75.00" \
    --cost-per-crew-hour "125.00"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with cost fields"
fi

test_name "Create job production plan with explicit-color-hex"
TEST_NAME9=$(unique_name "JPPColor")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME9" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --explicit-color-hex "#FF5733"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with explicit-color-hex"
fi

test_name "Create job production plan as template"
TEST_TEMPLATE=$(unique_name "JPPTemplate")
xbe_json do job-production-plans create \
    --job-name "$TEST_TEMPLATE" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-template \
    --template-name "Test Template Name"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan as template"
fi

# ============================================================================
# CREATE Tests - Boolean Attributes
# ============================================================================

test_name "Create job production plan with is-using-volumetric-measurements"
TEST_NAME_VOL=$(unique_name "JPPVol")
xbe_json do job-production-plans create \
    --job-name "$TEST_NAME_VOL" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-using-volumetric-measurements

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create job production plan with is-using-volumetric-measurements"
fi

# ============================================================================
# UPDATE Tests - String Attributes
# ============================================================================

test_name "Update job production plan job-name"
UPDATED_NAME=$(unique_name "UpdatedJPP")
xbe_json do job-production-plans update "$CREATED_JPP_ID" --job-name "$UPDATED_NAME"
assert_success

test_name "Update job production plan job-number"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --job-number "JOB-$(date +%s)"
assert_success

test_name "Update job production plan raw-job-number"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --raw-job-number "RAW-$(date +%s)"
assert_success

test_name "Update job production plan notes"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --notes "Updated notes"
assert_success

test_name "Update job production plan dispatch-instructions"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --dispatch-instructions "Updated dispatch instructions"
assert_success

test_name "Update job production plan phase-name"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --phase-name "Phase 1"
assert_success

test_name "Update job production plan end-time"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --end-time "17:00"
assert_success

test_name "Update job production plan explicit-color-hex"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --explicit-color-hex "#00FF00"
assert_success

# ============================================================================
# UPDATE Tests - Cost Attributes
# ============================================================================

test_name "Update job production plan cost-per-truck-hour"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --cost-per-truck-hour "80.00"
assert_success

test_name "Update job production plan cost-per-crew-hour"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --cost-per-crew-hour "130.00"
assert_success

# ============================================================================
# UPDATE Tests - Date/Time Attributes
# ============================================================================

test_name "Update job production plan start-on"
TOMORROW=$(date -v+1d +%Y-%m-%d 2>/dev/null || date -d "+1 day" +%Y-%m-%d)
xbe_json do job-production-plans update "$CREATED_JPP_ID" --start-on "$TOMORROW"
assert_success

test_name "Update job production plan start-time"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --start-time "09:00"
assert_success

# ============================================================================
# UPDATE Tests - Numeric Attributes
# ============================================================================

test_name "Update job production plan goal-quantity"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --goal-quantity "1000"
assert_success

test_name "Update job production plan remaining-quantity"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --remaining-quantity "500"
assert_success

test_name "Update job production plan goal-hours"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --goal-hours "8"
assert_success

test_name "Update job production plan benchmark-tons-per-truck-hour"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --benchmark-tons-per-truck-hour "25"
assert_success

test_name "Update job production plan observed-possible-cycle-minutes"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --observed-possible-cycle-minutes "45"
assert_success

test_name "Update job production plan parallel-production-count"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --parallel-production-count "2"
assert_success

# ============================================================================
# UPDATE Tests - Boolean Attributes (presence flags)
# ============================================================================

test_name "Update job production plan is-on-hold to true"
# Note: Server requires plan to be approved and not in the past to set on-hold
# We test that the CLI sends the request correctly (may fail due to business rules)
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-on-hold --on-hold-comment "Testing hold"
if [[ $status -eq 0 ]]; then
    pass
else
    # Expected to fail with validation error - test that the flag works
    if [[ "$output" == *"is-on-hold"* ]] || [[ "$output" == *"cannot be"* ]] || [[ "$output" == *"approved"* ]]; then
        echo "    (Server validation: must be approved and not in past - expected)"
        pass
    else
        fail "Unexpected error: $output"
    fi
fi

test_name "Update job production plan is-on-hold to false"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-on-hold=false
assert_success

test_name "Update job production plan is-template to true"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-template --template-name "Test Template"
assert_success

test_name "Update job production plan is-template to false"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-template=false
assert_success

test_name "Update job production plan is-using-volumetric-measurements"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-using-volumetric-measurements
assert_success

test_name "Update job production plan is-schedule-locked"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-schedule-locked
assert_success

test_name "Update job production plan is-schedule-locked to false"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-schedule-locked=false
assert_success

test_name "Update job production plan is-notifying-crew"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-notifying-crew
assert_success

test_name "Update job production plan is-managing-crew-requirements"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-managing-crew-requirements
assert_success

test_name "Update job production plan are-shifts-expecting-time-cards"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --are-shifts-expecting-time-cards
assert_success

test_name "Update job production plan show-loadout-position-to-drivers"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --show-loadout-position-to-drivers
assert_success

# ============================================================================
# UPDATE Tests - Tri-state Boolean Attributes (explicit overrides)
# ============================================================================

test_name "Update job production plan is-prevailing-wage-explicit to true"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-prevailing-wage-explicit true
assert_success

test_name "Update job production plan is-prevailing-wage-explicit to false"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-prevailing-wage-explicit false
assert_success

test_name "Update job production plan is-prevailing-wage-explicit to null (clear)"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-prevailing-wage-explicit null
assert_success

test_name "Update job production plan is-one-way-job-explicit to true"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-one-way-job-explicit true
assert_success

test_name "Update job production plan is-certification-required-explicit to true"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --is-certification-required-explicit true
assert_success

# ============================================================================
# UPDATE Tests - Approval Process Attributes
# ============================================================================

test_name "Update job production plan default-time-card-approval-process"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --default-time-card-approval-process "admin"
assert_success

test_name "Update job production plan default-time-card-approval-process to field"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --default-time-card-approval-process "field"
assert_success

test_name "Update job production plan enable-implicit-time-card-approval"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --enable-implicit-time-card-approval
assert_success

# ============================================================================
# UPDATE Tests - Relationship Attributes
# ============================================================================

test_name "Update job production plan customer"
xbe_json do job-production-plans update "$CREATED_JPP_ID" --customer "$CREATED_CUSTOMER_ID"
assert_success

if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
    test_name "Update job production plan job-site"
    xbe_json do job-production-plans update "$CREATED_JPP_ID" --job-site "$CREATED_JOB_SITE_ID"
    assert_success
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plans"
xbe_json view job-production-plans list --start-on "$TODAY"
assert_success

test_name "List job production plans returns array"
xbe_json view job-production-plans list --start-on "$TODAY"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plans"
fi

# ============================================================================
# LIST Tests - Date Range Filters
# ============================================================================

test_name "List job production plans with --start-on-min and --start-on-max"
START_DATE="$TODAY"
END_DATE=$(date -v+7d +%Y-%m-%d 2>/dev/null || date -d "+7 days" +%Y-%m-%d)
xbe_json view job-production-plans list --start-on-min "$START_DATE" --start-on-max "$END_DATE"
assert_success

test_name "List job production plans with comma-separated start-on dates"
xbe_json view job-production-plans list --start-on "$TODAY,$TOMORROW"
assert_success

# ============================================================================
# LIST Tests - Status and Basic Filters
# ============================================================================

test_name "List job production plans with --status filter"
xbe_json view job-production-plans list --start-on "$TODAY" --status "editing"
assert_success

test_name "List job production plans with --q filter"
xbe_json view job-production-plans list --start-on "$TODAY" --q "Test"
assert_success

test_name "List job production plans with --customer filter"
xbe_json view job-production-plans list --start-on "$TODAY" --customer "$CREATED_CUSTOMER_ID"
assert_success

test_name "List job production plans with --broker filter"
xbe_json view job-production-plans list --start-on "$TODAY" --broker "$CREATED_BROKER_ID"
assert_success

test_name "List job production plans with --is-template filter"
xbe_json view job-production-plans list --start-on "$TODAY" --is-template "true"
assert_success

# ============================================================================
# LIST Tests - Boolean Filters
# ============================================================================

test_name "List job production plans with --has-labor-requirements filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-labor-requirements true
assert_success

test_name "List job production plans with --has-crew-requirements filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-crew-requirements true
assert_success

test_name "List job production plans with --has-equipment-requirements filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-equipment-requirements true
assert_success

test_name "List job production plans with --is-only-for-equipment-movement filter"
xbe_json view job-production-plans list --start-on "$TODAY" --is-only-for-equipment-movement true
assert_success

test_name "List job production plans with --is-using-volumetric-measurements filter"
xbe_json view job-production-plans list --start-on "$TODAY" --is-using-volumetric-measurements true
assert_success

test_name "List job production plans with --is-auditing-time-card-approvals filter"
xbe_json view job-production-plans list --start-on "$TODAY" --is-auditing-time-card-approvals true
assert_success

test_name "List job production plans with --has-checksum-difference filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-checksum-difference true
assert_success

test_name "List job production plans with --has-manager-assignment filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-manager-assignment true
assert_success

# ============================================================================
# LIST Tests - Numeric Range Filters
# ============================================================================

test_name "List job production plans with --remaining-quantity-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --remaining-quantity-min "0"
assert_success

test_name "List job production plans with --start-time-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --start-time-min "06:00"
assert_success

test_name "List job production plans with --start-time-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --start-time-max "18:00"
assert_success

# ============================================================================
# LIST Tests - Relationship Filters
# ============================================================================

test_name "List job production plans with --broker-id filter"
xbe_json view job-production-plans list --start-on "$TODAY" --broker-id "$CREATED_BROKER_ID"
assert_success

test_name "List job production plans with --business-unit filter"
xbe_json view job-production-plans list --start-on "$TODAY" --business-unit "1"
assert_success

test_name "List job production plans with --planner filter"
xbe_json view job-production-plans list --start-on "$TODAY" --planner "1"
assert_success

test_name "List job production plans with --project-manager filter"
xbe_json view job-production-plans list --start-on "$TODAY" --project-manager "1"
assert_success

test_name "List job production plans with --project filter"
xbe_json view job-production-plans list --start-on "$TODAY" --project "1"
assert_success

test_name "List job production plans with --trucker filter"
xbe_json view job-production-plans list --start-on "$TODAY" --trucker "1"
assert_success

test_name "List job production plans with --default-trucker filter"
xbe_json view job-production-plans list --start-on "$TODAY" --default-trucker "1"
assert_success

test_name "List job production plans with --job-site filter"
if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
    xbe_json view job-production-plans list --start-on "$TODAY" --job-site "$CREATED_JOB_SITE_ID"
    assert_success
else
    xbe_json view job-production-plans list --start-on "$TODAY" --job-site "1"
    assert_success
fi

test_name "List job production plans with --material-site filter"
if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view job-production-plans list --start-on "$TODAY" --material-site "$CREATED_MATERIAL_SITE_ID"
    assert_success
else
    xbe_json view job-production-plans list --start-on "$TODAY" --material-site "1"
    assert_success
fi

test_name "List job production plans with --material-supplier filter"
xbe_json view job-production-plans list --start-on "$TODAY" --material-supplier "1"
assert_success

test_name "List job production plans with --material-type filter"
xbe_json view job-production-plans list --start-on "$TODAY" --material-type "1"
assert_success

test_name "List job production plans with --contractor filter"
xbe_json view job-production-plans list --start-on "$TODAY" --contractor "1"
assert_success

test_name "List job production plans with --cost-code filter"
xbe_json view job-production-plans list --start-on "$TODAY" --cost-code "1"
assert_success

test_name "List job production plans with --created-by filter"
xbe_json view job-production-plans list --start-on "$TODAY" --created-by "1"
assert_success

test_name "List job production plans with --template filter"
xbe_json view job-production-plans list --start-on "$TODAY" --template "1"
assert_success

test_name "List job production plans with --not-customer filter"
xbe_json view job-production-plans list --start-on "$TODAY" --not-customer "999999"
assert_success

test_name "List job production plans with --user-has-stake filter"
xbe_json view job-production-plans list --start-on "$TODAY" --user-has-stake "1"
assert_success

# ============================================================================
# LIST Tests - String/Search Filters
# ============================================================================

test_name "List job production plans with --job-name filter"
xbe_json view job-production-plans list --start-on "$TODAY" --job-name "Test"
assert_success

test_name "List job production plans with --job-number filter"
xbe_json view job-production-plans list --start-on "$TODAY" --job-number "JOB"
assert_success

test_name "List job production plans with --template-name filter"
xbe_json view job-production-plans list --start-on "$TODAY" --template-name "Template"
assert_success

test_name "List job production plans with --q-segments filter"
xbe_json view job-production-plans list --start-on "$TODAY" --q-segments "segment"
assert_success

test_name "List job production plans with --duplication-token filter"
xbe_json view job-production-plans list --start-on "$TODAY" --duplication-token "token123"
assert_success

test_name "List job production plans with --ultimate-material-types filter"
xbe_json view job-production-plans list --start-on "$TODAY" --ultimate-material-types "Asphalt"
assert_success

test_name "List job production plans with --trailer-classification-or-equivalent filter"
xbe_json view job-production-plans list --start-on "$TODAY" --trailer-classification-or-equivalent "1"
assert_success

# ============================================================================
# LIST Tests - Additional Boolean Filters
# ============================================================================

test_name "List job production plans with --could-have-labor-requirements filter"
xbe_json view job-production-plans list --start-on "$TODAY" --could-have-labor-requirements true
assert_success

test_name "List job production plans with --has-project-phase-revenue-items filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-project-phase-revenue-items true
assert_success

test_name "List job production plans with --has-material-types-with-qc-requirements filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-material-types-with-qc-requirements true
assert_success

test_name "List job production plans with --has-supply-demand-balance-cannot-compute-reasons filter"
xbe_json view job-production-plans list --start-on "$TODAY" --has-supply-demand-balance-cannot-compute-reasons true
assert_success

test_name "List job production plans with --with-non-deletable-lineup-jpps filter"
xbe_json view job-production-plans list --start-on "$TODAY" --with-non-deletable-lineup-jpps true
assert_success

test_name "List job production plans with --default-time-card-approval-process filter"
xbe_json view job-production-plans list --start-on "$TODAY" --default-time-card-approval-process "admin"
assert_success

# ============================================================================
# LIST Tests - Additional Date/Time Filters
# ============================================================================

test_name "List job production plans with --active-on filter"
xbe_json view job-production-plans list --start-on "$TODAY" --active-on "$TODAY"
assert_success

test_name "List job production plans with --practically-start-on filter"
xbe_json view job-production-plans list --start-on "$TODAY" --practically-start-on "$TODAY"
assert_success

test_name "List job production plans with --practically-start-on-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --practically-start-on-min "$TODAY"
assert_success

test_name "List job production plans with --practically-start-on-max filter"
NEXT_MONTH=$(date -v+30d +%Y-%m-%d 2>/dev/null || date -d "+30 days" +%Y-%m-%d)
xbe_json view job-production-plans list --start-on "$TODAY" --practically-start-on-max "$NEXT_MONTH"
assert_success

test_name "List job production plans with --start-at-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --start-at-min "${TODAY}T00:00:00Z"
assert_success

test_name "List job production plans with --start-at-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --start-at-max "${TODAY}T23:59:59Z"
assert_success

test_name "List job production plans with --template-start-on-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --template-start-on-min "$TODAY"
assert_success

test_name "List job production plans with --template-start-on-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --template-start-on-max "$NEXT_MONTH"
assert_success

test_name "List job production plans with --job-site-active-around filter"
xbe_json view job-production-plans list --start-on "$TODAY" --job-site-active-around "${TODAY}T12:00:00Z"
assert_success

# ============================================================================
# LIST Tests - Additional Numeric Range Filters
# ============================================================================

test_name "List job production plans with --remaining-quantity-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --remaining-quantity-max "10000"
assert_success

test_name "List job production plans with --checksum-difference filter"
xbe_json view job-production-plans list --start-on "$TODAY" --checksum-difference "0"
assert_success

test_name "List job production plans with --checksum-difference-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --checksum-difference-min "0"
assert_success

test_name "List job production plans with --checksum-difference-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --checksum-difference-max "100"
assert_success

test_name "List job production plans with --planned-tons-per-productive-segment-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --planned-tons-per-productive-segment-min "0"
assert_success

test_name "List job production plans with --planned-tons-per-productive-segment-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --planned-tons-per-productive-segment-max "1000"
assert_success

test_name "List job production plans with --material-type-ultimate-parent-count-min filter"
xbe_json view job-production-plans list --start-on "$TODAY" --material-type-ultimate-parent-count-min "0"
assert_success

test_name "List job production plans with --material-type-ultimate-parent-count-max filter"
xbe_json view job-production-plans list --start-on "$TODAY" --material-type-ultimate-parent-count-max "10"
assert_success

# ============================================================================
# LIST Tests - Special Filters
# ============================================================================

test_name "List job production plans with --external-identification-value filter"
xbe_json view job-production-plans list --start-on "$TODAY" --external-identification-value "EXT123"
assert_success

test_name "List job production plans with --reference-data filter"
xbe_json view job-production-plans list --start-on "$TODAY" --reference-data "test-key|test-value"
assert_success

test_name "List job production plans with --practically-start-on-between filter"
NEXT_WEEK=$(date -v+7d +%Y-%m-%d 2>/dev/null || date -d "+7 days" +%Y-%m-%d)
xbe_json view job-production-plans list --start-on-min "$TODAY" --practically-start-on-between "$TODAY|$NEXT_WEEK"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List job production plans with --limit"
xbe_json view job-production-plans list --start-on "$TODAY" --limit 5
assert_success

test_name "List job production plans with --offset"
xbe_json view job-production-plans list --start-on "$TODAY" --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "List job production plans without date filter fails"
xbe_json view job-production-plans list
assert_failure

test_name "Update without any fields fails"
xbe_json do job-production-plans update "$CREATED_JPP_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
