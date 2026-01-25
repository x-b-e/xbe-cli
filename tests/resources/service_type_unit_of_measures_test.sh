#!/bin/bash
#
# XBE CLI Integration Tests: Service Type Unit of Measures
#
# Tests view operations for the service_type_unit_of_measures resource.
# Service type unit of measures define how services are quantified and billed.
#
# COVERAGE: List + show + filters (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: service_type_unit_of_measures (view-only)"

SAMPLE_ID=""
SAMPLE_SERVICE_TYPE_ID=""
SAMPLE_UNIT_OF_MEASURE_ID=""
SAMPLE_QUANTIFIABLE=""

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List service type unit of measures"
xbe_json view service-type-unit-of-measures list --limit 5
assert_success

test_name "List service type unit of measures returns array"
xbe_json view service-type-unit-of-measures list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list service type unit of measures"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample service type unit of measure"
xbe_json view service-type-unit-of-measures list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_SERVICE_TYPE_ID=$(json_get ".[0].service_type_id")
    SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".[0].unit_of_measure_id")
    SAMPLE_QUANTIFIABLE=$(json_get ".[0].is_quantifiable")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No service type unit of measures available for follow-on tests"
    fi
else
    skip "Could not list service type unit of measures to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List service type unit of measures with --service-type filter"
if [[ -n "$SAMPLE_SERVICE_TYPE_ID" && "$SAMPLE_SERVICE_TYPE_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measures list --service-type "$SAMPLE_SERVICE_TYPE_ID" --limit 5
    assert_success
else
    skip "No sample service type ID available"
fi

test_name "List service type unit of measures with --unit-of-measure filter"
if [[ -n "$SAMPLE_UNIT_OF_MEASURE_ID" && "$SAMPLE_UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measures list --unit-of-measure "$SAMPLE_UNIT_OF_MEASURE_ID" --limit 5
    assert_success
else
    skip "No sample unit of measure ID available"
fi

test_name "List service type unit of measures with --quantifiable filter"
if [[ -n "$SAMPLE_QUANTIFIABLE" && "$SAMPLE_QUANTIFIABLE" != "null" ]]; then
    xbe_json view service-type-unit-of-measures list --quantifiable "$SAMPLE_QUANTIFIABLE" --limit 5
    assert_success
else
    skip "No sample quantifiable value available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show service type unit of measure"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measures show "$SAMPLE_ID"
    assert_success
else
    skip "No service type unit of measure ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
