#!/bin/bash
#
# XBE CLI Integration Tests: Incident Request Rejections
#
# Tests view and create operations for incident_request_rejections.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUBMITTED_INCIDENT_REQUEST_ID=""
SECOND_SUBMITTED_INCIDENT_REQUEST_ID=""
SAMPLE_REJECTION_ID=""
SKIP_MUTATION=0

describe "Resource: incident-request-rejections"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List incident request rejections"
xbe_json view incident-request-rejections list --limit 1
assert_success

test_name "Capture sample incident request rejection (if available)"
xbe_json view incident-request-rejections list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_REJECTION_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No incident request rejections available; skipping show test."
        pass
    fi
else
    fail "Failed to list incident request rejections"
fi

if [[ -n "$SAMPLE_REJECTION_ID" && "$SAMPLE_REJECTION_ID" != "null" ]]; then
    test_name "Show incident request rejection"
    xbe_json view incident-request-rejections show "$SAMPLE_REJECTION_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create incident request rejection requires --incident-request"
xbe_run do incident-request-rejections create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find submitted incident request"
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/incident-requests" \
        --data-urlencode "page[limit]=2" \
        --data-urlencode "filter[status]=submitted" \
        --data-urlencode "fields[incident-requests]=status"

    http_code="$output"
    if [[ $status -eq 0 && "$http_code" == 2* ]]; then
        mapfile -t submitted_ids < <(jq -r '.data[].id' "$response_file" 2>/dev/null)
        SUBMITTED_INCIDENT_REQUEST_ID="${submitted_ids[0]}"
        SECOND_SUBMITTED_INCIDENT_REQUEST_ID="${submitted_ids[1]}"
        if [[ -n "$SUBMITTED_INCIDENT_REQUEST_ID" && "$SUBMITTED_INCIDENT_REQUEST_ID" != "null" ]]; then
            pass
        else
            skip "No submitted incident request found"
        fi
    else
        skip "Unable to list incident requests (HTTP ${http_code})"
    fi
    rm -f "$response_file"
fi

if [[ -n "$SUBMITTED_INCIDENT_REQUEST_ID" && "$SUBMITTED_INCIDENT_REQUEST_ID" != "null" ]]; then
    if [[ -n "$SECOND_SUBMITTED_INCIDENT_REQUEST_ID" && "$SECOND_SUBMITTED_INCIDENT_REQUEST_ID" != "null" ]]; then
        test_name "Create incident request rejection (minimal)"
        xbe_json do incident-request-rejections create --incident-request "$SUBMITTED_INCIDENT_REQUEST_ID"
        assert_success

        test_name "Create incident request rejection with comment"
        COMMENT_TEXT="$(unique_name "IncidentRequestRejection")"
        xbe_json do incident-request-rejections create \
            --incident-request "$SECOND_SUBMITTED_INCIDENT_REQUEST_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to create incident request rejection with comment"
        fi
    else
        test_name "Create incident request rejection with comment"
        COMMENT_TEXT="$(unique_name "IncidentRequestRejection")"
        xbe_json do incident-request-rejections create \
            --incident-request "$SUBMITTED_INCIDENT_REQUEST_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to create incident request rejection"
        fi
    fi
else
    test_name "Create incident request rejection"
    skip "No submitted incident request available for rejection tests"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create incident request rejection with invalid ID fails"
xbe_run do incident-request-rejections create --incident-request "999999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
