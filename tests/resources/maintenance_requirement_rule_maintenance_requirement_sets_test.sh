#!/bin/bash
#
# XBE CLI Integration Tests: Maintenance Requirement Rule Maintenance Requirement Sets
#
# Tests CRUD operations for the maintenance_requirement_rule_maintenance_requirement_sets resource.
#
# COVERAGE: All filters + all create attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_LINK_ID=""
SAMPLE_RULE_ID=""
SAMPLE_SET_ID=""
CREATED_LINK_ID=""
CREATE_RULE_ID=""
CREATE_SET_ID=""
FILTER_RULE_ID=""
FILTER_SET_ID=""
FILTER_TARGET_ID=""

describe "Resource: maintenance-requirement-rule-maintenance-requirement-sets"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List maintenance requirement rule maintenance requirement sets"
xbe_json view maintenance-requirement-rule-maintenance-requirement-sets list --limit 5
assert_success

test_name "List maintenance requirement rule maintenance requirement sets returns array"
xbe_json view maintenance-requirement-rule-maintenance-requirement-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list maintenance requirement rule maintenance requirement sets"
fi

# ============================================================================
# Prerequisites - Locate sample link
# ============================================================================

test_name "Locate maintenance requirement rule maintenance requirement set for filters"
xbe_json view maintenance-requirement-rule-maintenance-requirement-sets list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_LINK_ID=$(json_get ".[0].id")
        SAMPLE_RULE_ID=$(json_get ".[0].maintenance_requirement_rule_id")
        SAMPLE_SET_ID=$(json_get ".[0].maintenance_requirement_set_id")
        pass
    else
        if [[ -n "$XBE_TEST_MAINTENANCE_REQUIREMENT_RULE_MAINTENANCE_REQUIREMENT_SET_ID" ]]; then
            xbe_json view maintenance-requirement-rule-maintenance-requirement-sets show "$XBE_TEST_MAINTENANCE_REQUIREMENT_RULE_MAINTENANCE_REQUIREMENT_SET_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_LINK_ID=$(json_get ".id")
                SAMPLE_RULE_ID=$(json_get ".maintenance_requirement_rule_id")
                SAMPLE_SET_ID=$(json_get ".maintenance_requirement_set_id")
                pass
            else
                skip "Failed to load XBE_TEST_MAINTENANCE_REQUIREMENT_RULE_MAINTENANCE_REQUIREMENT_SET_ID"
            fi
        else
            skip "No maintenance requirement rule maintenance requirement sets found. Set XBE_TEST_MAINTENANCE_REQUIREMENT_RULE_MAINTENANCE_REQUIREMENT_SET_ID to enable show/filter tests."
        fi
    fi
else
    fail "Failed to list maintenance requirement rule maintenance requirement sets for prerequisites"
fi

# ============================================================================
# CREATE Tests - Best Effort
# ============================================================================

CREATE_RULE_ID="$XBE_TEST_MAINTENANCE_REQUIREMENT_RULE_ID"
CREATE_SET_ID="$XBE_TEST_MAINTENANCE_REQUIREMENT_SET_ID"

if [[ -n "$CREATE_RULE_ID" && -n "$CREATE_SET_ID" ]]; then
    test_name "Create maintenance requirement rule maintenance requirement set"
    xbe_json do maintenance-requirement-rule-maintenance-requirement-sets create \
        --maintenance-requirement-rule "$CREATE_RULE_ID" \
        --maintenance-requirement-set "$CREATE_SET_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "maintenance-requirement-rule-maintenance-requirement-sets" "$CREATED_LINK_ID"
            pass
        else
            fail "Created link but no ID returned"
        fi
    else
        fail "Failed to create maintenance requirement rule maintenance requirement set"
    fi
else
    test_name "Create maintenance requirement rule maintenance requirement set"
    skip "Set XBE_TEST_MAINTENANCE_REQUIREMENT_RULE_ID and XBE_TEST_MAINTENANCE_REQUIREMENT_SET_ID (template) to enable create/delete tests."
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Show maintenance requirement rule maintenance requirement set (created)"
    xbe_json view maintenance-requirement-rule-maintenance-requirement-sets show "$CREATED_LINK_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show created maintenance requirement rule maintenance requirement set"
    fi
elif [[ -n "$SAMPLE_LINK_ID" && "$SAMPLE_LINK_ID" != "null" ]]; then
    test_name "Show maintenance requirement rule maintenance requirement set"
    xbe_json view maintenance-requirement-rule-maintenance-requirement-sets show "$SAMPLE_LINK_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show maintenance requirement rule maintenance requirement set"
    fi
else
    test_name "Show maintenance requirement rule maintenance requirement set"
    skip "No maintenance requirement rule maintenance requirement set available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    FILTER_TARGET_ID="$CREATED_LINK_ID"
    FILTER_RULE_ID="$CREATE_RULE_ID"
    FILTER_SET_ID="$CREATE_SET_ID"
else
    FILTER_TARGET_ID="$SAMPLE_LINK_ID"
    FILTER_RULE_ID="$SAMPLE_RULE_ID"
    FILTER_SET_ID="$SAMPLE_SET_ID"
fi

test_name "Filter by maintenance requirement rule"
if [[ -n "$FILTER_RULE_ID" && "$FILTER_RULE_ID" != "null" ]]; then
    xbe_json view maintenance-requirement-rule-maintenance-requirement-sets list --maintenance-requirement-rule "$FILTER_RULE_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_TARGET_ID" && "$FILTER_TARGET_ID" != "null" ]] && echo "$output" | jq -e --arg id "$FILTER_TARGET_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing expected link"
        fi
    else
        fail "Failed to filter by maintenance requirement rule"
    fi
else
    skip "No maintenance requirement rule ID available for filter test"
fi

test_name "Filter by maintenance requirement set"
if [[ -n "$FILTER_SET_ID" && "$FILTER_SET_ID" != "null" ]]; then
    xbe_json view maintenance-requirement-rule-maintenance-requirement-sets list --maintenance-requirement-set "$FILTER_SET_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_TARGET_ID" && "$FILTER_TARGET_ID" != "null" ]] && echo "$output" | jq -e --arg id "$FILTER_TARGET_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing expected link"
        fi
    else
        fail "Failed to filter by maintenance requirement set"
    fi
else
    skip "No maintenance requirement set ID available for filter test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create link without required flags fails"
xbe_json do maintenance-requirement-rule-maintenance-requirement-sets create
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Delete maintenance requirement rule maintenance requirement set"
    xbe_run do maintenance-requirement-rule-maintenance-requirement-sets delete "$CREATED_LINK_ID" --confirm
    assert_success
else
    test_name "Delete maintenance requirement rule maintenance requirement set"
    skip "No created link available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
