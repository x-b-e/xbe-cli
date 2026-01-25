#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Payroll Certifications
#
# Tests CRUD operations for the time_card_payroll_certifications resource.
#
# COVERAGE: All filters + create/delete attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CERT_ID=""
SAMPLE_CERT_ID=""
SAMPLE_TIME_CARD_ID=""
SAMPLE_CREATED_BY_ID=""
SKIP_CREATE=0

describe "Resource: time-card-payroll-certifications"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card payroll certifications"
xbe_json view time-card-payroll-certifications list --limit 5
assert_success

test_name "List time card payroll certifications returns array"
xbe_json view time-card-payroll-certifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time card payroll certifications"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate time card payroll certification for filters"
xbe_json view time-card-payroll-certifications list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_CERT_ID=$(json_get ".[0].id")
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
        SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
        pass
    else
        if [[ -n "$XBE_TEST_TIME_CARD_PAYROLL_CERTIFICATION_ID" ]]; then
            xbe_json view time-card-payroll-certifications show "$XBE_TEST_TIME_CARD_PAYROLL_CERTIFICATION_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_CERT_ID=$(json_get ".id")
                SAMPLE_TIME_CARD_ID=$(json_get ".time_card_id")
                SAMPLE_CREATED_BY_ID=$(json_get ".created_by_id")
                pass
            else
                skip "Failed to load XBE_TEST_TIME_CARD_PAYROLL_CERTIFICATION_ID"
            fi
        else
            skip "No time card payroll certifications found. Set XBE_TEST_TIME_CARD_PAYROLL_CERTIFICATION_ID for filter tests."
        fi
    fi
else
    fail "Failed to list time card payroll certifications for filters"
fi

# ============================================================================
# Show Tests
# ============================================================================

if [[ -n "$SAMPLE_CERT_ID" && "$SAMPLE_CERT_ID" != "null" ]]; then
    test_name "Show time card payroll certification"
    xbe_json view time-card-payroll-certifications show "$SAMPLE_CERT_ID"
    assert_success
else
    test_name "Show time card payroll certification"
    skip "No sample certification available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    test_name "Filter by time card"
    xbe_json view time-card-payroll-certifications list --time-card "$SAMPLE_TIME_CARD_ID"
    assert_success
else
    test_name "Filter by time card"
    skip "No time card ID available"
fi

if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    test_name "Filter by created-by"
    xbe_json view time-card-payroll-certifications list --created-by "$SAMPLE_CREATED_BY_ID"
    assert_success
else
    test_name "Filter by created-by"
    skip "No created-by user ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

CREATE_TIME_CARD_ID="${XBE_TEST_TIME_CARD_PAYROLL_CERTIFICATION_TIME_CARD_ID:-}"

if [[ -n "$CREATE_TIME_CARD_ID" ]]; then
    test_name "Create time card payroll certification"
    xbe_json do time-card-payroll-certifications create \
        --time-card "$CREATE_TIME_CARD_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_CERT_ID=$(json_get ".id")
        if [[ -n "$CREATED_CERT_ID" && "$CREATED_CERT_ID" != "null" ]]; then
            register_cleanup "time-card-payroll-certifications" "$CREATED_CERT_ID"
            pass
        else
            fail "Created certification but no ID returned"
        fi
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Create rejected (validation or authorization)"
            SKIP_CREATE=1
        else
            fail "Failed to create time card payroll certification"
            SKIP_CREATE=1
        fi
    fi
else
    test_name "Create time card payroll certification"
    skip "Set XBE_TEST_TIME_CARD_PAYROLL_CERTIFICATION_TIME_CARD_ID to enable create tests"
    SKIP_CREATE=1
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create time card payroll certification without required fields fails"
xbe_run do time-card-payroll-certifications create
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_CERT_ID" && "$CREATED_CERT_ID" != "null" ]]; then
    test_name "Delete time card payroll certification requires --confirm flag"
    xbe_run do time-card-payroll-certifications delete "$CREATED_CERT_ID"
    assert_failure

    test_name "Delete time card payroll certification with --confirm"
    xbe_run do time-card-payroll-certifications delete "$CREATED_CERT_ID" --confirm
    assert_success
else
    test_name "Delete time card payroll certification requires --confirm flag"
    skip "No certification created"
    test_name "Delete time card payroll certification with --confirm"
    skip "No certification created"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
