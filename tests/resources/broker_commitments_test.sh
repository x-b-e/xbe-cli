#!/bin/bash
#
# XBE CLI Integration Tests: Broker Commitments
#
# Tests CRUD operations for the broker-commitments resource.
# Broker commitments link brokers with truckers for capacity commitments.
#
# COVERAGE: Create + update + delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRUCK_SCOPE_ID=""
CREATED_COMMITMENT_ID=""

SECOND_BROKER_ID=""
SECOND_TRUCKER_ID=""

# Optional override
TEST_TRUCK_SCOPE_ID="${XBE_TEST_TRUCK_SCOPE_ID:-}"

describe "Resource: broker-commitments"

# =========================================================================
# Prerequisites - Broker + Trucker
# =========================================================================

test_name "Create prerequisite broker for broker commitments"
BROKER_NAME=$(unique_name "BrokerCommitments")

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

test_name "Create prerequisite trucker for broker commitments"
TRUCKER_NAME=$(unique_name "BrokerCommitmentTrucker")
TRUCKER_ADDRESS="100 Commitment Lane, Haul City, HC 55555"

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

# =========================================================================
# Optional Truck Scope
# =========================================================================

test_name "Create truck scope for broker commitments (optional)"
if [[ -n "$TEST_TRUCK_SCOPE_ID" ]]; then
    CREATED_TRUCK_SCOPE_ID="$TEST_TRUCK_SCOPE_ID"
    echo "    Using XBE_TEST_TRUCK_SCOPE_ID: $CREATED_TRUCK_SCOPE_ID"
    pass
else
    xbe_json do truck-scopes create \
        --organization-type brokers \
        --organization-id "$CREATED_BROKER_ID" \
        --authorized-state-codes "IL,IN"

    if [[ $status -eq 0 ]]; then
        CREATED_TRUCK_SCOPE_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRUCK_SCOPE_ID" && "$CREATED_TRUCK_SCOPE_ID" != "null" ]]; then
            register_cleanup "truck-scopes" "$CREATED_TRUCK_SCOPE_ID"
            pass
        else
            fail "Created truck scope but no ID returned"
        fi
    else
        skip "Truck scope creation failed (continuing without truck scope)"
    fi
fi

# =========================================================================
# CREATE Tests
# =========================================================================

test_name "Create broker commitment with required fields"
TEST_LABEL=$(unique_name "Commitment")
TEST_NOTES="Initial commitment notes"

xbe_json do broker-commitments create \
    --status active \
    --broker "$CREATED_BROKER_ID" \
    --trucker "$CREATED_TRUCKER_ID" \
    --label "$TEST_LABEL" \
    --notes "$TEST_NOTES"

if [[ $status -eq 0 ]]; then
    CREATED_COMMITMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_COMMITMENT_ID" && "$CREATED_COMMITMENT_ID" != "null" ]]; then
        register_cleanup "broker-commitments" "$CREATED_COMMITMENT_ID"
        pass
    else
        fail "Created broker commitment but no ID returned"
    fi
else
    fail "Failed to create broker commitment: $output"
fi

