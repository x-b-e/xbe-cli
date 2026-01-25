#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Mixing Lots
#
# Tests list/show operations for the material-site-mixing-lots resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

MIXING_LOT_ID=""
MATERIAL_SITE_ID=""
MATERIAL_SUPPLIER_ID=""
BROKER_ID=""
READING_MATERIAL_TYPE_ID=""
MATERIAL_TYPE_ID=""
START_AT=""
END_AT=""
START_ON=""
TONS_PER_HOUR_AVG=""
AC_TONS_PER_HOUR_AVG=""
AGG_TONS_PER_HOUR_AVG=""
TEMPERATURE_AVG=""
AC_TEMPERATURE_AVG=""

describe "Resource: material-site-mixing-lots"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material site mixing lots"
xbe_json view material-site-mixing-lots list
assert_success

test_name "List material site mixing lots returns array"
xbe_json view material-site-mixing-lots list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material site mixing lots"
fi

# ==========================================================================
# SHOW Tests - Get an ID
# ==========================================================================

test_name "Get a mixing lot ID for show tests"
xbe_json view material-site-mixing-lots list --limit 1
if [[ $status -eq 0 ]]; then
    MIXING_LOT_ID=$(json_get ".[0].id")
    MATERIAL_SITE_ID=$(json_get ".[0].material_site_id")
    MATERIAL_SUPPLIER_ID=$(json_get ".[0].material_supplier_id")
    BROKER_ID=$(json_get ".[0].broker_id")
    READING_MATERIAL_TYPE_ID=$(json_get ".[0].material_site_reading_material_type_id")
    MATERIAL_TYPE_ID=$(json_get ".[0].material_type_id")
    START_AT=$(json_get ".[0].start_at")
    END_AT=$(json_get ".[0].end_at")
    START_ON=$(json_get ".[0].start_on")
    TONS_PER_HOUR_AVG=$(json_get ".[0].tons_per_hour_avg")
    AC_TONS_PER_HOUR_AVG=$(json_get ".[0].ac_tons_per_hour_avg")
    AGG_TONS_PER_HOUR_AVG=$(json_get ".[0].agg_tons_per_hour_avg")
    TEMPERATURE_AVG=$(json_get ".[0].temperature_avg")
    AC_TEMPERATURE_AVG=$(json_get ".[0].ac_temperature_avg")

    if [[ -n "$MIXING_LOT_ID" && "$MIXING_LOT_ID" != "null" ]]; then
        pass
    else
        skip "No mixing lots found in the system"
        run_tests
    fi
else
    fail "Failed to list mixing lots"
    run_tests
fi

test_name "Show mixing lot by ID"
xbe_json view material-site-mixing-lots show "$MIXING_LOT_ID"
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List mixing lots with --material-site filter"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --material-site "$MATERIAL_SITE_ID"
    assert_success
else
    skip "No material site ID available for filter test"
fi

test_name "List mixing lots with --material-supplier-id filter"
if [[ -n "$MATERIAL_SUPPLIER_ID" && "$MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --material-supplier-id "$MATERIAL_SUPPLIER_ID"
    assert_success
else
    skip "No material supplier ID available for filter test"
fi

test_name "List mixing lots with --material-supplier filter"
if [[ -n "$MATERIAL_SUPPLIER_ID" && "$MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --material-supplier "$MATERIAL_SUPPLIER_ID"
    assert_success
else
    skip "No material supplier ID available for filter test"
fi

test_name "List mixing lots with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --broker "$BROKER_ID"
    assert_success
else
    skip "No broker ID available for filter test"
fi

test_name "List mixing lots with --material-site-reading-material-type filter"
if [[ -n "$READING_MATERIAL_TYPE_ID" && "$READING_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --material-site-reading-material-type "$READING_MATERIAL_TYPE_ID"
    assert_success
else
    skip "No material site reading material type ID available for filter test"
fi

test_name "List mixing lots with --material-type-id filter"
if [[ -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --material-type-id "$MATERIAL_TYPE_ID"
    assert_success
else
    skip "No material type ID available for filter test"
fi

test_name "List mixing lots with --material-type filter"
if [[ -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --material-type "$MATERIAL_TYPE_ID"
    assert_success
else
    skip "No material type ID available for filter test"
fi

test_name "List mixing lots with --start-at filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --start-at "$START_AT"
    assert_success
else
    skip "No start-at value available for filter test"
fi

test_name "List mixing lots with --end-at filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --end-at "$END_AT"
    assert_success
else
    skip "No end-at value available for filter test"
fi

test_name "List mixing lots with --start-on-cached filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --start-on-cached "$START_ON"
    assert_success
else
    skip "No start-on value available for filter test"
fi

test_name "List mixing lots with --tons-per-hour-avg filter"
if [[ -n "$TONS_PER_HOUR_AVG" && "$TONS_PER_HOUR_AVG" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --tons-per-hour-avg "$TONS_PER_HOUR_AVG"
    assert_success
else
    skip "No tons per hour avg available for filter test"
fi

test_name "List mixing lots with --ac-tons-per-hour-avg filter"
if [[ -n "$AC_TONS_PER_HOUR_AVG" && "$AC_TONS_PER_HOUR_AVG" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --ac-tons-per-hour-avg "$AC_TONS_PER_HOUR_AVG"
    assert_success
else
    skip "No AC tons per hour avg available for filter test"
fi

test_name "List mixing lots with --agg-tons-per-hour-avg filter"
if [[ -n "$AGG_TONS_PER_HOUR_AVG" && "$AGG_TONS_PER_HOUR_AVG" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --agg-tons-per-hour-avg "$AGG_TONS_PER_HOUR_AVG"
    assert_success
else
    skip "No aggregate tons per hour avg available for filter test"
fi

test_name "List mixing lots with --temperature-avg filter"
if [[ -n "$TEMPERATURE_AVG" && "$TEMPERATURE_AVG" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --temperature-avg "$TEMPERATURE_AVG"
    assert_success
else
    skip "No temperature avg available for filter test"
fi

test_name "List mixing lots with --ac-temperature-avg filter"
if [[ -n "$AC_TEMPERATURE_AVG" && "$AC_TEMPERATURE_AVG" != "null" ]]; then
    xbe_json view material-site-mixing-lots list --ac-temperature-avg "$AC_TEMPERATURE_AVG"
    assert_success
else
    skip "No AC temperature avg available for filter test"
fi

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List mixing lots with --limit"
xbe_json view material-site-mixing-lots list --limit 5
assert_success

test_name "List mixing lots with --offset"
xbe_json view material-site-mixing-lots list --limit 5 --offset 1
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
