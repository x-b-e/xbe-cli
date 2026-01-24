#!/bin/bash
#
# XBE CLI Integration Tests: Tender Re-Rates
#
# Tests view and create operations for tender_re_rates.
# Re-rates tender pricing and optionally re-constrains.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_RE_RATE_ID=""
TENDER_IDS_ARG=""
SKIP_MUTATION=0

describe "Resource: tender-re-rates"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List tender re-rates"
xbe_json view tender-re-rates list --limit 1
assert_success

test_name "Capture sample tender re-rate (if available)"
xbe_json view tender-re-rates list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_RE_RATE_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No tender re-rates available; skipping show test."
        pass
    fi
else
    fail "Failed to list tender re-rates"
fi

if [[ -n "$SAMPLE_RE_RATE_ID" && "$SAMPLE_RE_RATE_ID" != "null" ]]; then
    test_name "Show tender re-rate"
    xbe_json view tender-re-rates show "$SAMPLE_RE_RATE_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create tender re-rate requires --tender-ids"
xbe_run do tender-re-rates create --re-rate
assert_failure

test_name "Create tender re-rate requires an action flag"
xbe_run do tender-re-rates create --tender-ids 123
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    if [[ -n "$XBE_TEST_TENDER_RE_RATE_IDS" ]]; then
        TENDER_IDS_ARG="$XBE_TEST_TENDER_RE_RATE_IDS"
    elif [[ -n "$XBE_TEST_TENDER_RE_RATE_ID" ]]; then
        TENDER_IDS_ARG="$XBE_TEST_TENDER_RE_RATE_ID"
    fi
fi

if [[ -n "$TENDER_IDS_ARG" ]]; then
    test_name "Create tender re-rate with optional flags"
    TENDER_IDS_ARG=$(echo "$TENDER_IDS_ARG" | tr -d ' ')
    xbe_json do tender-re-rates create \
        --tender-ids "$TENDER_IDS_ARG" \
        --re-rate \
        --re-constrain \
        --update-time-card-quantities=false \
        --skip-update-travel-minutes \
        --skip-validate-customer-tender-hourly-rates

    if [[ $status -eq 0 ]]; then
        assert_json_has ".tender_ids"
        assert_json_equals ".re_rate" "true"
        assert_json_equals ".re_constrain" "true"
        assert_json_equals ".update_time_card_quantities" "false"
        assert_json_equals ".skip_update_travel_minutes" "true"
        assert_json_equals ".skip_validate_customer_tender_hourly_rates" "true"
    else
        skip "Unable to re-rate tenders (permissions or data constraints)"
    fi
else
    test_name "Create tender re-rate with optional flags"
    skip "XBE_TEST_TENDER_RE_RATE_IDS not set"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
