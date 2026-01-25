#!/bin/bash
#
# XBE CLI Integration Tests: Tractor Credentials
#
# Tests CRUD operations for the tractor_credentials resource.
# Tractor credentials track credentials (licenses, registrations) for tractors.
#
# NOTE: Each create test uses a unique classification because credentials
# for the same tractor+classification cannot have overlapping date ranges.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CREDENTIAL_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRACTOR_ID=""
CREATED_CLASSIFICATION_ID=""

describe "Resource: tractor_credentials"

# ============================================================================
# Prerequisites - Create broker, trucker, tractor, and classification
# ============================================================================

test_name "Create prerequisite broker for tractor credentials tests"
BROKER_NAME=$(unique_name "TCTestBroker")

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

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "TCTestTrucker")
TRUCKER_ADDRESS="350 N Orleans St, Chicago, IL 60654"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

test_name "Create prerequisite tractor"
TRACTOR_NUMBER=$(unique_name "TCTractor")

xbe_json do tractors create \
    --number "$TRACTOR_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRACTOR_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRACTOR_ID" && "$CREATED_TRACTOR_ID" != "null" ]]; then
        register_cleanup "tractors" "$CREATED_TRACTOR_ID"
        pass
    else
        fail "Created tractor but no ID returned"
        echo "Cannot continue without a tractor"
        run_tests
    fi
else
    fail "Failed to create tractor"
    echo "Cannot continue without a tractor"
    run_tests
fi

test_name "Create prerequisite tractor trailer credential classification"
CLASS_NAME=$(unique_name "TCTestClass")

xbe_json do tractor-trailer-credential-classifications create \
    --name "$CLASS_NAME" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "tractor-trailer-credential-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created classification but no ID returned"
        echo "Cannot continue without a classification"
        run_tests
    fi
else
    fail "Failed to create tractor trailer credential classification"
    echo "Cannot continue without a classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create tractor credential with required fields"

xbe_json do tractor-credentials create \
    --tractor "$CREATED_TRACTOR_ID" \
    --tractor-trailer-credential-classification "$CREATED_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CREDENTIAL_ID=$(json_get ".id")
    if [[ -n "$CREATED_CREDENTIAL_ID" && "$CREATED_CREDENTIAL_ID" != "null" ]]; then
        register_cleanup "tractor-credentials" "$CREATED_CREDENTIAL_ID"
        pass
    else
        fail "Created tractor credential but no ID returned"
    fi
else
    fail "Failed to create tractor credential"
fi

# Only continue if we successfully created a credential
if [[ -z "$CREATED_CREDENTIAL_ID" || "$CREATED_CREDENTIAL_ID" == "null" ]]; then
    echo "Cannot continue without a valid tractor credential ID"
    run_tests
fi

test_name "Create tractor credential with issued-on"
# Create a unique classification to avoid overlap
CLASS_NAME2=$(unique_name "TCTestClass2")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$CLASS_NAME2" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID2=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$CLASS_ID2"
    xbe_json do tractor-credentials create \
        --tractor "$CREATED_TRACTOR_ID" \
        --tractor-trailer-credential-classification "$CLASS_ID2" \
        --issued-on "2024-01-15"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "tractor-credentials" "$id"
        pass
    else
        fail "Failed to create tractor credential with issued-on"
    fi
else
    fail "Failed to create classification for issued-on test"
fi

test_name "Create tractor credential with expires-on"
# Create a unique classification to avoid overlap
CLASS_NAME3=$(unique_name "TCTestClass3")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$CLASS_NAME3" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID3=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$CLASS_ID3"
    xbe_json do tractor-credentials create \
        --tractor "$CREATED_TRACTOR_ID" \
        --tractor-trailer-credential-classification "$CLASS_ID3" \
        --expires-on "2025-01-15"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "tractor-credentials" "$id"
        pass
    else
        fail "Failed to create tractor credential with expires-on"
    fi
else
    fail "Failed to create classification for expires-on test"
fi

test_name "Create tractor credential with both dates"
# Create a unique classification to avoid overlap
CLASS_NAME4=$(unique_name "TCTestClass4")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$CLASS_NAME4" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID4=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$CLASS_ID4"
    xbe_json do tractor-credentials create \
        --tractor "$CREATED_TRACTOR_ID" \
        --tractor-trailer-credential-classification "$CLASS_ID4" \
        --issued-on "2024-01-01" \
        --expires-on "2025-12-31"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "tractor-credentials" "$id"
        pass
    else
        fail "Failed to create tractor credential with both dates"
    fi
else
    fail "Failed to create classification for both dates test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update tractor credential issued-on"
xbe_json do tractor-credentials update "$CREATED_CREDENTIAL_ID" --issued-on "2024-02-01"
assert_success

test_name "Update tractor credential expires-on"
xbe_json do tractor-credentials update "$CREATED_CREDENTIAL_ID" --expires-on "2026-02-01"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tractor credentials"
xbe_json view tractor-credentials list --limit 5
assert_success

test_name "List tractor credentials returns array"
xbe_json view tractor-credentials list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tractor credentials"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List tractor credentials with --tractor filter"
xbe_json view tractor-credentials list --tractor "$CREATED_TRACTOR_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List tractor credentials with --limit"
xbe_json view tractor-credentials list --limit 3
assert_success

test_name "List tractor credentials with --offset"
xbe_json view tractor-credentials list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete tractor credential requires --confirm flag"
xbe_run do tractor-credentials delete "$CREATED_CREDENTIAL_ID"
assert_failure

test_name "Delete tractor credential with --confirm"
# Create a credential specifically for deletion (using unique classification)
CLASS_NAME_DEL=$(unique_name "TCTestClassDel")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$CLASS_NAME_DEL" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID_DEL=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$CLASS_ID_DEL"
    xbe_json do tractor-credentials create \
        --tractor "$CREATED_TRACTOR_ID" \
        --tractor-trailer-credential-classification "$CLASS_ID_DEL"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do tractor-credentials delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create tractor credential for deletion test"
    fi
else
    skip "Could not create classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tractor credential without tractor fails"
xbe_json do tractor-credentials create \
    --tractor-trailer-credential-classification "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Create tractor credential without classification fails"
xbe_json do tractor-credentials create \
    --tractor "$CREATED_TRACTOR_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do tractor-credentials update "$CREATED_CREDENTIAL_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
