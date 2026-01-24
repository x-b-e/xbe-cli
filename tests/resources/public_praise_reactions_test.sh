#!/bin/bash
#
# XBE CLI Integration Tests: Public Praise Reactions
#
# Tests view and mutation operations for public praise reactions.
#
# COVERAGE: Create relationships + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_REACTION_ID=""
REACTION_CLASSIFICATION_ID=""
REACTION_CLASSIFICATION_ID_ALT=""
PUBLIC_PRAISE_ID=""
PUBLIC_PRAISE_ID_ALT=""
CREATED_REACTION_ID=""
CREATED_REACTION_ALT_ID=""
CREATED_BROKER_ID=""
CREATED_BY_ID=""
RECEIVED_BY_ID=""
SKIP_MUTATION=0

describe "Resource: public-praise-reactions"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List public praise reactions"
xbe_json view public-praise-reactions list --limit 1
assert_success

test_name "List public praise reactions with created-at filter"
xbe_json view public-praise-reactions list --created-at-min 2024-01-01T00:00:00Z --limit 1
assert_success

test_name "List public praise reactions with updated-at filter"
xbe_json view public-praise-reactions list --updated-at-min 2024-01-01T00:00:00Z --limit 1
assert_success

test_name "Capture sample public praise reaction (if available)"
xbe_json view public-praise-reactions list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_REACTION_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No public praise reactions available; skipping show test."
        pass
    fi
else
    fail "Failed to list public praise reactions"
fi

if [[ -n "$SAMPLE_REACTION_ID" && "$SAMPLE_REACTION_ID" != "null" ]]; then
    test_name "Show public praise reaction"
    xbe_json view public-praise-reactions show "$SAMPLE_REACTION_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create public praise reaction requires --public-praise"
xbe_run do public-praise-reactions create --reaction-classification 1
assert_failure

test_name "Create public praise reaction requires --reaction-classification"
xbe_run do public-praise-reactions create --public-praise 1
assert_failure

# ============================================================================
# CREATE / DELETE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find reaction classifications"
    xbe_json view reaction-classifications list --limit 2
    if [[ $status -eq 0 ]]; then
        REACTION_CLASSIFICATION_ID=$(json_get ".[0].id")
        REACTION_CLASSIFICATION_ID_ALT=$(json_get ".[1].id")
        if [[ -n "$REACTION_CLASSIFICATION_ID" && "$REACTION_CLASSIFICATION_ID" != "null" ]]; then
            pass
        else
            skip "No reaction classifications available"
            SKIP_MUTATION=1
        fi
    else
        skip "Unable to list reaction classifications"
        SKIP_MUTATION=1
    fi
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Resolve current user for public praise creation"
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".id")
    fi
    if [[ -z "$CREATED_BY_ID" || "$CREATED_BY_ID" == "null" ]]; then
        skip "Auth whoami failed"
        SKIP_MUTATION=1
    else
        pass
    fi

    if [[ $SKIP_MUTATION -eq 0 ]]; then
        xbe_json view users list --limit 2
        if [[ $status -eq 0 ]]; then
            candidate_id=$(json_get ".[1].id")
            if [[ -n "$candidate_id" && "$candidate_id" != "null" && "$candidate_id" != "$CREATED_BY_ID" ]]; then
                RECEIVED_BY_ID="$candidate_id"
            else
                RECEIVED_BY_ID="$CREATED_BY_ID"
            fi
        else
            RECEIVED_BY_ID="$CREATED_BY_ID"
        fi
    fi

    test_name "Create prerequisite broker for public praise reaction tests"
    BROKER_NAME=$(unique_name "PublicPraiseReactionBroker")

    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
            SKIP_MUTATION=1
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            skip "Failed to create broker and XBE_TEST_BROKER_ID not set"
            SKIP_MUTATION=1
        fi
    fi
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create prerequisite public praise"
    PRAISE_DESCRIPTION=$(unique_name "PublicPraiseReaction")

    xbe_json do public-praises create \
        --description "$PRAISE_DESCRIPTION" \
        --given-by "$CREATED_BY_ID" \
        --received-by "$RECEIVED_BY_ID" \
        --organization-type brokers \
        --organization-id "$CREATED_BROKER_ID"

    if [[ $status -eq 0 ]]; then
        PUBLIC_PRAISE_ID=$(json_get ".id")
        if [[ -n "$PUBLIC_PRAISE_ID" && "$PUBLIC_PRAISE_ID" != "null" ]]; then
            register_cleanup "public-praises" "$PUBLIC_PRAISE_ID"
            pass
        else
            fail "Created public praise but no ID returned"
            SKIP_MUTATION=1
        fi
    else
        skip "Failed to create public praise"
        SKIP_MUTATION=1
    fi
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create public praise reaction (required fields)"
    xbe_json do public-praise-reactions create \
        --public-praise "$PUBLIC_PRAISE_ID" \
        --reaction-classification "$REACTION_CLASSIFICATION_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_REACTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_REACTION_ID" && "$CREATED_REACTION_ID" != "null" ]]; then
            register_cleanup "public-praise-reactions" "$CREATED_REACTION_ID"
            pass
        else
            fail "Created public praise reaction but no ID returned"
        fi
    else
        fail "Failed to create public praise reaction"
    fi
