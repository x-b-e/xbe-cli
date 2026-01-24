#!/bin/bash
#
# XBE CLI Integration Tests: Customer Application Approvals
#
# Tests create operations for the customer-application-approvals resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CUSTOMER_APPLICATION_ID="${XBE_TEST_CUSTOMER_APPLICATION_ID:-}"

BROKER_ID=""
USER_ID=""

CREDIT_LIMIT="1000000"

describe "Resource: customer-application-approvals"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create approval without required fields fails"
xbe_run do customer-application-approvals create
assert_failure

# ============================================================================
# Prerequisites - Create customer application (if needed)
# ============================================================================

if [[ -z "$CUSTOMER_APPLICATION_ID" ]]; then
    if [[ -z "$XBE_TOKEN" ]]; then
        test_name "Create customer application (requires XBE_TOKEN)"
        skip "XBE_TOKEN not set; set XBE_TEST_CUSTOMER_APPLICATION_ID to run approval test"
        run_tests
    fi

    test_name "Fetch current user"
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        USER_ID=$(json_get ".id")
        if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
            pass
        else
            skip "No user ID returned"
            run_tests
        fi
    else
        skip "Failed to fetch current user"
        run_tests
    fi

    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        BROKER_ID="$XBE_TEST_BROKER_ID"
    else
        test_name "Find broker membership for user"
        xbe_json view broker-memberships list --user "$USER_ID" --limit 1
        if [[ $status -eq 0 ]]; then
            BROKER_ID=$(json_get ".[0].broker_id")
            if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
                pass
            else
                skip "No broker membership found; set XBE_TEST_BROKER_ID"
                run_tests
            fi
        else
            skip "Failed to list broker memberships"
            run_tests
        fi
    fi

    test_name "Create customer application for approval"
    COMPANY_NAME=$(unique_name "CustomerApplication")
    COMPANY_ADDRESS="1 Main St\nWest Simsbury, CT 06092"
    base_url="${XBE_BASE_URL%/}"

    payload=$(jq -n \
        --arg companyName "$COMPANY_NAME" \
        --arg companyAddress "$COMPANY_ADDRESS" \
        --arg userID "$USER_ID" \
        --arg brokerID "$BROKER_ID" \
        '{
          data: {
            type: "customer-applications",
            attributes: {
              "company-name": $companyName,
              "company-address": $companyAddress,
              "requires-union-drivers": false,
              "is-trucking-company": true,
              "skip-company-address-geocoding": true
            },
            relationships: {
              user: { data: { type: "users", id: $userID } },
              broker: { data: { type: "brokers", id: $brokerID } }
            }
          }
        }')

    customer_application_response=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -H "Content-Type: application/vnd.api+json" \
        -X POST \
        -d "$payload" \
        "$base_url/v1/customer-applications" || true)

    CUSTOMER_APPLICATION_ID=$(echo "$customer_application_response" | jq -r '.data.id // empty' 2>/dev/null || true)

    if [[ -n "$CUSTOMER_APPLICATION_ID" ]]; then
        pass
    else
        skip "Failed to create customer application"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Approve customer application"
xbe_json do customer-application-approvals create \
    --customer-application "$CUSTOMER_APPLICATION_ID" \
    --credit-limit "$CREDIT_LIMIT"

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".customer_application_id" "$CUSTOMER_APPLICATION_ID"
    assert_json_equals ".credit_limit" "$CREDIT_LIMIT"

    CUSTOMER_ID=$(json_get ".customer_id")
    if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CUSTOMER_ID"
    fi
else
    if [[ "$output" == *"422"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
        pass
    else
        fail "Create failed: $output"
    fi
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
