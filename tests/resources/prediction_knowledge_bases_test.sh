#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Knowledge Bases
#
# Tests list, show, and create operations for the prediction-knowledge-bases resource.
#
# COVERAGE: List filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_KB_ID=""
SAMPLE_ID=""
SAMPLE_BROKER_ID=""
LIST_SUPPORTED="false"
BROKER_CREATED="false"

describe "Resource: prediction-knowledge-bases"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction knowledge bases"
xbe_json view prediction-knowledge-bases list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Prediction knowledge bases list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List prediction knowledge bases returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view prediction-knowledge-bases list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list prediction knowledge bases"
    fi
else
    skip "Prediction knowledge bases list endpoint not available"
fi

# ==========================================================================
# Sample Record (used for show)
# ==========================================================================

test_name "Capture sample prediction knowledge base"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view prediction-knowledge-bases list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No prediction knowledge bases available for show"
        fi
    else
        skip "Could not list prediction knowledge bases to capture sample"
    fi
else
    skip "Prediction knowledge bases list endpoint not available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction knowledge bases with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view prediction-knowledge-bases list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter test"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction knowledge base"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-knowledge-bases show "$SAMPLE_ID"
    assert_success
else
    skip "No prediction knowledge base ID available"
fi

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "PredictionKnowledgeBaseBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        BROKER_CREATED="true"
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

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create prediction knowledge base without required broker fails"
xbe_run do prediction-knowledge-bases create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction knowledge base"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json do prediction-knowledge-bases create --broker "$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_KB_ID=$(json_get ".id")
        if [[ -n "$CREATED_KB_ID" && "$CREATED_KB_ID" != "null" ]]; then
            assert_json_equals ".broker_id" "$CREATED_BROKER_ID"
        else
            fail "Created prediction knowledge base but no ID returned"
        fi
    else
        if [[ "$BROKER_CREATED" != "true" ]]; then
            skip "Create failed (broker may already have a knowledge base)"
        else
            fail "Failed to create prediction knowledge base"
        fi
    fi
else
    skip "No broker available for create"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
