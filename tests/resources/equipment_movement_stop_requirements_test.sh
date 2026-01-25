#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Stop Requirements
#
# Tests list, show, create, and delete operations for the equipment-movement-stop-requirements resource.
#
# COVERAGE: List filters + show + create/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STOP_ID=""
SAMPLE_REQUIREMENT_ID=""
SAMPLE_KIND=""
CREATED_ID=""

describe "Resource: equipment-movement-stop-requirements"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment movement stop requirements"
xbe_json view equipment-movement-stop-requirements list --limit 5
assert_success

test_name "List equipment movement stop requirements returns array"
xbe_json view equipment-movement-stop-requirements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment movement stop requirements"
fi

# ============================================================================
# Sample Record (used for filters/show/create/delete)
# ============================================================================

test_name "Capture sample stop requirement"
xbe_json view equipment-movement-stop-requirements list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STOP_ID=$(json_get ".[0].stop_id")
    SAMPLE_REQUIREMENT_ID=$(json_get ".[0].requirement_id")
    SAMPLE_KIND=$(json_get ".[0].kind")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No stop requirements available for follow-on tests"
    fi
else
    skip "Could not list stop requirements to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List stop requirements with --stop filter"
if [[ -n "$SAMPLE_STOP_ID" && "$SAMPLE_STOP_ID" != "null" ]]; then
    xbe_json view equipment-movement-stop-requirements list --stop "$SAMPLE_STOP_ID" --limit 5
    assert_success
else
    skip "No stop ID available"
fi

test_name "List stop requirements with --requirement filter"
if [[ -n "$SAMPLE_REQUIREMENT_ID" && "$SAMPLE_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view equipment-movement-stop-requirements list --requirement "$SAMPLE_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No requirement ID available"
fi

test_name "List stop requirements with --kind filter"
if [[ -n "$SAMPLE_KIND" && "$SAMPLE_KIND" != "null" ]]; then
    xbe_json view equipment-movement-stop-requirements list --kind "$SAMPLE_KIND" --limit 5
    assert_success
else
    skip "No kind available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment movement stop requirement"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view equipment-movement-stop-requirements show "$SAMPLE_ID"
    assert_success
else
    skip "No stop requirement ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment movement stop requirement"
if [[ -n "$SAMPLE_STOP_ID" && "$SAMPLE_STOP_ID" != "null" && \
      -n "$SAMPLE_REQUIREMENT_ID" && "$SAMPLE_REQUIREMENT_ID" != "null" ]]; then
    if [[ -n "$SAMPLE_KIND" && "$SAMPLE_KIND" != "null" ]]; then
        xbe_json do equipment-movement-stop-requirements create \
            --stop "$SAMPLE_STOP_ID" \
            --requirement "$SAMPLE_REQUIREMENT_ID" \
            --kind "$SAMPLE_KIND"
    else
        xbe_json do equipment-movement-stop-requirements create \
            --stop "$SAMPLE_STOP_ID" \
            --requirement "$SAMPLE_REQUIREMENT_ID"
    fi

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "equipment-movement-stop-requirements" "$CREATED_ID"
            pass
        else
            fail "Created stop requirement but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"422"* ]] || [[ "$output" == *"has already been taken"* ]] || \
           [[ "$output" == *"must match the stop's location"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No stop/requirement ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment movement stop requirement"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do equipment-movement-stop-requirements delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created stop requirement available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create stop requirement without required fields fails"
xbe_run do equipment-movement-stop-requirements create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
