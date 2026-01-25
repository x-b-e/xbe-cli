#!/bin/bash
#
# XBE CLI Integration Tests: Project Rejections
#
# Tests create operations for project-rejections.
#
# COVERAGE: Writable attributes (comment)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

REJECTABLE_PROJECT_ID="${XBE_TEST_PROJECT_REJECTION_ID:-}"

describe "Resource: project-rejections"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create rejection requires project"
xbe_run do project-rejections create --comment "missing project"
assert_failure

test_name "Create project rejection"
if [[ -n "$REJECTABLE_PROJECT_ID" && "$REJECTABLE_PROJECT_ID" != "null" ]]; then
    COMMENT=$(unique_name "ProjectRejection")
    xbe_json do project-rejections create \
        --project "$REJECTABLE_PROJECT_ID" \
        --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".project_id" "$REJECTABLE_PROJECT_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create project rejection: $output"
        fi
    fi
else
    skip "No submitted project ID available for rejection. Set XBE_TEST_PROJECT_REJECTION_ID to enable create testing."
fi

run_tests
