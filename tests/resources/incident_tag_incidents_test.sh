#!/bin/bash
#
# XBE CLI Integration Tests: Incident Tag Incidents
#
# Tests list, show, create, and delete operations for incident_tag_incidents.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_INCIDENT_TAG_INCIDENT_ID=""
INCIDENT_ID="${XBE_TEST_INCIDENT_ID:-}"
INCIDENT_TAG_ID="${XBE_TEST_INCIDENT_TAG_ID:-}"
CREATE_INCIDENT_ID="${XBE_TEST_INCIDENT_ID:-}"
CREATE_INCIDENT_TAG_ID="${XBE_TEST_INCIDENT_TAG_ID:-}"
CREATED_INCIDENT_TAG_INCIDENT_ID=""

describe "Resource: incident-tag-incidents"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List incident tag incidents"
xbe_json view incident-tag-incidents list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_INCIDENT_TAG_INCIDENT_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$INCIDENT_ID" || "$INCIDENT_ID" == "null" ]]; then
            INCIDENT_ID=$(echo "$output" | jq -r '.[0].incident_id')
        fi
        if [[ -z "$INCIDENT_TAG_ID" || "$INCIDENT_TAG_ID" == "null" ]]; then
            INCIDENT_TAG_ID=$(echo "$output" | jq -r '.[0].incident_tag_id')
        fi
    fi
else
    fail "Failed to list incident tag incidents"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show incident tag incident"
if [[ -n "$SEED_INCIDENT_TAG_INCIDENT_ID" && "$SEED_INCIDENT_TAG_INCIDENT_ID" != "null" ]]; then
    xbe_json view incident-tag-incidents show "$SEED_INCIDENT_TAG_INCIDENT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        if [[ -z "$INCIDENT_ID" || "$INCIDENT_ID" == "null" ]]; then
            INCIDENT_ID=$(json_get ".incident_id")
        fi
        if [[ -z "$INCIDENT_TAG_ID" || "$INCIDENT_TAG_ID" == "null" ]]; then
            INCIDENT_TAG_ID=$(json_get ".incident_tag_id")
        fi
        pass
    else
        fail "Failed to show incident tag incident"
    fi
else
    skip "No incident tag incident available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create incident tag incident"
if [[ -n "$CREATE_INCIDENT_ID" && "$CREATE_INCIDENT_ID" != "null" && -n "$CREATE_INCIDENT_TAG_ID" && "$CREATE_INCIDENT_TAG_ID" != "null" ]]; then
    xbe_json do incident-tag-incidents create --incident "$CREATE_INCIDENT_ID" --incident-tag "$CREATE_INCIDENT_TAG_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_INCIDENT_TAG_INCIDENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_INCIDENT_TAG_INCIDENT_ID" && "$CREATED_INCIDENT_TAG_INCIDENT_ID" != "null" ]]; then
            register_cleanup "incident-tag-incidents" "$CREATED_INCIDENT_TAG_INCIDENT_ID"
            pass
        else
            fail "Created incident tag incident but no ID returned"
        fi
    else
        fail "Failed to create incident tag incident"
    fi
else
    skip "No incident or incident tag ID available for creation (set XBE_TEST_INCIDENT_ID and XBE_TEST_INCIDENT_TAG_ID to a compatible pair)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete incident tag incident"
if [[ -n "$CREATED_INCIDENT_TAG_INCIDENT_ID" && "$CREATED_INCIDENT_TAG_INCIDENT_ID" != "null" ]]; then
    xbe_run do incident-tag-incidents delete "$CREATED_INCIDENT_TAG_INCIDENT_ID" --confirm
    assert_success
else
    skip "No created incident tag incident to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by incident"
if [[ -n "$INCIDENT_ID" && "$INCIDENT_ID" != "null" ]]; then
    xbe_json view incident-tag-incidents list --incident "$INCIDENT_ID" --limit 5
    assert_success
else
    skip "No incident ID available for filter"
fi

test_name "Filter by incident tag"
if [[ -n "$INCIDENT_TAG_ID" && "$INCIDENT_TAG_ID" != "null" ]]; then
    xbe_json view incident-tag-incidents list --incident-tag "$INCIDENT_TAG_ID" --limit 5
    assert_success
else
    skip "No incident tag ID available for filter"
fi

run_tests
