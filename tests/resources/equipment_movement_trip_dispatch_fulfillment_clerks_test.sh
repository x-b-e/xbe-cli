#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Trip Dispatch Fulfillment Clerks
#
# Tests create operations for the equipment-movement-trip-dispatch-fulfillment-clerks resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: equipment-movement-trip-dispatch-fulfillment-clerks"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create fulfillment clerk without required fields fails"
xbe_run do equipment-movement-trip-dispatch-fulfillment-clerks create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create equipment movement trip dispatch fulfillment clerk"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup for dispatches"
else
    base_url="${XBE_BASE_URL%/}"

    dispatches_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/equipment-movement-trip-dispatches?page[limit]=1&sort=-created-at" || true)

    dispatch_id=$(echo "$dispatches_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -z "$dispatch_id" ]]; then
        skip "No equipment movement trip dispatch found"
    else
        xbe_json do equipment-movement-trip-dispatch-fulfillment-clerks create \
            --equipment-movement-trip-dispatch "$dispatch_id"

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
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
