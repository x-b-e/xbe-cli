#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Readings
#
# Tests create, update, delete operations and list filters for the
# material_site_readings resource.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_READING_ID=""
EXISTING_READING_ID=""

MATERIAL_SITE_ID="${XBE_TEST_MATERIAL_SITE_ID:-}"
MATERIAL_SITE_MEASURE_ID="${XBE_TEST_MATERIAL_SITE_MEASURE_ID:-}"
MATERIAL_SITE_READING_MATERIAL_TYPE_ID="${XBE_TEST_MATERIAL_SITE_READING_MATERIAL_TYPE_ID:-}"
MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID="${XBE_TEST_MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID:-}"
RAW_MATERIAL_KIND="${XBE_TEST_MATERIAL_SITE_READING_RAW_MATERIAL_KIND:-}"
RAW_MATERIAL_DESCRIPTION="${XBE_TEST_MATERIAL_SITE_READING_RAW_MATERIAL_DESCRIPTION:-}"
RAW_MATERIAL_FEEDER_NUMBER="${XBE_TEST_MATERIAL_SITE_READING_RAW_MATERIAL_FEEDER_NUMBER:-}"

HAS_EXPLICIT_IDS=false
if [[ -n "$XBE_TEST_MATERIAL_SITE_ID" && -n "$XBE_TEST_MATERIAL_SITE_MEASURE_ID" ]]; then
    HAS_EXPLICIT_IDS=true
fi

describe "Resource: material-site-readings"

# ==========================================================================
# Seed IDs from existing data if available
# ==========================================================================

test_name "Lookup existing material site reading (if any)"
xbe_json view material-site-readings list --limit 1
if [[ $status -eq 0 ]]; then
    EXISTING_READING_ID=$(json_get ".[0].id")
    if [[ -n "$EXISTING_READING_ID" && "$EXISTING_READING_ID" != "null" ]]; then
        if [[ -z "$MATERIAL_SITE_ID" || "$MATERIAL_SITE_ID" == "null" ]]; then
            MATERIAL_SITE_ID=$(json_get ".[0].material_site_id")
        fi
        if [[ -z "$MATERIAL_SITE_MEASURE_ID" || "$MATERIAL_SITE_MEASURE_ID" == "null" ]]; then
            MATERIAL_SITE_MEASURE_ID=$(json_get ".[0].material_site_measure_id")
        fi
        if [[ -z "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID" || "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID" == "null" ]]; then
            MATERIAL_SITE_READING_MATERIAL_TYPE_ID=$(json_get ".[0].material_site_reading_material_type_id")
        fi
        if [[ -z "$MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID" || "$MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID" == "null" ]]; then
            MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID=$(json_get ".[0].material_site_reading_raw_material_type_id")
        fi
    fi
    pass
else
    fail "Failed to list material site readings"
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create material site reading"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_SITE_MEASURE_ID" && "$MATERIAL_SITE_MEASURE_ID" != "null" ]]; then
    READING_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    xbe_json do material-site-readings create \
        --material-site "$MATERIAL_SITE_ID" \
        --material-site-measure "$MATERIAL_SITE_MEASURE_ID" \
        --reading-at "$READING_AT" \
        --value "1.23"

    if [[ $status -eq 0 ]]; then
        CREATED_READING_ID=$(json_get ".id")
        if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
            register_cleanup "material-site-readings" "$CREATED_READING_ID"
            pass
        else
            fail "Created reading but no ID returned"
        fi
    else
        if [[ "$HAS_EXPLICIT_IDS" == true ]]; then
            fail "Failed to create material site reading: $output"
        else
            skip "Create failed with inferred IDs"
        fi
    fi
else
    skip "Missing material site/measure IDs (set XBE_TEST_MATERIAL_SITE_ID and XBE_TEST_MATERIAL_SITE_MEASURE_ID)"
fi


