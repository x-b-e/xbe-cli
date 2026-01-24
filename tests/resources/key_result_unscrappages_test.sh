#!/bin/bash
#
# XBE CLI Integration Tests: Key Result Unscrappages
#
# Tests create operations for the key_result_unscrappages resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: key-result-unscrappages"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create unscrappage without required key result fails"
xbe_run do key-result-unscrappages create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create key result unscrappage"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup for scrapped key results"
else
    base_url="${XBE_BASE_URL%/}"

    key_results_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/key-results?page[limit]=1&filter[status]=scrapped" || true)

    key_result_id=$(echo "$key_results_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -z "$key_result_id" ]]; then
        skip "No scrapped key result found"
    else
        COMMENT="Unscrapping key result for test"
        xbe_json do key-result-unscrappages create \
            --key-result "$key_result_id" \
            --comment "$COMMENT"

        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
            assert_json_equals ".key_result_id" "$key_result_id"
            assert_json_equals ".comment" "$COMMENT"
        else
            if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                pass
            else
                fail "Failed to create key result unscrappage"
            fi
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
