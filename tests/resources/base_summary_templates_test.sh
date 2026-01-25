#!/bin/bash
#
# XBE CLI Integration Tests: Base Summary Templates
#
# Tests create/delete and view operations for base summary templates.
#
# COVERAGE: All create attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TEMPLATE_ID=""
CREATED_BY_ID=""

describe "Resource: base-summary-templates"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for base summary template tests"
BROKER_NAME=$(unique_name "BaseSummaryTemplateBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
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

# Resolve current user for created-by filter tests
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CREATED_BY_ID=$(json_get ".id")
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create base summary template with all attributes"
TEMPLATE_LABEL=$(unique_name "BaseSummaryTemplate")
CREATE_CMD=(do base-summary-templates create \
    --label "$TEMPLATE_LABEL" \
    --group-bys broker,customer \
    --explicit-metrics count,total_cost \
    --filters "{\"broker\":\"$CREATED_BROKER_ID\"}" \
    --start-date "2025-01-01" \
    --end-date "2025-01-31" \
    --broker "$CREATED_BROKER_ID")

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    CREATE_CMD+=(--created-by "$CREATED_BY_ID")
fi

xbe_json "${CREATE_CMD[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_TEMPLATE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TEMPLATE_ID" && "$CREATED_TEMPLATE_ID" != "null" ]]; then
        register_cleanup "base-summary-templates" "$CREATED_TEMPLATE_ID"
        pass
    else
        fail "Created base summary template but no ID returned"
    fi
else
    fail "Failed to create base summary template"
fi

if [[ -z "$CREATED_TEMPLATE_ID" || "$CREATED_TEMPLATE_ID" == "null" ]]; then
    echo "Cannot continue without a valid template ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show base summary template by ID"
xbe_json view base-summary-templates show "$CREATED_TEMPLATE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List base summary templates"
xbe_json view base-summary-templates list --limit 5
assert_success

test_name "List base summary templates returns array"
xbe_json view base-summary-templates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list base summary templates"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List base summary templates with --label filter"
xbe_json view base-summary-templates list --label "$TEMPLATE_LABEL" --limit 5
assert_success

test_name "List base summary templates with --broker filter"
xbe_json view base-summary-templates list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "List base summary templates with --created-by filter"
    xbe_json view base-summary-templates list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete base summary template requires --confirm flag"
xbe_run do base-summary-templates delete "$CREATED_TEMPLATE_ID"
assert_failure

test_name "Delete base summary template with --confirm"
DEL_LABEL=$(unique_name "BaseSummaryTemplateDel")
DEL_CMD=(do base-summary-templates create --label "$DEL_LABEL" --broker "$CREATED_BROKER_ID")
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    DEL_CMD+=(--created-by "$CREATED_BY_ID")
fi

xbe_json "${DEL_CMD[@]}"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do base-summary-templates delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create base summary template for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create base summary template without label fails"
xbe_json do base-summary-templates create --broker "$CREATED_BROKER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
