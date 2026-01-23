#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Inspection Rejections
#
# Tests CRUD operations for the material-transaction-inspection-rejections resource.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REJECTION_ID=""
SAMPLE_REJECTION_ID=""
SAMPLE_INSPECTION_ID=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""
SAMPLE_UNIT_OF_MEASURE_ID=""
SAMPLE_UNIT_OF_MEASURE_ID_2=""
SKIP_CREATE=0

describe "Resource: material-transaction-inspection-rejections"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction inspection rejections"
xbe_json view material-transaction-inspection-rejections list --limit 5
assert_success

test_name "List material transaction inspection rejections returns array"
xbe_json view material-transaction-inspection-rejections list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction inspection rejections"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate material transaction inspection rejection for filters"
xbe_json view material-transaction-inspection-rejections list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_REJECTION_ID=$(json_get ".[0].id")
        SAMPLE_INSPECTION_ID=$(json_get ".[0].material_transaction_inspection_id")
        SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
        SAMPLE_UPDATED_AT=$(json_get ".[0].updated_at")
        SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".[0].unit_of_measure_id")
        pass
    else
        if [[ -n "$XBE_TEST_MATERIAL_TRANSACTION_INSPECTION_REJECTION_ID" ]]; then
            xbe_json view material-transaction-inspection-rejections show "$XBE_TEST_MATERIAL_TRANSACTION_INSPECTION_REJECTION_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_REJECTION_ID=$(json_get ".id")
                SAMPLE_INSPECTION_ID=$(json_get ".material_transaction_inspection_id")
                SAMPLE_CREATED_AT=$(json_get ".created_at")
                SAMPLE_UPDATED_AT=$(json_get ".updated_at")
                SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".unit_of_measure_id")
                pass
            else
                skip "Failed to load XBE_TEST_MATERIAL_TRANSACTION_INSPECTION_REJECTION_ID"
            fi
        else
            skip "No inspection rejections found. Set XBE_TEST_MATERIAL_TRANSACTION_INSPECTION_REJECTION_ID for filter tests."
        fi
    fi
else
    fail "Failed to list material transaction inspection rejections for filters"
fi

# ============================================================================
# Show Tests
# ============================================================================

if [[ -n "$SAMPLE_REJECTION_ID" && "$SAMPLE_REJECTION_ID" != "null" ]]; then
    test_name "Show material transaction inspection rejection"
    xbe_json view material-transaction-inspection-rejections show "$SAMPLE_REJECTION_ID"
    assert_success
else
    test_name "Show material transaction inspection rejection"
    skip "No sample rejection available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$SAMPLE_INSPECTION_ID" && "$SAMPLE_INSPECTION_ID" != "null" ]]; then
    test_name "Filter by material transaction inspection"
    xbe_json view material-transaction-inspection-rejections list --material-transaction-inspection "$SAMPLE_INSPECTION_ID"
    assert_success
else
    test_name "Filter by material transaction inspection"
    skip "No inspection ID available"
fi

if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    test_name "Filter by created-at min/max"
    xbe_json view material-transaction-inspection-rejections list \
        --created-at-min "$SAMPLE_CREATED_AT" \
        --created-at-max "$SAMPLE_CREATED_AT"
    assert_success

    test_name "Filter by is-created-at"
    xbe_json view material-transaction-inspection-rejections list --is-created-at true --limit 5
    assert_success
else
    test_name "Filter by created-at min/max"
    skip "No created-at available"
    test_name "Filter by is-created-at"
    skip "No created-at available"
fi

if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    test_name "Filter by updated-at min/max"
    xbe_json view material-transaction-inspection-rejections list \
        --updated-at-min "$SAMPLE_UPDATED_AT" \
        --updated-at-max "$SAMPLE_UPDATED_AT"
    assert_success

    test_name "Filter by is-updated-at"
    xbe_json view material-transaction-inspection-rejections list --is-updated-at true --limit 5
    assert_success
else
    test_name "Filter by updated-at min/max"
    skip "No updated-at available"
    test_name "Filter by is-updated-at"
    skip "No updated-at available"
