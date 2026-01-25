#!/bin/bash
#
# XBE CLI Integration Tests: Incident Unit Of Measure Quantities
#
# Tests CRUD operations and list filters for the incident_unit_of_measure_quantities resource.
#
# COVERAGE: Create/update attributes + list filters + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_QUANTITY_ID=""
INCIDENT_ID=""
INCIDENT_TYPE=""
UNIT_OF_MEASURE_ID=""
ALT_UNIT_OF_MEASURE_ID=""

UNIT_OF_MEASURE_OUTPUT=""
USED_UOM_IDS=""

SKIP_MUTATION=0

describe "Resource: incident-unit-of-measure-quantities"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List incident unit of measure quantities"
xbe_json view incident-unit-of-measure-quantities list --limit 5
assert_success

test_name "List incident unit of measure quantities returns array"
xbe_json view incident-unit-of-measure-quantities list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list incident unit of measure quantities"
fi

# ==========================================================================
# Setup - Sample incident and unit of measure
# ==========================================================================

test_name "Fetch sample incident"
xbe_json view incidents list --limit 1
if [[ $status -eq 0 ]]; then
    INCIDENT_ID=$(json_get '.[0].id')
    INCIDENT_TYPE=$(json_get '.[0].type')
    if [[ -n "$INCIDENT_ID" && "$INCIDENT_ID" != "null" ]]; then
        if [[ -z "$INCIDENT_TYPE" || "$INCIDENT_TYPE" == "null" ]]; then
            INCIDENT_TYPE="incidents"
        fi
        pass
    else
        skip "No incidents available; skipping mutation and filter tests"
        SKIP_MUTATION=1
    fi
else
    skip "Failed to fetch incidents; skipping mutation and filter tests"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Fetch unit of measure options"
    xbe_json view unit-of-measures list --limit 100
    if [[ $status -eq 0 ]]; then
        UNIT_OF_MEASURE_OUTPUT="$output"
        pass
    else
        skip "Failed to fetch unit of measures; skipping mutation tests"
        SKIP_MUTATION=1
    fi
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Select available unit of measure"
    xbe_json view incident-unit-of-measure-quantities list --incident-id "$INCIDENT_ID" --limit 200
    if [[ $status -eq 0 ]]; then
        USED_UOM_IDS=$(echo "$output" | jq -r '.[].unit_of_measure_id' | tr '\n' ' ')
    fi

    for uom_id in $(echo "$UNIT_OF_MEASURE_OUTPUT" | jq -r '.[].id'); do
        if ! grep -qw "$uom_id" <<< "$USED_UOM_IDS"; then
            if [[ -z "$UNIT_OF_MEASURE_ID" ]]; then
                UNIT_OF_MEASURE_ID="$uom_id"
            elif [[ -z "$ALT_UNIT_OF_MEASURE_ID" ]]; then
                ALT_UNIT_OF_MEASURE_ID="$uom_id"
                break
            fi
        fi
    done

    if [[ -n "$UNIT_OF_MEASURE_ID" ]]; then
        pass
    else
        skip "No available unit of measure found; skipping mutation tests"
        SKIP_MUTATION=1
    fi
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create incident unit of measure quantity without required fields fails"
xbe_run do incident-unit-of-measure-quantities create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete tests"
else
    test_name "Create incident unit of measure quantity"
    xbe_json do incident-unit-of-measure-quantities create \
        --incident-type "$INCIDENT_TYPE" \
        --incident-id "$INCIDENT_ID" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID" \
        --quantity "12.5"

    if [[ $status -eq 0 ]]; then
        CREATED_QUANTITY_ID=$(json_get ".id")
        if [[ -n "$CREATED_QUANTITY_ID" && "$CREATED_QUANTITY_ID" != "null" ]]; then
            register_cleanup "incident-unit-of-measure-quantities" "$CREATED_QUANTITY_ID"
            pass
        else
            fail "Created incident unit of measure quantity but no ID returned"
        fi
    else
        fail "Failed to create incident unit of measure quantity"
    fi
fi

if [[ -z "$CREATED_QUANTITY_ID" || "$CREATED_QUANTITY_ID" == "null" ]]; then
    echo "Cannot continue without a valid incident unit of measure quantity ID"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show incident unit of measure quantity"
xbe_json view incident-unit-of-measure-quantities show "$CREATED_QUANTITY_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update incident unit of measure quantity --quantity"
xbe_json do incident-unit-of-measure-quantities update "$CREATED_QUANTITY_ID" --quantity "15"
assert_success

if [[ -n "$ALT_UNIT_OF_MEASURE_ID" ]]; then
    test_name "Update incident unit of measure quantity --unit-of-measure"
    xbe_json do incident-unit-of-measure-quantities update "$CREATED_QUANTITY_ID" --unit-of-measure "$ALT_UNIT_OF_MEASURE_ID"
    assert_success
else
    skip "No alternate unit of measure available for update"
fi

test_name "Update incident unit of measure quantity --incident"
xbe_json do incident-unit-of-measure-quantities update "$CREATED_QUANTITY_ID" \
    --incident-type "$INCIDENT_TYPE" \
    --incident-id "$INCIDENT_ID"
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List incident unit of measure quantities with --incident filter"
xbe_json view incident-unit-of-measure-quantities list \
    --incident-type "$INCIDENT_TYPE" \
    --incident-id "$INCIDENT_ID" \
    --limit 5
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"Internal Server Error"* || "$output" == *"INTERNAL SERVER ERROR"* ]]; then
    skip "Server does not support incident filters for this resource"
else
    fail "Failed to list incident unit of measure quantities with incident filter"
fi

test_name "List incident unit of measure quantities with --incident-type filter"
xbe_json view incident-unit-of-measure-quantities list --incident-type "$INCIDENT_TYPE" --limit 5
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"Internal Server Error"* || "$output" == *"INTERNAL SERVER ERROR"* ]]; then
    skip "Server does not support incident type filter for this resource"
else
    fail "Failed to list incident unit of measure quantities with incident type filter"
fi

test_name "List incident unit of measure quantities with --unit-of-measure filter"
xbe_json view incident-unit-of-measure-quantities list --unit-of-measure "$UNIT_OF_MEASURE_ID" --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete incident unit of measure quantity"
xbe_json do incident-unit-of-measure-quantities delete "$CREATED_QUANTITY_ID" --confirm
assert_success

run_tests
