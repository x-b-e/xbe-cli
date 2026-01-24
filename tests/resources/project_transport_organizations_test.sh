#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Organizations
#
# Tests CRUD operations for the project_transport_organizations resource.
#
# COVERAGE: All writable attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_PTO_ID=""
CREATED_PTO2_ID=""
EXTERNAL_ID_CREATE=""
EXTERNAL_ID_UPDATE=""

describe "Resource: project-transport-organizations"

# =========================================================================
# Prerequisites - Create broker for project transport organization tests
# =========================================================================

test_name "Create prerequisite broker for project transport organization tests"
BROKER_NAME=$(unique_name "PTOTestBroker")

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

# =========================================================================
# CREATE Tests
# =========================================================================

test_name "Create project transport organization with required fields"
TEST_NAME=$(unique_name "ProjTransportOrg")

xbe_json do project-transport-organizations create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PTO_ID=$(json_get ".id")
    if [[ -n "$CREATED_PTO_ID" && "$CREATED_PTO_ID" != "null" ]]; then
        register_cleanup "project-transport-organizations" "$CREATED_PTO_ID"
        pass
    else
        fail "Created project transport organization but no ID returned"
    fi
else
    fail "Failed to create project transport organization"
fi

if [[ -z "$CREATED_PTO_ID" || "$CREATED_PTO_ID" == "null" ]]; then
    echo "Cannot continue without a valid project transport organization ID"
    run_tests
fi

test_name "Create project transport organization with external TMS master company ID"
TEST_NAME2=$(unique_name "ProjTransportOrg2")
EXTERNAL_ID_CREATE="TMS-$(unique_suffix)"

xbe_json do project-transport-organizations create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --external-tms-master-company-id "$EXTERNAL_ID_CREATE"

if [[ $status -eq 0 ]]; then
    CREATED_PTO2_ID=$(json_get ".id")
    if [[ -n "$CREATED_PTO2_ID" && "$CREATED_PTO2_ID" != "null" ]]; then
        register_cleanup "project-transport-organizations" "$CREATED_PTO2_ID"
        pass
    else
        fail "Created project transport organization but no ID returned"
    fi
else
    fail "Failed to create project transport organization with external ID"
fi

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show project transport organization"
xbe_json view project-transport-organizations show "$CREATED_PTO_ID"
assert_success

# =========================================================================
# UPDATE Tests
# =========================================================================

test_name "Update project transport organization name"
UPDATED_NAME=$(unique_name "UpdatedPTO")
xbe_json do project-transport-organizations update "$CREATED_PTO_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update project transport organization external TMS master company ID"
EXTERNAL_ID_UPDATE="TMS-UPDATE-$(unique_suffix)"
xbe_json do project-transport-organizations update "$CREATED_PTO_ID" --external-tms-master-company-id "$EXTERNAL_ID_UPDATE"
assert_success

# =========================================================================
# LIST Tests - Basic
# =========================================================================

test_name "List project transport organizations"
xbe_json view project-transport-organizations list --limit 5
assert_success

test_name "List project transport organizations returns array"
xbe_json view project-transport-organizations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport organizations"
fi

# =========================================================================
# LIST Tests - Filters
# =========================================================================

test_name "List project transport organizations with --broker filter"
xbe_json view project-transport-organizations list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List project transport organizations with --q filter"
xbe_json view project-transport-organizations list --q "Transport" --limit 5
assert_success

test_name "List project transport organizations with --external-tms-master-company-id filter"
xbe_json view project-transport-organizations list --external-tms-master-company-id "$EXTERNAL_ID_UPDATE" --limit 5
assert_success

test_name "List project transport organizations with --external-identification-value filter"
xbe_json view project-transport-organizations list --external-identification-value "$EXTERNAL_ID_UPDATE" --limit 5
assert_success

# =========================================================================
# DELETE Tests
# =========================================================================

test_name "Delete project transport organization"
if [[ -n "$CREATED_PTO2_ID" ]]; then
    xbe_run do project-transport-organizations delete "$CREATED_PTO2_ID" --confirm
    assert_success
else
    skip "No project transport organization available for delete test"
fi

# =========================================================================
# Error Cases
# =========================================================================

test_name "Create project transport organization without name fails"
xbe_json do project-transport-organizations create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project transport organization without broker fails"
xbe_json do project-transport-organizations create --name "Missing Broker"
assert_failure

test_name "Update project transport organization with no fields fails"
xbe_json do project-transport-organizations update "$CREATED_PTO_ID"
assert_failure

test_name "Delete project transport organization without --confirm fails"
xbe_run do project-transport-organizations delete "$CREATED_PTO_ID"
assert_failure

# =========================================================================
# Summary
# =========================================================================

run_tests
