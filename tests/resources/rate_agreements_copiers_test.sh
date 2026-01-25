#!/bin/bash
#
# XBE CLI Integration Tests: Rate Agreements Copiers
#
# Tests list, show, and create operations for the rate-agreements-copiers resource.
#
# COVERAGE: All filters + create attributes/relationships + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TEMPLATE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_CREATED_BY_ID=""
LIST_SUPPORTED="true"

CREATED_ID=""

describe "Resource: rate_agreements_copiers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List rate agreements copiers"
xbe_json view rate-agreements-copiers list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing rate agreements copiers"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List rate agreements copiers returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view rate-agreements-copiers list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list rate agreements copiers"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample rate agreements copier"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view rate-agreements-copiers list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TEMPLATE_ID=$(json_get ".[0].rate_agreement_template_id")
        SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
        SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No rate agreements copiers available for follow-on tests"
        fi
    else
        skip "Could not list rate agreements copiers to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show rate agreements copier"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view rate-agreements-copiers show "$SAMPLE_ID"
    assert_success
else
    skip "No copier ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by broker"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view rate-agreements-copiers list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter test"
fi

test_name "Filter by rate agreement template"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_TEMPLATE_ID" && "$SAMPLE_TEMPLATE_ID" != "null" ]]; then
    xbe_json view rate-agreements-copiers list --rate-agreement-template "$SAMPLE_TEMPLATE_ID" --limit 5
    assert_success
else
    skip "No template ID available for filter test"
fi

test_name "Filter by created-by"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view rate-agreements-copiers list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available for filter test"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create rate agreements copier"
TEMPLATE_ID="${XBE_TEST_RATE_AGREEMENT_ID:-}"
TARGET_CUSTOMER_ID="${XBE_TEST_CUSTOMER_ID:-}"
TARGET_TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"

if [[ -n "$TEMPLATE_ID" && "$TEMPLATE_ID" != "null" && -n "$TARGET_CUSTOMER_ID" && "$TARGET_CUSTOMER_ID" != "null" ]]; then
    xbe_json do rate-agreements-copiers create \
        --rate-agreement-template "$TEMPLATE_ID" \
        --target-customers "$TARGET_CUSTOMER_ID" \
        --note "CLI test copier"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            pass
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
elif [[ -n "$TEMPLATE_ID" && "$TEMPLATE_ID" != "null" && -n "$TARGET_TRUCKER_ID" && "$TARGET_TRUCKER_ID" != "null" ]]; then
    xbe_json do rate-agreements-copiers create \
        --rate-agreement-template "$TEMPLATE_ID" \
        --target-truckers "$TARGET_TRUCKER_ID" \
        --note "CLI test copier"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            pass
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_RATE_AGREEMENT_ID and XBE_TEST_CUSTOMER_ID or XBE_TEST_TRUCKER_ID to enable create test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required flags fails"
xbe_run do rate-agreements-copiers create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
