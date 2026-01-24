#!/bin/bash
#
# XBE CLI Integration Tests: Profit Improvement Subscriptions
#
# Tests CRUD operations for the profit_improvement_subscriptions resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUBSCRIPTION_ID=""
CREATED_PROFIT_IMPROVEMENT_ID=""
CREATED_BROKER_ID=""
CREATED_CATEGORY_ID=""
CREATED_SUBSCRIPTION_USER_ID=""
CREATED_SUBSCRIPTION_USER_EMAIL=""
CREATED_SUBSCRIPTION_USER_MOBILE=""
CREATED_MEMBERSHIP_ID=""

CURRENT_USER_ID=""
CURRENT_USER_IS_ADMIN=""
CURRENT_USER_EMAIL=""
CURRENT_USER_MOBILE=""

SUBSCRIPTION_CONTACT_METHOD=""
NEEDS_CONTACTABLE_USER=false
API_TOKEN=""

api_body=""
api_http_code=""
api_curl_status=""

normalize_base_url() {
    echo "${XBE_BASE_URL%/}"
}

resolve_api_token() {
    API_TOKEN="$XBE_TOKEN"
    if [[ -n "$API_TOKEN" ]]; then
        return 0
    fi

    local config_dir="${XDG_CONFIG_HOME:-$HOME/.config}"
    local config_path="${config_dir}/xbe/config.json"
    local normalized_base_url
    normalized_base_url=$(normalize_base_url)

    if [[ -f "$config_path" ]]; then
        API_TOKEN=$(jq -r --arg url "$normalized_base_url" '.tokens[$url] // empty' "$config_path")
    fi
}

api_request() {
    local method="$1"
    local path="$2"
    local payload="$3"
    local url
    url="$(normalize_base_url)${path}"

    api_body=""
    api_http_code=""
    api_curl_status=""

    if [[ -z "$API_TOKEN" ]]; then
        return 1
    fi

    set +e
    if [[ -n "$payload" ]]; then
        local response
        response=$(curl -sS -X "$method" "$url" \
            -H "Authorization: Bearer ${API_TOKEN}" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -d "$payload" \
            -w "\n%{http_code}")
        api_curl_status=$?
        api_http_code="${response##*$'\n'}"
        api_body="${response%$'\n'*}"
    else
        local response
        response=$(curl -sS -X "$method" "$url" \
            -H "Authorization: Bearer ${API_TOKEN}" \
            -H "Accept: application/vnd.api+json" \
            -w "\n%{http_code}")
        api_curl_status=$?
        api_http_code="${response##*$'\n'}"
        api_body="${response%$'\n'*}"
    fi
    set -e

    return 0
}

run_profit_improvement_cleanup() {
    if [[ -n "$CREATED_PROFIT_IMPROVEMENT_ID" && -n "$API_TOKEN" ]]; then
        api_request "DELETE" "/v1/profit-improvements/${CREATED_PROFIT_IMPROVEMENT_ID}"
    fi
}

trap 'run_profit_improvement_cleanup; run_cleanup' EXIT

describe "Resource: profit-improvement-subscriptions"

# ============================================================================
# Prerequisites - Current user
# ============================================================================

test_name "Fetch current authenticated user"
xbe_json auth whoami

if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    CURRENT_USER_IS_ADMIN=$(json_get ".is_admin")
    CURRENT_USER_EMAIL=$(json_get ".email")
    CURRENT_USER_MOBILE=$(json_get ".mobile")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "Could not determine current user ID"
        run_tests
    fi
else
    fail "Failed to fetch authenticated user"
    run_tests
fi

if [[ -n "$CURRENT_USER_EMAIL" && "$CURRENT_USER_EMAIL" != "null" ]]; then
    SUBSCRIPTION_CONTACT_METHOD="email_address"
elif [[ -n "$CURRENT_USER_MOBILE" && "$CURRENT_USER_MOBILE" != "null" ]]; then
    SUBSCRIPTION_CONTACT_METHOD="mobile_number"
else
    SUBSCRIPTION_CONTACT_METHOD="email_address"
    NEEDS_CONTACTABLE_USER=true
fi

resolve_api_token

