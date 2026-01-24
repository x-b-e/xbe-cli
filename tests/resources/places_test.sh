#!/bin/bash
#
# XBE CLI Integration Tests: Places
#
# Tests view operations for the places resource.
# Places return formatted addresses and coordinates for a Google Place ID.
#
# COVERAGE: Show (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: places (view-only)"

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show place"
PLACE_ID="ChIJD7fiBh9u5kcRYJSMaMOCCwQ"

xbe_json view places show "$PLACE_ID"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$PLACE_ID"
    assert_json_equals ".place_id" "$PLACE_ID"
    assert_json_has ".formatted_address"
    assert_json_has ".latitude"
    assert_json_has ".longitude"
else
    fail "Failed to show place"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
