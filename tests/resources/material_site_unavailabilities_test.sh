#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Unavailabilities
#
# Tests CRUD operations for the material-site-unavailabilities resource.
# Material site unavailabilities track downtime windows for material sites.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_UNAVAILABILITY_ID=""
SAMPLE_UNAVAILABILITY_ID=""
SAMPLE_MATERIAL_SITE_ID=""
SAMPLE_START_AT=""
SAMPLE_END_AT=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""
SKIP_CREATE=0

describe "Resource: material-site-unavailabilities"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material site unavailabilities"
xbe_json view material-site-unavailabilities list --limit 5
assert_success

test_name "List material site unavailabilities returns array"
xbe_json view material-site-unavailabilities list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material site unavailabilities"
fi

# ============================================================================
# Prerequisites - Create material site and locate sample unavailability
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "MSU-Broker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        SKIP_CREATE=1
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        SKIP_CREATE=1
    fi
fi

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "MSU-Supplier")

if [[ $SKIP_CREATE -eq 0 ]]; then
    xbe_json do material-suppliers create --name "$SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
        if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
            register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
            pass
        else
            fail "Created material supplier but no ID returned"
            SKIP_CREATE=1
        fi
    else
        fail "Failed to create material supplier"
        SKIP_CREATE=1
    fi
else
    skip "Prerequisite broker missing"
fi

test_name "Create prerequisite material site"
SITE_NAME=$(unique_name "MSU-Site")

if [[ $SKIP_CREATE -eq 0 ]]; then
    xbe_json do material-sites create \
        --name "$SITE_NAME" \
        --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
        --address "100 Unavailability Rd, Chicago, IL 60601"

    if [[ $status -eq 0 ]]; then
        CREATED_MATERIAL_SITE_ID=$(json_get ".id")
        if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
            register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
            pass
        else
            fail "Created material site but no ID returned"
            SKIP_CREATE=1
        fi
    else
        fail "Failed to create material site"
        SKIP_CREATE=1
    fi
else
    skip "Prerequisite supplier missing"
fi

# Locate sample unavailability for filter tests

test_name "Locate material site unavailability for filters"
xbe_json view material-site-unavailabilities list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_UNAVAILABILITY_ID=$(json_get ".[0].id")
        SAMPLE_MATERIAL_SITE_ID=$(json_get ".[0].material_site_id")
        SAMPLE_START_AT=$(json_get ".[0].start_at")
        SAMPLE_END_AT=$(json_get ".[0].end_at")
        pass
    else
        if [[ -n "$XBE_TEST_MATERIAL_SITE_UNAVAILABILITY_ID" ]]; then
            xbe_json view material-site-unavailabilities show "$XBE_TEST_MATERIAL_SITE_UNAVAILABILITY_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_UNAVAILABILITY_ID=$(json_get ".id")
                SAMPLE_MATERIAL_SITE_ID=$(json_get ".material_site_id")
                SAMPLE_START_AT=$(json_get ".start_at")
                SAMPLE_END_AT=$(json_get ".end_at")
                pass
            else
                skip "Failed to load XBE_TEST_MATERIAL_SITE_UNAVAILABILITY_ID"
            fi
        else
            skip "No material site unavailabilities found. Set XBE_TEST_MATERIAL_SITE_UNAVAILABILITY_ID for filter tests."
        fi
    fi
else
    fail "Failed to list material site unavailabilities for prerequisites"
fi

# Capture created/updated timestamps for filters
if [[ -n "$SAMPLE_UNAVAILABILITY_ID" ]]; then
    test_name "Fetch material site unavailability timestamps"
    xbe_json view material-site-unavailabilities show "$SAMPLE_UNAVAILABILITY_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_CREATED_AT=$(json_get ".created_at")
        SAMPLE_UPDATED_AT=$(json_get ".updated_at")
        pass
    else
        skip "Unable to load timestamps for filter tests"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ $SKIP_CREATE -eq 0 ]]; then
    test_name "Create material site unavailability"
    START_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    END_AT="$(date -u -d "+2 hours" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+2H +%Y-%m-%dT%H:%M:%SZ)"

    xbe_json do material-site-unavailabilities create \
        --material-site "$CREATED_MATERIAL_SITE_ID" \
        --start-at "$START_AT" \
        --end-at "$END_AT" \
        --description "CLI test unavailability"

    if [[ $status -eq 0 ]]; then
        CREATED_UNAVAILABILITY_ID=$(json_get ".id")
        if [[ -n "$CREATED_UNAVAILABILITY_ID" && "$CREATED_UNAVAILABILITY_ID" != "null" ]]; then
            register_cleanup "material-site-unavailabilities" "$CREATED_UNAVAILABILITY_ID"
            pass
        else
            fail "Created unavailability but no ID returned"
        fi
    else
        fail "Failed to create material site unavailability"
    fi
else
    test_name "Create material site unavailability"
    skip "Prerequisites not available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_UNAVAILABILITY_ID" && "$CREATED_UNAVAILABILITY_ID" != "null" ]]; then
    test_name "Update material site unavailability description"
    xbe_json do material-site-unavailabilities update "$CREATED_UNAVAILABILITY_ID" --description "Updated description"
    assert_success

    test_name "Update material site unavailability end-at"
    UPDATED_END_AT="$(date -u -d "+3 hours" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+3H +%Y-%m-%dT%H:%M:%SZ)"
    xbe_json do material-site-unavailabilities update "$CREATED_UNAVAILABILITY_ID" --end-at "$UPDATED_END_AT"
    assert_success

    test_name "Update material site unavailability without fields fails"
    xbe_json do material-site-unavailabilities update "$CREATED_UNAVAILABILITY_ID"
    assert_failure
