#!/bin/bash
#
# XBE CLI Integration Tests: Missing Rates
#
# Tests list/show/create operations for missing-rates.
#
# COVERAGE: List + created/updated filters + show + create required flags
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_ID=""
SAMPLE_STUOM_ID=""
SAMPLE_CURRENCY_CODE=""
SAMPLE_CUSTOMER_PPU=""
SAMPLE_TRUCKER_PPU=""

JOB_ID="${XBE_TEST_JOB_ID:-}"
STUOM_ID="${XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID:-}"

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: missing-rates"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List missing rates"
xbe_json view missing-rates list --limit 5
assert_success

test_name "List missing rates returns array"
xbe_json view missing-rates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list missing rates"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample missing rate"
xbe_json view missing-rates list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_JOB_ID=$(json_get ".[0].job_id")
    SAMPLE_STUOM_ID=$(json_get ".[0].service_type_unit_of_measure_id")
    SAMPLE_CURRENCY_CODE=$(json_get ".[0].currency_code")
    SAMPLE_CUSTOMER_PPU=$(json_get ".[0].customer_price_per_unit")
    SAMPLE_TRUCKER_PPU=$(json_get ".[0].trucker_price_per_unit")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No missing rates available for follow-on tests"
    fi
else
    skip "Could not list missing rates to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List missing rates with --created-at-min"
xbe_json view missing-rates list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List missing rates with --created-at-max"
xbe_json view missing-rates list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List missing rates with --is-created-at=true"
xbe_json view missing-rates list --is-created-at true --limit 5
assert_success

test_name "List missing rates with --is-created-at=false"
xbe_json view missing-rates list --is-created-at false --limit 5
assert_success

test_name "List missing rates with --updated-at-min"
xbe_json view missing-rates list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List missing rates with --updated-at-max"
xbe_json view missing-rates list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List missing rates with --is-updated-at=true"
xbe_json view missing-rates list --is-updated-at true --limit 5
assert_success

test_name "List missing rates with --is-updated-at=false"
xbe_json view missing-rates list --is-updated-at false --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show missing rate"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view missing-rates show "$SAMPLE_ID"
    assert_success
else
    skip "No sample missing rate ID available"
fi

# ============================================================================
# CREATE Tests - Required Flags
# ============================================================================

test_name "Create missing rate requires --job"
xbe_run do missing-rates create \
    --service-type-unit-of-measure 1 \
    --currency-code USD \
    --customer-price-per-unit 1.00 \
    --trucker-price-per-unit 1.00
assert_failure

test_name "Create missing rate requires --service-type-unit-of-measure"
xbe_run do missing-rates create \
    --job 1 \
    --currency-code USD \
    --customer-price-per-unit 1.00 \
    --trucker-price-per-unit 1.00
assert_failure

test_name "Create missing rate requires --currency-code"
xbe_run do missing-rates create \
    --job 1 \
    --service-type-unit-of-measure 1 \
    --customer-price-per-unit 1.00 \
    --trucker-price-per-unit 1.00
assert_failure

test_name "Create missing rate requires --customer-price-per-unit"
xbe_run do missing-rates create \
    --job 1 \
    --service-type-unit-of-measure 1 \
    --currency-code USD \
    --trucker-price-per-unit 1.00
assert_failure

test_name "Create missing rate requires --trucker-price-per-unit"
xbe_run do missing-rates create \
    --job 1 \
    --service-type-unit-of-measure 1 \
    --currency-code USD \
    --customer-price-per-unit 1.00
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create missing rate"
if [[ -n "$JOB_ID" && "$JOB_ID" != "null" && -n "$STUOM_ID" && "$STUOM_ID" != "null" ]]; then
    xbe_json do missing-rates create \
        --job "$JOB_ID" \
        --service-type-unit-of-measure "$STUOM_ID" \
        --currency-code USD \
        --customer-price-per-unit 100.00 \
        --trucker-price-per-unit 85.00
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Created missing rate but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Validation"* ]] || [[ "$output" == *"unprocessable"* ]] || [[ "$output" == *"already related"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create missing rate: $output"
        fi
    fi
else
    skip "Missing prerequisites. Set XBE_TEST_JOB_ID and XBE_TEST_SERVICE_TYPE_UNIT_OF_MEASURE_ID to enable create testing."
fi

run_tests
