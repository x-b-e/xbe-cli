#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Cost Item Actuals
#
# Tests create/update/delete operations and list filters for project-phase-cost-item-actuals.
# Requires creating supporting project, revenue, and cost resources.
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_PHASE_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_PRODUCTION_PLAN_ID=""
CREATED_PROJECT_REVENUE_CLASSIFICATION_ID=""
CREATED_PROJECT_COST_CLASSIFICATION_ID=""
UNIT_OF_MEASURE_ID=""
CREATED_PROJECT_REVENUE_ITEM_ID=""
CREATED_PROJECT_PHASE_REVENUE_ITEM_ID=""
CREATED_PROJECT_PHASE_COST_ITEM_ID=""
CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID=""
CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID=""
CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2=""
CREATED_BY_ID=""

SKIP_MUTATION=0
if [[ -z "$XBE_TOKEN" ]]; then
    SKIP_MUTATION=1
fi

describe "Resource: project-phase-cost-item-actuals"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "PPciaBroker")

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

test_name "Create prerequisite developer"
DEV_NAME=$(unique_name "PPciaDeveloper")

xbe_json do developers create --name "$DEV_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

test_name "Create prerequisite project"
PROJECT_NAME=$(unique_name "PPciaProject")

xbe_json do projects create \
    --name "$PROJECT_NAME" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        echo "Cannot continue without a project"
        run_tests
    fi
else
    fail "Failed to create project"
    echo "Cannot continue without a project"
    run_tests
fi

test_name "Create prerequisite project phase"
PHASE_NAME=$(unique_name "PPciaPhase")

xbe_json do project-phases create \
    --project "$CREATED_PROJECT_ID" \
    --name "$PHASE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_PHASE_ID" && "$CREATED_PROJECT_PHASE_ID" != "null" ]]; then
        register_cleanup "project-phases" "$CREATED_PROJECT_PHASE_ID"
        pass
    else
        fail "Created project phase but no ID returned"
        echo "Cannot continue without a project phase"
        run_tests
    fi
else
    fail "Failed to create project phase"
    echo "Cannot continue without a project phase"
    run_tests
fi

test_name "Resolve unit of measure"
xbe_json view unit-of-measures list --limit 1
if [[ $status -eq 0 ]]; then
    UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        fail "No unit of measure found"
        echo "Cannot continue without a unit of measure"
        run_tests
    fi
else
    fail "Failed to list unit of measures"
    echo "Cannot continue without a unit of measure"
    run_tests
fi

# Project revenue and cost classifications
if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create project revenue classification"
    REV_CLASS_NAME=$(unique_name "PPciaRevClass")
    xbe_json do project-revenue-classifications create --name "$REV_CLASS_NAME" --broker "$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_REVENUE_CLASSIFICATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID" && "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID" != "null" ]]; then
            register_cleanup "project-revenue-classifications" "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID"
            pass
        else
            fail "Created project revenue classification but no ID returned"
        fi
    else
        fail "Failed to create project revenue classification"
    fi

    test_name "Create project cost classification"
    COST_CLASS_NAME=$(unique_name "PPciaCostClass")
    xbe_json do project-cost-classifications create --name "$COST_CLASS_NAME" --broker "$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_COST_CLASSIFICATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_PROJECT_COST_CLASSIFICATION_ID" && "$CREATED_PROJECT_COST_CLASSIFICATION_ID" != "null" ]]; then
            register_cleanup "project-cost-classifications" "$CREATED_PROJECT_COST_CLASSIFICATION_ID"
            pass
        else
            fail "Created project cost classification but no ID returned"
        fi
    else
        fail "Failed to create project cost classification"
    fi
fi

# Create customer + job production plan (needed for job production plan filters)
if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create prerequisite customer"
    CUSTOMER_NAME=$(unique_name "PPciaCustomer")
    xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID" --can-manage-crew-requirements true
    if [[ $status -eq 0 ]]; then
        CREATED_CUSTOMER_ID=$(json_get ".id")
        if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
            register_cleanup "customers" "$CREATED_CUSTOMER_ID"
            pass
        else
            fail "Created customer but no ID returned"
        fi
    else
        fail "Failed to create customer"
    fi

    test_name "Create job production plan"
    TODAY=$(date +%Y-%m-%d)
    JOB_NAME=$(unique_name "PPciaJPP")
    xbe_json do job-production-plans create \
        --job-name "$JOB_NAME" \
        --start-on "$TODAY" \
        --start-time "07:00" \
        --customer "$CREATED_CUSTOMER_ID" \
        --project "$CREATED_PROJECT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_JOB_PRODUCTION_PLAN_ID=$(json_get ".id")
        if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
            register_cleanup "job-production-plans" "$CREATED_JOB_PRODUCTION_PLAN_ID"
            pass
        else
            fail "Created job production plan but no ID returned"
        fi
    else
        fail "Failed to create job production plan"
    fi
fi

