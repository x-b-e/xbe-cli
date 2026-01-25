#!/bin/bash
#
# XBE CLI Integration Tests: Material Purchase Order Release Redemptions
#
# Tests view/do operations for the material-purchase-order-release-redemptions resource.
#
# COVERAGE: list/show filters + create/update/delete (ticket-number + material-transaction)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REDEMPTION_ID=""
SAMPLE_ID=""
SAMPLE_RELEASE_ID=""
SAMPLE_PURCHASE_ORDER_ID=""
SAMPLE_MATERIAL_TRANSACTION_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_DRIVER_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_MATERIAL_SUPPLIER_ID=""
SAMPLE_MATERIAL_TYPE_ID=""
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_TICKET_NUMBER=""

RELEASE_ID=""
MTXN_ID=""


describe "Resource: material-purchase-order-release-redemptions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material purchase order release redemptions"
xbe_json view material-purchase-order-release-redemptions list --limit 5
assert_success

test_name "List material purchase order release redemptions returns array"
xbe_json view material-purchase-order-release-redemptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get '.[0].id // empty')
    SAMPLE_RELEASE_ID=$(json_get '.[0].release_id // empty')
    SAMPLE_PURCHASE_ORDER_ID=$(json_get '.[0].purchase_order_id // empty')
    SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get '.[0].material_transaction_id // empty')
    SAMPLE_TRUCKER_ID=$(json_get '.[0].trucker_id // empty')
    SAMPLE_DRIVER_ID=$(json_get '.[0].driver_id // empty')
    SAMPLE_BROKER_ID=$(json_get '.[0].broker_id // empty')
    SAMPLE_MATERIAL_SUPPLIER_ID=$(json_get '.[0].material_supplier_id // empty')
    SAMPLE_MATERIAL_TYPE_ID=$(json_get '.[0].material_type_id // empty')
    SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get '.[0].tender_job_schedule_shift_id // empty')
    SAMPLE_TICKET_NUMBER=$(json_get '.[0].ticket_number // empty')
else
    fail "Failed to list release redemptions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material purchase order release redemption"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions show "$SAMPLE_ID"
    assert_success
else
    skip "No redemption ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create release redemption with ticket number"
RELEASE_ID="${XBE_TEST_MATERIAL_PURCHASE_ORDER_RELEASE_ID:-$SAMPLE_RELEASE_ID}"
if [[ -n "$RELEASE_ID" && "$RELEASE_ID" != "null" ]]; then
    TICKET_NUMBER="T-${RELEASE_ID}-$(unique_suffix)"
    xbe_json do material-purchase-order-release-redemptions create \
        --release "$RELEASE_ID" \
        --ticket-number "$TICKET_NUMBER"
    if [[ $status -eq 0 ]]; then
        CREATED_REDEMPTION_ID=$(json_get '.id')
        if [[ -n "$CREATED_REDEMPTION_ID" && "$CREATED_REDEMPTION_ID" != "null" ]]; then
            register_cleanup "material-purchase-order-release-redemptions" "$CREATED_REDEMPTION_ID"
            pass
        else
            fail "Created redemption but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"must be approved"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create redemption: $output"
        fi
    fi
else
    skip "No release ID available (set XBE_TEST_MATERIAL_PURCHASE_ORDER_RELEASE_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update release redemption ticket number"
if [[ -n "$CREATED_REDEMPTION_ID" && "$CREATED_REDEMPTION_ID" != "null" ]]; then
    UPDATED_TICKET_NUMBER="T-UPDATED-$(unique_suffix)"
    xbe_json do material-purchase-order-release-redemptions update "$CREATED_REDEMPTION_ID" --ticket-number "$UPDATED_TICKET_NUMBER"
    assert_success
else
    skip "No created redemption to update"
fi

test_name "Update release redemption material transaction"
MTXN_ID="${XBE_TEST_MATERIAL_TRANSACTION_ID:-$SAMPLE_MATERIAL_TRANSACTION_ID}"
if [[ -n "$CREATED_REDEMPTION_ID" && "$CREATED_REDEMPTION_ID" != "null" && -n "$MTXN_ID" && "$MTXN_ID" != "null" ]]; then
    xbe_json do material-purchase-order-release-redemptions update "$CREATED_REDEMPTION_ID" --material-transaction "$MTXN_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Material transaction update blocked by server policy/validation"
        else
            fail "Failed to update material transaction: $output"
        fi
    fi
