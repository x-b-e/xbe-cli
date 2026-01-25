#!/bin/bash
#
# XBE CLI Integration Tests: Key Result Scrappages
#
# Tests view and create operations for key_result_scrappages.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SCRAPPABLE_KEY_RESULT_ID=""
SAMPLE_SCRAPPAGE_ID=""
SKIP_MUTATION=0

describe "Resource: key-result-scrappages"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List key result scrappages"
xbe_json view key-result-scrappages list --limit 1
assert_success

test_name "Capture sample key result scrappage (if available)"
xbe_json view key-result-scrappages list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_SCRAPPAGE_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No key result scrappages available; skipping show test."
        pass
    fi
else
    fail "Failed to list key result scrappages"
fi

if [[ -n "$SAMPLE_SCRAPPAGE_ID" && "$SAMPLE_SCRAPPAGE_ID" != "null" ]]; then
    test_name "Show key result scrappage"
    xbe_json view key-result-scrappages show "$SAMPLE_SCRAPPAGE_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create key result scrappage requires --key-result"
xbe_run do key-result-scrappages create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find key result to scrap"
    response_file=$(mktemp)
    found=0
    statuses=("unknown" "not_started" "red" "yellow" "green")

    for status_value in "${statuses[@]}"; do
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -G "$XBE_BASE_URL/v1/key-results" \
            --data-urlencode "page[limit]=1" \
            --data-urlencode "filter[status]=${status_value}" \
            --data-urlencode "fields[key-results]=status"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            SCRAPPABLE_KEY_RESULT_ID=$(jq -r '.data[0].id' "$response_file")
            if [[ -n "$SCRAPPABLE_KEY_RESULT_ID" && "$SCRAPPABLE_KEY_RESULT_ID" != "null" ]]; then
                found=1
                break
            fi
        fi
    done

    if [[ $found -eq 1 ]]; then
        pass
    else
        skip "No key result found in scrappable statuses"
    fi

    rm -f "$response_file"
fi

if [[ -n "$SCRAPPABLE_KEY_RESULT_ID" && "$SCRAPPABLE_KEY_RESULT_ID" != "null" ]]; then
    test_name "Create key result scrappage"
    COMMENT="$(unique_name "Scrappage")"
    xbe_json do key-result-scrappages create --key-result "$SCRAPPABLE_KEY_RESULT_ID" --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create key result scrappage"
    fi
else
    test_name "Create key result scrappage"
    skip "No scrappable key result available for creation"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