# ============================================================================
# Admin setup - create broker/category/profit improvement if possible
# ============================================================================

if [[ "$CURRENT_USER_IS_ADMIN" == "true" && -n "$API_TOKEN" ]]; then
    test_name "Create prerequisite broker for subscription tests"
    BROKER_NAME=$(unique_name "PISubBroker")
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
        fail "Failed to create broker"
        run_tests
    fi

    test_name "Create prerequisite profit improvement category"
    CATEGORY_NAME=$(unique_name "PISubCategory")
    xbe_json do profit-improvement-categories create --name "$CATEGORY_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_CATEGORY_ID=$(json_get ".id")
        if [[ -n "$CREATED_CATEGORY_ID" && "$CREATED_CATEGORY_ID" != "null" ]]; then
            register_cleanup "profit-improvement-categories" "$CREATED_CATEGORY_ID"
            pass
        else
            fail "Created category but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create profit improvement category"
        run_tests
    fi

    if [[ "$NEEDS_CONTACTABLE_USER" == true ]]; then
        test_name "Create contactable user for subscription tests"
        USER_NAME=$(unique_name "PISubUser")
        USER_EMAIL=$(unique_email)
        USER_MOBILE=$(unique_mobile)

        xbe_json do users create \
            --name "$USER_NAME" \
            --email "$USER_EMAIL" \
            --mobile "$USER_MOBILE"

        if [[ $status -eq 0 ]]; then
            CREATED_SUBSCRIPTION_USER_ID=$(json_get ".id")
            CREATED_SUBSCRIPTION_USER_EMAIL="$USER_EMAIL"
            CREATED_SUBSCRIPTION_USER_MOBILE="$USER_MOBILE"
            if [[ -n "$CREATED_SUBSCRIPTION_USER_ID" && "$CREATED_SUBSCRIPTION_USER_ID" != "null" ]]; then
                pass
            else
                fail "Created user but no ID returned"
                run_tests
            fi
        else
            fail "Failed to create contactable user"
            run_tests
        fi

        test_name "Create broker membership for contactable user"
        xbe_json do memberships create \
            --user "$CREATED_SUBSCRIPTION_USER_ID" \
            --organization "Broker|$CREATED_BROKER_ID" \
            --kind "operations"

        if [[ $status -eq 0 ]]; then
            CREATED_MEMBERSHIP_ID=$(json_get ".id")
            if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
                register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
                pass
            else
                fail "Created membership but no ID returned"
                run_tests
            fi
        else
            fail "Failed to create broker membership"
            run_tests
        fi
    fi

    test_name "Create profit improvement via API"
    PI_TITLE=$(unique_name "PISub")
    PI_BODY=$(jq -n \
        --arg title "$PI_TITLE" \
        --arg category "$CREATED_CATEGORY_ID" \
        --arg broker "$CREATED_BROKER_ID" \
        '{
          data: {
            type: "profit-improvements",
            attributes: {
              title: $title,
              description: "CLI profit improvement subscription test"
            },
            relationships: {
              "profit-improvement-category": {
                data: {type: "profit-improvement-categories", id: $category}
              },
              organization: {
                data: {type: "brokers", id: $broker}
              }
            }
          }
        }')

    api_request "POST" "/v1/profit-improvements" "$PI_BODY"
    if [[ "$api_curl_status" -ne 0 ]]; then
        fail "Failed to call profit improvements API"
        run_tests
    fi

    if [[ "$api_http_code" =~ ^2 ]]; then
        CREATED_PROFIT_IMPROVEMENT_ID=$(echo "$api_body" | jq -r ".data.id")
        if [[ -n "$CREATED_PROFIT_IMPROVEMENT_ID" && "$CREATED_PROFIT_IMPROVEMENT_ID" != "null" ]]; then
            pass
        else
            fail "Created profit improvement but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create profit improvement (HTTP $api_http_code)"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create profit improvement subscription"

create_subscription() {
    local profit_improvement_id="$1"
    local user_id="$2"
    xbe_json do profit-improvement-subscriptions create \
        --user "$user_id" \
        --profit-improvement "$profit_improvement_id" \
        --contact-method "$SUBSCRIPTION_CONTACT_METHOD"
}

