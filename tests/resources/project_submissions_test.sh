#!/bin/bash
#
# XBE CLI Integration Tests: Project Submissions
#
# Tests view and create operations for project_submissions.
# Submissions transition projects from editing or rejected to submitted.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_COMMENT_PROJECT_ID=""
SAMPLE_SUBMISSION_ID=""
SKIP_MUTATION=0

describe "Resource: project-submissions"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List project submissions"
xbe_json view project-submissions list --limit 1
assert_success

test_name "Capture sample project submission (if available)"
xbe_json view project-submissions list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_SUBMISSION_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No project submissions available; skipping show test."
        pass
    fi
else
    fail "Failed to list project submissions"
fi

if [[ -n "$SAMPLE_SUBMISSION_ID" && "$SAMPLE_SUBMISSION_ID" != "null" ]]; then
    test_name "Show project submission"
    xbe_json view project-submissions show "$SAMPLE_SUBMISSION_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create project submission requires --project"
xbe_run do project-submissions create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create prerequisite broker for project submission tests"
    BROKER_NAME=$(unique_name "ProjectSubmissionBroker")

    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        fi
    fi

    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        test_name "Create prerequisite developer"
        if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
            CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
            echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
            pass
        else
            DEV_NAME=$(unique_name "ProjectSubmissionDev")
            xbe_json do developers create \
                --name "$DEV_NAME" \
                --broker "$CREATED_BROKER_ID"

            if [[ $status -eq 0 ]]; then
                CREATED_DEVELOPER_ID=$(json_get ".id")
                if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
                    register_cleanup "developers" "$CREATED_DEVELOPER_ID"
                    pass
                else
                    fail "Created developer but no ID returned"
                fi
            else
                fail "Failed to create developer"
            fi
        fi
    fi

    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        test_name "Create project for submission (minimal)"
        PROJECT_NAME=$(unique_name "ProjectSubmission")

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
            fi
        else
            fail "Failed to create project"
        fi
    fi

    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        test_name "Create project for submission with comment"
        PROJECT_NAME=$(unique_name "ProjectSubmissionComment")

        xbe_json do projects create \
            --name "$PROJECT_NAME" \
            --developer "$CREATED_DEVELOPER_ID"

        if [[ $status -eq 0 ]]; then
            CREATED_COMMENT_PROJECT_ID=$(json_get ".id")
            if [[ -n "$CREATED_COMMENT_PROJECT_ID" && "$CREATED_COMMENT_PROJECT_ID" != "null" ]]; then
                register_cleanup "projects" "$CREATED_COMMENT_PROJECT_ID"
                pass
            else
                fail "Created project but no ID returned"
            fi
        else
            fail "Failed to create project for comment test"
        fi
    fi
else
    echo "    (Missing prerequisites; skipping submission creation)"
fi

if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
    test_name "Create project submission (minimal)"
    xbe_json do project-submissions create --project "$CREATED_PROJECT_ID"
    assert_success
else
    skip "No project available for submission tests"
fi

if [[ -n "$CREATED_COMMENT_PROJECT_ID" && "$CREATED_COMMENT_PROJECT_ID" != "null" ]]; then
    test_name "Create project submission with comment"
    COMMENT_TEXT="Submitted by CLI test"
    xbe_json do project-submissions create \
        --project "$CREATED_COMMENT_PROJECT_ID" \
        --comment "$COMMENT_TEXT"

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT_TEXT"
    else
        fail "Failed to create submission with comment"
    fi
else
    skip "No project available for submission comment test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project submission with invalid ID fails"
xbe_run do project-submissions create --project "999999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
