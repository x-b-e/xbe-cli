#!/bin/bash
#
# XBE CLI Integration Tests: Project Abandonments
#
# Tests list, show, and create operations for the project-abandonments resource.
#
# COVERAGE: Create + list + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""

SAMPLE_ID=""
LIST_SUPPORTED="false"

describe "Resource: project-abandonments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project abandonments"
xbe_json view project-abandonments list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Project abandonments list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project abandonments returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-abandonments list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project abandonments"
    fi
else
    skip "Project abandonments list endpoint not available"
fi

# ==========================================================================
# Sample Record (used for show)
# ==========================================================================

test_name "Capture sample project abandonment"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-abandonments list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project abandonments available for show"
        fi
    else
        skip "Could not list project abandonments to capture sample"
    fi
else
    skip "Project abandonments list endpoint not available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project abandonment"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-abandonments show "$SAMPLE_ID"
    assert_success
else
    skip "No project abandonment ID available"
fi

# ============================================================================
# Prerequisites - Create broker, developer, and project
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "ProjectAbandonBroker")

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

test_name "Create prerequisite developer"
DEVELOPER_NAME=$(unique_name "ProjectAbandonDeveloper")

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

test_name "Create project for abandonment"
PROJECT_NAME=$(unique_name "ProjectAbandon")

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
# Error Cases
# ============================================================================

test_name "Create abandonment without required project fails"
xbe_run do project-abandonments create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project abandonment"
COMMENT="Abandoning project for test"

xbe_json do project-abandonments create \
    --project "$CREATED_PROJECT_ID" \
    --comment "$COMMENT"

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".project_id" "$CREATED_PROJECT_ID"
    assert_json_equals ".comment" "$COMMENT"
else
    fail "Failed to create project abandonment"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
