#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Applications
#
# Tests list, show, create, update, and delete operations for trucker applications.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_APPLICATION_ID=""
CREATED_BROKER_MEMBERSHIP_ID=""
WHOAMI_USER_ID=""
USER_ID=""
HAS_BROKER_MEMBERSHIP="false"

describe "Resource: trucker-applications"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Get current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    pass
else
    skip "Unable to resolve current user"
fi

USER_ID="${XBE_TEST_USER_ID:-$WHOAMI_USER_ID}"
if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
    fail "No user ID available for trucker application tests"
    run_tests
fi

test_name "Resolve broker for trucker applications"
if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
    echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
    pass
elif [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view memberships list --user "$USER_ID" --limit 50
    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(echo "$output" | jq -r '.[] | select(.organization_type=="Broker" or .organization_type=="brokers") | .organization_id' | head -n 1)
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            pass
        else
            fail "No broker membership found for current user"
        fi
    else
        fail "Failed to list memberships for broker lookup"
    fi
else
    skip "No user ID available for broker lookup"
fi

if [[ -z "$CREATED_BROKER_ID" || "$CREATED_BROKER_ID" == "null" ]]; then
    test_name "Create broker for trucker application tests"
    BROKER_NAME=$(unique_name "TruckerAppBroker")
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
        fail "Failed to create broker for trucker applications"
        run_tests
    fi
fi

test_name "Check broker membership for current user"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json view memberships list --user "$USER_ID" --limit 100
    if [[ $status -eq 0 ]]; then
        membership_count=$(echo "$output" | jq -r --arg broker "$CREATED_BROKER_ID" '[.[] | select((.organization_type=="Broker" or .organization_type=="brokers") and .organization_id==$broker)] | length')
        if [[ "$membership_count" -gt 0 ]]; then
            HAS_BROKER_MEMBERSHIP="true"
            pass
        else
            skip "No broker membership found for current user"
        fi
    else
        skip "Unable to list memberships to verify broker access"
    fi
else
    skip "No broker available for membership check"
fi

if [[ "$HAS_BROKER_MEMBERSHIP" != "true" && -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    test_name "Create broker membership for current user"
    xbe_json do memberships create --user "$USER_ID" --organization "Broker|$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_MEMBERSHIP_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_MEMBERSHIP_ID" && "$CREATED_BROKER_MEMBERSHIP_ID" != "null" ]]; then
            register_cleanup "memberships" "$CREATED_BROKER_MEMBERSHIP_ID"
            HAS_BROKER_MEMBERSHIP="true"
            pass
        else
            skip "Created broker membership but no ID returned"
        fi
    else
        skip "Unable to create broker membership; restricted fields will be skipped"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trucker application"
APP_NAME=$(unique_name "TruckerApplication")
APP_ADDRESS="123 Test Lane"

xbe_json do trucker-applications create \
    --name "$APP_NAME" \
    --company-address "$APP_ADDRESS" \
    --broker "$CREATED_BROKER_ID" \
    --user "$USER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_APPLICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_APPLICATION_ID" && "$CREATED_TRUCKER_APPLICATION_ID" != "null" ]]; then
        register_cleanup "trucker-applications" "$CREATED_TRUCKER_APPLICATION_ID"
        pass
    else
        fail "Created trucker application but no ID returned"
    fi
else
    fail "Failed to create trucker application"
fi

if [[ -z "$CREATED_TRUCKER_APPLICATION_ID" || "$CREATED_TRUCKER_APPLICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid trucker application ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests (Writable attributes)
# ============================================================================

test_name "Update trucker application"
UPDATED_NAME=$(unique_name "TruckerApplicationUpdated")
UPDATED_ADDRESS="456 Updated Lane"
UPDATE_ARGS=(do trucker-applications update "$CREATED_TRUCKER_APPLICATION_ID" \
    --name "$UPDATED_NAME" \
    --company-address "$UPDATED_ADDRESS" \
    --company-address-place-id "ChIJN1t_tDeuEmsRUsoyG83frY4" \
    --company-address-plus-code "849VCWC8+R9" \
    --skip-company-address-geocoding true \
    --has-union-drivers true \
    --estimated-trailer-capacity 12 \
    --notes "Updated notes" \
    --referral-code "REF-123")

if [[ "$HAS_BROKER_MEMBERSHIP" == "true" ]]; then
    UPDATE_ARGS+=(--status reviewing --via-dumptruckloadsdotcom true)
fi

xbe_json "${UPDATE_ARGS[@]}"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucker application"
xbe_json view trucker-applications show "$CREATED_TRUCKER_APPLICATION_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List trucker applications"
xbe_json view trucker-applications list --limit 25
assert_json_is_array

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by broker"
xbe_json view trucker-applications list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "Filter by status"
FILTER_STATUS="pending"
if [[ "$HAS_BROKER_MEMBERSHIP" == "true" ]]; then
    FILTER_STATUS="reviewing"
fi
xbe_json view trucker-applications list --status "$FILTER_STATUS" --limit 5
assert_success

test_name "Filter by search query"
xbe_json view trucker-applications list --q "$APP_NAME" --limit 5
assert_success

test_name "Filter by phone number"
xbe_json view trucker-applications list --phone-number "555" --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"column trucker_applications.phone_number does not exist"* ]]; then
        skip "Server does not support phone-number filter (missing column)"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "Filter by company address proximity"
xbe_json view trucker-applications list --company-address-within "0,0:50" --limit 5
assert_success

# ============================================================================
# Reset status to allow deletion if needed
# ============================================================================

if [[ "$HAS_BROKER_MEMBERSHIP" == "true" ]]; then
    test_name "Reset trucker application status to pending"
    xbe_json do trucker-applications update "$CREATED_TRUCKER_APPLICATION_ID" --status pending
    assert_success
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete trucker application"
xbe_run do trucker-applications delete "$CREATED_TRUCKER_APPLICATION_ID" --confirm
assert_success

run_tests
