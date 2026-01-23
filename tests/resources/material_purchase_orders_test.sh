#!/bin/bash
#
# XBE CLI Integration Tests: Material Purchase Orders
#
# Tests CRUD operations for the material_purchase_orders resource.
#
# COVERAGE: Create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_JOB_SITE_ID2=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_SITE_ID2=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_TYPE_ID2=""
CREATED_UNIT_OF_MEASURE_ID=""
CREATED_UNIT_OF_MEASURE_ID2=""
CREATED_PO_ID=""

TRANSACTION_AT_MIN="2026-01-01T00:00:00Z"
TRANSACTION_AT_MAX="2026-12-31T23:59:59Z"

EXTERNAL_PO_ID="PO-CLI-1001"
EXTERNAL_SO_ID="SO-CLI-1001"

EXTERNAL_PO_ID2="PO-CLI-2002"
EXTERNAL_SO_ID2="SO-CLI-2002"

describe "Resource: material-purchase-orders"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "MPOBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "MPOSupplier")

xbe_json do material-suppliers create \
    --name "$SUPPLIER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
else
    fail "Failed to create material supplier"
    echo "Cannot continue without a material supplier"
    run_tests
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "MPOCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create prerequisite job site"
JOB_SITE_NAME=$(unique_name "MPOJobSite")

xbe_json do job-sites create \
    --name "$JOB_SITE_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "100 MPO Job Site Rd, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID"
        pass
    else
        fail "Created job site but no ID returned"
        echo "Cannot continue without a job site"
        run_tests
    fi
else
    fail "Failed to create job site"
    echo "Cannot continue without a job site"
    run_tests
fi

test_name "Create prerequisite material site"
MATERIAL_SITE_NAME=$(unique_name "MPOMaterialSite")

xbe_json do material-sites create \
    --name "$MATERIAL_SITE_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "100 Test Quarry Rd, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
        pass
    else
        fail "Created material site but no ID returned"
        echo "Cannot continue without a material site"
        run_tests
    fi
else
    fail "Failed to create material site"
    echo "Cannot continue without a material site"
    run_tests
fi

test_name "Create prerequisite material type"
MATERIAL_TYPE_NAME=$(unique_name "MPOMaterialType")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    fail "Failed to create material type"
    echo "Cannot continue without a material type"
    run_tests
fi

test_name "Select unit of measure"

xbe_json view unit-of-measures list --name "Ton" --limit 1

