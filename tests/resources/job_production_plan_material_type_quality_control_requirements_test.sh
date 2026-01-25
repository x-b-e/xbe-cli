#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Material Type Quality Control Requirements
#
# Tests CRUD operations for the job_production_plan_material_type_quality_control_requirements resource.
#
# COVERAGE: All filters + all create/update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_UNIT_OF_MEASURE_ID=""
CREATED_JPP_MATERIAL_TYPE_ID=""
CREATED_QC_CLASSIFICATION_ID=""
CREATED_QC_CLASSIFICATION_ID2=""
CREATED_REQUIREMENT_ID=""

NOTE_INITIAL=""
NOTE_UPDATED=""

DIRECT_API_AVAILABLE=0
if [[ -n "$XBE_TOKEN" ]]; then
    DIRECT_API_AVAILABLE=1
fi

cleanup_api_resources() {
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        return
    fi

    if [[ -n "$CREATED_JPP_MATERIAL_TYPE_ID" && "$CREATED_JPP_MATERIAL_TYPE_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/job-production-plan-material-types/$CREATED_JPP_MATERIAL_TYPE_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi
}

trap 'cleanup_api_resources; run_cleanup' EXIT

api_post() {
    local path="$1"
    local body="$2"
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        output="Missing XBE_TOKEN for direct API calls"
        status=1
        return
    fi
    run curl -sS -X POST "$XBE_BASE_URL$path" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json" \
        -d "$body"
}

describe "Resource: job-production-plan-material-type-quality-control-requirements"

if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
    skip "Set XBE_TOKEN to create job production plan material types via direct API"
    run_tests
fi

# ============================================================================
# Prerequisites - Create broker, customer, job production plan, material type
# ============================================================================

test_name "Create prerequisite broker for quality control requirement tests"
BROKER_NAME=$(unique_name "JPPMTQCReqBroker")

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
        fail "Failed to create broker"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPMTQCReqCustomer")

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

test_name "Create prerequisite job production plan"
TODAY=$(date +%Y-%m-%d)
JPP_NAME=$(unique_name "JPPMTQCReqPlan")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        register_cleanup "job-production-plans" "$CREATED_JPP_ID"
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    echo "Cannot continue without a job production plan"
    run_tests
fi

test_name "Create prerequisite material type"
MATERIAL_TYPE_NAME=$(unique_name "JPPMTQCReqMaterialType")

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

test_name "Select unit of measure for material types"

xbe_json view unit-of-measures list --name "Ton" --limit 1

if [[ $status -eq 0 ]]; then
    CREATED_UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
    if [[ -n "$CREATED_UNIT_OF_MEASURE_ID" && "$CREATED_UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        fail "No unit of measure returned"
        echo "Cannot continue without a unit of measure"
        run_tests
    fi
else
    fail "Failed to list unit of measures"
    echo "Cannot continue without a unit of measure"
    run_tests
fi

test_name "Create prerequisite quality control classifications"
QC_NAME=$(unique_name "JPPMTQCReqQC")
QC_NAME2=$(unique_name "JPPMTQCReqQC2")

xbe_json do quality-control-classifications create \
    --name "$QC_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_QC_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_QC_CLASSIFICATION_ID" && "$CREATED_QC_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "quality-control-classifications" "$CREATED_QC_CLASSIFICATION_ID"
        pass
    else
        fail "Created quality control classification but no ID returned"
        echo "Cannot continue without quality control classification"
        run_tests
    fi
else
    fail "Failed to create quality control classification"
    echo "Cannot continue without quality control classification"
    run_tests
fi

xbe_json do quality-control-classifications create \
    --name "$QC_NAME2" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_QC_CLASSIFICATION_ID2=$(json_get ".id")
    if [[ -n "$CREATED_QC_CLASSIFICATION_ID2" && "$CREATED_QC_CLASSIFICATION_ID2" != "null" ]]; then
        register_cleanup "quality-control-classifications" "$CREATED_QC_CLASSIFICATION_ID2"
        pass
    else
        fail "Created second quality control classification but no ID returned"
        echo "Cannot continue without second quality control classification"
        run_tests
    fi
else
    fail "Failed to create second quality control classification"
    echo "Cannot continue without second quality control classification"
    run_tests
fi

# Create job production plan material type via direct API

NOTE_INITIAL="Temperature check"
NOTE_UPDATED="Updated temperature check"

JPP_MATERIAL_TYPE_BODY=$(cat <<JSON
{"data":{"type":"job-production-plan-material-types","attributes":{"quantity":100},"relationships":{"job-production-plan":{"data":{"type":"job-production-plans","id":"$CREATED_JPP_ID"}},"material-type":{"data":{"type":"material-types","id":"$CREATED_MATERIAL_TYPE_ID"}},"unit-of-measure":{"data":{"type":"unit-of-measures","id":"$CREATED_UNIT_OF_MEASURE_ID"}}}}}
JSON
)

test_name "Create job production plan material type via direct API"
api_post "/v1/job-production-plan-material-types" "$JPP_MATERIAL_TYPE_BODY"
if [[ $status -eq 0 ]]; then
    CREATED_JPP_MATERIAL_TYPE_ID=$(json_get ".data.id")
    if [[ -n "$CREATED_JPP_MATERIAL_TYPE_ID" && "$CREATED_JPP_MATERIAL_TYPE_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan material type but no ID returned"
        echo "Cannot continue without job production plan material type"
        run_tests
    fi
else
    fail "Failed to create job production plan material type"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan material type quality control requirement"

xbe_json do job-production-plan-material-type-quality-control-requirements create \
    --job-production-plan-material-type "$CREATED_JPP_MATERIAL_TYPE_ID" \
    --quality-control-classification "$CREATED_QC_CLASSIFICATION_ID" \
    --note "$NOTE_INITIAL"

if [[ $status -eq 0 ]]; then
    CREATED_REQUIREMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_REQUIREMENT_ID" && "$CREATED_REQUIREMENT_ID" != "null" ]]; then
        register_cleanup "job-production-plan-material-type-quality-control-requirements" "$CREATED_REQUIREMENT_ID"
        pass
    else
        fail "Created requirement but no ID returned"
    fi
else
    fail "Failed to create requirement"
fi

if [[ -z "$CREATED_REQUIREMENT_ID" || "$CREATED_REQUIREMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid requirement ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update requirement note"

xbe_json do job-production-plan-material-type-quality-control-requirements update "$CREATED_REQUIREMENT_ID" \
    --note "$NOTE_UPDATED"
assert_success

test_name "Update requirement quality control classification"

xbe_json do job-production-plan-material-type-quality-control-requirements update "$CREATED_REQUIREMENT_ID" \
    --quality-control-classification "$CREATED_QC_CLASSIFICATION_ID2"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show requirement details"

xbe_json view job-production-plan-material-type-quality-control-requirements show "$CREATED_REQUIREMENT_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show requirement"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List requirements"

xbe_json view job-production-plan-material-type-quality-control-requirements list --limit 5
assert_success

test_name "List requirements returns array"

xbe_json view job-production-plan-material-type-quality-control-requirements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list requirements"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List requirements with --job-production-plan-material-type filter"

xbe_json view job-production-plan-material-type-quality-control-requirements list \
    --job-production-plan-material-type "$CREATED_JPP_MATERIAL_TYPE_ID" --limit 10
assert_success

test_name "List requirements with --quality-control-classification filter"

xbe_json view job-production-plan-material-type-quality-control-requirements list \
    --quality-control-classification "$CREATED_QC_CLASSIFICATION_ID2" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List requirements with --limit"

xbe_json view job-production-plan-material-type-quality-control-requirements list --limit 3
assert_success

test_name "List requirements with --offset"

xbe_json view job-production-plan-material-type-quality-control-requirements list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create requirement without job production plan material type fails"

xbe_json do job-production-plan-material-type-quality-control-requirements create \
    --quality-control-classification "$CREATED_QC_CLASSIFICATION_ID"
assert_failure

test_name "Create requirement without quality control classification fails"

xbe_json do job-production-plan-material-type-quality-control-requirements create \
    --job-production-plan-material-type "$CREATED_JPP_MATERIAL_TYPE_ID"
assert_failure

test_name "Update requirement without changes fails"

xbe_json do job-production-plan-material-type-quality-control-requirements update "$CREATED_REQUIREMENT_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requirement requires --confirm flag"

xbe_run do job-production-plan-material-type-quality-control-requirements delete "$CREATED_REQUIREMENT_ID"
assert_failure

test_name "Delete requirement with --confirm"

xbe_run do job-production-plan-material-type-quality-control-requirements delete "$CREATED_REQUIREMENT_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
