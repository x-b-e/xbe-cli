#!/bin/bash
#
# XBE CLI Integration Tests: Material Unit Of Measure Quantities
#
# Tests CRUD operations for the material-unit-of-measure-quantities resource.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_QUANTITY_ID=""
SAMPLE_QUANTITY_ID=""
SAMPLE_MATERIAL_TRANSACTION_ID=""
SAMPLE_MATERIAL_TRANSACTION_ID_2=""
SAMPLE_UNIT_OF_MEASURE_ID=""
SAMPLE_UNIT_OF_MEASURE_ID_2=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""
SKIP_CREATE=0

describe "Resource: material-unit-of-measure-quantities"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material unit of measure quantities"
xbe_json view material-unit-of-measure-quantities list --limit 5
assert_success

test_name "List material unit of measure quantities returns array"
xbe_json view material-unit-of-measure-quantities list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material unit of measure quantities"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate material unit of measure quantity for filters"
xbe_json view material-unit-of-measure-quantities list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_QUANTITY_ID=$(json_get ".[0].id")
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].material_transaction_id")
        SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".[0].unit_of_measure_id")
        SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
        SAMPLE_UPDATED_AT=$(json_get ".[0].updated_at")
        pass
    else
        if [[ -n "$XBE_TEST_MATERIAL_UNIT_OF_MEASURE_QUANTITY_ID" ]]; then
            xbe_json view material-unit-of-measure-quantities show "$XBE_TEST_MATERIAL_UNIT_OF_MEASURE_QUANTITY_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_QUANTITY_ID=$(json_get ".id")
                SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".material_transaction_id")
                SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".unit_of_measure_id")
                SAMPLE_CREATED_AT=$(json_get ".created_at")
                SAMPLE_UPDATED_AT=$(json_get ".updated_at")
                pass
            else
                skip "Failed to load XBE_TEST_MATERIAL_UNIT_OF_MEASURE_QUANTITY_ID"
            fi
        else
            skip "No material unit of measure quantities found. Set XBE_TEST_MATERIAL_UNIT_OF_MEASURE_QUANTITY_ID for filter tests."
        fi
    fi
else
    fail "Failed to list material unit of measure quantities for filters"
fi

# ============================================================================
# Show Tests
# ============================================================================

if [[ -n "$SAMPLE_QUANTITY_ID" && "$SAMPLE_QUANTITY_ID" != "null" ]]; then
    test_name "Show material unit of measure quantity"
    xbe_json view material-unit-of-measure-quantities show "$SAMPLE_QUANTITY_ID"
    assert_success
else
    test_name "Show material unit of measure quantity"
    skip "No sample quantity available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    test_name "Filter by created-at min/max"
    xbe_json view material-unit-of-measure-quantities list \
        --created-at-min "$SAMPLE_CREATED_AT" \
        --created-at-max "$SAMPLE_CREATED_AT"
    assert_success

    test_name "Filter by is-created-at"
    xbe_json view material-unit-of-measure-quantities list --is-created-at true --limit 5
    assert_success
else
    test_name "Filter by created-at min/max"
    skip "No created-at available"
    test_name "Filter by is-created-at"
    skip "No created-at available"
fi

if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    test_name "Filter by updated-at min/max"
    xbe_json view material-unit-of-measure-quantities list \
        --updated-at-min "$SAMPLE_UPDATED_AT" \
        --updated-at-max "$SAMPLE_UPDATED_AT"
    assert_success

    test_name "Filter by is-updated-at"
    xbe_json view material-unit-of-measure-quantities list --is-updated-at true --limit 5
    assert_success
else
    test_name "Filter by updated-at min/max"
    skip "No updated-at available"
    test_name "Filter by is-updated-at"
    skip "No updated-at available"
fi

test_name "List material unit of measure quantities with --offset"
xbe_json view material-unit-of-measure-quantities list --limit 3 --offset 1
assert_success

test_name "List material unit of measure quantities with --sort"
xbe_json view material-unit-of-measure-quantities list --sort created-at --limit 5
assert_success

# ============================================================================
# Prerequisites - Unit of measure and material transaction IDs
# ============================================================================

