#!/bin/bash
#
# XBE CLI Integration Tests: File Attachments
#
# Tests CRUD operations for the file-attachments resource.
#
# NOTE: This test requires creating prerequisite resources: broker, developer, project
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FILE_ATTACHMENT_ID=""
CREATED_ATTACHED_FILE_ATTACHMENT_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_BY_ID=""

describe "Resource: file-attachments"

# ============================================================================
# Prerequisites - Create broker, developer, and project
# ============================================================================

test_name "Create prerequisite broker for file attachment tests"
BROKER_NAME=$(unique_name "FileAttachmentTestBroker")

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

test_name "Create prerequisite developer for file attachment tests"
DEVELOPER_NAME=$(unique_name "FileAttachmentTestDev")

xbe_json do developers create \
    --name "$DEVELOPER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

test_name "Create prerequisite project for file attachment tests"
PROJECT_NAME=$(unique_name "FileAttachmentTestProject")

xbe_json do projects create \
    --name "$PROJECT_NAME" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        echo "Cannot continue without a project"
        run_tests
    fi
else
    fail "Failed to create project"
    echo "Cannot continue without a project"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create file attachment with required fields"
FILE_SUFFIX=$(unique_suffix)
FILE_NAME="file-attachment-${FILE_SUFFIX}.txt"
OBJECT_KEY="uploads/${FILE_SUFFIX}/${FILE_NAME}"

xbe_json do file-attachments create \
    --file-name "$FILE_NAME" \
    --object-key "$OBJECT_KEY"

if [[ $status -eq 0 ]]; then
    CREATED_FILE_ATTACHMENT_ID=$(json_get ".id")
    CREATED_BY_ID=$(json_get ".created_by_id")
    if [[ -n "$CREATED_FILE_ATTACHMENT_ID" && "$CREATED_FILE_ATTACHMENT_ID" != "null" ]]; then
        register_cleanup "file-attachments" "$CREATED_FILE_ATTACHMENT_ID"
        pass
    else
        fail "Created file attachment but no ID returned"
    fi
else
    fail "Failed to create file attachment"
fi

# Only continue if we successfully created a file attachment
if [[ -z "$CREATED_FILE_ATTACHMENT_ID" || "$CREATED_FILE_ATTACHMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid file attachment ID"
    run_tests
fi

# Fetch created-by ID if not returned on create
if [[ -z "$CREATED_BY_ID" || "$CREATED_BY_ID" == "null" ]]; then
    xbe_json view file-attachments show "$CREATED_FILE_ATTACHMENT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".created_by_id")
    fi
fi

if [[ -z "$CREATED_BY_ID" || "$CREATED_BY_ID" == "null" ]]; then
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".id")
    fi
fi

test_name "Create file attachment attached to a project"
FILE_SUFFIX=$(unique_suffix)
ATTACHED_FILE_NAME="project-attachment-${FILE_SUFFIX}.pdf"
ATTACHED_OBJECT_KEY="uploads/${FILE_SUFFIX}/${ATTACHED_FILE_NAME}"

xbe_json do file-attachments create \
    --file-name "$ATTACHED_FILE_NAME" \
    --object-key "$ATTACHED_OBJECT_KEY" \
    --attached-to-type projects \
    --attached-to-id "$CREATED_PROJECT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ATTACHED_FILE_ATTACHMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_ATTACHED_FILE_ATTACHMENT_ID" && "$CREATED_ATTACHED_FILE_ATTACHMENT_ID" != "null" ]]; then
        register_cleanup "file-attachments" "$CREATED_ATTACHED_FILE_ATTACHMENT_ID"
        pass
    else
        fail "Created attached file attachment but no ID returned"
    fi
else
    fail "Failed to create attached file attachment"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update file attachment --file-name"
UPDATED_FILE_NAME="updated-${FILE_NAME}"

xbe_json do file-attachments update "$CREATED_FILE_ATTACHMENT_ID" \
    --file-name "$UPDATED_FILE_NAME"
assert_success

test_name "Update file attachment --object-key"
UPDATED_OBJECT_KEY="uploads/${FILE_SUFFIX}/updated-${FILE_NAME}"

xbe_json do file-attachments update "$CREATED_FILE_ATTACHMENT_ID" \
    --object-key "$UPDATED_OBJECT_KEY"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List file attachments"
xbe_json view file-attachments list --limit 5
assert_success

test_name "List file attachments returns array"
xbe_json view file-attachments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list file attachments"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List file attachments with --attached-to-type and --attached-to-id"
xbe_json view file-attachments list \
    --attached-to-type Project \
    --attached-to-id "$CREATED_PROJECT_ID" \
    --limit 10
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "List file attachments with --created-by"
    xbe_json view file-attachments list --created-by "$CREATED_BY_ID" --limit 10
    assert_success
else
    skip "created-by ID not available for filter test"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List file attachments with --limit"
xbe_json view file-attachments list --limit 3
assert_success

test_name "List file attachments with --offset"
xbe_json view file-attachments list --limit 3 --offset 3
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show file attachment"
xbe_json view file-attachments show "$CREATED_FILE_ATTACHMENT_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete file attachment requires --confirm flag"
xbe_run do file-attachments delete "$CREATED_FILE_ATTACHMENT_ID"
assert_failure

test_name "Delete file attachment with --confirm"
FILE_SUFFIX=$(unique_suffix)
DEL_FILE_NAME="delete-attachment-${FILE_SUFFIX}.txt"
DEL_OBJECT_KEY="uploads/${FILE_SUFFIX}/${DEL_FILE_NAME}"

xbe_json do file-attachments create \
    --file-name "$DEL_FILE_NAME" \
    --object-key "$DEL_OBJECT_KEY"

if [[ $status -eq 0 ]]; then
    DEL_FILE_ATTACHMENT_ID=$(json_get ".id")
    if [[ -n "$DEL_FILE_ATTACHMENT_ID" && "$DEL_FILE_ATTACHMENT_ID" != "null" ]]; then
        xbe_run do file-attachments delete "$DEL_FILE_ATTACHMENT_ID" --confirm
        assert_success
    else
        skip "Could not create file attachment for deletion test"
    fi
else
    skip "Could not create file attachment for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create file attachment without --file-name fails"
xbe_json do file-attachments create --object-key "uploads/missing/name.txt"
assert_failure

test_name "Create file attachment without --object-key fails"
xbe_json do file-attachments create --file-name "missing-object-key.txt"
assert_failure

test_name "Update file attachment without any fields fails"
xbe_run do file-attachments update "$CREATED_FILE_ATTACHMENT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