else
    skip "No created redemption or material transaction ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List release redemptions with --release filter"
if [[ -n "$SAMPLE_RELEASE_ID" && "$SAMPLE_RELEASE_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --release "$SAMPLE_RELEASE_ID" --limit 5
    assert_success
else
    skip "No release ID available"
fi

test_name "List release redemptions with --ticket-number filter"
if [[ -n "$SAMPLE_TICKET_NUMBER" && "$SAMPLE_TICKET_NUMBER" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --ticket-number "$SAMPLE_TICKET_NUMBER" --limit 5
    assert_success
else
    skip "No ticket number available"
fi

test_name "List release redemptions with --material-transaction filter"
if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --material-transaction "$SAMPLE_MATERIAL_TRANSACTION_ID" --limit 5
    assert_success
else
    skip "No material transaction ID available"
fi

test_name "List release redemptions with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List release redemptions with --broker-id filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --broker-id "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List release redemptions with --trucker filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --trucker "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List release redemptions with --trucker-id filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --trucker-id "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List release redemptions with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List release redemptions with --driver-id filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --driver-id "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List release redemptions with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List release redemptions with --tender-job-schedule-shift-id filter"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --tender-job-schedule-shift-id "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List release redemptions with --material-supplier filter"
if [[ -n "$SAMPLE_MATERIAL_SUPPLIER_ID" && "$SAMPLE_MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --material-supplier "$SAMPLE_MATERIAL_SUPPLIER_ID" --limit 5
    assert_success
else
    skip "No material supplier ID available"
fi

test_name "List release redemptions with --material-supplier-id filter"
if [[ -n "$SAMPLE_MATERIAL_SUPPLIER_ID" && "$SAMPLE_MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --material-supplier-id "$SAMPLE_MATERIAL_SUPPLIER_ID" --limit 5
    assert_success
else
    skip "No material supplier ID available"
fi

test_name "List release redemptions with --material-type filter"
if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --material-type "$SAMPLE_MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

test_name "List release redemptions with --material-type-id filter"
if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --material-type-id "$SAMPLE_MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

test_name "List release redemptions with --purchase-order filter"
if [[ -n "$SAMPLE_PURCHASE_ORDER_ID" && "$SAMPLE_PURCHASE_ORDER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --purchase-order "$SAMPLE_PURCHASE_ORDER_ID" --limit 5
    assert_success
else
    skip "No purchase order ID available"
fi

test_name "List release redemptions with --purchase-order-id filter"
if [[ -n "$SAMPLE_PURCHASE_ORDER_ID" && "$SAMPLE_PURCHASE_ORDER_ID" != "null" ]]; then
    xbe_json view material-purchase-order-release-redemptions list --purchase-order-id "$SAMPLE_PURCHASE_ORDER_ID" --limit 5
    assert_success
else
    skip "No purchase order ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete release redemption requires --confirm flag"
if [[ -n "$CREATED_REDEMPTION_ID" && "$CREATED_REDEMPTION_ID" != "null" ]]; then
    xbe_run do material-purchase-order-release-redemptions delete "$CREATED_REDEMPTION_ID"
    assert_failure
else
    skip "No created redemption for delete confirmation test"
fi

test_name "Delete release redemption with --confirm"
if [[ -n "$CREATED_REDEMPTION_ID" && "$CREATED_REDEMPTION_ID" != "null" ]]; then
    xbe_run do material-purchase-order-release-redemptions delete "$CREATED_REDEMPTION_ID" --confirm
    assert_success
else
    skip "No created redemption to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create release redemption without release fails"
xbe_run do material-purchase-order-release-redemptions create --ticket-number "T-$(unique_suffix)"
assert_failure

test_name "Create release redemption without ticket/material transaction fails"
if [[ -n "$RELEASE_ID" && "$RELEASE_ID" != "null" ]]; then
    xbe_run do material-purchase-order-release-redemptions create --release "$RELEASE_ID"
    assert_failure
else
    skip "No release ID available for missing ticket test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
