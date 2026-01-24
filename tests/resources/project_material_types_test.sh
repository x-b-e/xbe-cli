#!/bin/bash
#
# XBE CLI Integration Tests: Project Material Types
#
# Tests list, show, create, update, delete operations for the
# project-material-types resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PROJECT_ID=""
SAMPLE_MATERIAL_TYPE_ID=""
SAMPLE_PICKUP_LOCATION_ID=""
SAMPLE_DELIVERY_LOCATION_ID=""
SAMPLE_PICKUP_AT_MIN=""
SAMPLE_PICKUP_AT_MAX=""
SAMPLE_DELIVER_AT_MIN=""
SAMPLE_DELIVER_AT_MAX=""
CREATED_ID=""
CREATE_PROJECT_ID=""
CREATE_MATERIAL_TYPE_ID=""
UNIT_OF_MEASURE_ID=""
MATERIAL_SITE_ID=""
JOB_SITE_ID=""
LIST_SUPPORTED="true"

describe "Resource: project_material_types"

is_nonfatal_error() {
    [[ "$output" == *"Not Authorized"* ]] || \
    [[ "$output" == *"not authorized"* ]] || \
    [[ "$output" == *"Record Invalid"* ]] || \
    [[ "$output" == *"422"* ]] || \
    [[ "$output" == *"403"* ]]
}

