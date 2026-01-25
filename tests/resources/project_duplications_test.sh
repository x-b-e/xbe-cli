#!/bin/bash
#
# XBE CLI Integration Tests: Project Duplications
#
# Tests create operations for the project-duplications resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_DUPLICATED_PROJECT_ID=""

describe "Resource: project-duplications"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project duplication without required project template fails"
xbe_run do project-duplications create
assert_failure

# ============================================================================
# Prerequisites - Create broker, developer, and project
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "ProjectDuplicationBroker")

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
DEVELOPER_NAME=$(unique_name "ProjectDuplicationDeveloper")

xbe_json do developers create \
    --name "$DEVELOPER_NAME" \
    --broker "$CREATED_BROKER_ID"

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

test_name "Create template project"
PROJECT_NAME=$(unique_name "ProjectDuplicationTemplate")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project duplication"

xbe_json do project-duplications create \
    --project-template "$CREATED_PROJECT_ID"

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    CREATED_DUPLICATED_PROJECT_ID=$(json_get ".derived_project_id")
    if [[ -n "$CREATED_DUPLICATED_PROJECT_ID" && "$CREATED_DUPLICATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_DUPLICATED_PROJECT_ID"
    fi
else
    fail "Failed to create project duplication"
fi

# ============================================================================
# CREATE Tests - optional attributes + skip flags
# ============================================================================

test_name "Create project duplication with optional attributes and skip flags"

DERIVED_TEMPLATE_NAME=$(unique_name "ProjectDuplicationTemplateName")
DERIVED_PROJECT_NUMBER="DUP-$(date +%s)-$RANDOM"
DERIVED_DUE_ON="2026-02-01"

xbe_json do project-duplications create \
    --project-template "$CREATED_PROJECT_ID" \
    --new-developer "$CREATED_DEVELOPER_ID" \
    --derived-project-template-name "$DERIVED_TEMPLATE_NAME" \
    --derived-project-number "$DERIVED_PROJECT_NUMBER" \
    --derived-due-on "$DERIVED_DUE_ON" \
    --derived-is-prevailing-wage-applicable \
    --derived-is-time-card-payroll-certification-required \
    --skip-project-material-types \
    --skip-project-customers \
    --skip-project-truckers \
    --skip-project-trailer-classifications \
    --skip-project-labor-classifications \
    --skip-certification-requirements \
    --skip-project-cost-codes \
    --skip-project-revenue-items \
    --skip-project-phase-revenue-items

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".project_template_id" "$CREATED_PROJECT_ID"
    assert_json_equals ".new_developer_id" "$CREATED_DEVELOPER_ID"
    assert_json_equals ".derived_project_template_name" "$DERIVED_TEMPLATE_NAME"
    assert_json_equals ".derived_project_number" "$DERIVED_PROJECT_NUMBER"
    assert_json_equals ".derived_due_on" "$DERIVED_DUE_ON"
    assert_json_bool ".derived_is_prevailing_wage_applicable" "true"
    assert_json_bool ".derived_is_time_card_payroll_certification_required" "true"
    assert_json_bool ".skip_project_material_types" "true"
    assert_json_bool ".skip_project_customers" "true"
    assert_json_bool ".skip_project_truckers" "true"
    assert_json_bool ".skip_project_trailer_classifications" "true"
    assert_json_bool ".skip_project_labor_classifications" "true"
    assert_json_bool ".skip_certification_requirements" "true"
    assert_json_bool ".skip_project_cost_codes" "true"
    assert_json_bool ".skip_project_revenue_items" "true"
    assert_json_bool ".skip_project_phase_revenue_items" "true"
    DUP_PROJECT_ID=$(json_get ".derived_project_id")
    if [[ -n "$DUP_PROJECT_ID" && "$DUP_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$DUP_PROJECT_ID"
    fi
else
    fail "Failed to create project duplication with optional attributes"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
