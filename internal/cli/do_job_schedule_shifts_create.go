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

type doJobScheduleShiftsCreateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
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

func newDoJobScheduleShiftsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job schedule shift",
		Long: `Create a job schedule shift.

Required flags:
  --job        Job ID (required)
  --start-at   Shift start time (ISO 8601; required)
  --end-at     Shift end time (ISO 8601; required)

Optional attributes:
  --dispatch-instructions                 Dispatch instructions
  --is-planned-productive                 Planned productive shift (true/false)
  --cancelled-at                          Cancelled at time (ISO 8601)
  --suppress-automated-shift-feedback     Suppress automated shift feedback (true/false)
  --expected-material-transaction-count   Expected material transaction count
  --expected-material-transaction-tons    Expected material transaction tons
  --is-flexible                           Flexible shift (true/false)
  --start-at-min                          Earliest flexible start time (ISO 8601)
  --start-at-max                          Latest flexible start time (ISO 8601)
  --is-subsequent-shift-in-driver-day     Subsequent shift in driver day (true/false)
  --is-trucker-incident-creation-automated-explicit Explicit trucker incident automation override (true/false)

Optional relationships:
  --start-site-type        Start site type (job-sites or material-sites)
  --start-site             Start site ID (requires --start-site-type)
  --start-location         Job production plan location ID
  --trailer-classification Trailer classification ID
  --project-labor-classification Project labor classification ID`,
		Example: `  # Create a job schedule shift
  xbe do job-schedule-shifts create --job 123 \\
    --start-at 2025-01-01T08:00:00Z \\
    --end-at 2025-01-01T16:00:00Z

  # Create a flexible shift
  xbe do job-schedule-shifts create --job 123 \\
    --start-at 2025-01-01T08:00:00Z \\
    --end-at 2025-01-01T16:00:00Z \\
    --is-flexible true \\
    --start-at-min 2025-01-01T07:30:00Z \\
    --start-at-max 2025-01-01T08:30:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoJobScheduleShiftsCreate,
	}
	initDoJobScheduleShiftsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobScheduleShiftsCmd.AddCommand(newDoJobScheduleShiftsCreateCmd())
}

func initDoJobScheduleShiftsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job", "", "Job ID (required)")
	cmd.Flags().String("start-at", "", "Shift start time (ISO 8601; required)")
	cmd.Flags().String("end-at", "", "Shift end time (ISO 8601; required)")
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

	cmd.MarkFlagRequired("job")
	cmd.MarkFlagRequired("start-at")
	cmd.MarkFlagRequired("end-at")
}

func runDoJobScheduleShiftsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobScheduleShiftsCreateOptions(cmd)
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

	if opts.Job == "" {
		return fmt.Errorf("--job is required")
	}
	if opts.StartAt == "" {
		return fmt.Errorf("--start-at is required")
	}
	if opts.EndAt == "" {
		return fmt.Errorf("--end-at is required")
	}
	if (opts.StartSiteType != "" && opts.StartSite == "") || (opts.StartSiteType == "" && opts.StartSite != "") {
		return fmt.Errorf("--start-site and --start-site-type must be set together")
	}
	if opts.StartAtMin != "" || opts.StartAtMax != "" {
		if opts.IsFlexible == "" {
			return fmt.Errorf("--is-flexible must be set when using --start-at-min or --start-at-max")
		}
		isFlexible, err := parseJobScheduleShiftBool(opts.IsFlexible, "is-flexible")
		if err != nil {
			return err
		}
		if !isFlexible {
			return fmt.Errorf("--is-flexible must be true when using --start-at-min or --start-at-max")
		}
	}

	attributes := map[string]any{
		"start-at": opts.StartAt,
		"end-at":   opts.EndAt,
	}

	if opts.DispatchInstructions != "" {
		attributes["dispatch-instructions"] = opts.DispatchInstructions
	}
	if opts.IsPlannedProductive != "" {
		value, err := parseJobScheduleShiftBool(opts.IsPlannedProductive, "is-planned-productive")
		if err != nil {
			return err
		}
		attributes["is-planned-productive"] = value
	}
	if opts.CancelledAt != "" {
		attributes["cancelled-at"] = opts.CancelledAt
	}
	if opts.SuppressAutomatedShiftFeedback != "" {
		value, err := parseJobScheduleShiftBool(opts.SuppressAutomatedShiftFeedback, "suppress-automated-shift-feedback")
		if err != nil {
			return err
		}
		attributes["suppress-automated-shift-feedback"] = value
	}
	if opts.ExpectedMaterialTransactionCount != "" {
		attributes["expected-material-transaction-count"] = opts.ExpectedMaterialTransactionCount
	}
	if opts.ExpectedMaterialTransactionTons != "" {
		attributes["expected-material-transaction-tons"] = opts.ExpectedMaterialTransactionTons
	}
	if opts.IsFlexible != "" {
		value, err := parseJobScheduleShiftBool(opts.IsFlexible, "is-flexible")
		if err != nil {
			return err
		}
		attributes["is-flexible"] = value
	}
	if opts.StartAtMin != "" {
		attributes["start-at-min"] = opts.StartAtMin
	}
	if opts.StartAtMax != "" {
		attributes["start-at-max"] = opts.StartAtMax
	}
	if opts.IsSubsequentShiftInDriverDay != "" {
		value, err := parseJobScheduleShiftBool(opts.IsSubsequentShiftInDriverDay, "is-subsequent-shift-in-driver-day")
		if err != nil {
			return err
		}
		attributes["is-subsequent-shift-in-driver-day"] = value
	}
	if opts.IsTruckerIncidentCreationAutomatedExplicit != "" {
		value, err := parseJobScheduleShiftBool(opts.IsTruckerIncidentCreationAutomatedExplicit, "is-trucker-incident-creation-automated-explicit")
		if err != nil {
			return err
		}
		attributes["is-trucker-incident-creation-automated-explicit"] = value
	}

	relationships := map[string]any{
		"job": map[string]any{
			"data": map[string]any{
				"type": "jobs",
				"id":   opts.Job,
			},
		},
	}

	if opts.StartSiteType != "" && opts.StartSite != "" {
		relationships["start-site"] = map[string]any{
			"data": map[string]any{
				"type": opts.StartSiteType,
				"id":   opts.StartSite,
			},
		}
	}
	if opts.StartLocation != "" {
		relationships["start-location"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-locations",
				"id":   opts.StartLocation,
			},
		}
	}
	if opts.TrailerClassification != "" {
		relationships["trailer-classification"] = map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		}
	}
	if opts.ProjectLaborClassification != "" {
		relationships["project-labor-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-labor-classifications",
				"id":   opts.ProjectLaborClassification,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-schedule-shifts",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-schedule-shifts", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job schedule shift %s\n", row.ID)
	return nil
}

func parseDoJobScheduleShiftsCreateOptions(cmd *cobra.Command) (doJobScheduleShiftsCreateOptions, error) {
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

	return doJobScheduleShiftsCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
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

func jobScheduleShiftRowFromSingle(resp jsonAPISingleResponse) jobScheduleShiftRow {
	resource := resp.Data
	row := jobScheduleShiftRow{
		ID:        resource.ID,
		StartAt:   formatDateTime(stringAttr(resource.Attributes, "start-at")),
		EndAt:     formatDateTime(stringAttr(resource.Attributes, "end-at")),
		StartDate: stringAttr(resource.Attributes, "start-date"),
		IsManaged: boolAttr(resource.Attributes, "is-managed"),
		Cancelled: strings.TrimSpace(stringAttr(resource.Attributes, "cancelled-at")) != "",
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resource.Relationships["job"]; ok && rel.Data != nil {
		row.JobID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
		row.JobSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.JobSite = stringAttr(site.Attributes, "name")
		}
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Customer = stringAttr(customer.Attributes, "company-name")
		}
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Broker = stringAttr(broker.Attributes, "company-name")
		}
	}

	return row
}
