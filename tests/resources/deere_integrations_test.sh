#!/bin/bash
#
# XBE CLI Integration Tests: Deere Integrations
#
# Tests CRUD operations for the deere-integrations resource.
#
# COVERAGE: Create/update attributes + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_DEERE_INTEGRATION_ID=""
BROKER_ID=""
INTEGRATION_CONFIG_ID=""

describe "Resource: deere-integrations"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Resolve integration config and broker prerequisites"
if [[ -n "$XBE_TEST_DEERE_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_DEERE_INTEGRATION_CONFIG_ID"
elif [[ -n "$XBE_TEST_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_INTEGRATION_CONFIG_ID"
fi

if [[ -n "$XBE_TEST_DEERE_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_DEERE_BROKER_ID"
elif [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_BROKER_ID"
fi

if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    echo "    Using integration config: $INTEGRATION_CONFIG_ID"
    echo "    Using broker: $BROKER_ID"
    pass
else
    skip "Set XBE_TEST_DEERE_INTEGRATION_CONFIG_ID and XBE_TEST_DEERE_BROKER_ID (or XBE_TEST_INTEGRATION_CONFIG_ID/XBE_TEST_BROKER_ID) to enable create/update/delete tests"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create Deere integration with required fields"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    INTEGRATION_IDENTIFIER="deere-cli-$(date +%s)-${RANDOM}"
    FRIENDLY_NAME=$(unique_name "DeereIntegration")

    xbe_json do deere-integrations create \
        --integration-identifier "$INTEGRATION_IDENTIFIER" \
        --friendly-name "$FRIENDLY_NAME" \
        --broker "$BROKER_ID" \
        --integration-config "$INTEGRATION_CONFIG_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_DEERE_INTEGRATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_DEERE_INTEGRATION_ID" && "$CREATED_DEERE_INTEGRATION_ID" != "null" ]]; then
            register_cleanup "deere-integrations" "$CREATED_DEERE_INTEGRATION_ID"
            pass
        else
            fail "Created Deere integration but no ID returned"
        fi
    else
        fail "Failed to create Deere integration"
    fi
else
    skip "Missing prerequisites for create"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show Deere integration"
SHOW_ID="$CREATED_DEERE_INTEGRATION_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    xbe_json view deere-integrations list --limit 1
    if [[ $status -eq 0 ]]; then
        SHOW_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view deere-integrations show "$SHOW_ID"
    assert_success
else
    skip "No Deere integration ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update Deere integration friendly name"
if [[ -n "$CREATED_DEERE_INTEGRATION_ID" && "$CREATED_DEERE_INTEGRATION_ID" != "null" ]]; then
    UPDATED_NAME=$(unique_name "UpdatedDeereIntegration")
    xbe_json do deere-integrations update "$CREATED_DEERE_INTEGRATION_ID" --friendly-name "$UPDATED_NAME"
    assert_success
else
    skip "No Deere integration ID available for update"
fi

test_name "Update Deere integration identifier"
if [[ -n "$CREATED_DEERE_INTEGRATION_ID" && "$CREATED_DEERE_INTEGRATION_ID" != "null" ]]; then
    UPDATED_IDENTIFIER="deere-cli-updated-$(date +%s)-${RANDOM}"
    xbe_json do deere-integrations update "$CREATED_DEERE_INTEGRATION_ID" --integration-identifier "$UPDATED_IDENTIFIER"
    assert_success
else
    skip "No Deere integration ID available for update"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List Deere integrations"
xbe_json view deere-integrations list --limit 5
assert_success

test_name "List Deere integrations returns array"
xbe_json view deere-integrations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list Deere integrations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List Deere integrations with --broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view deere-integrations list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List Deere integrations with --limit"
xbe_json view deere-integrations list --limit 3
assert_success

test_name "List Deere integrations with --offset"
xbe_json view deere-integrations list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete Deere integration requires --confirm flag"
if [[ -n "$CREATED_DEERE_INTEGRATION_ID" && "$CREATED_DEERE_INTEGRATION_ID" != "null" ]]; then
    xbe_run do deere-integrations delete "$CREATED_DEERE_INTEGRATION_ID"
    assert_failure
else
    skip "No Deere integration ID available for delete"
fi

test_name "Delete Deere integration with --confirm"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    DEL_IDENTIFIER="deere-cli-delete-$(date +%s)-${RANDOM}"
    DEL_NAME=$(unique_name "DeleteDeereIntegration")
    xbe_json do deere-integrations create \
        --integration-identifier "$DEL_IDENTIFIER" \
        --friendly-name "$DEL_NAME" \
        --broker "$BROKER_ID" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do deere-integrations delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create Deere integration for deletion test"
    fi
else
    skip "Missing prerequisites for delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create Deere integration without integration identifier fails"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    xbe_json do deere-integrations create --friendly-name "Missing ID" --broker "$BROKER_ID" --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing prerequisites for error test"
fi

test_name "Create Deere integration without friendly name fails"
if [[ -n "$INTEGRATION_CONFIG_ID" && -n "$BROKER_ID" ]]; then
    xbe_json do deere-integrations create --integration-identifier "missing-name" --broker "$BROKER_ID" --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing prerequisites for error test"
fi

test_name "Update Deere integration without any fields fails"
if [[ -n "$CREATED_DEERE_INTEGRATION_ID" && "$CREATED_DEERE_INTEGRATION_ID" != "null" ]]; then
    xbe_json do deere-integrations update "$CREATED_DEERE_INTEGRATION_ID"
    assert_failure
else
    skip "No Deere integration ID available for update error test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
