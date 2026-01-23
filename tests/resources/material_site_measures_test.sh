#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Measures
#
# Tests list/show operations for the material_site_measures resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

MEASURE_ID=""
MEASURE_SLUG=""

describe "Resource: material-site-measures"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material site measures"
xbe_json view material-site-measures list
assert_success

test_name "List material site measures returns array"
xbe_json view material-site-measures list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material site measures"
fi

# ==========================================================================
# SHOW Tests - Get an ID
# ==========================================================================

test_name "Get a material site measure ID for show tests"
xbe_json view material-site-measures list --limit 1
if [[ $status -eq 0 ]]; then
    MEASURE_ID=$(json_get ".[0].id")
    MEASURE_SLUG=$(json_get ".[0].slug")
    if [[ -n "$MEASURE_ID" && "$MEASURE_ID" != "null" ]]; then
        pass
    else
        skip "No material site measures found in the system"
        run_tests
    fi
else
    fail "Failed to list material site measures"
    run_tests
fi

test_name "Show material site measure by ID"
xbe_json view material-site-measures show "$MEASURE_ID"
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List material site measures with --slug filter"
if [[ -n "$MEASURE_SLUG" && "$MEASURE_SLUG" != "null" ]]; then
    xbe_json view material-site-measures list --slug "$MEASURE_SLUG"
    assert_success
else
    skip "No slug available for filter test"
fi

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List material site measures with --limit"
xbe_json view material-site-measures list --limit 5
assert_success

test_name "List material site measures with --offset"
xbe_json view material-site-measures list --limit 5 --offset 1
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
