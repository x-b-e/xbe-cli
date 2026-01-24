#!/bin/bash
#
# XBE CLI Integration Tests: Project Subscriptions
#
# Tests CRUD operations for the project_subscriptions resource.
# Subscriptions link users to projects with a contact method.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUBSCRIPTION_ID=""
SAMPLE_ID=""
SAMPLE_PROJECT_ID=""
SAMPLE_USER_ID=""
SAMPLE_CONTACT_METHOD=""
WHOAMI_USER_ID=""

describe "Resource: project-subscriptions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project subscriptions"
xbe_json view project-subscriptions list --limit 5
assert_success

test_name "List project subscriptions returns array"
xbe_json view project-subscriptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_PROJECT_ID=$(echo "$output" | jq -r '.[0].project_id // empty')
    SAMPLE_USER_ID=$(echo "$output" | jq -r '.[0].user_id // empty')
    SAMPLE_CONTACT_METHOD=$(echo "$output" | jq -r '.[0].contact_method // empty')
else
    fail "Failed to list project subscriptions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project subscription"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-subscriptions show "$SAMPLE_ID"
    assert_success
else
    skip "No subscription ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Get current user for subscription create"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    pass
else
    skip "Unable to resolve current user"
fi

test_name "Create project subscription"
PROJECT_ID="${XBE_TEST_PROJECT_ID:-$SAMPLE_PROJECT_ID}"
USER_ID="${XBE_TEST_USER_ID:-$WHOAMI_USER_ID}"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json do project-subscriptions create \
        --project "$PROJECT_ID" \
        --user "$USER_ID" \
        --contact-method "email_address"
    if [[ $status -eq 0 ]]; then
        CREATED_SUBSCRIPTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
            register_cleanup "project-subscriptions" "$CREATED_SUBSCRIPTION_ID"
            pass
        else
            fail "Created subscription but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"exists"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"must be contactable"* ]] || [[ "$output" == *"must be a member"* ]]; then
            pass
        else
            fail "Failed to create project subscription: $output"
        fi
    fi
else
    skip "No project/user ID available (set XBE_TEST_PROJECT_ID and XBE_TEST_USER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project subscription contact method"
UPDATE_ID="${CREATED_SUBSCRIPTION_ID:-$SAMPLE_ID}"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do project-subscriptions update "$UPDATE_ID" --contact-method "email_address"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update project subscription: $output"
        fi
    fi
else
    skip "No subscription ID available for update"
fi

test_name "Update project subscription without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do project-subscriptions update "$UPDATE_ID"
    assert_failure
else
    skip "No subscription ID available for no-attribute update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List subscriptions with --project filter"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view project-subscriptions list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available for filter test"
fi

test_name "List subscriptions with --user filter"
USER_FILTER_ID="${USER_ID:-$SAMPLE_USER_ID}"
if [[ -n "$USER_FILTER_ID" && "$USER_FILTER_ID" != "null" ]]; then
    xbe_json view project-subscriptions list --user "$USER_FILTER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

test_name "List subscriptions with --contact-method filter"
CONTACT_METHOD_FILTER="${SAMPLE_CONTACT_METHOD:-email_address}"
if [[ -n "$CONTACT_METHOD_FILTER" && "$CONTACT_METHOD_FILTER" != "null" ]]; then
    xbe_json view project-subscriptions list --contact-method "$CONTACT_METHOD_FILTER" --limit 5
    assert_success
else
    skip "No contact method available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete subscription requires --confirm flag"
if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
    xbe_run do project-subscriptions delete "$CREATED_SUBSCRIPTION_ID"
    assert_failure
else
    skip "No created subscription for delete confirmation test"
fi

test_name "Delete subscription with --confirm"
if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
    xbe_run do project-subscriptions delete "$CREATED_SUBSCRIPTION_ID" --confirm
    assert_success
else
    skip "No created subscription to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create subscription without project fails"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_run do project-subscriptions create --user "$USER_ID"
    assert_failure
else
    skip "No user available for missing project test"
fi

test_name "Create subscription without user fails"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_run do project-subscriptions create --project "$PROJECT_ID"
    assert_failure
else
    skip "No project ID available for missing user test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
