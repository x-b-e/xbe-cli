#!/bin/bash
#
# XBE CLI Integration Tests: Project Estimate File Imports
#
# Tests list and create behavior for project estimate file imports.
# Requires a project ID, file import ID, and supported file import type.
#
# COVERAGE: List + required attributes + optional flags
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-estimate-file-imports"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List project estimate file imports"
xbe_json view project-estimate-file-imports list
assert_success

test_name "List project estimate file imports returns array"
xbe_json view project-estimate-file-imports list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project estimate file imports"
fi

# ============================================================================
# CREATE Tests (optional)
# ============================================================================

PROJECT_ID="${XBE_TEST_PROJECT_ID:-}"
FILE_IMPORT_ID="${XBE_TEST_FILE_IMPORT_ID:-}"
FILE_IMPORT_TYPE="${XBE_TEST_PROJECT_ESTIMATE_FILE_IMPORT_TYPE:-}"

if [[ -z "$PROJECT_ID" || -z "$FILE_IMPORT_ID" || -z "$FILE_IMPORT_TYPE" ]]; then
    test_name "Skip create tests (missing project estimate file import prerequisites)"
    skip "Set XBE_TEST_PROJECT_ID, XBE_TEST_FILE_IMPORT_ID, and XBE_TEST_PROJECT_ESTIMATE_FILE_IMPORT_TYPE to run create tests"
else
    test_name "Create project estimate file import (required fields)"
    xbe_json do project-estimate-file-imports create \
        --project "$PROJECT_ID" \
        --file-import "$FILE_IMPORT_ID" \
        --file-import-type "$FILE_IMPORT_TYPE"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create project estimate file import"
    fi

    test_name "Create project estimate file import with dry run and extraction update"
    xbe_json do project-estimate-file-imports create \
        --project "$PROJECT_ID" \
        --file-import "$FILE_IMPORT_ID" \
        --file-import-type "$FILE_IMPORT_TYPE" \
        --is-dry-run \
        --should-update-file-extraction

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_bool ".is_dry_run" "true"
        assert_json_bool ".should_update_file_extraction" "true"
    else
        fail "Failed to create project estimate file import with flags"
    fi
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Fail create without file-import"
xbe_run do project-estimate-file-imports create --project "1" --file-import-type "Bid2Win"
assert_failure

test_name "Fail create without project"
xbe_run do project-estimate-file-imports create --file-import "1" --file-import-type "Bid2Win"
assert_failure

test_name "Fail create without file-import-type"
xbe_run do project-estimate-file-imports create --project "1" --file-import "1"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
