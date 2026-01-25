#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Mergers
#
# Tests create operations for material-site-mergers.
#
# COVERAGE: create
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

ORPHAN_ID="${XBE_TEST_MATERIAL_SITE_ORPHAN_ID:-}"
SURVIVOR_ID="${XBE_TEST_MATERIAL_SITE_SURVIVOR_ID:-}"

describe "Resource: material-site-mergers"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires --orphan"
xbe_run do material-site-mergers create --survivor "123"
assert_failure

test_name "Create requires --survivor"
xbe_run do material-site-mergers create --orphan "123"
assert_failure

test_name "Merge material site"
if [[ -n "$ORPHAN_ID" && "$ORPHAN_ID" != "null" && -n "$SURVIVOR_ID" && "$SURVIVOR_ID" != "null" ]]; then
    xbe_json do material-site-mergers create --orphan "$ORPHAN_ID" --survivor "$SURVIVOR_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"same brokers"* ]] || [[ "$output" == *"not found"* ]] || [[ "$output" == *"Validation"* ]]; then
            skip "Unable to merge material sites with provided IDs"
        else
            fail "Failed to merge material sites"
        fi
    fi
else
    skip "Set XBE_TEST_MATERIAL_SITE_ORPHAN_ID and XBE_TEST_MATERIAL_SITE_SURVIVOR_ID to enable create testing."
fi

run_tests
