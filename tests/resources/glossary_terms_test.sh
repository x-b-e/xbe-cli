#!/bin/bash
#
# XBE CLI Integration Tests: Glossary Terms
#
# Tests CRUD operations for the glossary_terms resource.
# Glossary terms define terminology used in the application.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_GLOSSARY_TERM_ID=""

describe "Resource: glossary_terms"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create glossary term with required fields"
TEST_TERM=$(unique_name "GlossaryTerm")

xbe_json do glossary-terms create \
    --term "$TEST_TERM" \
    --definition "A test definition for this glossary term"

if [[ $status -eq 0 ]]; then
    CREATED_GLOSSARY_TERM_ID=$(json_get ".id")
    if [[ -n "$CREATED_GLOSSARY_TERM_ID" && "$CREATED_GLOSSARY_TERM_ID" != "null" ]]; then
        register_cleanup "glossary-terms" "$CREATED_GLOSSARY_TERM_ID"
        pass
    else
        fail "Created glossary term but no ID returned"
    fi
else
    fail "Failed to create glossary term"
fi

# Only continue if we successfully created a glossary term
if [[ -z "$CREATED_GLOSSARY_TERM_ID" || "$CREATED_GLOSSARY_TERM_ID" == "null" ]]; then
    echo "Cannot continue without a valid glossary term ID"
    run_tests
fi

test_name "Create glossary term with source"
TEST_TERM2=$(unique_name "GlossaryTerm2")
xbe_json do glossary-terms create \
    --term "$TEST_TERM2" \
    --definition "Definition with source" \
    --source "expert"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "glossary-terms" "$id"
    pass
else
    fail "Failed to create glossary term with source"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update glossary term term"
UPDATED_TERM=$(unique_name "UpdatedGT")
xbe_json do glossary-terms update "$CREATED_GLOSSARY_TERM_ID" --term "$UPDATED_TERM"
assert_success

test_name "Update glossary term definition"
xbe_json do glossary-terms update "$CREATED_GLOSSARY_TERM_ID" --definition "Updated definition text"
assert_success

test_name "Update glossary term source"
xbe_json do glossary-terms update "$CREATED_GLOSSARY_TERM_ID" --source "xbe"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List glossary terms"
xbe_json view glossary-terms list --limit 5
assert_success

test_name "List glossary terms returns array"
xbe_json view glossary-terms list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list glossary terms"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List glossary terms with --source filter"
xbe_json view glossary-terms list --source "xbe" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List glossary terms with --limit"
xbe_json view glossary-terms list --limit 3
assert_success

test_name "List glossary terms with --offset"
xbe_json view glossary-terms list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete glossary term requires --confirm flag"
xbe_run do glossary-terms delete "$CREATED_GLOSSARY_TERM_ID"
assert_failure

test_name "Delete glossary term with --confirm"
# Create a glossary term specifically for deletion
TEST_DEL_TERM=$(unique_name "DeleteGT")
xbe_json do glossary-terms create \
    --term "$TEST_DEL_TERM" \
    --definition "Term to be deleted"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do glossary-terms delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create glossary term for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create glossary term without term fails"
xbe_json do glossary-terms create --definition "No term"
assert_failure

test_name "Create glossary term without definition fails"
xbe_json do glossary-terms create --term "NoDefinition"
assert_failure

test_name "Update without any fields fails"
xbe_json do glossary-terms update "$CREATED_GLOSSARY_TERM_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
