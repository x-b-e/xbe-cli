package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMeetingsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	OrganizerID        string
	Subject            string
	Description        string
	Transcript         string
	Summary            string
	StartAt            string
	EndAt              string
	ExplicitTimeZoneID string
	Address            string
	AddressLatitude    string
	AddressLongitude   string
	AddressPlaceID     string
	AddressPlusCode    string
	SkipGeocoding      bool
}

func newDoMeetingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a meeting",
		Long: `Update a meeting.

Optional flags:
  --organizer              Organizer user ID (empty clears)
  --subject                Meeting subject
  --description            Meeting description
  --transcript             Meeting transcript
  --summary                Meeting summary
  --start-at               Start time (ISO 8601)
  --end-at                 End time (ISO 8601)
  --explicit-time-zone-id  Explicit time zone ID (e.g. America/Chicago)
  --address                Full address (will be geocoded unless --skip-geocoding)
  --address-latitude       Address latitude
  --address-longitude      Address longitude
  --address-place-id       Address place ID
  --address-plus-code      Address plus code
  --skip-geocoding         Skip address geocoding

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update meeting subject
  xbe do meetings update 123 --subject "Updated Subject"

  # Update meeting schedule
  xbe do meetings update 123 --start-at 2025-01-15T15:00:00Z --end-at 2025-01-15T16:00:00Z

  # Update address and skip geocoding
  xbe do meetings update 123 --address "456 Oak Ave, Springfield, IL" \\
    --skip-geocoding --address-latitude 39.78 --address-longitude -89.64`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMeetingsUpdate,
	}
	initDoMeetingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMeetingsCmd.AddCommand(newDoMeetingsUpdateCmd())
}

func initDoMeetingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organizer", "", "Organizer user ID (empty clears)")
	cmd.Flags().String("subject", "", "Meeting subject")
	cmd.Flags().String("description", "", "Meeting description")
	cmd.Flags().String("transcript", "", "Meeting transcript")
	cmd.Flags().String("summary", "", "Meeting summary")
	cmd.Flags().String("start-at", "", "Start time (ISO 8601)")
	cmd.Flags().String("end-at", "", "End time (ISO 8601)")
	cmd.Flags().String("explicit-time-zone-id", "", "Explicit time zone ID (e.g. America/Chicago)")
	cmd.Flags().String("address", "", "Full address (will be geocoded unless --skip-geocoding)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-place-id", "", "Address place ID")
	cmd.Flags().String("address-plus-code", "", "Address plus code")
	cmd.Flags().Bool("skip-geocoding", false, "Skip address geocoding")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMeetingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMeetingsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("subject") {
		attributes["subject"] = opts.Subject
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("transcript") {
		attributes["transcript"] = opts.Transcript
	}
	if cmd.Flags().Changed("summary") {
		attributes["summary"] = opts.Summary
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("explicit-time-zone-id") {
		attributes["explicit-time-zone-id"] = opts.ExplicitTimeZoneID
	}
	if cmd.Flags().Changed("address") {
		attributes["address"] = opts.Address
	}
	if cmd.Flags().Changed("address-latitude") {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if cmd.Flags().Changed("address-longitude") {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if cmd.Flags().Changed("address-place-id") {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if cmd.Flags().Changed("address-plus-code") {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if cmd.Flags().Changed("skip-geocoding") {
		attributes["skip-geocoding"] = opts.SkipGeocoding
	}

	if cmd.Flags().Changed("organizer") {
		if strings.TrimSpace(opts.OrganizerID) == "" {
			relationships["organizer"] = map[string]any{"data": nil}
		} else {
			relationships["organizer"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.OrganizerID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "meetings",
		"id":         opts.ID,
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/meetings/"+opts.ID, jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := meetingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated meeting %s\n", row.ID)
	return nil
}

func parseDoMeetingsUpdateOptions(cmd *cobra.Command, args []string) (doMeetingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizer, _ := cmd.Flags().GetString("organizer")
	subject, _ := cmd.Flags().GetString("subject")
	description, _ := cmd.Flags().GetString("description")
	transcript, _ := cmd.Flags().GetString("transcript")
	summary, _ := cmd.Flags().GetString("summary")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	explicitTimeZoneID, _ := cmd.Flags().GetString("explicit-time-zone-id")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	addressPlaceID, _ := cmd.Flags().GetString("address-place-id")
	addressPlusCode, _ := cmd.Flags().GetString("address-plus-code")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMeetingsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		OrganizerID:        organizer,
		Subject:            subject,
		Description:        description,
		Transcript:         transcript,
		Summary:            summary,
		StartAt:            startAt,
		EndAt:              endAt,
		ExplicitTimeZoneID: explicitTimeZoneID,
		Address:            address,
		AddressLatitude:    addressLatitude,
		AddressLongitude:   addressLongitude,
		AddressPlaceID:     addressPlaceID,
		AddressPlusCode:    addressPlusCode,
		SkipGeocoding:      skipGeocoding,
	}, nil
}
