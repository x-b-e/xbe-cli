#!/bin/bash
#
# XBE CLI Integration Tests: Prompt Prescriptions
#
# Tests list, show, and create operations for
# prompt-prescriptions.
#
# COVERAGE: List filters + show + create (email/name/org/location/role/symptoms)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_EMAIL=""
CREATED_ID=""

PROMPT_EMAIL="${XBE_TEST_PROMPT_PRESCRIPTION_EMAIL:-}"
PROMPT_NAME="${XBE_TEST_PROMPT_PRESCRIPTION_NAME:-}"
PROMPT_ORG="${XBE_TEST_PROMPT_PRESCRIPTION_ORGANIZATION:-}"
PROMPT_LOCATION="${XBE_TEST_PROMPT_PRESCRIPTION_LOCATION:-}"
PROMPT_ROLE="${XBE_TEST_PROMPT_PRESCRIPTION_ROLE:-}"
PROMPT_SYMPTOMS="${XBE_TEST_PROMPT_PRESCRIPTION_SYMPTOMS:-}"

describe "Resource: prompt-prescriptions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prompt prescriptions"
xbe_json view prompt-prescriptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_EMAIL=$(json_get ".[0].email_address")
    fi
else
    fail "Failed to list prompt prescriptions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prompt prescriptions with --email-address filter"
FILTER_EMAIL="$SAMPLE_EMAIL"
if [[ -z "$FILTER_EMAIL" || "$FILTER_EMAIL" == "null" ]]; then
    FILTER_EMAIL="$PROMPT_EMAIL"
fi

if [[ -n "$FILTER_EMAIL" && "$FILTER_EMAIL" != "null" ]]; then
    xbe_json view prompt-prescriptions list --email-address "$FILTER_EMAIL" --limit 5
    assert_success
else
    skip "No email address available"
fi

test_name "List prompt prescriptions with created-at filters"
xbe_json view prompt-prescriptions list --created-at-min "2000-01-01T00:00:00Z" --created-at-max "2100-01-01T00:00:00Z" --limit 5
assert_success

test_name "List prompt prescriptions with updated-at filters"
xbe_json view prompt-prescriptions list --updated-at-min "2000-01-01T00:00:00Z" --updated-at-max "2100-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prompt prescription"
SHOW_ID="$SAMPLE_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$CREATED_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view prompt-prescriptions show "$SHOW_ID"
    assert_success
else
    skip "No prompt prescription ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prompt prescription requires required fields"
xbe_run do prompt-prescriptions create
assert_failure

test_name "Create prompt prescription"
if [[ -n "$PROMPT_EMAIL" && "$PROMPT_EMAIL" != "null" ]]; then
    if [[ -z "$PROMPT_NAME" || "$PROMPT_NAME" == "null" ]]; then
        PROMPT_NAME=$(unique_name "PromptPrescriber")
    fi
    if [[ -z "$PROMPT_ORG" || "$PROMPT_ORG" == "null" ]]; then
        PROMPT_ORG=$(unique_name "PromptOrg")
    fi
    if [[ -z "$PROMPT_LOCATION" || "$PROMPT_LOCATION" == "null" ]]; then
        PROMPT_LOCATION="Austin, TX"
    fi
    if [[ -z "$PROMPT_ROLE" || "$PROMPT_ROLE" == "null" ]]; then
        PROMPT_ROLE="Operations Manager"
    fi

    if [[ -n "$PROMPT_SYMPTOMS" && "$PROMPT_SYMPTOMS" != "null" ]]; then
        xbe_json do prompt-prescriptions create \
            --email-address "$PROMPT_EMAIL" \
            --name "$PROMPT_NAME" \
            --organization-name "$PROMPT_ORG" \
            --location-name "$PROMPT_LOCATION" \
            --role "$PROMPT_ROLE" \
            --symptoms "$PROMPT_SYMPTOMS"
    else
        xbe_json do prompt-prescriptions create \
            --email-address "$PROMPT_EMAIL" \
            --name "$PROMPT_NAME" \
            --organization-name "$PROMPT_ORG" \
            --location-name "$PROMPT_LOCATION" \
            --role "$PROMPT_ROLE"
    fi

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".email_address" "$PROMPT_EMAIL"
        CREATED_ID=$(json_get ".id")
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
            skip "Not authorized to create prompt prescriptions"
        else
            fail "Failed to create prompt prescription"
        fi
    fi
else
    skip "Set XBE_TEST_PROMPT_PRESCRIPTION_EMAIL to enable create testing."
fi

run_tests
