#!/bin/bash
#
# XBE CLI Integration Tests: Organization Formatters
#
# Tests create/update operations and list filters for organization-formatters.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FORMATTER_ID=""
INACTIVE_FORMATTER_ID=""
CREATED_BROKER_ID=""

describe "Resource: organization-formatters"

# ============================================================================
# Prerequisites - Create broker for organization
# ============================================================================

test_name "Create prerequisite broker for organization formatter tests"
BROKER_NAME=$(unique_name "OrgFormatterBroker")

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

FORMATTER_FUNCTION='function format(lineItemsJson, timestamp) { return lineItemsJson; }'
FORMATTER_FUNCTION_ALT='function format(lineItemsJson, timestamp) { return "ok"; }'

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create organization formatter with required fields"
xbe_json do organization-formatters create \
    --formatter-type TimeSheetsExportFormatter \
    --organization "Broker|$CREATED_BROKER_ID" \
    --formatter-function "$FORMATTER_FUNCTION"

if [[ $status -eq 0 ]]; then
    CREATED_FORMATTER_ID=$(json_get ".id")
    if [[ -n "$CREATED_FORMATTER_ID" && "$CREATED_FORMATTER_ID" != "null" ]]; then
        pass
    else
        fail "Created formatter but no ID returned"
    fi
else
    fail "Failed to create organization formatter"
fi

# Only continue if we successfully created a formatter
if [[ -z "$CREATED_FORMATTER_ID" || "$CREATED_FORMATTER_ID" == "null" ]]; then
    echo "Cannot continue without a valid formatter ID"
    run_tests
fi

test_name "Create organization formatter with optional fields"
xbe_json do organization-formatters create \
    --formatter-type TimeSheetsExportFormatter \
    --organization "Broker|$CREATED_BROKER_ID" \
    --description "CLI Test Formatter" \
    --status inactive \
    --is-library \
    --mime-types "text/csv,application/json" \
    --formatter-function "$FORMATTER_FUNCTION"

if [[ $status -eq 0 ]]; then
    INACTIVE_FORMATTER_ID=$(json_get ".id")
    if [[ -n "$INACTIVE_FORMATTER_ID" && "$INACTIVE_FORMATTER_ID" != "null" ]]; then
        pass
    else
        fail "Created inactive formatter but no ID returned"
    fi
else
    fail "Failed to create organization formatter with optional fields"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization formatter"
xbe_json view organization-formatters show "$CREATED_FORMATTER_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update organization formatter description"
xbe_json do organization-formatters update "$CREATED_FORMATTER_ID" --description "Updated description"
assert_success

test_name "Update organization formatter status"
xbe_json do organization-formatters update "$CREATED_FORMATTER_ID" --status inactive
assert_success

test_name "Update organization formatter is-library"
xbe_json do organization-formatters update "$CREATED_FORMATTER_ID" --is-library
assert_success

test_name "Update organization formatter mime types"
xbe_json do organization-formatters update "$CREATED_FORMATTER_ID" --mime-types "text/csv"
assert_success

test_name "Update organization formatter function"
xbe_json do organization-formatters update "$CREATED_FORMATTER_ID" --formatter-function "$FORMATTER_FUNCTION_ALT"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List organization formatters"
xbe_json view organization-formatters list --limit 5
assert_success

test_name "List organization formatters returns array"
xbe_json view organization-formatters list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list organization formatters"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter organization formatters by formatter type"
xbe_json view organization-formatters list --formatter-type TimeSheetsExportFormatter --limit 20
assert_success

test_name "Filter organization formatters by organization"
xbe_json view organization-formatters list --organization "Broker|$CREATED_BROKER_ID" --limit 20
assert_success

test_name "Filter organization formatters by status"
xbe_json view organization-formatters list --status inactive --limit 20
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to filter by status"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create organization formatter without formatter type fails"
xbe_json do organization-formatters create \
    --organization "Broker|$CREATED_BROKER_ID" \
    --formatter-function "$FORMATTER_FUNCTION"
assert_failure

test_name "Create organization formatter without organization fails"
xbe_json do organization-formatters create \
    --formatter-type TimeSheetsExportFormatter \
    --formatter-function "$FORMATTER_FUNCTION"
assert_failure

test_name "Create organization formatter without formatter function fails"
xbe_json do organization-formatters create \
    --formatter-type TimeSheetsExportFormatter \
    --organization "Broker|$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do organization-formatters update "$CREATED_FORMATTER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
