#!/bin/bash
#
# XBE CLI Integration Tests: Application Settings
#
# Tests CRUD operations for the application-settings resource.
# Application settings are global key/value pairs and require admin access.
#
# COVERAGE: Create, update, list pagination, show, delete failure
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SETTING_ID=""
CREATED_SETTING_KEY=""

describe "Resource: application-settings"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create application setting with required fields"
CREATED_SETTING_KEY="CLI_TEST_SETTING_$(date +%s)_${RANDOM}"
CREATED_SETTING_VALUE="test-value-$(date +%s)"

xbe_json do application-settings create \
    --key "$CREATED_SETTING_KEY" \
    --value "$CREATED_SETTING_VALUE" \
    --description "CLI test setting"

if [[ $status -eq 0 ]]; then
    CREATED_SETTING_ID=$(json_get ".id")
    if [[ -n "$CREATED_SETTING_ID" && "$CREATED_SETTING_ID" != "null" ]]; then
        pass
    else
        fail "Created application setting but no ID returned"
    fi
else
    fail "Failed to create application setting"
fi

# Only continue if we successfully created a setting
if [[ -z "$CREATED_SETTING_ID" || "$CREATED_SETTING_ID" == "null" ]]; then
    echo "Cannot continue without a valid application setting ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show application setting"
xbe_json view application-settings show "$CREATED_SETTING_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update application setting value"
xbe_json do application-settings update "$CREATED_SETTING_ID" --value "updated-value"
assert_success

test_name "Update application setting description"
xbe_json do application-settings update "$CREATED_SETTING_ID" --description "Updated description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List application settings"
xbe_json view application-settings list --limit 5
assert_success

test_name "List application settings returns array"
xbe_json view application-settings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list application settings"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List application settings with --limit"
xbe_json view application-settings list --limit 3
assert_success

test_name "List application settings with --offset"
xbe_json view application-settings list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests - Expected failure
# ============================================================================

test_name "Delete application setting requires --confirm flag"
xbe_run do application-settings delete "$CREATED_SETTING_ID"
assert_failure

test_name "Delete application setting fails (server prohibits deletion)"
xbe_run do application-settings delete "$CREATED_SETTING_ID" --confirm
assert_failure

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create application setting without key fails"
xbe_json do application-settings create --value "missing-key"
assert_failure

test_name "Create application setting without value fails"
xbe_json do application-settings create --key "MISSING_VALUE"
assert_failure

test_name "Update without any fields fails"
xbe_json do application-settings update "$CREATED_SETTING_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
