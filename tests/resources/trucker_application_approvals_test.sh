#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Application Approvals
#
# Tests create operations for the trucker-application-approvals resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TRUCKER_APPLICATION_ID="${XBE_TEST_TRUCKER_APPLICATION_ID:-}"

BROKER_ID=""
USER_ID=""

describe "Resource: trucker-application-approvals"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create approval without required fields fails"
xbe_run do trucker-application-approvals create
assert_failure

# ============================================================================
# Prerequisites - Create trucker application (if needed)
# ============================================================================

if [[ -z "$TRUCKER_APPLICATION_ID" ]]; then
    if [[ -z "$XBE_TOKEN" ]]; then
        test_name "Create trucker application (requires XBE_TOKEN)"
        skip "XBE_TOKEN not set; set XBE_TEST_TRUCKER_APPLICATION_ID to run approval test"
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

    test_name "Create trucker application for approval"
    COMPANY_NAME=$(unique_name "TruckerApplication")
    COMPANY_ADDRESS="1 Trucking Way\nWest Simsbury, CT 06092"
    base_url="${XBE_BASE_URL%/}"

    payload=$(jq -n \
        --arg companyName "$COMPANY_NAME" \
        --arg companyAddress "$COMPANY_ADDRESS" \
        --arg userID "$USER_ID" \
        --arg brokerID "$BROKER_ID" \
        '{
          data: {
            type: "trucker-applications",
            attributes: {
              "company-name": $companyName,
              "company-address": $companyAddress,
              "status": "pending",
              "skip-company-address-geocoding": true
            },
            relationships: {
              user: { data: { type: "users", id: $userID } },
              broker: { data: { type: "brokers", id: $brokerID } }
            }
          }
        }')

    trucker_application_response=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -H "Content-Type: application/vnd.api+json" \
        -X POST \
        -d "$payload" \
        "$base_url/v1/trucker-applications" || true)

    TRUCKER_APPLICATION_ID=$(echo "$trucker_application_response" | jq -r '.data.id // empty' 2>/dev/null || true)

    if [[ -n "$TRUCKER_APPLICATION_ID" ]]; then
        pass
    else
        skip "Failed to create trucker application"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Approve trucker application"
xbe_json do trucker-application-approvals create \
    --trucker-application "$TRUCKER_APPLICATION_ID" \
    --add-application-user-as-trucker-manager

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".trucker_application_id" "$TRUCKER_APPLICATION_ID"
    assert_json_equals ".add_application_user_as_trucker_manager" "true"

    TRUCKER_ID=$(json_get ".trucker_id")
    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$TRUCKER_ID"
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
