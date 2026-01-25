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

type jobProductionPlanInspectableSummariesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanInspectableSummaryDetails struct {
	ID                                     string   `json:"id"`
	DeveloperID                            string   `json:"developer_id,omitempty"`
	CustomerCompanyName                    string   `json:"customer_company_name,omitempty"`
	PlannerName                            string   `json:"planner_name,omitempty"`
	JobName                                string   `json:"job_name,omitempty"`
	JobNumber                              string   `json:"job_number,omitempty"`
	StartOn                                string   `json:"start_on,omitempty"`
	StartTime                              string   `json:"start_time,omitempty"`
	JobSiteName                            string   `json:"job_site_name,omitempty"`
	JobSiteLatitude                        *float64 `json:"job_site_latitude,omitempty"`
	JobSiteLongitude                       *float64 `json:"job_site_longitude,omitempty"`
	JobSiteTimeZoneID                      string   `json:"job_site_time_zone_id,omitempty"`
	IsEticketingCycleTimeEnabled           bool     `json:"is_eticketing_cycle_time_enabled"`
	IsEticketingRawEnabled                 bool     `json:"is_eticketing_raw_enabled"`
	IsMaterialTransactionInspectionEnabled bool     `json:"is_material_transaction_inspection_enabled"`
	CurrentUserCanInspect                  bool     `json:"current_user_can_inspect"`
	CurrentUserCanShowPlan                 bool     `json:"current_user_can_show_plan"`
	DeveloperReferences                    any      `json:"developer_references,omitempty"`
	ProjectID                              string   `json:"project_id,omitempty"`
	ProjectName                            string   `json:"project_name,omitempty"`
	ProjectNumber                          string   `json:"project_number,omitempty"`
	MaterialTransactions                   any      `json:"material_transactions,omitempty"`
}

func newJobProductionPlanInspectableSummariesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <job-production-plan-id>",
		Short: "Show job production plan inspectable summary details",
		Long: `Show the full details of a job production plan inspectable summary.

Output Fields:
  ID
  Developer ID
  Customer Company Name
  Planner Name
  Job Name
  Job Number
  Start On
  Start Time
  Job Site Name
  Job Site Latitude
  Job Site Longitude
  Job Site Time Zone ID
  Is Eticketing Cycle Time Enabled
  Is Eticketing Raw Enabled
  Is Material Transaction Inspection Enabled
  Current User Can Inspect
  Current User Can Show Plan
  Project ID
  Project Name
  Project Number
  Developer References
  Material Transactions

Arguments:
  <job-production-plan-id>    The job production plan ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an inspectable summary
  xbe view job-production-plan-inspectable-summaries show 123

  # Get JSON output
  xbe view job-production-plan-inspectable-summaries show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanInspectableSummariesShow,
	}
	initJobProductionPlanInspectableSummariesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanInspectableSummariesCmd.AddCommand(newJobProductionPlanInspectableSummariesShowCmd())
}

func initJobProductionPlanInspectableSummariesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanInspectableSummariesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanInspectableSummariesShowOptions(cmd)
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
		return fmt.Errorf("job production plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-inspectable-summaries]", "developer-id,customer-company-name,planner-name,job-name,job-number,start-on,start-time,job-site-name,job-site-latitude,job-site-longitude,job-site-time-zone-id,is-eticketing-cycle-time-enabled,is-eticketing-raw-enabled,is-material-transaction-inspection-enabled,material-transactions,current-user-can-inspect,current-user-can-show-plan,developer-references,project-id,project-name,project-number")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-inspectable-summaries/"+id, query)
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

	details := buildJobProductionPlanInspectableSummaryDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanInspectableSummaryDetails(cmd, details)
}

func parseJobProductionPlanInspectableSummariesShowOptions(cmd *cobra.Command) (jobProductionPlanInspectableSummariesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanInspectableSummariesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanInspectableSummaryDetails(resp jsonAPISingleResponse) jobProductionPlanInspectableSummaryDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanInspectableSummaryDetails{
		ID:                                     resource.ID,
		DeveloperID:                            stringAttr(attrs, "developer-id"),
		CustomerCompanyName:                    stringAttr(attrs, "customer-company-name"),
		PlannerName:                            stringAttr(attrs, "planner-name"),
		JobName:                                stringAttr(attrs, "job-name"),
		JobNumber:                              stringAttr(attrs, "job-number"),
		StartOn:                                formatDate(stringAttr(attrs, "start-on")),
		StartTime:                              stringAttr(attrs, "start-time"),
		JobSiteName:                            stringAttr(attrs, "job-site-name"),
		JobSiteLatitude:                        floatAttrPointer(attrs, "job-site-latitude"),
		JobSiteLongitude:                       floatAttrPointer(attrs, "job-site-longitude"),
		JobSiteTimeZoneID:                      stringAttr(attrs, "job-site-time-zone-id"),
		IsEticketingCycleTimeEnabled:           boolAttr(attrs, "is-eticketing-cycle-time-enabled"),
		IsEticketingRawEnabled:                 boolAttr(attrs, "is-eticketing-raw-enabled"),
		IsMaterialTransactionInspectionEnabled: boolAttr(attrs, "is-material-transaction-inspection-enabled"),
		CurrentUserCanInspect:                  boolAttr(attrs, "current-user-can-inspect"),
		CurrentUserCanShowPlan:                 boolAttr(attrs, "current-user-can-show-plan"),
		ProjectID:                              stringAttr(attrs, "project-id"),
		ProjectName:                            stringAttr(attrs, "project-name"),
		ProjectNumber:                          stringAttr(attrs, "project-number"),
	}

	details.DeveloperReferences = attrs["developer-references"]
	details.MaterialTransactions = attrs["material-transactions"]

	return details
}

func renderJobProductionPlanInspectableSummaryDetails(cmd *cobra.Command, details jobProductionPlanInspectableSummaryDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DeveloperID != "" {
		fmt.Fprintf(out, "Developer ID: %s\n", details.DeveloperID)
	}
	if details.CustomerCompanyName != "" {
		fmt.Fprintf(out, "Customer Company Name: %s\n", details.CustomerCompanyName)
	}
	if details.PlannerName != "" {
		fmt.Fprintf(out, "Planner Name: %s\n", details.PlannerName)
	}
	if details.JobName != "" {
		fmt.Fprintf(out, "Job Name: %s\n", details.JobName)
	}
	if details.JobNumber != "" {
		fmt.Fprintf(out, "Job Number: %s\n", details.JobNumber)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.StartTime != "" {
		fmt.Fprintf(out, "Start Time: %s\n", details.StartTime)
	}
	if details.JobSiteName != "" {
		fmt.Fprintf(out, "Job Site Name: %s\n", details.JobSiteName)
	}
	if details.JobSiteLatitude != nil {
		fmt.Fprintf(out, "Job Site Latitude: %s\n", formatFloat(details.JobSiteLatitude, 6))
	}
	if details.JobSiteLongitude != nil {
		fmt.Fprintf(out, "Job Site Longitude: %s\n", formatFloat(details.JobSiteLongitude, 6))
	}
	if details.JobSiteTimeZoneID != "" {
		fmt.Fprintf(out, "Job Site Time Zone ID: %s\n", details.JobSiteTimeZoneID)
	}

	fmt.Fprintf(out, "Is Eticketing Cycle Time Enabled: %s\n", formatBool(details.IsEticketingCycleTimeEnabled))
	fmt.Fprintf(out, "Is Eticketing Raw Enabled: %s\n", formatBool(details.IsEticketingRawEnabled))
	fmt.Fprintf(out, "Is Material Transaction Inspection Enabled: %s\n", formatBool(details.IsMaterialTransactionInspectionEnabled))
	fmt.Fprintf(out, "Current User Can Inspect: %s\n", formatBool(details.CurrentUserCanInspect))
	fmt.Fprintf(out, "Current User Can Show Plan: %s\n", formatBool(details.CurrentUserCanShowPlan))

	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.ProjectName != "" {
		fmt.Fprintf(out, "Project Name: %s\n", details.ProjectName)
	}
	if details.ProjectNumber != "" {
		fmt.Fprintf(out, "Project Number: %s\n", details.ProjectNumber)
	}

	if details.DeveloperReferences != nil {
		fmt.Fprintln(out, "\nDeveloper References:")
		fmt.Fprintln(out, formatJSONBlock(details.DeveloperReferences, "  "))
	}
	if details.MaterialTransactions != nil {
		fmt.Fprintln(out, "\nMaterial Transactions:")
		fmt.Fprintln(out, formatJSONBlock(details.MaterialTransactions, "  "))
	}

	return nil
}
