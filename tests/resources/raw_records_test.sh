#!/bin/bash
#
# XBE CLI Integration Tests: Raw Records
#
# Tests list and show operations for raw-records.
#
# COVERAGE: All list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_RAW_RECORD_ID=""
SEED_EXTERNAL_RECORD_TYPE=""
SEED_EXTERNAL_RECORD_ID=""
SEED_INTERNAL_RECORD_TYPE=""
SEED_INTERNAL_RECORD_ID=""
SEED_BROKER_ID=""
SEED_INTEGRATION_CONFIG_ID=""
SEED_INTEGRATION_CONFIG_ORG_TYPE=""
SEED_INTEGRATION_CONFIG_ORG_ID=""

describe "Resource: raw-records"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List raw records"
xbe_json view raw-records list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SEED_RAW_RECORD_ID=$(json_get ".[0].id")
    SEED_EXTERNAL_RECORD_TYPE=$(json_get ".[0].external_record_type")
    SEED_EXTERNAL_RECORD_ID=$(json_get ".[0].external_record_id")
    SEED_INTERNAL_RECORD_TYPE=$(json_get ".[0].internal_record_type")
    SEED_INTERNAL_RECORD_ID=$(json_get ".[0].internal_record_id")
    SEED_BROKER_ID=$(json_get ".[0].broker_id")
    SEED_INTEGRATION_CONFIG_ID=$(json_get ".[0].integration_config_id")
else
    fail "Failed to list raw records"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show raw record"
if [[ -n "$SEED_RAW_RECORD_ID" && "$SEED_RAW_RECORD_ID" != "null" ]]; then
    xbe_json view raw-records show "$SEED_RAW_RECORD_ID"
    assert_success
    if [[ $status -eq 0 ]]; then
        if [[ -z "$SEED_INTERNAL_RECORD_TYPE" || "$SEED_INTERNAL_RECORD_TYPE" == "null" ]]; then
            SEED_INTERNAL_RECORD_TYPE=$(json_get ".internal_record_type")
        fi
        if [[ -z "$SEED_INTERNAL_RECORD_ID" || "$SEED_INTERNAL_RECORD_ID" == "null" ]]; then
            SEED_INTERNAL_RECORD_ID=$(json_get ".internal_record_id")
        fi
        if [[ -z "$SEED_EXTERNAL_RECORD_TYPE" || "$SEED_EXTERNAL_RECORD_TYPE" == "null" ]]; then
            SEED_EXTERNAL_RECORD_TYPE=$(json_get ".external_record_type")
        fi
        if [[ -z "$SEED_EXTERNAL_RECORD_ID" || "$SEED_EXTERNAL_RECORD_ID" == "null" ]]; then
            SEED_EXTERNAL_RECORD_ID=$(json_get ".external_record_id")
        fi
        SEED_INTEGRATION_CONFIG_ORG_TYPE=$(json_get ".integration_config_organization_type")
        SEED_INTEGRATION_CONFIG_ORG_ID=$(json_get ".integration_config_organization_id")
    fi
else
    skip "No raw record available to show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by broker"
if [[ -n "$SEED_BROKER_ID" && "$SEED_BROKER_ID" != "null" ]]; then
    xbe_json view raw-records list --broker "$SEED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

test_name "Filter by integration config"
if [[ -n "$SEED_INTEGRATION_CONFIG_ID" && "$SEED_INTEGRATION_CONFIG_ID" != "null" ]]; then
    xbe_json view raw-records list --integration-config "$SEED_INTEGRATION_CONFIG_ID" --limit 5
    assert_success
else
    skip "No integration config ID available for filter"
fi

test_name "Filter by integration config organization"
if [[ -n "$SEED_INTEGRATION_CONFIG_ORG_TYPE" && "$SEED_INTEGRATION_CONFIG_ORG_TYPE" != "null" && -n "$SEED_INTEGRATION_CONFIG_ORG_ID" && "$SEED_INTEGRATION_CONFIG_ORG_ID" != "null" ]]; then
    xbe_json view raw-records list --integration-config-organization "${SEED_INTEGRATION_CONFIG_ORG_TYPE}|${SEED_INTEGRATION_CONFIG_ORG_ID}" --limit 5
    assert_success
else
    skip "No integration config organization available for filter"
fi

test_name "Filter by internal record"
if [[ -n "$SEED_INTERNAL_RECORD_TYPE" && "$SEED_INTERNAL_RECORD_TYPE" != "null" && -n "$SEED_INTERNAL_RECORD_ID" && "$SEED_INTERNAL_RECORD_ID" != "null" ]]; then
    xbe_json view raw-records list --internal-record "${SEED_INTERNAL_RECORD_TYPE}|${SEED_INTERNAL_RECORD_ID}" --limit 5
    assert_success
else
    skip "No internal record available for filter"
fi

test_name "Filter by external record type"
if [[ -n "$SEED_EXTERNAL_RECORD_TYPE" && "$SEED_EXTERNAL_RECORD_TYPE" != "null" ]]; then
    xbe_json view raw-records list --external-record-type "$SEED_EXTERNAL_RECORD_TYPE" --limit 5
    assert_success
else
    skip "No external record type available for filter"
fi

test_name "Filter by external record ID"
if [[ -n "$SEED_EXTERNAL_RECORD_ID" && "$SEED_EXTERNAL_RECORD_ID" != "null" ]]; then
    xbe_json view raw-records list --external-record-id "$SEED_EXTERNAL_RECORD_ID" --limit 5
    assert_success
else
    skip "No external record ID available for filter"
fi

test_name "Filter by processed status"
xbe_json view raw-records list --is-processed true --limit 5
assert_success

test_name "Filter by failed status"
xbe_json view raw-records list --is-failed false --limit 5
assert_success

test_name "Filter by skipped status"
xbe_json view raw-records list --is-skipped false --limit 5
assert_success

test_name "Filter by data presence"
xbe_json view raw-records list --has-data true --limit 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
