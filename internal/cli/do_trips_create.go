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

type doTripsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
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

func newDoTripsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trip",
		Long: `Create a trip.

Required:
  --origin-type         Origin type (e.g., material-sites, job-sites, parking-sites)
  --origin-id           Origin ID
  --destination-type    Destination type
  --destination-id      Destination ID

Optional:
  --tender-job-schedule-shift  Tender job schedule shift ID
  --driver-day                 Driver day ID
  --origin-at                  Origin time (ISO 8601)
  --origin-notes               Origin notes
  --destination-at             Destination time (ISO 8601)
  --destination-notes          Destination notes
  --submitted-mileage          Submitted mileage
  --submitted-minutes          Submitted minutes`,
		Example: `  # Create a trip
  xbe do trips create --origin-type material-sites --origin-id 123 --destination-type job-sites --destination-id 456

  # Create with times
  xbe do trips create --origin-type material-sites --origin-id 123 --destination-type job-sites --destination-id 456 \
    --origin-at "2024-01-15T08:00:00Z" --destination-at "2024-01-15T09:00:00Z"

  # Create with driver day
  xbe do trips create --origin-type material-sites --origin-id 123 --destination-type job-sites --destination-id 456 \
    --driver-day 789`,
		RunE: runDoTripsCreate,
	}
	initDoTripsCreateFlags(cmd)
	return cmd
}

func init() {
	doTripsCmd.AddCommand(newDoTripsCreateCmd())
}

func initDoTripsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("origin-type", "", "Origin type (e.g., material-sites, job-sites, parking-sites)")
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

	_ = cmd.MarkFlagRequired("origin-type")
	_ = cmd.MarkFlagRequired("origin-id")
	_ = cmd.MarkFlagRequired("destination-type")
	_ = cmd.MarkFlagRequired("destination-id")
}

func runDoTripsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTripsCreateOptions(cmd)
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

	if opts.OriginAt != "" {
		attributes["origin-at"] = opts.OriginAt
	}
	if opts.OriginNotes != "" {
		attributes["origin-notes"] = opts.OriginNotes
	}
	if opts.DestinationAt != "" {
		attributes["destination-at"] = opts.DestinationAt
	}
	if opts.DestinationNotes != "" {
		attributes["destination-notes"] = opts.DestinationNotes
	}
	if opts.SubmittedMileage != "" {
		attributes["submitted-mileage"] = opts.SubmittedMileage
	}
	if opts.SubmittedMinutes != "" {
		attributes["submitted-minutes"] = opts.SubmittedMinutes
	}

	relationships := map[string]any{
		"origin": map[string]any{
			"data": map[string]any{
				"type": opts.OriginType,
				"id":   opts.OriginID,
			},
		},
		"destination": map[string]any{
			"data": map[string]any{
				"type": opts.DestinationType,
				"id":   opts.DestinationID,
			},
		},
	}

	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if opts.DriverDay != "" {
		relationships["driver-day"] = map[string]any{
			"data": map[string]any{
				"type": "driver-days",
				"id":   opts.DriverDay,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trips",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trips", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trip %s\n", resp.Data.ID)
	return nil
}

func parseDoTripsCreateOptions(cmd *cobra.Command) (doTripsCreateOptions, error) {
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

	return doTripsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
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
