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

type jobProductionPlanMaterialSitesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanMaterialSiteDetails struct {
	ID                                        string   `json:"id"`
	JobProductionPlanID                       string   `json:"job_production_plan_id,omitempty"`
	JobProductionPlan                         string   `json:"job_production_plan,omitempty"`
	MaterialSiteID                            string   `json:"material_site_id,omitempty"`
	MaterialSite                              string   `json:"material_site,omitempty"`
	IsDefault                                 bool     `json:"is_default"`
	Miles                                     string   `json:"miles,omitempty"`
	CalculatedTravelMiles                     string   `json:"calculated_travel_miles,omitempty"`
	CalculatedTravelMinutes                   string   `json:"calculated_travel_minutes,omitempty"`
	DefaultTicketMaker                        string   `json:"default_ticket_maker,omitempty"`
	UserTicketMakerMaterialTypeIDs            []string `json:"user_ticket_maker_material_type_ids,omitempty"`
	HasUserScaleTickets                       bool     `json:"has_user_scale_tickets"`
	PlanRequiresSiteSpecificMaterialTypes     bool     `json:"plan_requires_site_specific_material_types"`
	PlanRequiresSupplierSpecificMaterialTypes bool     `json:"plan_requires_supplier_specific_material_types"`
}

func newJobProductionPlanMaterialSitesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan material site details",
		Long: `Show the full details of a job production plan material site.

Output Fields:
  ID                                         Resource identifier
  Job Production Plan                         Job production plan (job number or name)
  Material Site                               Material site name
  Is Default                                  Whether this is the default site
  Miles                                       Planned travel miles
  Calculated Travel Miles                     Calculated travel miles
  Calculated Travel Minutes                   Calculated travel minutes
  Default Ticket Maker                        Default ticket maker (user, material_site)
  User Ticket Maker Material Type IDs         Material types where user is ticket maker
  Has User Scale Tickets                      Whether users have scale tickets
  Plan Requires Site Specific Material Types  Plan-level setting
  Plan Requires Supplier Specific Material Types Plan-level setting

Arguments:
  <id>          The job production plan material site ID (required).`,
		Example: `  # Show details
  xbe view job-production-plan-material-sites show 123

  # Output as JSON
  xbe view job-production-plan-material-sites show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanMaterialSitesShow,
	}
	initJobProductionPlanMaterialSitesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialSitesCmd.AddCommand(newJobProductionPlanMaterialSitesShowCmd())
}

func initJobProductionPlanMaterialSitesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialSitesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanMaterialSitesShowOptions(cmd)
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
		return fmt.Errorf("job production plan material site id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-material-sites]", "job-production-plan,material-site,is-default,miles,default-ticket-maker,user-ticket-maker-material-type-ids,has-user-scale-tickets,plan-requires-site-specific-material-types,plan-requires-supplier-specific-material-types,calculated-travel-miles,calculated-travel-minutes")
	query.Set("include", "job-production-plan,material-site")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-sites]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-sites/"+id, query)
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

	details := buildJobProductionPlanMaterialSiteDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanMaterialSiteDetails(cmd, details)
}

func parseJobProductionPlanMaterialSitesShowOptions(cmd *cobra.Command) (jobProductionPlanMaterialSitesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlanMaterialSitesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlanMaterialSitesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlanMaterialSitesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlanMaterialSitesShowOptions{}, err
	}

	return jobProductionPlanMaterialSitesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanMaterialSiteDetails(resp jsonAPISingleResponse) jobProductionPlanMaterialSiteDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := jobProductionPlanMaterialSiteDetails{
		ID:                                    resp.Data.ID,
		IsDefault:                             boolAttr(attrs, "is-default"),
		Miles:                                 stringAttr(attrs, "miles"),
		CalculatedTravelMiles:                 stringAttr(attrs, "calculated-travel-miles"),
		CalculatedTravelMinutes:               stringAttr(attrs, "calculated-travel-minutes"),
		DefaultTicketMaker:                    stringAttr(attrs, "default-ticket-maker"),
		UserTicketMakerMaterialTypeIDs:        stringSliceAttr(attrs, "user-ticket-maker-material-type-ids"),
		HasUserScaleTickets:                   boolAttr(attrs, "has-user-scale-tickets"),
		PlanRequiresSiteSpecificMaterialTypes: boolAttr(attrs, "plan-requires-site-specific-material-types"),
		PlanRequiresSupplierSpecificMaterialTypes: boolAttr(attrs, "plan-requires-supplier-specific-material-types"),
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobProductionPlan = firstNonEmpty(
				stringAttr(plan.Attributes, "job-number"),
				stringAttr(plan.Attributes, "job-name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(site.Attributes, "name")
		}
	}

	return details
}

func renderJobProductionPlanMaterialSiteDetails(cmd *cobra.Command, details jobProductionPlanMaterialSiteDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" || details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", formatRelated(details.JobProductionPlan, details.JobProductionPlanID))
	}
	if details.MaterialSiteID != "" || details.MaterialSite != "" {
		fmt.Fprintf(out, "Material Site: %s\n", formatRelated(details.MaterialSite, details.MaterialSiteID))
	}
	fmt.Fprintf(out, "Is Default: %t\n", details.IsDefault)
	fmt.Fprintf(out, "Has User Scale Tickets: %t\n", details.HasUserScaleTickets)
	fmt.Fprintf(out, "Plan Requires Site Specific Material Types: %t\n", details.PlanRequiresSiteSpecificMaterialTypes)
	fmt.Fprintf(out, "Plan Requires Supplier Specific Material Types: %t\n", details.PlanRequiresSupplierSpecificMaterialTypes)

	if details.Miles != "" {
		fmt.Fprintf(out, "Miles: %s\n", details.Miles)
	}
	if details.CalculatedTravelMiles != "" {
		fmt.Fprintf(out, "Calculated Travel Miles: %s\n", details.CalculatedTravelMiles)
	}
	if details.CalculatedTravelMinutes != "" {
		fmt.Fprintf(out, "Calculated Travel Minutes: %s\n", details.CalculatedTravelMinutes)
	}
	if details.DefaultTicketMaker != "" {
		fmt.Fprintf(out, "Default Ticket Maker: %s\n", details.DefaultTicketMaker)
	}
	if len(details.UserTicketMakerMaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "User Ticket Maker Material Type IDs: %s\n", strings.Join(details.UserTicketMakerMaterialTypeIDs, ", "))
	}

	return nil
}