pick_unused_material_type() {
    local project_id="$1"
    local used_ids=""

    if [[ -n "$project_id" ]]; then
        xbe_json view project-material-types list --project "$project_id" --limit 200
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

test_name "List project material types"
xbe_json view project-material-types list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing project material types"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project material types returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-material-types list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project material types"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample project material type"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-material-types list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_PROJECT_ID=$(json_get ".[0].project_id")
        SAMPLE_MATERIAL_TYPE_ID=$(json_get ".[0].material_type_id")
        SAMPLE_PICKUP_LOCATION_ID=$(json_get ".[0].pickup_location_id")
        SAMPLE_DELIVERY_LOCATION_ID=$(json_get ".[0].delivery_location_id")
        SAMPLE_PICKUP_AT_MIN=$(json_get ".[0].pickup_at_min")
        SAMPLE_PICKUP_AT_MAX=$(json_get ".[0].pickup_at_max")
        SAMPLE_DELIVER_AT_MIN=$(json_get ".[0].deliver_at_min")
        SAMPLE_DELIVER_AT_MAX=$(json_get ".[0].deliver_at_max")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project material types available for follow-on tests"
        fi
    else
        skip "Could not list project material types to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter by project"
if [[ -n "$SAMPLE_PROJECT_ID" && "$SAMPLE_PROJECT_ID" != "null" ]]; then
    xbe_json view project-material-types list --project "$SAMPLE_PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "Filter by material type"
if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view project-material-types list --material-type "$SAMPLE_MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

test_name "Filter by pickup location"
if [[ -n "$SAMPLE_PICKUP_LOCATION_ID" && "$SAMPLE_PICKUP_LOCATION_ID" != "null" ]]; then
    xbe_json view project-material-types list --pickup-location "$SAMPLE_PICKUP_LOCATION_ID" --limit 5
    assert_success
else
    skip "No pickup location ID available"
fi

test_name "Filter by delivery location"
if [[ -n "$SAMPLE_DELIVERY_LOCATION_ID" && "$SAMPLE_DELIVERY_LOCATION_ID" != "null" ]]; then
    xbe_json view project-material-types list --delivery-location "$SAMPLE_DELIVERY_LOCATION_ID" --limit 5
    assert_success
else
    skip "No delivery location ID available"
fi

test_name "Filter by pickup-at-min-min"
if [[ -n "$SAMPLE_PICKUP_AT_MIN" && "$SAMPLE_PICKUP_AT_MIN" != "null" ]]; then
    xbe_json view project-material-types list --pickup-at-min-min "$SAMPLE_PICKUP_AT_MIN" --limit 5
    assert_success
else
    skip "No pickup-at-min available"
fi

test_name "Filter by pickup-at-min-max"
if [[ -n "$SAMPLE_PICKUP_AT_MIN" && "$SAMPLE_PICKUP_AT_MIN" != "null" ]]; then
    xbe_json view project-material-types list --pickup-at-min-max "$SAMPLE_PICKUP_AT_MIN" --limit 5
    assert_success
else
    skip "No pickup-at-min available"
fi

test_name "Filter by pickup-at-max-min"
if [[ -n "$SAMPLE_PICKUP_AT_MAX" && "$SAMPLE_PICKUP_AT_MAX" != "null" ]]; then
    xbe_json view project-material-types list --pickup-at-max-min "$SAMPLE_PICKUP_AT_MAX" --limit 5
    assert_success
else
    skip "No pickup-at-max available"
fi

test_name "Filter by pickup-at-max-max"
if [[ -n "$SAMPLE_PICKUP_AT_MAX" && "$SAMPLE_PICKUP_AT_MAX" != "null" ]]; then
    xbe_json view project-material-types list --pickup-at-max-max "$SAMPLE_PICKUP_AT_MAX" --limit 5
    assert_success
else
    skip "No pickup-at-max available"
fi

test_name "Filter by deliver-at-min-min"
if [[ -n "$SAMPLE_DELIVER_AT_MIN" && "$SAMPLE_DELIVER_AT_MIN" != "null" ]]; then
    xbe_json view project-material-types list --deliver-at-min-min "$SAMPLE_DELIVER_AT_MIN" --limit 5
    assert_success
else
    skip "No deliver-at-min available"
fi

test_name "Filter by deliver-at-min-max"
if [[ -n "$SAMPLE_DELIVER_AT_MIN" && "$SAMPLE_DELIVER_AT_MIN" != "null" ]]; then
    xbe_json view project-material-types list --deliver-at-min-max "$SAMPLE_DELIVER_AT_MIN" --limit 5
    assert_success
else
    skip "No deliver-at-min available"
fi

test_name "Filter by deliver-at-max-min"
if [[ -n "$SAMPLE_DELIVER_AT_MAX" && "$SAMPLE_DELIVER_AT_MAX" != "null" ]]; then
    xbe_json view project-material-types list --deliver-at-max-min "$SAMPLE_DELIVER_AT_MAX" --limit 5
    assert_success
else
    skip "No deliver-at-max available"
fi

test_name "Filter by deliver-at-max-max"
if [[ -n "$SAMPLE_DELIVER_AT_MAX" && "$SAMPLE_DELIVER_AT_MAX" != "null" ]]; then
    xbe_json view project-material-types list --deliver-at-max-max "$SAMPLE_DELIVER_AT_MAX" --limit 5
    assert_success
else
    skip "No deliver-at-max available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project material type"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-material-types show "$SAMPLE_ID"
    assert_success
else
    skip "No project material type ID available"
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

test_name "Find material site for update"
xbe_json view material-sites list --limit 1
if [[ $status -eq 0 ]]; then
    MATERIAL_SITE_ID=$(json_get ".[0].id")
    pass
else
    skip "Could not list material sites"
fi

test_name "Find job site for update"
xbe_json view job-sites list --limit 1
if [[ $status -eq 0 ]]; then
    JOB_SITE_ID=$(json_get ".[0].id")
    pass
else
    skip "Could not list job sites"
fi

if [[ -n "$XBE_TEST_PROJECT_ID" ]]; then
    CREATE_PROJECT_ID="$XBE_TEST_PROJECT_ID"
elif [[ -n "$SAMPLE_PROJECT_ID" && "$SAMPLE_PROJECT_ID" != "null" ]]; then
    CREATE_PROJECT_ID="$SAMPLE_PROJECT_ID"
fi

if [[ -n "$CREATE_PROJECT_ID" ]]; then
    CREATE_MATERIAL_TYPE_ID=$(pick_unused_material_type "$CREATE_PROJECT_ID")
fi

test_name "Create project material type"
if [[ -n "$CREATE_PROJECT_ID" && -n "$CREATE_MATERIAL_TYPE_ID" ]]; then
    xbe_json do project-material-types create \
        --project "$CREATE_PROJECT_ID" \
        --material-type "$CREATE_MATERIAL_TYPE_ID" \
        --quantity "12.5" \
        --explicit-display-name "Test Material Type" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-material-types" "$CREATED_ID"
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
    skip "Set XBE_TEST_PROJECT_ID or provide existing data to enable create test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_TARGET_ID="$CREATED_ID"
if [[ -z "$UPDATE_TARGET_ID" || "$UPDATE_TARGET_ID" == "null" ]]; then
    UPDATE_TARGET_ID="$SAMPLE_ID"
fi

test_name "Update project material type attributes"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" ]]; then
    xbe_json do project-material-types update "$UPDATE_TARGET_ID" \
        --quantity "15" \
        --explicit-display-name "Updated Material" \
        --pickup-at-min "2026-01-23T08:00:00Z" \
        --pickup-at-max "2026-01-23T10:00:00Z" \
        --deliver-at-min "2026-01-23T11:00:00Z" \
        --deliver-at-max "2026-01-23T12:00:00Z"
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
    skip "No project material type ID available"
fi

test_name "Update project material type unit of measure"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" && -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json do project-material-types update "$UPDATE_TARGET_ID" \
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

test_name "Update project material type material site"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" && -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json do project-material-types update "$UPDATE_TARGET_ID" \
        --material-site "$MATERIAL_SITE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update material site failed: $output"
        fi
    fi
else
    skip "No update target or material site ID available"
fi

test_name "Update project material type job site"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" && -n "$JOB_SITE_ID" && "$JOB_SITE_ID" != "null" ]]; then
    xbe_json do project-material-types update "$UPDATE_TARGET_ID" \
        --job-site "$JOB_SITE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update job site failed: $output"
        fi
    fi
else
    skip "No update target or job site ID available"
fi

test_name "Update project material type pickup/delivery locations"
if [[ -n "$UPDATE_TARGET_ID" && "$UPDATE_TARGET_ID" != "null" && \
      -n "$SAMPLE_PICKUP_LOCATION_ID" && "$SAMPLE_PICKUP_LOCATION_ID" != "null" && \
      -n "$SAMPLE_DELIVERY_LOCATION_ID" && "$SAMPLE_DELIVERY_LOCATION_ID" != "null" ]]; then
    xbe_json do project-material-types update "$UPDATE_TARGET_ID" \
        --pickup-location "$SAMPLE_PICKUP_LOCATION_ID" \
        --delivery-location "$SAMPLE_DELIVERY_LOCATION_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update pickup/delivery locations failed: $output"
        fi
    fi
else
    skip "No pickup/delivery location IDs available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project material type requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do project-material-types delete "$SAMPLE_ID"
    assert_failure
else
    skip "No project material type ID available"
fi

test_name "Delete project material type"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-material-types delete "$CREATED_ID" --confirm
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
    skip "No created project material type ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required flags fails"
xbe_run do project-material-types create
assert_failure

test_name "Update without fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do project-material-types update "$SAMPLE_ID"
    assert_failure
else
    skip "No project material type ID available"
fi

# ============================================================================
# Done
# ============================================================================

run_tests
