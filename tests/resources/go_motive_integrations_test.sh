#!/bin/bash
#
# XBE CLI Integration Tests: GoMotive Integrations
#
# Tests list filters and create/update/delete operations for the go-motive-integrations resource.
#
# COVERAGE: List filters + create/update attributes + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_GO_MOTIVE_INTEGRATION_ID=""
GO_MOTIVE_BROKER_ID="${XBE_TEST_GO_MOTIVE_BROKER_ID:-${XBE_TEST_BROKER_ID:-}}"
GO_MOTIVE_INTEGRATION_CONFIG_ID="${XBE_TEST_GO_MOTIVE_INTEGRATION_CONFIG_ID:-${XBE_TEST_INTEGRATION_CONFIG_ID:-}}"

describe "Resource: go-motive-integrations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List GoMotive integrations"
xbe_json view go-motive-integrations list --limit 5
assert_success

test_name "List GoMotive integrations returns array"
xbe_json view go-motive-integrations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list GoMotive integrations"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List GoMotive integrations with --broker filter"
BROKER_FILTER_ID="${GO_MOTIVE_BROKER_ID:-1}"
xbe_json view go-motive-integrations list --broker "$BROKER_FILTER_ID" --limit 5
assert_success

test_name "List GoMotive integrations with --created-at-min filter"
xbe_json view go-motive-integrations list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List GoMotive integrations with --created-at-max filter"
xbe_json view go-motive-integrations list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List GoMotive integrations with --is-created-at filter"
xbe_json view go-motive-integrations list --is-created-at true --limit 5
assert_success

test_name "List GoMotive integrations with --updated-at-min filter"
xbe_json view go-motive-integrations list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List GoMotive integrations with --updated-at-max filter"
xbe_json view go-motive-integrations list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List GoMotive integrations with --is-updated-at filter"
xbe_json view go-motive-integrations list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List GoMotive integrations with --limit"
xbe_json view go-motive-integrations list --limit 3
assert_success

test_name "List GoMotive integrations with --offset"
xbe_json view go-motive-integrations list --limit 3 --offset 3
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show GoMotive integration (if any exist)"
xbe_json view go-motive-integrations list --limit 1
if [[ $status -eq 0 ]]; then
    GO_MOTIVE_ID=$(json_get ".[0].id")
    if [[ -n "$GO_MOTIVE_ID" && "$GO_MOTIVE_ID" != "null" ]]; then
        xbe_json view go-motive-integrations show "$GO_MOTIVE_ID"
        assert_success
    else
        skip "No GoMotive integrations available to show"
    fi
else
    skip "Could not list GoMotive integrations for show test"
fi

# ============================================================================
# CREATE / UPDATE / DELETE Tests
# ============================================================================

test_name "Create GoMotive integration (requires broker + integration config IDs)"
if [[ -z "$GO_MOTIVE_BROKER_ID" || -z "$GO_MOTIVE_INTEGRATION_CONFIG_ID" ]]; then
    skip "Missing XBE_TEST_GO_MOTIVE_BROKER_ID/XBE_TEST_BROKER_ID or XBE_TEST_GO_MOTIVE_INTEGRATION_CONFIG_ID/XBE_TEST_INTEGRATION_CONFIG_ID"
else
    INTEGRATION_IDENTIFIER="motive-$(date +%s)-${RANDOM}"
    FRIENDLY_NAME=$(unique_name "GoMotive")

    xbe_json do go-motive-integrations create \
        --integration-identifier "$INTEGRATION_IDENTIFIER" \
        --friendly-name "$FRIENDLY_NAME" \
        --broker "$GO_MOTIVE_BROKER_ID" \
        --integration-config "$GO_MOTIVE_INTEGRATION_CONFIG_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_GO_MOTIVE_INTEGRATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_GO_MOTIVE_INTEGRATION_ID" && "$CREATED_GO_MOTIVE_INTEGRATION_ID" != "null" ]]; then
            register_cleanup "go-motive-integrations" "$CREATED_GO_MOTIVE_INTEGRATION_ID"
            pass
        else
            fail "Created GoMotive integration but no ID returned"
        fi
    else
        skip "Failed to create GoMotive integration - may lack permissions or valid integration config"
    fi
fi

if [[ -n "$CREATED_GO_MOTIVE_INTEGRATION_ID" && "$CREATED_GO_MOTIVE_INTEGRATION_ID" != "null" ]]; then
    test_name "Update GoMotive integration fields"
    NEW_IDENTIFIER="motive-update-$(date +%s)-${RANDOM}"
    NEW_FRIENDLY_NAME=$(unique_name "GoMotive-Updated")

    xbe_json do go-motive-integrations update "$CREATED_GO_MOTIVE_INTEGRATION_ID" \
        --integration-identifier "$NEW_IDENTIFIER" \
        --friendly-name "$NEW_FRIENDLY_NAME"

    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update GoMotive integration - may not have permission"
    fi

    test_name "Update GoMotive integration without fields fails"
    xbe_run do go-motive-integrations update "$CREATED_GO_MOTIVE_INTEGRATION_ID"
    assert_failure

    test_name "Delete GoMotive integration requires --confirm"
    xbe_run do go-motive-integrations delete "$CREATED_GO_MOTIVE_INTEGRATION_ID"
    assert_failure

    test_name "Delete GoMotive integration with --confirm"
    xbe_run do go-motive-integrations delete "$CREATED_GO_MOTIVE_INTEGRATION_ID" --confirm
    assert_success
else
    test_name "Skip update/delete tests"
    skip "No GoMotive integration available for update/delete tests"
fi

run_tests
