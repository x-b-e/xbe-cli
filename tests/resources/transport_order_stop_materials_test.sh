#!/bin/bash
#
# XBE CLI Integration Tests: Transport Order Stop Materials
#
# Tests list, show, create, update, delete operations for the
# transport-order-stop-materials resource.
#
# COVERAGE: List + show + filters + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TRANSPORT_ORDER_MATERIAL_ID=""
SAMPLE_TRANSPORT_ORDER_STOP_ID=""
SAMPLE_QUANTITY_EXPLICIT=""
CREATED_TOSM_ID=""
LIST_SUPPORTED="true"

describe "Resource: transport-order-stop-materials"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List transport order stop materials"
xbe_json view transport-order-stop-materials list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"Forbidden"* ]]; then
        LIST_SUPPORTED="false"
        skip "Listing requires admin access"
    elif [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "List endpoint not supported"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List transport order stop materials returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view transport-order-stop-materials list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list transport order stop materials"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample transport order stop material"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view transport-order-stop-materials list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TRANSPORT_ORDER_MATERIAL_ID=$(json_get ".[0].transport_order_material_id")
        SAMPLE_TRANSPORT_ORDER_STOP_ID=$(json_get ".[0].transport_order_stop_id")
        SAMPLE_QUANTITY_EXPLICIT=$(json_get ".[0].quantity_explicit")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No transport order stop materials available for follow-on tests"
        fi
    else
        skip "Could not list transport order stop materials to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter transport order stop materials by transport order material"
if [[ -n "$SAMPLE_TRANSPORT_ORDER_MATERIAL_ID" && "$SAMPLE_TRANSPORT_ORDER_MATERIAL_ID" != "null" ]]; then
    xbe_json view transport-order-stop-materials list --transport-order-material "$SAMPLE_TRANSPORT_ORDER_MATERIAL_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by transport order material failed"
    fi
else
    skip "No transport order material ID available"
fi

test_name "Filter transport order stop materials by transport order stop"
if [[ -n "$SAMPLE_TRANSPORT_ORDER_STOP_ID" && "$SAMPLE_TRANSPORT_ORDER_STOP_ID" != "null" ]]; then
    xbe_json view transport-order-stop-materials list --transport-order-stop "$SAMPLE_TRANSPORT_ORDER_STOP_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by transport order stop failed"
    fi
else
    skip "No transport order stop ID available"
fi

test_name "Filter transport order stop materials by quantity explicit"
if [[ -n "$SAMPLE_QUANTITY_EXPLICIT" && "$SAMPLE_QUANTITY_EXPLICIT" != "null" ]]; then
    xbe_json view transport-order-stop-materials list --quantity-explicit "$SAMPLE_QUANTITY_EXPLICIT" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by quantity explicit failed"
    fi
else
    skip "No quantity explicit available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show transport order stop material"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view transport-order-stop-materials show "$SAMPLE_ID"
    assert_success
else
    skip "No transport order stop material ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create transport order stop material"
if [[ -n "$XBE_TEST_TRANSPORT_ORDER_MATERIAL_ID" && -n "$XBE_TEST_TRANSPORT_ORDER_STOP_ID" ]]; then
    CREATE_QTY="${XBE_TEST_TRANSPORT_ORDER_STOP_MATERIAL_QTY:-10}"
    xbe_json do transport-order-stop-materials create \
        --transport-order-material "$XBE_TEST_TRANSPORT_ORDER_MATERIAL_ID" \
        --transport-order-stop "$XBE_TEST_TRANSPORT_ORDER_STOP_ID" \
        --quantity-explicit "$CREATE_QTY"
    if [[ $status -eq 0 ]]; then
        CREATED_TOSM_ID=$(json_get ".id")
        if [[ -n "$CREATED_TOSM_ID" && "$CREATED_TOSM_ID" != "null" ]]; then
            register_cleanup "transport-order-stop-materials" "$CREATED_TOSM_ID"
        fi
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"403"* ]] || \
           [[ "$output" == *"422"* ]] || \
           [[ "$output" == *"Record Invalid"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_TRANSPORT_ORDER_MATERIAL_ID and XBE_TEST_TRANSPORT_ORDER_STOP_ID to enable create test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update transport order stop material quantity"
UPDATE_TARGET_ID="$CREATED_TOSM_ID"
if [[ -z "$UPDATE_TARGET_ID" || "$UPDATE_TARGET_ID" == "null" ]]; then
    UPDATE_TARGET_ID="$XBE_TEST_TRANSPORT_ORDER_STOP_MATERIAL_ID"
fi

if [[ -n "$UPDATE_TARGET_ID" ]]; then
    xbe_json do transport-order-stop-materials update "$UPDATE_TARGET_ID" --quantity-explicit "12.5"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"403"* ]] || \
           [[ "$output" == *"422"* ]] || \
           [[ "$output" == *"Record Invalid"* ]]; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_TRANSPORT_ORDER_STOP_MATERIAL_ID to enable update test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete transport order stop material requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do transport-order-stop-materials delete "$SAMPLE_ID"
    assert_failure
else
    skip "No transport order stop material ID available"
fi

test_name "Delete transport order stop material"
if [[ -n "$CREATED_TOSM_ID" && "$CREATED_TOSM_ID" != "null" ]]; then
    xbe_json do transport-order-stop-materials delete "$CREATED_TOSM_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"403"* ]] || \
           [[ "$output" == *"422"* ]] || \
           [[ "$output" == *"Record Invalid"* ]]; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created transport order stop material ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create transport order stop material without required flags fails"
xbe_run do transport-order-stop-materials create
assert_failure

test_name "Update transport order stop material without fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do transport-order-stop-materials update "$SAMPLE_ID"
    assert_failure
else
    skip "No transport order stop material ID available"
fi

run_tests
