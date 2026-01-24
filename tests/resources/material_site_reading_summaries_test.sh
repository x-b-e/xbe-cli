#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Reading Summaries
#
# Tests create operations and filters for the material_site_reading_summaries resource.
#
# COVERAGE: Create + filters + sort/limit/metrics
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

MATERIAL_SITE_ID=""
MATERIAL_SITE_MEASURE_ID=""
MATERIAL_SITE_READING_MATERIAL_TYPE_ID=""

READING_AT_MIN="2025-01-01T00:00:00Z"
READING_AT_MAX="2025-01-01T00:10:00Z"

describe "Resource: material-site-reading-summaries"

# ==========================================================================
# Seed IDs from existing data if available
# ==========================================================================

test_name "Lookup existing material site reading (if any)"
xbe_json view material-site-readings list --limit 1
if [[ $status -eq 0 ]]; then
    MATERIAL_SITE_ID=$(json_get ".[0].material_site_id")
    MATERIAL_SITE_MEASURE_ID=$(json_get ".[0].material_site_measure_id")
    MATERIAL_SITE_READING_MATERIAL_TYPE_ID=$(json_get ".[0].material_site_reading_material_type_id")
    pass
else
    fail "Failed to list material site readings"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create material site reading summary without required filters fails"
xbe_run summarize material-site-reading-summary create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create material site reading summary with required filters"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_SITE_MEASURE_ID" && "$MATERIAL_SITE_MEASURE_ID" != "null" ]]; then
    filters_json=$(cat <<FILTERS
{"material_site":"$MATERIAL_SITE_ID","material_site_measure":"$MATERIAL_SITE_MEASURE_ID","reading_at_min":"$READING_AT_MIN","reading_at_max":"$READING_AT_MAX","material_site_reading_material_type_presence":"false"}
FILTERS
)

    xbe_json summarize material-site-reading-summary create \
        --group-by minute \
        --sort value_avg:desc \
        --limit 5 \
        --metrics value_avg \
        --filters "$filters_json"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".headers"
        assert_json_has ".values"
    else
        fail "Failed to create material site reading summary"
    fi
else
    skip "Missing material site or measure IDs"
fi


test_name "Create material site reading summary with material type filter"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_SITE_MEASURE_ID" && "$MATERIAL_SITE_MEASURE_ID" != "null" && -n "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID" && "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json summarize material-site-reading-summary create \
        --group-by minute \
        --filter material_site="$MATERIAL_SITE_ID" \
        --filter material_site_measure="$MATERIAL_SITE_MEASURE_ID" \
        --filter reading_at_min="$READING_AT_MIN" \
        --filter reading_at_max="$READING_AT_MAX" \
        --filter material_site_reading_material_type="$MATERIAL_SITE_READING_MATERIAL_TYPE_ID"

    assert_success
else
    skip "Missing material site reading material type ID"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
