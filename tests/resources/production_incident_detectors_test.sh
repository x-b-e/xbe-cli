#!/bin/bash
#
# XBE CLI Integration Tests: Production Incident Detectors
#
# Tests create operations for production-incident-detectors.
#
# COVERAGE: Writable attributes (job-production-plan, lookahead-offset, minutes-threshold, quantity-threshold)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PLAN_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_ID:-}"
CREATED_DETECTOR_ID=""

describe "Resource: production-incident-detectors"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires job production plan"
xbe_run do production-incident-detectors create --minutes-threshold 30
assert_failure

test_name "Create detector run with custom thresholds"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json do production-incident-detectors create \
        --job-production-plan "$PLAN_ID" \
        --lookahead-offset 30 \
        --minutes-threshold 45 \
        --quantity-threshold 50
    if [[ $status -eq 0 ]]; then
        CREATED_DETECTOR_ID=$(json_get ".id")
        assert_json_has ".id"
        assert_json_equals ".job_production_plan_id" "$PLAN_ID"
        assert_json_equals ".lookahead_offset" "30"
        assert_json_equals ".minutes_threshold" "45"
        assert_json_equals ".quantity_threshold" "50"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create production incident detector: $output"
        fi
    fi
else
    skip "No job production plan ID available. Set XBE_TEST_JOB_PRODUCTION_PLAN_ID to enable create testing."
fi

# ============================================================================
# LIST/SHOW Tests
# ============================================================================

test_name "List production incident detectors"
xbe_json view production-incident-detectors list --limit 5
if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e 'type == "array"' >/dev/null; then
        pass
    else
        fail "List output was not a JSON array"
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"404"* ]] || [[ "$output" == *"Method Not Allowed"* ]]; then
        skip "List endpoint not available or not authorized"
    else
        fail "Failed to list production incident detectors: $output"
    fi
fi

test_name "Show production incident detector details"
if [[ -n "$CREATED_DETECTOR_ID" && "$CREATED_DETECTOR_ID" != "null" ]]; then
    xbe_json view production-incident-detectors show "$CREATED_DETECTOR_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".id" "$CREATED_DETECTOR_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"404"* ]] || [[ "$output" == *"Method Not Allowed"* ]]; then
            skip "Show endpoint not available or not authorized"
        else
            fail "Failed to show production incident detector: $output"
        fi
    fi
else
    skip "No detector ID available from create."
fi

run_tests
