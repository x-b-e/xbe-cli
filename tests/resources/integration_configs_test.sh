#!/bin/bash
#
# XBE CLI Integration Tests: Integration Configs
#
# Tests view operations for integration configs.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

FIRST_CONFIG_ID=""
FRIENDLY_NAME=""
BROKER_ID=""
ORG_TYPE=""
ORG_ID=""

describe "Resource: integration-configs (view-only)"

normalize_org_type() {
    case "$1" in
        brokers|Broker|BROKER) echo "Broker" ;;
        customers|Customer|CUSTOMER) echo "Customer" ;;
        truckers|Trucker|TRUCKER) echo "Trucker" ;;
        developers|Developer|DEVELOPER) echo "Developer" ;;
        material-suppliers|MaterialSupplier|MATERIALSUPPLIER) echo "MaterialSupplier" ;;
        business-units|BusinessUnit|BUSINESSUNIT) echo "BusinessUnit" ;;
        *) echo "$1" ;;
    esac
}

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List integration configs"
xbe_json view integration-configs list --limit 5
assert_success

test_name "List integration configs returns array"
xbe_json view integration-configs list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list integration configs"
fi

# Capture IDs for downstream tests
xbe_json view integration-configs list --limit 5
if [[ $status -eq 0 ]]; then
    FIRST_CONFIG_ID=$(json_get ".[0].id")
    FRIENDLY_NAME=$(json_get ".[0].friendly_name")
    BROKER_ID=$(json_get ".[0].broker_id")
    ORG_TYPE=$(json_get ".[0].organization_type")
    ORG_ID=$(json_get ".[0].organization_id")
    if [[ -n "$ORG_TYPE" && "$ORG_TYPE" != "null" ]]; then
        ORG_TYPE=$(normalize_org_type "$ORG_TYPE")
    fi
else
    FIRST_CONFIG_ID=""
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show integration config"
if [[ -n "$FIRST_CONFIG_ID" && "$FIRST_CONFIG_ID" != "null" ]]; then
    xbe_json view integration-configs show "$FIRST_CONFIG_ID"
    if [[ $status -eq 0 ]]; then
        FRIENDLY_NAME=$(json_get ".friendly_name")
        BROKER_ID=$(json_get ".broker_id")
        ORG_TYPE=$(json_get ".organization_type")
        ORG_ID=$(json_get ".organization_id")
        if [[ -n "$ORG_TYPE" && "$ORG_TYPE" != "null" ]]; then
            ORG_TYPE=$(normalize_org_type "$ORG_TYPE")
        fi
        pass
    else
        fail "Failed to show integration config"
    fi
else
    skip "No integration config ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List integration configs with --friendly-name filter"
if [[ -n "$FRIENDLY_NAME" && "$FRIENDLY_NAME" != "null" ]]; then
    xbe_json view integration-configs list --friendly-name "$FRIENDLY_NAME" --limit 5
    assert_success
else
    skip "No friendly name available for filter test"
fi

test_name "List integration configs with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view integration-configs list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter test"
fi

test_name "List integration configs with --organization filter"
if [[ -n "$ORG_TYPE" && "$ORG_TYPE" != "null" && -n "$ORG_ID" && "$ORG_ID" != "null" ]]; then
    xbe_json view integration-configs list --organization "${ORG_TYPE}|${ORG_ID}" --limit 5
    assert_success
else
    skip "No organization type/id available for filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
