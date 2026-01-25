#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Referral Codes
#
# Tests CRUD operations for the trucker_referral_codes resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
SECOND_BROKER_ID=""
CREATED_REFERRAL_CODE_ID=""
CREATED_BROKER_MEMBERSHIP_ID=""
SECOND_BROKER_MEMBERSHIP_ID=""
USER_ID=""
HAS_BROKER_MEMBERSHIP="false"
HAS_SECOND_BROKER_MEMBERSHIP="false"
MEMBERSHIP_BROKER_IDS=""

describe "Resource: trucker_referral_codes"

# ============================================================================
# Prerequisites - Resolve broker + membership
# ============================================================================

test_name "Get current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    USER_ID=$(json_get ".id")
    pass
else
    skip "Unable to resolve current user"
fi

if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view memberships list --user "$USER_ID" --limit 200
    if [[ $status -eq 0 ]]; then
        MEMBERSHIP_BROKER_IDS=$(echo "$output" | jq -r '.[] | select(.organization_type=="Broker" or .organization_type=="brokers") | .organization_id')
    fi
fi

test_name "Resolve broker for trucker referral code tests"
if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
    echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
    pass
elif [[ -n "$MEMBERSHIP_BROKER_IDS" ]]; then
    CREATED_BROKER_ID=$(echo "$MEMBERSHIP_BROKER_IDS" | head -n 1)
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        HAS_BROKER_MEMBERSHIP="true"
        pass
    else
        fail "No broker membership found for current user"
    fi
else
    skip "No broker membership found for current user"
fi

if [[ -z "$CREATED_BROKER_ID" || "$CREATED_BROKER_ID" == "null" ]]; then
    test_name "Create prerequisite broker for trucker referral code tests"
    BROKER_NAME=$(unique_name "TRCBroker")

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
        fail "Failed to create broker for trucker referral code tests"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Ensure broker membership for current user"
if [[ "$HAS_BROKER_MEMBERSHIP" == "true" ]]; then
    pass
