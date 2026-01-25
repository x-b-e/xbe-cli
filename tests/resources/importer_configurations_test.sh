#!/bin/bash
#
# XBE CLI Integration Tests: Importer Configurations
#
# Tests CRUD operations for the importer-configurations resource.
#
# COVERAGE: Create/update attributes + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_IMPORTER_CONFIGURATION_ID=""
INTEGRATION_CONFIG_ID=""
BROKER_ID=""

describe "Resource: importer-configurations"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Resolve integration config and broker prerequisites"
if [[ -n "$XBE_TEST_IMPORTER_CONFIGURATION_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_IMPORTER_CONFIGURATION_INTEGRATION_CONFIG_ID"
elif [[ -n "$XBE_TEST_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_INTEGRATION_CONFIG_ID"
fi

if [[ -n "$XBE_TEST_IMPORTER_CONFIGURATION_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_IMPORTER_CONFIGURATION_BROKER_ID"
elif [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_BROKER_ID"
fi

if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    echo "    Using integration config: $INTEGRATION_CONFIG_ID"
    if [[ -n "$BROKER_ID" ]]; then
        echo "    Using broker: $BROKER_ID"
    fi
    pass
else
    skip "Set XBE_TEST_IMPORTER_CONFIGURATION_INTEGRATION_CONFIG_ID (or XBE_TEST_INTEGRATION_CONFIG_ID) to enable create/update/delete tests"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create importer configuration with required fields"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    DATA_SOURCE_TYPE="test-source-$(date +%s)"
    TICKET_FIELD="ticket_id"
    LATEST_QUERIES='[{"name":"recent","limit":50}]'
    ADDITIONAL_CONFIGS='{"mode":"full","retry":2}'

    xbe_json do importer-configurations create \
        --importer-data-source-type "$DATA_SOURCE_TYPE" \
        --ticket-identifier-field "$TICKET_FIELD" \
        --integration-config "$INTEGRATION_CONFIG_ID" \
        --latest-tickets-queries "$LATEST_QUERIES" \
        --additional-configurations "$ADDITIONAL_CONFIGS"

    if [[ $status -eq 0 ]]; then
        CREATED_IMPORTER_CONFIGURATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
            register_cleanup "importer-configurations" "$CREATED_IMPORTER_CONFIGURATION_ID"
            pass
        else
            fail "Created importer configuration but no ID returned"
        fi
    else
        fail "Failed to create importer configuration"
    fi
else
    skip "Missing integration config prerequisite for create"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show importer configuration"
SHOW_ID="$CREATED_IMPORTER_CONFIGURATION_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    xbe_json view importer-configurations list --limit 1
    if [[ $status -eq 0 ]]; then
        SHOW_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view importer-configurations show "$SHOW_ID"
    assert_success
else
    skip "No importer configuration ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update importer configuration data source type"
if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_DATA_SOURCE_TYPE="updated-source-$(date +%s)"
    xbe_json do importer-configurations update "$CREATED_IMPORTER_CONFIGURATION_ID" --importer-data-source-type "$UPDATED_DATA_SOURCE_TYPE"
    assert_success
else
    skip "No importer configuration ID available for update"
fi

test_name "Update importer configuration ticket identifier field"
if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_TICKET_FIELD="ticket_reference"
    xbe_json do importer-configurations update "$CREATED_IMPORTER_CONFIGURATION_ID" --ticket-identifier-field "$UPDATED_TICKET_FIELD"
    assert_success
else
    skip "No importer configuration ID available for update"
fi

test_name "Update importer configuration latest tickets queries"
if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_QUERIES='[{"name":"recent","limit":25}]'
    xbe_json do importer-configurations update "$CREATED_IMPORTER_CONFIGURATION_ID" --latest-tickets-queries "$UPDATED_QUERIES"
    assert_success
else
    skip "No importer configuration ID available for update"
fi

test_name "Update importer configuration additional configurations"
if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_CONFIGS='{"mode":"delta","retry":1}'
    xbe_json do importer-configurations update "$CREATED_IMPORTER_CONFIGURATION_ID" --additional-configurations "$UPDATED_CONFIGS"
    assert_success
else
    skip "No importer configuration ID available for update"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List importer configurations"
xbe_json view importer-configurations list --limit 5
assert_success

test_name "List importer configurations returns array"
xbe_json view importer-configurations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list importer configurations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List importer configurations with --integration-config filter"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    xbe_json view importer-configurations list --integration-config "$INTEGRATION_CONFIG_ID" --limit 5
    assert_success
else
    skip "No integration config ID available for filter"
fi

test_name "List importer configurations with --broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view importer-configurations list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List importer configurations with --limit"
xbe_json view importer-configurations list --limit 3
assert_success

test_name "List importer configurations with --offset"
xbe_json view importer-configurations list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete importer configuration requires --confirm flag"
if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
    xbe_run do importer-configurations delete "$CREATED_IMPORTER_CONFIGURATION_ID"
    assert_failure
else
    skip "No importer configuration ID available for delete"
fi

test_name "Delete importer configuration with --confirm"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    DATA_SOURCE_TYPE="delete-source-$(date +%s)"
    xbe_json do importer-configurations create \
        --importer-data-source-type "$DATA_SOURCE_TYPE" \
        --ticket-identifier-field "ticket_id" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do importer-configurations delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create importer configuration for deletion test"
    fi
else
    skip "Missing integration config prerequisite for delete test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create importer configuration without data source type fails"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    xbe_json do importer-configurations create \
        --ticket-identifier-field "ticket_id" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing integration config prerequisite for error test"
fi

test_name "Create importer configuration without ticket identifier field fails"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    xbe_json do importer-configurations create \
        --importer-data-source-type "tms" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing integration config prerequisite for error test"
fi

test_name "Update importer configuration without any fields fails"
if [[ -n "$CREATED_IMPORTER_CONFIGURATION_ID" && "$CREATED_IMPORTER_CONFIGURATION_ID" != "null" ]]; then
    xbe_json do importer-configurations update "$CREATED_IMPORTER_CONFIGURATION_ID"
    assert_failure
else
    skip "No importer configuration ID available for update error test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
