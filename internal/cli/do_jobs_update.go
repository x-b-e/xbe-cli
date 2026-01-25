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

type doJobsUpdateOptions struct {
	BaseURL                                 string
	Token                                   string
	JSON                                    bool
	ID                                      string
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
	ExternalJobNumber                       string
}

func newDoJobsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job",
		Long: `Update a job.

Optional flags:
  --customer               Customer ID
  --job-site               Job site ID
  --start-site-type         Start site type (job-sites or material-sites)
  --start-site              Start site ID (requires --start-site-type)
  --job-production-plan     Job production plan ID
  --foreman                 Foreman user ID
  --material-types          Material type IDs (comma-separated)
  --trailer-classifications Trailer classification IDs (comma-separated)
  --material-sites           Material site IDs (comma-separated)
  --service-type-unit-of-measures Service type unit of measure IDs (comma-separated)
  --notes                   Notes
  --dispatch-instructions    Dispatch instructions
  --loaded-miles             Loaded miles
  --is-prevailing-wage        Prevailing wage job
  --requires-certified-payroll Requires certified payroll
  --prevailing-wage-hourly-rate Prevailing wage hourly rate
  --skip-material-type-start-site-type-validation Skip start site type validation
  --validate-job-schedule-shifts Validate job schedule shifts
  --external-job-number       External job number (update-only)`,
		Example: `  # Update notes
  xbe do jobs update 123 --notes "Updated notes"

  # Update relationships
  xbe do jobs update 123 --material-types 1,2 --trailer-classifications 3

  # Set prevailing wage details
  xbe do jobs update 123 --is-prevailing-wage --requires-certified-payroll --prevailing-wage-hourly-rate 50`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobsUpdate,
	}
	initDoJobsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobsCmd.AddCommand(newDoJobsUpdateCmd())
}

func initDoJobsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("job-site", "", "Job site ID")
	cmd.Flags().String("start-site-type", "", "Start site type (job-sites or material-sites)")
	cmd.Flags().String("start-site", "", "Start site ID (requires --start-site-type)")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("foreman", "", "Foreman user ID")
	cmd.Flags().String("material-types", "", "Material type IDs (comma-separated)")
	cmd.Flags().String("trailer-classifications", "", "Trailer classification IDs (comma-separated)")
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
	cmd.Flags().String("external-job-number", "", "External job number (update-only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobsUpdateOptions(cmd, args)
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

	if (opts.StartSiteType != "" && opts.StartSiteID == "") || (opts.StartSiteType == "" && opts.StartSiteID != "") {
		return fmt.Errorf("--start-site and --start-site-type must be set together")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("is-prevailing-wage") {
		attributes["is-prevailing-wage"] = opts.IsPrevailingWage
	}
	if cmd.Flags().Changed("requires-certified-payroll") {
		attributes["requires-certified-payroll"] = opts.RequiresCertifiedPayroll
	}
	if cmd.Flags().Changed("prevailing-wage-hourly-rate") {
		attributes["prevailing-wage-hourly-rate"] = opts.PrevailingWageHourlyRate
	}
	if cmd.Flags().Changed("dispatch-instructions") {
		attributes["dispatch-instructions"] = opts.DispatchInstructions
	}
	if cmd.Flags().Changed("loaded-miles") {
		attributes["loaded-miles"] = opts.LoadedMiles
	}
	if cmd.Flags().Changed("skip-material-type-start-site-type-validation") {
		attributes["skip-material-type-start-site-type-validation"] = opts.SkipMaterialTypeStartSiteTypeValidation
	}
	if cmd.Flags().Changed("validate-job-schedule-shifts") {
		attributes["validate-job-schedule-shifts"] = opts.ValidateJobScheduleShifts
	}
	if cmd.Flags().Changed("external-job-number") {
		attributes["external-job-number"] = opts.ExternalJobNumber
	}

	if cmd.Flags().Changed("customer") {
		if opts.CustomerID == "" {
			relationships["customer"] = map[string]any{"data": nil}
		} else {
			relationships["customer"] = map[string]any{
				"data": map[string]any{
					"type": "customers",
					"id":   opts.CustomerID,
				},
			}
		}
	}
	if cmd.Flags().Changed("job-site") {
		if opts.JobSiteID == "" {
			relationships["job-site"] = map[string]any{"data": nil}
		} else {
			relationships["job-site"] = map[string]any{
				"data": map[string]any{
					"type": "job-sites",
					"id":   opts.JobSiteID,
				},
			}
		}
	}
	if cmd.Flags().Changed("start-site") || cmd.Flags().Changed("start-site-type") {
		if opts.StartSiteID == "" {
			relationships["start-site"] = map[string]any{"data": nil}
		} else {
			relationships["start-site"] = map[string]any{
				"data": map[string]any{
					"type": opts.StartSiteType,
					"id":   opts.StartSiteID,
				},
			}
		}
	}
	if cmd.Flags().Changed("job-production-plan") {
		if opts.JobProductionPlanID == "" {
			relationships["job-production-plan"] = map[string]any{"data": nil}
		} else {
			relationships["job-production-plan"] = map[string]any{
				"data": map[string]any{
					"type": "job-production-plans",
					"id":   opts.JobProductionPlanID,
				},
			}
		}
	}
	if cmd.Flags().Changed("foreman") {
		if opts.ForemanID == "" {
			relationships["foreman"] = map[string]any{"data": nil}
		} else {
			relationships["foreman"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.ForemanID,
				},
			}
		}
	}
	if cmd.Flags().Changed("material-types") {
		if opts.MaterialTypes == "" {
			relationships["material-types"] = map[string]any{"data": []any{}}
		} else {
			ids := parseCommaSeparatedIDs(opts.MaterialTypes)
			relationships["material-types"] = map[string]any{"data": buildRelationshipDataList(ids, "material-types")}
		}
	}
	if cmd.Flags().Changed("trailer-classifications") {
		if opts.TrailerClassifications == "" {
			relationships["trailer-classifications"] = map[string]any{"data": []any{}}
		} else {
			ids := parseCommaSeparatedIDs(opts.TrailerClassifications)
			relationships["trailer-classifications"] = map[string]any{"data": buildRelationshipDataList(ids, "trailer-classifications")}
		}
	}
	if cmd.Flags().Changed("material-sites") {
		if opts.MaterialSites == "" {
			relationships["material-sites"] = map[string]any{"data": []any{}}
		} else {
			ids := parseCommaSeparatedIDs(opts.MaterialSites)
			relationships["material-sites"] = map[string]any{"data": buildRelationshipDataList(ids, "material-sites")}
		}
	}
	if cmd.Flags().Changed("service-type-unit-of-measures") {
		if opts.ServiceTypeUnitOfMeasures == "" {
			relationships["service-type-unit-of-measures"] = map[string]any{"data": []any{}}
		} else {
			ids := parseCommaSeparatedIDs(opts.ServiceTypeUnitOfMeasures)
			relationships["service-type-unit-of-measures"] = map[string]any{"data": buildRelationshipDataList(ids, "service-type-unit-of-measures")}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		return fmt.Errorf("no attributes or relationships to update")
	}

	data := map[string]any{
		"type": "jobs",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/jobs/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job %s\n", row.ID)
	return nil
}

func parseDoJobsUpdateOptions(cmd *cobra.Command, args []string) (doJobsUpdateOptions, error) {
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
	externalJobNumber, _ := cmd.Flags().GetString("external-job-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobsUpdateOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		ID:                                      args[0],
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
		ExternalJobNumber:                       externalJobNumber,
	}, nil
}
