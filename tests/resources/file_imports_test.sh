#!/bin/bash
#
# XBE CLI Integration Tests: File Imports
#
# Tests CRUD operations for the file_imports resource.
#
# COVERAGE: Create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FILE_IMPORT_ID=""
CREATED_BROKER_ID=""
FILE_ATTACHMENT_ID=""
CREATED_BY_ID=""

describe "Resource: file_imports"

# ============================================================================
# Prerequisites - Create broker and find a file attachment
# ============================================================================

test_name "Create prerequisite broker for file import tests"
BROKER_NAME=$(unique_name "FileImportBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Find action item with attachment"
xbe_json view action-items list --limit 10

if [[ $status -eq 0 ]]; then
    action_item_ids=$(echo "$output" | jq -r '.[].id' 2>/dev/null)
    for id in $action_item_ids; do
        xbe_json view action-items show "$id"
        if [[ $status -eq 0 ]]; then
            file_id=$(json_get ".attachments[0].id")
            if [[ -n "$file_id" && "$file_id" != "null" ]]; then
                FILE_ATTACHMENT_ID="$file_id"
                break
            fi
        fi
    done

    if [[ -n "$FILE_ATTACHMENT_ID" && "$FILE_ATTACHMENT_ID" != "null" ]]; then
        pass
    else
        skip "No action items with attachments available"
    fi
else
    fail "Failed to list action items"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create file import without required flags fails"
xbe_run do file-imports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create file import with required fields"
if [[ -n "$FILE_ATTACHMENT_ID" && "$FILE_ATTACHMENT_ID" != "null" ]]; then
    xbe_json do file-imports create \
        --broker "$CREATED_BROKER_ID" \
        --file-attachment "$FILE_ATTACHMENT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_FILE_IMPORT_ID=$(json_get ".id")
        if [[ -n "$CREATED_FILE_IMPORT_ID" && "$CREATED_FILE_IMPORT_ID" != "null" ]]; then
            register_cleanup "file-imports" "$CREATED_FILE_IMPORT_ID"
            pass
        else
            fail "Created file import but no ID returned"
        fi
    else
        fail "Failed to create file import"
    fi
else
    skip "No file attachment available"
fi

# Only continue if we successfully created a file import
if [[ -z "$CREATED_FILE_IMPORT_ID" || "$CREATED_FILE_IMPORT_ID" == "null" ]]; then
    echo "Cannot continue without a valid file import ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update file import note"
UPDATED_NOTE="CLI file import note"
xbe_json do file-imports update "$CREATED_FILE_IMPORT_ID" --note "$UPDATED_NOTE"
assert_success

test_name "Update file import processed-at"
UPDATED_PROCESSED_AT="2024-01-02T03:04:05Z"
xbe_json do file-imports update "$CREATED_FILE_IMPORT_ID" --processed-at "$UPDATED_PROCESSED_AT"
assert_success

test_name "Show file import details"
xbe_json view file-imports show "$CREATED_FILE_IMPORT_ID"
if [[ $status -eq 0 ]]; then
    CREATED_BY_ID=$(json_get ".created_by_id")
    note_value=$(json_get ".note")
    if [[ "$note_value" == "$UPDATED_NOTE" ]]; then
        pass
    else
        fail "Expected note to be updated"
    fi

    if echo "$output" | jq -e --arg date "2024-01-02" '.processed_at | contains($date)' > /dev/null; then
        pass
    else
        fail "Expected processed_at to include 2024-01-02"
    fi
else
    fail "Failed to show file import"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List file imports filtered by broker"
xbe_json view file-imports list --broker "$CREATED_BROKER_ID"
if echo "$output" | jq -e --arg id "$CREATED_FILE_IMPORT_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
    pass
else
    fail "Expected file import in broker filter results"
fi

test_name "List file imports filtered by file attachment"
xbe_json view file-imports list --file-attachment "$FILE_ATTACHMENT_ID"
if echo "$output" | jq -e --arg id "$CREATED_FILE_IMPORT_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
    pass
else
    fail "Expected file import in file attachment filter results"
fi

test_name "List file imports filtered by created-by"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view file-imports list --created-by "$CREATED_BY_ID"
    if echo "$output" | jq -e --arg id "$CREATED_FILE_IMPORT_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
        pass
    else
        fail "Expected file import in created-by filter results"
    fi
else
    skip "No created-by ID available from show output"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete file import"
xbe_run do file-imports delete "$CREATED_FILE_IMPORT_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
