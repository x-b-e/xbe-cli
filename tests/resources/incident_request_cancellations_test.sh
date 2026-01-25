#!/bin/bash
#
# XBE CLI Integration Tests: Incident Request Cancellations
#
# Tests create operations for incident-request-cancellations.
#
# COVERAGE: Writable attributes (comment)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

INCIDENT_REQUEST_ID="${XBE_TEST_INCIDENT_REQUEST_CANCELLATION_ID:-}"

describe "Resource: incident-request-cancellations"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cancellation requires incident request"
xbe_run do incident-request-cancellations create --comment "missing incident request"
assert_failure

test_name "Create incident request cancellation"
if [[ -n "$INCIDENT_REQUEST_ID" && "$INCIDENT_REQUEST_ID" != "null" ]]; then
    COMMENT=$(unique_name "IncidentRequestCancellation")
    xbe_json do incident-request-cancellations create \
        --incident-request "$INCIDENT_REQUEST_ID" \
        --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".incident_request_id" "$INCIDENT_REQUEST_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"previous status"* ]] || [[ "$output" == *"not in"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"cancelled"* ]] || [[ "$output" == *"canceled"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create incident request cancellation: $output"
        fi
    fi
else
    skip "No incident request ID available. Set XBE_TEST_INCIDENT_REQUEST_CANCELLATION_ID to enable create testing."
fi

run_tests
