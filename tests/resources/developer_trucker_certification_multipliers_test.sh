#!/bin/bash
#
# XBE CLI Integration Tests: Developer Trucker Certification Multipliers
#
# Tests CRUD operations for the developer_trucker_certification_multipliers resource.
# Multipliers link trailers to developer trucker certifications.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MULTIPLIER_ID=""

DTC_ID="${XBE_TEST_DEVELOPER_TRUCKER_CERTIFICATION_ID:-}"
TRAILER_ID="${XBE_TEST_TRAILER_ID:-}"

CREATE_READY=true
if [[ -z "$DTC_ID" ]]; then
    CREATE_READY=false
fi
if [[ -z "$TRAILER_ID" ]]; then
    CREATE_READY=false
fi

describe "Resource: developer_trucker_certification_multipliers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List developer trucker certification multipliers"
xbe_json view developer-trucker-certification-multipliers list --limit 5
assert_success

test_name "List developer trucker certification multipliers returns array"
xbe_json view developer-trucker-certification-multipliers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list developer trucker certification multipliers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List multipliers with --developer-trucker-certification filter"
if [[ -n "$DTC_ID" ]]; then
    xbe_json view developer-trucker-certification-multipliers list --developer-trucker-certification "$DTC_ID" --limit 10
    assert_success
else
    skip "Set XBE_TEST_DEVELOPER_TRUCKER_CERTIFICATION_ID to test developer trucker certification filter"
fi

test_name "List multipliers with --trailer filter"
if [[ -n "$TRAILER_ID" ]]; then
    xbe_json view developer-trucker-certification-multipliers list --trailer "$TRAILER_ID" --limit 10
    assert_success
else
    skip "Set XBE_TEST_TRAILER_ID to test trailer filter"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create developer trucker certification multiplier with required fields"
if [[ "$CREATE_READY" == true ]]; then
    xbe_json do developer-trucker-certification-multipliers create \
        --developer-trucker-certification "$DTC_ID" \
        --trailer "$TRAILER_ID" \
        --multiplier 0.85

    if [[ $status -eq 0 ]]; then
        CREATED_MULTIPLIER_ID=$(json_get ".id")
        if [[ -n "$CREATED_MULTIPLIER_ID" && "$CREATED_MULTIPLIER_ID" != "null" ]]; then
            register_cleanup "developer-trucker-certification-multipliers" "$CREATED_MULTIPLIER_ID"
            pass
        else
            fail "Created multiplier but no ID returned"
        fi
    else
        fail "Failed to create developer trucker certification multiplier"
    fi
else
    skip "Set XBE_TEST_DEVELOPER_TRUCKER_CERTIFICATION_ID and XBE_TEST_TRAILER_ID to run create tests"
fi

if [[ -z "$CREATED_MULTIPLIER_ID" || "$CREATED_MULTIPLIER_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer trucker certification multiplier ID"
else
    # ============================================================================
    # UPDATE Tests
    # ============================================================================

    test_name "Update multiplier value"
    xbe_json do developer-trucker-certification-multipliers update "$CREATED_MULTIPLIER_ID" --multiplier 0.9
    assert_success

    test_name "Update multiplier trailer"
    xbe_json do developer-trucker-certification-multipliers update "$CREATED_MULTIPLIER_ID" --trailer "$TRAILER_ID"
    assert_success

    test_name "Update without any fields fails"
    xbe_json do developer-trucker-certification-multipliers update "$CREATED_MULTIPLIER_ID"
    assert_failure

    # ============================================================================
    # DELETE Tests
    # ============================================================================

    test_name "Delete multiplier requires --confirm flag"
    xbe_run do developer-trucker-certification-multipliers delete "$CREATED_MULTIPLIER_ID"
    assert_failure

    test_name "Delete multiplier with --confirm"
    xbe_run do developer-trucker-certification-multipliers delete "$CREATED_MULTIPLIER_ID" --confirm
    assert_success
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create multiplier without developer trucker certification fails"
if [[ -n "$TRAILER_ID" ]]; then
    xbe_json do developer-trucker-certification-multipliers create \
        --trailer "$TRAILER_ID" \
        --multiplier 0.5
    assert_failure
else
    skip "Set XBE_TEST_TRAILER_ID to run create error tests"
fi

test_name "Create multiplier without trailer fails"
if [[ -n "$DTC_ID" ]]; then
    xbe_json do developer-trucker-certification-multipliers create \
        --developer-trucker-certification "$DTC_ID" \
        --multiplier 0.5
    assert_failure
else
    skip "Set XBE_TEST_DEVELOPER_TRUCKER_CERTIFICATION_ID to run create error tests"
fi

test_name "Create multiplier without multiplier fails"
if [[ -n "$DTC_ID" && -n "$TRAILER_ID" ]]; then
    xbe_json do developer-trucker-certification-multipliers create \
        --developer-trucker-certification "$DTC_ID" \
        --trailer "$TRAILER_ID"
    assert_failure
else
    skip "Set XBE_TEST_DEVELOPER_TRUCKER_CERTIFICATION_ID and XBE_TEST_TRAILER_ID to run create error tests"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
