#!/bin/bash
#
# XBE CLI Integration Tests: Raw Transport Projects
#
# Tests create, list filters, show, and delete operations for raw transport projects.
#
# COVERAGE: Create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_ID=""
CREATED_BROKER_ID=""

EXTERNAL_PROJECT_NUMBER=""

ROWVERSION_MIN=""
ROWVERSION_MAX=""


describe "Resource: raw-transport-projects"

# ==========================================================================
# Prerequisites - Create broker
# ==========================================================================

test_name "Create prerequisite broker for raw transport projects"
BROKER_NAME=$(unique_name "RawTransportProjectBroker")

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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create raw transport project with required fields"
EXTERNAL_PROJECT_NUMBER=$(unique_name "PROJ")
ROWVERSION_MIN="100"
ROWVERSION_MAX="200"

xbe_json do raw-transport-projects create \
    --broker "$CREATED_BROKER_ID" \
    --external-project-number "$EXTERNAL_PROJECT_NUMBER" \
    --importer "quantix_tmw" \
    --tables-rowversion-min "$ROWVERSION_MIN" \
    --tables-rowversion-max "$ROWVERSION_MAX" \
    --is-managed \
    --tables '[]'

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "raw-transport-projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created raw transport project but no ID returned"
    fi
else
    fail "Failed to create raw transport project"
fi

# Only continue if we successfully created a project
if [[ -z "$CREATED_PROJECT_ID" || "$CREATED_PROJECT_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw transport project ID"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show raw transport project"
xbe_json view raw-transport-projects show "$CREATED_PROJECT_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List raw transport projects"
xbe_json view raw-transport-projects list --limit 5
assert_success


test_name "List raw transport projects returns array"
xbe_json view raw-transport-projects list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list raw transport projects"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List raw transport projects with --broker filter"
xbe_json view raw-transport-projects list --broker "$CREATED_BROKER_ID" --limit 5
assert_success


test_name "List raw transport projects with --tables-rowversion-min filter"
xbe_json view raw-transport-projects list --tables-rowversion-min "$ROWVERSION_MIN" --limit 5
assert_success


test_name "List raw transport projects with --tables-rowversion-max filter"
xbe_json view raw-transport-projects list --tables-rowversion-max "$ROWVERSION_MAX" --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete raw transport project requires --confirm flag"
xbe_run do raw-transport-projects delete "$CREATED_PROJECT_ID"
assert_failure


test_name "Delete raw transport project with --confirm"
xbe_run do raw-transport-projects delete "$CREATED_PROJECT_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
