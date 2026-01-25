#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Revenue Item Actuals
#
# Tests create/update/delete operations and list filters for project-phase-revenue-item-actuals.
# Requires creating supporting project and revenue resources.
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_PHASE_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_PRODUCTION_PLAN_ID=""
CREATED_PROJECT_REVENUE_CLASSIFICATION_ID=""
UNIT_OF_MEASURE_ID=""
CREATED_PROJECT_REVENUE_ITEM_ID=""
CREATED_PROJECT_PHASE_REVENUE_ITEM_ID=""
CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID=""
CREATED_BY_ID=""

SKIP_MUTATION=0
if [[ -z "$XBE_TOKEN" ]]; then
    SKIP_MUTATION=1
fi

describe "Resource: project-phase-revenue-item-actuals"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "PPriaBroker")

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
DEV_NAME=$(unique_name "PPriaDeveloper")

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
PROJECT_NAME=$(unique_name "PPriaProject")

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
PHASE_NAME=$(unique_name "PPriaPhase")

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

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create project revenue classification"
    REV_CLASS_NAME=$(unique_name "PPriaRevClass")
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
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create prerequisite customer"
    CUSTOMER_NAME=$(unique_name "PPriaCustomer")
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
    JOB_NAME=$(unique_name "PPriaJPP")
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
else
    TODAY=$(date +%Y-%m-%d)
fi

cleanup_project_phase_revenue_item_actuals() {
    if [[ -n "$XBE_TOKEN" ]]; then
        if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" && "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/job-production-plan-project-phase-revenue-items/$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" \
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
trap cleanup_project_phase_revenue_item_actuals EXIT

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project phase revenue item actual without required fields fails"
xbe_json do project-phase-revenue-item-actuals create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete/filter tests without XBE_TOKEN"
    test_name "List project phase revenue item actuals"
    xbe_json view project-phase-revenue-item-actuals list --limit 5
    assert_success
    run_tests
fi

if [[ -z "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID" ]]; then
    skip "Missing project revenue classification; skipping mutation tests"
    run_tests
fi

# Resolve current user for created-by filter tests
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CREATED_BY_ID=$(json_get ".id")
fi

# Create project revenue item (direct API)
PROJECT_REVENUE_ITEM_DESC=$(unique_name "PPriaRevItem")
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

# Create project phase revenue item actual
ACTUAL_QUANTITY="9.99"
xbe_json do project-phase-revenue-item-actuals create \
    --project-phase-revenue-item "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" \
    --job-production-plan-project-phase-revenue-item "$CREATED_JOB_PRODUCTION_PLAN_PROJECT_PHASE_REVENUE_ITEM_ID" \
    --quantity "$ACTUAL_QUANTITY" \
    --revenue-date "$TODAY"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" && "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" != "null" ]]; then
        register_cleanup "project-phase-revenue-item-actuals" "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID"
        pass
    else
        fail "Created project phase revenue item actual but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project phase revenue item actual"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project phase revenue item actual"
xbe_json view project-phase-revenue-item-actuals show "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project phase revenue item actual quantity"
UPDATED_QUANTITY="12.25"
xbe_json do project-phase-revenue-item-actuals update "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" --quantity "$UPDATED_QUANTITY"
assert_success

test_name "Update project phase revenue item actual revenue-date"
xbe_json do project-phase-revenue-item-actuals update "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" --revenue-date "$TODAY"
assert_success

test_name "Update project phase revenue item actual quantity strategy"
xbe_json do project-phase-revenue-item-actuals update "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" --quantity-strategy-explicit direct
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "Update project phase revenue item actual created-by"
    xbe_json do project-phase-revenue-item-actuals update "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" --created-by "$CREATED_BY_ID"
    assert_success
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List project phase revenue item actuals"
xbe_json view project-phase-revenue-item-actuals list --limit 5
assert_success

test_name "List project phase revenue item actuals with --project-phase-revenue-item"
xbe_json view project-phase-revenue-item-actuals list --project-phase-revenue-item "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
assert_success

test_name "List project phase revenue item actuals with --job-production-plan"
xbe_json view project-phase-revenue-item-actuals list --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" --limit 5
assert_success

test_name "List project phase revenue item actuals with --project"
xbe_json view project-phase-revenue-item-actuals list --project "$CREATED_PROJECT_ID" --limit 5
assert_success

test_name "List project phase revenue item actuals with --revenue-date"
xbe_json view project-phase-revenue-item-actuals list --revenue-date "$TODAY" --limit 5
assert_success

test_name "List project phase revenue item actuals with --revenue-date-min"
xbe_json view project-phase-revenue-item-actuals list --revenue-date-min "$TODAY" --limit 5
assert_success

test_name "List project phase revenue item actuals with --revenue-date-max"
xbe_json view project-phase-revenue-item-actuals list --revenue-date-max "$TODAY" --limit 5
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "List project phase revenue item actuals with --created-by"
    xbe_json view project-phase-revenue-item-actuals list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
fi

test_name "List project phase revenue item actuals with --limit"
xbe_json view project-phase-revenue-item-actuals list --limit 2
assert_success

test_name "List project phase revenue item actuals with --offset"
xbe_json view project-phase-revenue-item-actuals list --limit 2 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project phase revenue item actual requires --confirm"
xbe_json do project-phase-revenue-item-actuals delete "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID"
assert_failure

test_name "Delete project phase revenue item actual with --confirm"
xbe_json do project-phase-revenue-item-actuals delete "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID" --confirm
if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to delete project phase revenue item actual"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do project-phase-revenue-item-actuals update "$CREATED_PROJECT_PHASE_REVENUE_ITEM_ACTUAL_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
