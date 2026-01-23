#!/bin/bash
#
# XBE CLI Integration Tests: Resource Unavailabilities
#
# Tests list, show, create, update, and delete operations for the resource-unavailabilities resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_RESOURCE_TYPE=""
SAMPLE_RESOURCE_ID=""
CREATED_ID=""

TEST_START_AT="2025-01-02T08:00:00Z"
TEST_END_AT="2025-01-02T17:00:00Z"
UPDATED_END_AT="2025-01-02T18:00:00Z"


describe "Resource: resource-unavailabilities"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List resource unavailabilities"
xbe_json view resource-unavailabilities list --limit 5
assert_success

test_name "List resource unavailabilities returns array"
xbe_json view resource-unavailabilities list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list resource unavailabilities"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample resource unavailability"
xbe_json view resource-unavailabilities list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_RESOURCE_TYPE=$(json_get ".[0].resource_type")
    SAMPLE_RESOURCE_ID=$(json_get ".[0].resource_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No resource unavailabilities available for follow-on tests"
    fi
else
    skip "Could not list resource unavailabilities to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List resource unavailabilities with --resource-type filter"
if [[ -n "$SAMPLE_RESOURCE_TYPE" && "$SAMPLE_RESOURCE_TYPE" != "null" ]]; then
    xbe_json view resource-unavailabilities list --resource-type "$SAMPLE_RESOURCE_TYPE" --limit 5
    assert_success
else
    skip "No resource type available"
fi


test_name "List resource unavailabilities with --resource-type/--resource-id filter"
if [[ -n "$SAMPLE_RESOURCE_TYPE" && "$SAMPLE_RESOURCE_TYPE" != "null" && -n "$SAMPLE_RESOURCE_ID" && "$SAMPLE_RESOURCE_ID" != "null" ]]; then
    xbe_json view resource-unavailabilities list --resource-type "$SAMPLE_RESOURCE_TYPE" --resource-id "$SAMPLE_RESOURCE_ID" --limit 5
    assert_success
else
    skip "No resource type/id available"
fi


test_name "List resource unavailabilities with --organization filter"
xbe_json view resource-unavailabilities list --organization "Broker|1" --limit 5
assert_success

test_name "List resource unavailabilities with --start-at-min filter"
xbe_json view resource-unavailabilities list --start-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List resource unavailabilities with --start-at-max filter"
xbe_json view resource-unavailabilities list --start-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success


test_name "List resource unavailabilities with --end-at-min filter"
xbe_json view resource-unavailabilities list --end-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List resource unavailabilities with --end-at-max filter"
xbe_json view resource-unavailabilities list --end-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show resource unavailability"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view resource-unavailabilities show "$SAMPLE_ID"
    assert_success
else
    skip "No resource unavailability ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create resource unavailability"
if [[ -n "$SAMPLE_RESOURCE_TYPE" && "$SAMPLE_RESOURCE_TYPE" != "null" && -n "$SAMPLE_RESOURCE_ID" && "$SAMPLE_RESOURCE_ID" != "null" ]]; then
    xbe_json do resource-unavailabilities create \
        --resource-type "$SAMPLE_RESOURCE_TYPE" \
        --resource-id "$SAMPLE_RESOURCE_ID" \
        --start-at "$TEST_START_AT" \
        --end-at "$TEST_END_AT" \
        --description "CLI test unavailability"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "resource-unavailabilities" "$CREATED_ID"
            pass
        else
            fail "Created resource unavailability but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"overlapping"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No resource type/id available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update resource unavailability"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do resource-unavailabilities update "$CREATED_ID" --end-at "$UPDATED_END_AT" --description "Updated via CLI"
    assert_success
elif [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do resource-unavailabilities update "$SAMPLE_ID" --description "Updated via CLI"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update resource unavailability (permissions or policy)"
    fi
else
    skip "No resource unavailability ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete resource unavailability"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do resource-unavailabilities delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created resource unavailability to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create resource unavailability without required fields fails"
xbe_run do resource-unavailabilities create
assert_failure


test_name "Update resource unavailability without any fields fails"
xbe_run do resource-unavailabilities update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