fi

if [[ $SKIP_MUTATION -eq 0 && -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    attempted_created_by=0
    if [[ -z "$REACTION_CLASSIFICATION_ID_ALT" || "$REACTION_CLASSIFICATION_ID_ALT" == "null" ]]; then
        test_name "Create secondary public praise for created-by test"
        PRAISE_DESCRIPTION_ALT=$(unique_name "PublicPraiseReactionAlt")

        xbe_json do public-praises create \
            --description "$PRAISE_DESCRIPTION_ALT" \
            --given-by "$CREATED_BY_ID" \
            --received-by "$RECEIVED_BY_ID" \
            --organization-type brokers \
            --organization-id "$CREATED_BROKER_ID"

        if [[ $status -eq 0 ]]; then
            PUBLIC_PRAISE_ID_ALT=$(json_get ".id")
            if [[ -n "$PUBLIC_PRAISE_ID_ALT" && "$PUBLIC_PRAISE_ID_ALT" != "null" ]]; then
                register_cleanup "public-praises" "$PUBLIC_PRAISE_ID_ALT"
            fi
        fi
    fi

    if [[ -n "$PUBLIC_PRAISE_ID_ALT" && "$PUBLIC_PRAISE_ID_ALT" != "null" ]]; then
        test_name "Create public praise reaction with --created-by"
        xbe_json do public-praise-reactions create \
            --public-praise "$PUBLIC_PRAISE_ID_ALT" \
            --reaction-classification "$REACTION_CLASSIFICATION_ID" \
            --created-by "$CREATED_BY_ID"
        attempted_created_by=1
    elif [[ -n "$REACTION_CLASSIFICATION_ID_ALT" && "$REACTION_CLASSIFICATION_ID_ALT" != "null" ]]; then
        test_name "Create public praise reaction with --created-by"
        xbe_json do public-praise-reactions create \
            --public-praise "$PUBLIC_PRAISE_ID" \
            --reaction-classification "$REACTION_CLASSIFICATION_ID_ALT" \
            --created-by "$CREATED_BY_ID"
        attempted_created_by=1
    else
        test_name "Create public praise reaction with --created-by"
        skip "No alternate reaction classification or public praise available"
    fi

    if [[ $attempted_created_by -eq 1 ]]; then
        if [[ $status -eq 0 ]]; then
            CREATED_REACTION_ALT_ID=$(json_get ".id")
            if [[ -n "$CREATED_REACTION_ALT_ID" && "$CREATED_REACTION_ALT_ID" != "null" ]]; then
                register_cleanup "public-praise-reactions" "$CREATED_REACTION_ALT_ID"
                pass
            else
                fail "Created public praise reaction but no ID returned"
            fi
        else
            skip "Create with --created-by not permitted"
        fi
    fi
fi

if [[ -n "$CREATED_REACTION_ID" && "$CREATED_REACTION_ID" != "null" ]]; then
    test_name "Delete public praise reaction requires --confirm flag"
    xbe_run do public-praise-reactions delete "$CREATED_REACTION_ID"
    assert_failure

    test_name "Delete public praise reaction"
    xbe_json do public-praise-reactions delete "$CREATED_REACTION_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
