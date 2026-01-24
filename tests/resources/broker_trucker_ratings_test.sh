#!/bin/bash
#
# XBE CLI Integration Tests: Broker Trucker Ratings
#
# Tests CRUD operations for the broker_trucker_ratings resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_RATING_ID=""

describe "Resource: broker_trucker_ratings"

# ============================================================================
# Prerequisites - Create broker and trucker
# ============================================================================

test_name "Create prerequisite broker for broker trucker rating tests"
BROKER_NAME=$(unique_name "BTRBroker")

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

test_name "Create prerequisite trucker for broker trucker rating tests"
TRUCKER_NAME=$(unique_name "BTRTrucker")
TRUCKER_ADDRESS="100 Rating Way, Haul City, HC 55555"

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
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
        CREATED_TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
        echo "    Using XBE_TEST_TRUCKER_ID: $CREATED_TRUCKER_ID"
        pass
    else
        fail "Failed to create trucker and XBE_TEST_TRUCKER_ID not set"
        echo "Cannot continue without a trucker"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker trucker rating with required fields"
xbe_json do broker-trucker-ratings create \
    --broker "$CREATED_BROKER_ID" \
    --trucker "$CREATED_TRUCKER_ID" \
    --rating 5

if [[ $status -eq 0 ]]; then
    CREATED_RATING_ID=$(json_get ".id")
    if [[ -n "$CREATED_RATING_ID" && "$CREATED_RATING_ID" != "null" ]]; then
        register_cleanup "broker-trucker-ratings" "$CREATED_RATING_ID"
        pass
    else
        fail "Created broker trucker rating but no ID returned"
    fi
else
    fail "Failed to create broker trucker rating"
fi

if [[ -z "$CREATED_RATING_ID" || "$CREATED_RATING_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker trucker rating ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker trucker rating"
xbe_json view broker-trucker-ratings show "$CREATED_RATING_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update broker trucker rating"
xbe_json do broker-trucker-ratings update "$CREATED_RATING_ID" --rating 4
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker trucker ratings"
xbe_json view broker-trucker-ratings list --limit 5
assert_success

test_name "List broker trucker ratings returns array"
xbe_json view broker-trucker-ratings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker trucker ratings"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List broker trucker ratings with --broker filter"
xbe_json view broker-trucker-ratings list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List broker trucker ratings with --trucker filter"
xbe_json view broker-trucker-ratings list --trucker "$CREATED_TRUCKER_ID" --limit 5
assert_success

test_name "List broker trucker ratings with --rating filter"
xbe_json view broker-trucker-ratings list --rating 4 --limit 5
assert_success

test_name "List broker trucker ratings with --created-at-min filter"
xbe_json view broker-trucker-ratings list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker trucker ratings with --created-at-max filter"
xbe_json view broker-trucker-ratings list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker trucker ratings with --is-created-at filter"
xbe_json view broker-trucker-ratings list --is-created-at true --limit 5
assert_success

test_name "List broker trucker ratings with --updated-at-min filter"
xbe_json view broker-trucker-ratings list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker trucker ratings with --updated-at-max filter"
xbe_json view broker-trucker-ratings list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker trucker ratings with --is-updated-at filter"
xbe_json view broker-trucker-ratings list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broker trucker ratings with --limit"
xbe_json view broker-trucker-ratings list --limit 3
assert_success

test_name "List broker trucker ratings with --offset"
xbe_json view broker-trucker-ratings list --limit 3 --offset 3
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create broker trucker rating without broker fails"
xbe_json do broker-trucker-ratings create --trucker "$CREATED_TRUCKER_ID" --rating 5
assert_failure

test_name "Create broker trucker rating without trucker fails"
xbe_json do broker-trucker-ratings create --broker "$CREATED_BROKER_ID" --rating 5
assert_failure

test_name "Create broker trucker rating without rating fails"
xbe_json do broker-trucker-ratings create --broker "$CREATED_BROKER_ID" --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Update broker trucker rating without any fields fails"
xbe_json do broker-trucker-ratings update "$CREATED_RATING_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker trucker rating requires --confirm flag"
xbe_run do broker-trucker-ratings delete "$CREATED_RATING_ID"
assert_failure

test_name "Delete broker trucker rating with --confirm"
xbe_run do broker-trucker-ratings delete "$CREATED_RATING_ID" --confirm
assert_success

run_tests
