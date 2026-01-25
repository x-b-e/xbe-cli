#!/bin/bash
#
# XBE CLI Integration Tests: Model Filter Infos
#
# Tests create behavior for model filter infos.
#
# COVERAGE: Create + required flag + filter keys + scope filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: model-filter-infos"

test_name "Create model filter info requires --resource-type"
xbe_run do model-filter-infos create
assert_failure

test_name "Create model filter info"
xbe_json do model-filter-infos create --resource-type projects
if [[ $status -eq 0 ]]; then
    assert_json_equals ".resource_type" "projects"
    assert_json_has ".options"
else
    fail "Failed to create model filter info"
fi

test_name "Create model filter info with filter keys and scope filters"
xbe_json do model-filter-infos create --resource-type projects --filter-keys customer,project_manager --scope-filter broker=123
if [[ $status -eq 0 ]]; then
    assert_json_equals ".resource_type" "projects"
    assert_json_has ".filter_keys[] | select(. == \"customer\")"
    assert_json_has ".scope_filters.broker"
else
    fail "Failed to create model filter info with filter keys"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
