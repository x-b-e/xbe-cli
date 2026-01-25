#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheets
#
# Tests CRUD operations for the time-sheets resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_WORK_ORDER_ID=""
CREATED_TIME_SHEET_ID=""
DRIVER_ID=""

TEST_START_AT="2026-01-01T08:00:00Z"
TEST_END_AT="2026-01-01T16:00:00Z"
UPDATED_START_AT="2026-01-01T09:00:00Z"
UPDATED_END_AT="2026-01-01T17:00:00Z"

describe "Resource: time-sheets"

# ============================================================================
# Prerequisites - Create broker, business unit, work order
# ============================================================================

test_name "Create prerequisite broker for time sheet tests"
BROKER_NAME=$(unique_name "TimeSheetBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite business unit for time sheet tests"
BUSINESS_UNIT_NAME=$(unique_name "TimeSheetBU")

xbe_json do business-units create \
    --name "$BUSINESS_UNIT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business unit but no ID returned"
        echo "Cannot continue without a business unit"
        run_tests
    fi
else
    fail "Failed to create business unit"
    echo "Cannot continue without a business unit"
    run_tests
fi

test_name "Create work order for time sheet tests"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_WORK_ORDER_ID=$(json_get ".id")
    if [[ -n "$CREATED_WORK_ORDER_ID" && "$CREATED_WORK_ORDER_ID" != "null" ]]; then
        register_cleanup "work-orders" "$CREATED_WORK_ORDER_ID"
        pass
    else
        fail "Created work order but no ID returned"
        echo "Cannot continue without a work order"
        run_tests
    fi
else
    fail "Failed to create work order"
    echo "Cannot continue without a work order"
    run_tests
fi

test_name "Fetch current user for driver relationship"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    DRIVER_ID=$(json_get ".id")
    if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
        pass
    else
        fail "Auth whoami returned no ID"
        echo "Cannot continue without a driver ID"
        run_tests
    fi
else
    fail "Failed to fetch current user"
    echo "Cannot continue without a driver ID"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet with required and optional fields"
xbe_json do time-sheets create \
    --subject-type WorkOrder \
    --subject-id "$CREATED_WORK_ORDER_ID" \
    --driver "$DRIVER_ID" \
    --start-at "$TEST_START_AT" \
    --end-at "$TEST_END_AT" \
    --break-minutes 30 \
    --notes "Test time sheet"

if [[ $status -eq 0 ]]; then
    CREATED_TIME_SHEET_ID=$(json_get ".id")
    if [[ -n "$CREATED_TIME_SHEET_ID" && "$CREATED_TIME_SHEET_ID" != "null" ]]; then
        register_cleanup "time-sheets" "$CREATED_TIME_SHEET_ID"
        pass
    else
        fail "Created time sheet but no ID returned"
    fi
else
    fail "Failed to create time sheet"
fi

if [[ -z "$CREATED_TIME_SHEET_ID" || "$CREATED_TIME_SHEET_ID" == "null" ]]; then
    echo "Cannot continue without a valid time sheet ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update time sheet attributes"
xbe_json do time-sheets update "$CREATED_TIME_SHEET_ID" \
    --start-at "$UPDATED_START_AT" \
    --end-at "$UPDATED_END_AT" \
    --break-minutes 45 \
    --notes "Updated time sheet" \
    --skip-validate-overlap
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time sheet"
xbe_json view time-sheets show "$CREATED_TIME_SHEET_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheets"
xbe_json view time-sheets list --limit 5
assert_success

test_name "List time sheets returns array"
xbe_json view time-sheets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time sheets"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List time sheets with --subject-type and --subject-id"
xbe_json view time-sheets list --subject-type WorkOrder --subject-id "$CREATED_WORK_ORDER_ID" --limit 10
assert_success

test_name "List time sheets with --start-at-min"
xbe_json view time-sheets list --start-at-min "2026-01-01T00:00:00Z" --limit 10
assert_success

test_name "List time sheets with --start-at-max"
xbe_json view time-sheets list --start-at-max "2026-01-31T23:59:59Z" --limit 10
assert_success

test_name "List time sheets with --end-at-min"
xbe_json view time-sheets list --end-at-min "2026-01-01T00:00:00Z" --limit 10
assert_success

test_name "List time sheets with --end-at-max"
xbe_json view time-sheets list --end-at-max "2026-01-31T23:59:59Z" --limit 10
assert_success

test_name "List time sheets with --start-on-min"
xbe_json view time-sheets list --start-on-min "2026-01-01" --limit 10
assert_success

test_name "List time sheets with --start-on-max"
xbe_json view time-sheets list --start-on-max "2026-01-31" --limit 10
assert_success

test_name "List time sheets with --broker"
xbe_json view time-sheets list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List time sheets with --trucker"
xbe_json view time-sheets list --trucker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List time sheets with --driver"
xbe_json view time-sheets list --driver "$DRIVER_ID" --limit 10
assert_success

test_name "List time sheets with --laborer"
xbe_json view time-sheets list --laborer "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List time sheets with --laborer-user"
xbe_json view time-sheets list --laborer-user "$DRIVER_ID" --limit 10
assert_success

test_name "List time sheets with --equipment"
xbe_json view time-sheets list --equipment "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List time sheets with --has-exports"
xbe_json view time-sheets list --has-exports "false" --limit 10
assert_success

test_name "List time sheets with --status"
xbe_json view time-sheets list --status "editing" --limit 10
assert_success

test_name "List time sheets with --missing-craft-class"
xbe_json view time-sheets list --missing-craft-class "true" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List time sheets with --limit"
xbe_json view time-sheets list --limit 3
assert_success

test_name "List time sheets with --offset"
xbe_json view time-sheets list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete time sheet"
xbe_run do time-sheets delete "$CREATED_TIME_SHEET_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
