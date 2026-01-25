#!/bin/bash
#
# XBE CLI Integration Tests: Project Completions
#
# Tests list, show, and create operations for the project-completions resource.
#
# COVERAGE: Create + list + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
LIST_SUPPORTED="false"
APPROVED_PROJECT_ID="${XBE_TEST_APPROVED_PROJECT_ID:-}"

describe "Resource: project-completions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project completions"
xbe_json view project-completions list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Project completions list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project completions returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-completions list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project completions"
    fi
else
    skip "Project completions list endpoint not available"
fi

# ==========================================================================
# Sample Record (used for show)
# ==========================================================================

test_name "Capture sample project completion"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-completions list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project completions available for show"
        fi
    else
        skip "Could not list project completions to capture sample"
    fi
else
    skip "Project completions list endpoint not available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project completion"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-completions show "$SAMPLE_ID"
    assert_success
else
    skip "No project completion ID available"
fi

# ============================================================================
# Prerequisites - Find an approved project
# ============================================================================

if [[ -z "$APPROVED_PROJECT_ID" ]]; then
    test_name "Find approved project for completion"
    xbe_json view projects list --status approved --limit 1
    if [[ $status -eq 0 ]]; then
        APPROVED_PROJECT_ID=$(json_get ".[0].id")
        if [[ -n "$APPROVED_PROJECT_ID" && "$APPROVED_PROJECT_ID" != "null" ]]; then
            pass
        else
            skip "No approved projects available (set XBE_TEST_APPROVED_PROJECT_ID to override)"
        fi
    else
        fail "Failed to list approved projects"
    fi
else
    test_name "Using approved project from XBE_TEST_APPROVED_PROJECT_ID"
    echo "    Using XBE_TEST_APPROVED_PROJECT_ID: $APPROVED_PROJECT_ID"
    pass
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create completion without required project fails"
xbe_run do project-completions create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project completion"
if [[ -n "$APPROVED_PROJECT_ID" && "$APPROVED_PROJECT_ID" != "null" ]]; then
    COMMENT="Completing project for test"

    xbe_json do project-completions create \
        --project "$APPROVED_PROJECT_ID" \
        --comment "$COMMENT"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".project_id" "$APPROVED_PROJECT_ID"
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create project completion"
    fi
else
    skip "No approved project available for completion"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
