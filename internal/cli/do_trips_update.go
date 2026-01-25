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

type doTripsUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	OriginType             string
	OriginID               string
	DestinationType        string
	DestinationID          string
	TenderJobScheduleShift string
	DriverDay              string
	OriginAt               string
	OriginNotes            string
	DestinationAt          string
	DestinationNotes       string
	SubmittedMileage       string
	SubmittedMinutes       string
}

func newDoTripsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trip",
		Long: `Update a trip.

Optional:
  --origin-type         Origin type
  --origin-id           Origin ID
  --destination-type    Destination type
  --destination-id      Destination ID
  --tender-job-schedule-shift  Tender job schedule shift ID
  --driver-day                 Driver day ID
  --origin-at                  Origin time (ISO 8601)
  --origin-notes               Origin notes
  --destination-at             Destination time (ISO 8601)
  --destination-notes          Destination notes
  --submitted-mileage          Submitted mileage
  --submitted-minutes          Submitted minutes`,
		Example: `  # Update origin time
  xbe do trips update 123 --origin-at "2024-01-15T08:00:00Z"

  # Update destination
  xbe do trips update 123 --destination-type job-sites --destination-id 456

  # Update submitted mileage
  xbe do trips update 123 --submitted-mileage "25.5"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTripsUpdate,
	}
	initDoTripsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTripsCmd.AddCommand(newDoTripsUpdateCmd())
}

func initDoTripsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("origin-type", "", "Origin type")
	cmd.Flags().String("origin-id", "", "Origin ID")
	cmd.Flags().String("destination-type", "", "Destination type")
	cmd.Flags().String("destination-id", "", "Destination ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("driver-day", "", "Driver day ID")
	cmd.Flags().String("origin-at", "", "Origin time (ISO 8601)")
	cmd.Flags().String("origin-notes", "", "Origin notes")
	cmd.Flags().String("destination-at", "", "Destination time (ISO 8601)")
	cmd.Flags().String("destination-notes", "", "Destination notes")
	cmd.Flags().String("submitted-mileage", "", "Submitted mileage")
	cmd.Flags().String("submitted-minutes", "", "Submitted minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTripsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTripsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("origin-at") {
		attributes["origin-at"] = opts.OriginAt
	}
	if cmd.Flags().Changed("origin-notes") {
		attributes["origin-notes"] = opts.OriginNotes
	}
	if cmd.Flags().Changed("destination-at") {
		attributes["destination-at"] = opts.DestinationAt
	}
	if cmd.Flags().Changed("destination-notes") {
		attributes["destination-notes"] = opts.DestinationNotes
	}
	if cmd.Flags().Changed("submitted-mileage") {
		attributes["submitted-mileage"] = opts.SubmittedMileage
	}
	if cmd.Flags().Changed("submitted-minutes") {
		attributes["submitted-minutes"] = opts.SubmittedMinutes
	}

	if cmd.Flags().Changed("origin-type") && cmd.Flags().Changed("origin-id") {
		relationships["origin"] = map[string]any{
			"data": map[string]any{
				"type": opts.OriginType,
				"id":   opts.OriginID,
			},
		}
	}
	if cmd.Flags().Changed("destination-type") && cmd.Flags().Changed("destination-id") {
		relationships["destination"] = map[string]any{
			"data": map[string]any{
				"type": opts.DestinationType,
				"id":   opts.DestinationID,
			},
		}
	}
	if cmd.Flags().Changed("tender-job-schedule-shift") {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if cmd.Flags().Changed("driver-day") {
		relationships["driver-day"] = map[string]any{
			"data": map[string]any{
				"type": "driver-days",
				"id":   opts.DriverDay,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "trips",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trips/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := tripRow{
			ID:               resp.Data.ID,
			OriginAt:         stringAttr(resp.Data.Attributes, "origin-at"),
			OriginNotes:      stringAttr(resp.Data.Attributes, "origin-notes"),
			DestinationAt:    stringAttr(resp.Data.Attributes, "destination-at"),
			DestinationNotes: stringAttr(resp.Data.Attributes, "destination-notes"),
			Mileage:          stringAttr(resp.Data.Attributes, "mileage"),
			Minutes:          stringAttr(resp.Data.Attributes, "minutes"),
		}
		if rel, ok := resp.Data.Relationships["origin"]; ok && rel.Data != nil {
			row.OriginType = rel.Data.Type
			row.OriginID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["destination"]; ok && rel.Data != nil {
			row.DestinationType = rel.Data.Type
			row.DestinationID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trip %s\n", resp.Data.ID)
	return nil
}

func parseDoTripsUpdateOptions(cmd *cobra.Command, args []string) (doTripsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	originType, _ := cmd.Flags().GetString("origin-type")
	originID, _ := cmd.Flags().GetString("origin-id")
	destinationType, _ := cmd.Flags().GetString("destination-type")
	destinationID, _ := cmd.Flags().GetString("destination-id")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	originAt, _ := cmd.Flags().GetString("origin-at")
	originNotes, _ := cmd.Flags().GetString("origin-notes")
	destinationAt, _ := cmd.Flags().GetString("destination-at")
	destinationNotes, _ := cmd.Flags().GetString("destination-notes")
	submittedMileage, _ := cmd.Flags().GetString("submitted-mileage")
	submittedMinutes, _ := cmd.Flags().GetString("submitted-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTripsUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		OriginType:             originType,
		OriginID:               originID,
		DestinationType:        destinationType,
		DestinationID:          destinationID,
		TenderJobScheduleShift: tenderJobScheduleShift,
		DriverDay:              driverDay,
		OriginAt:               originAt,
		OriginNotes:            originNotes,
		DestinationAt:          destinationAt,
		DestinationNotes:       destinationNotes,
		SubmittedMileage:       submittedMileage,
		SubmittedMinutes:       submittedMinutes,
	}, nil
}
