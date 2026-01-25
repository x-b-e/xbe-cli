#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Referrals
#
# Tests create/update/delete operations and list filters for the
# trucker-referrals resource.
#
# COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
UPDATED_TRUCKER_ID=""
CREATED_USER_ID=""
UPDATED_USER_ID=""
CREATED_TRUCKER_REFERRAL_ID=""

describe "Resource: trucker-referrals"

# ============================================================================
# Prerequisites - Create broker, truckers, users
# ============================================================================

test_name "Create prerequisite broker for trucker referral tests"
BROKER_NAME=$(unique_name "TruckerReferralBroker")

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

test_name "Create referred trucker"
TRUCKER_NAME=$(unique_name "ReferralTrucker")
TRUCKER_ADDRESS="123 Referral St"

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

test_name "Create updated trucker"
UPDATED_TRUCKER_NAME=$(unique_name "UpdatedReferralTrucker")
UPDATED_TRUCKER_ADDRESS="456 Updated Referral Ave"

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

test_name "Create referring user"
REFERRAL_USER_EMAIL=$(unique_email)
REFERRAL_USER_NAME=$(unique_name "ReferralUser")

xbe_json do users create \
    --email "$REFERRAL_USER_EMAIL" \
    --name "$REFERRAL_USER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

test_name "Create updated referring user"
UPDATED_USER_EMAIL=$(unique_email)
UPDATED_USER_NAME=$(unique_name "UpdatedReferralUser")

xbe_json do users create \
    --email "$UPDATED_USER_EMAIL" \
    --name "$UPDATED_USER_NAME"

if [[ $status -eq 0 ]]; then
    UPDATED_USER_ID=$(json_get ".id")
    if [[ -n "$UPDATED_USER_ID" && "$UPDATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created updated user but no ID returned"
        echo "Cannot continue without an updated user"
        run_tests
    fi
else
    fail "Failed to create updated user"
    echo "Cannot continue without an updated user"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trucker referral"
REFERRED_ON=$(date -u +"%Y-%m-%d")

xbe_json do trucker-referrals create \
    --trucker "$CREATED_TRUCKER_ID" \
    --user "$CREATED_USER_ID" \
    --notes "Initial referral" \
    --referred-on "$REFERRED_ON" \
    --trucker-first-shift-bonus-amount "250.00" \
    --truck-first-shift-bonus-amount "100.00"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_REFERRAL_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_REFERRAL_ID" && "$CREATED_TRUCKER_REFERRAL_ID" != "null" ]]; then
        register_cleanup "trucker-referrals" "$CREATED_TRUCKER_REFERRAL_ID"
        pass
    else
        fail "Created trucker referral but no ID returned"
        echo "Cannot continue without a trucker referral"
        run_tests
    fi
else
    fail "Failed to create trucker referral"
    echo "Cannot continue without a trucker referral"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update trucker referral"
UPDATED_REFERRED_ON="$REFERRED_ON"

xbe_json do trucker-referrals update "$CREATED_TRUCKER_REFERRAL_ID" \
    --trucker "$UPDATED_TRUCKER_ID" \
    --user "$UPDATED_USER_ID" \
    --notes "Updated referral" \
    --referred-on "$UPDATED_REFERRED_ON" \
    --trucker-first-shift-bonus-amount "300.00" \
    --truck-first-shift-bonus-amount "150.00"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucker referral"
xbe_json view trucker-referrals show "$CREATED_TRUCKER_REFERRAL_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trucker referrals"
xbe_json view trucker-referrals list --limit 10
assert_success

test_name "List trucker referrals returns array"
xbe_json view trucker-referrals list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker referrals"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trucker referrals filtered by trucker"
xbe_json view trucker-referrals list --trucker "$UPDATED_TRUCKER_ID" --limit 10
assert_success

test_name "List trucker referrals filtered by user"
xbe_json view trucker-referrals list --user "$UPDATED_USER_ID" --limit 10
assert_success

test_name "List trucker referrals filtered by broker"
xbe_json view trucker-referrals list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List trucker referrals filtered by referred-on"
xbe_json view trucker-referrals list --referred-on "$REFERRED_ON" --limit 10
assert_success

test_name "List trucker referrals with --referred-on-min filter"
xbe_json view trucker-referrals list --referred-on-min "$REFERRED_ON" --limit 10
assert_success

test_name "List trucker referrals with --referred-on-max filter"
xbe_json view trucker-referrals list --referred-on-max "$REFERRED_ON" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
xbe_run do trucker-referrals delete "$CREATED_TRUCKER_REFERRAL_ID"
assert_failure

test_name "Delete trucker referral"
xbe_json do trucker-referrals delete "$CREATED_TRUCKER_REFERRAL_ID" --confirm
assert_success

run_tests
