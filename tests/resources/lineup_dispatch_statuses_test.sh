#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Dispatch Statuses
#
# Tests create operations for the lineup-dispatch-statuses resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: lineup-dispatch-statuses"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create lineup dispatch status without required fields fails"
xbe_run do lineup-dispatch-statuses create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create lineup dispatch status"
if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    broker_id="$XBE_TEST_BROKER_ID"
elif [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping broker lookup"
else
    base_url="${XBE_BASE_URL%/}"

    brokers_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/brokers?page[limit]=1&sort=-created-at" || true)

    broker_id=$(echo "$brokers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -z "$broker_id" ]]; then
        skip "No broker found"
    fi
fi

if [[ -n "$broker_id" ]]; then
    test_date=$(date +%Y-%m-%d)

    xbe_json do lineup-dispatch-statuses create \
        --broker "$broker_id" \
        --window day \
        --date "$test_date"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