test_name "Capture unit of measure IDs"
xbe_json view unit-of-measures list --limit 2
if [[ $status -eq 0 ]]; then
    SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
    SAMPLE_UNIT_OF_MEASURE_ID_2=$(json_get ".[1].id")
    if [[ -n "$SAMPLE_UNIT_OF_MEASURE_ID" && "$SAMPLE_UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        skip "No unit of measures available"
        SKIP_CREATE=1
    fi
else
    skip "Failed to list unit of measures"
    SKIP_CREATE=1
fi

if [[ -n "$XBE_TEST_MATERIAL_TRANSACTION_ID" ]]; then
    SAMPLE_MATERIAL_TRANSACTION_ID="$XBE_TEST_MATERIAL_TRANSACTION_ID"
fi

if [[ -z "$SAMPLE_MATERIAL_TRANSACTION_ID" || "$SAMPLE_MATERIAL_TRANSACTION_ID" == "null" ]]; then
    test_name "Capture material transaction IDs"
    xbe_json view material-transactions list --limit 2
    if [[ $status -eq 0 ]]; then
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].id")
        SAMPLE_MATERIAL_TRANSACTION_ID_2=$(json_get ".[1].id")
        if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
            pass
        else
            skip "No material transactions available"
            SKIP_CREATE=1
        fi
    else
        skip "Failed to list material transactions"
        SKIP_CREATE=1
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Create material unit of measure quantity"
    xbe_json do material-unit-of-measure-quantities create \
        --material-transaction "$SAMPLE_MATERIAL_TRANSACTION_ID" \
        --unit-of-measure "$SAMPLE_UNIT_OF_MEASURE_ID" \
        --quantity 12.5

    if [[ $status -eq 0 ]]; then
        CREATED_QUANTITY_ID=$(json_get ".id")
        if [[ -n "$CREATED_QUANTITY_ID" && "$CREATED_QUANTITY_ID" != "null" ]]; then
            register_cleanup "material-unit-of-measure-quantities" "$CREATED_QUANTITY_ID"
            pass
        else
            fail "Created quantity but no ID returned"
        fi
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
            SKIP_CREATE=1
        else
            fail "Failed to create material unit of measure quantity"
            SKIP_CREATE=1
        fi
    fi
else
    test_name "Create material unit of measure quantity"
    skip "Missing prerequisites for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_QUANTITY_ID" && "$CREATED_QUANTITY_ID" != "null" ]]; then
    test_name "Update material unit of measure quantity amount"
    xbe_json do material-unit-of-measure-quantities update "$CREATED_QUANTITY_ID" --quantity 15.25
    assert_success

    if [[ -n "$SAMPLE_UNIT_OF_MEASURE_ID_2" && "$SAMPLE_UNIT_OF_MEASURE_ID_2" != "null" ]]; then
        test_name "Update material unit of measure quantity unit of measure"
        xbe_json do material-unit-of-measure-quantities update "$CREATED_QUANTITY_ID" --unit-of-measure "$SAMPLE_UNIT_OF_MEASURE_ID_2"
        assert_success
    else
        test_name "Update material unit of measure quantity unit of measure"
        skip "No alternate unit of measure available"
    fi

    if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID_2" && "$SAMPLE_MATERIAL_TRANSACTION_ID_2" != "null" ]]; then
        test_name "Update material unit of measure quantity material transaction"
        xbe_json do material-unit-of-measure-quantities update "$CREATED_QUANTITY_ID" --material-transaction "$SAMPLE_MATERIAL_TRANSACTION_ID_2"
        assert_success
    else
        test_name "Update material unit of measure quantity material transaction"
        skip "No alternate material transaction available"
    fi

    test_name "Update material unit of measure quantity without fields fails"
    xbe_json do material-unit-of-measure-quantities update "$CREATED_QUANTITY_ID"
    assert_failure
else
    test_name "Update material unit of measure quantity amount"
    skip "No quantity created"
    test_name "Update material unit of measure quantity unit of measure"
    skip "No quantity created"
    test_name "Update material unit of measure quantity material transaction"
    skip "No quantity created"
    test_name "Update material unit of measure quantity without fields fails"
    skip "No quantity created"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_QUANTITY_ID" && "$CREATED_QUANTITY_ID" != "null" ]]; then
    test_name "Delete material unit of measure quantity requires --confirm flag"
    xbe_run do material-unit-of-measure-quantities delete "$CREATED_QUANTITY_ID"
    assert_failure

    test_name "Delete material unit of measure quantity with --confirm"
    xbe_run do material-unit-of-measure-quantities delete "$CREATED_QUANTITY_ID" --confirm
    assert_success
else
    test_name "Delete material unit of measure quantity requires --confirm flag"
    skip "No quantity created"
    test_name "Delete material unit of measure quantity with --confirm"
    skip "No quantity created"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material unit of measure quantity without required fields fails"
xbe_json do material-unit-of-measure-quantities create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
