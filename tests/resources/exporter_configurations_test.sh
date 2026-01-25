#!/bin/bash
#
# XBE CLI Integration Tests: Exporter Configurations
#
# Tests CRUD operations for the exporter-configurations resource.
#
# COVERAGE: Create/update attributes + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EXPORTER_CONFIGURATION_ID=""
INTEGRATION_CONFIG_ID=""
BROKER_ID=""

describe "Resource: exporter-configurations"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Resolve integration config and broker prerequisites"
if [[ -n "$XBE_TEST_EXPORTER_CONFIGURATION_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_EXPORTER_CONFIGURATION_INTEGRATION_CONFIG_ID"
elif [[ -n "$XBE_TEST_INTEGRATION_CONFIG_ID" ]]; then
    INTEGRATION_CONFIG_ID="$XBE_TEST_INTEGRATION_CONFIG_ID"
fi

if [[ -n "$XBE_TEST_EXPORTER_CONFIGURATION_BROKER_ID" ]]; then
    BROKER_ID="$XBE_TEST_EXPORTER_CONFIGURATION_BROKER_ID"
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
    skip "Set XBE_TEST_EXPORTER_CONFIGURATION_INTEGRATION_CONFIG_ID (or XBE_TEST_INTEGRATION_CONFIG_ID) to enable create/update/delete tests"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create exporter configuration with required fields"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    NAME=$(unique_name "ExporterConfig")
    API_URL="https://example.com/api/export/$(date +%s)"
    TICKET_FIELD="ticket_id"
    EXPORTER_HEADERS='{"Authorization":"Bearer token","X-Test":"true"}'
    ADDITIONAL_CONFIGS='{"mode":"full","retry":3}'

    xbe_json do exporter-configurations create \
        --name "$NAME" \
        --api-url "$API_URL" \
        --ticket-identifier-field "$TICKET_FIELD" \
        --integration-config "$INTEGRATION_CONFIG_ID" \
        --template "default" \
        --exporter-headers "$EXPORTER_HEADERS" \
        --additional-configurations "$ADDITIONAL_CONFIGS"

    if [[ $status -eq 0 ]]; then
        CREATED_EXPORTER_CONFIGURATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
            register_cleanup "exporter-configurations" "$CREATED_EXPORTER_CONFIGURATION_ID"
            pass
        else
            fail "Created exporter configuration but no ID returned"
        fi
    else
        fail "Failed to create exporter configuration"
    fi
else
    skip "Missing integration config prerequisite for create"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show exporter configuration"
SHOW_ID="$CREATED_EXPORTER_CONFIGURATION_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    xbe_json view exporter-configurations list --limit 1
    if [[ $status -eq 0 ]]; then
        SHOW_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view exporter-configurations show "$SHOW_ID"
    assert_success
else
    skip "No exporter configuration ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update exporter configuration name"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_NAME=$(unique_name "ExporterConfigUpdated")
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID" --name "$UPDATED_NAME"
    assert_success
else
    skip "No exporter configuration ID available for update"
fi

test_name "Update exporter configuration API URL"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_API_URL="https://example.com/api/export/updated-$(date +%s)"
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID" --api-url "$UPDATED_API_URL"
    assert_success
else
    skip "No exporter configuration ID available for update"
fi

test_name "Update exporter configuration ticket identifier field"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_TICKET_FIELD="ticket_reference"
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID" --ticket-identifier-field "$UPDATED_TICKET_FIELD"
    assert_success
else
    skip "No exporter configuration ID available for update"
fi

test_name "Update exporter configuration template"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID" --template "updated-template"
    assert_success
else
    skip "No exporter configuration ID available for update"
fi

test_name "Update exporter configuration headers"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_HEADERS='{"Authorization":"Bearer updated","X-Trace":"true"}'
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID" --exporter-headers "$UPDATED_HEADERS"
    assert_success
else
    skip "No exporter configuration ID available for update"
fi

test_name "Update exporter configuration additional configurations"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    UPDATED_CONFIGS='{"mode":"delta","retry":1}'
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID" --additional-configurations "$UPDATED_CONFIGS"
    assert_success
else
    skip "No exporter configuration ID available for update"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List exporter configurations"
xbe_json view exporter-configurations list --limit 5
assert_success

test_name "List exporter configurations returns array"
xbe_json view exporter-configurations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list exporter configurations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List exporter configurations with --integration-config filter"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    xbe_json view exporter-configurations list --integration-config "$INTEGRATION_CONFIG_ID" --limit 5
    assert_success
else
    skip "No integration config ID available for filter"
fi

test_name "List exporter configurations with --broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view exporter-configurations list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List exporter configurations with --limit"
xbe_json view exporter-configurations list --limit 3
assert_success

test_name "List exporter configurations with --offset"
xbe_json view exporter-configurations list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete exporter configuration requires --confirm flag"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    xbe_run do exporter-configurations delete "$CREATED_EXPORTER_CONFIGURATION_ID"
    assert_failure
else
    skip "No exporter configuration ID available for delete"
fi

test_name "Delete exporter configuration with --confirm"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    NAME=$(unique_name "ExporterConfigDelete")
    API_URL="https://example.com/api/export/delete-$(date +%s)"
    xbe_json do exporter-configurations create \
        --name "$NAME" \
        --api-url "$API_URL" \
        --ticket-identifier-field "ticket_id" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do exporter-configurations delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create exporter configuration for deletion test"
    fi
else
    skip "Missing integration config prerequisite for delete test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create exporter configuration without name fails"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    xbe_json do exporter-configurations create \
        --api-url "https://example.com/api/export" \
        --ticket-identifier-field "ticket_id" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing integration config prerequisite for error test"
fi

test_name "Create exporter configuration without ticket identifier field fails"
if [[ -n "$INTEGRATION_CONFIG_ID" ]]; then
    xbe_json do exporter-configurations create \
        --name "Missing ticket field" \
        --api-url "https://example.com/api/export" \
        --integration-config "$INTEGRATION_CONFIG_ID"
    assert_failure
else
    skip "Missing integration config prerequisite for error test"
fi

test_name "Update exporter configuration without any fields fails"
if [[ -n "$CREATED_EXPORTER_CONFIGURATION_ID" && "$CREATED_EXPORTER_CONFIGURATION_ID" != "null" ]]; then
    xbe_json do exporter-configurations update "$CREATED_EXPORTER_CONFIGURATION_ID"
    assert_failure
else
    skip "No exporter configuration ID available for update error test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
