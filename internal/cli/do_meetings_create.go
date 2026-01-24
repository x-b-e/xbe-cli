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

type doMeetingsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	OrganizationType   string
	OrganizationID     string
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

func newDoMeetingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new meeting",
		Long: `Create a new meeting.

Required flags:
  --organization-type  Organization type (brokers, customers, truckers, material-suppliers, developers)
  --organization-id    Organization ID

Optional flags:
  --organizer            Organizer user ID
  --subject              Meeting subject
  --description          Meeting description
  --transcript           Meeting transcript
  --summary              Meeting summary
  --start-at             Start time (ISO 8601)
  --end-at               End time (ISO 8601)
  --explicit-time-zone-id Explicit time zone ID (e.g. America/Chicago)
  --address              Full address (will be geocoded unless --skip-geocoding)
  --address-latitude     Address latitude (use with --skip-geocoding)
  --address-longitude    Address longitude (use with --skip-geocoding)
  --address-place-id     Address place ID
  --address-plus-code    Address plus code
  --skip-geocoding       Skip address geocoding

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a meeting for a broker
  xbe do meetings create --organization-type brokers --organization-id 123 \\
    --subject "Weekly Safety Meeting" \\
    --start-at 2025-01-15T14:00:00Z --end-at 2025-01-15T15:00:00Z

  # Create with organizer and location
  xbe do meetings create --organization-type brokers --organization-id 123 \\
    --organizer 456 \\
    --address "123 Main St, Chicago, IL" \\
    --skip-geocoding --address-latitude 41.88 --address-longitude -87.63`,
		Args: cobra.NoArgs,
		RunE: runDoMeetingsCreate,
	}
	initDoMeetingsCreateFlags(cmd)
	return cmd
}

func init() {
	doMeetingsCmd.AddCommand(newDoMeetingsCreateCmd())
}

func initDoMeetingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-type", "", "Organization type (brokers, customers, truckers, material-suppliers, developers)")
	cmd.Flags().String("organization-id", "", "Organization ID")
	cmd.Flags().String("organizer", "", "Organizer user ID")
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

func runDoMeetingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMeetingsCreateOptions(cmd)
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

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.Subject != "" {
		attributes["subject"] = opts.Subject
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Transcript != "" {
		attributes["transcript"] = opts.Transcript
	}
	if opts.Summary != "" {
		attributes["summary"] = opts.Summary
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.ExplicitTimeZoneID != "" {
		attributes["explicit-time-zone-id"] = opts.ExplicitTimeZoneID
	}
	if opts.Address != "" {
		attributes["address"] = opts.Address
	}
	if opts.AddressLatitude != "" {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if opts.AddressLongitude != "" {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if opts.AddressPlaceID != "" {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if opts.AddressPlusCode != "" {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if opts.SkipGeocoding {
		attributes["skip-geocoding"] = true
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	if opts.OrganizerID != "" {
		relationships["organizer"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.OrganizerID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "meetings",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/meetings", jsonBody)
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

	if row.Subject != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created meeting %s (%s)\n", row.ID, row.Subject)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created meeting %s\n", row.ID)
	return nil
}

func parseDoMeetingsCreateOptions(cmd *cobra.Command) (doMeetingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
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

	return doMeetingsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		OrganizationType:   organizationType,
		OrganizationID:     organizationID,
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

func meetingRowFromSingle(resp jsonAPISingleResponse) meetingRow {
	attrs := resp.Data.Attributes
	return meetingRow{
		ID:      resp.Data.ID,
		Subject: strings.TrimSpace(stringAttr(attrs, "subject")),
		StartAt: formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:   formatDateTime(stringAttr(attrs, "end-at")),
	}
}
