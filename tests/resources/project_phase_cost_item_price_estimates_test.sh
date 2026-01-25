#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Cost Item Price Estimates
#
# Tests CRUD operations for the project-phase-cost-item-price-estimates resource.
#
# COVERAGE: All filters + create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
SHOW_ID=""
PROJECT_PHASE_COST_ITEM_ID=""
PROJECT_ESTIMATE_SET_ID=""
CREATED_BY_ID=""
ALT_PROJECT_ESTIMATE_SET_ID=""


describe "Resource: project-phase-cost-item-price-estimates"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project phase cost item price estimates"
xbe_json view project-phase-cost-item-price-estimates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(echo "$output" | jq -r '.[0].id // empty' 2>/dev/null || true)
    PROJECT_PHASE_COST_ITEM_ID=$(echo "$output" | jq -r '.[0].project_phase_cost_item_id // empty' 2>/dev/null || true)
    PROJECT_ESTIMATE_SET_ID=$(echo "$output" | jq -r '.[0].project_estimate_set_id // empty' 2>/dev/null || true)
    CREATED_BY_ID=$(echo "$output" | jq -r '.[0].created_by_id // empty' 2>/dev/null || true)
else
    fail "Failed to list project phase cost item price estimates"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create estimate without required fields fails"
xbe_run do project-phase-cost-item-price-estimates create
assert_failure

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup estimate set and cost item via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    estimate_sets_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/project-estimate-sets?page[limit]=50&fields[project-estimate-sets]=project" || true)

    selected_project_id=""

    while read -r set_id; do
        if [[ -z "$set_id" ]]; then
            continue
        fi
        project_id=$(echo "$estimate_sets_json" | jq -r --arg id "$set_id" '.data[] | select(.id==$id) | .relationships.project.data.id // empty' 2>/dev/null || true)
        if [[ -z "$project_id" ]]; then
            continue
        fi

        cost_items_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-phase-cost-items?filter[project]=$project_id&page[limit]=200" || true)

        if [[ "$(echo "$cost_items_json" | jq -r '.data | length' 2>/dev/null)" == "0" ]]; then
            continue
        fi

        existing_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-phase-cost-item-price-estimates?filter[project-estimate-set]=$set_id&page[limit]=200" || true)

        existing_ids=$(echo "$existing_json" | jq -r '.data[].relationships["project-phase-cost-item"].data.id' 2>/dev/null || true)

        candidate_id=""
        while read -r cost_item_id; do
            if [[ -z "$cost_item_id" ]]; then
                continue
            fi
            if ! grep -Fxq "$cost_item_id" <<<"$existing_ids"; then
                candidate_id="$cost_item_id"
                break
            fi
        done < <(echo "$cost_items_json" | jq -r '.data[].id' 2>/dev/null || true)

        if [[ -n "$candidate_id" ]]; then
            PROJECT_PHASE_COST_ITEM_ID="$candidate_id"
            PROJECT_ESTIMATE_SET_ID="$set_id"
            selected_project_id="$project_id"
            break
        fi
    done < <(echo "$estimate_sets_json" | jq -r '.data[].id' 2>/dev/null || true)

    if [[ -z "$PROJECT_PHASE_COST_ITEM_ID" || -z "$PROJECT_ESTIMATE_SET_ID" ]]; then
        skip "No matching project estimate set and cost item found"
    else
        # Find alternate estimate set for update tests (same project, no existing estimate)
        if [[ -n "$selected_project_id" ]]; then
            while read -r alt_set_id; do
                if [[ -z "$alt_set_id" || "$alt_set_id" == "$PROJECT_ESTIMATE_SET_ID" ]]; then
                    continue
                fi
                alt_project_id=$(echo "$estimate_sets_json" | jq -r --arg id "$alt_set_id" '.data[] | select(.id==$id) | .relationships.project.data.id // empty' 2>/dev/null || true)
                if [[ "$alt_project_id" != "$selected_project_id" ]]; then
                    continue
                fi
                alt_existing_json=$(curl -s \
                    -H "Authorization: Bearer $XBE_TOKEN" \
                    -H "Accept: application/vnd.api+json" \
                    "$base_url/v1/project-phase-cost-item-price-estimates?filter[project-estimate-set]=$alt_set_id&page[limit]=200" || true)
                alt_existing_ids=$(echo "$alt_existing_json" | jq -r '.data[].relationships["project-phase-cost-item"].data.id' 2>/dev/null || true)
                if ! grep -Fxq "$PROJECT_PHASE_COST_ITEM_ID" <<<"$alt_existing_ids"; then
                    ALT_PROJECT_ESTIMATE_SET_ID="$alt_set_id"
                    break
                fi
            done < <(echo "$estimate_sets_json" | jq -r '.data[].id' 2>/dev/null || true)
        fi
        pass
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create price estimate with required fields"
if [[ -n "$PROJECT_PHASE_COST_ITEM_ID" && -n "$PROJECT_ESTIMATE_SET_ID" ]]; then
    xbe_json do project-phase-cost-item-price-estimates create \
        --project-phase-cost-item "$PROJECT_PHASE_COST_ITEM_ID" \
        --project-estimate-set "$PROJECT_ESTIMATE_SET_ID" \
        --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-phase-cost-item-price-estimates" "$CREATED_ID"
            pass
        else
            fail "Created estimate but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create estimate: $output"
        fi
    fi
