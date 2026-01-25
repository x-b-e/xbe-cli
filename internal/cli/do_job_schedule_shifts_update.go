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

type doJobScheduleShiftsUpdateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	ID                                         string
	Job                                        string
	StartAt                                    string
	EndAt                                      string
	DispatchInstructions                       string
	IsPlannedProductive                        string
	CancelledAt                                string
	SuppressAutomatedShiftFeedback             string
	ExpectedMaterialTransactionCount           string
	ExpectedMaterialTransactionTons            string
	IsFlexible                                 string
	StartAtMin                                 string
	StartAtMax                                 string
	IsSubsequentShiftInDriverDay               string
	IsTruckerIncidentCreationAutomatedExplicit string
	TrailerClassification                      string
	ProjectLaborClassification                 string
	StartSiteType                              string
	StartSite                                  string
	StartLocation                              string
}

func newDoJobScheduleShiftsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job schedule shift",
		Long: `Update a job schedule shift.

Optional attributes:
  --job                                  Job ID
  --start-at                             Shift start time (ISO 8601)
  --end-at                               Shift end time (ISO 8601)
  --dispatch-instructions                Dispatch instructions
  --is-planned-productive                Planned productive shift (true/false)
  --cancelled-at                         Cancelled at time (ISO 8601)
  --suppress-automated-shift-feedback    Suppress automated shift feedback (true/false)
  --expected-material-transaction-count  Expected material transaction count
  --expected-material-transaction-tons   Expected material transaction tons
  --is-flexible                          Flexible shift (true/false)
  --start-at-min                         Earliest flexible start time (ISO 8601)
  --start-at-max                         Latest flexible start time (ISO 8601)
  --is-subsequent-shift-in-driver-day    Subsequent shift in driver day (true/false)
  --is-trucker-incident-creation-automated-explicit Explicit trucker incident automation override (true/false)

Optional relationships:
  --start-site-type        Start site type (job-sites or material-sites)
  --start-site             Start site ID (requires --start-site-type)
  --start-location         Job production plan location ID
  --trailer-classification Trailer classification ID
  --project-labor-classification Project labor classification ID`,
		Example: `  # Update dispatch instructions
  xbe do job-schedule-shifts update 123 --dispatch-instructions "Gate 3"

  # Update flexible window
  xbe do job-schedule-shifts update 123 --is-flexible true \\
    --start-at-min 2025-01-01T07:30:00Z --start-at-max 2025-01-01T08:30:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobScheduleShiftsUpdate,
	}
	initDoJobScheduleShiftsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobScheduleShiftsCmd.AddCommand(newDoJobScheduleShiftsUpdateCmd())
}

func initDoJobScheduleShiftsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job", "", "Job ID")
	cmd.Flags().String("start-at", "", "Shift start time (ISO 8601)")
	cmd.Flags().String("end-at", "", "Shift end time (ISO 8601)")
	cmd.Flags().String("dispatch-instructions", "", "Dispatch instructions")
	cmd.Flags().String("is-planned-productive", "", "Planned productive shift (true/false)")
	cmd.Flags().String("cancelled-at", "", "Cancelled at time (ISO 8601)")
	cmd.Flags().String("suppress-automated-shift-feedback", "", "Suppress automated shift feedback (true/false)")
	cmd.Flags().String("expected-material-transaction-count", "", "Expected material transaction count")
	cmd.Flags().String("expected-material-transaction-tons", "", "Expected material transaction tons")
	cmd.Flags().String("is-flexible", "", "Flexible shift (true/false)")
	cmd.Flags().String("start-at-min", "", "Earliest flexible start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Latest flexible start time (ISO 8601)")
	cmd.Flags().String("is-subsequent-shift-in-driver-day", "", "Subsequent shift in driver day (true/false)")
	cmd.Flags().String("is-trucker-incident-creation-automated-explicit", "", "Explicit trucker incident automation override (true/false)")
	cmd.Flags().String("start-site-type", "", "Start site type (job-sites or material-sites)")
	cmd.Flags().String("start-site", "", "Start site ID (requires --start-site-type)")
	cmd.Flags().String("start-location", "", "Job production plan location ID")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().String("project-labor-classification", "", "Project labor classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobScheduleShiftsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobScheduleShiftsUpdateOptions(cmd, args)
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

	if (opts.StartSiteType != "" && opts.StartSite == "") || (opts.StartSiteType == "" && opts.StartSite != "") {
		return fmt.Errorf("--start-site and --start-site-type must be set together")
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("dispatch-instructions") {
		attributes["dispatch-instructions"] = opts.DispatchInstructions
	}
	if cmd.Flags().Changed("is-planned-productive") {
		value, err := parseJobScheduleShiftBool(opts.IsPlannedProductive, "is-planned-productive")
		if err != nil {
			return err
		}
		attributes["is-planned-productive"] = value
	}
	if cmd.Flags().Changed("cancelled-at") {
		attributes["cancelled-at"] = opts.CancelledAt
	}
	if cmd.Flags().Changed("suppress-automated-shift-feedback") {
		value, err := parseJobScheduleShiftBool(opts.SuppressAutomatedShiftFeedback, "suppress-automated-shift-feedback")
		if err != nil {
			return err
		}
		attributes["suppress-automated-shift-feedback"] = value
	}
	if cmd.Flags().Changed("expected-material-transaction-count") {
		attributes["expected-material-transaction-count"] = opts.ExpectedMaterialTransactionCount
	}
	if cmd.Flags().Changed("expected-material-transaction-tons") {
		attributes["expected-material-transaction-tons"] = opts.ExpectedMaterialTransactionTons
	}
	if cmd.Flags().Changed("is-flexible") {
		value, err := parseJobScheduleShiftBool(opts.IsFlexible, "is-flexible")
		if err != nil {
			return err
		}
		attributes["is-flexible"] = value
	}
	if cmd.Flags().Changed("start-at-min") {
		attributes["start-at-min"] = opts.StartAtMin
	}
	if cmd.Flags().Changed("start-at-max") {
		attributes["start-at-max"] = opts.StartAtMax
	}
	if cmd.Flags().Changed("is-subsequent-shift-in-driver-day") {
		value, err := parseJobScheduleShiftBool(opts.IsSubsequentShiftInDriverDay, "is-subsequent-shift-in-driver-day")
		if err != nil {
			return err
		}
		attributes["is-subsequent-shift-in-driver-day"] = value
	}
	if cmd.Flags().Changed("is-trucker-incident-creation-automated-explicit") {
		value, err := parseJobScheduleShiftBool(opts.IsTruckerIncidentCreationAutomatedExplicit, "is-trucker-incident-creation-automated-explicit")
		if err != nil {
			return err
		}
		attributes["is-trucker-incident-creation-automated-explicit"] = value
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("job") {
		relationships["job"] = map[string]any{
			"data": map[string]any{
				"type": "jobs",
				"id":   opts.Job,
			},
		}
	}
	if opts.StartSiteType != "" && opts.StartSite != "" {
		relationships["start-site"] = map[string]any{
			"data": map[string]any{
				"type": opts.StartSiteType,
				"id":   opts.StartSite,
			},
		}
	}
	if cmd.Flags().Changed("start-location") {
		relationships["start-location"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-locations",
				"id":   opts.StartLocation,
			},
		}
	}
	if cmd.Flags().Changed("trailer-classification") {
		relationships["trailer-classification"] = map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		}
	}
	if cmd.Flags().Changed("project-labor-classification") {
		relationships["project-labor-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-labor-classifications",
				"id":   opts.ProjectLaborClassification,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "job-schedule-shifts",
			"id":   opts.ID,
		},
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-schedule-shifts/"+opts.ID, jsonBody)
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

	row := jobScheduleShiftRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job schedule shift %s\n", row.ID)
	return nil
}

func parseDoJobScheduleShiftsUpdateOptions(cmd *cobra.Command, args []string) (doJobScheduleShiftsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	job, _ := cmd.Flags().GetString("job")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	dispatchInstructions, _ := cmd.Flags().GetString("dispatch-instructions")
	isPlannedProductive, _ := cmd.Flags().GetString("is-planned-productive")
	cancelledAt, _ := cmd.Flags().GetString("cancelled-at")
	suppressAutomatedShiftFeedback, _ := cmd.Flags().GetString("suppress-automated-shift-feedback")
	expectedMaterialTransactionCount, _ := cmd.Flags().GetString("expected-material-transaction-count")
	expectedMaterialTransactionTons, _ := cmd.Flags().GetString("expected-material-transaction-tons")
	isFlexible, _ := cmd.Flags().GetString("is-flexible")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	isSubsequentShiftInDriverDay, _ := cmd.Flags().GetString("is-subsequent-shift-in-driver-day")
	isTruckerIncidentCreationAutomatedExplicit, _ := cmd.Flags().GetString("is-trucker-incident-creation-automated-explicit")
	startSiteType, _ := cmd.Flags().GetString("start-site-type")
	startSite, _ := cmd.Flags().GetString("start-site")
	startLocation, _ := cmd.Flags().GetString("start-location")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	projectLaborClassification, _ := cmd.Flags().GetString("project-labor-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobScheduleShiftsUpdateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		ID:                               args[0],
		Job:                              job,
		StartAt:                          startAt,
		EndAt:                            endAt,
		DispatchInstructions:             dispatchInstructions,
		IsPlannedProductive:              isPlannedProductive,
		CancelledAt:                      cancelledAt,
		SuppressAutomatedShiftFeedback:   suppressAutomatedShiftFeedback,
		ExpectedMaterialTransactionCount: expectedMaterialTransactionCount,
		ExpectedMaterialTransactionTons:  expectedMaterialTransactionTons,
		IsFlexible:                       isFlexible,
		StartAtMin:                       startAtMin,
		StartAtMax:                       startAtMax,
		IsSubsequentShiftInDriverDay:     isSubsequentShiftInDriverDay,
		IsTruckerIncidentCreationAutomatedExplicit: isTruckerIncidentCreationAutomatedExplicit,
		StartSiteType:              startSiteType,
		StartSite:                  startSite,
		StartLocation:              startLocation,
		TrailerClassification:      trailerClassification,
		ProjectLaborClassification: projectLaborClassification,
	}, nil
}
