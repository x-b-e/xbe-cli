#!/bin/bash
#
# XBE CLI Integration Tests: Raw Material Transactions
#
# Tests list/show/update operations for the raw-material-transactions resource.
# Raw material transactions represent ingested ticket data before normalization.
#
# COVERAGE: List, show, update, filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

RAW_MATERIAL_TXN_ID=""
RAW_MATERIAL_SITE_ID=""


describe "Resource: raw-material-transactions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List raw material transactions"
xbe_json view raw-material-transactions list --limit 5
assert_success

test_name "List raw material transactions returns array"
xbe_json view raw-material-transactions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list raw material transactions"
fi

# ============================================================================
# SHOW / UPDATE Setup
# ============================================================================

test_name "Get a raw material transaction ID for show/update tests"
xbe_json view raw-material-transactions list --limit 1
if [[ $status -eq 0 ]]; then
    RAW_MATERIAL_TXN_ID=$(json_get ".[0].id")
    RAW_MATERIAL_SITE_ID=$(json_get ".[0].material_site_id")
    if [[ -n "$RAW_MATERIAL_TXN_ID" && "$RAW_MATERIAL_TXN_ID" != "null" ]]; then
        pass
    else
        skip "No raw material transactions found"
    fi
else
    fail "Failed to list raw material transactions"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$RAW_MATERIAL_TXN_ID" && "$RAW_MATERIAL_TXN_ID" != "null" ]]; then
    test_name "Show raw material transaction"
    xbe_json view raw-material-transactions show "$RAW_MATERIAL_TXN_ID"
    assert_success
else
    test_name "Show raw material transaction"
    skip "No raw material transactions available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$RAW_MATERIAL_TXN_ID" && "$RAW_MATERIAL_TXN_ID" != "null" ]]; then
    test_name "Update raw material transaction ticket-job-number"
    NEW_JOB_NUMBER="CLI-RAW-JOB-$(date +%s)"
    xbe_json do raw-material-transactions update "$RAW_MATERIAL_TXN_ID" --ticket-job-number "$NEW_JOB_NUMBER"
    if [[ $status -ne 0 && "$output" == *"must be an existing source"* ]]; then
        skip "Raw material transaction source is missing; skipping update"
    else
        assert_success
    fi

    test_name "Update raw material transaction without fields fails"
    xbe_json do raw-material-transactions update "$RAW_MATERIAL_TXN_ID"
    assert_failure
else
    test_name "Update raw material transaction ticket-job-number"
    skip "No raw material transactions available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

TODAY="$(date +%Y-%m-%d)"
NOW_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Date filters

test_name "List raw material transactions with --date filter"
xbe_json view raw-material-transactions list --date "$TODAY" --limit 5
assert_success

test_name "List raw material transactions with --date-min and --date-max"
xbe_json view raw-material-transactions list --date-min "$NOW_UTC" --date-max "$NOW_UTC" --limit 5
assert_success

# ID filters

test_name "List raw material transactions with --material-site"
if [[ -n "$RAW_MATERIAL_SITE_ID" && "$RAW_MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view raw-material-transactions list --material-site "$RAW_MATERIAL_SITE_ID" --limit 5
else
    xbe_json view raw-material-transactions list --material-site 1 --limit 5
fi
assert_success

test_name "List raw material transactions with --material-site-id"
if [[ -n "$RAW_MATERIAL_SITE_ID" && "$RAW_MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view raw-material-transactions list --material-site-id "$RAW_MATERIAL_SITE_ID" --limit 5
else
    xbe_json view raw-material-transactions list --material-site-id 1 --limit 5
fi
assert_success

test_name "List raw material transactions with --broker"
xbe_json view raw-material-transactions list --broker 1 --limit 5
assert_success

test_name "List raw material transactions with --broker-id"
xbe_json view raw-material-transactions list --broker-id 1 --limit 5
assert_success

# Attribute filters

test_name "List raw material transactions with --material-supplier-name"
xbe_json view raw-material-transactions list --material-supplier-name "Test" --limit 5
assert_success

test_name "List raw material transactions with --ticket-number"
xbe_json view raw-material-transactions list --ticket-number "T123" --limit 5
assert_success

test_name "List raw material transactions with --job-number"
xbe_json view raw-material-transactions list --job-number "J123" --limit 5
assert_success

test_name "List raw material transactions with --hauler-type"
xbe_json view raw-material-transactions list --hauler-type "internal" --limit 5
assert_success

test_name "List raw material transactions with --sales-customer-id"
xbe_json view raw-material-transactions list --sales-customer-id 1 --limit 5
assert_success

test_name "List raw material transactions with --truck-name"
xbe_json view raw-material-transactions list --truck-name "Truck" --limit 5
assert_success

test_name "List raw material transactions with --site-id"
xbe_json view raw-material-transactions list --site-id 1 --limit 5
assert_success

test_name "List raw material transactions with --material-name"
xbe_json view raw-material-transactions list --material-name "MAT" --limit 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
