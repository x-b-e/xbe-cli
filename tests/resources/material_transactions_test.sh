#!/bin/bash
#
# XBE CLI Integration Tests: Material Transactions
#
# Tests create and delete operations for the material_transactions resource.
# Material transactions track material deliveries and pickups.
#
# COMPLETE COVERAGE: Create, delete + list filters (no update)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MT_ID=""

describe "Resource: material-transactions"

# ============================================================================
# CREATE Tests
# ============================================================================

# Note: Creating material transactions may require existing material-types/material-sites
# We'll test basic creation - relationships are optional

test_name "Create material transaction with ticket number"
TICKET_NUM="T$(date +%s)"
TRANS_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
xbe_json do material-transactions create --ticket-number "$TICKET_NUM" --transaction-at "$TRANS_AT"

if [[ $status -eq 0 ]]; then
    CREATED_MT_ID=$(json_get ".id")
    if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
        register_cleanup "material-transactions" "$CREATED_MT_ID"
        pass
    else
        fail "Created material transaction but no ID returned"
    fi
else
    fail "Failed to create material transaction: $output"
fi

# Only continue if we successfully created a material transaction
if [[ -z "$CREATED_MT_ID" || "$CREATED_MT_ID" == "null" ]]; then
    echo "Cannot continue without a valid material transaction ID"
    run_tests
fi

test_name "Create material transaction with weights"
TICKET_NUM2="TW$(date +%s)"
TRANS_AT2="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
xbe_json do material-transactions create \
    --ticket-number "$TICKET_NUM2" \
    --transaction-at "$TRANS_AT2" \
    --tare-weight-lbs "15000" \
    --gross-weight-lbs "55000" \
    --net-weight-lbs "40000"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-transactions" "$id"
    pass
else
    fail "Failed to create material transaction with weights"
fi

test_name "Create material transaction with transaction-at"
TICKET_NUM3="TT$(date +%s)"
xbe_json do material-transactions create \
    --ticket-number "$TICKET_NUM3" \
    --transaction-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-transactions" "$id"
    pass
else
    fail "Failed to create material transaction with transaction-at"
fi

test_name "Create material transaction with BOL number"
TICKET_NUM4="TB$(date +%s)"
TRANS_AT4="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
xbe_json do material-transactions create \
    --ticket-number "$TICKET_NUM4" \
    --transaction-at "$TRANS_AT4" \
    --ticket-bol-number "BOL-$(date +%s)"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-transactions" "$id"
    pass
else
    fail "Failed to create material transaction with BOL number"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transactions"
xbe_json view material-transactions list --limit 10
assert_success

test_name "List material transactions returns array"
xbe_json view material-transactions list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transactions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List material transactions with --is-voided filter (false)"
xbe_json view material-transactions list --is-voided false --limit 10
assert_success

test_name "List material transactions with --is-voided filter (true)"
xbe_json view material-transactions list --is-voided true --limit 10
assert_success

test_name "List material transactions with --has-shift filter"
xbe_json view material-transactions list --has-shift true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List material transactions with --limit"
xbe_json view material-transactions list --limit 5
assert_success

test_name "List material transactions with --offset"
xbe_json view material-transactions list --limit 5 --offset 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material transaction requires --confirm flag"
xbe_json do material-transactions delete "$CREATED_MT_ID"
assert_failure

test_name "Delete material transaction with --confirm (soft delete)"
# Create a material transaction specifically for deletion
TICKET_DEL="TD$(date +%s)"
TRANS_AT_DEL="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
xbe_json do material-transactions create --ticket-number "$TICKET_DEL" --transaction-at "$TRANS_AT_DEL"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do material-transactions delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create material transaction for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
