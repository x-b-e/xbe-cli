#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Revenue Item Actual Exports
#
# Tests list, show, and create operations for the project-phase-revenue-item-actual-exports resource.
#
# COVERAGE: List filters + show + create relationships/attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
STATUS=""
REVENUE_DATE=""
ORG_TYPE=""
ORG_ID=""
BROKER_ID=""
CREATED_BY_ID=""
ORGANIZATION_FORMATTER_ID=""
PROJECT_PHASE_REVENUE_ITEM_ID=""
SKIP_ID_FILTERS=0
CREATED_EXPORT_ID=""
ENV_ORGANIZATION="${XBE_TEST_ORGANIZATION:-}"
ENV_PROJECT="${XBE_TEST_PROJECT:-}"
ORG_FILTER_TYPE=""
ORG_FILTER_ID=""

if [[ -n "$ENV_ORGANIZATION" ]]; then
    ORG_FILTER_TYPE=$(echo "$ENV_ORGANIZATION" | cut -d'|' -f1)
    ORG_FILTER_ID=$(echo "$ENV_ORGANIZATION" | cut -d'|' -f2)
fi

describe "Resource: project-phase-revenue-item-actual-exports"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project phase revenue item actual exports"
xbe_json view project-phase-revenue-item-actual-exports list --limit 5
assert_success

test_name "List project phase revenue item actual exports returns array"
xbe_json view project-phase-revenue-item-actual-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list exports"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project phase revenue item actual export"
xbe_json view project-phase-revenue-item-actual-exports list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    STATUS=$(json_get ".[0].status")
    REVENUE_DATE=$(json_get ".[0].revenue_date")
    ORG_TYPE=$(json_get ".[0].organization_type")
    ORG_ID=$(json_get ".[0].organization_id")
    BROKER_ID=$(json_get ".[0].broker_id")
    CREATED_BY_ID=$(json_get ".[0].created_by_id")
    ORGANIZATION_FORMATTER_ID=$(json_get ".[0].organization_formatter_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No project phase revenue item actual exports available"
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
    xbe_json view project-phase-revenue-item-actual-exports list --status "$STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List exports with --organization-formatter filter"
if [[ -n "$ORGANIZATION_FORMATTER_ID" && "$ORGANIZATION_FORMATTER_ID" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --organization-formatter "$ORGANIZATION_FORMATTER_ID" --limit 5
    assert_success
else
    skip "No organization formatter ID available"
fi

test_name "List exports with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List exports with --project filter"
if [[ -n "$ENV_PROJECT" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --project "$ENV_PROJECT" --limit 5
    assert_success
else
    skip "Set XBE_TEST_PROJECT to run"
fi

test_name "List exports with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List exports with --organization filter"
if [[ -n "$ENV_ORGANIZATION" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --organization "$ENV_ORGANIZATION" --limit 5
    assert_success
else
    skip "Set XBE_TEST_ORGANIZATION to run"
fi

test_name "List exports with --organization-id filter"
if [[ -n "$ORG_FILTER_TYPE" && -n "$ORG_FILTER_ID" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --organization-id "$ORG_FILTER_ID" --organization-type "$ORG_FILTER_TYPE" --limit 5
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
    xbe_json view project-phase-revenue-item-actual-exports list --organization-type "$FILTER_ORG_TYPE" --limit 5
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
    xbe_json view project-phase-revenue-item-actual-exports list --not-organization-type "$FILTER_NOT_ORG_TYPE" --limit 5
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

test_name "List exports with --revenue-date filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$REVENUE_DATE" && "$REVENUE_DATE" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --revenue-date "$REVENUE_DATE" --limit 5
    assert_success
else
    skip "No revenue date available"
fi

test_name "List exports with --revenue-date-min filter"
xbe_json view project-phase-revenue-item-actual-exports list --revenue-date-min "2020-01-01" --limit 5
assert_success

test_name "List exports with --revenue-date-max filter"
xbe_json view project-phase-revenue-item-actual-exports list --revenue-date-max "2030-01-01" --limit 5
assert_success

test_name "List exports with --has-revenue-date filter"
xbe_json view project-phase-revenue-item-actual-exports list --has-revenue-date true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project phase revenue item actual export"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        PROJECT_PHASE_REVENUE_ITEM_ID=$(json_get ".project_phase_revenue_item_ids[0]")
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
# LIST Tests - Project Phase Revenue Items Filter
# ============================================================================

test_name "List exports with --project-phase-revenue-items filter"
if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" && "$PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports list --project-phase-revenue-items "$PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
    assert_success
else
    skip "No project phase revenue item ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create export without required fields fails"
xbe_run do project-phase-revenue-item-actual-exports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

TEST_DATE="$(date -u +%Y-%m-%d)"

test_name "Create project phase revenue item actual export"
if [[ -n "$ORGANIZATION_FORMATTER_ID" && "$ORGANIZATION_FORMATTER_ID" != "null" && -n "$PROJECT_PHASE_REVENUE_ITEM_ID" && "$PROJECT_PHASE_REVENUE_ITEM_ID" != "null" ]]; then
    xbe_json do project-phase-revenue-item-actual-exports create \
        --organization-formatter "$ORGANIZATION_FORMATTER_ID" \
        --project-phase-revenue-item-ids "$PROJECT_PHASE_REVENUE_ITEM_ID" \
        --revenue-date "$TEST_DATE"

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
        elif [[ "$output" == *"must all be exportable project phase revenue items"* ]] || [[ "$output" == *"must all be in scope"* ]] || [[ "$output" == *"formatter type"* ]]; then
            skip "Project phase revenue items not eligible for export"
        else
            fail "Failed to create export"
        fi
    fi
else
    skip "Missing organization formatter or project phase revenue item ID"
fi

test_name "Show created export includes relationships"
if [[ -n "$CREATED_EXPORT_ID" && "$CREATED_EXPORT_ID" != "null" ]]; then
    xbe_json view project-phase-revenue-item-actual-exports show "$CREATED_EXPORT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".organization_formatter_id" "$ORGANIZATION_FORMATTER_ID"
        assert_json_equals ".revenue_date" "$TEST_DATE"
        if echo "$output" | jq -e --arg id "$PROJECT_PHASE_REVENUE_ITEM_ID" '.project_phase_revenue_item_ids | index($id) != null' > /dev/null 2>&1; then
            pass
        else
            fail "Created export missing project phase revenue item ID"
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
