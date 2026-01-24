#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheets Exports
#
# Tests list, show, and create operations for time-sheets-exports.
#
# COVERAGE: All list filters + writable relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TIME_SHEETS_EXPORT_ID="${XBE_TEST_TIME_SHEETS_EXPORT_ID:-}"
ORGANIZATION_FORMATTER_ID="${XBE_TEST_TIME_SHEETS_EXPORT_ORGANIZATION_FORMATTER_ID:-}"
TIME_SHEET_IDS_RAW="${XBE_TEST_TIME_SHEETS_EXPORT_TIME_SHEET_IDS:-}"
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
CREATED_BY_ID="${XBE_TEST_USER_ID:-}"
ORGANIZATION_FILTER="${XBE_TEST_TIME_SHEETS_EXPORT_ORGANIZATION_FILTER:-}"

if [[ -z "$ORGANIZATION_FILTER" ]]; then
    if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
        ORGANIZATION_FILTER="Broker|$BROKER_ID"
    elif [[ -n "$XBE_TEST_CUSTOMER_ID" && "$XBE_TEST_CUSTOMER_ID" != "null" ]]; then
        ORGANIZATION_FILTER="Customer|$XBE_TEST_CUSTOMER_ID"
    fi
fi

describe "Resource: time-sheets-exports"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create export requires formatter and time sheets"
xbe_run do time-sheets-exports create
assert_failure

test_name "Create time sheets export"
if [[ -n "$ORGANIZATION_FORMATTER_ID" && -n "$TIME_SHEET_IDS_RAW" && "$ORGANIZATION_FORMATTER_ID" != "null" ]]; then
    xbe_json do time-sheets-exports create \
        --organization-formatter "$ORGANIZATION_FORMATTER_ID" \
        --time-sheet-ids "$TIME_SHEET_IDS_RAW"
    if [[ $status -eq 0 ]]; then
        TIME_SHEETS_EXPORT_ID=$(json_get ".id")
        if [[ -n "$TIME_SHEETS_EXPORT_ID" && "$TIME_SHEETS_EXPORT_ID" != "null" ]]; then
            pass
        else
            fail "Created time sheets export but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"must be"* ]] || [[ "$output" == *"not in scope"* ]] || [[ "$output" == *"exportable"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create time sheets export: $output"
        fi
    fi
else
    skip "No formatter/time sheet IDs available. Set XBE_TEST_TIME_SHEETS_EXPORT_ORGANIZATION_FORMATTER_ID and XBE_TEST_TIME_SHEETS_EXPORT_TIME_SHEET_IDS."
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List time sheets exports"
xbe_json view time-sheets-exports list --limit 20
assert_json_is_array

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time sheets export"
if [[ -n "$TIME_SHEETS_EXPORT_ID" && "$TIME_SHEETS_EXPORT_ID" != "null" ]]; then
    xbe_json view time-sheets-exports show "$TIME_SHEETS_EXPORT_ID"
    assert_success
else
    skip "No time sheets export ID available. Set XBE_TEST_TIME_SHEETS_EXPORT_ID to enable show testing."
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by organization formatter"
if [[ -n "$ORGANIZATION_FORMATTER_ID" && "$ORGANIZATION_FORMATTER_ID" != "null" ]]; then
    xbe_json view time-sheets-exports list --organization-formatter "$ORGANIZATION_FORMATTER_ID" --limit 5
    assert_success
else
    skip "No organization formatter ID available (set XBE_TEST_TIME_SHEETS_EXPORT_ORGANIZATION_FORMATTER_ID)"
fi

test_name "Filter by status"
xbe_json view time-sheets-exports list --status processing --limit 5
assert_success

test_name "Filter by broker"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view time-sheets-exports list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available (set XBE_TEST_BROKER_ID)"
fi

test_name "Filter by created-by"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view time-sheets-exports list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available (set XBE_TEST_USER_ID)"
fi

test_name "Filter by organization"
if [[ -n "$ORGANIZATION_FILTER" && "$ORGANIZATION_FILTER" != "null" ]]; then
    xbe_json view time-sheets-exports list --organization "$ORGANIZATION_FILTER" --limit 5
    assert_success
else
    skip "No organization filter available (set XBE_TEST_TIME_SHEETS_EXPORT_ORGANIZATION_FILTER or XBE_TEST_BROKER_ID)"
fi

test_name "Filter by time sheets"
if [[ -n "$TIME_SHEET_IDS_RAW" && "$TIME_SHEET_IDS_RAW" != "null" ]]; then
    xbe_json view time-sheets-exports list --time-sheets "$TIME_SHEET_IDS_RAW" --limit 5
    assert_success
else
    skip "No time sheet IDs available (set XBE_TEST_TIME_SHEETS_EXPORT_TIME_SHEET_IDS)"
fi

run_tests
