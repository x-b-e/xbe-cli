#!/bin/bash
#
# XBE CLI Integration Tests: Projects File Imports
#
# Tests list/show/create behavior for projects file imports.
# Requires a file import ID and supported file import type for create.
#
# COVERAGE: List + filters + required attributes + optional flags
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: projects-file-imports"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List projects file imports"
xbe_json view projects-file-imports list
assert_success

test_name "List projects file imports returns array"
xbe_json view projects-file-imports list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list projects file imports"
fi

# ============================================================================
# LIST Filters
# ============================================================================

test_name "List projects file imports with --is-dry-run filter"
xbe_json view projects-file-imports list --is-dry-run true
assert_success

FILE_IMPORT_ID="${XBE_TEST_FILE_IMPORT_ID:-}"
FILE_IMPORT_TYPE="${XBE_TEST_PROJECTS_FILE_IMPORT_TYPE:-}"
SUBJECT_TYPE="${XBE_TEST_PROJECTS_FILE_IMPORT_SUBJECT_TYPE:-}"
SUBJECT_ID="${XBE_TEST_PROJECTS_FILE_IMPORT_SUBJECT_ID:-}"

if [[ -n "$FILE_IMPORT_TYPE" ]]; then
    test_name "List projects file imports with --file-import-type filter"
    xbe_json view projects-file-imports list --file-import-type "$FILE_IMPORT_TYPE"
    assert_success
else
    test_name "Skip --file-import-type filter (missing XBE_TEST_PROJECTS_FILE_IMPORT_TYPE)"
    skip "Set XBE_TEST_PROJECTS_FILE_IMPORT_TYPE to test file-import-type filter"
fi

if [[ -n "$FILE_IMPORT_ID" ]]; then
    test_name "List projects file imports with --file-import filter"
    xbe_json view projects-file-imports list --file-import "$FILE_IMPORT_ID"
    assert_success
else
    test_name "Skip --file-import filter (missing XBE_TEST_FILE_IMPORT_ID)"
    skip "Set XBE_TEST_FILE_IMPORT_ID to test file-import filter"
fi

if [[ -n "$SUBJECT_TYPE" && -n "$SUBJECT_ID" ]]; then
    test_name "List projects file imports with --subject filter"
    xbe_json view projects-file-imports list --subject-type "$SUBJECT_TYPE" --subject-id "$SUBJECT_ID"
    assert_success
else
    test_name "Skip --subject filter (missing subject env vars)"
    skip "Set XBE_TEST_PROJECTS_FILE_IMPORT_SUBJECT_TYPE and XBE_TEST_PROJECTS_FILE_IMPORT_SUBJECT_ID to test subject filter"
fi

# ============================================================================
# CREATE Tests (optional)
# ============================================================================

if [[ -z "$FILE_IMPORT_ID" || -z "$FILE_IMPORT_TYPE" ]]; then
    test_name "Skip create tests (missing projects file import prerequisites)"
    skip "Set XBE_TEST_FILE_IMPORT_ID and XBE_TEST_PROJECTS_FILE_IMPORT_TYPE to run create tests"
else
    test_name "Create projects file import (required fields)"
    xbe_json do projects-file-imports create \
        --file-import "$FILE_IMPORT_ID" \
        --file-import-type "$FILE_IMPORT_TYPE"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create projects file import"
    fi

    test_name "Create projects file import with dry run"
    xbe_json do projects-file-imports create \
        --file-import "$FILE_IMPORT_ID" \
        --file-import-type "$FILE_IMPORT_TYPE" \
        --is-dry-run

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_bool ".is_dry_run" "true"
    else
        fail "Failed to create projects file import with dry run"
    fi
fi

# ============================================================================
# SHOW Tests (optional)
# ============================================================================

if [[ -n "$XBE_TEST_PROJECTS_FILE_IMPORT_ID" ]]; then
    test_name "Show projects file import"
    xbe_json view projects-file-imports show "$XBE_TEST_PROJECTS_FILE_IMPORT_ID"
    assert_success
else
    test_name "Skip show tests (missing XBE_TEST_PROJECTS_FILE_IMPORT_ID)"
    skip "Set XBE_TEST_PROJECTS_FILE_IMPORT_ID to run show tests"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Fail create without file-import"
xbe_run do projects-file-imports create --file-import-type "SageProjectsFileImport"
assert_failure

test_name "Fail create without file-import-type"
xbe_run do projects-file-imports create --file-import "1"
assert_failure

test_name "Fail create with subject-type only"
xbe_run do projects-file-imports create --file-import "1" --file-import-type "SageProjectsFileImport" --subject-type brokers
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
