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

type doJobsCreateOptions struct {
	BaseURL                                 string
	Token                                   string
	JSON                                    bool
	CustomerID                              string
	JobSiteID                               string
	StartSiteType                           string
	StartSiteID                             string
	JobProductionPlanID                     string
	ForemanID                               string
	MaterialTypes                           string
	TrailerClassifications                  string
	MaterialSites                           string
	ServiceTypeUnitOfMeasures               string
	Notes                                   string
	IsPrevailingWage                        bool
	RequiresCertifiedPayroll                bool
	PrevailingWageHourlyRate                string
	DispatchInstructions                    string
	LoadedMiles                             string
	SkipMaterialTypeStartSiteTypeValidation bool
	ValidateJobScheduleShifts               bool
}

func newDoJobsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job",
		Long: `Create a new job.

Required flags:
  --customer               Customer ID (required)
  --job-site               Job site ID (required)
  --material-types         Material type IDs (comma-separated, required)
  --trailer-classifications Trailer classification IDs (comma-separated, required)

Optional flags:
  --start-site-type         Start site type (job-sites or material-sites)
  --start-site              Start site ID (requires --start-site-type)
  --job-production-plan     Job production plan ID
  --foreman                 Foreman user ID
  --material-sites           Material site IDs (comma-separated)
  --service-type-unit-of-measures Service type unit of measure IDs (comma-separated)
  --notes                   Notes
  --dispatch-instructions    Dispatch instructions
  --loaded-miles             Loaded miles
  --is-prevailing-wage        Prevailing wage job
  --requires-certified-payroll Requires certified payroll
  --prevailing-wage-hourly-rate Prevailing wage hourly rate
  --skip-material-type-start-site-type-validation Skip start site type validation
  --validate-job-schedule-shifts Validate job schedule shifts`,
		Example: `  # Create a job
  xbe do jobs create \
    --customer 123 \
    --job-site 456 \
    --material-types 1,2 \
    --trailer-classifications 3

  # Create with prevailing wage requirements
  xbe do jobs create \
    --customer 123 \
    --job-site 456 \
    --material-types 1 \
    --trailer-classifications 3 \
    --is-prevailing-wage \
    --requires-certified-payroll \
    --prevailing-wage-hourly-rate 45`,
		Args: cobra.NoArgs,
		RunE: runDoJobsCreate,
	}
	initDoJobsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobsCmd.AddCommand(newDoJobsCreateCmd())
}

func initDoJobsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("job-site", "", "Job site ID (required)")
	cmd.Flags().String("start-site-type", "", "Start site type (job-sites or material-sites)")
	cmd.Flags().String("start-site", "", "Start site ID (requires --start-site-type)")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("foreman", "", "Foreman user ID")
	cmd.Flags().String("material-types", "", "Material type IDs (comma-separated, required)")
	cmd.Flags().String("trailer-classifications", "", "Trailer classification IDs (comma-separated, required)")
	cmd.Flags().String("material-sites", "", "Material site IDs (comma-separated)")
	cmd.Flags().String("service-type-unit-of-measures", "", "Service type unit of measure IDs (comma-separated)")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().Bool("is-prevailing-wage", false, "Prevailing wage job")
	cmd.Flags().Bool("requires-certified-payroll", false, "Requires certified payroll")
	cmd.Flags().String("prevailing-wage-hourly-rate", "", "Prevailing wage hourly rate")
	cmd.Flags().String("dispatch-instructions", "", "Dispatch instructions")
	cmd.Flags().String("loaded-miles", "", "Loaded miles")
	cmd.Flags().Bool("skip-material-type-start-site-type-validation", false, "Skip start site type validation")
	cmd.Flags().Bool("validate-job-schedule-shifts", false, "Validate job schedule shifts")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobsCreateOptions(cmd)
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

	if opts.CustomerID == "" {
		return fmt.Errorf("--customer is required")
	}
	if opts.JobSiteID == "" {
		return fmt.Errorf("--job-site is required")
	}
	if strings.TrimSpace(opts.MaterialTypes) == "" {
		return fmt.Errorf("--material-types is required")
	}
	if strings.TrimSpace(opts.TrailerClassifications) == "" {
		return fmt.Errorf("--trailer-classifications is required")
	}
	if (opts.StartSiteType != "" && opts.StartSiteID == "") || (opts.StartSiteType == "" && opts.StartSiteID != "") {
		return fmt.Errorf("--start-site and --start-site-type must be set together")
	}

	attributes := map[string]any{}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("is-prevailing-wage") {
		attributes["is-prevailing-wage"] = opts.IsPrevailingWage
	}
	if cmd.Flags().Changed("requires-certified-payroll") {
		attributes["requires-certified-payroll"] = opts.RequiresCertifiedPayroll
	}
	if opts.PrevailingWageHourlyRate != "" {
		attributes["prevailing-wage-hourly-rate"] = opts.PrevailingWageHourlyRate
	}
	if opts.DispatchInstructions != "" {
		attributes["dispatch-instructions"] = opts.DispatchInstructions
	}
	if opts.LoadedMiles != "" {
		attributes["loaded-miles"] = opts.LoadedMiles
	}
	if cmd.Flags().Changed("skip-material-type-start-site-type-validation") {
		attributes["skip-material-type-start-site-type-validation"] = opts.SkipMaterialTypeStartSiteTypeValidation
	}
	if cmd.Flags().Changed("validate-job-schedule-shifts") {
		attributes["validate-job-schedule-shifts"] = opts.ValidateJobScheduleShifts
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		},
		"job-site": map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   opts.JobSiteID,
			},
		},
	}

	if opts.StartSiteType != "" {
		relationships["start-site"] = map[string]any{
			"data": map[string]any{
				"type": opts.StartSiteType,
				"id":   opts.StartSiteID,
			},
		}
	}
	if opts.JobProductionPlanID != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		}
	}
	if opts.ForemanID != "" {
		relationships["foreman"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ForemanID,
			},
		}
	}

	materialTypeIDs := parseCommaSeparatedIDs(opts.MaterialTypes)
	trailerClassificationIDs := parseCommaSeparatedIDs(opts.TrailerClassifications)

	relationships["material-types"] = map[string]any{"data": buildRelationshipDataList(materialTypeIDs, "material-types")}
	relationships["trailer-classifications"] = map[string]any{"data": buildRelationshipDataList(trailerClassificationIDs, "trailer-classifications")}

	if opts.MaterialSites != "" {
		materialSiteIDs := parseCommaSeparatedIDs(opts.MaterialSites)
		relationships["material-sites"] = map[string]any{"data": buildRelationshipDataList(materialSiteIDs, "material-sites")}
	}
	if opts.ServiceTypeUnitOfMeasures != "" {
		stuomIDs := parseCommaSeparatedIDs(opts.ServiceTypeUnitOfMeasures)
		relationships["service-type-unit-of-measures"] = map[string]any{"data": buildRelationshipDataList(stuomIDs, "service-type-unit-of-measures")}
	}

	data := map[string]any{
		"type":          "jobs",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/jobs", jsonBody)
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

	row := jobRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job %s\n", row.ID)
	return nil
}

