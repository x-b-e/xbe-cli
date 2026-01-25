#!/bin/bash
#
# XBE CLI Integration Tests: File Attachment Signed URLs
#
# Tests create operations for the file_attachment_signed_urls resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

FILE_ATTACHMENT_ID=""

describe "Resource: file-attachment-signed-urls"

# ============================================================================
# Prerequisites - Find a file attachment from action items
# ============================================================================

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

test_name "Create signed URL without required file attachment fails"
xbe_run do file-attachment-signed-urls create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create signed URL"
if [[ -n "$FILE_ATTACHMENT_ID" && "$FILE_ATTACHMENT_ID" != "null" ]]; then
    xbe_json do file-attachment-signed-urls create --file-attachment-id "$FILE_ATTACHMENT_ID"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".signed_url"
        assert_json_equals ".file_attachment_id" "$FILE_ATTACHMENT_ID"
    else
        fail "Failed to create signed URL"
    fi
else
    skip "No file attachment available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
