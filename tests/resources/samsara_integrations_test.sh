#!/bin/bash
#
# XBE CLI Integration Tests: Samsara Integrations
#
# Tests CRUD operations for the samsara-integrations resource.
#
# COVERAGE: Create/update attributes + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SAMSARA_INTEGRATION_ID=""
BROKER_ID=""
INTEGRATION_CONFIG_ID=""

describe "Resource: samsara-integrations"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Resolve integration config and broker prerequisites"
if [[ -n "$XBE_TEST_SAMSARA_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_SAMSARA_INTEGRATION_CONFIG_ID"
elif [[ -n "$XBE_TEST_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_INTEGRATION_CONFIG_ID"
fi

if [[ -n "$XBE_TEST_SAMSARA_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_SAMSARA_BROKER_ID"
elif [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_BROKER_ID"
fi

if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    echo "    Using integration config: $INTEGRATION_CONFIG_ID"
    echo "    Using broker: $BROKER_ID"
    pass
else
    skip "Set XBE_TEST_SAMSARA_INTEGRATION_CONFIG_ID and XBE_TEST_SAMSARA_BROKER_ID (or XBE_TEST_INTEGRATION_CONFIG_ID/XBE_TEST_BROKER_ID) to enable create/update/delete tests"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create Samsara integration with required fields"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    INTEGRATION_IDENTIFIER="samsara-cli-$(date +%s)-${RANDOM}"
    FRIENDLY_NAME=$(unique_name "SamsaraIntegration")

    xbe_json do samsara-integrations create \
        --integration-identifier "$INTEGRATION_IDENTIFIER" \
        --friendly-name "$FRIENDLY_NAME" \
        --broker "$BROKER_ID" \
        --integration-config "$INTEGRATION_CONFIG_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_SAMSARA_INTEGRATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_SAMSARA_INTEGRATION_ID" && "$CREATED_SAMSARA_INTEGRATION_ID" != "null" ]]; then
            register_cleanup "samsara-integrations" "$CREATED_SAMSARA_INTEGRATION_ID"
            pass
        else
            fail "Created Samsara integration but no ID returned"
        fi
    else
        fail "Failed to create Samsara integration"
    fi
else
    skip "Missing prerequisites for create"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show Samsara integration"
SHOW_ID="$CREATED_SAMSARA_INTEGRATION_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    xbe_json view samsara-integrations list --limit 1
    if [[ $status -eq 0 ]]; then
        SHOW_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view samsara-integrations show "$SHOW_ID"
    assert_success
else
    skip "No Samsara integration ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update Samsara integration friendly name"
if [[ -n "$CREATED_SAMSARA_INTEGRATION_ID" && "$CREATED_SAMSARA_INTEGRATION_ID" != "null" ]]; then
    UPDATED_NAME=$(unique_name "UpdatedSamsaraIntegration")
    xbe_json do samsara-integrations update "$CREATED_SAMSARA_INTEGRATION_ID" --friendly-name "$UPDATED_NAME"
    assert_success
else
    skip "No Samsara integration ID available for update"
fi

test_name "Update Samsara integration identifier"
if [[ -n "$CREATED_SAMSARA_INTEGRATION_ID" && "$CREATED_SAMSARA_INTEGRATION_ID" != "null" ]]; then
    UPDATED_IDENTIFIER="samsara-cli-updated-$(date +%s)-${RANDOM}"
    xbe_json do samsara-integrations update "$CREATED_SAMSARA_INTEGRATION_ID" --integration-identifier "$UPDATED_IDENTIFIER"
    assert_success
else
    skip "No Samsara integration ID available for update"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List Samsara integrations"
xbe_json view samsara-integrations list --limit 5
assert_success

test_name "List Samsara integrations returns array"
xbe_json view samsara-integrations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list Samsara integrations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List Samsara integrations with --broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view samsara-integrations list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List Samsara integrations with --limit"
xbe_json view samsara-integrations list --limit 3
assert_success

test_name "List Samsara integrations with --offset"
xbe_json view samsara-integrations list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete Samsara integration requires --confirm flag"
if [[ -n "$CREATED_SAMSARA_INTEGRATION_ID" && "$CREATED_SAMSARA_INTEGRATION_ID" != "null" ]]; then
    xbe_run do samsara-integrations delete "$CREATED_SAMSARA_INTEGRATION_ID"
    assert_failure
else
    skip "No Samsara integration ID available for delete"
fi

test_name "Delete Samsara integration with --confirm"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    DEL_IDENTIFIER="samsara-cli-delete-$(date +%s)-${RANDOM}"
    DEL_NAME=$(unique_name "DeleteSamsaraIntegration")
    xbe_json do samsara-integrations create \
        --integration-identifier "$DEL_IDENTIFIER" \
        --friendly-name "$DEL_NAME" \
        --broker "$BROKER_ID" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do samsara-integrations delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create Samsara integration for deletion test"
    fi
else
    skip "Missing prerequisites for delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create Samsara integration without integration identifier fails"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    xbe_json do samsara-integrations create --friendly-name "Missing ID" --broker "$BROKER_ID" --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing prerequisites for error test"
fi

test_name "Create Samsara integration without friendly name fails"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    xbe_json do samsara-integrations create --integration-identifier "missing-name" --broker "$BROKER_ID" --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing prerequisites for error test"
fi

test_name "Update Samsara integration without any fields fails"
if [[ -n "$CREATED_SAMSARA_INTEGRATION_ID" && "$CREATED_SAMSARA_INTEGRATION_ID" != "null" ]]; then
    xbe_json do samsara-integrations update "$CREATED_SAMSARA_INTEGRATION_ID"
    assert_failure
else
    skip "No Samsara integration ID available for update error test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