if [[ -n "$CREATED_SUBSCRIPTION_USER_ID" && "$CREATED_SUBSCRIPTION_USER_ID" != "null" ]]; then
    SUBSCRIPTION_USER_ID="$CREATED_SUBSCRIPTION_USER_ID"
else
    SUBSCRIPTION_USER_ID="$CURRENT_USER_ID"
    CREATED_SUBSCRIPTION_USER_EMAIL="$CURRENT_USER_EMAIL"
    CREATED_SUBSCRIPTION_USER_MOBILE="$CURRENT_USER_MOBILE"
fi

if [[ -n "$CREATED_PROFIT_IMPROVEMENT_ID" ]]; then
    create_subscription "$CREATED_PROFIT_IMPROVEMENT_ID" "$SUBSCRIPTION_USER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_SUBSCRIPTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
            register_cleanup "profit-improvement-subscriptions" "$CREATED_SUBSCRIPTION_ID"
            pass
        else
            fail "Created subscription but no ID returned"
        fi
    else
        fail "Failed to create profit improvement subscription"
    fi
else
    # Try using existing subscriptions to find a profit improvement we are not subscribed to
    xbe_json view profit-improvement-subscriptions list --limit 100
    if [[ $status -ne 0 ]]; then
        fail "Failed to list profit improvement subscriptions"
        run_tests
    fi

    CURRENT_USER_SUBS=$(echo "$output" | jq -r --arg uid "$CURRENT_USER_ID" '.[] | select(.user_id == $uid) | .profit_improvement_id' | sort -u)
    CANDIDATES=$(echo "$output" | jq -r --arg uid "$CURRENT_USER_ID" '.[] | select(.user_id != $uid) | .profit_improvement_id' | sort -u)

    for profit_improvement_id in $CANDIDATES; do
        if echo "$CURRENT_USER_SUBS" | grep -qx "$profit_improvement_id"; then
            continue
        fi
        create_subscription "$profit_improvement_id" "$SUBSCRIPTION_USER_ID"
        if [[ $status -eq 0 ]]; then
            CREATED_SUBSCRIPTION_ID=$(json_get ".id")
            CREATED_PROFIT_IMPROVEMENT_ID="$profit_improvement_id"
            if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
                register_cleanup "profit-improvement-subscriptions" "$CREATED_SUBSCRIPTION_ID"
                pass
                break
            else
                fail "Created subscription but no ID returned"
                run_tests
            fi
        fi
    done

    if [[ -z "$CREATED_SUBSCRIPTION_ID" && -n "$API_TOKEN" ]]; then
        api_request "GET" "/v1/profit-improvements?page[limit]=20"
        if [[ "$api_curl_status" -eq 0 && "$api_http_code" =~ ^2 ]]; then
            PROFIT_IMPROVEMENT_IDS=$(echo "$api_body" | jq -r ".data[].id")
            for profit_improvement_id in $PROFIT_IMPROVEMENT_IDS; do
                create_subscription "$profit_improvement_id" "$SUBSCRIPTION_USER_ID"
                if [[ $status -eq 0 ]]; then
                    CREATED_SUBSCRIPTION_ID=$(json_get ".id")
                    CREATED_PROFIT_IMPROVEMENT_ID="$profit_improvement_id"
                    if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
                        register_cleanup "profit-improvement-subscriptions" "$CREATED_SUBSCRIPTION_ID"
                        pass
                        break
                    fi
                fi
            done
        fi
    fi

    if [[ -z "$CREATED_SUBSCRIPTION_ID" || "$CREATED_SUBSCRIPTION_ID" == "null" ]]; then
        if [[ -z "$API_TOKEN" ]]; then
            skip "No existing profit improvement subscriptions and no API token to provision data"
        else
            fail "Unable to create a profit improvement subscription for the current user"
        fi
        run_tests
    fi
fi

