#!/bin/bash
#
# XBE CLI Integration Tests: Project Truckers
#
# Tests CRUD operations for the project_truckers resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRUCKER_ID_2=""
CREATED_PROJECT_TRUCKER_ID=""
CREATED_PROJECT_TRUCKER_ID_2=""

describe "Resource: project_truckers"

# ============================================================================
# Prerequisites - Create broker, developer, project, trucker
# ============================================================================

test_name "Create prerequisite broker for project trucker tests"
BROKER_NAME=$(unique_name "ProjectTruckerBroker")

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

test_name "Create developer for project trucker tests"
if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
    CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
    echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
    pass
else
    DEV_NAME=$(unique_name "ProjectTruckerDev")
    xbe_json do developers create \
        --name "$DEV_NAME" \
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
fi

test_name "Create project for project trucker tests"
if [[ -n "$XBE_TEST_PROJECT_ID" ]]; then
    CREATED_PROJECT_ID="$XBE_TEST_PROJECT_ID"
    echo "    Using XBE_TEST_PROJECT_ID: $CREATED_PROJECT_ID"
    pass
else
    PROJECT_NAME=$(unique_name "ProjectTruckerProject")
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
fi

test_name "Create trucker for project trucker tests"
if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
    CREATED_TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
    echo "    Using XBE_TEST_TRUCKER_ID: $CREATED_TRUCKER_ID"
    pass
else
    TRUCKER_NAME=$(unique_name "ProjectTrucker")
    xbe_json do truckers create \
        --name "$TRUCKER_NAME" \
        --broker "$CREATED_BROKER_ID" \
        --company-address "123 Project Trucker St"

    if [[ $status -eq 0 ]]; then
        CREATED_TRUCKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
            register_cleanup "truckers" "$CREATED_TRUCKER_ID"
            pass
        else
            fail "Created trucker but no ID returned"
            echo "Cannot continue without a trucker"
            run_tests
        fi
    else
        fail "Failed to create trucker"
        echo "Cannot continue without a trucker"
        run_tests
    fi
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create project trucker with required fields"
xbe_json do project-truckers create \
    --project "$CREATED_PROJECT_ID" \
    --trucker "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_TRUCKER_ID" && "$CREATED_PROJECT_TRUCKER_ID" != "null" ]]; then
        register_cleanup "project-truckers" "$CREATED_PROJECT_TRUCKER_ID"
        pass
    else
        fail "Created project trucker but no ID returned"
    fi
else
    fail "Failed to create project trucker"
fi

if [[ -z "$CREATED_PROJECT_TRUCKER_ID" || "$CREATED_PROJECT_TRUCKER_ID" == "null" ]]; then
    echo "Cannot continue without a valid project trucker ID"
    run_tests
fi

test_name "Create second trucker for optional create fields"
TRUCKER_NAME_2=$(unique_name "ProjectTrucker2")
xbe_json do truckers create \
    --name "$TRUCKER_NAME_2" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "456 Project Trucker Ave"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID_2" && "$CREATED_TRUCKER_ID_2" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID_2"
        pass
    else
        fail "Created trucker but no ID returned"
    fi
else
    fail "Failed to create second trucker"
fi

if [[ -n "$CREATED_TRUCKER_ID_2" && "$CREATED_TRUCKER_ID_2" != "null" ]]; then
    test_name "Create project trucker with exclusion flag"
    xbe_json do project-truckers create \
        --project "$CREATED_PROJECT_ID" \
        --trucker "$CREATED_TRUCKER_ID_2" \
        --is-excluded-from-time-card-payroll-certification-requirements=true

    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_TRUCKER_ID_2=$(json_get ".id")
        if [[ -n "$CREATED_PROJECT_TRUCKER_ID_2" && "$CREATED_PROJECT_TRUCKER_ID_2" != "null" ]]; then
            register_cleanup "project-truckers" "$CREATED_PROJECT_TRUCKER_ID_2"
            pass
        else
            fail "Created project trucker but no ID returned"
        fi
    else
        fail "Failed to create project trucker with exclusion flag"
    fi
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update project trucker exclusion flag to true"
xbe_json do project-truckers update "$CREATED_PROJECT_TRUCKER_ID" \
    --is-excluded-from-time-card-payroll-certification-requirements=true
assert_success

test_name "Update project trucker exclusion flag to false"
xbe_json do project-truckers update "$CREATED_PROJECT_TRUCKER_ID" \
    --is-excluded-from-time-card-payroll-certification-requirements=false
assert_success

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show project trucker"
xbe_json view project-truckers show "$CREATED_PROJECT_TRUCKER_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List project truckers"
xbe_json view project-truckers list --limit 5
assert_success

test_name "List project truckers returns array"
xbe_json view project-truckers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project truckers"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List project truckers with --project filter"
xbe_json view project-truckers list --project "$CREATED_PROJECT_ID" --limit 5
assert_success

test_name "List project truckers with --trucker filter"
xbe_json view project-truckers list --trucker "$CREATED_TRUCKER_ID" --limit 5
assert_success

test_name "List project truckers with exclusion flag filter"
xbe_json view project-truckers list \
    --is-excluded-from-time-card-payroll-certification-requirements false \
    --limit 5
assert_success

test_name "List project truckers with --created-at-min filter"
xbe_json view project-truckers list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project truckers with --created-at-max filter"
xbe_json view project-truckers list --created-at-max "2100-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project truckers with --updated-at-min filter"
xbe_json view project-truckers list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project truckers with --updated-at-max filter"
xbe_json view project-truckers list --updated-at-max "2100-01-01T00:00:00Z" --limit 5
assert_success

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create project trucker without project fails"
xbe_json do project-truckers create --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create project trucker without trucker fails"
xbe_json do project-truckers create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update project trucker without any fields fails"
xbe_json do project-truckers update "$CREATED_PROJECT_TRUCKER_ID"
assert_failure

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete project trucker requires --confirm flag"
xbe_run do project-truckers delete "$CREATED_PROJECT_TRUCKER_ID"
assert_failure

test_name "Delete project trucker with --confirm"
xbe_run do project-truckers delete "$CREATED_PROJECT_TRUCKER_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
