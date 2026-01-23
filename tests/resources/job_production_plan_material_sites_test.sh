#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Material Sites
#
# Tests list, show, create, update, and delete operations for
# job_production_plan_material_sites.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_JPPMS_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_JPP_ID=""
CREATED_MATERIAL_TYPE_ID=""

describe "Resource: job-production-plan-material-sites"

# ============================================================================
# Prerequisites - Create broker, customer, material supplier/site, job plan
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPPMSTestBroker")

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
CUSTOMER_NAME=$(unique_name "JPPMSTestCustomer")

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
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "JPPMSTestSupplier")

xbe_json do material-suppliers create \
    --name "$SUPPLIER_NAME" \
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
    if [[ -n "$XBE_TEST_MATERIAL_SUPPLIER_ID" ]]; then
        CREATED_MATERIAL_SUPPLIER_ID="$XBE_TEST_MATERIAL_SUPPLIER_ID"
        echo "    Using XBE_TEST_MATERIAL_SUPPLIER_ID: $CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Failed to create material supplier and XBE_TEST_MATERIAL_SUPPLIER_ID not set"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
fi

test_name "Create prerequisite material site"
MATERIAL_SITE_NAME=$(unique_name "JPPMSTestMaterialSite")

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
        echo "Cannot continue without a material site"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_SITE_ID" ]]; then
        CREATED_MATERIAL_SITE_ID="$XBE_TEST_MATERIAL_SITE_ID"
        echo "    Using XBE_TEST_MATERIAL_SITE_ID: $CREATED_MATERIAL_SITE_ID"
        pass
    else
        fail "Failed to create material site and XBE_TEST_MATERIAL_SITE_ID not set"
        echo "Cannot continue without a material site"
        run_tests
    fi
fi

test_name "Create prerequisite job production plan"
PLAN_NAME=$(unique_name "JPPMSTestPlan")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$PLAN_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_JOB_PRODUCTION_PLAN_ID" ]]; then
        CREATED_JPP_ID="$XBE_TEST_JOB_PRODUCTION_PLAN_ID"
        echo "    Using XBE_TEST_JOB_PRODUCTION_PLAN_ID: $CREATED_JPP_ID"
        pass
    else
        fail "Failed to create job production plan and XBE_TEST_JOB_PRODUCTION_PLAN_ID not set"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
fi

test_name "Create prerequisite material type"
MATERIAL_TYPE_NAME=$(unique_name "JPPMSTestMaterialType")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_TYPE_ID" ]]; then
        CREATED_MATERIAL_TYPE_ID="$XBE_TEST_MATERIAL_TYPE_ID"
        echo "    Using XBE_TEST_MATERIAL_TYPE_ID: $CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Failed to create material type and XBE_TEST_MATERIAL_TYPE_ID not set"
        echo "Cannot continue without a material type"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan material site with required fields"
xbe_json do job-production-plan-material-sites create \
    --job-production-plan "$CREATED_JPP_ID" \
    --material-site "$CREATED_MATERIAL_SITE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPPMS_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPPMS_ID" && "$CREATED_JPPMS_ID" != "null" ]]; then
        register_cleanup "job-production-plan-material-sites" "$CREATED_JPPMS_ID"
        pass
    else
        fail "Created job production plan material site but no ID returned"
    fi
else
    fail "Failed to create job production plan material site"
fi

if [[ -z "$CREATED_JPPMS_ID" || "$CREATED_JPPMS_ID" == "null" ]]; then
    echo "Cannot continue without a job production plan material site ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update is-default"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --is-default
assert_success

test_name "Update miles"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --miles 12.5
assert_success

test_name "Update default-ticket-maker"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --default-ticket-maker material_site
assert_success

test_name "Update user-ticket-maker-material-type-ids"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --user-ticket-maker-material-type-ids "$CREATED_MATERIAL_TYPE_ID"
assert_success

test_name "Update has-user-scale-tickets"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --has-user-scale-tickets
assert_success

test_name "Update plan-requires-site-specific-material-types"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --plan-requires-site-specific-material-types
assert_success

test_name "Update plan-requires-supplier-specific-material-types"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID" --plan-requires-supplier-specific-material-types
assert_success

test_name "Update without any fields fails"
xbe_json do job-production-plan-material-sites update "$CREATED_JPPMS_ID"
assert_failure

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan material site"
xbe_json view job-production-plan-material-sites show "$CREATED_JPPMS_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan material sites"
xbe_json view job-production-plan-material-sites list --limit 5
assert_success

test_name "List job production plan material sites returns array"
xbe_json view job-production-plan-material-sites list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan material sites"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job production plan material sites with --job-production-plan"
xbe_json view job-production-plan-material-sites list --job-production-plan "$CREATED_JPP_ID"
assert_success

test_name "List job production plan material sites with --material-site"
xbe_json view job-production-plan-material-sites list --material-site "$CREATED_MATERIAL_SITE_ID"
assert_success

test_name "List job production plan material sites with --is-default"
xbe_json view job-production-plan-material-sites list --is-default true
assert_success

test_name "List job production plan material sites with --miles-min"
xbe_json view job-production-plan-material-sites list --miles-min 10
assert_success

test_name "List job production plan material sites with --miles-max"
xbe_json view job-production-plan-material-sites list --miles-max 20
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job production plan material site requires --confirm flag"
xbe_json do job-production-plan-material-sites delete "$CREATED_JPPMS_ID"
assert_failure

test_name "Delete job production plan material site with --confirm"
xbe_run do job-production-plan-material-sites delete "$CREATED_JPPMS_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without job-production-plan fails"
xbe_json do job-production-plan-material-sites create --material-site "$CREATED_MATERIAL_SITE_ID"
assert_failure

test_name "Create without material-site fails"
xbe_json do job-production-plan-material-sites create --job-production-plan "$CREATED_JPP_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
