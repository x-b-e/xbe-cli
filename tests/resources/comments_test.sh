#!/bin/bash
#
# XBE CLI Integration Tests: Comments
#
# Tests CRUD operations for the comments resource.
# Comments can be attached to various resources (projects, truckers, etc.).
#
# NOTE: This test requires creating prerequisite resources: broker and project
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_COMMENT_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""

describe "Resource: comments"

# ============================================================================
# Prerequisites - Create broker, developer, and project
# ============================================================================

test_name "Create prerequisite broker for comment tests"
BROKER_NAME=$(unique_name "CommentTestBroker")

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

test_name "Create prerequisite developer for comment tests"
DEVELOPER_NAME=$(unique_name "CommentTestDev")

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

test_name "Create prerequisite project for comment tests"
PROJECT_NAME=$(unique_name "CommentTestProject")

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

test_name "Create comment with required fields"

xbe_json do comments create \
    --body "Test comment body" \
    --commentable-type "projects" \
    --commentable-id "$CREATED_PROJECT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_COMMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_COMMENT_ID" && "$CREATED_COMMENT_ID" != "null" ]]; then
        register_cleanup "comments" "$CREATED_COMMENT_ID"
        pass
    else
        fail "Created comment but no ID returned"
    fi
else
    fail "Failed to create comment"
fi

# Only continue if we successfully created a comment
if [[ -z "$CREATED_COMMENT_ID" || "$CREATED_COMMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid comment ID"
    run_tests
fi

test_name "Create comment with --do-not-notify"
xbe_json do comments create \
    --body "Silent comment" \
    --commentable-type "projects" \
    --commentable-id "$CREATED_PROJECT_ID" \
    --do-not-notify

if [[ $status -eq 0 ]]; then
    SILENT_COMMENT_ID=$(json_get ".id")
    if [[ -n "$SILENT_COMMENT_ID" && "$SILENT_COMMENT_ID" != "null" ]]; then
        register_cleanup "comments" "$SILENT_COMMENT_ID"
        pass
    else
        fail "Created comment but no ID returned"
    fi
else
    fail "Failed to create comment with --do-not-notify"
fi

test_name "Create comment with --include-in-recap"
xbe_json do comments create \
    --body "Recap comment" \
    --commentable-type "projects" \
    --commentable-id "$CREATED_PROJECT_ID" \
    --include-in-recap

if [[ $status -eq 0 ]]; then
    RECAP_COMMENT_ID=$(json_get ".id")
    if [[ -n "$RECAP_COMMENT_ID" && "$RECAP_COMMENT_ID" != "null" ]]; then
        register_cleanup "comments" "$RECAP_COMMENT_ID"
        pass
    else
        fail "Created comment but no ID returned"
    fi
else
    fail "Failed to create comment with --include-in-recap"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update comment --body"
xbe_json do comments update "$CREATED_COMMENT_ID" --body "Updated comment body"
assert_success

test_name "Update comment --do-not-notify"
xbe_json do comments update "$CREATED_COMMENT_ID" --do-not-notify
assert_success

test_name "Update comment --include-in-recap"
xbe_json do comments update "$CREATED_COMMENT_ID" --include-in-recap
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List comments"
xbe_json view comments list --limit 5
assert_success

test_name "List comments returns array"
xbe_json view comments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list comments"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List comments with --commentable-type filter"
xbe_json view comments list --commentable-type "projects" --limit 10
assert_success

test_name "List comments with --commentable-type and --commentable-id filter"
xbe_json view comments list \
    --commentable-type "Project" \
    --commentable-id "$CREATED_PROJECT_ID" \
    --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List comments with --limit"
xbe_json view comments list --limit 3
assert_success

test_name "List comments with --offset"
xbe_json view comments list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete comment requires --confirm flag"
xbe_run do comments delete "$CREATED_COMMENT_ID"
assert_failure

test_name "Delete comment with --confirm"
# Create a comment specifically for deletion
xbe_json do comments create \
    --body "Comment to delete" \
    --commentable-type "projects" \
    --commentable-id "$CREATED_PROJECT_ID"

if [[ $status -eq 0 ]]; then
    DEL_COMMENT_ID=$(json_get ".id")
    if [[ -n "$DEL_COMMENT_ID" && "$DEL_COMMENT_ID" != "null" ]]; then
        xbe_run do comments delete "$DEL_COMMENT_ID" --confirm
        assert_success
    else
        skip "Could not create comment for deletion test"
    fi
else
    skip "Could not create comment for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create comment without --body fails"
xbe_json do comments create \
    --commentable-type "projects" \
    --commentable-id "$CREATED_PROJECT_ID"
assert_failure

test_name "Create comment without --commentable-type fails"
xbe_json do comments create \
    --body "Missing type" \
    --commentable-id "$CREATED_PROJECT_ID"
assert_failure

test_name "Create comment without --commentable-id fails"
xbe_json do comments create \
    --body "Missing ID" \
    --commentable-type "projects"
assert_failure

test_name "Update comment without any fields fails"
xbe_run do comments update "$CREATED_COMMENT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
