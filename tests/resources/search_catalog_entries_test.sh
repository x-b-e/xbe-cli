#!/bin/bash
#
# XBE CLI Integration Tests: Search Catalog Entries
#
# Tests view operations for the search_catalog_entries resource.
# Search catalog entries index entities for full-text and fuzzy search.
#
# COVERAGE: list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: search_catalog_entries"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List search catalog entries"
xbe_json view search-catalog-entries list --limit 5
assert_success

test_name "List search catalog entries returns array"
xbe_json view search-catalog-entries list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list search catalog entries"
fi

ENTRY_ID=$(json_get ".[0].id")
ENTITY_TYPE_FILTER=$(json_get ".[0].entity_type")
ENTITY_ID_FILTER=$(json_get ".[0].entity_id")
BROKER_FILTER=$(json_get ".[0].broker_id")
CUSTOMER_FILTER=$(json_get ".[0].customer_id")
TRUCKER_FILTER=$(json_get ".[0].trucker_id")
SEARCH_FILTER=$(json_get ".[0].display_text")

if [[ -z "$ENTITY_TYPE_FILTER" || "$ENTITY_TYPE_FILTER" == "null" ]]; then
    ENTITY_TYPE_FILTER="Customer"
fi

if [[ -z "$ENTITY_ID_FILTER" || "$ENTITY_ID_FILTER" == "null" ]]; then
    ENTITY_ID_FILTER="1"
fi

if [[ -z "$BROKER_FILTER" || "$BROKER_FILTER" == "null" ]]; then
    BROKER_FILTER="1"
fi

if [[ -z "$CUSTOMER_FILTER" || "$CUSTOMER_FILTER" == "null" ]]; then
    CUSTOMER_FILTER="1"
fi

if [[ -z "$TRUCKER_FILTER" || "$TRUCKER_FILTER" == "null" ]]; then
    TRUCKER_FILTER="1"
fi

if [[ -z "$SEARCH_FILTER" || "$SEARCH_FILTER" == "null" ]]; then
    SEARCH_FILTER="john"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List search catalog entries with --entity-type filter"
xbe_json view search-catalog-entries list --entity-type "$ENTITY_TYPE_FILTER" --limit 5
assert_success

test_name "List search catalog entries with --entity-id filter"
xbe_json view search-catalog-entries list --entity-id "$ENTITY_ID_FILTER" --limit 5
assert_success

test_name "List search catalog entries with --broker filter"
xbe_json view search-catalog-entries list --broker "$BROKER_FILTER" --limit 5
assert_success

test_name "List search catalog entries with --customer filter"
xbe_json view search-catalog-entries list --customer "$CUSTOMER_FILTER" --limit 5
assert_success

test_name "List search catalog entries with --trucker filter"
xbe_json view search-catalog-entries list --trucker "$TRUCKER_FILTER" --limit 5
assert_success

test_name "List search catalog entries with --search filter"
xbe_json view search-catalog-entries list --search "$SEARCH_FILTER" --limit 5
assert_success

test_name "List search catalog entries with --fuzzy-search filter"
xbe_json view search-catalog-entries list --fuzzy-search "jo" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show search catalog entry"
if [[ -n "$ENTRY_ID" && "$ENTRY_ID" != "null" ]]; then
    xbe_json view search-catalog-entries show "$ENTRY_ID"
    assert_success
else
    skip "No search catalog entry available for show test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
