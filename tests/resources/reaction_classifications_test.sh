#!/bin/bash
#
# XBE CLI Integration Tests: Reaction Classifications
#
# Tests create and list operations for the reaction_classifications resource.
# Reaction classifications categorize user reactions.
#
# COVERAGE: Create + list (create-only, no update/delete, minimal writable attributes)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RC_ID=""

describe "Resource: reaction-classifications"

# ============================================================================
# CREATE Tests
# ============================================================================

# Note: Reaction classifications have read-only label, utf8, and external_reference
# attributes. Only created_by can be set on create.

test_name "Create reaction classification"
# Note: Server may not support creating reaction classifications (reference data)
xbe_json do reaction-classifications create

if [[ $status -eq 0 ]]; then
    CREATED_RC_ID=$(json_get ".id")
    if [[ -n "$CREATED_RC_ID" && "$CREATED_RC_ID" != "null" ]]; then
        # Note: No delete available for reaction-classifications
        pass
    else
        fail "Created reaction classification but no ID returned"
    fi
else
    # 404 indicates server doesn't support creating this resource (reference data only)
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        echo "    (Server does not support creating reaction classifications - skipping)"
        skip "Server does not support create"
    else
        fail "Failed to create reaction classification: $output"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List reaction classifications"
xbe_json view reaction-classifications list
assert_success

test_name "List reaction classifications returns array"
xbe_json view reaction-classifications list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list reaction classifications"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List reaction classifications with --limit"
xbe_json view reaction-classifications list --limit 5
assert_success

test_name "List reaction classifications with --offset"
xbe_json view reaction-classifications list --limit 5 --offset 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
