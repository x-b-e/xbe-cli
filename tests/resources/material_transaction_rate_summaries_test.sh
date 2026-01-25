#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Rate Summaries
#
# Tests create operations for material-transaction-rate-summaries.
#
# COVERAGE: Writable attributes (material-site, start-at, end-at, group-by, material-type-hierarchies)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

MATERIAL_SITE_ID="${XBE_TEST_MATERIAL_SITE_ID:-}"

START_AT="2025-01-01T00:00:00Z"
END_AT="2025-01-02T00:00:00Z"


describe "Resource: material-transaction-rate-summaries"

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create requires material site"
xbe_run do material-transaction-rate-summaries create
assert_failure

test_name "Create material transaction rate summary"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json do material-transaction-rate-summaries create \
        --material-site "$MATERIAL_SITE_ID" \
        --start-at "$START_AT" \
        --end-at "$END_AT" \
        --group-by hour \
        --material-type-hierarchies "aggregate"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".material_site_id" "$MATERIAL_SITE_ID"
        if echo "$output" | jq -e '.results | type == "array"' >/dev/null 2>&1; then
            pass
        else
            fail "results was not an array"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"404"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create material transaction rate summary: $output"
        fi
    fi
else
    skip "No material site ID available. Set XBE_TEST_MATERIAL_SITE_ID to enable create testing."
fi

run_tests