# Custom cleanup for resources created via direct API
cleanup_project_phase_cost_item_actuals() {
    if [[ -n "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" && "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" != "null" ]]; then
        xbe_run do project-phase-cost-item-actuals delete "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" --confirm 2>/dev/null || true
    fi

    if [[ -n "$XBE_TOKEN" ]]; then
        if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-phase-revenue-item-actuals/$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-phase-revenue-item-actuals/$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" && "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/job-production-plan-project-phase-revenue-items/$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_PROJECT_PHASE_COST_ITEM_ID" && "$CREATED_PROJECT_PHASE_COST_ITEM_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-phase-cost-items/$CREATED_PROJECT_PHASE_COST_ITEM_ID" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-phase-revenue-items/$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_PROJECT_REVENUE_ITEM_ID" && "$CREATED_PROJECT_REVENUE_ITEM_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-revenue-items/$CREATED_PROJECT_REVENUE_ITEM_ID" \
                >/dev/null 2>&1 || true
        fi
    fi

    run_cleanup
}
trap cleanup_project_phase_cost_item_actuals EXIT

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project phase cost item actual without required fields fails"
xbe_json do project-phase-cost-item-actuals create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete/filter tests without XBE_TOKEN"
    # Still run basic list test
    test_name "List project phase cost item actuals"
    xbe_json view project-phase-cost-item-actuals list --limit 5
    assert_success
    run_tests
fi

if [[ -z "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID" || -z "$CREATED_PROJECT_COST_CLASSIFICATION_ID" ]]; then
    skip "Missing project classifications; skipping mutation tests"
    run_tests
fi

# Resolve current user for created-by filter tests
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CREATED_BY_ID=$(json_get ".id")
fi

