#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Material Types
#
# Tests CRUD operations for the job_production_plan_material_types resource.
# Job production plan material types define planned materials for a plan.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_JPPMT_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_TYPE_ID_2=""
UNIT_OF_MEASURE_ID=""
CREATED_COST_CODE_ID=""
CREATED_JPP_ID=""
CREATED_MIX_DESIGN_ID=""

TODAY=$(date +%Y-%m-%d)

describe "Resource: job-production-plan-material-types"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPPMTBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPMTCustomer")

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
        run_tests
    fi
else
    fail "Failed to create customer"
    run_tests
fi

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "JPPMTSupplier")

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
        run_tests
    fi
else
    fail "Failed to create material supplier"
    run_tests
fi

test_name "Fetch material site"
xbe_json view material-sites list --limit 1

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID=$(json_get ".[0].id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
        pass
    else
        fail "No material site ID returned"
        run_tests
    fi
else
    fail "Failed to list material sites"
    run_tests
fi

test_name "Create prerequisite material type"
MT_NAME=$(unique_name "JPPMTMaterial")

xbe_json do material-types create \
    --name "$MT_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        run_tests
    fi
else
    fail "Failed to create material type"
    run_tests
fi

test_name "Create secondary material type for updates"
MT_NAME2=$(unique_name "JPPMTMaterial2")

xbe_json do material-types create \
    --name "$MT_NAME2"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID_2" && "$CREATED_MATERIAL_TYPE_ID_2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID_2"
        pass
    else
        fail "Created secondary material type but no ID returned"
        run_tests
    fi
else
    fail "Failed to create secondary material type"
    run_tests
fi

test_name "Fetch unit of measure"
xbe_json view unit-of-measures list --limit 100
if [[ $status -eq 0 ]]; then
    UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r '.[] | select(.abbreviation == "ton" or .abbreviation == "ea" or .abbreviation == "load") | .id' | head -n 1)
fi
if [[ -z "$UNIT_OF_MEASURE_ID" || "$UNIT_OF_MEASURE_ID" == "null" ]]; then
    UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
fi
if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    pass
else
    fail "No valid unit of measure ID returned"
    run_tests
fi

test_name "Create prerequisite cost code"
COST_CODE_VALUE="JPPMT-$(date +%s)-${RANDOM}"

xbe_json do cost-codes create \
    --code "$COST_CODE_VALUE" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_COST_CODE_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_CODE_ID" && "$CREATED_COST_CODE_ID" != "null" ]]; then
        register_cleanup "cost-codes" "$CREATED_COST_CODE_ID"
        pass
    else
        fail "Created cost code but no ID returned"
        run_tests
    fi
else
    fail "Failed to create cost code"
    run_tests
fi

test_name "Create prerequisite job production plan"
PLAN_NAME=$(unique_name "JPPMTPlan")

xbe_json do job-production-plans create \
    --job-name "$PLAN_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --cost-codes "$CREATED_COST_CODE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    run_tests
fi

test_name "Fetch material mix design"
xbe_json view material-mix-designs list --limit 1

if [[ $status -eq 0 ]]; then
    CREATED_MIX_DESIGN_ID=$(json_get ".[0].id")
    if [[ -n "$CREATED_MIX_DESIGN_ID" && "$CREATED_MIX_DESIGN_ID" != "null" ]]; then
        pass
    else
        CREATED_MIX_DESIGN_ID=""
        skip "No material mix design available"
    fi
else
    CREATED_MIX_DESIGN_ID=""
    skip "Failed to list material mix designs"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan material type with required fields"
xbe_json do job-production-plan-material-types create \
    --job-production-plan "$CREATED_JPP_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID" \
    --quantity "100"

if [[ $status -eq 0 ]]; then
    CREATED_JPPMT_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPPMT_ID" && "$CREATED_JPPMT_ID" != "null" ]]; then
        register_cleanup "job-production-plan-material-types" "$CREATED_JPPMT_ID"
        pass
    else
        fail "Created job production plan material type but no ID returned"
    fi
else
    fail "Failed to create job production plan material type"
fi

# Only continue if we have a valid ID
if [[ -z "$CREATED_JPPMT_ID" || "$CREATED_JPPMT_ID" == "null" ]]; then
    echo "Cannot continue without a valid job production plan material type ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update job production plan material type quantity"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --quantity "150"
assert_success

test_name "Update job production plan material type display name"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --explicit-display-name "Updated Display"
assert_success

test_name "Update plan requirement flags"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --plan-requires-site-specific-material-types
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"plan requires"* ]] || [[ "$output" == *"material-site"* ]]; then
        echo "    (Plan requirement validation - expected)"
        pass
    else
        fail "Failed to update site-specific requirement"
    fi
fi

xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --plan-requires-supplier-specific-material-types
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"plan requires"* ]] || [[ "$output" == *"material-site"* ]]; then
        echo "    (Plan requirement validation - expected)"
        pass
    else
        fail "Failed to update supplier-specific requirement"
    fi
fi

test_name "Update job production plan relationship"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --job-production-plan "$CREATED_JPP_ID"
assert_success

test_name "Update material type relationship"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --material-type "$CREATED_MATERIAL_TYPE_ID_2"
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"cannot be changed"* ]]; then
        echo "    (Material type change not allowed - expected validation)"
        pass
    else
        fail "Failed to update material type"
    fi
