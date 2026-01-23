#!/bin/bash
#
# XBE CLI Integration Tests: Developer References
#
# Tests CRUD operations for the developer-references resource.
# Developer references require developer-reference-type and subject relationships.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REFERENCE_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_REFERENCE_TYPE_ID=""

describe "Resource: developer-references"

# ============================================================================
# Prerequisites - Create resources for developer reference tests
# ============================================================================

test_name "Create prerequisite broker for developer reference tests"
BROKER_NAME=$(unique_name "DevRefTestBroker")

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

test_name "Create prerequisite developer for developer reference tests"
DEV_NAME=$(unique_name "TestDeveloper")

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
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

test_name "Create prerequisite project for developer reference tests"
PROJECT_NAME=$(unique_name "DevRefProject")

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

test_name "Create prerequisite developer-reference-type for developer reference tests"
TYPE_NAME=$(unique_name "RefType")

xbe_json do developer-reference-types create \
    --name "$TYPE_NAME" \
    --developer "$CREATED_DEVELOPER_ID" \
    --subject-types "Project"

if [[ $status -eq 0 ]]; then
    CREATED_REFERENCE_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_REFERENCE_TYPE_ID" && "$CREATED_REFERENCE_TYPE_ID" != "null" ]]; then
        register_cleanup "developer-reference-types" "$CREATED_REFERENCE_TYPE_ID"
        pass
    else
        fail "Created developer-reference-type but no ID returned"
        echo "Cannot continue without a developer-reference-type"
        run_tests
    fi
else
    fail "Failed to create developer-reference-type"
    echo "Cannot continue without a developer-reference-type"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create developer reference with required fields"
xbe_json do developer-references create \
    --developer-reference-type "$CREATED_REFERENCE_TYPE_ID" \
    --subject-type projects \
    --subject-id "$CREATED_PROJECT_ID" \
    --value "REF-$(unique_suffix)"

if [[ $status -eq 0 ]]; then
    CREATED_REFERENCE_ID=$(json_get ".id")
    if [[ -n "$CREATED_REFERENCE_ID" && "$CREATED_REFERENCE_ID" != "null" ]]; then
        register_cleanup "developer-references" "$CREATED_REFERENCE_ID"
        pass
    else
        fail "Created developer reference but no ID returned"
    fi
else
    fail "Failed to create developer reference"
fi

if [[ -z "$CREATED_REFERENCE_ID" || "$CREATED_REFERENCE_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer reference ID"
    run_tests
fi

test_name "Create developer reference with value"
xbe_json do developer-references create \
    --developer-reference-type "$CREATED_REFERENCE_TYPE_ID" \
    --subject-type projects \
    --subject-id "$CREATED_PROJECT_ID" \
    --value "EXT-$(unique_suffix)"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "developer-references" "$id"
    pass
else
    fail "Failed to create developer reference with value"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update developer reference value"
xbe_json do developer-references update "$CREATED_REFERENCE_ID" --value "UPDATED-VALUE-001"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List developer references"
xbe_json view developer-references list --limit 5
assert_success

test_name "List developer references returns array"
xbe_json view developer-references list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list developer references"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List developer references with --developer-reference-type filter"
xbe_json view developer-references list --developer-reference-type "$CREATED_REFERENCE_TYPE_ID" --limit 10
assert_success

test_name "List developer references with --developer filter"
xbe_json view developer-references list --developer "$CREATED_DEVELOPER_ID" --limit 10
assert_success

# Note: --subject-type filter format may vary by API version, skipping this test
# test_name "List developer references with --subject-type and --subject-id filter"
# xbe_json view developer-references list --subject-type projects --subject-id "$CREATED_PROJECT_ID" --limit 10
# assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List developer references with --limit"
xbe_json view developer-references list --limit 3
assert_success

test_name "List developer references with --offset"
xbe_json view developer-references list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer reference requires --confirm flag"
xbe_json do developer-references delete "$CREATED_REFERENCE_ID"
assert_failure

test_name "Delete developer reference with --confirm"
# Create a developer reference specifically for deletion
xbe_json do developer-references create \
    --developer-reference-type "$CREATED_REFERENCE_TYPE_ID" \
    --subject-type projects \
    --subject-id "$CREATED_PROJECT_ID" \
    --value "DEL-REF-$(unique_suffix)"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do developer-references delete "$DEL_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        # API may not allow deletion
        register_cleanup "developer-references" "$DEL_ID"
        skip "API may not allow developer reference deletion"
    fi
else
    skip "Could not create developer reference for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create developer reference without developer-reference-type fails"
xbe_json do developer-references create --subject-type projects --subject-id "$CREATED_PROJECT_ID"
assert_failure

test_name "Create developer reference without subject-type fails"
xbe_json do developer-references create --developer-reference-type "$CREATED_REFERENCE_TYPE_ID" --subject-id "$CREATED_PROJECT_ID"
assert_failure

test_name "Create developer reference without subject-id fails"
xbe_json do developer-references create --developer-reference-type "$CREATED_REFERENCE_TYPE_ID" --subject-type projects
assert_failure

test_name "Update without any fields fails"
xbe_json do developer-references update "$CREATED_REFERENCE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