# Create project revenue item (direct API)
PROJECT_REVENUE_ITEM_DESC=$(unique_name "PPciaRevItem")
revenue_item_payload=$(cat <<JSON
{"data":{"type":"project-revenue-items","attributes":{"description":"$PROJECT_REVENUE_ITEM_DESC"},"relationships":{"project":{"data":{"type":"projects","id":"$CREATED_PROJECT_ID"}},"revenue-classification":{"data":{"type":"project-revenue-classifications","id":"$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID"}},"unit-of-measure":{"data":{"type":"unit-of-measures","id":"$UNIT_OF_MEASURE_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-revenue-items" \
    -d "$revenue_item_payload"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_REVENUE_ITEM_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_REVENUE_ITEM_ID" && "$CREATED_PROJECT_REVENUE_ITEM_ID" != "null" ]]; then
        pass
    else
        fail "Created project revenue item but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project revenue item"
    run_tests
fi

# Create project phase revenue item (direct API)
phase_revenue_item_payload=$(cat <<JSON
{"data":{"type":"project-phase-revenue-items","relationships":{"project-phase":{"data":{"type":"project-phases","id":"$CREATED_PROJECT_PHASE_ID"}},"project-revenue-item":{"data":{"type":"project-revenue-items","id":"$CREATED_PROJECT_REVENUE_ITEM_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-phase-revenue-items" \
    -d "$phase_revenue_item_payload"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_REVENUE_ITEM_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
        pass
    else
        fail "Created project phase revenue item but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project phase revenue item"
    run_tests
fi

# Create project phase cost item (direct API)
phase_cost_item_payload=$(cat <<JSON
{"data":{"type":"project-phase-cost-items","relationships":{"project-phase-revenue-item":{"data":{"type":"project-phase-revenue-items","id":"$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID"}},"project-cost-classification":{"data":{"type":"project-cost-classifications","id":"$CREATED_PROJECT_COST_CLASSIFICATION_ID"}},"unit-of-measure":{"data":{"type":"unit-of-measures","id":"$UNIT_OF_MEASURE_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-phase-cost-items" \
    -d "$phase_cost_item_payload"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_COST_ITEM_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_PHASE_COST_ITEM_ID" && "$CREATED_PROJECT_PHASE_COST_ITEM_ID" != "null" ]]; then
        pass
    else
        fail "Created project phase cost item but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project phase cost item"
    run_tests
fi

# Create job production plan project phase revenue item (direct API)
jpppri_payload=$(cat <<JSON
{"data":{"type":"job-production-plan-project-phase-revenue-items","relationships":{"job-production-plan":{"data":{"type":"job-production-plans","id":"$CREATED_JOB_PRODUCTION_PLAN_ID"}},"project-phase-revenue-item":{"data":{"type":"project-phase-revenue-items","id":"$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/job-production-plan-project-phase-revenue-items" \
    -d "$jpppri_payload"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" && "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan project phase revenue item but no ID returned"
        run_tests
    fi
else
    fail "Failed to create job production plan project phase revenue item"
    run_tests
fi

# Create project phase revenue item actual (direct API)
ACTUAL_QUANTITY="9.99"
revenue_item_actual_payload=$(cat <<JSON
{"data":{"type":"project-phase-revenue-item-actuals","attributes":{"quantity":"$ACTUAL_QUANTITY","revenue-date":"$TODAY"},"relationships":{"project-phase-revenue-item":{"data":{"type":"project-phase-revenue-items","id":"$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID"}},"job-production-plan-project-phase-revenue-item":{"data":{"type":"job-production-plan-project-phase-revenue-items","id":"$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-phase-revenue-item-actuals" \
    -d "$revenue_item_actual_payload"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" != "null" ]]; then
        pass
    else
        fail "Created project phase revenue item actual but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project phase revenue item actual"
    run_tests
fi

# Create second project phase revenue item actual (direct API)
revenue_item_actual_payload_2=$(cat <<JSON
{"data":{"type":"project-phase-revenue-item-actuals","attributes":{"quantity":"10.5","revenue-date":"$TODAY"},"relationships":{"project-phase-revenue-item":{"data":{"type":"project-phase-revenue-items","id":"$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID"}},"job-production-plan-project-phase-revenue-item":{"data":{"type":"job-production-plan-project-phase-revenue-items","id":"$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-phase-revenue-item-actuals" \
    -d "$revenue_item_actual_payload_2"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2" != "null" ]]; then
        pass
    else
        fail "Created second project phase revenue item actual but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second project phase revenue item actual"
    run_tests
fi

# Create project phase cost item actual
PRICE_PER_UNIT="15.25"
create_cmd=(do project-phase-cost-item-actuals create \
    --project-phase-cost-item "$CREATED_PROJECT_PHASE_COST_ITEM_ID" \
    --project-phase-revenue-item-actual "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" \
    --quantity "$ACTUAL_QUANTITY" \
    --price-per-unit-explicit "$PRICE_PER_UNIT")

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    create_cmd+=(--created-by "$CREATED_BY_ID")
fi

xbe_json "${create_cmd[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" && "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" != "null" ]]; then
        register_cleanup "project-phase-cost-item-actuals" "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID"
        pass
    else
        fail "Created project phase cost item actual but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project phase cost item actual"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project phase cost item actual"
xbe_json view project-phase-cost-item-actuals show "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project phase cost item actual quantity"
UPDATED_QUANTITY="12.25"
xbe_json do project-phase-cost-item-actuals update "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" --quantity "$UPDATED_QUANTITY"
assert_success

test_name "Update project phase cost item actual price-per-unit-explicit"
UPDATED_PRICE="22.5"
xbe_json do project-phase-cost-item-actuals update "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" --price-per-unit-explicit "$UPDATED_PRICE"
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "Update project phase cost item actual created-by"
    xbe_json do project-phase-cost-item-actuals update "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" --created-by "$CREATED_BY_ID"
    assert_success
fi

test_name "Update project phase cost item actual revenue item actual"
xbe_json do project-phase-cost-item-actuals update "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" --project-phase-revenue-item-actual "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List project phase cost item actuals"
xbe_json view project-phase-cost-item-actuals list --limit 5
assert_success

test_name "List project phase cost item actuals with --project-phase-cost-item"
xbe_json view project-phase-cost-item-actuals list --project-phase-cost-item "$CREATED_PROJECT_PHASE_COST_ITEM_ID" --limit 5
assert_success

test_name "List project phase cost item actuals with --project-phase-revenue-item-actual"
xbe_json view project-phase-cost-item-actuals list --project-phase-revenue-item-actual "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID_2" --limit 5
assert_success

test_name "List project phase cost item actuals with --job-production-plan"
xbe_json view project-phase-cost-item-actuals list --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" --limit 5
assert_success

test_name "List project phase cost item actuals with --job-production-plan-id"
xbe_json view project-phase-cost-item-actuals list --job-production-plan-id "$CREATED_JOB_PRODUCTION_PLAN_ID" --limit 5
assert_success

test_name "List project phase cost item actuals with --job-production-plan-project-phase-revenue-item"
xbe_json view project-phase-cost-item-actuals list --job-production-plan-project-phase-revenue-item "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
assert_success

test_name "List project phase cost item actuals with --job-production-plan-project-phase-revenue-item-id"
xbe_json view project-phase-cost-item-actuals list --job-production-plan-project-phase-revenue-item-id "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
assert_success

test_name "List project phase cost item actuals with --project"
xbe_json view project-phase-cost-item-actuals list --project "$CREATED_PROJECT_ID" --limit 5
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "List project phase cost item actuals with --created-by"
    xbe_json view project-phase-cost-item-actuals list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
fi

test_name "List project phase cost item actuals with --quantity"
xbe_json view project-phase-cost-item-actuals list --quantity "$UPDATED_QUANTITY" --limit 5
assert_success

# Pagination

test_name "List project phase cost item actuals with --limit"
xbe_json view project-phase-cost-item-actuals list --limit 2
assert_success

test_name "List project phase cost item actuals with --offset"
xbe_json view project-phase-cost-item-actuals list --limit 2 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project phase cost item actual requires --confirm"
xbe_json do project-phase-cost-item-actuals delete "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID"
assert_failure

test_name "Delete project phase cost item actual with --confirm"
xbe_json do project-phase-cost-item-actuals delete "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID" --confirm
if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to delete project phase cost item actual"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do project-phase-cost-item-actuals update "$CREATED_PROJECT_PHASE_COST_ITEM_ACTUAL_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
