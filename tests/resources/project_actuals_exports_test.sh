#!/bin/bash
#
# XBE CLI Integration Tests: Project Actuals Exports
#
# Tests list, show, and create operations for the project-actuals-exports resource.
#
# COVERAGE: List filters + show + create relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
STATUS=""
ORG_TYPE=""
ORG_ID=""
BROKER_ID=""
PROJECT_ID=""
CREATED_BY_ID=""
ORGANIZATION_FORMATTER_ID=""
JOB_PRODUCTION_PLAN_ID=""
SKIP_ID_FILTERS=0
CREATED_EXPORT_ID=""
ENV_ORGANIZATION="${XBE_TEST_ORGANIZATION:-}"
ORG_FILTER_TYPE=""
ORG_FILTER_ID=""

if [[ -n "$ENV_ORGANIZATION" ]]; then
    ORG_FILTER_TYPE=$(echo "$ENV_ORGANIZATION" | cut -d'|' -f1)
    ORG_FILTER_ID=$(echo "$ENV_ORGANIZATION" | cut -d'|' -f2)
fi

describe "Resource: project-actuals-exports"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project actuals exports"
xbe_json view project-actuals-exports list --limit 5
assert_success

test_name "List project actuals exports returns array"
xbe_json view project-actuals-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list exports"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project actuals export"
xbe_json view project-actuals-exports list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    STATUS=$(json_get ".[0].status")
    ORG_TYPE=$(json_get ".[0].organization_type")
    ORG_ID=$(json_get ".[0].organization_id")
    BROKER_ID=$(json_get ".[0].broker_id")
    PROJECT_ID=$(json_get ".[0].project_id")
    CREATED_BY_ID=$(json_get ".[0].created_by_id")
    ORGANIZATION_FORMATTER_ID=$(json_get ".[0].organization_formatter_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No project actuals exports available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list exports"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List exports with --status filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$STATUS" && "$STATUS" != "null" ]]; then
    xbe_json view project-actuals-exports list --status "$STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List exports with --organization-formatter filter"
if [[ -n "$ORGANIZATION_FORMATTER_ID" && "$ORGANIZATION_FORMATTER_ID" != "null" ]]; then
    xbe_json view project-actuals-exports list --organization-formatter "$ORGANIZATION_FORMATTER_ID" --limit 5
    assert_success
else
    skip "No organization formatter ID available"
fi

test_name "List exports with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view project-actuals-exports list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List exports with --project filter"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view project-actuals-exports list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List exports with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view project-actuals-exports list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List exports with --organization filter"
if [[ -n "$ENV_ORGANIZATION" ]]; then
    xbe_json view project-actuals-exports list --organization "$ENV_ORGANIZATION" --limit 5
    assert_success
else
    skip "Set XBE_TEST_ORGANIZATION to run"
fi

test_name "List exports with --organization-id filter"
if [[ -n "$ORG_FILTER_TYPE" && -n "$ORG_FILTER_ID" ]]; then
    xbe_json view project-actuals-exports list --organization-id "$ORG_FILTER_ID" --organization-type "$ORG_FILTER_TYPE" --limit 5
    assert_success
else
    skip "Set XBE_TEST_ORGANIZATION to run"
fi

test_name "List exports with --organization-type filter"
FILTER_ORG_TYPE="$ORG_FILTER_TYPE"
if [[ -z "$FILTER_ORG_TYPE" || "$FILTER_ORG_TYPE" == "null" ]]; then
    FILTER_ORG_TYPE="$ORG_TYPE"
fi
if [[ -n "$FILTER_ORG_TYPE" && "$FILTER_ORG_TYPE" != "null" ]]; then
    xbe_json view project-actuals-exports list --organization-type "$FILTER_ORG_TYPE" --limit 5
    assert_success
else
    skip "No organization type available"
fi

test_name "List exports with --not-organization-type filter"
FILTER_NOT_ORG_TYPE="$ORG_FILTER_TYPE"
if [[ -z "$FILTER_NOT_ORG_TYPE" || "$FILTER_NOT_ORG_TYPE" == "null" ]]; then
    FILTER_NOT_ORG_TYPE="$ORG_TYPE"
fi
if [[ -n "$FILTER_NOT_ORG_TYPE" && "$FILTER_NOT_ORG_TYPE" != "null" ]]; then
    xbe_json view project-actuals-exports list --not-organization-type "$FILTER_NOT_ORG_TYPE" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    elif [[ "$output" == *"not_organization_type"* ]] || [[ "$output" == *"Internal Server Error"* ]]; then
        skip "Server does not support not-organization-type filter"
    else
        fail "Failed to list exports with not-organization-type filter"
    fi
else
    skip "No organization type available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project actuals export"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-actuals-exports show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        JOB_PRODUCTION_PLAN_ID=$(json_get ".job_production_plan_ids[0]")
        if [[ -z "$ORGANIZATION_FORMATTER_ID" || "$ORGANIZATION_FORMATTER_ID" == "null" ]]; then
            ORGANIZATION_FORMATTER_ID=$(json_get ".organization_formatter_id")
        fi
        pass
    else
        fail "Failed to show export"
    fi
else
    skip "No export ID available"
fi

# ============================================================================
# LIST Tests - Job Production Plans Filter
# ============================================================================

test_name "List exports with --job-production-plans filter"
if [[ -n "$JOB_PRODUCTION_PLAN_ID" && "$JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view project-actuals-exports list --job-production-plans "$JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create export without required fields fails"
xbe_run do project-actuals-exports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project actuals export"
if [[ -n "$ORGANIZATION_FORMATTER_ID" && "$ORGANIZATION_FORMATTER_ID" != "null" && -n "$JOB_PRODUCTION_PLAN_ID" && "$JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json do project-actuals-exports create \
        --organization-formatter "$ORGANIZATION_FORMATTER_ID" \
        --job-production-plan-ids "$JOB_PRODUCTION_PLAN_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_EXPORT_ID=$(json_get ".id")
        if [[ -n "$CREATED_EXPORT_ID" && "$CREATED_EXPORT_ID" != "null" ]]; then
            pass
        else
            fail "Created export but no ID returned"
        fi
    else
        if [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to create export"
        elif [[ "$output" == *"must all be exportable job production plans"* ]] || [[ "$output" == *"must all be in scope"* ]] || [[ "$output" == *"formatter type"* ]]; then
            skip "Job production plans not eligible for export"
        else
            fail "Failed to create export"
        fi
    fi
else
    skip "Missing organization formatter or job production plan ID"
fi

test_name "Show created export includes relationships"
if [[ -n "$CREATED_EXPORT_ID" && "$CREATED_EXPORT_ID" != "null" ]]; then
    xbe_json view project-actuals-exports show "$CREATED_EXPORT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".organization_formatter_id" "$ORGANIZATION_FORMATTER_ID"
        if echo "$output" | jq -e --arg id "$JOB_PRODUCTION_PLAN_ID" '.job_production_plan_ids | index($id) != null' > /dev/null 2>&1; then
            pass
        else
            fail "Created export missing job production plan ID"
        fi
    else
        fail "Failed to show created export"
    fi
else
    skip "No created export ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