# Only continue if we successfully created a subscription
if [[ -z "$CREATED_SUBSCRIPTION_ID" || "$CREATED_SUBSCRIPTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid subscription ID"
    run_tests
fi

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show profit improvement subscription"
xbe_json view profit-improvement-subscriptions show "$CREATED_SUBSCRIPTION_ID"
assert_success

# =========================================================================
# UPDATE Tests
# =========================================================================

test_name "Update subscription contact-method"
xbe_json do profit-improvement-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "$SUBSCRIPTION_CONTACT_METHOD"
assert_success

HAS_SUBSCRIPTION_USER_EMAIL=false
HAS_SUBSCRIPTION_USER_MOBILE=false
if [[ -n "$CREATED_SUBSCRIPTION_USER_EMAIL" && "$CREATED_SUBSCRIPTION_USER_EMAIL" != "null" ]]; then
    HAS_SUBSCRIPTION_USER_EMAIL=true
fi
if [[ -n "$CREATED_SUBSCRIPTION_USER_MOBILE" && "$CREATED_SUBSCRIPTION_USER_MOBILE" != "null" ]]; then
    HAS_SUBSCRIPTION_USER_MOBILE=true
fi

if [[ "$HAS_SUBSCRIPTION_USER_EMAIL" == true && "$HAS_SUBSCRIPTION_USER_MOBILE" == true ]]; then
    if [[ "$SUBSCRIPTION_CONTACT_METHOD" == "email_address" ]]; then
        test_name "Update subscription contact-method to mobile_number"
        xbe_json do profit-improvement-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "mobile_number"
        assert_success
    else
        test_name "Update subscription contact-method to email_address"
        xbe_json do profit-improvement-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "email_address"
        assert_success
    fi
else
    test_name "Update subscription contact-method to alternate method"
    skip "Subscription user does not have both email and mobile"
fi

# =========================================================================
# LIST Tests - Basic
# =========================================================================

test_name "List profit improvement subscriptions"
xbe_json view profit-improvement-subscriptions list
assert_success

test_name "List profit improvement subscriptions returns array"
xbe_json view profit-improvement-subscriptions list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list profit improvement subscriptions"
fi

# =========================================================================
# LIST Tests - Filters
# =========================================================================

test_name "List profit improvement subscriptions with --user filter"
xbe_json view profit-improvement-subscriptions list --user "$SUBSCRIPTION_USER_ID"
assert_success

test_name "List profit improvement subscriptions with --profit-improvement filter"
xbe_json view profit-improvement-subscriptions list --profit-improvement "$CREATED_PROFIT_IMPROVEMENT_ID"
assert_success

test_name "List profit improvement subscriptions with --contact-method filter"
xbe_json view profit-improvement-subscriptions list --contact-method "$SUBSCRIPTION_CONTACT_METHOD"
assert_success

# =========================================================================
# LIST Tests - Pagination
# =========================================================================

test_name "List profit improvement subscriptions with --limit"
xbe_json view profit-improvement-subscriptions list --limit 3
assert_success

test_name "List profit improvement subscriptions with --offset"
xbe_json view profit-improvement-subscriptions list --limit 3 --offset 3
assert_success

# =========================================================================
# DELETE Tests
# =========================================================================

test_name "Delete profit improvement subscription requires --confirm flag"
xbe_json do profit-improvement-subscriptions delete "$CREATED_SUBSCRIPTION_ID"
assert_failure

test_name "Delete profit improvement subscription with --confirm"
# Create a subscription specifically for deletion
CREATE_DELETE_ID="$CREATED_PROFIT_IMPROVEMENT_ID"
if [[ -n "$CREATE_DELETE_ID" ]]; then
    create_subscription "$CREATE_DELETE_ID" "$SUBSCRIPTION_USER_ID"
    if [[ $status -eq 0 ]]; then
        DELETE_ID=$(json_get ".id")
        xbe_json do profit-improvement-subscriptions delete "$DELETE_ID" --confirm
        assert_success
    else
        skip "Could not create subscription for deletion test"
    fi
else
    skip "No profit improvement ID available for deletion test"
fi

# =========================================================================
# Error Cases
# =========================================================================

test_name "Create profit improvement subscription without required flags fails"
xbe_json do profit-improvement-subscriptions create
assert_failure

test_name "Update without any fields fails"
xbe_json do profit-improvement-subscriptions update "$CREATED_SUBSCRIPTION_ID"
assert_failure

# =========================================================================
# Summary
# =========================================================================

run_tests
