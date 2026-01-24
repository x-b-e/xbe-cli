#!/bin/bash
#
# XBE CLI Integration Tests: Incident Headline Suggestions
#
# Tests list, show, and create operations for
# incident-headline-suggestions.
#
# COVERAGE: List filters + show + create (options, is-async)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
INCIDENT_ID="${XBE_TEST_INCIDENT_ID:-}"
CREATED_ID=""

OPTIONS_JSON='{"temperature":0.4,"max_tokens":256}'

describe "Resource: incident-headline-suggestions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List incident headline suggestions"
xbe_json view incident-headline-suggestions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -z "$INCIDENT_ID" || "$INCIDENT_ID" == "null" ]]; then
            INCIDENT_ID=$(json_get ".[0].incident_id")
        fi
    fi
else
    fail "Failed to list incident headline suggestions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List incident headline suggestions with --incident filter"
if [[ -z "$INCIDENT_ID" || "$INCIDENT_ID" == "null" ]]; then
    xbe_json view incidents list --limit 1
    if [[ $status -eq 0 ]]; then
        INCIDENT_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$INCIDENT_ID" && "$INCIDENT_ID" != "null" ]]; then
    xbe_json view incident-headline-suggestions list --incident "$INCIDENT_ID" --limit 5
    assert_success
else
    skip "No incident ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show incident headline suggestion"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view incident-headline-suggestions show "$SAMPLE_ID"
    assert_success
else
    skip "No incident headline suggestion ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create incident headline suggestion requires --incident"
xbe_run do incident-headline-suggestions create
assert_failure

test_name "Create incident headline suggestion"
if [[ -n "$INCIDENT_ID" && "$INCIDENT_ID" != "null" ]]; then
    xbe_json do incident-headline-suggestions create --incident "$INCIDENT_ID" --options "$OPTIONS_JSON" --is-async
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_bool ".is_async" "true"
        CREATED_ID=$(json_get ".id")
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to create incident headline suggestions"
        else
            fail "Failed to create incident headline suggestion"
        fi
    fi
else
    skip "No incident ID available. Set XBE_TEST_INCIDENT_ID to enable create testing."
fi

# ============================================================================
# CREATE Tests - Validate Options
# ============================================================================

test_name "Show created incident headline suggestion includes options"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json view incident-headline-suggestions show "$CREATED_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".options"
        assert_json_has ".options.temperature"
        assert_json_has ".options.max_tokens"
    else
        fail "Failed to show created incident headline suggestion"
    fi
else
    skip "No created incident headline suggestion ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete incident headline suggestion"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do incident-headline-suggestions delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created incident headline suggestion ID available"
fi

run_tests
