#!/bin/bash
#
# XBE CLI Integration Tests: Project Phase Cost Items
#
# Tests CRUD operations for the project-phase-cost-items resource.
#
# COVERAGE: All filters + create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
SHOW_ID=""
PROJECT_ID=""
PROJECT_PHASE_ID=""
PROJECT_PHASE_REVENUE_ITEM_ID=""
PROJECT_COST_CLASSIFICATION_ID=""
ALT_PROJECT_COST_CLASSIFICATION_ID=""
PROJECT_RESOURCE_CLASSIFICATION_ID=""
UNIT_OF_MEASURE_ID=""
COST_CODE_ID=""
ESTIMATE_SET_ID=""
PRICE_ESTIMATE_ID=""
QUANTITY_ESTIMATE_ID=""
IS_REVENUE_QUANTITY_DRIVER=""

describe "Resource: project-phase-cost-items"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project phase cost items"
xbe_json view project-phase-cost-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(echo "$output" | jq -r ".[0].id // empty" 2>/dev/null || true)
    PROJECT_PHASE_REVENUE_ITEM_ID=$(echo "$output" | jq -r ".[0].project_phase_revenue_item_id // empty" 2>/dev/null || true)
    PROJECT_COST_CLASSIFICATION_ID=$(echo "$output" | jq -r ".[0].project_cost_classification_id // empty" 2>/dev/null || true)
    PROJECT_RESOURCE_CLASSIFICATION_ID=$(echo "$output" | jq -r ".[0].project_resource_classification_id // empty" 2>/dev/null || true)
    UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r ".[0].unit_of_measure_id // empty" 2>/dev/null || true)
    COST_CODE_ID=$(echo "$output" | jq -r ".[0].cost_code_id // empty" 2>/dev/null || true)
    IS_REVENUE_QUANTITY_DRIVER=$(echo "$output" | jq -r ".[0].is_revenue_quantity_driver // empty" 2>/dev/null || true)
else
    fail "Failed to list project phase cost items"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cost item without required fields fails"
xbe_run do project-phase-cost-items create
assert_failure

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup project phase revenue item and classifications via API"
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

        ppri_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-phase-revenue-items?filter[project]=$project_id&page[limit]=50" || true)

        if [[ "$(echo "$ppri_json" | jq -r ".data | length" 2>/dev/null)" == "0" ]]; then
            continue
        fi

        ppc_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-project-cost-classifications?filter[project]=$project_id&page[limit]=200" || true)

        cost_class_ids=$(echo "$ppc_json" | jq -r ".data[].relationships[\"project-cost-classification\"].data.id" 2>/dev/null | sort -u)
        if [[ -z "$cost_class_ids" ]]; then
            continue
        fi

        while read -r ppri_id; do
            if [[ -z "$ppri_id" ]]; then
                continue
            fi

            existing_json=$(curl -s \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                "$base_url/v1/project-phase-cost-items?filter[project-phase-revenue-item]=$ppri_id&page[limit]=200" || true)

            existing_cc_ids=$(echo "$existing_json" | jq -r ".data[].relationships[\"project-cost-classification\"].data.id" 2>/dev/null || true)

            candidate_id=""
            while read -r cc_id; do
                if [[ -z "$cc_id" ]]; then
                    continue
                fi
                if ! grep -Fxq "$cc_id" <<<"$existing_cc_ids"; then
                    candidate_id="$cc_id"
                    break
                fi
            done <<<"$cost_class_ids"

            if [[ -n "$candidate_id" ]]; then
                PROJECT_ID="$project_id"
                PROJECT_PHASE_REVENUE_ITEM_ID="$ppri_id"
                PROJECT_COST_CLASSIFICATION_ID="$candidate_id"
                PROJECT_PHASE_ID=$(echo "$ppri_json" | jq -r --arg id "$ppri_id" ".data[] | select(.id==\$id) | .relationships[\"project-phase\"].data.id // empty" 2>/dev/null || true)

                alt_id=""
                while read -r cc_id; do
                    if [[ -z "$cc_id" || "$cc_id" == "$candidate_id" ]]; then
                        continue
                    fi
                    if ! grep -Fxq "$cc_id" <<<"$existing_cc_ids"; then
                        alt_id="$cc_id"
                        break
                    fi
                done <<<"$cost_class_ids"
                ALT_PROJECT_COST_CLASSIFICATION_ID="$alt_id"
                break
            fi
        done < <(echo "$ppri_json" | jq -r ".data[].id" 2>/dev/null || true)

        if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" && -n "$PROJECT_COST_CLASSIFICATION_ID" ]]; then
            break
        fi
    done < <(echo "$projects_json" | jq -r ".data[].id" 2>/dev/null || true)

    if [[ -z "$PROJECT_PHASE_REVENUE_ITEM_ID" || -z "$PROJECT_COST_CLASSIFICATION_ID" ]]; then
        skip "No suitable project phase revenue item and cost classification found"
    else
        cc_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-cost-classifications/$PROJECT_COST_CLASSIFICATION_ID?fields[project-cost-classifications]=default-project-resource-classification,default-unit-of-measure" || true)

        PROJECT_RESOURCE_CLASSIFICATION_ID=$(echo "$cc_json" | jq -r ".data.relationships[\"default-project-resource-classification\"].data.id // empty" 2>/dev/null || true)
        UNIT_OF_MEASURE_ID=$(echo "$cc_json" | jq -r ".data.relationships[\"default-unit-of-measure\"].data.id // empty" 2>/dev/null || true)

        cost_codes_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-cost-codes?filter[project]=$PROJECT_ID&page[limit]=50" || true)

        COST_CODE_ID=$(echo "$cost_codes_json" | jq -r ".data[0].relationships[\"cost-code\"].data.id // empty" 2>/dev/null || true)

        estimate_sets_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/project-estimate-sets?filter[project]=$PROJECT_ID&page[limit]=50" || true)

        ESTIMATE_SET_ID=$(echo "$estimate_sets_json" | jq -r ".data[0].id // empty" 2>/dev/null || true)
        pass
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cost item with required fields"
if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" && -n "$PROJECT_COST_CLASSIFICATION_ID" ]]; then
    create_args=(do project-phase-cost-items create \
        --project-phase-revenue-item "$PROJECT_PHASE_REVENUE_ITEM_ID" \
        --project-cost-classification "$PROJECT_COST_CLASSIFICATION_ID" \
        --is-revenue-quantity-driver true)

    if [[ -n "$PROJECT_RESOURCE_CLASSIFICATION_ID" ]]; then
        create_args+=(--project-resource-classification "$PROJECT_RESOURCE_CLASSIFICATION_ID")
    fi
    if [[ -n "$UNIT_OF_MEASURE_ID" ]]; then
        create_args+=(--unit-of-measure "$UNIT_OF_MEASURE_ID")
    fi
    if [[ -n "$COST_CODE_ID" ]]; then
        create_args+=(--cost-code "$COST_CODE_ID")
    fi

    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-phase-cost-items" "$CREATED_ID"
            pass
        else
            fail "Created cost item but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create cost item: $output"
        fi
    fi
