#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Unscrappages
#
# Tests view and create operations for time_card_unscrappages.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SCRAPPED_TIME_CARD_ID=""
SAMPLE_UNSCRAPPAGE_ID=""
SKIP_MUTATION=0

describe "Resource: time-card-unscrappages"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List time card unscrappages"
xbe_json view time-card-unscrappages list --limit 1
assert_success

test_name "Capture sample time card unscrappage (if available)"
xbe_json view time-card-unscrappages list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_UNSCRAPPAGE_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No time card unscrappages available; skipping show test."
        pass
    fi
else
    fail "Failed to list time card unscrappages"
fi

if [[ -n "$SAMPLE_UNSCRAPPAGE_ID" && "$SAMPLE_UNSCRAPPAGE_ID" != "null" ]]; then
    test_name "Show time card unscrappage"
    xbe_json view time-card-unscrappages show "$SAMPLE_UNSCRAPPAGE_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create time card unscrappage requires --time-card"
xbe_run do time-card-unscrappages create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find scrapped time card"
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/time-cards" \
        --data-urlencode "page[limit]=25" \
        --data-urlencode "filter[status]=scrapped" \
        --data-urlencode "fields[time-cards]=status"

    http_code="$output"
    if [[ $status -eq 0 && "$http_code" == 2* ]]; then
        SCRAPPED_TIME_CARD_ID=$(jq -r '.data[0].id' "$response_file")
        if [[ -n "$SCRAPPED_TIME_CARD_ID" && "$SCRAPPED_TIME_CARD_ID" != "null" ]]; then
            pass
        else
            skip "No scrapped time card found"
        fi
    else
        skip "Unable to list time cards (HTTP ${http_code})"
    fi
    rm -f "$response_file"
fi

if [[ -n "$SCRAPPED_TIME_CARD_ID" && "$SCRAPPED_TIME_CARD_ID" != "null" ]]; then
    test_name "Create time card unscrappage"
    COMMENT="$(unique_name "Unscrappage")"
    xbe_json do time-card-unscrappages create --time-card "$SCRAPPED_TIME_CARD_ID" --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create time card unscrappage"
    fi
else
    test_name "Create time card unscrappage"
    skip "No scrapped time card available for creation"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
