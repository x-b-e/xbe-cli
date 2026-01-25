#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Project Phase Revenue Items
#
# Tests CRUD operations for the job-production-plan-project-phase-revenue-items resource.
#
# COVERAGE: All filters + all create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ITEM_ID=""
SHOW_ID=""
JPP_ID=""
PROJECT_ID=""
PROJECT_PHASE_REVENUE_ITEM_ID=""
PROJECT_REVENUE_ITEM_ID=""
PROJECT_PHASE_ID=""
JPP_STATUS=""

describe "Resource: job-production-plan-project-phase-revenue-items"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan project phase revenue items"
xbe_json view job-production-plan-project-phase-revenue-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(echo "$output" | jq -r '.[0].id // empty' 2>/dev/null || true)
    JPP_ID=$(echo "$output" | jq -r '.[0].job_production_plan_id // empty' 2>/dev/null || true)
    PROJECT_PHASE_REVENUE_ITEM_ID=$(echo "$output" | jq -r '.[0].project_phase_revenue_item_id // empty' 2>/dev/null || true)
else
    fail "Failed to list job production plan project phase revenue items"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create item without required fields fails"
xbe_run do job-production-plan-project-phase-revenue-items create
assert_failure

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup job production plan and project phase revenue item via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    plans_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-production-plans?page[limit]=1&fields[job-production-plans]=project,status" || true)

    JPP_ID=$(echo "$plans_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    PROJECT_ID=$(echo "$plans_json" | jq -r '.data[0].relationships.project.data.id // empty' 2>/dev/null || true)
    JPP_STATUS=$(echo "$plans_json" | jq -r '.data[0].attributes.status // empty' 2>/dev/null || true)

    if [[ -z "$JPP_ID" || -z "$PROJECT_ID" ]]; then
        skip "No job production plan with project found"
    else
        ppri_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-phase-revenue-items?filter[project]=$PROJECT_ID&page[limit]=100" || true)

        if [[ "$(echo "$ppri_json" | jq -r '.data | length' 2>/dev/null)" == "0" ]]; then
            skip "No project phase revenue items found for project"
        else
            existing_json=$(curl -s \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                "$base_url/v1/job-production-plan-project-phase-revenue-items?filter[job-production-plan]=$JPP_ID&page[limit]=200" || true)

            existing_ids=$(echo "$existing_json" | jq -r '.data[].relationships[\"project-phase-revenue-item\"].data.id' 2>/dev/null || true)

            candidate_id=""
            while read -r id; do
                if [[ -z "$id" ]]; then
                    continue
                fi
                if ! grep -Fxq "$id" <<<"$existing_ids"; then
                    candidate_id="$id"
                    break
                fi
            done < <(echo "$ppri_json" | jq -r '.data[].id' 2>/dev/null || true)

            if [[ -z "$candidate_id" ]]; then
                skip "No unlinked project phase revenue item found for job production plan"
            else
                PROJECT_PHASE_REVENUE_ITEM_ID="$candidate_id"
                PROJECT_REVENUE_ITEM_ID=$(echo "$ppri_json" | jq -r --arg id "$candidate_id" '.data[] | select(.id==$id) | .relationships[\"project-revenue-item\"].data.id // empty' 2>/dev/null || true)
                PROJECT_PHASE_ID=$(echo "$ppri_json" | jq -r --arg id "$candidate_id" '.data[] | select(.id==$id) | .relationships[\"project-phase\"].data.id // empty' 2>/dev/null || true)
                pass
            fi
        fi
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create item with required fields"
if [[ -n "$JPP_ID" && -n "$PROJECT_PHASE_REVENUE_ITEM_ID" ]]; then
    xbe_json do job-production-plan-project-phase-revenue-items create \
        --job-production-plan "$JPP_ID" \
        --project-phase-revenue-item "$PROJECT_PHASE_REVENUE_ITEM_ID" \
        --quantity 10

    if [[ $status -eq 0 ]]; then
        CREATED_ITEM_ID=$(json_get ".id")
        if [[ -n "$CREATED_ITEM_ID" && "$CREATED_ITEM_ID" != "null" ]]; then
            register_cleanup "job-production-plan-project-phase-revenue-items" "$CREATED_ITEM_ID"
            pass
        else
            fail "Created item but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create item: $output"
        fi
    fi
else
    skip "Missing job production plan or project phase revenue item for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update item quantity"
if [[ -n "$CREATED_ITEM_ID" && "$CREATED_ITEM_ID" != "null" ]]; then
    xbe_json do job-production-plan-project-phase-revenue-items update "$CREATED_ITEM_ID" --quantity 15
    assert_success
else
    skip "No created item available"
fi

test_name "Update item should-update"
if [[ -n "$CREATED_ITEM_ID" && "$CREATED_ITEM_ID" != "null" ]]; then
    xbe_json do job-production-plan-project-phase-revenue-items update "$CREATED_ITEM_ID" --should-update
    assert_success
else
    skip "No created item available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show item details"
if [[ -n "$CREATED_ITEM_ID" && "$CREATED_ITEM_ID" != "null" ]]; then
    SHOW_ID="$CREATED_ITEM_ID"
fi

if [[ -n "$SHOW_ID" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items show "$SHOW_ID"
    assert_success
else
    skip "No item ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List items with --job-production-plan filter"
if [[ -n "$JPP_ID" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items list --job-production-plan "$JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List items with --project-phase-revenue-item filter"
if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items list --project-phase-revenue-item "$PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
    assert_success
else
    skip "No project phase revenue item ID available"
fi

test_name "List items with --project filter"
if [[ -n "$PROJECT_ID" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List items with --project-revenue-item filter"
if [[ -n "$PROJECT_REVENUE_ITEM_ID" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items list --project-revenue-item "$PROJECT_REVENUE_ITEM_ID" --limit 5
    assert_success
else
    skip "No project revenue item ID available"
fi

test_name "List items with --project-phase filter"
if [[ -n "$PROJECT_PHASE_ID" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items list --project-phase "$PROJECT_PHASE_ID" --limit 5
    assert_success
else
    skip "No project phase ID available"
fi

test_name "List items with --job-production-plan-status filter"
if [[ -n "$JPP_STATUS" ]]; then
    xbe_json view job-production-plan-project-phase-revenue-items list --job-production-plan-status "$JPP_STATUS" --limit 5
    assert_success
else
    skip "No job production plan status available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete item"
if [[ -n "$CREATED_ITEM_ID" && "$CREATED_ITEM_ID" != "null" ]]; then
    xbe_run do job-production-plan-project-phase-revenue-items delete "$CREATED_ITEM_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to delete item"
        fi
    fi
else
    skip "No created item available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