func parseDoJobsCreateOptions(cmd *cobra.Command) (doJobsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customerID, _ := cmd.Flags().GetString("customer")
	jobSiteID, _ := cmd.Flags().GetString("job-site")
	startSiteType, _ := cmd.Flags().GetString("start-site-type")
	startSiteID, _ := cmd.Flags().GetString("start-site")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	foremanID, _ := cmd.Flags().GetString("foreman")
	materialTypes, _ := cmd.Flags().GetString("material-types")
	trailerClassifications, _ := cmd.Flags().GetString("trailer-classifications")
	materialSites, _ := cmd.Flags().GetString("material-sites")
	serviceTypeUnitOfMeasures, _ := cmd.Flags().GetString("service-type-unit-of-measures")
	notes, _ := cmd.Flags().GetString("notes")
	isPrevailingWage, _ := cmd.Flags().GetBool("is-prevailing-wage")
	requiresCertifiedPayroll, _ := cmd.Flags().GetBool("requires-certified-payroll")
	prevailingWageHourlyRate, _ := cmd.Flags().GetString("prevailing-wage-hourly-rate")
	dispatchInstructions, _ := cmd.Flags().GetString("dispatch-instructions")
	loadedMiles, _ := cmd.Flags().GetString("loaded-miles")
	skipMaterialTypeStartSiteTypeValidation, _ := cmd.Flags().GetBool("skip-material-type-start-site-type-validation")
	validateJobScheduleShifts, _ := cmd.Flags().GetBool("validate-job-schedule-shifts")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobsCreateOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		CustomerID:                              customerID,
		JobSiteID:                               jobSiteID,
		StartSiteType:                           startSiteType,
		StartSiteID:                             startSiteID,
		JobProductionPlanID:                     jobProductionPlanID,
		ForemanID:                               foremanID,
		MaterialTypes:                           materialTypes,
		TrailerClassifications:                  trailerClassifications,
		MaterialSites:                           materialSites,
		ServiceTypeUnitOfMeasures:               serviceTypeUnitOfMeasures,
		Notes:                                   notes,
		IsPrevailingWage:                        isPrevailingWage,
		RequiresCertifiedPayroll:                requiresCertifiedPayroll,
		PrevailingWageHourlyRate:                prevailingWageHourlyRate,
		DispatchInstructions:                    dispatchInstructions,
		LoadedMiles:                             loadedMiles,
		SkipMaterialTypeStartSiteTypeValidation: skipMaterialTypeStartSiteTypeValidation,
		ValidateJobScheduleShifts:               validateJobScheduleShifts,
	}, nil
}