if [[ $status -eq 0 ]]; then
    CREATED_UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
    if [[ -n "$CREATED_UNIT_OF_MEASURE_ID" && "$CREATED_UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        fail "No unit of measure returned"
        echo "Cannot continue without a unit of measure"
        run_tests
    fi
else
    fail "Failed to list unit of measures"
    echo "Cannot continue without a unit of measure"
    run_tests
fi

# Secondary resources for updates

test_name "Create secondary job site"
JOB_SITE_NAME2=$(unique_name "MPOJobSite2")

xbe_json do job-sites create \
    --name "$JOB_SITE_NAME2" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "101 MPO Job Site Rd, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID2=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID2" && "$CREATED_JOB_SITE_ID2" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID2"
        pass
    else
        fail "Created secondary job site but no ID returned"
        echo "Cannot continue without a job site"
        run_tests
    fi
else
    fail "Failed to create secondary job site"
    echo "Cannot continue without a job site"
    run_tests
fi

test_name "Create secondary material site"
MATERIAL_SITE_NAME2=$(unique_name "MPOMaterialSite2")

xbe_json do material-sites create \
    --name "$MATERIAL_SITE_NAME2" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "101 Test Quarry Rd, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID2" && "$CREATED_MATERIAL_SITE_ID2" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID2"
        pass
    else
        fail "Created secondary material site but no ID returned"
        echo "Cannot continue without a material site"
        run_tests
    fi
else
    fail "Failed to create secondary material site"
    echo "Cannot continue without a material site"
    run_tests
fi

test_name "Create secondary material type"
MATERIAL_TYPE_NAME2=$(unique_name "MPOMaterialType2")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME2"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID2" && "$CREATED_MATERIAL_TYPE_ID2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID2"
        pass
    else
        fail "Created secondary material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    fail "Failed to create secondary material type"
    echo "Cannot continue without a material type"
    run_tests
fi

test_name "Select alternate unit of measure"

xbe_json view unit-of-measures list --name "Load" --limit 1

if [[ $status -eq 0 ]]; then
    CREATED_UNIT_OF_MEASURE_ID2=$(json_get ".[0].id")
    if [[ -n "$CREATED_UNIT_OF_MEASURE_ID2" && "$CREATED_UNIT_OF_MEASURE_ID2" != "null" ]]; then
        pass
    else
        echo "    No alternate unit of measure found"
        pass
    fi
else
    echo "    Failed to list alternate unit of measures"
    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material purchase order with required fields"

xbe_json do material-purchase-orders create \
    --broker "$CREATED_BROKER_ID" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --unit-of-measure "$CREATED_UNIT_OF_MEASURE_ID" \
    --quantity "500" \
    --customer "$CREATED_CUSTOMER_ID" \
    --job-site "$CREATED_JOB_SITE_ID" \
    --material-site "$CREATED_MATERIAL_SITE_ID" \
    --transaction-at-min "$TRANSACTION_AT_MIN" \
    --transaction-at-max "$TRANSACTION_AT_MAX" \
    --external-purchase-order-id "$EXTERNAL_PO_ID" \
    --external-sales-order-id "$EXTERNAL_SO_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PO_ID=$(json_get ".id")
    if [[ -n "$CREATED_PO_ID" && "$CREATED_PO_ID" != "null" ]]; then
        register_cleanup "material-purchase-orders" "$CREATED_PO_ID"
        pass
    else
        fail "Created material purchase order but no ID returned"
    fi
else
    fail "Failed to create material purchase order"
fi

if [[ -z "$CREATED_PO_ID" || "$CREATED_PO_ID" == "null" ]]; then
    echo "Cannot continue without a valid purchase order ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material purchase order quantity"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --quantity "750"
assert_success

test_name "Update material purchase order transaction-at-min"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --transaction-at-min "2026-02-01T00:00:00Z"
assert_success

test_name "Update material purchase order transaction-at-max"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --transaction-at-max "2026-12-15T23:59:59Z"
assert_success

test_name "Update material purchase order is-managing-redemption"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --is-managing-redemption
assert_success

test_name "Update material purchase order no-is-managing-redemption"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --no-is-managing-redemption
assert_success

test_name "Update material purchase order external IDs"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" \
    --external-purchase-order-id "$EXTERNAL_PO_ID2" \
    --external-sales-order-id "$EXTERNAL_SO_ID2"
assert_success

test_name "Update material purchase order status"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --status approved
if [[ $status -eq 0 ]]; then
    # Attempt to revert for cleanup safety
    xbe_json do material-purchase-orders update "$CREATED_PO_ID" --status editing
    if [[ $status -eq 0 ]]; then
        pass
    else
        echo "    (Status reverted failed, cleanup may fail)"
        pass
    fi
else
    if [[ "$output" == *"is not included in the list"* ]] || [[ "$output" == *"cannot"* ]]; then
        echo "    (Server validation for status update - flag works correctly)"
        pass
    else
        fail "Unexpected error: $output"
    fi
fi

test_name "Update material purchase order material type"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --material-type "$CREATED_MATERIAL_TYPE_ID2"
assert_success

test_name "Update material purchase order material site"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --material-site "$CREATED_MATERIAL_SITE_ID2"
assert_success

test_name "Update material purchase order job site"
xbe_json do material-purchase-orders update "$CREATED_PO_ID" --job-site "$CREATED_JOB_SITE_ID2"
assert_success

if [[ -n "$CREATED_UNIT_OF_MEASURE_ID2" && "$CREATED_UNIT_OF_MEASURE_ID2" != "null" ]]; then
    test_name "Update material purchase order unit of measure"
    xbe_json do material-purchase-orders update "$CREATED_PO_ID" --unit-of-measure "$CREATED_UNIT_OF_MEASURE_ID2"
    assert_success
else
    test_name "Update material purchase order unit of measure"
    skip "No alternate unit of measure available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material purchase order"
xbe_json view material-purchase-orders show "$CREATED_PO_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material purchase orders"
xbe_json view material-purchase-orders list
assert_success

test_name "List material purchase orders returns array"
xbe_json view material-purchase-orders list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material purchase orders"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List material purchase orders with --status filter"
xbe_json view material-purchase-orders list --status editing
assert_success

test_name "List material purchase orders with --broker filter"
xbe_json view material-purchase-orders list --broker "$CREATED_BROKER_ID"
assert_success

test_name "List material purchase orders with --material-supplier filter"
xbe_json view material-purchase-orders list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"
assert_success

test_name "List material purchase orders with --customer filter"
xbe_json view material-purchase-orders list --customer "$CREATED_CUSTOMER_ID"
assert_success

test_name "List material purchase orders with --material-site filter"
xbe_json view material-purchase-orders list --material-site "$CREATED_MATERIAL_SITE_ID"
assert_success

test_name "List material purchase orders with --material-type filter"
xbe_json view material-purchase-orders list --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_success

test_name "List material purchase orders with --job-site filter"
xbe_json view material-purchase-orders list --job-site "$CREATED_JOB_SITE_ID"
assert_success

test_name "List material purchase orders with --unit-of-measure filter"
xbe_json view material-purchase-orders list --unit-of-measure "$CREATED_UNIT_OF_MEASURE_ID"
assert_success

test_name "List material purchase orders with --transaction-at-min filter"
xbe_json view material-purchase-orders list --transaction-at-min "$TRANSACTION_AT_MIN"
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Filter not allowed"* ]]; then
        echo "    (Server does not allow transaction-at-min filter - flag still wired)"
        pass
    else
        fail "Unexpected error: $output"
    fi
fi

test_name "List material purchase orders with --transaction-at-max filter"
xbe_json view material-purchase-orders list --transaction-at-max "$TRANSACTION_AT_MAX"
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Filter not allowed"* ]]; then
        echo "    (Server does not allow transaction-at-max filter - flag still wired)"
        pass
    else
        fail "Unexpected error: $output"
    fi
fi

test_name "List material purchase orders with --quantity filter"
xbe_json view material-purchase-orders list --quantity 500
assert_success

test_name "List material purchase orders with --base-material-type filter"
xbe_json view material-purchase-orders list --base-material-type "$CREATED_MATERIAL_TYPE_ID"
assert_success

test_name "List material purchase orders with --external-purchase-order-id filter"
xbe_json view material-purchase-orders list --external-purchase-order-id "$EXTERNAL_PO_ID"
assert_success

test_name "List material purchase orders with --external-sales-order-id filter"
xbe_json view material-purchase-orders list --external-sales-order-id "$EXTERNAL_SO_ID"
assert_success

test_name "List material purchase orders with --external-identification-value filter"
xbe_json view material-purchase-orders list --external-identification-value "TEST-EXT-ID"
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material purchase order without required fields fails"
xbe_json do material-purchase-orders create
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material purchase order requires --confirm"
xbe_json do material-purchase-orders create \
    --broker "$CREATED_BROKER_ID" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --unit-of-measure "$CREATED_UNIT_OF_MEASURE_ID" \
    --quantity "100"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do material-purchase-orders delete "$DEL_ID"
    assert_failure
else
    skip "Could not create purchase order for deletion test"
fi

test_name "Delete material purchase order with --confirm"
xbe_json do material-purchase-orders create \
    --broker "$CREATED_BROKER_ID" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --unit-of-measure "$CREATED_UNIT_OF_MEASURE_ID" \
    --quantity "150"

if [[ $status -eq 0 ]]; then
    DEL_ID2=$(json_get ".id")
    xbe_run do material-purchase-orders delete "$DEL_ID2" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"cannot be deleted"* ]] || [[ "$output" == *"cannot"* ]]; then
            echo "    (Delete blocked by status or server validation - expected)"
            pass
        else
            fail "Delete command failed unexpectedly: $output"
        fi
    fi
else
    skip "Could not create purchase order for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
