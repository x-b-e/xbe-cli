#!/bin/bash
#
# XBE CLI Integration Tests: Material Purchase Order Releases
#
# Tests create, update, delete operations and list filters for the
# material_purchase_order_releases resource.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RELEASE_ID=""
PURCHASE_ORDER_ID="${XBE_TEST_MATERIAL_PURCHASE_ORDER_ID:-}"
EXISTING_RELEASE_ID=""

# Optional relationship IDs for update tests
RELATIONSHIP_RELEASE_ID="${XBE_TEST_ACTIVE_MATERIAL_PURCHASE_ORDER_RELEASE_ID:-}"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
TENDER_JOB_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
JOB_SHIFT_ID="${XBE_TEST_JOB_SCHEDULE_SHIFT_ID:-}"

describe "Resource: material-purchase-order-releases"

# ============================================================================
# Seed IDs from existing data if available
# ============================================================================

test_name "Lookup existing release (if any)"
xbe_json view material-purchase-order-releases list --limit 1
if [[ $status -eq 0 ]]; then
    EXISTING_RELEASE_ID=$(json_get ".[0].id")
    if [[ -z "$PURCHASE_ORDER_ID" ]]; then
        PURCHASE_ORDER_ID=$(json_get ".[0].purchase_order_id")
    fi
    pass
else
    fail "Failed to list material purchase order releases"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material purchase order release"
if [[ -n "$PURCHASE_ORDER_ID" && "$PURCHASE_ORDER_ID" != "null" ]]; then
    xbe_json do material-purchase-order-releases create \
        --purchase-order "$PURCHASE_ORDER_ID" \
        --quantity "0.10"

    if [[ $status -eq 0 ]]; then
        CREATED_RELEASE_ID=$(json_get ".id")
        if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
            register_cleanup "material-purchase-order-releases" "$CREATED_RELEASE_ID"
            pass
        else
            fail "Created release but no ID returned"
        fi
    else
        if [[ -z "$XBE_TEST_MATERIAL_PURCHASE_ORDER_ID" ]]; then
            skip "Create failed with inferred purchase order ID"
        else
            fail "Failed to create material purchase order release: $output"
        fi
    fi
else
    skip "No purchase order ID available (set XBE_TEST_MATERIAL_PURCHASE_ORDER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material purchase order release quantity"
if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
    xbe_json do material-purchase-order-releases update "$CREATED_RELEASE_ID" --quantity "0.20"
    assert_success
else
    skip "No release ID available"
fi


test_name "Update material purchase order release status"
if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
    xbe_json do material-purchase-order-releases update "$CREATED_RELEASE_ID" --status editing
    assert_success
else
    skip "No release ID available"
fi


test_name "Update material purchase order release skip-validate flag"
if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
    xbe_json do material-purchase-order-releases update "$CREATED_RELEASE_ID" --skip-validate-tender-job-schedule-shift-match
    assert_success
else
    skip "No release ID available"
fi

# Relationship updates require active releases; run only when explicitly provided

test_name "Update release trucker relationship"
if [[ -n "$RELATIONSHIP_RELEASE_ID" && -n "$TRUCKER_ID" ]]; then
    xbe_json do material-purchase-order-releases update "$RELATIONSHIP_RELEASE_ID" --trucker "$TRUCKER_ID"
    assert_success
else
    skip "Set XBE_TEST_ACTIVE_MATERIAL_PURCHASE_ORDER_RELEASE_ID and XBE_TEST_TRUCKER_ID to run"
fi


test_name "Update release tender job schedule shift relationship"
if [[ -n "$RELATIONSHIP_RELEASE_ID" && -n "$TENDER_JOB_SHIFT_ID" ]]; then
    xbe_json do material-purchase-order-releases update "$RELATIONSHIP_RELEASE_ID" --tender-job-schedule-shift "$TENDER_JOB_SHIFT_ID"
    assert_success
else
    skip "Set XBE_TEST_ACTIVE_MATERIAL_PURCHASE_ORDER_RELEASE_ID and XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID to run"
fi


test_name "Update release job schedule shift relationship"
if [[ -n "$RELATIONSHIP_RELEASE_ID" && -n "$JOB_SHIFT_ID" ]]; then
    xbe_json do material-purchase-order-releases update "$RELATIONSHIP_RELEASE_ID" --job-schedule-shift "$JOB_SHIFT_ID"
    assert_success
else
    skip "Set XBE_TEST_ACTIVE_MATERIAL_PURCHASE_ORDER_RELEASE_ID and XBE_TEST_JOB_SCHEDULE_SHIFT_ID to run"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material purchase order releases"
xbe_json view material-purchase-order-releases list --limit 10
assert_success

test_name "List material purchase order releases returns array"
xbe_json view material-purchase-order-releases list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material purchase order releases"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List releases with --status filter"
xbe_json view material-purchase-order-releases list --status editing --limit 10
assert_success

test_name "List releases with --release-token filter"
xbe_json view material-purchase-order-releases list --release-token ABC123 --limit 10
assert_success

test_name "List releases with --purchase-order filter"
xbe_json view material-purchase-order-releases list --purchase-order 1 --limit 10
assert_success

test_name "List releases with --trucker filter"
xbe_json view material-purchase-order-releases list --trucker 1 --limit 10
assert_success

test_name "List releases with --tender-job-schedule-shift filter"
xbe_json view material-purchase-order-releases list --tender-job-schedule-shift 1 --limit 10
assert_success

test_name "List releases with --job-schedule-shift filter"
xbe_json view material-purchase-order-releases list --job-schedule-shift 1 --limit 10
assert_success

test_name "List releases with --quantity filter"
xbe_json view material-purchase-order-releases list --quantity 1 --limit 10
assert_success

test_name "List releases with --broker filter"
xbe_json view material-purchase-order-releases list --broker 1 --limit 10
assert_success

test_name "List releases with --customer filter"
xbe_json view material-purchase-order-releases list --customer 1 --limit 10
assert_success

test_name "List releases with --valid-for-customer filter"
xbe_json view material-purchase-order-releases list --valid-for-customer 1 --limit 10
assert_success

test_name "List releases with --is-assigned filter (true)"
xbe_json view material-purchase-order-releases list --is-assigned true --limit 10
assert_success

test_name "List releases with --material-supplier filter"
xbe_json view material-purchase-order-releases list --material-supplier 1 --limit 10
assert_success

test_name "List releases with --active filter"
xbe_json view material-purchase-order-releases list --active true --limit 10
assert_success

test_name "List releases with --not-active filter"
xbe_json view material-purchase-order-releases list --not-active true --limit 10
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material purchase order release"
if [[ -n "$EXISTING_RELEASE_ID" && "$EXISTING_RELEASE_ID" != "null" ]]; then
    xbe_json view material-purchase-order-releases show "$EXISTING_RELEASE_ID"
    assert_success
else
    skip "No existing release ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material purchase order release requires --confirm"
if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
    xbe_json do material-purchase-order-releases delete "$CREATED_RELEASE_ID"
    assert_failure
else
    skip "No release ID available"
fi


test_name "Delete material purchase order release with --confirm"
if [[ -n "$CREATED_RELEASE_ID" && "$CREATED_RELEASE_ID" != "null" ]]; then
    xbe_json do material-purchase-order-releases delete "$CREATED_RELEASE_ID" --confirm
    assert_success
else
    skip "No release ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
