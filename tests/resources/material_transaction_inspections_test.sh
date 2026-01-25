#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Inspections
#
# Tests create, update, delete operations and list filters for the
# material_transaction_inspections resource.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MT_ID=""
CREATED_INSPECTION_ID=""

describe "Resource: material-transaction-inspections"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material transaction for inspection"
TICKET_NUM="TI$(date +%s)"
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

if [[ -z "$CREATED_MT_ID" || "$CREATED_MT_ID" == "null" ]]; then
    echo "Cannot continue without a valid material transaction ID"
    run_tests
fi

test_name "Create material transaction inspection"
NOTE="Inspection $(date +%s)"
xbe_json do material-transaction-inspections create \
    --material-transaction "$CREATED_MT_ID" \
    --status open \
    --strategy delivery_site_personnel \
    --note "$NOTE"

if [[ $status -eq 0 ]]; then
    CREATED_INSPECTION_ID=$(json_get ".id")
    if [[ -n "$CREATED_INSPECTION_ID" && "$CREATED_INSPECTION_ID" != "null" ]]; then
        register_cleanup "material-transaction-inspections" "$CREATED_INSPECTION_ID"
        pass
    else
        fail "Created inspection but no ID returned"
    fi
else
    fail "Failed to create material transaction inspection: $output"
fi

if [[ -z "$CREATED_INSPECTION_ID" || "$CREATED_INSPECTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid inspection ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material transaction inspection"
UPDATED_NOTE="Updated inspection $(date +%s)"
xbe_json do material-transaction-inspections update "$CREATED_INSPECTION_ID" \
    --note "$UPDATED_NOTE" \
    --status closed
assert_success

test_name "Show inspection reflects updates"
xbe_json view material-transaction-inspections show "$CREATED_INSPECTION_ID"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".status" "closed"
else
    fail "Failed to show material transaction inspection"
fi

CHANGED_BY_ID=$(json_get ".changed_by_id")

# ============================================================================
# Lookup Helpers
# ============================================================================

MATERIAL_SITE_ID=""

test_name "Fetch material site for delivery-site filter"
xbe_json view material-sites list --limit 1
if [[ $status -eq 0 ]]; then
    MATERIAL_SITE_ID=$(json_get ".[0].id")
    if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
        pass
    else
        skip "No material site ID available"
    fi
else
    skip "Failed to list material sites"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction inspections"
xbe_json view material-transaction-inspections list --limit 10
assert_success

test_name "List inspections returns array"
xbe_json view material-transaction-inspections list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction inspections"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List inspections with --material-transaction filter"
xbe_json view material-transaction-inspections list --material-transaction "$CREATED_MT_ID" --limit 10
assert_success

test_name "List inspections with --status filter"
xbe_json view material-transaction-inspections list --status open --limit 10
assert_success

test_name "List inspections with --strategy filter"
xbe_json view material-transaction-inspections list --strategy delivery_site_personnel --limit 10
assert_success

test_name "List inspections with --changed-by filter"
if [[ -n "$CHANGED_BY_ID" && "$CHANGED_BY_ID" != "null" ]]; then
    xbe_json view material-transaction-inspections list --changed-by "$CHANGED_BY_ID" --limit 10
else
    xbe_json view material-transaction-inspections list --changed-by 123 --limit 10
fi
assert_success

test_name "List inspections with --delivery-site filter"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view material-transaction-inspections list --delivery-site-type MaterialSite --delivery-site-id "$MATERIAL_SITE_ID" --limit 10
    assert_success
else
    skip "No material site ID available for delivery-site filter"
fi

test_name "List inspections with --trip filter"
xbe_json view material-transaction-inspections list --trip 123 --limit 10
assert_success

test_name "List inspections with --trip-id filter"
xbe_json view material-transaction-inspections list --trip-id 123 --limit 10
assert_success

test_name "List inspections with --tender-job-schedule-shift filter"
xbe_json view material-transaction-inspections list --tender-job-schedule-shift 123 --limit 10
assert_success

test_name "List inspections with --tender-job-schedule-shift-id filter"
xbe_json view material-transaction-inspections list --tender-job-schedule-shift-id 123 --limit 10
assert_success

test_name "List inspections with --customer filter"
xbe_json view material-transaction-inspections list --customer 123 --limit 10
assert_success

test_name "List inspections with --customer-id filter"
xbe_json view material-transaction-inspections list --customer-id 123 --limit 10
assert_success

test_name "List inspections with --broker filter"
xbe_json view material-transaction-inspections list --broker 123 --limit 10
assert_success

test_name "List inspections with --broker-id filter"
xbe_json view material-transaction-inspections list --broker-id 123 --limit 10
assert_success

test_name "List inspections with --material-supplier filter"
xbe_json view material-transaction-inspections list --material-supplier 123 --limit 10
assert_success

test_name "List inspections with --material-supplier-id filter"
xbe_json view material-transaction-inspections list --material-supplier-id 123 --limit 10
assert_success

test_name "List inspections with --job-production-plan filter"
xbe_json view material-transaction-inspections list --job-production-plan 123 --limit 10
assert_success

test_name "List inspections with --job-production-plan-id filter"
xbe_json view material-transaction-inspections list --job-production-plan-id 123 --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete inspection requires --confirm flag"
xbe_json do material-transaction-inspections delete "$CREATED_INSPECTION_ID"
assert_failure

test_name "Delete inspection with --confirm"
xbe_json do material-transaction-inspections delete "$CREATED_INSPECTION_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
