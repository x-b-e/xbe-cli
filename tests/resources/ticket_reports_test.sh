#!/bin/bash
#
# XBE CLI Integration Tests: Ticket Reports
#
# Tests CRUD operations for the ticket_reports resource.
#
# COVERAGE: Create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TICKET_REPORT_ID=""
CREATED_TICKET_REPORT_BROKER_ID=""
CREATED_TICKET_REPORT_TYPE_ID=""

LIST_TICKET_REPORT_ID=""
LIST_BROKER_ID=""
LIST_TICKET_REPORT_TYPE_ID=""

ENV_TICKET_REPORT_TYPE_ID="${XBE_TEST_TICKET_REPORT_TYPE_ID:-}"
ENV_TICKET_REPORT_FILE="${XBE_TEST_TICKET_REPORT_FILE:-}"
ENV_BROKER_ID="${XBE_TEST_BROKER_ID:-}"

describe "Resource: ticket_reports"

# ============================================================================
# LIST Tests (smoke)
# ============================================================================

test_name "List ticket reports"
xbe_json view ticket-reports list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_TICKET_REPORT_ID=$(json_get ".[0].id")
    LIST_BROKER_ID=$(json_get ".[0].broker_id")
    LIST_TICKET_REPORT_TYPE_ID=$(json_get ".[0].ticket_report_type_id")
    if [[ -n "$LIST_TICKET_REPORT_ID" && "$LIST_TICKET_REPORT_ID" != "null" ]]; then
        pass
    else
        skip "No ticket reports returned"
    fi
else
    fail "Failed to list ticket reports"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create ticket report without required flags fails"
xbe_run do ticket-reports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create ticket report with required fields"
if [[ -n "$ENV_TICKET_REPORT_TYPE_ID" && -n "$ENV_TICKET_REPORT_FILE" && -f "$ENV_TICKET_REPORT_FILE" ]]; then
    if [[ -n "$ENV_BROKER_ID" ]]; then
        xbe_json do ticket-reports create \
            --ticket-report-type "$ENV_TICKET_REPORT_TYPE_ID" \
            --broker "$ENV_BROKER_ID" \
            --file-path "$ENV_TICKET_REPORT_FILE"
    else
        xbe_json do ticket-reports create \
            --ticket-report-type "$ENV_TICKET_REPORT_TYPE_ID" \
            --file-path "$ENV_TICKET_REPORT_FILE"
    fi

    if [[ $status -eq 0 ]]; then
        CREATED_TICKET_REPORT_ID=$(json_get ".id")
        CREATED_TICKET_REPORT_BROKER_ID=$(json_get ".broker_id")
        CREATED_TICKET_REPORT_TYPE_ID=$(json_get ".ticket_report_type_id")
        if [[ -n "$CREATED_TICKET_REPORT_ID" && "$CREATED_TICKET_REPORT_ID" != "null" ]]; then
            register_cleanup "ticket-reports" "$CREATED_TICKET_REPORT_ID"
            pass
        else
            fail "Created ticket report but no ID returned"
        fi
    else
        fail "Failed to create ticket report"
    fi
else
    skip "Set XBE_TEST_TICKET_REPORT_TYPE_ID and XBE_TEST_TICKET_REPORT_FILE to run"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update ticket report file name"
if [[ -n "$CREATED_TICKET_REPORT_ID" && "$CREATED_TICKET_REPORT_ID" != "null" ]]; then
    UPDATED_NAME="CLI-TicketReport-$(date +%s)-${RANDOM}.csv"
    xbe_json do ticket-reports update "$CREATED_TICKET_REPORT_ID" --file-name "$UPDATED_NAME"
    assert_success
else
    skip "No ticket report available for update"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show ticket report details"
SHOW_ID="$CREATED_TICKET_REPORT_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$LIST_TICKET_REPORT_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view ticket-reports show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        actual_id=$(json_get ".id")
        if [[ "$actual_id" == "$SHOW_ID" ]]; then
            pass
        else
            fail "Expected show id $SHOW_ID, got $actual_id"
        fi
    else
        fail "Failed to show ticket report"
    fi
else
    skip "No ticket report ID available"
fi

# ============================================================================
# LIST Filter Tests
# ============================================================================

test_name "List ticket reports filtered by broker"
FILTER_BROKER_ID="$CREATED_TICKET_REPORT_BROKER_ID"
FILTER_REPORT_ID="$CREATED_TICKET_REPORT_ID"
if [[ -z "$FILTER_BROKER_ID" || "$FILTER_BROKER_ID" == "null" ]]; then
    FILTER_BROKER_ID="$LIST_BROKER_ID"
    FILTER_REPORT_ID="$LIST_TICKET_REPORT_ID"
fi

if [[ -n "$FILTER_BROKER_ID" && "$FILTER_BROKER_ID" != "null" ]]; then
    xbe_json view ticket-reports list --broker "$FILTER_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_REPORT_ID" && "$FILTER_REPORT_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_REPORT_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected ticket report in broker filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list ticket reports by broker"
    fi
else
    skip "No broker ID available for filter test"
fi

test_name "List ticket reports filtered by ticket report type"
FILTER_TYPE_ID="$CREATED_TICKET_REPORT_TYPE_ID"
FILTER_REPORT_ID="$CREATED_TICKET_REPORT_ID"
if [[ -z "$FILTER_TYPE_ID" || "$FILTER_TYPE_ID" == "null" ]]; then
    FILTER_TYPE_ID="$LIST_TICKET_REPORT_TYPE_ID"
    FILTER_REPORT_ID="$LIST_TICKET_REPORT_ID"
fi

if [[ -n "$FILTER_TYPE_ID" && "$FILTER_TYPE_ID" != "null" ]]; then
    xbe_json view ticket-reports list --ticket-report-type "$FILTER_TYPE_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_REPORT_ID" && "$FILTER_REPORT_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_REPORT_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected ticket report in type filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list ticket reports by type"
    fi
else
    skip "No ticket report type ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete ticket report"
if [[ -n "$CREATED_TICKET_REPORT_ID" && "$CREATED_TICKET_REPORT_ID" != "null" ]]; then
    xbe_run do ticket-reports delete "$CREATED_TICKET_REPORT_ID" --confirm
    assert_success
else
    skip "No ticket report available for deletion"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
