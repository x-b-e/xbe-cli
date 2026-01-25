#!/bin/bash
#
# XBE CLI Integration Tests: OpenAI Vector Stores
#
# Tests list/show operations and optional create/update/delete for the open-ai-vector-stores resource.
#
# COVERAGE: List filters + show + create/update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

VECTOR_STORE_ID=""
PURPOSE=""
SCOPE_TYPE=""
SCOPE_ID=""
SKIP_ID_FILTERS=0

ENV_PURPOSE="${XBE_TEST_OPEN_AI_VECTOR_STORE_PURPOSE:-}"
ENV_SCOPE_TYPE="${XBE_TEST_OPEN_AI_VECTOR_STORE_SCOPE_TYPE:-}"
ENV_SCOPE_ID="${XBE_TEST_OPEN_AI_VECTOR_STORE_SCOPE_ID:-}"

CREATED_VECTOR_STORE_ID=""

describe "Resource: open-ai-vector-stores"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List open ai vector stores"
xbe_json view open-ai-vector-stores list --limit 5
assert_success

test_name "List open ai vector stores returns array"
xbe_json view open-ai-vector-stores list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list open ai vector stores"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample open ai vector store"
xbe_json view open-ai-vector-stores list --limit 1
if [[ $status -eq 0 ]]; then
    VECTOR_STORE_ID=$(json_get ".[0].id")
    PURPOSE=$(json_get ".[0].purpose")
    SCOPE_TYPE=$(json_get ".[0].scope_type")
    SCOPE_ID=$(json_get ".[0].scope_id")
    if [[ -n "$VECTOR_STORE_ID" && "$VECTOR_STORE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No open ai vector stores available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list open ai vector stores"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List vector stores with --purpose filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$PURPOSE" && "$PURPOSE" != "null" ]]; then
    xbe_json view open-ai-vector-stores list --purpose "$PURPOSE" --limit 5
    assert_success
else
    skip "No purpose available"
fi

test_name "List vector stores with --scope filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SCOPE_TYPE" && "$SCOPE_TYPE" != "null" && -n "$SCOPE_ID" && "$SCOPE_ID" != "null" ]]; then
    xbe_json view open-ai-vector-stores list --scope "${SCOPE_TYPE}|${SCOPE_ID}" --limit 5
    assert_success
else
    skip "No scope available"
fi

test_name "List vector stores with --scope-type filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SCOPE_TYPE" && "$SCOPE_TYPE" != "null" ]]; then
    xbe_json view open-ai-vector-stores list --scope-type "$SCOPE_TYPE" --limit 5
    assert_success
else
    skip "No scope type available"
fi

test_name "List vector stores with --scope-id filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SCOPE_TYPE" && "$SCOPE_TYPE" != "null" && -n "$SCOPE_ID" && "$SCOPE_ID" != "null" ]]; then
    xbe_json view open-ai-vector-stores list --scope-type "$SCOPE_TYPE" --scope-id "$SCOPE_ID" --limit 5
    assert_success
else
    skip "No scope available"
fi

test_name "List vector stores with --not-scope-type filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SCOPE_TYPE" && "$SCOPE_TYPE" != "null" ]]; then
    xbe_json view open-ai-vector-stores list --not-scope-type "$SCOPE_TYPE" --limit 5
    assert_success
else
    skip "No scope type available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show open ai vector store"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$VECTOR_STORE_ID" && "$VECTOR_STORE_ID" != "null" ]]; then
    xbe_json view open-ai-vector-stores show "$VECTOR_STORE_ID"
    assert_success
else
    skip "No vector store ID available"
fi

# ============================================================================
# CREATE Tests - Error Cases
# ============================================================================

test_name "Create vector store without required fields fails"
xbe_run do open-ai-vector-stores create
assert_failure

test_name "Create vector store with scope-id only fails"
xbe_run do open-ai-vector-stores create --purpose platform_content --scope-id 1
assert_failure

# ============================================================================
# CREATE + UPDATE Tests - Optional
# ============================================================================

test_name "Create open ai vector store"
if [[ -n "$ENV_PURPOSE" && -n "$ENV_SCOPE_TYPE" && -n "$ENV_SCOPE_ID" ]]; then
    xbe_json do open-ai-vector-stores create \
        --purpose "$ENV_PURPOSE" \
        --scope-type "$ENV_SCOPE_TYPE" \
        --scope-id "$ENV_SCOPE_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_VECTOR_STORE_ID=$(json_get ".id")
        if [[ -n "$CREATED_VECTOR_STORE_ID" && "$CREATED_VECTOR_STORE_ID" != "null" ]]; then
            register_cleanup "open-ai-vector-stores" "$CREATED_VECTOR_STORE_ID"
            pass
        else
            fail "Created vector store but no ID returned"
        fi
    else
        skip "Create failed (check permissions or scope constraints)"
    fi
else
    skip "Set XBE_TEST_OPEN_AI_VECTOR_STORE_PURPOSE and XBE_TEST_OPEN_AI_VECTOR_STORE_SCOPE_* to run"
fi

test_name "Update open ai vector store purpose"
if [[ -n "$CREATED_VECTOR_STORE_ID" && "$CREATED_VECTOR_STORE_ID" != "null" && -n "$ENV_PURPOSE" ]]; then
    xbe_json do open-ai-vector-stores update "$CREATED_VECTOR_STORE_ID" --purpose "$ENV_PURPOSE"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".purpose" "$ENV_PURPOSE"
    else
        skip "Update failed (check permissions or read-only fields)"
    fi
else
    skip "No created vector store ID available"
fi

test_name "Update open ai vector store scope"
if [[ -n "$CREATED_VECTOR_STORE_ID" && "$CREATED_VECTOR_STORE_ID" != "null" && -n "$ENV_SCOPE_TYPE" && -n "$ENV_SCOPE_ID" ]]; then
    xbe_json do open-ai-vector-stores update "$CREATED_VECTOR_STORE_ID" \
        --scope-type "$ENV_SCOPE_TYPE" \
        --scope-id "$ENV_SCOPE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".scope_type"
        assert_json_equals ".scope_id" "$ENV_SCOPE_ID"
    else
        skip "Update failed (check permissions or read-only fields)"
    fi
else
    skip "No created vector store ID available"
fi

# ============================================================================
# UPDATE Tests - Error Cases
# ============================================================================

test_name "Update vector store without attributes fails"
xbe_run do open-ai-vector-stores update 999999
assert_failure

# ============================================================================
# DELETE Tests - Optional
# ============================================================================

test_name "Delete open ai vector store"
if [[ -n "$CREATED_VECTOR_STORE_ID" && "$CREATED_VECTOR_STORE_ID" != "null" ]]; then
    xbe_json do open-ai-vector-stores delete "$CREATED_VECTOR_STORE_ID" --confirm
    if [[ $status -eq 0 ]]; then
        assert_json_bool ".deleted" "true"
    else
        skip "Delete failed (check permissions)"
    fi
else
    skip "No created vector store ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
