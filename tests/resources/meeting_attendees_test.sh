#!/bin/bash
#
# XBE CLI Integration Tests: Meeting Attendees
#
# Tests list/show and create/update/delete behavior for meeting-attendees.
#
# COVERAGE: List filters + show + create/update attributes + delete
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: meeting-attendees"

CREATED_BROKER_ID=""
CREATED_MEETING_ID=""
CREATED_MEETING_VIA_API=0
CREATED_USER_ID=""
CREATED_ATTENDEE_ID=""

cleanup_meeting() {
	if [[ "$CREATED_MEETING_VIA_API" -eq 1 && -n "$CREATED_MEETING_ID" && "$CREATED_MEETING_ID" != "null" && -n "$XBE_TOKEN" ]]; then
		curl -s -o /dev/null \
			-H "Authorization: Bearer $XBE_TOKEN" \
			-H "Accept: application/vnd.api+json" \
			-X DELETE "$XBE_BASE_URL/v1/meetings/$CREATED_MEETING_ID" >/dev/null 2>&1 || true
	fi
}

trap 'cleanup_meeting; run_cleanup' EXIT

# ==========================================================================
# Prerequisites - Create or locate meeting and user
# ==========================================================================

if [[ -n "$XBE_TOKEN" ]]; then
	test_name "Create prerequisite broker for meeting attendee tests"
	BROKER_NAME=$(unique_name "MeetingAttendeeBroker")

	xbe_json do brokers create --name "$BROKER_NAME"
	if [[ $status -eq 0 ]]; then
		CREATED_BROKER_ID=$(json_get ".id")
		if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
			register_cleanup "brokers" "$CREATED_BROKER_ID"
			pass
		else
			fail "Created broker but no ID returned"
			run_tests
		fi
	else
		if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
			CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
			echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
			pass
		else
			fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
			run_tests
		fi
	fi

	test_name "Create prerequisite meeting for meeting attendee tests"
	MEETING_SUBJECT=$(unique_name "MeetingAttendee")
	response_file=$(mktemp)
	payload=$(cat <<PAYLOAD
{"data":{"type":"meetings","attributes":{"subject":"$MEETING_SUBJECT"},"relationships":{"organization":{"data":{"type":"brokers","id":"$CREATED_BROKER_ID"}}}}}
PAYLOAD
)
	run curl -s -o "$response_file" -w "%{http_code}" \
		-H "Authorization: Bearer $XBE_TOKEN" \
		-H "Accept: application/vnd.api+json" \
		-H "Content-Type: application/vnd.api+json" \
		-X POST "$XBE_BASE_URL/v1/meetings" \
		-d "$payload"

	http_code="$output"
	if [[ $status -eq 0 && "$http_code" == 2* ]]; then
		CREATED_MEETING_ID=$(jq -r '.data.id' "$response_file" 2>/dev/null)
		rm -f "$response_file"
		if [[ -n "$CREATED_MEETING_ID" && "$CREATED_MEETING_ID" != "null" ]]; then
			CREATED_MEETING_VIA_API=1
			pass
		else
			fail "Created meeting but no ID returned"
			run_tests
		fi
	else
		rm -f "$response_file"
		if [[ -n "$XBE_TEST_MEETING_ID" ]]; then
			CREATED_MEETING_ID="$XBE_TEST_MEETING_ID"
			echo "    Using XBE_TEST_MEETING_ID: $CREATED_MEETING_ID"
			pass
		else
			fail "Failed to create meeting (HTTP ${http_code})"
			run_tests
		fi
	fi

	test_name "Create prerequisite user for meeting attendee tests"
	USER_EMAIL=$(unique_email)
	USER_NAME=$(unique_name "MeetingAttendeeUser")

	xbe_json do users create --email "$USER_EMAIL" --name "$USER_NAME"
	if [[ $status -eq 0 ]]; then
		CREATED_USER_ID=$(json_get ".id")
		if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
			pass
		else
			fail "Created user but no ID returned"
			run_tests
		fi
	else
		fail "Failed to create user"
		run_tests
	fi