# Only continue if we successfully created a commitment
if [[ -z "$CREATED_COMMITMENT_ID" || "$CREATED_COMMITMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker commitment ID"
    run_tests
fi

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show broker commitment"
xbe_json view broker-commitments show "$CREATED_COMMITMENT_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show broker commitment"
fi

# =========================================================================
# UPDATE Tests
# =========================================================================

test_name "Update broker commitment attributes"
NEW_LABEL=$(unique_name "CommitmentUpdated")

xbe_json do broker-commitments update "$CREATED_COMMITMENT_ID" \
    --status inactive \
    --label "$NEW_LABEL" \
    --notes "Updated notes"

assert_success

if [[ -n "$CREATED_TRUCK_SCOPE_ID" ]]; then
    test_name "Update broker commitment truck scope"
    xbe_json do broker-commitments update "$CREATED_COMMITMENT_ID" --truck-scope "$CREATED_TRUCK_SCOPE_ID"
    assert_success
fi

# Create second broker + trucker to test broker/trucker update

test_name "Create second broker for broker commitment update"
SECOND_BROKER_NAME=$(unique_name "BrokerCommitments2")

xbe_json do brokers create --name "$SECOND_BROKER_NAME"

if [[ $status -eq 0 ]]; then
    SECOND_BROKER_ID=$(json_get ".id")
    if [[ -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$SECOND_BROKER_ID"
        pass
    else
        fail "Created second broker but no ID returned"
    fi
else
    fail "Failed to create second broker"
fi

test_name "Create second trucker for broker commitment update"
SECOND_TRUCKER_NAME=$(unique_name "BrokerCommitmentTrucker2")
SECOND_TRUCKER_ADDRESS="200 Commitment Ave, Haul City, HC 55555"

xbe_json do truckers create \
    --name "$SECOND_TRUCKER_NAME" \
    --broker "$SECOND_BROKER_ID" \
    --company-address "$SECOND_TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    SECOND_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$SECOND_TRUCKER_ID" && "$SECOND_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$SECOND_TRUCKER_ID"
        pass
    else
        fail "Created second trucker but no ID returned"
    fi
else
    fail "Failed to create second trucker"
fi

if [[ -n "$SECOND_BROKER_ID" && -n "$SECOND_TRUCKER_ID" ]]; then
    test_name "Update broker commitment broker/trucker"
    xbe_json do broker-commitments update "$CREATED_COMMITMENT_ID" \
        --broker "$SECOND_BROKER_ID" \
        --trucker "$SECOND_TRUCKER_ID" \
        --truck-scope ""
    assert_success
else
    skip "Missing second broker or trucker ID for update test"
fi

# =========================================================================
# LIST Tests
# =========================================================================

test_name "List broker commitments"
xbe_json view broker-commitments list
assert_success

test_name "List broker commitments returns array"
xbe_json view broker-commitments list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker commitments"
fi

test_name "List broker commitments with --status filter"
xbe_json view broker-commitments list --status active
assert_success

test_name "List broker commitments with --broker-id filter"
xbe_json view broker-commitments list --broker-id "$CREATED_BROKER_ID"
assert_success

test_name "List broker commitments with --broker filter"
xbe_json view broker-commitments list --broker "$CREATED_BROKER_ID"
assert_success

test_name "List broker commitments with --trucker-id filter"
xbe_json view broker-commitments list --trucker-id "$CREATED_TRUCKER_ID"
assert_success

test_name "List broker commitments with --trucker filter"
xbe_json view broker-commitments list --trucker "$CREATED_TRUCKER_ID"
assert_success

test_name "List broker commitments with --created-at-min filter"
xbe_json view broker-commitments list --created-at-min "2000-01-01T00:00:00Z"
assert_success

test_name "List broker commitments with --updated-at-max filter"
xbe_json view broker-commitments list --updated-at-max "2100-01-01T00:00:00Z"
assert_success

test_name "List broker commitments with --limit"
xbe_json view broker-commitments list --limit 5
assert_success

test_name "List broker commitments with --offset"
xbe_json view broker-commitments list --limit 5 --offset 5
assert_success

test_name "List broker commitments with --sort"
xbe_json view broker-commitments list --sort status
assert_success

# =========================================================================
# DELETE Tests
# =========================================================================

test_name "Delete broker commitment"
xbe_run do broker-commitments delete "$CREATED_COMMITMENT_ID" --confirm
assert_success

# =========================================================================
# Error Cases
# =========================================================================

test_name "Create broker commitment without status fails"
xbe_json do broker-commitments create --broker "$CREATED_BROKER_ID" --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create broker commitment without broker fails"
xbe_json do broker-commitments create --status active --trucker "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create broker commitment without trucker fails"
xbe_json do broker-commitments create --status active --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update broker commitment with no fields fails"
xbe_json do broker-commitments update "$CREATED_COMMITMENT_ID"
assert_failure

test_name "Delete broker commitment without confirm fails"
xbe_run do broker-commitments delete "$CREATED_COMMITMENT_ID"
assert_failure

# =========================================================================
# Summary
# =========================================================================

run_tests
