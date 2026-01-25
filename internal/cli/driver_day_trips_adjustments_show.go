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

type driverDayTripsAdjustmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverDayTripsAdjustmentDetails struct {
	ID                     string   `json:"id"`
	Description            string   `json:"description,omitempty"`
	Status                 string   `json:"status,omitempty"`
	OrderedShiftIDs        []string `json:"ordered_shift_ids,omitempty"`
	OldTripsAttributes     any      `json:"old_trips_attributes,omitempty"`
	NewTripsAttributes     any      `json:"new_trips_attributes,omitempty"`
	DriverDayID            string   `json:"driver_day_id,omitempty"`
	TenderJobScheduleShift string   `json:"tender_job_schedule_shift_id,omitempty"`
	TruckerID              string   `json:"trucker_id,omitempty"`
	BrokerID               string   `json:"broker_id,omitempty"`
	CreatedByID            string   `json:"created_by_id,omitempty"`
}

func newDriverDayTripsAdjustmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver day trips adjustment details",
		Long: `Show the full details of a driver day trips adjustment.

Output Fields:
  ID
  Description
  Status
  Ordered Shift IDs
  Old Trips Attributes
  New Trips Attributes
  Driver Day
  Tender Job Schedule Shift
  Trucker
  Broker
  Created By

Arguments:
  <id>    The adjustment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an adjustment
  xbe view driver-day-trips-adjustments show 123

  # Get JSON output
  xbe view driver-day-trips-adjustments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverDayTripsAdjustmentsShow,
	}
	initDriverDayTripsAdjustmentsShowFlags(cmd)
	return cmd
}

func init() {
	driverDayTripsAdjustmentsCmd.AddCommand(newDriverDayTripsAdjustmentsShowCmd())
}

func initDriverDayTripsAdjustmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayTripsAdjustmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverDayTripsAdjustmentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("driver day trips adjustment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-trips-adjustments/"+id, nil)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildDriverDayTripsAdjustmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverDayTripsAdjustmentDetails(cmd, details)
}

func parseDriverDayTripsAdjustmentsShowOptions(cmd *cobra.Command) (driverDayTripsAdjustmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayTripsAdjustmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverDayTripsAdjustmentDetails(resp jsonAPISingleResponse) driverDayTripsAdjustmentDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := driverDayTripsAdjustmentDetails{
		ID:              resource.ID,
		Description:     stringAttr(attrs, "description"),
		Status:          stringAttr(attrs, "status"),
		OrderedShiftIDs: stringSliceAttr(attrs, "ordered-shift-ids"),
	}

	if value, ok := attrs["old-trips-attributes"]; ok {
		details.OldTripsAttributes = value
	}
	if value, ok := attrs["new-trips-attributes"]; ok {
		details.NewTripsAttributes = value
	}

	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		details.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderDriverDayTripsAdjustmentDetails(cmd *cobra.Command, details driverDayTripsAdjustmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if len(details.OrderedShiftIDs) > 0 {
		fmt.Fprintf(out, "Ordered Shift IDs: %s\n", strings.Join(details.OrderedShiftIDs, ", "))
	}
	if details.DriverDayID != "" {
		fmt.Fprintf(out, "Driver Day: %s\n", details.DriverDayID)
	}
	if details.TenderJobScheduleShift != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShift)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.OldTripsAttributes != nil {
		fmt.Fprintln(out, "Old Trips Attributes:")
		fmt.Fprintln(out, formatDriverDayTripsAdjustmentJSON(details.OldTripsAttributes))
	}
	if details.NewTripsAttributes != nil {
		fmt.Fprintln(out, "New Trips Attributes:")
		fmt.Fprintln(out, formatDriverDayTripsAdjustmentJSON(details.NewTripsAttributes))
	}

	return nil
}

func formatDriverDayTripsAdjustmentJSON(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
