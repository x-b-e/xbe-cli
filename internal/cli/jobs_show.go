package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobDetails struct {
	ID                                      string `json:"id"`
	ExternalJobNumber                       string `json:"external_job_number,omitempty"`
	Notes                                   string `json:"notes,omitempty"`
	IsPrevailingWage                        bool   `json:"is_prevailing_wage"`
	RequiresCertifiedPayroll                bool   `json:"requires_certified_payroll"`
	PrevailingWageHourlyRate                string `json:"prevailing_wage_hourly_rate,omitempty"`
	DispatchInstructions                    string `json:"dispatch_instructions,omitempty"`
	LoadedMiles                             string `json:"loaded_miles,omitempty"`
	Tenderable                              bool   `json:"tenderable"`
	ExcludeTravelMinutesFromTotalHours      bool   `json:"exclude_travel_minutes_from_total_hours"`
	SkipMaterialTypeStartSiteTypeValidation bool   `json:"skip_material_type_start_site_type_validation"`
	JobProductionPlanReferenceData          any    `json:"job_production_plan_reference_data,omitempty"`

	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	JobSiteID    string `json:"job_site_id,omitempty"`
	JobSiteName  string `json:"job_site_name,omitempty"`

	StartSiteType string `json:"start_site_type,omitempty"`
	StartSiteID   string `json:"start_site_id,omitempty"`
	StartSiteName string `json:"start_site_name,omitempty"`

	JobProductionPlanID    string `json:"job_production_plan_id,omitempty"`
	JobProductionPlanLabel string `json:"job_production_plan,omitempty"`
	ForemanID              string `json:"foreman_id,omitempty"`
	ForemanName            string `json:"foreman_name,omitempty"`

	MaterialTypeIDs                           []string `json:"material_type_ids,omitempty"`
	TrailerClassificationIDs                  []string `json:"trailer_classification_ids,omitempty"`
	MaterialSiteIDs                           []string `json:"material_site_ids,omitempty"`
	ServiceTypeUnitOfMeasureIDs               []string `json:"service_type_unit_of_measure_ids,omitempty"`
	JobScheduleShiftIDs                       []string `json:"job_schedule_shift_ids,omitempty"`
	JobProductionPlanTrailerClassificationIDs []string `json:"job_production_plan_trailer_classification_ids,omitempty"`
	JobProductionPlanMaterialTypeIDs          []string `json:"job_production_plan_material_type_ids,omitempty"`
	CustomerTenderIDs                         []string `json:"customer_tender_ids,omitempty"`
	BrokerTenderIDs                           []string `json:"broker_tender_ids,omitempty"`
	CertificationRequirementIDs               []string `json:"certification_requirement_ids,omitempty"`
	ExternalIdentificationIDs                 []string `json:"external_identification_ids,omitempty"`
}

func newJobsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job details",
		Long: `Show the full details of a job.

Output Fields:
  ID
  External Job Number
  Notes
  Prevailing Wage flags and hourly rate
  Dispatch Instructions
  Loaded Miles
  Tenderable / Exclude Travel Minutes
  Skip Material Type Start Site Validation
  Job Production Plan Reference Data
  Customer / Job Site / Start Site
  Job Production Plan / Foreman
  Material Types / Trailer Classifications
  Material Sites / Service Type Unit Of Measures
  Job Schedule Shifts
  Job Production Plan trailer/material types
  Customer/Broker Tenders
  Certification Requirements
  External Identifications

Arguments:
  <id>    The job ID (required). Use the list command to find IDs.`,
		Example: `  # Show a job
  xbe view jobs show 123

  # Show as JSON
  xbe view jobs show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobsShow,
	}
	initJobsShowFlags(cmd)
	return cmd
}

func init() {
	jobsCmd.AddCommand(newJobsShowCmd())
}

func initJobsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobsShowOptions(cmd)
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
		return fmt.Errorf("job id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[jobs]", strings.Join([]string{
		"external-job-number",
		"notes",
		"is-prevailing-wage",
		"requires-certified-payroll",
		"prevailing-wage-hourly-rate",
		"dispatch-instructions",
		"loaded-miles",
		"tenderable",
		"exclude-travel-minutes-from-total-hours",
		"skip-material-type-start-site-type-validation",
		"job-production-plan-reference-data",
		"customer",
		"job-site",
		"start-site",
		"job-production-plan",
		"foreman",
		"material-types",
		"trailer-classifications",
		"material-sites",
		"service-type-unit-of-measures",
		"job-schedule-shifts",
		"job-production-plan-trailer-classifications",
		"job-production-plan-material-types",
		"customer-tenders",
		"broker-tenders",
		"certification-requirements",
		"external-identifications",
	}, ","))
	query.Set("include", "customer,job-site,start-site,job-production-plan,foreman,material-types,trailer-classifications,material-sites,service-type-unit-of-measures")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[trailer-classifications]", "name,abbreviation")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[service-type-unit-of-measures]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/jobs/"+id, query)
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

	details := buildJobDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobDetails(cmd, details)
}

func parseJobsShowOptions(cmd *cobra.Command) (jobsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobsShowOptions{}, err
	}

	return jobsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobDetails(resp jsonAPISingleResponse) jobDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := jobDetails{
		ID:                                      resp.Data.ID,
		ExternalJobNumber:                       stringAttr(attrs, "external-job-number"),
		Notes:                                   stringAttr(attrs, "notes"),
		IsPrevailingWage:                        boolAttr(attrs, "is-prevailing-wage"),
		RequiresCertifiedPayroll:                boolAttr(attrs, "requires-certified-payroll"),
		PrevailingWageHourlyRate:                stringAttr(attrs, "prevailing-wage-hourly-rate"),
		DispatchInstructions:                    stringAttr(attrs, "dispatch-instructions"),
		LoadedMiles:                             stringAttr(attrs, "loaded-miles"),
		Tenderable:                              boolAttr(attrs, "tenderable"),
		ExcludeTravelMinutesFromTotalHours:      boolAttr(attrs, "exclude-travel-minutes-from-total-hours"),
		SkipMaterialTypeStartSiteTypeValidation: boolAttr(attrs, "skip-material-type-start-site-type-validation"),
		JobProductionPlanReferenceData:          anyAttr(attrs, "job-production-plan-reference-data"),
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerName = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
		}
	}
	if rel, ok := resp.Data.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobSiteName = strings.TrimSpace(stringAttr(site.Attributes, "name"))
		}
	}
	if rel, ok := resp.Data.Relationships["start-site"]; ok && rel.Data != nil {
		details.StartSiteType = rel.Data.Type
		details.StartSiteID = rel.Data.ID
		if startSite, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.StartSiteName = strings.TrimSpace(firstNonEmpty(
				stringAttr(startSite.Attributes, "name"),
				stringAttr(startSite.Attributes, "company-name"),
			))
		}
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobProductionPlanLabel = jobProductionPlanLabel(plan.Attributes)
		}
	}
	if rel, ok := resp.Data.Relationships["foreman"]; ok && rel.Data != nil {
		details.ForemanID = rel.Data.ID
		if foreman, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ForemanName = strings.TrimSpace(stringAttr(foreman.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["material-types"]; ok {
		details.MaterialTypeIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["trailer-classifications"]; ok {
		details.TrailerClassificationIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["material-sites"]; ok {
		details.MaterialSiteIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["service-type-unit-of-measures"]; ok {
		details.ServiceTypeUnitOfMeasureIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["job-schedule-shifts"]; ok {
		details.JobScheduleShiftIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["job-production-plan-trailer-classifications"]; ok {
		details.JobProductionPlanTrailerClassificationIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["job-production-plan-material-types"]; ok {
		details.JobProductionPlanMaterialTypeIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["customer-tenders"]; ok {
		details.CustomerTenderIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["broker-tenders"]; ok {
		details.BrokerTenderIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["certification-requirements"]; ok {
		details.CertificationRequirementIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDList(rel)
	}

	return details
}

func renderJobDetails(cmd *cobra.Command, details jobDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalJobNumber != "" {
		fmt.Fprintf(out, "External Job Number: %s\n", details.ExternalJobNumber)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	fmt.Fprintf(out, "Is Prevailing Wage: %v\n", details.IsPrevailingWage)
	fmt.Fprintf(out, "Requires Certified Payroll: %v\n", details.RequiresCertifiedPayroll)
	if details.PrevailingWageHourlyRate != "" {
		fmt.Fprintf(out, "Prevailing Wage Hourly Rate: %s\n", details.PrevailingWageHourlyRate)
	}
	if details.DispatchInstructions != "" {
		fmt.Fprintf(out, "Dispatch Instructions: %s\n", details.DispatchInstructions)
	}
	if details.LoadedMiles != "" {
		fmt.Fprintf(out, "Loaded Miles: %s\n", details.LoadedMiles)
	}
	fmt.Fprintf(out, "Tenderable: %v\n", details.Tenderable)
	fmt.Fprintf(out, "Exclude Travel Minutes From Total Hours: %v\n", details.ExcludeTravelMinutesFromTotalHours)
	fmt.Fprintf(out, "Skip Material Type Start Site Validation: %v\n", details.SkipMaterialTypeStartSiteTypeValidation)

	if details.JobProductionPlanReferenceData != nil {
		fmt.Fprintln(out, "Job Production Plan Reference Data:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatAnyJSON(details.JobProductionPlanReferenceData))
	}

	if details.CustomerID != "" {
		if details.CustomerName != "" {
			fmt.Fprintf(out, "Customer: %s (ID: %s)\n", details.CustomerName, details.CustomerID)
		} else {
			fmt.Fprintf(out, "Customer: %s\n", details.CustomerID)
		}
	}
	if details.JobSiteID != "" {
		if details.JobSiteName != "" {
			fmt.Fprintf(out, "Job Site: %s (ID: %s)\n", details.JobSiteName, details.JobSiteID)
		} else {
			fmt.Fprintf(out, "Job Site: %s\n", details.JobSiteID)
		}
	}
	if details.StartSiteID != "" {
		label := details.StartSiteType + "/" + details.StartSiteID
		if details.StartSiteName != "" {
			label = details.StartSiteName + " (" + label + ")"
		}
		fmt.Fprintf(out, "Start Site: %s\n", label)
	}
	if details.JobProductionPlanID != "" {
		if details.JobProductionPlanLabel != "" {
			fmt.Fprintf(out, "Job Production Plan: %s (ID: %s)\n", details.JobProductionPlanLabel, details.JobProductionPlanID)
		} else {
			fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
		}
	}
	if details.ForemanID != "" {
		if details.ForemanName != "" {
			fmt.Fprintf(out, "Foreman: %s (ID: %s)\n", details.ForemanName, details.ForemanID)
		} else {
			fmt.Fprintf(out, "Foreman: %s\n", details.ForemanID)
		}
	}

	if len(details.MaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "Material Types: %s\n", strings.Join(details.MaterialTypeIDs, ", "))
	}
	if len(details.TrailerClassificationIDs) > 0 {
		fmt.Fprintf(out, "Trailer Classifications: %s\n", strings.Join(details.TrailerClassificationIDs, ", "))
	}
	if len(details.MaterialSiteIDs) > 0 {
		fmt.Fprintf(out, "Material Sites: %s\n", strings.Join(details.MaterialSiteIDs, ", "))
	}
	if len(details.ServiceTypeUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Service Type Unit Of Measures: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureIDs, ", "))
	}
	if len(details.JobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Job Schedule Shifts: %s\n", strings.Join(details.JobScheduleShiftIDs, ", "))
	}
	if len(details.JobProductionPlanTrailerClassificationIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan Trailer Classifications: %s\n", strings.Join(details.JobProductionPlanTrailerClassificationIDs, ", "))
	}
	if len(details.JobProductionPlanMaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan Material Types: %s\n", strings.Join(details.JobProductionPlanMaterialTypeIDs, ", "))
	}
	if len(details.CustomerTenderIDs) > 0 {
		fmt.Fprintf(out, "Customer Tenders: %s\n", strings.Join(details.CustomerTenderIDs, ", "))
	}
	if len(details.BrokerTenderIDs) > 0 {
		fmt.Fprintf(out, "Broker Tenders: %s\n", strings.Join(details.BrokerTenderIDs, ", "))
	}
	if len(details.CertificationRequirementIDs) > 0 {
		fmt.Fprintf(out, "Certification Requirements: %s\n", strings.Join(details.CertificationRequirementIDs, ", "))
	}
	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintf(out, "External Identifications: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}

	return nil
}
