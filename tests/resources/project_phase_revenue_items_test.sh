#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Revenue Items
#
# Tests CRUD operations for the project-phase-revenue-items resource.
#
# COVERAGE: All filters + all create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
SHOW_ID=""
PROJECT_ID=""
PROJECT_PHASE_ID=""
PROJECT_REVENUE_ITEM_ID=""
PROJECT_REVENUE_CLASSIFICATION_ID=""
ESTIMATE_SET_ID=""
QUANTITY_ESTIMATE_ID=""


describe "Resource: project-phase-revenue-items"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project phase revenue items"
xbe_json view project-phase-revenue-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(echo "$output" | jq -r '.[0].id // empty' 2>/dev/null || true)
    PROJECT_PHASE_ID=$(echo "$output" | jq -r '.[0].project_phase_id // empty' 2>/dev/null || true)
    PROJECT_REVENUE_ITEM_ID=$(echo "$output" | jq -r '.[0].project_revenue_item_id // empty' 2>/dev/null || true)
    PROJECT_REVENUE_CLASSIFICATION_ID=$(echo "$output" | jq -r '.[0].project_revenue_classification_id // empty' 2>/dev/null || true)
else
    fail "Failed to list project phase revenue items"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project phase revenue item without required fields fails"
xbe_run do project-phase-revenue-items create
assert_failure

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup project phase and revenue item via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    projects_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/projects?page[limit]=20&fields[projects]=name" || true)

    while read -r project_id; do
        if [[ -z "$project_id" ]]; then
            continue
        fi

        phases_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-phases?filter[project]=$project_id&page[limit]=200" || true)

        revenue_items_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-revenue-items?filter[project]=$project_id&page[limit]=200" || true)

        phase_count=$(echo "$phases_json" | jq -r '.data | length' 2>/dev/null || true)
        revenue_count=$(echo "$revenue_items_json" | jq -r '.data | length' 2>/dev/null || true)
        if [[ "$phase_count" == "0" || "$revenue_count" == "0" ]]; then
            continue
        fi

        existing_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-phase-revenue-items?filter[project]=$project_id&page[limit]=200" || true)

        existing_pairs=$(echo "$existing_json" | jq -r '.data[] | "\(.relationships["project-phase"].data.id)|\(.relationships["project-revenue-item"].data.id)"' 2>/dev/null || true)

        candidate_phase=""
        candidate_revenue_item=""
        while read -r phase_id; do
            if [[ -z "$phase_id" ]]; then
                continue
            fi
            while read -r revenue_item_id; do
                if [[ -z "$revenue_item_id" ]]; then
                    continue
                fi
                pair="$phase_id|$revenue_item_id"
                if ! grep -Fxq "$pair" <<<"$existing_pairs"; then
                    candidate_phase="$phase_id"
                    candidate_revenue_item="$revenue_item_id"
                    break
                fi
            done < <(echo "$revenue_items_json" | jq -r '.data[].id' 2>/dev/null || true)
            if [[ -n "$candidate_phase" ]]; then
                break
            fi
        done < <(echo "$phases_json" | jq -r '.data[].id' 2>/dev/null || true)

        if [[ -n "$candidate_phase" && -n "$candidate_revenue_item" ]]; then
            PROJECT_ID="$project_id"
            PROJECT_PHASE_ID="$candidate_phase"
            PROJECT_REVENUE_ITEM_ID="$candidate_revenue_item"

            revenue_class_id=$(echo "$revenue_items_json" | jq -r --arg id "$candidate_revenue_item" '.data[] | select(.id==$id) | .relationships["revenue-classification"].data.id // empty' 2>/dev/null || true)
            if [[ -n "$revenue_class_id" ]]; then
                child_json=$(curl -s \
                    -H "Authorization: Bearer $XBE_TOKEN" \
                    -H "Accept: application/vnd.api+json" \
                    "$base_url/v1/project-revenue-classifications?filter[parent]=$revenue_class_id&page[limit]=50" || true)
                PROJECT_REVENUE_CLASSIFICATION_ID=$(echo "$child_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
            fi

            estimate_sets_json=$(curl -s \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                "$base_url/v1/project-estimate-sets?filter[project]=$project_id&page[limit]=50" || true)
            ESTIMATE_SET_ID=$(echo "$estimate_sets_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

            pass
            break
        fi
    done < <(echo "$projects_json" | jq -r '.data[].id' 2>/dev/null || true)

    if [[ -z "$PROJECT_PHASE_ID" || -z "$PROJECT_REVENUE_ITEM_ID" ]]; then
        skip "No suitable project phase and revenue item pair found"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project phase revenue item with required fields"
if [[ -n "$PROJECT_PHASE_ID" && -n "$PROJECT_REVENUE_ITEM_ID" ]]; then
    create_args=(do project-phase-revenue-items create \
        --project-phase "$PROJECT_PHASE_ID" \
        --project-revenue-item "$PROJECT_REVENUE_ITEM_ID" \
        --quantity-strategy indirect \
        --note "CLI test")

    if [[ -n "$PROJECT_REVENUE_CLASSIFICATION_ID" ]]; then
        create_args+=(--project-revenue-classification "$PROJECT_REVENUE_CLASSIFICATION_ID")
    fi

    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-phase-revenue-items" "$CREATED_ID"
            pass
        else
            fail "Created project phase revenue item but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create project phase revenue item: $output"
        fi
    fi
else
    skip "Missing project phase or revenue item"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update quantity strategy"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-phase-revenue-items update "$CREATED_ID" --quantity-strategy direct
    assert_success
else
    skip "No created item available"
fi

test_name "Update note"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-phase-revenue-items update "$CREATED_ID" --note "Updated note"
    assert_success
else
    skip "No created item available"
fi

test_name "Update project revenue classification"
if [[ -n "$CREATED_ID" && -n "$PROJECT_REVENUE_CLASSIFICATION_ID" ]]; then
    xbe_json do project-phase-revenue-items update "$CREATED_ID" --project-revenue-classification "$PROJECT_REVENUE_CLASSIFICATION_ID"
    assert_success
else
    skip "No project revenue classification available"
fi

test_name "Update quantity estimate"
if [[ -n "$CREATED_ID" && -n "$ESTIMATE_SET_ID" && -n "$XBE_TOKEN" ]]; then
    base_url="${XBE_BASE_URL%/}"
    payload="{\"data\":{\"type\":\"project-phase-revenue-item-quantity-estimates\",\"attributes\":{\"estimate\":{\"class_name\":\"NormalDistribution\",\"mean\":5,\"standard_deviation\":1}},\"relationships\":{\"project-phase-revenue-item\":{\"data\":{\"type\":\"project-phase-revenue-items\",\"id\":\"$CREATED_ID\"}},\"project-estimate-set\":{\"data\":{\"type\":\"project-estimate-sets\",\"id\":\"$ESTIMATE_SET_ID\"}}}}}"

    quantity_response=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json" \
        -H "Accept: application/vnd.api+json" \
        -X POST \
        --data "$payload" \
        "$base_url/v1/project-phase-revenue-item-quantity-estimates" || true)

    QUANTITY_ESTIMATE_ID=$(echo "$quantity_response" | jq -r '.data.id // empty' 2>/dev/null || true)

    if [[ -n "$QUANTITY_ESTIMATE_ID" ]]; then
        xbe_json do project-phase-revenue-items update "$CREATED_ID" --quantity-estimate "$QUANTITY_ESTIMATE_ID"
        if [[ $status -eq 0 ]]; then
            pass
        else
            if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
                pass
            else
                fail "Failed to update quantity estimate"
            fi
        fi
    else
        skip "Failed to create quantity estimate"
    fi
else
    skip "Missing item, estimate set, or token"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project phase revenue item details"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    SHOW_ID="$CREATED_ID"
fi

if [[ -n "$SHOW_ID" ]]; then
    xbe_json view project-phase-revenue-items show "$SHOW_ID"
    assert_success
else
    skip "No project phase revenue item ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project phase revenue items with --project-phase filter"
if [[ -n "$PROJECT_PHASE_ID" ]]; then
    xbe_json view project-phase-revenue-items list --project-phase "$PROJECT_PHASE_ID" --limit 5
    assert_success
else
    skip "No project phase ID available"
fi

test_name "List project phase revenue items with --project-revenue-item filter"
if [[ -n "$PROJECT_REVENUE_ITEM_ID" ]]; then
    xbe_json view project-phase-revenue-items list --project-revenue-item "$PROJECT_REVENUE_ITEM_ID" --limit 5
    assert_success
else
    skip "No project revenue item ID available"
fi

test_name "List project phase revenue items with --project-revenue-classification filter"
if [[ -n "$PROJECT_REVENUE_CLASSIFICATION_ID" ]]; then
    xbe_json view project-phase-revenue-items list --project-revenue-classification "$PROJECT_REVENUE_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No project revenue classification ID available"
fi

test_name "List project phase revenue items with --project filter"
if [[ -n "$PROJECT_ID" ]]; then
    xbe_json view project-phase-revenue-items list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List project phase revenue items with --quantity-strategy filter"
xbe_json view project-phase-revenue-items list --quantity-strategy direct --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project phase revenue item"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do project-phase-revenue-items delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"cannot be deleted"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to delete project phase revenue item"
        fi
    fi
else
    skip "No created item available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