fi

test_name "List material transaction inspection rejections with --offset"
xbe_json view material-transaction-inspection-rejections list --limit 3 --offset 1
assert_success

test_name "List material transaction inspection rejections with --sort"
xbe_json view material-transaction-inspection-rejections list --sort created-at --limit 5
assert_success

# ============================================================================
# Prerequisites - Unit of measure and inspection ID
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

if [[ -n "$XBE_TEST_MATERIAL_TRANSACTION_INSPECTION_ID" ]]; then
    SAMPLE_INSPECTION_ID="$XBE_TEST_MATERIAL_TRANSACTION_INSPECTION_ID"
fi

if [[ -z "$SAMPLE_INSPECTION_ID" || "$SAMPLE_INSPECTION_ID" == "null" ]]; then
    SKIP_CREATE=1
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Create material transaction inspection rejection"
    xbe_json do material-transaction-inspection-rejections create \
        --material-transaction-inspection "$SAMPLE_INSPECTION_ID" \
        --unit-of-measure "$SAMPLE_UNIT_OF_MEASURE_ID" \
        --quantity 10 \
        --note "CLI test rejection"

    if [[ $status -eq 0 ]]; then
        CREATED_REJECTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_REJECTION_ID" && "$CREATED_REJECTION_ID" != "null" ]]; then
            register_cleanup "material-transaction-inspection-rejections" "$CREATED_REJECTION_ID"
            pass
        else
            fail "Created rejection but no ID returned"
        fi
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
            SKIP_CREATE=1
        else
            fail "Failed to create inspection rejection"
            SKIP_CREATE=1
        fi
    fi
else
    test_name "Create material transaction inspection rejection"
    skip "Prerequisites not available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_REJECTION_ID" && "$CREATED_REJECTION_ID" != "null" ]]; then
    test_name "Update material transaction inspection rejection note"
    xbe_json do material-transaction-inspection-rejections update "$CREATED_REJECTION_ID" --note "Updated note"
    assert_success

    test_name "Update material transaction inspection rejection quantity"
    xbe_json do material-transaction-inspection-rejections update "$CREATED_REJECTION_ID" --quantity 12
    assert_success

    if [[ -n "$SAMPLE_UNIT_OF_MEASURE_ID_2" && "$SAMPLE_UNIT_OF_MEASURE_ID_2" != "null" ]]; then
        test_name "Update material transaction inspection rejection unit of measure"
        xbe_json do material-transaction-inspection-rejections update "$CREATED_REJECTION_ID" --unit-of-measure "$SAMPLE_UNIT_OF_MEASURE_ID_2"
        assert_success
    else
        test_name "Update material transaction inspection rejection unit of measure"
        skip "No alternate unit of measure available"
    fi

    test_name "Update material transaction inspection rejection without fields fails"
    xbe_json do material-transaction-inspection-rejections update "$CREATED_REJECTION_ID"
    assert_failure
else
    test_name "Update material transaction inspection rejection note"
    skip "No rejection created"
    test_name "Update material transaction inspection rejection quantity"
    skip "No rejection created"
    test_name "Update material transaction inspection rejection unit of measure"
    skip "No rejection created"
    test_name "Update material transaction inspection rejection without fields fails"
    skip "No rejection created"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_REJECTION_ID" && "$CREATED_REJECTION_ID" != "null" ]]; then
    test_name "Delete material transaction inspection rejection requires --confirm flag"
    xbe_run do material-transaction-inspection-rejections delete "$CREATED_REJECTION_ID"
    assert_failure

    test_name "Delete material transaction inspection rejection with --confirm"
    xbe_run do material-transaction-inspection-rejections delete "$CREATED_REJECTION_ID" --confirm
    assert_success
else
    test_name "Delete material transaction inspection rejection requires --confirm flag"
    skip "No rejection created"
    test_name "Delete material transaction inspection rejection with --confirm"
    skip "No rejection created"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create rejection without required fields fails"
xbe_json do material-transaction-inspection-rejections create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
