#!/bin/bash
#
# XBE CLI Integration Tests: Objective Stakeholder Classifications
#
# Tests CRUD operations for the objective_stakeholder_classifications resource.
# Objective stakeholder classifications link objective templates to stakeholder
# classifications with an interest degree between 0 and 1.
#
# COVERAGE: All create/update attributes + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

OBJECTIVE_ID=""
STAKEHOLDER_CLASSIFICATION_ID=""
CREATED_CLASSIFICATION_ID=""
SKIP_OBJECTIVE_TESTS=0
INTEREST_DEGREE="0.5"
UPDATED_INTEREST_DEGREE="0.8"

describe "Resource: objective_stakeholder_classifications"

# ============================================================================
# Objective Template Lookup
# ============================================================================

test_name "Find objective template for objective stakeholder classification tests"
if [[ -n "$XBE_TEST_OBJECTIVE_ID" ]]; then
    OBJECTIVE_ID="$XBE_TEST_OBJECTIVE_ID"
    pass
elif [[ -n "$XBE_TOKEN" ]]; then
    base_url="${XBE_BASE_URL%/}"
    objectives_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/objectives?page[limit]=1&filter[is-template]=true" || true)
    OBJECTIVE_ID=$(echo "$objectives_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -n "$OBJECTIVE_ID" && "$OBJECTIVE_ID" != "null" ]]; then
        pass
    else
        SKIP_OBJECTIVE_TESTS=1
        skip "No objective template found"
    fi
else
    SKIP_OBJECTIVE_TESTS=1
    skip "XBE_TOKEN not set and XBE_TEST_OBJECTIVE_ID not provided"
fi

# ============================================================================
# Prerequisites - Create stakeholder classification
# ============================================================================

test_name "Create stakeholder classification for objective stakeholder classification tests"
TEST_TITLE=$(unique_name "ObjStakeholder")

xbe_json do stakeholder-classifications create --title "$TEST_TITLE"

if [[ $status -eq 0 ]]; then
    STAKEHOLDER_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$STAKEHOLDER_CLASSIFICATION_ID" && "$STAKEHOLDER_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "stakeholder-classifications" "$STAKEHOLDER_CLASSIFICATION_ID"
        pass
    else
        fail "Created stakeholder classification but no ID returned"
    fi
else
    fail "Failed to create stakeholder classification"
fi

if [[ -z "$STAKEHOLDER_CLASSIFICATION_ID" || "$STAKEHOLDER_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid stakeholder classification ID"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create objective stakeholder classification with required fields"
if [[ $SKIP_OBJECTIVE_TESTS -eq 1 ]]; then
    skip "No objective template available"
else
    xbe_json do objective-stakeholder-classifications create \
        --objective "$OBJECTIVE_ID" \
        --stakeholder-classification "$STAKEHOLDER_CLASSIFICATION_ID" \
        --interest-degree "$INTEREST_DEGREE"

    if [[ $status -eq 0 ]]; then
        CREATED_CLASSIFICATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
            register_cleanup "objective-stakeholder-classifications" "$CREATED_CLASSIFICATION_ID"
            pass
        else
            fail "Created objective stakeholder classification but no ID returned"
        fi
    else
        fail "Failed to create objective stakeholder classification"
    fi
fi

# Only continue if we successfully created a classification
if [[ $SKIP_OBJECTIVE_TESTS -eq 0 && (-z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null") ]]; then
    echo "Cannot continue without a valid objective stakeholder classification ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update objective stakeholder classification interest degree"
if [[ $SKIP_OBJECTIVE_TESTS -eq 0 ]]; then
    xbe_json do objective-stakeholder-classifications update "$CREATED_CLASSIFICATION_ID" --interest-degree "$UPDATED_INTEREST_DEGREE"
    assert_success
else
    skip "No objective stakeholder classification available"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List objective stakeholder classifications"
xbe_json view objective-stakeholder-classifications list --limit 5
assert_success

test_name "List objective stakeholder classifications returns array"
xbe_json view objective-stakeholder-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list objective stakeholder classifications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List objective stakeholder classifications with --objective filter"
if [[ -n "$OBJECTIVE_ID" && "$OBJECTIVE_ID" != "null" ]]; then
    xbe_json view objective-stakeholder-classifications list --objective "$OBJECTIVE_ID" --limit 5
    assert_success
else
    skip "No objective ID available"
fi

test_name "List objective stakeholder classifications with --stakeholder-classification filter"
if [[ -n "$STAKEHOLDER_CLASSIFICATION_ID" && "$STAKEHOLDER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json view objective-stakeholder-classifications list --stakeholder-classification "$STAKEHOLDER_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No stakeholder classification ID available"
fi

test_name "List objective stakeholder classifications with --interest-degree filter"
xbe_json view objective-stakeholder-classifications list --interest-degree "$INTEREST_DEGREE" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show objective stakeholder classification"
if [[ $SKIP_OBJECTIVE_TESTS -eq 0 && -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json view objective-stakeholder-classifications show "$CREATED_CLASSIFICATION_ID"
    assert_success
else
    skip "No objective stakeholder classification ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete objective stakeholder classification requires --confirm flag"
if [[ $SKIP_OBJECTIVE_TESTS -eq 0 && -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
    xbe_run do objective-stakeholder-classifications delete "$CREATED_CLASSIFICATION_ID"
    assert_failure
else
    skip "No objective stakeholder classification ID available"
fi

test_name "Delete objective stakeholder classification with --confirm"
if [[ $SKIP_OBJECTIVE_TESTS -eq 0 && -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
    xbe_run do objective-stakeholder-classifications delete "$CREATED_CLASSIFICATION_ID" --confirm
    assert_success
else
    skip "No objective stakeholder classification ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create objective stakeholder classification without objective fails"
xbe_json do objective-stakeholder-classifications create \
    --stakeholder-classification "$STAKEHOLDER_CLASSIFICATION_ID" \
    --interest-degree "$INTEREST_DEGREE"
assert_failure

test_name "Create objective stakeholder classification without stakeholder classification fails"
xbe_json do objective-stakeholder-classifications create \
    --objective "${OBJECTIVE_ID:-123}" \
    --interest-degree "$INTEREST_DEGREE"
assert_failure

test_name "Create objective stakeholder classification without interest degree fails"
xbe_json do objective-stakeholder-classifications create \
    --objective "${OBJECTIVE_ID:-123}" \
    --stakeholder-classification "$STAKEHOLDER_CLASSIFICATION_ID"
assert_failure

test_name "Update objective stakeholder classification without any fields fails"
if [[ $SKIP_OBJECTIVE_TESTS -eq 0 && -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do objective-stakeholder-classifications update "$CREATED_CLASSIFICATION_ID"
    assert_failure
else
    skip "No objective stakeholder classification ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
