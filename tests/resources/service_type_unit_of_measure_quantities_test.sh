#!/bin/bash
#
# XBE CLI Integration Tests: Service Type Unit of Measure Quantities
#
# Tests list/show/create/update/delete operations for the service-type-unit-of-measure-quantities resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STUOM_ID=""
SAMPLE_QUANTITY=""
SAMPLE_CALCULATED_QUANTITY=""
SAMPLE_EXPLICIT_QUANTITY=""
SAMPLE_QUANTIFIES_TYPE=""
SAMPLE_QUANTIFIES_ID=""

CREATED_ID=""

QUANTIFIES_TYPE="${XBE_TEST_QUANTIFIES_TYPE:-time-cards}"
QUANTIFIES_ID="${XBE_TEST_TIME_CARD_ID:-${XBE_TEST_QUANTIFIES_ID:-}}"
STUOM_ID="${XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID:-}"
MATERIAL_TYPE_ID="${XBE_TEST_MATERIAL_TYPE_ID:-}"
TRAILER_CLASSIFICATION_ID="${XBE_TEST_TRAILER_CLASSIFICATION_ID:-}"
BROKER_ID="${XBE_TEST_BROKER_ID:-}"

describe "Resource: service-type-unit-of-measure-quantities"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List service type unit of measure quantities"
xbe_json view service-type-unit-of-measure-quantities list --limit 5
assert_success