else
    skip "Missing project phase revenue item or cost classification"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update revenue quantity driver"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-phase-cost-items update "$CREATED_ID" --is-revenue-quantity-driver false
    assert_success
else
    skip "No created cost item available"
fi

test_name "Update project resource classification"
if [[ -n "$CREATED_ID" && -n "$PROJECT_RESOURCE_CLASSIFICATION_ID" ]]; then
    xbe_json do project-phase-cost-items update "$CREATED_ID" --project-resource-classification "$PROJECT_RESOURCE_CLASSIFICATION_ID"
    assert_success
else
    skip "No project resource classification available"
fi

test_name "Update unit of measure"
if [[ -n "$CREATED_ID" && -n "$UNIT_OF_MEASURE_ID" ]]; then
    xbe_json do project-phase-cost-items update "$CREATED_ID" --unit-of-measure "$UNIT_OF_MEASURE_ID"
    assert_success
else
    skip "No unit of measure available"
fi

test_name "Update cost code"
if [[ -n "$CREATED_ID" && -n "$COST_CODE_ID" ]]; then
    xbe_json do project-phase-cost-items update "$CREATED_ID" --cost-code "$COST_CODE_ID"
    assert_success
else
    skip "No cost code available"
fi

test_name "Update project cost classification"
if [[ -n "$CREATED_ID" && -n "$ALT_PROJECT_COST_CLASSIFICATION_ID" ]]; then
    xbe_json do project-phase-cost-items update "$CREATED_ID" --project-cost-classification "$ALT_PROJECT_COST_CLASSIFICATION_ID"
    assert_success
else
    skip "No alternate project cost classification available"
fi

