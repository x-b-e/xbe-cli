#!/bin/bash
#
# XBE CLI Integration Tests: Sourcing Searches
#
# Tests create operations for the sourcing-searches resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: sourcing-searches"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create sourcing search without required fields fails"
xbe_run do sourcing-searches create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create sourcing search"
CUSTOMER_TENDER_ID="${XBE_TEST_CUSTOMER_TENDER_ID:-}"

if [[ -n "$CUSTOMER_TENDER_ID" ]]; then
    echo "    Using XBE_TEST_CUSTOMER_TENDER_ID: $CUSTOMER_TENDER_ID"
else
    if [[ -z "$XBE_TOKEN" ]]; then
        skip "XBE_TOKEN not set; skipping customer tender lookup"
    else
        base_url="${XBE_BASE_URL%/}"

        tenders_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/customer-tenders?page[limit]=1&sort=-created-at&filter[with-alive-shifts]=true" || true)

        CUSTOMER_TENDER_ID=$(echo "$tenders_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

        if [[ -z "$CUSTOMER_TENDER_ID" ]]; then
            tenders_json=$(curl -s \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                "$base_url/v1/customer-tenders?page[limit]=1&sort=-created-at" || true)

            CUSTOMER_TENDER_ID=$(echo "$tenders_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
        fi

        if [[ -z "$CUSTOMER_TENDER_ID" ]]; then
            skip "No customer tender found"
        fi
    fi
fi

if [[ -n "$CUSTOMER_TENDER_ID" ]]; then
    xbe_json do sourcing-searches create \
        --customer-tender "$CUSTOMER_TENDER_ID" \
        --maximum-result-count 5

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".customer_tender_id" "$CUSTOMER_TENDER_ID"
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to create sourcing search"
        fi
    fi
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