else
	test_name "Resolve current user for meeting attendee tests"
	xbe_json auth whoami
	if [[ $status -eq 0 ]]; then
		CREATED_USER_ID=$(json_get ".id")
		if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
			pass
		else
			fail "Could not resolve current user ID"
			run_tests
		fi
	else
		fail "Failed to resolve current user"
		run_tests
	fi

	test_name "Select meeting for meeting attendee tests"
	if [[ -n "$XBE_TEST_MEETING_ID" ]]; then
		CREATED_MEETING_ID="$XBE_TEST_MEETING_ID"
		echo "    Using XBE_TEST_MEETING_ID: $CREATED_MEETING_ID"
		pass
	else
		xbe_json view meeting-attendees list --limit 25
		if [[ $status -eq 0 ]]; then
			CREATED_MEETING_ID=$(echo "$output" | jq -r --arg user "$CREATED_USER_ID" 'map(select(.user_id != $user)) | .[0].meeting_id // empty')
			if [[ -n "$CREATED_MEETING_ID" && "$CREATED_MEETING_ID" != "null" ]]; then
				pass
			else
				skip "No meeting found via meeting-attendees list; set XBE_TEST_MEETING_ID to run tests"
				run_tests
			fi
		else
			skip "Unable to list meeting attendees; set XBE_TEST_MEETING_ID to run tests"
			run_tests
		fi
	fi
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create meeting attendee without required fields fails"
xbe_json do meeting-attendees create
assert_failure

test_name "Create meeting attendee"
xbe_json do meeting-attendees create \
	--meeting "$CREATED_MEETING_ID" \
	--user "$CREATED_USER_ID" \
	--location-kind on_site \
	--is-presence-required=false \
	--is-present=true \
	--location-latitude "41.0" \
	--location-longitude "-87.0"

if [[ $status -eq 0 ]]; then
	CREATED_ATTENDEE_ID=$(json_get ".id")
	if [[ -n "$CREATED_ATTENDEE_ID" && "$CREATED_ATTENDEE_ID" != "null" ]]; then
		register_cleanup "meeting-attendees" "$CREATED_ATTENDEE_ID"
		pass
	else
		fail "Created meeting attendee but no ID returned"
		run_tests
	fi
else
	fail "Failed to create meeting attendee"
	run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show meeting attendee"
xbe_json view meeting-attendees show "$CREATED_ATTENDEE_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update meeting attendee"
xbe_json do meeting-attendees update "$CREATED_ATTENDEE_ID" \
	--location-kind remote \
	--is-presence-required=true \
	--is-present=false \
	--location-latitude "41.1" \
	--location-longitude "-87.1"
assert_success

# ==========================================================================
# LIST Tests
# ==========================================================================

test_name "List meeting attendees"
xbe_json view meeting-attendees list --limit 5
assert_success

test_name "List meeting attendees with meeting filter"
xbe_json view meeting-attendees list --meeting "$CREATED_MEETING_ID" --limit 5
assert_success

test_name "List meeting attendees with user filter"
xbe_json view meeting-attendees list --user "$CREATED_USER_ID" --limit 5
assert_success

test_name "List meeting attendees with location kind filter"
xbe_json view meeting-attendees list --location-kind remote --limit 5
assert_success

test_name "List meeting attendees with is-presence-required filter"
xbe_json view meeting-attendees list --is-presence-required true --limit 5
assert_success

test_name "List meeting attendees with is-present filter"
xbe_json view meeting-attendees list --is-present false --limit 5
assert_success

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Delete meeting attendee without confirm fails"
xbe_run do meeting-attendees delete "$CREATED_ATTENDEE_ID"
assert_failure

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete meeting attendee"
xbe_run do meeting-attendees delete "$CREATED_ATTENDEE_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
