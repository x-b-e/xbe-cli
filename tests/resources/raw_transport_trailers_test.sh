#!/bin/bash
#
# XBE CLI Integration Tests: Raw Transport Trailers
#
# Tests create, list filters, show, and delete operations for raw transport trailers.
#
# COVERAGE: Create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TRAILER_ID=""
CREATED_BROKER_ID=""
IMPORT_STATUS=""

EXTERNAL_TRAILER_ID=""


describe "Resource: raw-transport-trailers"

# ==========================================================================
# Prerequisites - Create broker
# ==========================================================================

test_name "Create prerequisite broker for raw transport trailers"
BROKER_NAME=$(unique_name "RawTransportTrailerBroker")

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

test_name "Create raw transport trailer with required fields"
EXTERNAL_TRAILER_ID=$(unique_name "TRL")

xbe_json do raw-transport-trailers create \
    --broker "$CREATED_BROKER_ID" \
    --external-trailer-id "$EXTERNAL_TRAILER_ID" \
    --importer "quantix_tmw" \
    --tables '[]'

if [[ $status -eq 0 ]]; then
    CREATED_TRAILER_ID=$(json_get ".id")
    IMPORT_STATUS=$(json_get ".import_status")
    if [[ -n "$CREATED_TRAILER_ID" && "$CREATED_TRAILER_ID" != "null" ]]; then
        register_cleanup "raw-transport-trailers" "$CREATED_TRAILER_ID"
        pass
    else
        fail "Created raw transport trailer but no ID returned"
    fi
else
    fail "Failed to create raw transport trailer"
fi

# Only continue if we successfully created a trailer
if [[ -z "$CREATED_TRAILER_ID" || "$CREATED_TRAILER_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw transport trailer ID"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show raw transport trailer"
xbe_json view raw-transport-trailers show "$CREATED_TRAILER_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List raw transport trailers"
xbe_json view raw-transport-trailers list --limit 5
assert_success


test_name "List raw transport trailers returns array"
xbe_json view raw-transport-trailers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list raw transport trailers"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List raw transport trailers with --broker filter"
xbe_json view raw-transport-trailers list --broker "$CREATED_BROKER_ID" --limit 5
assert_success


test_name "List raw transport trailers with --importer filter"
xbe_json view raw-transport-trailers list --importer "quantix_tmw" --limit 5
assert_success


test_name "List raw transport trailers with --import-status filter"
STATUS_FILTER="$IMPORT_STATUS"
if [[ -z "$STATUS_FILTER" || "$STATUS_FILTER" == "null" ]]; then
    STATUS_FILTER="pending"
fi
xbe_json view raw-transport-trailers list --import-status "$STATUS_FILTER" --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete raw transport trailer requires --confirm flag"
xbe_run do raw-transport-trailers delete "$CREATED_TRAILER_ID"
assert_failure


test_name "Delete raw transport trailer with --confirm"
xbe_run do raw-transport-trailers delete "$CREATED_TRAILER_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
