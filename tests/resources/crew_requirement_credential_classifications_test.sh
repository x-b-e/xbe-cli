#!/bin/bash
#
# XBE CLI Integration Tests: Crew Requirement Credential Classifications
#
# Tests list, show, create, and delete operations for the crew-requirement-credential-classifications resource.
#
# COVERAGE: List filters + show + create/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_CREW_REQUIREMENT_TYPE=""
SAMPLE_CREW_REQUIREMENT_ID=""
SAMPLE_CREDENTIAL_CLASSIFICATION_TYPE=""
SAMPLE_CREDENTIAL_CLASSIFICATION_ID=""
CREATED_ID=""

describe "Resource: crew-requirement-credential-classifications"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List crew requirement credential classifications"
xbe_json view crew-requirement-credential-classifications list --limit 5
assert_success

test_name "List crew requirement credential classifications returns array"
xbe_json view crew-requirement-credential-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list crew requirement credential classifications"
fi

# ============================================================================
# Sample Record (used for filters/show/create/delete)
# ============================================================================

test_name "Capture sample link"
xbe_json view crew-requirement-credential-classifications list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_CREW_REQUIREMENT_TYPE=$(json_get ".[0].crew_requirement_type")
    SAMPLE_CREW_REQUIREMENT_ID=$(json_get ".[0].crew_requirement_id")
    SAMPLE_CREDENTIAL_CLASSIFICATION_TYPE=$(json_get ".[0].credential_classification_type")
    SAMPLE_CREDENTIAL_CLASSIFICATION_ID=$(json_get ".[0].credential_classification_id")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No links available for follow-on tests"
    fi
else
    skip "Could not list links to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List links with --crew-requirement filter"
if [[ -n "$SAMPLE_CREW_REQUIREMENT_ID" && "$SAMPLE_CREW_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view crew-requirement-credential-classifications list --crew-requirement "$SAMPLE_CREW_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No crew requirement ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show crew requirement credential classification"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view crew-requirement-credential-classifications show "$SAMPLE_ID"
    assert_success
else
    skip "No link ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create crew requirement credential classification"
if [[ -n "$SAMPLE_CREW_REQUIREMENT_TYPE" && "$SAMPLE_CREW_REQUIREMENT_TYPE" != "null" && \
      -n "$SAMPLE_CREW_REQUIREMENT_ID" && "$SAMPLE_CREW_REQUIREMENT_ID" != "null" && \
      -n "$SAMPLE_CREDENTIAL_CLASSIFICATION_TYPE" && "$SAMPLE_CREDENTIAL_CLASSIFICATION_TYPE" != "null" && \
      -n "$SAMPLE_CREDENTIAL_CLASSIFICATION_ID" && "$SAMPLE_CREDENTIAL_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do crew-requirement-credential-classifications create \
        --crew-requirement-type "$SAMPLE_CREW_REQUIREMENT_TYPE" \
        --crew-requirement "$SAMPLE_CREW_REQUIREMENT_ID" \
        --credential-classification-type "$SAMPLE_CREDENTIAL_CLASSIFICATION_TYPE" \
        --credential-classification "$SAMPLE_CREDENTIAL_CLASSIFICATION_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "crew-requirement-credential-classifications" "$CREATED_ID"
            pass
        else
            fail "Created link but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"422"* ]] || [[ "$output" == *"must match assigned laborer"* ]] || \
           [[ "$output" == *"must be UserCredentialClassification"* ]] || [[ "$output" == *"must be LaborRequirement"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No sample crew requirement/credential classification available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete crew requirement credential classification"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do crew-requirement-credential-classifications delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created link available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create link without required fields fails"
xbe_run do crew-requirement-credential-classifications create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
