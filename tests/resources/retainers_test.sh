#!/bin/bash
#
# XBE CLI Integration Tests: Retainers
#
# Tests CRUD operations for the retainers resource.
#
# COVERAGE: All create/update attributes + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRUCKER_ID=""
CREATED_BROKER_RETAINER_ID=""
CREATED_CUSTOMER_RETAINER_ID=""

describe "Resource: retainers"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "RetainerBroker")
xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "RetainerCustomer")
xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        run_tests
    fi
fi

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "RetainerTrucker")
TRUCKER_ADDRESS="100 Retainer Way, Haul City, HC 55555"
xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS" \
    --skip-company-address-geocoding true

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
        CREATED_TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
        echo "    Using XBE_TEST_TRUCKER_ID: $CREATED_TRUCKER_ID"
        pass
    else
        fail "Failed to create trucker and XBE_TEST_TRUCKER_ID not set"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker retainer with required fields"
xbe_json do retainers create \
    --buyer "Broker|$CREATED_BROKER_ID" \
    --seller "Trucker|$CREATED_TRUCKER_ID" \
    --status editing \
    --maximum-expected-daily-hours 8 \
    --maximum-travel-minutes 60 \
    --billable-travel-minutes-per-travel-mile 1.5

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_RETAINER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_RETAINER_ID" && "$CREATED_BROKER_RETAINER_ID" != "null" ]]; then
        register_cleanup "retainers" "$CREATED_BROKER_RETAINER_ID"
        pass
    else
        fail "Created broker retainer but no ID returned"
    fi
else
    fail "Failed to create broker retainer"
fi

test_name "Create customer retainer"
xbe_json do retainers create \
    --buyer "Customer|$CREATED_CUSTOMER_ID" \
    --seller "Broker|$CREATED_BROKER_ID" \
    --status editing

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_RETAINER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_RETAINER_ID" && "$CREATED_CUSTOMER_RETAINER_ID" != "null" ]]; then
        register_cleanup "retainers" "$CREATED_CUSTOMER_RETAINER_ID"
        pass
    else
        fail "Created customer retainer but no ID returned"
    fi
else
    fail "Failed to create customer retainer"
fi

if [[ -z "$CREATED_BROKER_RETAINER_ID" || "$CREATED_BROKER_RETAINER_ID" == "null" ]]; then
    echo "Cannot continue without a valid retainer ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update retainer status"
xbe_json do retainers update "$CREATED_BROKER_RETAINER_ID" --status editing
assert_success

test_name "Update retainer terminated-on"
xbe_json do retainers update "$CREATED_BROKER_RETAINER_ID" --terminated-on "2025-01-15"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".terminated_on" "2025-01-15"
else
    fail "Failed to update terminated-on"
fi

test_name "Update maximum expected daily hours"
xbe_json do retainers update "$CREATED_BROKER_RETAINER_ID" --maximum-expected-daily-hours 10
if [[ $status -eq 0 ]]; then
    value=$(json_get ".maximum_expected_daily_hours")
    if [[ "$value" == "10" || "$value" == "10.0" ]]; then
        pass
    else
        fail "Expected maximum_expected_daily_hours to be 10 or 10.0, got '$value'"
    fi
else
    fail "Failed to update maximum expected daily hours"
fi

test_name "Update maximum travel minutes"
xbe_json do retainers update "$CREATED_BROKER_RETAINER_ID" --maximum-travel-minutes 90
if [[ $status -eq 0 ]]; then
    assert_json_equals ".maximum_travel_minutes" "90"
else
    fail "Failed to update maximum travel minutes"
fi

test_name "Update billable travel minutes per travel mile"
xbe_json do retainers update "$CREATED_BROKER_RETAINER_ID" --billable-travel-minutes-per-travel-mile 2.25
if [[ $status -eq 0 ]]; then
    assert_json_equals ".billable_travel_minutes_per_travel_mile" "2.25"
else
    fail "Failed to update billable travel minutes per travel mile"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List retainers"
xbe_json view retainers list --limit 5
assert_success

test_name "List retainers returns array"
xbe_json view retainers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list retainers"
fi

test_name "List retainers with --status filter"
xbe_json view retainers list --status editing --limit 5
assert_success

test_name "List retainers with --buyer filter"
xbe_json view retainers list --buyer "Broker|$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List retainers with --seller filter"
xbe_json view retainers list --seller "Trucker|$CREATED_TRUCKER_ID" --limit 5
assert_success

test_name "List retainers with --created-at-min filter"
xbe_json view retainers list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List retainers with --created-at-max filter"
xbe_json view retainers list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List retainers with --updated-at-min filter"
xbe_json view retainers list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List retainers with --updated-at-max filter"
xbe_json view retainers list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show retainer"
xbe_json view retainers show "$CREATED_BROKER_RETAINER_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker retainer"
xbe_run do retainers delete "$CREATED_BROKER_RETAINER_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
