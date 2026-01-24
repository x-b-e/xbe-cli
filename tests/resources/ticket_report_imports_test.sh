#!/bin/bash
#
# XBE CLI Integration Tests: Ticket Report Imports
#
# Tests list, show, create, and delete operations for ticket-report-imports.
#
# COVERAGE: List filters + show + create + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_TICKET_REPORT_IMPORT_ID=""
CREATE_TICKET_REPORT_ID="${XBE_TEST_TICKET_REPORT_ID:-}"
CREATED_TICKET_REPORT_IMPORT_ID=""

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: ticket-report-imports"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List ticket report imports"
xbe_json view ticket-report-imports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_TICKET_REPORT_IMPORT_ID=$(echo "$output" | jq -r '.[0].id')
    fi
else
    fail "Failed to list ticket report imports"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List ticket report imports with --created-at-min"
xbe_json view ticket-report-imports list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List ticket report imports with --created-at-max"
xbe_json view ticket-report-imports list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List ticket report imports with --is-created-at=true"
xbe_json view ticket-report-imports list --is-created-at true --limit 5
assert_success

test_name "List ticket report imports with --is-created-at=false"
xbe_json view ticket-report-imports list --is-created-at false --limit 5
assert_success

test_name "List ticket report imports with --updated-at-min"
xbe_json view ticket-report-imports list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List ticket report imports with --updated-at-max"
xbe_json view ticket-report-imports list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List ticket report imports with --is-updated-at=true"
xbe_json view ticket-report-imports list --is-updated-at true --limit 5
assert_success

test_name "List ticket report imports with --is-updated-at=false"
xbe_json view ticket-report-imports list --is-updated-at false --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show ticket report import"
if [[ -n "$SEED_TICKET_REPORT_IMPORT_ID" && "$SEED_TICKET_REPORT_IMPORT_ID" != "null" ]]; then
    xbe_json view ticket-report-imports show "$SEED_TICKET_REPORT_IMPORT_ID"
    assert_success
else
    skip "No ticket report import available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create ticket report import"
if [[ -n "$CREATE_TICKET_REPORT_ID" && "$CREATE_TICKET_REPORT_ID" != "null" ]]; then
    xbe_json do ticket-report-imports create --ticket-report "$CREATE_TICKET_REPORT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_TICKET_REPORT_IMPORT_ID=$(json_get ".id")
        if [[ -n "$CREATED_TICKET_REPORT_IMPORT_ID" && "$CREATED_TICKET_REPORT_IMPORT_ID" != "null" ]]; then
            register_cleanup "ticket-report-imports" "$CREATED_TICKET_REPORT_IMPORT_ID"
            pass
        else
            fail "Created ticket report import but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Validation"* ]] || [[ "$output" == *"unprocessable"* ]] || [[ "$output" == *"already"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create ticket report import: $output"
        fi
    fi
else
    skip "No ticket report ID available for creation (set XBE_TEST_TICKET_REPORT_ID)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete ticket report import"
if [[ -n "$CREATED_TICKET_REPORT_IMPORT_ID" && "$CREATED_TICKET_REPORT_IMPORT_ID" != "null" ]]; then
    xbe_run do ticket-report-imports delete "$CREATED_TICKET_REPORT_IMPORT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"cannot be deleted"* ]] || [[ "$output" == *"related material transactions"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Delete blocked by server policy/validation"
        else
            fail "Failed to delete ticket report import"
        fi
    fi
else
    skip "No created ticket report import to delete"
fi

run_tests
