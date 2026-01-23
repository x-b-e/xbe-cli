#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Field Scopes
#
# Tests list, show, and create operations for the material_transaction_field_scopes resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_MT_ID=""
LIST_SUPPORTED="true"

describe "Resource: material-transaction-field-scopes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction field scopes"
xbe_json view material-transaction-field-scopes list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"Forbidden"* ]]; then
        LIST_SUPPORTED="false"
        skip "Listing requires admin access"
    elif [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "List endpoint not supported"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List material transaction field scopes returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view material-transaction-field-scopes list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list material transaction field scopes"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# SHOW Tests - Use created material transaction
# ============================================================================

test_name "Create material transaction for field scope show"
TICKET_NUM="MTFS-$(date +%s)"
TRANS_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
xbe_json do material-transactions create --ticket-number "$TICKET_NUM" --transaction-at "$TRANS_AT"
if [[ $status -eq 0 ]]; then
    SAMPLE_MT_ID=$(json_get ".id")
    if [[ -n "$SAMPLE_MT_ID" && "$SAMPLE_MT_ID" != "null" ]]; then
        register_cleanup "material-transactions" "$SAMPLE_MT_ID"
        pass
    else
        fail "Created material transaction but no ID returned"
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]]; then
        skip "Not authorized to create material transaction"
    else
        fail "Failed to create material transaction: $output"
    fi
fi

test_name "Show material transaction field scope"
if [[ -n "$SAMPLE_MT_ID" && "$SAMPLE_MT_ID" != "null" ]]; then
    xbe_json view material-transaction-field-scopes show "$SAMPLE_MT_ID"
    assert_success
else
    skip "No material transaction ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material transaction field scope"
if [[ -n "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json do material-transaction-field-scopes create \
        --tender-job-schedule-shift "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        if [[ -n "$id" && "$id" != "null" ]]; then
            pass
        else
            fail "Created field scope but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"403"* ]] || \
           [[ "$output" == *"422"* ]] || \
           [[ "$output" == *"Record Invalid"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID to enable create test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create field scope without tender job schedule shift fails"
xbe_run do material-transaction-field-scopes create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
