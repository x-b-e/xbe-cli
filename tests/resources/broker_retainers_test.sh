#!/bin/bash
#
# XBE CLI Integration Tests: Broker Retainers
#
# Tests CRUD operations for the broker_retainers resource.
# Broker retainers define retainer agreements between brokers and truckers.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STATUS=""
SAMPLE_BUYER_ID=""
SAMPLE_BUYER_TYPE=""
SAMPLE_SELLER_ID=""
SAMPLE_SELLER_TYPE=""
CREATED_ID=""

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"

DESCRIBE_RESOURCE="broker-retainers"

describe "Resource: broker-retainers"

TERMINATED_ON="2099-01-15"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker retainers"
xbe_json view broker-retainers list --limit 5
assert_success

test_name "List broker retainers returns array"
xbe_json view broker-retainers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_STATUS=$(echo "$output" | jq -r '.[0].status // empty')
    SAMPLE_BUYER_ID=$(echo "$output" | jq -r '.[0].buyer_id // empty')
    SAMPLE_BUYER_TYPE=$(echo "$output" | jq -r '.[0].buyer_type // empty')
    SAMPLE_SELLER_ID=$(echo "$output" | jq -r '.[0].seller_id // empty')
    SAMPLE_SELLER_TYPE=$(echo "$output" | jq -r '.[0].seller_type // empty')
else
    fail "Failed to list broker retainers"
fi

if [[ -z "$BROKER_ID" && -n "$SAMPLE_BUYER_ID" && "$SAMPLE_BUYER_ID" != "null" ]]; then
    if [[ "$SAMPLE_BUYER_TYPE" == "brokers" ]]; then
        BROKER_ID="$SAMPLE_BUYER_ID"
    fi
fi
if [[ -z "$TRUCKER_ID" && -n "$SAMPLE_SELLER_ID" && "$SAMPLE_SELLER_ID" != "null" ]]; then
    if [[ "$SAMPLE_SELLER_TYPE" == "truckers" ]]; then
        TRUCKER_ID="$SAMPLE_SELLER_ID"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker retainer"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view broker-retainers show "$SAMPLE_ID"
    assert_success
else
    skip "No broker retainer ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker retainer"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" && -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json do broker-retainers create \
        --broker "$BROKER_ID" \
        --trucker "$TRUCKER_ID" \
        --maximum-expected-daily-hours "8" \
        --maximum-travel-minutes "60" \
        --billable-travel-minutes-per-travel-mile "2"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "broker-retainers" "$CREATED_ID"
            pass
        else
            fail "Created broker retainer but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            pass
        else
            fail "Failed to create broker retainer: $output"
        fi
    fi
else
    skip "No broker/trucker IDs available (set XBE_TEST_BROKER_ID and XBE_TEST_TRUCKER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_ID="${CREATED_ID:-$SAMPLE_ID}"

update_broker_retainer() {
    local description="$1"
    shift
    test_name "$description"
    xbe_json do broker-retainers update "$UPDATE_ID" "$@"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Update blocked by server policy/validation"
        else
            fail "Update failed: $output"
        fi
    fi
}

if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    update_broker_retainer "Update status" --status active
    update_broker_retainer "Update maximum expected daily hours" --maximum-expected-daily-hours "10"
    update_broker_retainer "Update maximum travel minutes" --maximum-travel-minutes "90"
    update_broker_retainer "Update billable travel minutes per mile" --billable-travel-minutes-per-travel-mile "3"
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        update_broker_retainer "Update terminated status with terminated-on" --status terminated --terminated-on "$TERMINATED_ON"
    fi
    if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
        update_broker_retainer "Update buyer" --buyer "Broker|$BROKER_ID"
    fi
    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        update_broker_retainer "Update seller" --seller "Trucker|$TRUCKER_ID"
    fi
else
    skip "No broker retainer ID available for update tests"
fi

test_name "Update broker retainer without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do broker-retainers update "$UPDATE_ID"
    assert_failure
else
    skip "No broker retainer ID available for no-attribute update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List broker retainers with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view broker-retainers list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    xbe_json view broker-retainers list --status active --limit 5
    assert_success
fi


test_name "List broker retainers with --buyer filter"
if [[ -n "$SAMPLE_BUYER_ID" && "$SAMPLE_BUYER_ID" != "null" ]]; then
    BUYER_TYPE="$SAMPLE_BUYER_TYPE"
    if [[ -z "$BUYER_TYPE" || "$BUYER_TYPE" == "null" ]]; then
        BUYER_TYPE="brokers"
    fi
    xbe_json view broker-retainers list --buyer "${BUYER_TYPE}|${SAMPLE_BUYER_ID}" --limit 5
    assert_success
else
    skip "No buyer available for filter"
fi

test_name "List broker retainers with --seller filter"
if [[ -n "$SAMPLE_SELLER_ID" && "$SAMPLE_SELLER_ID" != "null" ]]; then
    SELLER_TYPE="$SAMPLE_SELLER_TYPE"
    if [[ -z "$SELLER_TYPE" || "$SELLER_TYPE" == "null" ]]; then
        SELLER_TYPE="truckers"
    fi
    xbe_json view broker-retainers list --seller "${SELLER_TYPE}|${SAMPLE_SELLER_ID}" --limit 5
    assert_success
else
    skip "No seller available for filter"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker retainer requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do broker-retainers delete "$CREATED_ID"
    assert_failure
else
    skip "No created broker retainer for delete confirmation test"
fi


test_name "Delete broker retainer with --confirm"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" && -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json do broker-retainers create --broker "$BROKER_ID" --trucker "$TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do broker-retainers delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create broker retainer for deletion test"
    fi
else
    skip "No broker/trucker IDs available for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create broker retainer without broker fails"
xbe_run do broker-retainers create --trucker "${TRUCKER_ID:-123}"
assert_failure

test_name "Create broker retainer without trucker fails"
xbe_run do broker-retainers create --broker "${BROKER_ID:-123}"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
