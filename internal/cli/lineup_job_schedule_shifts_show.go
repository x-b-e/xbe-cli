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

type lineupJobScheduleShiftsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupJobScheduleShiftDetails struct {
	ID                                  string   `json:"id"`
	LineupID                            string   `json:"lineup_id,omitempty"`
	JobScheduleShiftID                  string   `json:"job_schedule_shift_id,omitempty"`
	LineupDispatchShiftID               string   `json:"lineup_dispatch_shift_id,omitempty"`
	TruckerID                           string   `json:"trucker_id,omitempty"`
	DriverID                            string   `json:"driver_id,omitempty"`
	TrailerClassificationID             string   `json:"trailer_classification_id,omitempty"`
	TicketReportID                      string   `json:"ticket_report_id,omitempty"`
	TruckerAssignmentRecommendationIDs  []string `json:"trucker_assignment_recommendation_ids,omitempty"`
	TrailerClassificationEquivalentType string   `json:"trailer_classification_equivalent_type,omitempty"`
	IsBrokered                          bool     `json:"is_brokered"`
	IsReadyToDispatch                   bool     `json:"is_ready_to_dispatch"`
	ExcludeFromLineupScenarios          bool     `json:"exclude_from_lineup_scenarios"`
	TravelMinutes                       string   `json:"travel_minutes,omitempty"`
	LoadedTonsMax                       string   `json:"loaded_tons_max,omitempty"`
	ExplicitMaterialTransactionTonsMax  string   `json:"explicit_material_transaction_tons_max,omitempty"`
	NotifyDriverOnLateShiftAssignment   bool     `json:"notify_driver_on_late_shift_assignment"`
	IsExpectingTimeCard                 bool     `json:"is_expecting_time_card"`
	HasLineupDispatchShift              bool     `json:"has_lineup_dispatch_shift"`
	CreatedAt                           string   `json:"created_at,omitempty"`
	UpdatedAt                           string   `json:"updated_at,omitempty"`
}

func newLineupJobScheduleShiftsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup job schedule shift details",
		Long: `Show the full details of a lineup job schedule shift.

Output Fields:
  ID
  Lineup ID
  Job Schedule Shift ID
  Lineup Dispatch Shift ID
  Trucker ID
  Driver ID
  Trailer Classification ID
  Ticket Report ID
  Trucker Assignment Recommendation IDs
  Trailer Classification Equivalent Type
  Is Brokered
  Is Ready To Dispatch
  Exclude From Lineup Scenarios
  Travel Minutes
  Loaded Tons Max
  Explicit Material Transaction Tons Max
  Notify Driver On Late Shift Assignment
  Is Expecting Time Card
  Has Lineup Dispatch Shift
  Created At
  Updated At

Arguments:
  <id>    The lineup job schedule shift ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a lineup job schedule shift
  xbe view lineup-job-schedule-shifts show 123

  # JSON output
  xbe view lineup-job-schedule-shifts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupJobScheduleShiftsShow,
	}
	initLineupJobScheduleShiftsShowFlags(cmd)
	return cmd
}

func init() {
	lineupJobScheduleShiftsCmd.AddCommand(newLineupJobScheduleShiftsShowCmd())
}

func initLineupJobScheduleShiftsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupJobScheduleShiftsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupJobScheduleShiftsShowOptions(cmd)
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
		return fmt.Errorf("lineup job schedule shift id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-job-schedule-shifts/"+id, nil)
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

	details := buildLineupJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupJobScheduleShiftDetails(cmd, details)
}

func parseLineupJobScheduleShiftsShowOptions(cmd *cobra.Command) (lineupJobScheduleShiftsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupJobScheduleShiftsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupJobScheduleShiftDetails(resp jsonAPISingleResponse) lineupJobScheduleShiftDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := lineupJobScheduleShiftDetails{
		ID:                                  resource.ID,
		TrailerClassificationEquivalentType: stringAttr(attrs, "trailer-classification-equivalent-type"),
		IsBrokered:                          boolAttr(attrs, "is-brokered"),
		IsReadyToDispatch:                   boolAttr(attrs, "is-ready-to-dispatch"),
		ExcludeFromLineupScenarios:          boolAttr(attrs, "exclude-from-lineup-scenarios"),
		TravelMinutes:                       stringAttr(attrs, "travel-minutes"),
		LoadedTonsMax:                       stringAttr(attrs, "loaded-tons-max"),
		ExplicitMaterialTransactionTonsMax:  stringAttr(attrs, "explicit-material-transaction-tons-max"),
		NotifyDriverOnLateShiftAssignment:   boolAttr(attrs, "notify-driver-on-late-shift-assignment"),
		IsExpectingTimeCard:                 boolAttr(attrs, "is-expecting-time-card"),
		HasLineupDispatchShift:              boolAttr(attrs, "has-lineup-dispatch-shift"),
		CreatedAt:                           stringAttr(attrs, "created-at"),
		UpdatedAt:                           stringAttr(attrs, "updated-at"),
	}

	details.LineupID = relationshipIDFromMap(resource.Relationships, "lineup")
	details.JobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "job-schedule-shift")
	details.LineupDispatchShiftID = relationshipIDFromMap(resource.Relationships, "lineup-dispatch-shift")
	details.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
	details.DriverID = relationshipIDFromMap(resource.Relationships, "driver")
	details.TrailerClassificationID = relationshipIDFromMap(resource.Relationships, "trailer-classification")
	details.TicketReportID = relationshipIDFromMap(resource.Relationships, "ticket-report")
	details.TruckerAssignmentRecommendationIDs = relationshipIDsFromMap(resource.Relationships, "trucker-assignment-recommendations")

	return details
}

func renderLineupJobScheduleShiftDetails(cmd *cobra.Command, details lineupJobScheduleShiftDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupID != "" {
		fmt.Fprintf(out, "Lineup ID: %s\n", details.LineupID)
	}
	if details.JobScheduleShiftID != "" {
		fmt.Fprintf(out, "Job Schedule Shift ID: %s\n", details.JobScheduleShiftID)
	}
	if details.LineupDispatchShiftID != "" {
		fmt.Fprintf(out, "Lineup Dispatch Shift ID: %s\n", details.LineupDispatchShiftID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.TrailerClassificationID != "" {
		fmt.Fprintf(out, "Trailer Classification ID: %s\n", details.TrailerClassificationID)
	}
	if details.TicketReportID != "" {
		fmt.Fprintf(out, "Ticket Report ID: %s\n", details.TicketReportID)
	}
	if len(details.TruckerAssignmentRecommendationIDs) > 0 {
		fmt.Fprintf(out, "Trucker Assignment Recommendation IDs: %s\n", strings.Join(details.TruckerAssignmentRecommendationIDs, ", "))
	}
	if details.TrailerClassificationEquivalentType != "" {
		fmt.Fprintf(out, "Trailer Classification Equivalent Type: %s\n", details.TrailerClassificationEquivalentType)
	}
	fmt.Fprintf(out, "Is Brokered: %t\n", details.IsBrokered)
	fmt.Fprintf(out, "Is Ready To Dispatch: %t\n", details.IsReadyToDispatch)
	fmt.Fprintf(out, "Exclude From Lineup Scenarios: %t\n", details.ExcludeFromLineupScenarios)
	if details.TravelMinutes != "" {
		fmt.Fprintf(out, "Travel Minutes: %s\n", details.TravelMinutes)
	}
	if details.LoadedTonsMax != "" {
		fmt.Fprintf(out, "Loaded Tons Max: %s\n", details.LoadedTonsMax)
	}
	if details.ExplicitMaterialTransactionTonsMax != "" {
		fmt.Fprintf(out, "Explicit Material Transaction Tons Max: %s\n", details.ExplicitMaterialTransactionTonsMax)
	}
	fmt.Fprintf(out, "Notify Driver On Late Shift Assignment: %t\n", details.NotifyDriverOnLateShiftAssignment)
	fmt.Fprintf(out, "Is Expecting Time Card: %t\n", details.IsExpectingTimeCard)
	fmt.Fprintf(out, "Has Lineup Dispatch Shift: %t\n", details.HasLineupDispatchShift)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