test_name "Create material site reading with raw material details"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_SITE_MEASURE_ID" && "$MATERIAL_SITE_MEASURE_ID" != "null" && -n "$RAW_MATERIAL_KIND" && -n "$RAW_MATERIAL_DESCRIPTION" && -n "$RAW_MATERIAL_FEEDER_NUMBER" ]]; then
    READING_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    xbe_json do material-site-readings create \
        --material-site "$MATERIAL_SITE_ID" \
        --material-site-measure "$MATERIAL_SITE_MEASURE_ID" \
        --reading-at "$READING_AT" \
        --value "2.75" \
        --raw-material-kind "$RAW_MATERIAL_KIND" \
        --raw-material-description "$RAW_MATERIAL_DESCRIPTION" \
        --raw-material-feeder-number "$RAW_MATERIAL_FEEDER_NUMBER"

    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "material-site-readings" "$id"
        pass
    else
        fail "Failed to create raw material site reading: $output"
    fi
else
    skip "Set XBE_TEST_MATERIAL_SITE_READING_RAW_MATERIAL_* to run"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show material site reading"
SHOW_ID="$CREATED_READING_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$EXISTING_READING_ID"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view material-site-readings show "$SHOW_ID"
    assert_success
else
    skip "No material site reading ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update material site reading material type relationship"
UPDATE_ID="$CREATED_READING_ID"
if [[ -z "$UPDATE_ID" || "$UPDATE_ID" == "null" ]]; then
    UPDATE_ID="${XBE_TEST_MATERIAL_SITE_READING_ID:-}"
fi
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID" && "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json do material-site-readings update "$UPDATE_ID" --material-site-reading-material-type "$MATERIAL_SITE_READING_MATERIAL_TYPE_ID"
    assert_success
else
    skip "Set XBE_TEST_MATERIAL_SITE_READING_ID and XBE_TEST_MATERIAL_SITE_READING_MATERIAL_TYPE_ID to run"
fi


test_name "Update material site reading raw material type relationship"
UPDATE_ID="$CREATED_READING_ID"
if [[ -z "$UPDATE_ID" || "$UPDATE_ID" == "null" ]]; then
    UPDATE_ID="${XBE_TEST_MATERIAL_SITE_READING_ID:-}"
fi
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID" && "$MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json do material-site-readings update "$UPDATE_ID" --material-site-reading-raw-material-type "$MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID"
    assert_success
else
    skip "Set XBE_TEST_MATERIAL_SITE_READING_ID and XBE_TEST_MATERIAL_SITE_READING_RAW_MATERIAL_TYPE_ID to run"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material site readings"
xbe_json view material-site-readings list --limit 10
assert_success

test_name "List material site readings returns array"
xbe_json view material-site-readings list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material site readings"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List material site readings with --material-site filter"
xbe_json view material-site-readings list --material-site 1 --limit 10
assert_success

test_name "List material site readings with --material-site-measure filter"
xbe_json view material-site-readings list --material-site-measure 1 --limit 10
assert_success

test_name "List material site readings with --material-site-reading-material-type filter"
xbe_json view material-site-readings list --material-site-reading-material-type 1 --limit 10
assert_success

test_name "List material site readings with --material-type filter"
xbe_json view material-site-readings list --material-type 1 --limit 10
assert_success

test_name "List material site readings with --material-type-id filter"
xbe_json view material-site-readings list --material-type-id 1 --limit 10
assert_success

test_name "List material site readings with --broker filter"
xbe_json view material-site-readings list --broker 1 --limit 10
assert_success

test_name "List material site readings with --reading-at filter"
xbe_json view material-site-readings list --reading-at 2025-01-01T00:00:00Z --limit 10
assert_success

test_name "List material site readings with --reading-at-min filter"
xbe_json view material-site-readings list --reading-at-min 2025-01-01T00:00:00Z --limit 10
assert_success

test_name "List material site readings with --reading-at-max filter"
xbe_json view material-site-readings list --reading-at-max 2025-01-31T23:59:59Z --limit 10
assert_success

test_name "List material site readings with --value filter"
xbe_json view material-site-readings list --value 1 --limit 10
assert_success

test_name "List material site readings with --value-min filter"
xbe_json view material-site-readings list --value-min 1 --limit 10
assert_success

test_name "List material site readings with --value-max filter"
xbe_json view material-site-readings list --value-max 100 --limit 10
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete material site reading requires --confirm flag"
if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
    xbe_json do material-site-readings delete "$CREATED_READING_ID"
    assert_failure
else
    skip "No reading ID available"
fi


test_name "Delete material site reading with --confirm"
if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
    xbe_json do material-site-readings delete "$CREATED_READING_ID" --confirm
    assert_success
else
    skip "No reading ID available"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