elif [[ -n "$USER_ID" && "$USER_ID" != "null" && -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    if [[ -n "$MEMBERSHIP_BROKER_IDS" ]]; then
        if echo "$MEMBERSHIP_BROKER_IDS" | grep -q "^${CREATED_BROKER_ID}$"; then
            HAS_BROKER_MEMBERSHIP="true"
            pass
        else
            xbe_json do memberships create --user "$USER_ID" --organization "Broker|$CREATED_BROKER_ID"
            if [[ $status -eq 0 ]]; then
                CREATED_BROKER_MEMBERSHIP_ID=$(json_get ".id")
                if [[ -n "$CREATED_BROKER_MEMBERSHIP_ID" && "$CREATED_BROKER_MEMBERSHIP_ID" != "null" ]]; then
                    register_cleanup "memberships" "$CREATED_BROKER_MEMBERSHIP_ID"
                    HAS_BROKER_MEMBERSHIP="true"
                    pass
                else
                    fail "Created broker membership but no ID returned"
                fi
            else
                fail "Unable to create broker membership for current user"
            fi
        fi
    else
        xbe_json do memberships create --user "$USER_ID" --organization "Broker|$CREATED_BROKER_ID"
        if [[ $status -eq 0 ]]; then
            CREATED_BROKER_MEMBERSHIP_ID=$(json_get ".id")
            if [[ -n "$CREATED_BROKER_MEMBERSHIP_ID" && "$CREATED_BROKER_MEMBERSHIP_ID" != "null" ]]; then
                register_cleanup "memberships" "$CREATED_BROKER_MEMBERSHIP_ID"
                HAS_BROKER_MEMBERSHIP="true"
                pass
            else
                fail "Created broker membership but no ID returned"
            fi
        else
            fail "Unable to create broker membership for current user"
        fi
    fi
else
    fail "No user available to verify broker membership"
fi

if [[ "$HAS_BROKER_MEMBERSHIP" != "true" ]]; then
    echo "Cannot continue without broker membership for trucker referral codes"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create trucker referral code with required fields"
REFERRAL_CODE="REF-$(date +%s)-${RANDOM}"

xbe_json do trucker-referral-codes create \
    --broker "$CREATED_BROKER_ID" \
    --code "$REFERRAL_CODE" \
    --value 25

if [[ $status -eq 0 ]]; then
    CREATED_REFERRAL_CODE_ID=$(json_get ".id")
    if [[ -n "$CREATED_REFERRAL_CODE_ID" && "$CREATED_REFERRAL_CODE_ID" != "null" ]]; then
        register_cleanup "trucker-referral-codes" "$CREATED_REFERRAL_CODE_ID"
        pass
    else
        fail "Created trucker referral code but no ID returned"
    fi
else
    fail "Failed to create trucker referral code"
fi

if [[ -z "$CREATED_REFERRAL_CODE_ID" || "$CREATED_REFERRAL_CODE_ID" == "null" ]]; then
    echo "Cannot continue without a valid trucker referral code ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucker referral code"
xbe_json view trucker-referral-codes show "$CREATED_REFERRAL_CODE_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Resolve secondary broker for update"
if [[ -n "$MEMBERSHIP_BROKER_IDS" ]]; then
    SECOND_BROKER_ID=$(echo "$MEMBERSHIP_BROKER_IDS" | grep -v "^${CREATED_BROKER_ID}$" | head -n 1)
    if [[ -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
        HAS_SECOND_BROKER_MEMBERSHIP="true"
        pass
    else
        skip "No second broker membership found"
    fi
else
    skip "No membership list available for second broker lookup"
fi

if [[ -z "$SECOND_BROKER_ID" || "$SECOND_BROKER_ID" == "null" ]]; then
    test_name "Create secondary broker for update"
    SECOND_BROKER_NAME=$(unique_name "TRCBroker2")

    xbe_json do brokers create --name "$SECOND_BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        SECOND_BROKER_ID=$(json_get ".id")
        if [[ -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$SECOND_BROKER_ID"
            pass
        else
            fail "Created second broker but no ID returned"
            echo "Cannot continue without a second broker"
            run_tests
        fi
    else
        fail "Failed to create second broker"
        echo "Cannot continue without a second broker"
        run_tests
    fi
fi

if [[ "$HAS_SECOND_BROKER_MEMBERSHIP" != "true" && -n "$USER_ID" && "$USER_ID" != "null" && -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
    test_name "Ensure secondary broker membership for update"
    xbe_json do memberships create --user "$USER_ID" --organization "Broker|$SECOND_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        SECOND_BROKER_MEMBERSHIP_ID=$(json_get ".id")
        if [[ -n "$SECOND_BROKER_MEMBERSHIP_ID" && "$SECOND_BROKER_MEMBERSHIP_ID" != "null" ]]; then
            register_cleanup "memberships" "$SECOND_BROKER_MEMBERSHIP_ID"
            HAS_SECOND_BROKER_MEMBERSHIP="true"
            pass
        else
            skip "Created second broker membership but no ID returned"
        fi
    else
        skip "Unable to create second broker membership"
    fi
fi

test_name "Update trucker referral code attributes"
UPDATED_CODE="REF-UPDATED-$(date +%s)-${RANDOM}"

xbe_json do trucker-referral-codes update "$CREATED_REFERRAL_CODE_ID" \
    --code "$UPDATED_CODE" \
    --value 75
assert_success

if [[ "$HAS_SECOND_BROKER_MEMBERSHIP" == "true" ]]; then
    test_name "Update trucker referral code broker"
    xbe_json do trucker-referral-codes update "$CREATED_REFERRAL_CODE_ID" \
        --broker "$SECOND_BROKER_ID"
    assert_success
else
    test_name "Update trucker referral code broker"
    skip "No secondary broker membership available"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trucker referral codes"
xbe_json view trucker-referral-codes list --limit 5
assert_success

test_name "List trucker referral codes returns array"
xbe_json view trucker-referral-codes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker referral codes"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List trucker referral codes with --code filter"
xbe_json view trucker-referral-codes list --code "$UPDATED_CODE" --limit 5
assert_success

test_name "List trucker referral codes with --created-at-min filter"
xbe_json view trucker-referral-codes list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List trucker referral codes with --created-at-max filter"
xbe_json view trucker-referral-codes list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List trucker referral codes with --is-created-at filter"
xbe_json view trucker-referral-codes list --is-created-at true --limit 5
assert_success

test_name "List trucker referral codes with --updated-at-min filter"
xbe_json view trucker-referral-codes list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List trucker referral codes with --updated-at-max filter"
xbe_json view trucker-referral-codes list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List trucker referral codes with --is-updated-at filter"
xbe_json view trucker-referral-codes list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List trucker referral codes with --limit"
xbe_json view trucker-referral-codes list --limit 3
assert_success

test_name "List trucker referral codes with --offset"
xbe_json view trucker-referral-codes list --limit 3 --offset 3
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create trucker referral code without broker fails"
xbe_json do trucker-referral-codes create --code "MISSING-BROKER"
assert_failure

test_name "Create trucker referral code without code fails"
xbe_json do trucker-referral-codes create --broker "$CREATED_BROKER_ID"
assert_failure