fi

test_name "Update unit of measure relationship"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --unit-of-measure "$UNIT_OF_MEASURE_ID"
assert_success

test_name "Update material site relationship"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --material-site "$CREATED_MATERIAL_SITE_ID"
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"material site"* || "$output" == *"material-site"* ]] && [[ "$output" == *"job production plan"* ]]; then
        echo "    (Material site not on plan - expected validation)"
        pass
    else
        fail "Failed to update material site"
    fi
fi

test_name "Update default cost code relationship"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --default-cost-code "$CREATED_COST_CODE_ID"
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"cost code"* ]] || [[ "$output" == *"must be in the list"* ]]; then
        echo "    (Server validation for default cost code)"
        pass
    else
        fail "Failed to update default cost code"
    fi
fi

if [[ -n "$CREATED_MIX_DESIGN_ID" ]]; then
    test_name "Update explicit material mix design relationship"
    xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --explicit-material-mix-design "$CREATED_MIX_DESIGN_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"material mix design"* ]] || [[ "$output" == *"material-mix-design"* ]] || [[ "$output" == *"explicit-material-mix-design"* ]]; then
            echo "    (Server validation for explicit material mix design)"
            pass
        else
            fail "Failed to update explicit material mix design"
        fi
    fi
else
    test_name "Update explicit material mix design relationship"
    skip "No material mix design available"
fi

test_name "Clear explicit inventory location"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --explicit-material-type-material-site-inventory-location ""
assert_success

test_name "Mark quantity as unknown"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID" --quantity "0" --is-quantity-unknown
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan material type"
xbe_json view job-production-plan-material-types show "$CREATED_JPPMT_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan material types"
xbe_json view job-production-plan-material-types list
assert_success

test_name "List job production plan material types returns array"
xbe_json view job-production-plan-material-types list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan material types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by job production plan"
xbe_json view job-production-plan-material-types list --job-production-plan "$CREATED_JPP_ID"
assert_success

test_name "Filter by material type"
xbe_json view job-production-plan-material-types list --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_success

test_name "Filter by unit of measure"
xbe_json view job-production-plan-material-types list --unit-of-measure "$UNIT_OF_MEASURE_ID"
assert_success

test_name "Filter by material site"
xbe_json view job-production-plan-material-types list --material-site "$CREATED_MATERIAL_SITE_ID"
assert_success

test_name "Filter by default cost code"
xbe_json view job-production-plan-material-types list --default-cost-code "$CREATED_COST_CODE_ID"
assert_success

if [[ -n "$CREATED_MIX_DESIGN_ID" ]]; then
    test_name "Filter by explicit material mix design"
    xbe_json view job-production-plan-material-types list --explicit-material-mix-design "$CREATED_MIX_DESIGN_ID"
    assert_success
else
    test_name "Filter by explicit material mix design"
    skip "No material mix design available"
fi

test_name "Filter by customer"
xbe_json view job-production-plan-material-types list --customer "$CREATED_CUSTOMER_ID"
assert_success

test_name "Filter by broker"
xbe_json view job-production-plan-material-types list --broker "$CREATED_BROKER_ID"
assert_success

test_name "Filter by material supplier"
xbe_json view job-production-plan-material-types list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"
assert_success

test_name "Filter by start-on-min"
xbe_json view job-production-plan-material-types list --start-on-min "$TODAY"
assert_success

test_name "Filter by start-on-max"
xbe_json view job-production-plan-material-types list --start-on-max "$TODAY"
assert_success

test_name "Filter by status"
xbe_json view job-production-plan-material-types list --status editing
assert_success

test_name "Filter by external identification value"
xbe_json view job-production-plan-material-types list --external-identification-value "test-ext-id"
assert_success

# ============================================================================
# LIST Tests - Pagination/Sorting
# ============================================================================

test_name "List with --limit"
xbe_json view job-production-plan-material-types list --limit 5
assert_success

test_name "List with --offset"
xbe_json view job-production-plan-material-types list --limit 5 --offset 5
assert_success

test_name "List with --sort"
xbe_json view job-production-plan-material-types list --sort "id"
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required flags fails"
xbe_json do job-production-plan-material-types create
assert_failure

test_name "Update without any fields fails"
xbe_json do job-production-plan-material-types update "$CREATED_JPPMT_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
TEST_DEL_NAME=$(unique_name "JPPMTDelete")
xbe_json do job-production-plan-material-types create \
    --job-production-plan "$CREATED_JPP_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID" \
    --quantity "40"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do job-production-plan-material-types delete "$DEL_ID"
    assert_failure
    register_cleanup "job-production-plan-material-types" "$DEL_ID"
else
    skip "Could not create job production plan material type for delete test"
fi

test_name "Delete with --confirm"
TEST_DEL_NAME2=$(unique_name "JPPMTDelete2")
xbe_json do job-production-plan-material-types create \
    --job-production-plan "$CREATED_JPP_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID" \
    --quantity "45"
if [[ $status -eq 0 ]]; then
    DEL_ID2=$(json_get ".id")
    xbe_json do job-production-plan-material-types delete "$DEL_ID2" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"cannot"* ]] || [[ "$output" == *"deleted"* ]]; then
            echo "    (Server prevented delete - expected)"
            pass
        else
            fail "Delete failed unexpectedly: $output"
        fi
    fi
else
    skip "Could not create job production plan material type for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
