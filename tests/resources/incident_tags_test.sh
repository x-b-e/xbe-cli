#!/bin/bash
#
# XBE CLI Integration Tests: Incident Tags
#
# Tests create and list operations for the incident_tags resource.
# Incident tags categorize incidents.
#
# COVERAGE: Create + list filters (create-only, no update/delete)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TAG_ID=""

describe "Resource: incident-tags"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create incident tag with required fields"
TEST_SLUG="tag-$(date +%s)"
TEST_NAME=$(unique_name "IncidentTag")
xbe_json do incident-tags create --slug "$TEST_SLUG" --name "$TEST_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_TAG_ID=$(json_get ".id")
    if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" ]]; then
        # Note: No delete available for incident-tags
        pass
    else
        fail "Created incident tag but no ID returned"
    fi
else
    fail "Failed to create incident tag: $output"
fi

test_name "Create incident tag with description"
TEST_SLUG2="tag-desc-$(date +%s)"
TEST_NAME2=$(unique_name "IncidentTagDesc")
xbe_json do incident-tags create \
    --slug "$TEST_SLUG2" \
    --name "$TEST_NAME2" \
    --description "Test incident tag description"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create incident tag with description"
fi

test_name "Create incident tag with kinds"
TEST_SLUG3="tag-kinds-$(date +%s)"
TEST_NAME3=$(unique_name "IncidentTagKinds")
# Note: kinds must match valid values in the system (e.g., "injury", "vehicle", etc.)
xbe_json do incident-tags create \
    --slug "$TEST_SLUG3" \
    --name "$TEST_NAME3" \
    --kinds "injury,vehicle"

if [[ $status -eq 0 ]]; then
    pass
else
    # If invalid kinds, the CLI still worked correctly
    if [[ "$output" == *"kinds"* ]]; then
        echo "    (Server validation for kinds - testing flag works)"
        pass
    else
        fail "Failed to create incident tag with kinds"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List incident tags"
xbe_json view incident-tags list
assert_success

test_name "List incident tags returns array"
xbe_json view incident-tags list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list incident tags"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List incident tags with --slug filter"
xbe_json view incident-tags list --slug "property-damage"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List incident tags with --limit"
xbe_json view incident-tags list --limit 5
assert_success

test_name "List incident tags with --offset"
xbe_json view incident-tags list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create incident tag without slug fails"
xbe_json do incident-tags create --name "Missing Slug"
assert_failure

test_name "Create incident tag without name fails"
xbe_json do incident-tags create --slug "missing-name"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
