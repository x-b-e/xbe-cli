#!/bin/bash
#
# XBE CLI Integration Tests: Project Import File Verifications
#
# Tests list and create operations for the project-import-file-verifications resource.
#
# COVERAGE: list filter + create (dry run) + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PROJECT_ID=""
FILE_IMPORT_ID=""
VERIFICATION_TYPE=""

describe "Resource: project-import-file-verifications"

# ============================================================================
# Prerequisites - Find a project
# ============================================================================

test_name "Find project for verification list"
xbe_json view projects list --limit 10
if [[ $status -eq 0 ]]; then
    PROJECT_ID=$(json_get ".[0].id")
    if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
        pass
    else
        skip "No projects available"
    fi
else
    skip "Could not list projects"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List project import file verifications"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view project-import-file-verifications list --project "$PROJECT_ID"
    assert_success
else
    skip "No project available"
fi

test_name "List project import file verifications returns array"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view project-import-file-verifications list --project "$PROJECT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list verifications"
    fi
else
    skip "No project available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Find verification to reuse for create"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view project-import-file-verifications list --project "$PROJECT_ID"
    if [[ $status -eq 0 ]]; then
        FILE_IMPORT_ID=$(json_get ".[0].file_import_id")
        VERIFICATION_TYPE=$(json_get ".[0].verification_type")
    fi

    if [[ -z "$FILE_IMPORT_ID" || "$FILE_IMPORT_ID" == "null" || -z "$VERIFICATION_TYPE" || "$VERIFICATION_TYPE" == "null" ]]; then
        xbe_json view projects list --limit 25
        if [[ $status -eq 0 ]]; then
            for pid in $(echo "$output" | jq -r '.[].id'); do
                xbe_json view project-import-file-verifications list --project "$pid"
                if [[ $status -ne 0 ]]; then
                    continue
                fi
                FILE_IMPORT_ID=$(json_get ".[0].file_import_id")
                VERIFICATION_TYPE=$(json_get ".[0].verification_type")
                if [[ -n "$FILE_IMPORT_ID" && "$FILE_IMPORT_ID" != "null" && -n "$VERIFICATION_TYPE" && "$VERIFICATION_TYPE" != "null" ]]; then
                    PROJECT_ID="$pid"
                    break
                fi
            done
        fi
    fi

    if [[ -n "$FILE_IMPORT_ID" && "$FILE_IMPORT_ID" != "null" && -n "$VERIFICATION_TYPE" && "$VERIFICATION_TYPE" != "null" ]]; then
        pass
    else
        skip "No existing verifications available"
    fi
else
    skip "No project available"
fi

test_name "Create project import file verification (dry run)"
if [[ -n "$FILE_IMPORT_ID" && "$FILE_IMPORT_ID" != "null" && -n "$VERIFICATION_TYPE" && "$VERIFICATION_TYPE" != "null" ]]; then
    xbe_json do project-import-file-verifications create \
        --verification-type "$VERIFICATION_TYPE" \
        --file-import "$FILE_IMPORT_ID" \
        --project "$PROJECT_ID" \
        --is-dry-run
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No existing verification data to reuse"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "List verifications without project filter fails"
xbe_run view project-import-file-verifications list
assert_failure

test_name "Create verification without required flags fails"
xbe_run do project-import-file-verifications create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
