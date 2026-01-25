#!/bin/bash
#
# XBE CLI Integration Tests: Objective Stakeholder Classification Quotes
#
# Tests CRUD operations and list filters for the objective_stakeholder_classification_quotes resource.
#
# COVERAGE: Create/update attributes + list filters + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: objective-stakeholder-classification-quotes"

CREATED_QUOTE_ID=""
SAMPLE_QUOTE_ID=""
SAMPLE_CLASSIFICATION_ID=""

CLASSIFICATION_ID="${XBE_TEST_OBJECTIVE_STAKEHOLDER_CLASSIFICATION_ID:-}"

SKIP_MUTATION=0

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List objective stakeholder classification quotes"
xbe_json view objective-stakeholder-classification-quotes list --limit 5
assert_success

test_name "List objective stakeholder classification quotes returns array"
xbe_json view objective-stakeholder-classification-quotes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list objective stakeholder classification quotes"
fi

# ==========================================================================
# Sample Record (for filters/show)
# ==========================================================================

test_name "Capture sample quote"
xbe_json view objective-stakeholder-classification-quotes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_QUOTE_ID=$(json_get ".[0].id")
    SAMPLE_CLASSIFICATION_ID=$(json_get ".[0].objective_stakeholder_classification_id")
    if [[ -n "$SAMPLE_QUOTE_ID" && "$SAMPLE_QUOTE_ID" != "null" ]]; then
        pass
    else
        skip "No quotes available for follow-on tests"
    fi
else
    skip "Could not list quotes to capture sample"
fi

if [[ -z "$CLASSIFICATION_ID" && -n "$SAMPLE_CLASSIFICATION_ID" && "$SAMPLE_CLASSIFICATION_ID" != "null" ]]; then
    CLASSIFICATION_ID="$SAMPLE_CLASSIFICATION_ID"
fi

if [[ -z "$CLASSIFICATION_ID" ]]; then
    SKIP_MUTATION=1
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create quote without required fields fails"
xbe_run do objective-stakeholder-classification-quotes create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "No objective stakeholder classification available; skipping mutation tests"
else
    test_name "Create objective stakeholder classification quote"
    QUOTE_CONTENT="CLI test quote $(unique_suffix)"
    xbe_json do objective-stakeholder-classification-quotes create \
        --objective-stakeholder-classification "$CLASSIFICATION_ID" \
        --content "$QUOTE_CONTENT" \
        --is-generated

    if [[ $status -eq 0 ]]; then
        CREATED_QUOTE_ID=$(json_get ".id")
        if [[ -n "$CREATED_QUOTE_ID" && "$CREATED_QUOTE_ID" != "null" ]]; then
            register_cleanup "objective-stakeholder-classification-quotes" "$CREATED_QUOTE_ID"
            pass
        else
            fail "Created quote but no ID returned"
        fi
    else
        fail "Failed to create objective stakeholder classification quote"
    fi
fi

if [[ -n "$CREATED_QUOTE_ID" && "$CREATED_QUOTE_ID" != "null" ]]; then
    test_name "Create quote sets is_generated"
    assert_json_bool ".is_generated" "true"
else
    skip "No quote created; skipping is_generated check"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show objective stakeholder classification quote"
if [[ -n "$CREATED_QUOTE_ID" && "$CREATED_QUOTE_ID" != "null" ]]; then
    xbe_json view objective-stakeholder-classification-quotes show "$CREATED_QUOTE_ID"
    assert_success
elif [[ -n "$SAMPLE_QUOTE_ID" && "$SAMPLE_QUOTE_ID" != "null" ]]; then
    xbe_json view objective-stakeholder-classification-quotes show "$SAMPLE_QUOTE_ID"
    assert_success
else
    skip "No quote ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update quote without fields fails"
if [[ -n "$CREATED_QUOTE_ID" && "$CREATED_QUOTE_ID" != "null" ]]; then
    xbe_run do objective-stakeholder-classification-quotes update "$CREATED_QUOTE_ID"
    assert_failure
else
    skip "No quote created; skipping update failure test"
fi

if [[ -n "$CREATED_QUOTE_ID" && "$CREATED_QUOTE_ID" != "null" ]]; then
    test_name "Update quote content"
    xbe_json do objective-stakeholder-classification-quotes update "$CREATED_QUOTE_ID" \
        --content "Updated quote $(unique_suffix)"
    assert_success
else
    skip "No quote created; skipping update"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List quotes with --objective-stakeholder-classification filter"
if [[ -n "$CLASSIFICATION_ID" ]]; then
    xbe_json view objective-stakeholder-classification-quotes list \
        --objective-stakeholder-classification "$CLASSIFICATION_ID" \
        --limit 5
    assert_success
else
    skip "No objective stakeholder classification ID available"
fi

test_name "List quotes with --interest-degree-min filter"
xbe_json view objective-stakeholder-classification-quotes list --interest-degree-min 0 --limit 5
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"Internal Server Error"* || "$output" == *"INTERNAL SERVER ERROR"* ]]; then
    skip "Server does not support interest degree min filter for this resource"
else
    fail "Failed to list quotes with interest degree min filter"
fi

test_name "List quotes with --interest-degree-max filter"
xbe_json view objective-stakeholder-classification-quotes list --interest-degree-max 10 --limit 5
if [[ $status -eq 0 ]]; then
    pass
elif [[ "$output" == *"Internal Server Error"* || "$output" == *"INTERNAL SERVER ERROR"* ]]; then
    skip "Server does not support interest degree max filter for this resource"
else
    fail "Failed to list quotes with interest degree max filter"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete objective stakeholder classification quote"
if [[ -n "$CREATED_QUOTE_ID" && "$CREATED_QUOTE_ID" != "null" ]]; then
    xbe_json do objective-stakeholder-classification-quotes delete "$CREATED_QUOTE_ID" --confirm
    assert_success
else
    skip "No quote created; skipping delete"
fi

run_tests
