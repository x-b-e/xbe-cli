package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanMaterialSitesListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	MaterialSite      string
	IsDefault         string
	MilesMin          string
	MilesMax          string
}

type jobProductionPlanMaterialSiteRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	MaterialSiteID      string `json:"material_site_id,omitempty"`
	MaterialSite        string `json:"material_site,omitempty"`
	IsDefault           bool   `json:"is_default"`
	Miles               string `json:"miles,omitempty"`
	DefaultTicketMaker  string `json:"default_ticket_maker,omitempty"`
}

func newJobProductionPlanMaterialSitesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan material sites",
		Long: `List job production plan material sites with filtering and pagination.

Output Columns:
  ID             Job production plan material site identifier
  JOB PLAN       Job production plan (job number or name)
  MATERIAL SITE  Material site name
  DEFAULT        Whether this is the default material site for the plan
  MILES          Planned miles between job site and material site
  TICKET MAKER   Default ticket maker (user, material_site)

Filters:
  --job-production-plan  Filter by job production plan ID
  --material-site        Filter by material site ID
  --is-default           Filter by default status (true/false)
  --miles-min            Filter by minimum miles
  --miles-max            Filter by maximum miles

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan material sites
  xbe view job-production-plan-material-sites list

  # Filter by job production plan
  xbe view job-production-plan-material-sites list --job-production-plan 123

  # Filter by material site
  xbe view job-production-plan-material-sites list --material-site 456

  # Filter by default status
  xbe view job-production-plan-material-sites list --is-default true

  # Output as JSON
  xbe view job-production-plan-material-sites list --json`,
		RunE: runJobProductionPlanMaterialSitesList,
	}
	initJobProductionPlanMaterialSitesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialSitesCmd.AddCommand(newJobProductionPlanMaterialSitesListCmd())
}

func initJobProductionPlanMaterialSitesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("is-default", "", "Filter by default status (true/false)")
	cmd.Flags().String("miles-min", "", "Filter by minimum miles")
	cmd.Flags().String("miles-max", "", "Filter by maximum miles")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialSitesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanMaterialSitesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-material-sites]", "job-production-plan,material-site,is-default,miles,default-ticket-maker")
	query.Set("include", "job-production-plan,material-site")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-sites]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[is-default]", opts.IsDefault)
	setFilterIfPresent(query, "filter[miles-min]", opts.MilesMin)
	setFilterIfPresent(query, "filter[miles-max]", opts.MilesMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-sites", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildJobProductionPlanMaterialSiteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanMaterialSitesTable(cmd, rows)
}

func parseJobProductionPlanMaterialSitesListOptions(cmd *cobra.Command) (jobProductionPlanMaterialSitesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	jobProductionPlan, err := cmd.Flags().GetString("job-production-plan")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	materialSite, err := cmd.Flags().GetString("material-site")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	isDefault, err := cmd.Flags().GetString("is-default")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	milesMin, err := cmd.Flags().GetString("miles-min")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	milesMax, err := cmd.Flags().GetString("miles-max")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlanMaterialSitesListOptions{}, err
	}

	return jobProductionPlanMaterialSitesListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		MaterialSite:      materialSite,
		IsDefault:         isDefault,
		MilesMin:          milesMin,
		MilesMax:          milesMax,
	}, nil
}

func buildJobProductionPlanMaterialSiteRows(resp jsonAPIResponse) []jobProductionPlanMaterialSiteRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanMaterialSiteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanMaterialSiteRow(resource, included))
	}
	return rows
}

func jobProductionPlanMaterialSiteRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanMaterialSiteRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildJobProductionPlanMaterialSiteRow(resp.Data, included)
}

func buildJobProductionPlanMaterialSiteRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanMaterialSiteRow {
	attrs := resource.Attributes
	row := jobProductionPlanMaterialSiteRow{
		ID:                 resource.ID,
		IsDefault:          boolAttr(attrs, "is-default"),
		Miles:              stringAttr(attrs, "miles"),
		DefaultTicketMaker: stringAttr(attrs, "default-ticket-maker"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.JobProductionPlan = firstNonEmpty(
				stringAttr(plan.Attributes, "job-number"),
				stringAttr(plan.Attributes, "job-name"),
			)
		}
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialSite = stringAttr(site.Attributes, "name")
		}
	}

	return row
}

func renderJobProductionPlanMaterialSitesTable(cmd *cobra.Command, rows []jobProductionPlanMaterialSiteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan material sites found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tMATERIAL SITE\tDEFAULT\tMILES\tTICKET MAKER")

	for _, row := range rows {
		jobPlan := row.JobProductionPlan
		if jobPlan == "" {
			jobPlan = row.JobProductionPlanID
		}
		materialSite := row.MaterialSite
		if materialSite == "" {
			materialSite = row.MaterialSiteID
		}
		miles := row.Miles
		if miles == "" {
			miles = "-"
		}
		ticketMaker := row.DefaultTicketMaker
		if ticketMaker == "" {
			ticketMaker = "-"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%t\t%s\t%s\n",
			row.ID,
			jobPlan,
			materialSite,
			row.IsDefault,
			miles,
			ticketMaker,
		)
	}

	return writer.Flush()
}
