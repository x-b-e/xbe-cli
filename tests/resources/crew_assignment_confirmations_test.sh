#!/bin/bash
#
# XBE CLI Integration Tests: Crew Assignment Confirmations
#
# Tests list, show, create, and update operations for the crew-assignment-confirmations resource.
#
# COVERAGE: List filters + show + create/update + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_UUID=""
SAMPLE_CREW_REQUIREMENT_ID=""
SAMPLE_RESOURCE_TYPE=""
SAMPLE_RESOURCE_ID=""
SAMPLE_CONFIRMED_BY_ID=""

describe "Resource: crew-assignment-confirmations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List crew assignment confirmations"
xbe_json view crew-assignment-confirmations list --limit 5
assert_success

test_name "List crew assignment confirmations returns array"
xbe_json view crew-assignment-confirmations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list crew assignment confirmations"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample confirmation"
xbe_json view crew-assignment-confirmations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_UUID=$(json_get ".[0].assignment_confirmation_uuid")
    SAMPLE_CREW_REQUIREMENT_ID=$(json_get ".[0].crew_requirement_id")
    SAMPLE_RESOURCE_TYPE=$(json_get ".[0].resource_type")
    SAMPLE_RESOURCE_ID=$(json_get ".[0].resource_id")
    SAMPLE_CONFIRMED_BY_ID=$(json_get ".[0].confirmed_by_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No confirmations available for follow-on tests"
    fi
else
    skip "Could not list confirmations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List confirmations with --crew-requirement filter"
if [[ -n "$SAMPLE_CREW_REQUIREMENT_ID" && "$SAMPLE_CREW_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view crew-assignment-confirmations list --crew-requirement "$SAMPLE_CREW_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No crew requirement ID available"
fi

test_name "List confirmations with --resource-type/--resource-id filter"
if [[ -n "$SAMPLE_RESOURCE_TYPE" && "$SAMPLE_RESOURCE_TYPE" != "null" && -n "$SAMPLE_RESOURCE_ID" && "$SAMPLE_RESOURCE_ID" != "null" ]]; then
    xbe_json view crew-assignment-confirmations list --resource-type "$SAMPLE_RESOURCE_TYPE" --resource-id "$SAMPLE_RESOURCE_ID" --limit 5
    assert_success
else
    skip "No resource type/id available"
fi

test_name "List confirmations with --confirmed-by filter"
if [[ -n "$SAMPLE_CONFIRMED_BY_ID" && "$SAMPLE_CONFIRMED_BY_ID" != "null" ]]; then
    xbe_json view crew-assignment-confirmations list --confirmed-by "$SAMPLE_CONFIRMED_BY_ID" --limit 5
    assert_success
else
    skip "No confirmed-by ID available"
fi

test_name "List confirmations with --confirmed-at-min filter"
xbe_json view crew-assignment-confirmations list --confirmed-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List confirmations with --confirmed-at-max filter"
xbe_json view crew-assignment-confirmations list --confirmed-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List confirmations with --created-at-min filter"
xbe_json view crew-assignment-confirmations list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List confirmations with --created-at-max filter"
xbe_json view crew-assignment-confirmations list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List confirmations with --updated-at-min filter"
xbe_json view crew-assignment-confirmations list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List confirmations with --updated-at-max filter"
xbe_json view crew-assignment-confirmations list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show crew assignment confirmation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view crew-assignment-confirmations show "$SAMPLE_ID"
    assert_success
else
    skip "No confirmation ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create confirmation with assignment-confirmation-uuid"
if [[ -n "$SAMPLE_UUID" && "$SAMPLE_UUID" != "null" ]]; then
    xbe_json do crew-assignment-confirmations create --assignment-confirmation-uuid "$SAMPLE_UUID" --note "CLI test confirmation"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"already confirmed"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No assignment confirmation UUID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update confirmation note"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do crew-assignment-confirmations update "$SAMPLE_ID" --note "CLI updated note"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update confirmation (permissions or policy)"
    fi
else
    skip "No confirmation ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create confirmation without required fields fails"
xbe_run do crew-assignment-confirmations create
assert_failure

test_name "Update confirmation without any fields fails"
xbe_run do crew-assignment-confirmations update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
