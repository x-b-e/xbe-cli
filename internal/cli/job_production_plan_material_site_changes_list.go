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

type jobProductionPlanMaterialSiteChangesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type jobProductionPlanMaterialSiteChangeRow struct {
	ID                     string `json:"id"`
	JobProductionPlanID    string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan      string `json:"job_production_plan,omitempty"`
	OldMaterialSiteID      string `json:"old_material_site_id,omitempty"`
	OldMaterialSite        string `json:"old_material_site,omitempty"`
	NewMaterialSiteID      string `json:"new_material_site_id,omitempty"`
	NewMaterialSite        string `json:"new_material_site,omitempty"`
	OldMaterialTypeID      string `json:"old_material_type_id,omitempty"`
	OldMaterialType        string `json:"old_material_type,omitempty"`
	NewMaterialTypeID      string `json:"new_material_type_id,omitempty"`
	NewMaterialType        string `json:"new_material_type,omitempty"`
	NewMaterialMixDesignID string `json:"new_material_mix_design_id,omitempty"`
	NewMaterialMixDesign   string `json:"new_material_mix_design,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
	CreatedBy              string `json:"created_by,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
}

func newJobProductionPlanMaterialSiteChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan material site changes",
		Long: `List job production plan material site changes.

Output Columns:
  ID                   Material site change identifier
  JOB PRODUCTION PLAN  Job production plan name/number
  OLD SITE             Old material site
  NEW SITE             New material site
  OLD TYPE             Old material type (if applicable)
  NEW TYPE             New material type (if applicable)
  MIX DESIGN           New material mix design (if applicable)
  CREATED BY           User who created the change
  CREATED              Created timestamp

Filters:
  --created-at-min      Filter by created-at on/after (ISO 8601)
  --created-at-max      Filter by created-at on/before (ISO 8601)
  --updated-at-min      Filter by updated-at on/after (ISO 8601)
  --updated-at-max      Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material site changes
  xbe view job-production-plan-material-site-changes list

  # Filter by created-at range
  xbe view job-production-plan-material-site-changes list --created-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view job-production-plan-material-site-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanMaterialSiteChangesList,
	}
	initJobProductionPlanMaterialSiteChangesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialSiteChangesCmd.AddCommand(newJobProductionPlanMaterialSiteChangesListCmd())
}

func initJobProductionPlanMaterialSiteChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialSiteChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanMaterialSiteChangesListOptions(cmd)
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
	query.Set("fields[job-production-plan-material-site-changes]", "created-at,job-production-plan,old-material-site,new-material-site,old-material-type,new-material-type,new-material-mix-design,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "display-name,name,fully-qualified-name")
	query.Set("fields[material-mix-designs]", "description,mix")
	query.Set("fields[users]", "name")
	query.Set("include", "job-production-plan,old-material-site,new-material-site,old-material-type,new-material-type,new-material-mix-design,created-by")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-site-changes", query)
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

	rows := buildJobProductionPlanMaterialSiteChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanMaterialSiteChangesTable(cmd, rows)
}

func parseJobProductionPlanMaterialSiteChangesListOptions(cmd *cobra.Command) (jobProductionPlanMaterialSiteChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanMaterialSiteChangesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildJobProductionPlanMaterialSiteChangeRows(resp jsonAPIResponse) []jobProductionPlanMaterialSiteChangeRow {
	rows := make([]jobProductionPlanMaterialSiteChangeRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanMaterialSiteChangeRow(resource, included))
	}
	return rows
}

func buildJobProductionPlanMaterialSiteChangeRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanMaterialSiteChangeRow {
	attrs := resource.Attributes
	row := jobProductionPlanMaterialSiteChangeRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	jppType := ""
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}
	oldSiteType := ""
	if rel, ok := resource.Relationships["old-material-site"]; ok && rel.Data != nil {
		row.OldMaterialSiteID = rel.Data.ID
		oldSiteType = rel.Data.Type
	}
	newSiteType := ""
	if rel, ok := resource.Relationships["new-material-site"]; ok && rel.Data != nil {
		row.NewMaterialSiteID = rel.Data.ID
		newSiteType = rel.Data.Type
	}
	oldTypeType := ""
	if rel, ok := resource.Relationships["old-material-type"]; ok && rel.Data != nil {
		row.OldMaterialTypeID = rel.Data.ID
		oldTypeType = rel.Data.Type
	}
	newTypeType := ""
	if rel, ok := resource.Relationships["new-material-type"]; ok && rel.Data != nil {
		row.NewMaterialTypeID = rel.Data.ID
		newTypeType = rel.Data.Type
	}
	mixDesignType := ""
	if rel, ok := resource.Relationships["new-material-mix-design"]; ok && rel.Data != nil {
		row.NewMaterialMixDesignID = rel.Data.ID
		mixDesignType = rel.Data.Type
	}
	createdByType := ""
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
	}

	if len(included) == 0 {
		return row
	}

	if row.JobProductionPlanID != "" && jppType != "" {
		if jpp, ok := included[resourceKey(jppType, row.JobProductionPlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(jpp.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(jpp.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				row.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				row.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if row.OldMaterialSiteID != "" && oldSiteType != "" {
		if site, ok := included[resourceKey(oldSiteType, row.OldMaterialSiteID)]; ok {
			row.OldMaterialSite = strings.TrimSpace(stringAttr(site.Attributes, "name"))
		}
	}

	if row.NewMaterialSiteID != "" && newSiteType != "" {
		if site, ok := included[resourceKey(newSiteType, row.NewMaterialSiteID)]; ok {
			row.NewMaterialSite = strings.TrimSpace(stringAttr(site.Attributes, "name"))
		}
	}

	if row.OldMaterialTypeID != "" && oldTypeType != "" {
		if materialType, ok := included[resourceKey(oldTypeType, row.OldMaterialTypeID)]; ok {
			row.OldMaterialType = firstNonEmpty(
				strings.TrimSpace(stringAttr(materialType.Attributes, "display-name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "fully-qualified-name")),
			)
		}
	}

	if row.NewMaterialTypeID != "" && newTypeType != "" {
		if materialType, ok := included[resourceKey(newTypeType, row.NewMaterialTypeID)]; ok {
			row.NewMaterialType = firstNonEmpty(
				strings.TrimSpace(stringAttr(materialType.Attributes, "display-name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "fully-qualified-name")),
			)
		}
	}

	if row.NewMaterialMixDesignID != "" && mixDesignType != "" {
		if mixDesign, ok := included[resourceKey(mixDesignType, row.NewMaterialMixDesignID)]; ok {
			row.NewMaterialMixDesign = firstNonEmpty(
				strings.TrimSpace(stringAttr(mixDesign.Attributes, "description")),
				strings.TrimSpace(stringAttr(mixDesign.Attributes, "mix")),
			)
		}
	}

	if row.CreatedByID != "" && createdByType != "" {
		if user, ok := included[resourceKey(createdByType, row.CreatedByID)]; ok {
			row.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return row
}

func buildJobProductionPlanMaterialSiteChangeRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanMaterialSiteChangeRow {
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	return buildJobProductionPlanMaterialSiteChangeRow(resp.Data, included)
}

func renderJobProductionPlanMaterialSiteChangesTable(cmd *cobra.Command, rows []jobProductionPlanMaterialSiteChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan material site changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PRODUCTION PLAN\tOLD SITE\tNEW SITE\tOLD TYPE\tNEW TYPE\tMIX DESIGN\tCREATED BY\tCREATED")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.JobProductionPlan, 32),
			truncateString(firstNonEmpty(row.OldMaterialSite, row.OldMaterialSiteID), 24),
			truncateString(firstNonEmpty(row.NewMaterialSite, row.NewMaterialSiteID), 24),
			truncateString(firstNonEmpty(row.OldMaterialType, row.OldMaterialTypeID), 24),
			truncateString(firstNonEmpty(row.NewMaterialType, row.NewMaterialTypeID), 24),
			truncateString(firstNonEmpty(row.NewMaterialMixDesign, row.NewMaterialMixDesignID), 24),
			truncateString(firstNonEmpty(row.CreatedBy, row.CreatedByID), 20),
			truncateString(row.CreatedAt, 20),
		)
	}
	return writer.Flush()
}
