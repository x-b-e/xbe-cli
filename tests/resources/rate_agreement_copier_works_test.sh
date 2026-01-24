#!/bin/bash
#
# XBE CLI Integration Tests: Rate Agreement Copier Works
#
# Tests list/show operations and optional create/update for the rate-agreement-copier-works resource.
#
# COVERAGE: List filters + show + create/update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

WORK_ID=""
TEMPLATE_ID=""
TARGET_TYPE=""
TARGET_ID=""
BROKER_ID=""
CREATED_BY_ID=""
SKIP_ID_FILTERS=0

ENV_TEMPLATE_ID="${XBE_TEST_RATE_AGREEMENT_TEMPLATE_ID:-}"
ENV_TARGET_TYPE="${XBE_TEST_TARGET_ORGANIZATION_TYPE:-}"
ENV_TARGET_ID="${XBE_TEST_TARGET_ORGANIZATION_ID:-}"

CREATED_WORK_ID=""

describe "Resource: rate-agreement-copier-works"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List rate agreement copier works"
xbe_json view rate-agreement-copier-works list --limit 5
assert_success

test_name "List rate agreement copier works returns array"
xbe_json view rate-agreement-copier-works list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list rate agreement copier works"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample rate agreement copier work"
xbe_json view rate-agreement-copier-works list --limit 1
if [[ $status -eq 0 ]]; then
    WORK_ID=$(json_get ".[0].id")
    TEMPLATE_ID=$(json_get ".[0].rate_agreement_template_id")
    TARGET_TYPE=$(json_get ".[0].target_organization_type")
    TARGET_ID=$(json_get ".[0].target_organization_id")
    BROKER_ID=$(json_get ".[0].broker_id")
    CREATED_BY_ID=$(json_get ".[0].created_by_id")
    if [[ -n "$WORK_ID" && "$WORK_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No rate agreement copier works available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list rate agreement copier works"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List copier works with --rate-agreement-template filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TEMPLATE_ID" && "$TEMPLATE_ID" != "null" ]]; then
    xbe_json view rate-agreement-copier-works list --rate-agreement-template "$TEMPLATE_ID" --limit 5
    assert_success
else
    skip "No rate agreement template ID available"
fi

test_name "List copier works with --target-organization filters"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TARGET_TYPE" && "$TARGET_TYPE" != "null" && -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json view rate-agreement-copier-works list \
        --target-organization-type "$TARGET_TYPE" \
        --target-organization-id "$TARGET_ID" \
        --limit 5
    assert_success
else
    skip "No target organization available"
fi

test_name "List copier works with --broker filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view rate-agreement-copier-works list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List copier works with --created-by filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view rate-agreement-copier-works list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show rate agreement copier work"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$WORK_ID" && "$WORK_ID" != "null" ]]; then
    xbe_json view rate-agreement-copier-works show "$WORK_ID"
    assert_success
else
    skip "No copier work ID available"
fi

# ============================================================================
# CREATE Tests - Error Cases
# ============================================================================

test_name "Create copier work without required fields fails"
xbe_run do rate-agreement-copier-works create
assert_failure

test_name "Create copier work without rate agreement template fails"
xbe_run do rate-agreement-copier-works create \
    --target-organization-type customers \
    --target-organization-id 1
assert_failure

test_name "Create copier work without target organization type fails"
xbe_run do rate-agreement-copier-works create \
    --rate-agreement-template 1 \
    --target-organization-id 1
assert_failure

test_name "Create copier work without target organization ID fails"
xbe_run do rate-agreement-copier-works create \
    --rate-agreement-template 1 \
    --target-organization-type customers
assert_failure

# ============================================================================
# CREATE + UPDATE Tests - Optional
# ============================================================================

test_name "Create rate agreement copier work"
if [[ -n "$ENV_TEMPLATE_ID" && -n "$ENV_TARGET_TYPE" && -n "$ENV_TARGET_ID" ]]; then
    NOTE_CREATE=$(unique_name "RateAgreementCopy")
    xbe_json do rate-agreement-copier-works create \
        --rate-agreement-template "$ENV_TEMPLATE_ID" \
        --target-organization-type "$ENV_TARGET_TYPE" \
        --target-organization-id "$ENV_TARGET_ID" \
        --note "$NOTE_CREATE"

    if [[ $status -eq 0 ]]; then
        CREATED_WORK_ID=$(json_get ".id")
        if [[ -n "$CREATED_WORK_ID" && "$CREATED_WORK_ID" != "null" ]]; then
            pass
        else
            fail "Created copier work but no ID returned"
        fi
    else
        skip "Create failed (check permissions or overlap constraints)"
    fi
else
    skip "Set XBE_TEST_RATE_AGREEMENT_TEMPLATE_ID and XBE_TEST_TARGET_ORGANIZATION_* to run"
fi

test_name "Update rate agreement copier work note"
if [[ -n "$CREATED_WORK_ID" && "$CREATED_WORK_ID" != "null" ]]; then
    NOTE_UPDATE=$(unique_name "RateAgreementCopyUpdate")
    xbe_json do rate-agreement-copier-works update "$CREATED_WORK_ID" --note "$NOTE_UPDATE"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".note" "$NOTE_UPDATE"
    else
        fail "Failed to update copier work note"
    fi
else
    skip "No created copier work ID available"
fi

# ============================================================================
# UPDATE Tests - Error Cases
# ============================================================================

test_name "Update copier work without attributes fails"
xbe_run do rate-agreement-copier-works update 999999
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
