#!/bin/bash
#
# XBE CLI Integration Tests: Objective Changes
#
# Tests list/show operations for objective-changes.
#
# COVERAGE: List filters (objective, organization, organization-type/id, broker, changed-by) + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

OBJECTIVE_ID="${XBE_TEST_OBJECTIVE_CHANGE_OBJECTIVE_ID:-}"
ORGANIZATION_TYPE="${XBE_TEST_OBJECTIVE_CHANGE_ORGANIZATION_TYPE:-}"
ORGANIZATION_ID="${XBE_TEST_OBJECTIVE_CHANGE_ORGANIZATION_ID:-}"
BROKER_ID="${XBE_TEST_OBJECTIVE_CHANGE_BROKER_ID:-}"
CHANGED_BY_ID="${XBE_TEST_OBJECTIVE_CHANGE_CHANGED_BY_ID:-}"
SAMPLE_ID=""

describe "Resource: objective-changes"

normalize_org_type() {
    case "$1" in
        brokers|Broker|BROKER) echo "Broker" ;;
        customers|Customer|CUSTOMER) echo "Customer" ;;
        truckers|Trucker|TRUCKER) echo "Trucker" ;;
        *) echo "$1" ;;
    esac
}

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List objective changes"
xbe_json view objective-changes list --limit 5
assert_success

test_name "List objective changes returns array"
xbe_json view objective-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list objective changes"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample objective change"
xbe_json view objective-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -z "$OBJECTIVE_ID" || "$OBJECTIVE_ID" == "null" ]]; then
        OBJECTIVE_ID=$(json_get ".[0].objective_id")
    fi
    if [[ -z "$ORGANIZATION_TYPE" || "$ORGANIZATION_TYPE" == "null" ]]; then
        ORGANIZATION_TYPE=$(json_get ".[0].organization_type")
    fi
    if [[ -z "$ORGANIZATION_ID" || "$ORGANIZATION_ID" == "null" ]]; then
        ORGANIZATION_ID=$(json_get ".[0].organization_id")
    fi
    if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
        BROKER_ID=$(json_get ".[0].broker_id")
    fi
    if [[ -z "$CHANGED_BY_ID" || "$CHANGED_BY_ID" == "null" ]]; then
        CHANGED_BY_ID=$(json_get ".[0].changed_by_id")
    fi
    if [[ -n "$ORGANIZATION_TYPE" && "$ORGANIZATION_TYPE" != "null" ]]; then
        ORGANIZATION_TYPE=$(normalize_org_type "$ORGANIZATION_TYPE")
    fi
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No objective changes available for show test"
    fi
else
    skip "Could not list objective changes to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List objective changes with --objective filter"
if [[ -n "$OBJECTIVE_ID" && "$OBJECTIVE_ID" != "null" ]]; then
    xbe_json view objective-changes list --objective "$OBJECTIVE_ID" --limit 5
    assert_success
else
    skip "No objective ID available for --objective filter"
fi

test_name "List objective changes with --organization filter"
if [[ -n "$ORGANIZATION_TYPE" && "$ORGANIZATION_TYPE" != "null" && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" ]]; then
    xbe_json view objective-changes list --organization "${ORGANIZATION_TYPE}|${ORGANIZATION_ID}" --limit 5
    assert_success
else
    skip "No organization available for --organization filter"
fi

test_name "List objective changes with --organization-type and --organization-id filters"
if [[ -n "$ORGANIZATION_TYPE" && "$ORGANIZATION_TYPE" != "null" && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" ]]; then
    xbe_json view objective-changes list --organization-type "$ORGANIZATION_TYPE" --organization-id "$ORGANIZATION_ID" --limit 5
    assert_success
else
    skip "No organization available for organization-type/id filters"
fi

test_name "List objective changes with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view objective-changes list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for --broker filter"
fi

test_name "List objective changes with --changed-by filter"
if [[ -n "$CHANGED_BY_ID" && "$CHANGED_BY_ID" != "null" ]]; then
    xbe_json view objective-changes list --changed-by "$CHANGED_BY_ID" --limit 5
    assert_success
else
    skip "No changed-by ID available for --changed-by filter"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show objective change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view objective-changes show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show objective change: $output"
        fi
    fi
else
    skip "No objective change ID available for show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
