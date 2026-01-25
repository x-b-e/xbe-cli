#!/bin/bash
#
# XBE CLI Integration Tests: Incident Request Approvals
#
# Tests view and create operations for incident_request_approvals.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUBMITTED_INCIDENT_REQUEST_ID=""
SECOND_SUBMITTED_INCIDENT_REQUEST_ID=""
SAMPLE_APPROVAL_ID=""
SKIP_MUTATION=0

describe "Resource: incident-request-approvals"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List incident request approvals"
xbe_json view incident-request-approvals list --limit 1
assert_success

test_name "Capture sample incident request approval (if available)"
xbe_json view incident-request-approvals list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_APPROVAL_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No incident request approvals available; skipping show test."
        pass
    fi
else
    fail "Failed to list incident request approvals"
fi

if [[ -n "$SAMPLE_APPROVAL_ID" && "$SAMPLE_APPROVAL_ID" != "null" ]]; then
    test_name "Show incident request approval"
    xbe_json view incident-request-approvals show "$SAMPLE_APPROVAL_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create incident request approval requires --incident-request"
xbe_run do incident-request-approvals create
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
        test_name "Create incident request approval (minimal)"
        xbe_json do incident-request-approvals create --incident-request "$SUBMITTED_INCIDENT_REQUEST_ID"
        assert_success

        test_name "Create incident request approval with comment"
        COMMENT_TEXT="$(unique_name "IncidentRequestApproval")"
        xbe_json do incident-request-approvals create \
            --incident-request "$SECOND_SUBMITTED_INCIDENT_REQUEST_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to create incident request approval with comment"
        fi
    else
        test_name "Create incident request approval with comment"
        COMMENT_TEXT="$(unique_name "IncidentRequestApproval")"
        xbe_json do incident-request-approvals create \
            --incident-request "$SUBMITTED_INCIDENT_REQUEST_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to create incident request approval"
        fi
    fi
else
    test_name "Create incident request approval"
    skip "No submitted incident request available for approval tests"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create incident request approval with invalid ID fails"
xbe_run do incident-request-approvals create --incident-request "999999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