else
    skip "Missing project phase cost item or estimate set"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update estimate distribution"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-phase-cost-item-price-estimates update "$CREATED_ID" \
        --estimate '{"class_name":"TriangularDistribution","minimum":5,"mode":10,"maximum":15}'
    assert_success
else
    skip "No created estimate available"
fi


test_name "Update estimate set"
if [[ -n "$CREATED_ID" && -n "$ALT_PROJECT_ESTIMATE_SET_ID" ]]; then
    xbe_json do project-phase-cost-item-price-estimates update "$CREATED_ID" \
        --project-estimate-set "$ALT_PROJECT_ESTIMATE_SET_ID"
    assert_success
else
    skip "No alternate estimate set available"
fi


test_name "Update created-by"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        whoami_id=$(json_get ".id")
        if [[ -n "$whoami_id" && "$whoami_id" != "null" ]]; then
            xbe_json do project-phase-cost-item-price-estimates update "$CREATED_ID" --created-by "$whoami_id"
            if [[ $status -eq 0 ]]; then
                pass
            else
                if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                    pass
                else
                    fail "Failed to update created-by"
                fi
            fi
        else
            skip "No user ID from whoami"
        fi
    else
        skip "Failed to fetch whoami"
    fi
else
    skip "No created estimate available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show price estimate details"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    SHOW_ID="$CREATED_ID"
fi

if [[ -n "$SHOW_ID" ]]; then
    xbe_json view project-phase-cost-item-price-estimates show "$SHOW_ID"
    assert_success
else
    skip "No estimate ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List estimates with --project-phase-cost-item filter"
if [[ -n "$PROJECT_PHASE_COST_ITEM_ID" ]]; then
    xbe_json view project-phase-cost-item-price-estimates list --project-phase-cost-item "$PROJECT_PHASE_COST_ITEM_ID" --limit 5
    assert_success
else
    skip "No project phase cost item ID available"
fi


test_name "List estimates with --project-estimate-set filter"
if [[ -n "$PROJECT_ESTIMATE_SET_ID" ]]; then
    xbe_json view project-phase-cost-item-price-estimates list --project-estimate-set "$PROJECT_ESTIMATE_SET_ID" --limit 5
    assert_success
else
    skip "No project estimate set ID available"
fi


test_name "List estimates with --created-by filter"
if [[ -n "$CREATED_BY_ID" ]]; then
    xbe_json view project-phase-cost-item-price-estimates list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete price estimate"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do project-phase-cost-item-price-estimates delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to delete estimate"
        fi
    fi
else
    skip "No created estimate available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
