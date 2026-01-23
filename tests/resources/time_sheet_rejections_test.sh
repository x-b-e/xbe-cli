#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Rejections
#
# Tests view and create operations for time_sheet_rejections.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUBMITTED_TIME_SHEET_ID=""
SAMPLE_REJECTION_ID=""
SKIP_MUTATION=0

describe "Resource: time-sheet-rejections"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List time sheet rejections"
xbe_json view time-sheet-rejections list --limit 1
assert_success

test_name "Capture sample time sheet rejection (if available)"
xbe_json view time-sheet-rejections list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_REJECTION_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No time sheet rejections available; skipping show test."
        pass
    fi
else
    fail "Failed to list time sheet rejections"
fi

if [[ -n "$SAMPLE_REJECTION_ID" && "$SAMPLE_REJECTION_ID" != "null" ]]; then
    test_name "Show time sheet rejection"
    xbe_json view time-sheet-rejections show "$SAMPLE_REJECTION_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create time sheet rejection requires --time-sheet"
xbe_run do time-sheet-rejections create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find submitted time sheet"
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/time-sheets" \
        --data-urlencode "page[limit]=25" \
        --data-urlencode "filter[status]=submitted" \
        --data-urlencode "fields[time-sheets]=status"

    http_code="$output"
    if [[ $status -eq 0 && "$http_code" == 2* ]]; then
        SUBMITTED_TIME_SHEET_ID=$(jq -r '.data[0].id' "$response_file")
        if [[ -n "$SUBMITTED_TIME_SHEET_ID" && "$SUBMITTED_TIME_SHEET_ID" != "null" ]]; then
            pass
        else
            skip "No submitted time sheet found"
        fi
    else
        skip "Unable to list time sheets (HTTP ${http_code})"
    fi
    rm -f "$response_file"
fi

if [[ -n "$SUBMITTED_TIME_SHEET_ID" && "$SUBMITTED_TIME_SHEET_ID" != "null" ]]; then
    test_name "Create time sheet rejection"
    COMMENT="$(unique_name "Rejection")"
    xbe_json do time-sheet-rejections create --time-sheet "$SUBMITTED_TIME_SHEET_ID" --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create time sheet rejection"
    fi
else
    test_name "Create time sheet rejection"
    skip "No submitted time sheet available for creation"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
