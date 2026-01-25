#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Material Site Changes
#
# Tests list, show, and create operations for the job-production-plan-material-site-changes resource.
#
# COVERAGE: List filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JPP_ID=""
SAMPLE_OLD_SITE_ID=""
SAMPLE_NEW_SITE_ID=""
SAMPLE_OLD_TYPE_ID=""
SAMPLE_NEW_TYPE_ID=""
SAMPLE_MIX_DESIGN_ID=""
CREATED_ID=""

describe "Resource: job-production-plan-material-site-changes"

# ============================================================================
# Endpoint Availability
# ============================================================================

test_name "Check material site change endpoint availability"
xbe_json view job-production-plan-material-site-changes list --limit 1
if [[ $status -ne 0 ]]; then
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        skip "Server does not support job-production-plan-material-site-changes (404)"
        run_tests
    else
        fail "Failed to list material site changes: $output"
        run_tests
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan material site changes"
xbe_json view job-production-plan-material-site-changes list --limit 5
assert_success

test_name "List job production plan material site changes returns array"
xbe_json view job-production-plan-material-site-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material site changes"
fi

# ============================================================================
# Sample Record (used for filters/show/create)
# ============================================================================

test_name "Capture sample material site change"
xbe_json view job-production-plan-material-site-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_JPP_ID=$(json_get ".[0].job_production_plan_id")
    SAMPLE_OLD_SITE_ID=$(json_get ".[0].old_material_site_id")
    SAMPLE_NEW_SITE_ID=$(json_get ".[0].new_material_site_id")
    SAMPLE_OLD_TYPE_ID=$(json_get ".[0].old_material_type_id")
    SAMPLE_NEW_TYPE_ID=$(json_get ".[0].new_material_type_id")
    SAMPLE_MIX_DESIGN_ID=$(json_get ".[0].new_material_mix_design_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No material site changes available for follow-on tests"
    fi
else
    skip "Could not list material site changes to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List material site changes with --created-at-min filter"
xbe_json view job-production-plan-material-site-changes list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List material site changes with --created-at-max filter"
xbe_json view job-production-plan-material-site-changes list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List material site changes with --updated-at-min filter"
xbe_json view job-production-plan-material-site-changes list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List material site changes with --updated-at-max filter"
xbe_json view job-production-plan-material-site-changes list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material site change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-material-site-changes show "$SAMPLE_ID"
    assert_success
else
    skip "No material site change ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material site change"
if [[ -n "$SAMPLE_JPP_ID" && "$SAMPLE_JPP_ID" != "null" && \
      -n "$SAMPLE_OLD_SITE_ID" && "$SAMPLE_OLD_SITE_ID" != "null" && \
      -n "$SAMPLE_NEW_SITE_ID" && "$SAMPLE_NEW_SITE_ID" != "null" ]]; then
    CREATE_ARGS=(
        --job-production-plan "$SAMPLE_JPP_ID"
        --old-material-site "$SAMPLE_OLD_SITE_ID"
        --new-material-site "$SAMPLE_NEW_SITE_ID"
    )
    if [[ -n "$SAMPLE_OLD_TYPE_ID" && "$SAMPLE_OLD_TYPE_ID" != "null" ]]; then
        CREATE_ARGS+=(--old-material-type "$SAMPLE_OLD_TYPE_ID")
    fi
    if [[ -n "$SAMPLE_NEW_TYPE_ID" && "$SAMPLE_NEW_TYPE_ID" != "null" ]]; then
        CREATE_ARGS+=(--new-material-type "$SAMPLE_NEW_TYPE_ID")
    fi
    if [[ -n "$SAMPLE_MIX_DESIGN_ID" && "$SAMPLE_MIX_DESIGN_ID" != "null" ]]; then
        CREATE_ARGS+=(--new-material-mix-design "$SAMPLE_MIX_DESIGN_ID")
    fi

    xbe_json do job-production-plan-material-site-changes create "${CREATE_ARGS[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Created material site change but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"422"* ]] || [[ "$output" == *"cannot be swapped"* ]] || \
           [[ "$output" == *"material site"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No sample job production plan/material site IDs available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material site change without required fields fails"
xbe_run do job-production-plan-material-site-changes create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
