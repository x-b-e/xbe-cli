#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Preloads
#
# Tests create/delete operations and list filters for the
# material_transaction_preloads resource.
#
# COMPLETE COVERAGE: Create, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRAILER_CLASSIFICATION_ID=""
CREATED_TRAILER_ID=""
CREATED_MT_ID=""
CREATED_PRELOAD_ID=""

describe "Resource: material-transaction-preloads"

# ==========================================================================
# Prerequisites - Create broker, trucker, trailer, material transaction
# ==========================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "MTPreloadBroker")

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
    fail "Failed to create broker"
    echo "Cannot continue without a broker"
    run_tests
fi

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "MTPreloadTrucker")

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "123 Preload Rd, Austin, TX"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

test_name "Fetch trailer classification"
xbe_json view trailer-classifications list --limit 1

if [[ $status -eq 0 ]]; then
    CREATED_TRAILER_CLASSIFICATION_ID=$(json_get ".[0].id")
    if [[ -n "$CREATED_TRAILER_CLASSIFICATION_ID" && "$CREATED_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        pass
    else
        fail "No trailer classification ID returned"
        echo "Cannot continue without a trailer classification"
        run_tests
    fi
else
    fail "Failed to list trailer classifications"
    echo "Cannot continue without a trailer classification"
    run_tests
fi

test_name "Create prerequisite trailer"
TRAILER_NUMBER="MTP-$(unique_suffix)"

xbe_json do trailers create \
    --number "$TRAILER_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID" \
    --trailer-classification "$CREATED_TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRAILER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRAILER_ID" && "$CREATED_TRAILER_ID" != "null" ]]; then
        register_cleanup "trailers" "$CREATED_TRAILER_ID"
        pass
    else
        fail "Created trailer but no ID returned"
        echo "Cannot continue without a trailer"
        run_tests
    fi
else
    fail "Failed to create trailer"
    echo "Cannot continue without a trailer"
    run_tests
fi

test_name "Create material transaction"
TICKET_NUM="MTP-T$(date +%s)"
TRANS_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

xbe_json do material-transactions create --ticket-number "$TICKET_NUM" --transaction-at "$TRANS_AT"

if [[ $status -eq 0 ]]; then
    CREATED_MT_ID=$(json_get ".id")
    if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
        register_cleanup "material-transactions" "$CREATED_MT_ID"
        pass
    else
        fail "Created material transaction but no ID returned"
        echo "Cannot continue without a material transaction"
        run_tests
    fi
else
    fail "Failed to create material transaction"
    echo "Cannot continue without a material transaction"
    run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create material transaction preload"
xbe_json do material-transaction-preloads create \
    --material-transaction "$CREATED_MT_ID" \
    --trailer "$CREATED_TRAILER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PRELOAD_ID=$(json_get ".id")
    if [[ -n "$CREATED_PRELOAD_ID" && "$CREATED_PRELOAD_ID" != "null" ]]; then
        register_cleanup "material-transaction-preloads" "$CREATED_PRELOAD_ID"
        pass
    else
        fail "Created preload but no ID returned"
        echo "Cannot continue without a material transaction preload"
        run_tests
    fi
else
    fail "Failed to create material transaction preload"
    echo "Cannot continue without a material transaction preload"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show material transaction preload"
xbe_json view material-transaction-preloads show "$CREATED_PRELOAD_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material transaction preloads"
xbe_json view material-transaction-preloads list --limit 10
assert_success

test_name "List material transaction preloads returns array"
xbe_json view material-transaction-preloads list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction preloads"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List material transaction preloads with --trailer filter"
xbe_json view material-transaction-preloads list --trailer "$CREATED_TRAILER_ID" --limit 10
assert_success

test_name "List material transaction preloads with --material-transaction filter"
xbe_json view material-transaction-preloads list --material-transaction "$CREATED_MT_ID" --limit 10
assert_success

test_name "List material transaction preloads with --preloaded-at-min filter"
xbe_json view material-transaction-preloads list --preloaded-at-min 2000-01-01T00:00:00Z --limit 10
assert_success

test_name "List material transaction preloads with --preloaded-at-max filter"
xbe_json view material-transaction-preloads list --preloaded-at-max 2100-01-01T00:00:00Z --limit 10
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete material transaction preload requires --confirm flag"
xbe_json do material-transaction-preloads delete "$CREATED_PRELOAD_ID"
assert_failure

test_name "Delete material transaction preload with --confirm"
xbe_json do material-transaction-preloads delete "$CREATED_PRELOAD_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
