#!/bin/bash
#
# XBE CLI Integration Tests: Taggings
#
# Tests create, list, show, and delete operations for the taggings resource.
#
# COVERAGE: All create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TAG_CATEGORY_ID=""
CREATED_TAG_ID=""
CREATED_TAGGING_ID=""
PREDICTION_SUBJECT_ID=""

describe "Resource: taggings"

# ============================================================================
# Prerequisites - Tag category + tag + prediction subject
# ============================================================================

test_name "Create tag category for taggings"
TEST_CAT_NAME=$(unique_name "TaggingsCat")
TEST_CAT_SLUG="taggings-$(date +%s | tail -c 6)"
xbe_json do tag-categories create \
    --name "$TEST_CAT_NAME" \
    --slug "$TEST_CAT_SLUG" \
    --can-apply-to "PredictionSubject"

if [[ $status -eq 0 ]]; then
    CREATED_TAG_CATEGORY_ID=$(json_get ".id")
    if [[ -n "$CREATED_TAG_CATEGORY_ID" && "$CREATED_TAG_CATEGORY_ID" != "null" ]]; then
        register_cleanup "tag-categories" "$CREATED_TAG_CATEGORY_ID"
        pass
    else
        fail "Created tag category but no ID returned"
    fi
else
    fail "Failed to create tag category"
fi

test_name "Create tag for taggings"
TEST_TAG_NAME=$(unique_name "TaggingsTag")
xbe_json do tags create --name "$TEST_TAG_NAME" --tag-category "$CREATED_TAG_CATEGORY_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TAG_ID=$(json_get ".id")
    if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" ]]; then
        register_cleanup "tags" "$CREATED_TAG_ID"
        pass
    else
        fail "Created tag but no ID returned"
    fi
else
    fail "Failed to create tag"
fi

test_name "Fetch prediction subject ID for tagging"
if [[ -n "$XBE_TEST_PREDICTION_SUBJECT_ID" ]]; then
    PREDICTION_SUBJECT_ID="$XBE_TEST_PREDICTION_SUBJECT_ID"
    pass
else
    xbe_json view predictions list --limit 10
    if [[ $status -eq 0 ]]; then
        PREDICTION_SUBJECT_ID=$(echo "$output" | jq -r 'map(select(.prediction_subject_id != null and .prediction_subject_id != "")) | .[0].prediction_subject_id')
        if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
            pass
        else
            skip "No prediction subject found; set XBE_TEST_PREDICTION_SUBJECT_ID to run tagging create tests"
        fi
    else
        skip "Failed to list predictions; cannot locate prediction subject"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" && -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    test_name "Create tagging with required fields"
    xbe_json do taggings create \
        --tag "$CREATED_TAG_ID" \
        --taggable-type "prediction-subjects" \
        --taggable-id "$PREDICTION_SUBJECT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_TAGGING_ID=$(json_get ".id")
        if [[ -n "$CREATED_TAGGING_ID" && "$CREATED_TAGGING_ID" != "null" ]]; then
            register_cleanup "taggings" "$CREATED_TAGGING_ID"
            pass
        else
            fail "Created tagging but no ID returned"
        fi
    else
        fail "Failed to create tagging"
    fi
else
    skip "Missing tag or prediction subject; skipping tagging create tests"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_TAGGING_ID" && "$CREATED_TAGGING_ID" != "null" ]]; then
    test_name "Show tagging"
    xbe_json view taggings show "$CREATED_TAGGING_ID"
    assert_success
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List taggings"
xbe_json view taggings list --limit 5
assert_success

test_name "List taggings returns array"
xbe_json view taggings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list taggings"
fi

if [[ -n "$CREATED_TAG_ID" && "$CREATED_TAG_ID" != "null" ]]; then
    test_name "List taggings with --tag-id filter"
    xbe_json view taggings list --tag-id "$CREATED_TAG_ID" --limit 10
    assert_success
fi

if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    test_name "List taggings with --taggable-type filter"
    xbe_json view taggings list --taggable-type "PredictionSubject" --limit 10
    assert_success

    test_name "List taggings with --taggable-type and --taggable-id filter"
    xbe_json view taggings list --taggable-type "PredictionSubject" --taggable-id "$PREDICTION_SUBJECT_ID" --limit 10
    assert_success
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_TAGGING_ID" && "$CREATED_TAGGING_ID" != "null" ]]; then
    test_name "Delete tagging requires --confirm flag"
    xbe_run do taggings delete "$CREATED_TAGGING_ID"
    assert_failure

    test_name "Delete tagging with --confirm"
    xbe_run do taggings delete "$CREATED_TAGGING_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
