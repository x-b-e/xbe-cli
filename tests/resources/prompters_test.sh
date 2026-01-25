#!/bin/bash
#
# XBE CLI Integration Tests: Prompters
#
# Tests CRUD operations for the prompters resource.
#
# COVERAGE: Create (all writable attributes), update (all writable attributes), list filters, show, delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROMPTER_ID=""
CREATED_PROMPTER_NAME=""

create_template_value() {
    local suffix="$1"
    echo "cli-${suffix}-$(date +%s)-${RANDOM}"
}

describe "Resource: prompters"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prompter with all attributes"
CREATED_PROMPTER_NAME="CLI_PROMPTER_$(date +%s)_${RANDOM}"

RELEASE_NOTE_NAV_TEMPLATE=$(create_template_value "rn-nav")
RELEASE_NOTE_HEADLINE_TEMPLATE=$(create_template_value "rn-headline")
RELEASE_NOTE_GLOSSARY_TEMPLATE=$(create_template_value "rn-glossary")
JPP_SAFETY_RISKS_TEMPLATE=$(create_template_value "jpp-risks")
JPP_SAFETY_COMM_TEMPLATE=$(create_template_value "jpp-comm")
INCIDENT_HEADLINE_TEMPLATE=$(create_template_value "incident")
GLOSSARY_DEFINITION_TEMPLATE=$(create_template_value "glossary-def")
CONDENSABLE_TEMPLATE=$(create_template_value "condense")
ANSWER_TEMPLATE=$(create_template_value "answer")
ACTION_ITEM_TEMPLATE=$(create_template_value "action-item")

xbe_json do prompters create \
    --name "$CREATED_PROMPTER_NAME" \
    --is-active=false \
    --release-note-guess-has-navigation-instructions-prompt-template "$RELEASE_NOTE_NAV_TEMPLATE" \
    --release-note-headline-suggestions-prompt-template "$RELEASE_NOTE_HEADLINE_TEMPLATE" \
    --release-note-glossary-term-suggestions-prompt-template "$RELEASE_NOTE_GLOSSARY_TEMPLATE" \
    --jpp-safety-risks-suggestion-suggestion-prompt-template "$JPP_SAFETY_RISKS_TEMPLATE" \
    --jpp-safety-risk-comm-suggestion-suggestion-prompt-template "$JPP_SAFETY_COMM_TEMPLATE" \
    --incident-headline-suggestion-suggestion-prompt-template "$INCIDENT_HEADLINE_TEMPLATE" \
    --glossary-term-definition-suggestions-prompt-template "$GLOSSARY_DEFINITION_TEMPLATE" \
    --condensable-condense-prompt-template "$CONDENSABLE_TEMPLATE" \
    --answer-answer-prompt-template "$ANSWER_TEMPLATE" \
    --action-item-summary-prompt-template "$ACTION_ITEM_TEMPLATE"

if [[ $status -eq 0 ]]; then
    CREATED_PROMPTER_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROMPTER_ID" && "$CREATED_PROMPTER_ID" != "null" ]]; then
        pass
    else
        fail "Created prompter but no ID returned"
    fi
else
    fail "Failed to create prompter"
fi

if [[ -z "$CREATED_PROMPTER_ID" || "$CREATED_PROMPTER_ID" == "null" ]]; then
    echo "Cannot continue without a valid prompter ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prompter"
xbe_json view prompters show "$CREATED_PROMPTER_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update prompter with all attributes"
UPDATED_PROMPTER_NAME="${CREATED_PROMPTER_NAME}_UPDATED"

UPDATED_RELEASE_NOTE_NAV_TEMPLATE=$(create_template_value "rn-nav-updated")
UPDATED_RELEASE_NOTE_HEADLINE_TEMPLATE=$(create_template_value "rn-headline-updated")
UPDATED_RELEASE_NOTE_GLOSSARY_TEMPLATE=$(create_template_value "rn-glossary-updated")
UPDATED_JPP_SAFETY_RISKS_TEMPLATE=$(create_template_value "jpp-risks-updated")
UPDATED_JPP_SAFETY_COMM_TEMPLATE=$(create_template_value "jpp-comm-updated")
UPDATED_INCIDENT_HEADLINE_TEMPLATE=$(create_template_value "incident-updated")
UPDATED_GLOSSARY_DEFINITION_TEMPLATE=$(create_template_value "glossary-def-updated")
UPDATED_CONDENSABLE_TEMPLATE=$(create_template_value "condense-updated")
UPDATED_ANSWER_TEMPLATE=$(create_template_value "answer-updated")
UPDATED_ACTION_ITEM_TEMPLATE=$(create_template_value "action-item-updated")

xbe_json do prompters update "$CREATED_PROMPTER_ID" \
    --name "$UPDATED_PROMPTER_NAME" \
    --is-active=false \
    --release-note-guess-has-navigation-instructions-prompt-template "$UPDATED_RELEASE_NOTE_NAV_TEMPLATE" \
    --release-note-headline-suggestions-prompt-template "$UPDATED_RELEASE_NOTE_HEADLINE_TEMPLATE" \
    --release-note-glossary-term-suggestions-prompt-template "$UPDATED_RELEASE_NOTE_GLOSSARY_TEMPLATE" \
    --jpp-safety-risks-suggestion-suggestion-prompt-template "$UPDATED_JPP_SAFETY_RISKS_TEMPLATE" \
    --jpp-safety-risk-comm-suggestion-suggestion-prompt-template "$UPDATED_JPP_SAFETY_COMM_TEMPLATE" \
    --incident-headline-suggestion-suggestion-prompt-template "$UPDATED_INCIDENT_HEADLINE_TEMPLATE" \
    --glossary-term-definition-suggestions-prompt-template "$UPDATED_GLOSSARY_DEFINITION_TEMPLATE" \
    --condensable-condense-prompt-template "$UPDATED_CONDENSABLE_TEMPLATE" \
    --answer-answer-prompt-template "$UPDATED_ANSWER_TEMPLATE" \
    --action-item-summary-prompt-template "$UPDATED_ACTION_ITEM_TEMPLATE"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List prompters"
xbe_json view prompters list --limit 5
assert_success

test_name "List prompters filtered by active status"
xbe_json view prompters list --is-active false --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete prompter requires --confirm"
xbe_run do prompters delete "$CREATED_PROMPTER_ID"
assert_failure

test_name "Delete prompter"
xbe_run do prompters delete "$CREATED_PROMPTER_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_run do prompters update "$CREATED_PROMPTER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