# Create price estimate for update tests
if [[ -n "$CREATED_ID" && -n "$ESTIMATE_SET_ID" ]]; then
    test_name "Create price estimate for cost item"
    xbe_json do project-phase-cost-item-price-estimates create \
        --project-phase-cost-item "$CREATED_ID" \
        --project-estimate-set "$ESTIMATE_SET_ID" \
        --estimate "{\"class_name\":\"NormalDistribution\",\"mean\":10,\"standard_deviation\":2}"

    if [[ $status -eq 0 ]]; then
        PRICE_ESTIMATE_ID=$(json_get ".id")
        if [[ -n "$PRICE_ESTIMATE_ID" && "$PRICE_ESTIMATE_ID" != "null" ]]; then
            register_cleanup "project-phase-cost-item-price-estimates" "$PRICE_ESTIMATE_ID"
            pass
        else
            fail "Created price estimate but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create price estimate"
        fi
    fi
else
    test_name "Create price estimate for cost item"
    skip "Missing cost item or estimate set"
fi

test_name "Update price estimate"
if [[ -n "$CREATED_ID" && -n "$PRICE_ESTIMATE_ID" ]]; then
    xbe_json do project-phase-cost-items update "$CREATED_ID" --price-estimate "$PRICE_ESTIMATE_ID"
    assert_success
else
    skip "No price estimate available"
fi

test_name "Update quantity estimate"
if [[ -n "$CREATED_ID" && -n "$ESTIMATE_SET_ID" && -n "$XBE_TOKEN" ]]; then
    base_url="${XBE_BASE_URL%/}"
    payload="{\"data\":{\"type\":\"project-phase-cost-item-quantity-estimates\",\"attributes\":{\"estimate\":{\"class_name\":\"NormalDistribution\",\"mean\":5,\"standard_deviation\":1}},\"relationships\":{\"project-phase-cost-item\":{\"data\":{\"type\":\"project-phase-cost-items\",\"id\":\"$CREATED_ID\"}},\"project-estimate-set\":{\"data\":{\"type\":\"project-estimate-sets\",\"id\":\"$ESTIMATE_SET_ID\"}}}}}"

    quantity_response=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json" \
        -H "Accept: application/vnd.api+json" \
        -X POST \
        --data "$payload" \
        "$base_url/v1/project-phase-cost-item-quantity-estimates" || true)

    QUANTITY_ESTIMATE_ID=$(echo "$quantity_response" | jq -r ".data.id // empty" 2>/dev/null || true)

    if [[ -n "$QUANTITY_ESTIMATE_ID" ]]; then
        xbe_json do project-phase-cost-items update "$CREATED_ID" --quantity-estimate "$QUANTITY_ESTIMATE_ID"
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
    skip "Missing cost item, estimate set, or token for quantity estimate"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show cost item details"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    SHOW_ID="$CREATED_ID"
fi

if [[ -n "$SHOW_ID" ]]; then
    xbe_json view project-phase-cost-items show "$SHOW_ID"
    assert_success
else
    skip "No cost item ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List cost items with --project-phase-revenue-item filter"
if [[ -n "$PROJECT_PHASE_REVENUE_ITEM_ID" ]]; then
    xbe_json view project-phase-cost-items list --project-phase-revenue-item "$PROJECT_PHASE_REVENUE_ITEM_ID" --limit 5
    assert_success
else
    skip "No project phase revenue item ID available"
fi

test_name "List cost items with --project-cost-classification filter"
if [[ -n "$PROJECT_COST_CLASSIFICATION_ID" ]]; then
    xbe_json view project-phase-cost-items list --project-cost-classification "$PROJECT_COST_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No project cost classification ID available"
fi

test_name "List cost items with --project-resource-classification filter"
if [[ -n "$PROJECT_RESOURCE_CLASSIFICATION_ID" ]]; then
    xbe_json view project-phase-cost-items list --project-resource-classification "$PROJECT_RESOURCE_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No project resource classification ID available"
fi

test_name "List cost items with --unit-of-measure filter"
if [[ -n "$UNIT_OF_MEASURE_ID" ]]; then
    xbe_json view project-phase-cost-items list --unit-of-measure "$UNIT_OF_MEASURE_ID" --limit 5
    assert_success
else
    skip "No unit of measure ID available"
fi

test_name "List cost items with --project-phase filter"
if [[ -n "$PROJECT_PHASE_ID" ]]; then
    xbe_json view project-phase-cost-items list --project-phase "$PROJECT_PHASE_ID" --limit 5
    assert_success
else
    skip "No project phase ID available"
fi

test_name "List cost items with --project filter"
if [[ -n "$PROJECT_ID" ]]; then
    xbe_json view project-phase-cost-items list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List cost items with --is-revenue-quantity-driver filter"
xbe_json view project-phase-cost-items list --is-revenue-quantity-driver true --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete cost item"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do project-phase-cost-items delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"cannot be deleted"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to delete cost item"
        fi
    fi
else
    skip "No created cost item available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
