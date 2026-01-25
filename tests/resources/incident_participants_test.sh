#!/bin/bash
#
# XBE CLI Integration Tests: Incident Participants
#
# Tests CRUD operations and list filters for the incident-participants resource.
#
# COVERAGE: Create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_INCIDENT_ID=""
CREATED_PARTICIPANT_ID=""
CURRENT_USER_ID=""

TEST_EMAIL=""
UPDATED_EMAIL=""
UPDATED_MOBILE=""

START_AT="2025-01-01T08:00:00Z"


describe "Resource: incident-participants"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List incident participants"
xbe_json view incident-participants list --limit 5
assert_success

test_name "List incident participants returns array"
xbe_json view incident-participants list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list incident participants"
fi

# ============================================================================
# Prerequisites - Broker, incident, user
# ============================================================================

test_name "Create prerequisite broker for incident participant tests"
BROKER_NAME=$(unique_name "IncidentParticipantBroker")

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

# Fetch current user (for user filter coverage)
test_name "Fetch current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "No user ID returned"
    fi
else
    fail "Failed to fetch current user"
fi

# Create prerequisite incident

test_name "Create prerequisite efficiency incident"
SUBJECT_VALUE="Broker|$CREATED_BROKER_ID"
xbe_json do efficiency-incidents create \
    --subject "$SUBJECT_VALUE" \
    --start-at "$START_AT" \
    --status open \
    --kind over_trucking \
    --headline "Incident participant test"

if [[ $status -eq 0 ]]; then
    CREATED_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_INCIDENT_ID" && "$CREATED_INCIDENT_ID" != "null" ]]; then
        register_cleanup "efficiency-incidents" "$CREATED_INCIDENT_ID"
        pass
    else
        fail "Created incident but no ID returned"
        echo "Cannot continue without an incident"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_INCIDENT_ID" ]]; then
        CREATED_INCIDENT_ID="$XBE_TEST_INCIDENT_ID"
        echo "    Using XBE_TEST_INCIDENT_ID: $CREATED_INCIDENT_ID"
        pass
    else
        fail "Failed to create incident and XBE_TEST_INCIDENT_ID not set"
        echo "Cannot continue without an incident"
        run_tests
    fi
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create incident participant requires --incident"
TEST_EMAIL=$(unique_email)
xbe_json do incident-participants create --name "Test Participant" --email-address "$TEST_EMAIL"
assert_failure

test_name "Create incident participant requires --name"
if [[ -n "$CREATED_INCIDENT_ID" && "$CREATED_INCIDENT_ID" != "null" ]]; then
    xbe_json do incident-participants create --incident "$CREATED_INCIDENT_ID" --email-address "$TEST_EMAIL"
    assert_failure
else
    skip "No incident ID available"
fi

test_name "Create incident participant requires email or mobile"
if [[ -n "$CREATED_INCIDENT_ID" && "$CREATED_INCIDENT_ID" != "null" ]]; then
    xbe_json do incident-participants create --incident "$CREATED_INCIDENT_ID" --name "Test Participant"
    assert_failure
else
    skip "No incident ID available"
fi

test_name "Create incident participant"
if [[ -n "$CREATED_INCIDENT_ID" && "$CREATED_INCIDENT_ID" != "null" ]]; then
    TEST_EMAIL=$(unique_email)
    xbe_json do incident-participants create \
        --incident "$CREATED_INCIDENT_ID" \
        --name "Incident Participant" \
        --email-address "$TEST_EMAIL" \
        --involvement witness

    if [[ $status -eq 0 ]]; then
        CREATED_PARTICIPANT_ID=$(json_get ".id")
        if [[ -n "$CREATED_PARTICIPANT_ID" && "$CREATED_PARTICIPANT_ID" != "null" ]]; then
            register_cleanup "incident-participants" "$CREATED_PARTICIPANT_ID"
            pass
        else
            fail "Created incident participant but no ID returned"
        fi
    else
        fail "Failed to create incident participant"
    fi
else
    skip "No incident ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update incident participant attributes"
if [[ -n "$CREATED_PARTICIPANT_ID" && "$CREATED_PARTICIPANT_ID" != "null" ]]; then
    UPDATED_EMAIL=$(unique_email)
    UPDATED_MOBILE=$(unique_mobile)
    xbe_json do incident-participants update "$CREATED_PARTICIPANT_ID" \
        --name "Incident Participant Updated" \
        --email-address "$UPDATED_EMAIL" \
        --mobile-number "$UPDATED_MOBILE" \
        --involvement observer
    assert_success
else
    skip "No incident participant ID available"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show incident participant"
if [[ -n "$CREATED_PARTICIPANT_ID" && "$CREATED_PARTICIPANT_ID" != "null" ]]; then
    xbe_json view incident-participants show "$CREATED_PARTICIPANT_ID"
    assert_success
else
    skip "No incident participant ID available"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List incident participants with --incident filter"
if [[ -n "$CREATED_INCIDENT_ID" && "$CREATED_INCIDENT_ID" != "null" ]]; then
    xbe_json view incident-participants list --incident "$CREATED_INCIDENT_ID" --limit 5
    assert_success
else
    skip "No incident ID available"
fi

test_name "List incident participants with --user filter"
if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    xbe_json view incident-participants list --user "$CURRENT_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List incident participants with --email-address filter"
if [[ -n "$UPDATED_EMAIL" ]]; then
    xbe_json view incident-participants list --email-address "$UPDATED_EMAIL" --limit 5
    assert_success
else
    skip "No email address available"
fi

test_name "List incident participants with --mobile-number filter"
if [[ -n "$UPDATED_MOBILE" ]]; then
    xbe_json view incident-participants list --mobile-number "$UPDATED_MOBILE" --limit 5
    assert_success
else
    skip "No mobile number available"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete incident participant requires --confirm"
if [[ -n "$CREATED_PARTICIPANT_ID" && "$CREATED_PARTICIPANT_ID" != "null" ]]; then
    xbe_run do incident-participants delete "$CREATED_PARTICIPANT_ID"
    assert_failure
else
    skip "No incident participant ID available"
fi

test_name "Delete incident participant"
if [[ -n "$CREATED_PARTICIPANT_ID" && "$CREATED_PARTICIPANT_ID" != "null" ]]; then
    xbe_run do incident-participants delete "$CREATED_PARTICIPANT_ID" --confirm
    assert_success
else
    skip "No incident participant ID available"
fi

run_tests
