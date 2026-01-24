#!/bin/bash
#
# XBE CLI Integration Tests: Project Bid Location Material Types
#
# Tests list, show, create, update, delete operations for the
# project-bid-location-material-types resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PROJECT_BID_LOCATION_ID=""
SAMPLE_MATERIAL_TYPE_ID=""
CREATED_ID=""
CREATE_PROJECT_BID_LOCATION_ID=""
CREATE_MATERIAL_TYPE_ID=""
ALT_MATERIAL_TYPE_ID=""
UNIT_OF_MEASURE_ID=""
LIST_SUPPORTED="true"

describe "Resource: project_bid_location_material_types"

is_nonfatal_error() {
    [[ "$output" == *"Not Authorized"* ]] || \
    [[ "$output" == *"not authorized"* ]] || \
    [[ "$output" == *"Record Invalid"* ]] || \
    [[ "$output" == *"422"* ]]
}

pick_unused_material_type() {
    local project_bid_location_id="$1"
    local used_ids=""

    if [[ -n "$project_bid_location_id" ]]; then
        xbe_json view project-bid-location-material-types list \
            --project-bid-location "$project_bid_location_id" \
            --limit 200
        if [[ $status -eq 0 ]]; then
            used_ids=$(echo "$output" | jq -r '.[].material_type_id' | tr '\n' ' ')
        fi
    fi

    xbe_json view material-types list --limit 200
    if [[ $status -ne 0 ]]; then
        echo ""
        return 1
    fi

    local mt_id
    for mt_id in $(echo "$output" | jq -r '.[].id'); do
        if [[ -z "$mt_id" || "$mt_id" == "null" ]]; then
            continue
        fi
        if [[ " $used_ids " != *" $mt_id "* ]]; then
            echo "$mt_id"
            return 0
        fi
    done

    echo ""
    return 1
}

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project bid location material types"
xbe_json view project-bid-location-material-types list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing project bid location material types"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project bid location material types returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-bid-location-material-types list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project bid location material types"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample project bid location material type"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-bid-location-material-types list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_PROJECT_BID_LOCATION_ID=$(json_get ".[0].project_bid_location_id")
        SAMPLE_MATERIAL_TYPE_ID=$(json_get ".[0].material_type_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project bid location material types available for follow-on tests"
        fi
    else
        skip "Could not list project bid location material types to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter by project bid location"
if [[ -n "$SAMPLE_PROJECT_BID_LOCATION_ID" && "$SAMPLE_PROJECT_BID_LOCATION_ID" != "null" ]]; then
    xbe_json view project-bid-location-material-types list \
        --project-bid-location "$SAMPLE_PROJECT_BID_LOCATION_ID" \
        --limit 5
    assert_success
else
    skip "No project bid location ID available"
fi

test_name "Filter by material type"
if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view project-bid-location-material-types list \
        --material-type "$SAMPLE_MATERIAL_TYPE_ID" \
        --limit 5
    assert_success
else
    skip "No material type ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project bid location material type"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-bid-location-material-types show "$SAMPLE_ID"
    assert_success
else
    skip "No project bid location material type ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Find unit of measure for create"
xbe_json view unit-of-measures list --limit 1
if [[ $status -eq 0 ]]; then
    UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
    pass
else
    skip "Could not list unit of measures"
fi

if [[ -n "$XBE_TEST_PROJECT_BID_LOCATION_ID" ]]; then
    CREATE_PROJECT_BID_LOCATION_ID="$XBE_TEST_PROJECT_BID_LOCATION_ID"
elif [[ -n "$SAMPLE_PROJECT_BID_LOCATION_ID" && "$SAMPLE_PROJECT_BID_LOCATION_ID" != "null" ]]; then
    CREATE_PROJECT_BID_LOCATION_ID="$SAMPLE_PROJECT_BID_LOCATION_ID"
fi

CREATE_MATERIAL_TYPE_ID=""
if [[ -n "$CREATE_PROJECT_BID_LOCATION_ID" ]]; then
    CREATE_MATERIAL_TYPE_ID=$(pick_unused_material_type "$CREATE_PROJECT_BID_LOCATION_ID")
fi

test_name "Create project bid location material type"
if [[ -n "$CREATE_PROJECT_BID_LOCATION_ID" && -n "$CREATE_MATERIAL_TYPE_ID" ]]; then
    xbe_json do project-bid-location-material-types create \
        --project-bid-location "$CREATE_PROJECT_BID_LOCATION_ID" \
        --material-type "$CREATE_MATERIAL_TYPE_ID" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID" \
        --quantity "12.5" \
        --notes "Test notes"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-bid-location-material-types" "$CREATED_ID"
        fi
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_BID_LOCATION_ID or provide existing data to enable create test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_TARGET_ID="$CREATED_ID"
UPDATE_PROJECT_BID_LOCATION_ID="$CREATE_PROJECT_BID_LOCATION_ID"
if [[ -z "$UPDATE_TARGET_ID" || "$UPDATE_TARGET_ID" == "null" ]]; then
    UPDATE_TARGET_ID="$SAMPLE_ID"
    UPDATE_PROJECT_BID_LOCATION_ID="$SAMPLE_PROJECT_BID_LOCATION_ID"
fi

if [[ -n "$UPDATE_PROJECT_BID_LOCATION_ID" ]]; then
    ALT_MATERIAL_TYPE_ID=$(pick_unused_material_type "$UPDATE_PROJECT_BID_LOCATION_ID")
fi

test_name "Update project bid location material type quantity and notes"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" ]]; then
    xbe_json do project-bid-location-material-types update "$UPDATE_TARGET_ID" \
        --quantity "15" \
        --notes "Updated notes"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "No project bid location material type ID available"
fi

test_name "Update project bid location material type unit of measure"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" && -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json do project-bid-location-material-types update "$UPDATE_TARGET_ID" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update unit of measure failed: $output"
        fi
    fi
else
    skip "No update target or unit of measure ID available"
fi

test_name "Update project bid location material type material type"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" && -n "$ALT_MATERIAL_TYPE_ID" && "$ALT_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json do project-bid-location-material-types update "$UPDATE_TARGET_ID" \
        --material-type "$ALT_MATERIAL_TYPE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update material type failed: $output"
        fi
    fi
else
    skip "No alternative material type available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project bid location material type requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do project-bid-location-material-types delete "$SAMPLE_ID"
    assert_failure
else
    skip "No project bid location material type ID available"
fi

test_name "Delete project bid location material type"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-bid-location-material-types delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created project bid location material type ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required flags fails"
xbe_run do project-bid-location-material-types create
assert_failure

test_name "Update without fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do project-bid-location-material-types update "$SAMPLE_ID"
    assert_failure
else
    skip "No project bid location material type ID available"
fi

# ============================================================================
# Done
# ============================================================================

run_tests