else
    test_name "Update material site unavailability description"
    skip "No unavailability created"
    test_name "Update material site unavailability end-at"
    skip "No unavailability created"
    test_name "Update material site unavailability without fields fails"
    skip "No unavailability created"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_UNAVAILABILITY_ID" && "$CREATED_UNAVAILABILITY_ID" != "null" ]]; then
    test_name "Show material site unavailability"
    xbe_json view material-site-unavailabilities show "$CREATED_UNAVAILABILITY_ID"
    assert_success
else
    test_name "Show material site unavailability"
    skip "No unavailability created"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

if [[ -n "$SAMPLE_MATERIAL_SITE_ID" ]]; then
    test_name "Filter by material site"
    xbe_json view material-site-unavailabilities list --material-site "$SAMPLE_MATERIAL_SITE_ID" --limit 5
    assert_success
else
    test_name "Filter by material site"
    skip "No sample material site ID"
fi

if [[ -n "$SAMPLE_START_AT" ]]; then
    test_name "Filter by start-at min"
    xbe_json view material-site-unavailabilities list --start-at-min "$SAMPLE_START_AT" --limit 5
    assert_success

    test_name "Filter by start-at max"
    xbe_json view material-site-unavailabilities list --start-at-max "$SAMPLE_START_AT" --limit 5
    assert_success

    test_name "Filter by is-start-at"
    xbe_json view material-site-unavailabilities list --is-start-at true --limit 5
    assert_success
else
    test_name "Filter by start-at min"
    skip "No sample start-at"
    test_name "Filter by start-at max"
    skip "No sample start-at"
    test_name "Filter by is-start-at"
    skip "No sample start-at"
fi

if [[ -n "$SAMPLE_END_AT" ]]; then
    test_name "Filter by end-at min"
    xbe_json view material-site-unavailabilities list --end-at-min "$SAMPLE_END_AT" --limit 5
    assert_success

    test_name "Filter by end-at max"
    xbe_json view material-site-unavailabilities list --end-at-max "$SAMPLE_END_AT" --limit 5
    assert_success

    test_name "Filter by is-end-at"
    xbe_json view material-site-unavailabilities list --is-end-at true --limit 5
    assert_success
else
    test_name "Filter by end-at min"
    skip "No sample end-at"
    test_name "Filter by end-at max"
    skip "No sample end-at"
    test_name "Filter by is-end-at"
    skip "No sample end-at"
fi

if [[ -n "$SAMPLE_CREATED_AT" ]]; then
    test_name "Filter by created-at min"
    xbe_json view material-site-unavailabilities list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success

    test_name "Filter by created-at max"
    xbe_json view material-site-unavailabilities list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success

    test_name "Filter by is-created-at"
    xbe_json view material-site-unavailabilities list --is-created-at true --limit 5
    assert_success
else
    test_name "Filter by created-at min"
    skip "No sample created-at"
    test_name "Filter by created-at max"
    skip "No sample created-at"
    test_name "Filter by is-created-at"
    skip "No sample created-at"
fi

if [[ -n "$SAMPLE_UPDATED_AT" ]]; then
    test_name "Filter by updated-at min"
    xbe_json view material-site-unavailabilities list --updated-at-min "$SAMPLE_UPDATED_AT" --limit 5
    assert_success

    test_name "Filter by updated-at max"
    xbe_json view material-site-unavailabilities list --updated-at-max "$SAMPLE_UPDATED_AT" --limit 5
    assert_success

    test_name "Filter by is-updated-at"
    xbe_json view material-site-unavailabilities list --is-updated-at true --limit 5
    assert_success
else
    test_name "Filter by updated-at min"
    skip "No sample updated-at"
    test_name "Filter by updated-at max"
    skip "No sample updated-at"
    test_name "Filter by is-updated-at"
    skip "No sample updated-at"
fi

# ============================================================================
# LIST Tests - Pagination / Sorting
# ============================================================================

test_name "List material site unavailabilities with --limit"
xbe_json view material-site-unavailabilities list --limit 3
assert_success

test_name "List material site unavailabilities with --offset"
xbe_json view material-site-unavailabilities list --limit 3 --offset 1
assert_success

test_name "List material site unavailabilities with --sort"
xbe_json view material-site-unavailabilities list --sort start-at --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_UNAVAILABILITY_ID" && "$CREATED_UNAVAILABILITY_ID" != "null" ]]; then
    test_name "Delete material site unavailability requires --confirm flag"
    xbe_run do material-site-unavailabilities delete "$CREATED_UNAVAILABILITY_ID"
    assert_failure

    test_name "Delete material site unavailability with --confirm"
    xbe_run do material-site-unavailabilities delete "$CREATED_UNAVAILABILITY_ID" --confirm
    assert_success
else
    test_name "Delete material site unavailability requires --confirm flag"
    skip "No unavailability created"
    test_name "Delete material site unavailability with --confirm"
    skip "No unavailability created"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create unavailability without material site fails"
xbe_json do material-site-unavailabilities create --start-at "2026-01-23T08:00:00Z"
assert_failure

test_name "Create unavailability without start/end fails"
xbe_json do material-site-unavailabilities create --material-site 123
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