test_name "List service type unit of measure quantities returns array"
xbe_json view service-type-unit-of-measure-quantities list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list service type unit of measure quantities"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample service type unit of measure quantity"
xbe_json view service-type-unit-of-measure-quantities list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STUOM_ID=$(json_get ".[0].service_type_unit_of_measure_id")
    SAMPLE_QUANTITY=$(json_get ".[0].quantity")
    SAMPLE_CALCULATED_QUANTITY=$(json_get ".[0].calculated_quantity")
    SAMPLE_EXPLICIT_QUANTITY=$(json_get ".[0].explicit_quantity")
    SAMPLE_QUANTIFIES_TYPE=$(json_get ".[0].quantifies_type")
    SAMPLE_QUANTIFIES_ID=$(json_get ".[0].quantifies_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No service type unit of measure quantities available for follow-on tests"
    fi
else
    skip "Could not list service type unit of measure quantities to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List quantities with --service-type-unit-of-measure filter"
if [[ -n "$SAMPLE_STUOM_ID" && "$SAMPLE_STUOM_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities list --service-type-unit-of-measure "$SAMPLE_STUOM_ID" --limit 5
    assert_success
else
    skip "No service type unit of measure ID available"
fi

test_name "List quantities with --quantity filter"
if [[ -n "$SAMPLE_QUANTITY" && "$SAMPLE_QUANTITY" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities list --quantity "$SAMPLE_QUANTITY" --limit 5
    assert_success
else
    skip "No quantity available"
fi

test_name "List quantities with --explicit-quantity filter"
if [[ -n "$SAMPLE_EXPLICIT_QUANTITY" && "$SAMPLE_EXPLICIT_QUANTITY" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities list --explicit-quantity "$SAMPLE_EXPLICIT_QUANTITY" --limit 5
    assert_success
else
    skip "No explicit quantity available"
fi

test_name "List quantities with --calculated-quantity filter"
if [[ -n "$SAMPLE_CALCULATED_QUANTITY" && "$SAMPLE_CALCULATED_QUANTITY" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities list --calculated-quantity "$SAMPLE_CALCULATED_QUANTITY" --limit 5
    assert_success
else
    skip "No calculated quantity available"
fi

test_name "List quantities with --quantifies-type/--quantifies-id filter"
if [[ -n "$SAMPLE_QUANTIFIES_TYPE" && "$SAMPLE_QUANTIFIES_TYPE" != "null" && -n "$SAMPLE_QUANTIFIES_ID" && "$SAMPLE_QUANTIFIES_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities list --quantifies-type "$SAMPLE_QUANTIFIES_TYPE" --quantifies-id "$SAMPLE_QUANTIFIES_ID" --limit 5
    assert_success
else
    skip "No quantifies type/id available"
fi

test_name "List quantities with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available (set XBE_TEST_BROKER_ID to enable)"
fi

test_name "List quantities with --quantifies-start-on-min filter"
xbe_json view service-type-unit-of-measure-quantities list --quantifies-start-on-min "2020-01-01" --limit 5
assert_success

test_name "List quantities with --quantifies-start-on-max filter"
xbe_json view service-type-unit-of-measure-quantities list --quantifies-start-on-max "2030-01-01" --limit 5
assert_success

test_name "List quantities with --limit"
xbe_json view service-type-unit-of-measure-quantities list --limit 5
assert_success

test_name "List quantities with --offset"
xbe_json view service-type-unit-of-measure-quantities list --limit 5 --offset 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show service type unit of measure quantity"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view service-type-unit-of-measure-quantities show "$SAMPLE_ID"
    assert_success
else
    skip "No sample quantity ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires --service-type-unit-of-measure"
xbe_run do service-type-unit-of-measure-quantities create --quantifies-type time-cards --quantifies-id "123"
assert_failure

test_name "Create requires --quantifies-type"
xbe_run do service-type-unit-of-measure-quantities create --service-type-unit-of-measure "123" --quantifies-id "123"
assert_failure

test_name "Create requires --quantifies-id"
xbe_run do service-type-unit-of-measure-quantities create --service-type-unit-of-measure "123" --quantifies-type time-cards
assert_failure

test_name "Create service type unit of measure quantity"
if [[ -n "$STUOM_ID" && "$STUOM_ID" != "null" && -n "$QUANTIFIES_ID" && "$QUANTIFIES_ID" != "null" ]]; then
    xbe_json do service-type-unit-of-measure-quantities create \
        --service-type-unit-of-measure "$STUOM_ID" \
        --quantifies-type "$QUANTIFIES_TYPE" \
        --quantifies-id "$QUANTIFIES_ID" \
        --quantity 5 \
        --explicit-quantity 5.5

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "service-type-unit-of-measure-quantities" "$CREATED_ID"
            pass
        else
            fail "Created quantity but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Validation"* ]] || [[ "$output" == *"unprocessable"* ]]; then
            skip "Create failed due to permissions or validation"
        else
            fail "Failed to create service type unit of measure quantity: $output"
        fi
    fi
else
    skip "Missing prerequisites. Set XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID and XBE_TEST_TIME_CARD_ID to enable create testing."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update without fields fails"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do service-type-unit-of-measure-quantities update "$CREATED_ID"
    assert_failure
else
    skip "No created quantity to update"
fi

test_name "Update quantity"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do service-type-unit-of-measure-quantities update "$CREATED_ID" --quantity 6
    assert_success
else
    skip "No created quantity to update"
fi

test_name "Update explicit quantity"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do service-type-unit-of-measure-quantities update "$CREATED_ID" --explicit-quantity 6.5
    assert_success
else
    skip "No created quantity to update"
fi

test_name "Update material type"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" && -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json do service-type-unit-of-measure-quantities update "$CREATED_ID" --material-type "$MATERIAL_TYPE_ID"
    assert_success

    xbe_json do service-type-unit-of-measure-quantities update "$CREATED_ID" --material-type ""
    assert_success
else
    skip "No material type ID available (set XBE_TEST_MATERIAL_TYPE_ID to enable)"
fi

test_name "Update trailer classification"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" && -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do service-type-unit-of-measure-quantities update "$CREATED_ID" --trailer-classification "$TRAILER_CLASSIFICATION_ID"
    assert_success

    xbe_json do service-type-unit-of-measure-quantities update "$CREATED_ID" --trailer-classification ""
    assert_success
else
    skip "No trailer classification ID available (set XBE_TEST_TRAILER_CLASSIFICATION_ID to enable)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do service-type-unit-of-measure-quantities delete "$CREATED_ID"
    assert_failure
else
    skip "No created quantity to delete"
fi

test_name "Delete service type unit of measure quantity with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do service-type-unit-of-measure-quantities delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created quantity to delete"
fi

run_tests
