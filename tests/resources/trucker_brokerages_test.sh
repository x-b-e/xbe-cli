#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Brokerages
#
# Tests create/update/delete operations and list filters for the
# trucker-brokerages resource.
#
# COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_BROKERED_TRUCKER_ID=""
UPDATED_TRUCKER_ID=""
UPDATED_BROKERED_TRUCKER_ID=""
CREATED_TRUCKER_BROKERAGE_ID=""

describe "Resource: trucker-brokerages"

# ============================================================================
# Prerequisites - Create broker and truckers
# ============================================================================

test_name "Create prerequisite broker for trucker brokerage tests"
BROKER_NAME=$(unique_name "TruckerBrokerageBroker")

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

test_name "Create brokering trucker"
TRUCKER_NAME=$(unique_name "BrokerageTrucker")
TRUCKER_ADDRESS="123 Brokerage St"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

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

test_name "Create brokered trucker"
BROKERED_TRUCKER_NAME=$(unique_name "BrokeredTrucker")
BROKERED_TRUCKER_ADDRESS="456 Brokered Ave"

xbe_json do truckers create \
    --name "$BROKERED_TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$BROKERED_TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_BROKERED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKERED_TRUCKER_ID" && "$CREATED_BROKERED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_BROKERED_TRUCKER_ID"
        pass
    else
        fail "Created brokered trucker but no ID returned"
        echo "Cannot continue without a brokered trucker"
        run_tests
    fi
else
    fail "Failed to create brokered trucker"
    echo "Cannot continue without a brokered trucker"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trucker brokerage"
xbe_json do trucker-brokerages create \
    --trucker "$CREATED_TRUCKER_ID" \
    --brokered-trucker "$CREATED_BROKERED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_BROKERAGE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_BROKERAGE_ID" && "$CREATED_TRUCKER_BROKERAGE_ID" != "null" ]]; then
        register_cleanup "trucker-brokerages" "$CREATED_TRUCKER_BROKERAGE_ID"
        pass
    else
        fail "Created trucker brokerage but no ID returned"
        echo "Cannot continue without a trucker brokerage"
        run_tests
    fi
else
    fail "Failed to create trucker brokerage"
    echo "Cannot continue without a trucker brokerage"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Create updated brokering trucker"
UPDATED_TRUCKER_NAME=$(unique_name "UpdatedTrucker")
UPDATED_TRUCKER_ADDRESS="789 Updated Blvd"

xbe_json do truckers create \
    --name "$UPDATED_TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$UPDATED_TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    UPDATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$UPDATED_TRUCKER_ID" && "$UPDATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$UPDATED_TRUCKER_ID"
        pass
    else
        fail "Created updated trucker but no ID returned"
        echo "Cannot continue without an updated trucker"
        run_tests
    fi
else
    fail "Failed to create updated trucker"
    echo "Cannot continue without an updated trucker"
    run_tests
fi

test_name "Create updated brokered trucker"
UPDATED_BROKERED_TRUCKER_NAME=$(unique_name "UpdatedBrokeredTrucker")
UPDATED_BROKERED_TRUCKER_ADDRESS="101 Updated Way"

xbe_json do truckers create \
    --name "$UPDATED_BROKERED_TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$UPDATED_BROKERED_TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    UPDATED_BROKERED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$UPDATED_BROKERED_TRUCKER_ID" && "$UPDATED_BROKERED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$UPDATED_BROKERED_TRUCKER_ID"
        pass
    else
        fail "Created updated brokered trucker but no ID returned"
        echo "Cannot continue without an updated brokered trucker"
        run_tests
    fi
else
    fail "Failed to create updated brokered trucker"
    echo "Cannot continue without an updated brokered trucker"
    run_tests
fi

test_name "Update trucker brokerage relationships"
xbe_json do trucker-brokerages update "$CREATED_TRUCKER_BROKERAGE_ID" \
    --trucker "$UPDATED_TRUCKER_ID" \
    --brokered-trucker "$UPDATED_BROKERED_TRUCKER_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucker brokerage"
xbe_json view trucker-brokerages show "$CREATED_TRUCKER_BROKERAGE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trucker brokerages"
xbe_json view trucker-brokerages list --limit 10
assert_success

test_name "List trucker brokerages returns array"
xbe_json view trucker-brokerages list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker brokerages"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trucker brokerages filtered by trucker"
xbe_json view trucker-brokerages list --trucker "$UPDATED_TRUCKER_ID" --limit 10
assert_success

test_name "List trucker brokerages filtered by brokered trucker"
xbe_json view trucker-brokerages list --brokered-trucker "$UPDATED_BROKERED_TRUCKER_ID" --limit 10
assert_success

NOW_ISO=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

test_name "List trucker brokerages with --created-at-min filter"
xbe_json view trucker-brokerages list --created-at-min "$NOW_ISO" --limit 10
assert_success

test_name "List trucker brokerages with --created-at-max filter"
xbe_json view trucker-brokerages list --created-at-max "$NOW_ISO" --limit 10
assert_success

test_name "List trucker brokerages with --updated-at-min filter"
xbe_json view trucker-brokerages list --updated-at-min "$NOW_ISO" --limit 10
assert_success

test_name "List trucker brokerages with --updated-at-max filter"
xbe_json view trucker-brokerages list --updated-at-max "$NOW_ISO" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
xbe_run do trucker-brokerages delete "$CREATED_TRUCKER_BROKERAGE_ID"
assert_failure

test_name "Delete trucker brokerage"
xbe_json do trucker-brokerages delete "$CREATED_TRUCKER_BROKERAGE_ID" --confirm
assert_success

run_tests
