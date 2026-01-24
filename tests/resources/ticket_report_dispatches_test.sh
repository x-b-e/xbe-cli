#!/bin/bash
#
# XBE CLI Integration Tests: Ticket Report Dispatches
#
# Tests list, show, create, and delete operations for ticket_report_dispatches.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_TICKET_REPORT_DISPATCH_ID=""
TICKET_REPORT_ID="${XBE_TEST_TICKET_REPORT_ID:-}"
CREATE_TICKET_REPORT_ID="${XBE_TEST_TICKET_REPORT_ID:-}"
CREATED_TICKET_REPORT_DISPATCH_ID=""

describe "Resource: ticket-report-dispatches"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List ticket report dispatches"
xbe_json view ticket-report-dispatches list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_TICKET_REPORT_DISPATCH_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$TICKET_REPORT_ID" || "$TICKET_REPORT_ID" == "null" ]]; then
            TICKET_REPORT_ID=$(echo "$output" | jq -r '.[0].ticket_report_id')
        fi
    fi
else
    fail "Failed to list ticket report dispatches"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show ticket report dispatch"
if [[ -n "$SEED_TICKET_REPORT_DISPATCH_ID" && "$SEED_TICKET_REPORT_DISPATCH_ID" != "null" ]]; then
    xbe_json view ticket-report-dispatches show "$SEED_TICKET_REPORT_DISPATCH_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        if [[ -z "$TICKET_REPORT_ID" || "$TICKET_REPORT_ID" == "null" ]]; then
            TICKET_REPORT_ID=$(json_get ".ticket_report_id")
        fi
        pass
    else
        fail "Failed to show ticket report dispatch"
    fi
else
    skip "No ticket report dispatch available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create ticket report dispatch"
if [[ -n "$CREATE_TICKET_REPORT_ID" && "$CREATE_TICKET_REPORT_ID" != "null" ]]; then
    xbe_json do ticket-report-dispatches create --ticket-report "$CREATE_TICKET_REPORT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_TICKET_REPORT_DISPATCH_ID=$(json_get ".id")
        if [[ -n "$CREATED_TICKET_REPORT_DISPATCH_ID" && "$CREATED_TICKET_REPORT_DISPATCH_ID" != "null" ]]; then
            register_cleanup "ticket-report-dispatches" "$CREATED_TICKET_REPORT_DISPATCH_ID"
            pass
        else
            fail "Created ticket report dispatch but no ID returned"
        fi
    else
        fail "Failed to create ticket report dispatch"
    fi
else
    skip "No ticket report ID available for creation (set XBE_TEST_TICKET_REPORT_ID to a valid ticket report)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete ticket report dispatch"
if [[ -n "$CREATED_TICKET_REPORT_DISPATCH_ID" && "$CREATED_TICKET_REPORT_DISPATCH_ID" != "null" ]]; then
    xbe_run do ticket-report-dispatches delete "$CREATED_TICKET_REPORT_DISPATCH_ID" --confirm
    assert_success
else
    skip "No created ticket report dispatch to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by ticket report"
if [[ -n "$TICKET_REPORT_ID" && "$TICKET_REPORT_ID" != "null" ]]; then
    xbe_json view ticket-report-dispatches list --ticket-report "$TICKET_REPORT_ID" --limit 5
    assert_success
else
    skip "No ticket report ID available for filter"
fi

run_tests
