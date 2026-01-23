#!/bin/bash
#
# XBE CLI Integration Tests: Maintenance Requirement Parts
#
# Tests CRUD operations for the maintenance_requirement_parts resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PART_ID=""
CREATED_BROKER_ID=""
UPDATED_BROKER_ID=""
CREATED_EQUIP_CLASS_ID=""
UPDATED_EQUIP_CLASS_ID=""

describe "Resource: maintenance_requirement_parts"

# ==========================================================================
# Prerequisites - Create broker and equipment classifications
# ==========================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "MRPBroker")

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

test_name "Create prerequisite equipment classification"
EQUIP_CLASS_NAME=$(unique_name "MRPClass")

xbe_json do equipment-classifications create --name "$EQUIP_CLASS_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIP_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIP_CLASS_ID" && "$CREATED_EQUIP_CLASS_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIP_CLASS_ID"
        pass
    else
        fail "Created equipment classification but no ID returned"
        echo "Cannot continue without an equipment classification"
        run_tests
    fi
else
    fail "Failed to create equipment classification"
    echo "Cannot continue without an equipment classification"
    run_tests
fi

test_name "Create secondary broker for updates"
UPDATED_BROKER_NAME=$(unique_name "MRPBrokerUpdate")

xbe_json do brokers create --name "$UPDATED_BROKER_NAME"

if [[ $status -eq 0 ]]; then
    UPDATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$UPDATED_BROKER_ID" && "$UPDATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$UPDATED_BROKER_ID"
        pass
    else
        fail "Created update broker but no ID returned"
        echo "Cannot continue without a broker for updates"
        run_tests
    fi
else
    fail "Failed to create update broker"
    echo "Cannot continue without a broker for updates"
    run_tests
fi

test_name "Create secondary equipment classification for updates"
UPDATED_EQUIP_CLASS_NAME=$(unique_name "MRPClassUpdate")

xbe_json do equipment-classifications create --name "$UPDATED_EQUIP_CLASS_NAME"

if [[ $status -eq 0 ]]; then
    UPDATED_EQUIP_CLASS_ID=$(json_get ".id")
    if [[ -n "$UPDATED_EQUIP_CLASS_ID" && "$UPDATED_EQUIP_CLASS_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$UPDATED_EQUIP_CLASS_ID"
        pass
    else
        fail "Created update equipment classification but no ID returned"
        echo "Cannot continue without an equipment classification for updates"
        run_tests
    fi
else
    fail "Failed to create update equipment classification"
    echo "Cannot continue without an equipment classification for updates"
    run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create maintenance requirement part with full details"
PART_NAME=$(unique_name "MRPPart")
PART_NUMBER="MRP-$(unique_suffix)"

xbe_json do maintenance-requirement-parts create \
    --name "$PART_NAME" \
    --part-number "$PART_NUMBER" \
    --description "Template part for testing" \
    --notes "Initial notes" \
    --is-template true \
    --make "Acme" \
    --model "X1" \
    --year "2024" \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIP_CLASS_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PART_ID=$(json_get ".id")
    if [[ -n "$CREATED_PART_ID" && "$CREATED_PART_ID" != "null" ]]; then
        register_cleanup "maintenance-requirement-parts" "$CREATED_PART_ID"
        pass
    else
        fail "Created maintenance requirement part but no ID returned"
    fi
else
    fail "Failed to create maintenance requirement part"
fi

if [[ -z "$CREATED_PART_ID" || "$CREATED_PART_ID" == "null" ]]; then
    echo "Cannot continue without a maintenance requirement part"
    run_tests
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update maintenance requirement part name"
NEW_PART_NAME=$(unique_name "MRPPartUpdated")

xbe_json do maintenance-requirement-parts update "$CREATED_PART_ID" --name "$NEW_PART_NAME"
assert_success

test_name "Update maintenance requirement part part-number"
NEW_PART_NUMBER="MRP-UPD-$(unique_suffix)"

xbe_json do maintenance-requirement-parts update "$CREATED_PART_ID" --part-number "$NEW_PART_NUMBER"
assert_success

test_name "Update maintenance requirement part description"

xbe_json do maintenance-requirement-parts update "$CREATED_PART_ID" --description "Updated description"
assert_success

test_name "Update maintenance requirement part notes"

xbe_json do maintenance-requirement-parts update "$CREATED_PART_ID" --notes "Updated notes"
assert_success

test_name "Update maintenance requirement part make/model/year/is-template"

xbe_json do maintenance-requirement-parts update "$CREATED_PART_ID" \
    --make "Acme-Updated" \
    --model "X2" \
    --year "2025" \
    --is-template true
assert_success

test_name "Update maintenance requirement part relationships"

xbe_json do maintenance-requirement-parts update "$CREATED_PART_ID" \
    --broker "$UPDATED_BROKER_ID" \
    --equipment-classification "$UPDATED_EQUIP_CLASS_ID"
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List maintenance requirement parts with --broker filter"
xbe_json view maintenance-requirement-parts list --broker "$UPDATED_BROKER_ID" --limit 10
assert_success

test_name "List maintenance requirement parts with --equipment-classification filter"
xbe_json view maintenance-requirement-parts list --equipment-classification "$UPDATED_EQUIP_CLASS_ID" --limit 10
assert_success

test_name "List maintenance requirement parts with --make filter"
xbe_json view maintenance-requirement-parts list --make "Acme-Updated" --limit 10
assert_success

test_name "List maintenance requirement parts with --model filter"
xbe_json view maintenance-requirement-parts list --model "X2" --limit 10
assert_success

test_name "List maintenance requirement parts with --year filter"
xbe_json view maintenance-requirement-parts list --year "2025" --limit 10
assert_success

test_name "List maintenance requirement parts with --is-template filter"
xbe_json view maintenance-requirement-parts list --is-template true --limit 10
assert_success

test_name "List maintenance requirement parts with --maintenance-requirements filter"
MAINT_REQUIREMENT_ID="${XBE_TEST_MAINTENANCE_REQUIREMENT_ID:-1}"
xbe_json view maintenance-requirement-parts list --maintenance-requirements "$MAINT_REQUIREMENT_ID" --limit 10
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List maintenance requirement parts with --limit"
xbe_json view maintenance-requirement-parts list --limit 3
assert_success

test_name "List maintenance requirement parts with --offset"
xbe_json view maintenance-requirement-parts list --limit 3 --offset 1
assert_success

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create maintenance requirement part without name fails"
xbe_json do maintenance-requirement-parts create --is-template true --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update without any fields fails"
# Use a placeholder ID - it should fail for missing fields before hitting the API
xbe_json do maintenance-requirement-parts update "99999999"
assert_failure

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete maintenance requirement part"
xbe_run do maintenance-requirement-parts delete "$CREATED_PART_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
