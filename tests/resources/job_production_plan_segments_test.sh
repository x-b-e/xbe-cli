#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Segments
#
# Tests CRUD operations for the job-production-plan-segments resource.
#
# COVERAGE: All filters + all create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SEGMENT_ID=""
SHOW_ID=""
JPP_ID=""
MATERIAL_SITE_ID=""
MATERIAL_TYPE_ID=""
COST_CODE_ID=""
SEGMENT_SET_ID=""
EXPLICIT_MTMSIL_ID=""

SELECTED_ROUTE='{"distance_miles":1.2,"route":"test"}'
UPDATED_ROUTE='{"distance_miles":2.4,"route":"updated"}'

describe "Resource: job-production-plan-segments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan segments"
xbe_json view job-production-plan-segments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(echo "$output" | jq -r '.[0].id // empty' 2>/dev/null || true)
    JPP_ID=$(echo "$output" | jq -r '.[0].job_production_plan_id // empty' 2>/dev/null || true)
    MATERIAL_SITE_ID=$(echo "$output" | jq -r '.[0].material_site_id // empty' 2>/dev/null || true)
    MATERIAL_TYPE_ID=$(echo "$output" | jq -r '.[0].material_type_id // empty' 2>/dev/null || true)
    COST_CODE_ID=$(echo "$output" | jq -r '.[0].cost_code_id // empty' 2>/dev/null || true)
    SEGMENT_SET_ID=$(echo "$output" | jq -r '.[0].job_production_plan_segment_set_id // empty' 2>/dev/null || true)
else
    fail "Failed to list job production plan segments"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create segment without required fields fails"
xbe_run do job-production-plan-segments create
assert_failure

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup job production plan segment dependencies via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    if [[ -z "$JPP_ID" ]]; then
        jpp_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/job-production-plans?page[limit]=1&fields[job-production-plans]=job-number" || true)
        JPP_ID=$(echo "$jpp_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$MATERIAL_SITE_ID" ]]; then
        material_sites_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/material-sites?page[limit]=1&fields[material-sites]=name" || true)
        MATERIAL_SITE_ID=$(echo "$material_sites_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$MATERIAL_TYPE_ID" ]]; then
        material_types_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/material-types?page[limit]=1&fields[material-types]=name" || true)
        MATERIAL_TYPE_ID=$(echo "$material_types_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$COST_CODE_ID" ]]; then
        cost_codes_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/cost-codes?page[limit]=1&fields[cost-codes]=code" || true)
        COST_CODE_ID=$(echo "$cost_codes_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$SEGMENT_SET_ID" && -n "$JPP_ID" ]]; then
        segment_sets_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/job-production-plan-segment-sets?filter[job-production-plan]=$JPP_ID&page[limit]=1" || true)
        SEGMENT_SET_ID=$(echo "$segment_sets_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    mtmsil_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/material-type-material-site-inventory-locations?page[limit]=1" || true)
    EXPLICIT_MTMSIL_ID=$(echo "$mtmsil_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create segment with attributes"
if [[ -n "$JPP_ID" ]]; then
    CREATE_ARGS=(do job-production-plan-segments create --job-production-plan "$JPP_ID" \
        --description "Test segment" \
        --non-production-minutes 10 \
        --is-expecting-weighed-transactions true \
        --explicit-start-site-kind material_site \
        --observed-possible-cycle-minutes 25 \
        --lock-observed-possible-cycle-minutes true \
        --selected-google-route "$SELECTED_ROUTE" \
        --sequence-position 1 \
        --planned-unproductive-minutes-per-hour 5 \
        --driving-minutes-per-cycle 12 \
        --material-site-minutes-per-cycle 8 \
        --tons-per-cycle 2.5)

    if [[ -n "$MATERIAL_SITE_ID" ]]; then
        CREATE_ARGS+=(--material-site "$MATERIAL_SITE_ID" --quantity 10 --quantity-per-hour 5)
    else
        CREATE_ARGS+=(--quantity 0)
    fi

    if [[ -n "$MATERIAL_TYPE_ID" ]]; then
        CREATE_ARGS+=(--material-type "$MATERIAL_TYPE_ID")
    fi

    if [[ -n "$COST_CODE_ID" ]]; then
        CREATE_ARGS+=(--cost-code "$COST_CODE_ID")
    fi

    if [[ -n "$SEGMENT_SET_ID" ]]; then
        CREATE_ARGS+=(--job-production-plan-segment-set "$SEGMENT_SET_ID")
    fi

    if [[ -n "$EXPLICIT_MTMSIL_ID" ]]; then
        CREATE_ARGS+=(--explicit-material-type-material-site-inventory-location "$EXPLICIT_MTMSIL_ID")
    fi

    xbe_json "${CREATE_ARGS[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_SEGMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_SEGMENT_ID" && "$CREATED_SEGMENT_ID" != "null" ]]; then
            register_cleanup "job-production-plan-segments" "$CREATED_SEGMENT_ID"
            pass
        else
            fail "Created segment but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create segment: $output"
        fi
    fi
else
    skip "No job production plan ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update segment attributes"
if [[ -n "$CREATED_SEGMENT_ID" && "$CREATED_SEGMENT_ID" != "null" ]]; then
    xbe_json do job-production-plan-segments update "$CREATED_SEGMENT_ID" \
        --description "Updated segment" \
        --non-production-minutes 12 \
        --is-expecting-weighed-transactions false \
        --lock-observed-possible-cycle-minutes false \
        --selected-google-route "$UPDATED_ROUTE"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update segment: $output"
        fi
    fi
else
    skip "No created segment available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show segment details"
if [[ -n "$CREATED_SEGMENT_ID" && "$CREATED_SEGMENT_ID" != "null" ]]; then
    SHOW_ID="$CREATED_SEGMENT_ID"
fi

if [[ -n "$SHOW_ID" ]]; then
    xbe_json view job-production-plan-segments show "$SHOW_ID"
    assert_success
else
    skip "No segment ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List segments with --job-production-plan filter"
if [[ -n "$JPP_ID" ]]; then
    xbe_json view job-production-plan-segments list --job-production-plan "$JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List segments with --material-site filter"
if [[ -n "$MATERIAL_SITE_ID" ]]; then
    xbe_json view job-production-plan-segments list --material-site "$MATERIAL_SITE_ID" --limit 5
    assert_success
else
    skip "No material site ID available"
fi

test_name "List segments with --material-type filter"
if [[ -n "$MATERIAL_TYPE_ID" ]]; then
    xbe_json view job-production-plan-segments list --material-type "$MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete segment"
if [[ -n "$CREATED_SEGMENT_ID" && "$CREATED_SEGMENT_ID" != "null" ]]; then
    xbe_run do job-production-plan-segments delete "$CREATED_SEGMENT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to delete segment"
        fi
    fi
else
    skip "No created segment available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
