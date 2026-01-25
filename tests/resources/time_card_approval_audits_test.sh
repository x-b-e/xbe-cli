#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Approval Audits
#
# Tests CRUD operations for the time_card_approval_audits resource.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_AUDIT_ID=""
SAMPLE_AUDIT_ID=""
SAMPLE_TIME_CARD_ID=""
SAMPLE_USER_ID=""
SKIP_CREATE=0

describe "Resource: time-card-approval-audits"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card approval audits"
xbe_json view time-card-approval-audits list --limit 5
assert_success

test_name "List time card approval audits returns array"
xbe_json view time-card-approval-audits list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time card approval audits"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate time card approval audit for filters"
xbe_json view time-card-approval-audits list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_AUDIT_ID=$(json_get ".[0].id")
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
        SAMPLE_USER_ID=$(json_get ".[0].user_id")
        pass
    else
        if [[ -n "$XBE_TEST_TIME_CARD_APPROVAL_AUDIT_ID" ]]; then
            xbe_json view time-card-approval-audits show "$XBE_TEST_TIME_CARD_APPROVAL_AUDIT_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_AUDIT_ID=$(json_get ".id")
                SAMPLE_TIME_CARD_ID=$(json_get ".time_card_id")
                SAMPLE_USER_ID=$(json_get ".user_id")
                pass
            else
                skip "Failed to load XBE_TEST_TIME_CARD_APPROVAL_AUDIT_ID"
            fi
        else
            skip "No time card approval audits found. Set XBE_TEST_TIME_CARD_APPROVAL_AUDIT_ID for filter tests."
        fi
    fi
else
    fail "Failed to list time card approval audits for filters"
fi

# ============================================================================
# Show Tests
# ============================================================================

if [[ -n "$SAMPLE_AUDIT_ID" && "$SAMPLE_AUDIT_ID" != "null" ]]; then
    test_name "Show time card approval audit"
    xbe_json view time-card-approval-audits show "$SAMPLE_AUDIT_ID"
    assert_success
else
    test_name "Show time card approval audit"
    skip "No sample audit available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    test_name "Filter by time card"
    xbe_json view time-card-approval-audits list --time-card "$SAMPLE_TIME_CARD_ID"
    assert_success
else
    test_name "Filter by time card"
    skip "No time card ID available"
fi

if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    test_name "Filter by user"
    xbe_json view time-card-approval-audits list --user "$SAMPLE_USER_ID"
    assert_success
else
    test_name "Filter by user"
    skip "No user ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

CREATE_TIME_CARD_ID="${XBE_TEST_TIME_CARD_ID:-}"
CREATE_USER_ID="${XBE_TEST_TIME_CARD_AUDITOR_USER_ID:-}"

if [[ -n "$CREATE_TIME_CARD_ID" && -n "$CREATE_USER_ID" ]]; then
    test_name "Create time card approval audit"
    xbe_json do time-card-approval-audits create \
        --time-card "$CREATE_TIME_CARD_ID" \
        --user "$CREATE_USER_ID" \
        --note "CLI audit test"

    if [[ $status -eq 0 ]]; then
        CREATED_AUDIT_ID=$(json_get ".id")
        if [[ -n "$CREATED_AUDIT_ID" && "$CREATED_AUDIT_ID" != "null" ]]; then
            register_cleanup "time-card-approval-audits" "$CREATED_AUDIT_ID"
            pass
        else
            fail "Created audit but no ID returned"
        fi
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Create rejected (validation or authorization)"
            SKIP_CREATE=1
        else
            fail "Failed to create time card approval audit"
            SKIP_CREATE=1
        fi
    fi
else
    test_name "Create time card approval audit"
    skip "Set XBE_TEST_TIME_CARD_ID and XBE_TEST_TIME_CARD_AUDITOR_USER_ID to enable create tests"
    SKIP_CREATE=1
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_AUDIT_ID" && "$CREATED_AUDIT_ID" != "null" ]]; then
    test_name "Update time card approval audit note"
    xbe_json do time-card-approval-audits update "$CREATED_AUDIT_ID" --note "Updated audit note"
    assert_success

    test_name "Update time card approval audit without fields fails"
    xbe_json do time-card-approval-audits update "$CREATED_AUDIT_ID"
    assert_failure
else
    test_name "Update time card approval audit note"
    skip "No audit created"
    test_name "Update time card approval audit without fields fails"
    skip "No audit created"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_AUDIT_ID" && "$CREATED_AUDIT_ID" != "null" ]]; then
    test_name "Delete time card approval audit requires --confirm flag"
    xbe_run do time-card-approval-audits delete "$CREATED_AUDIT_ID"
    assert_failure

    test_name "Delete time card approval audit with --confirm"
    xbe_run do time-card-approval-audits delete "$CREATED_AUDIT_ID" --confirm
    assert_success
else
    test_name "Delete time card approval audit requires --confirm flag"
    skip "No audit created"
    test_name "Delete time card approval audit with --confirm"
    skip "No audit created"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
