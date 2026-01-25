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

type projectProjectCostClassificationsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	Project                   string
	ProjectCostClassification string
	CreatedAtMin              string
	CreatedAtMax              string
	UpdatedAtMin              string
	UpdatedAtMax              string
}

type projectProjectCostClassificationRow struct {
	ID                          string `json:"id"`
	ProjectID                   string `json:"project_id,omitempty"`
	ProjectCostClassificationID string `json:"project_cost_classification_id,omitempty"`
	NameOverride                string `json:"name_override,omitempty"`
}

func newProjectProjectCostClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project project cost classifications",
		Long: `List project project cost classifications with filtering and pagination.

Project project cost classifications link a project to a project cost classification
and optionally override the classification name for that project.

Output Columns:
  ID                   Project project cost classification identifier
  PROJECT              Project ID
  COST CLASSIFICATION  Project cost classification ID
  NAME OVERRIDE        Optional project-specific name

Filters:
  --project                      Filter by project ID
  --project-cost-classification  Filter by project cost classification ID
  --created-at-min               Filter by created-at on/after (ISO 8601)
  --created-at-max               Filter by created-at on/before (ISO 8601)
  --updated-at-min               Filter by updated-at on/after (ISO 8601)
  --updated-at-max               Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project project cost classifications
  xbe view project-project-cost-classifications list

  # Filter by project
  xbe view project-project-cost-classifications list --project 123

  # Filter by project cost classification
  xbe view project-project-cost-classifications list --project-cost-classification 456

  # Output as JSON
  xbe view project-project-cost-classifications list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectProjectCostClassificationsList,
	}
	initProjectProjectCostClassificationsListFlags(cmd)
	return cmd
}

func init() {
	projectProjectCostClassificationsCmd.AddCommand(newProjectProjectCostClassificationsListCmd())
}

func initProjectProjectCostClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("project-cost-classification", "", "Filter by project cost classification ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectProjectCostClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectProjectCostClassificationsListOptions(cmd)
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
	query.Set("fields[project-project-cost-classifications]", "name-override,project,project-cost-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[project-cost-classification]", opts.ProjectCostClassification)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/project-project-cost-classifications", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildProjectProjectCostClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectProjectCostClassificationsTable(cmd, rows)
}

func parseProjectProjectCostClassificationsListOptions(cmd *cobra.Command) (projectProjectCostClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectProjectCostClassificationsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		Project:                   project,
		ProjectCostClassification: projectCostClassification,
		CreatedAtMin:              createdAtMin,
		CreatedAtMax:              createdAtMax,
		UpdatedAtMin:              updatedAtMin,
		UpdatedAtMax:              updatedAtMax,
	}, nil
}

func buildProjectProjectCostClassificationRows(resp jsonAPIResponse) []projectProjectCostClassificationRow {
	rows := make([]projectProjectCostClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectProjectCostClassificationRow{
			ID:           resource.ID,
			NameOverride: stringAttr(attrs, "name-override"),
		}
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-cost-classification"]; ok && rel.Data != nil {
			row.ProjectCostClassificationID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectProjectCostClassificationsTable(cmd *cobra.Command, rows []projectProjectCostClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project project cost classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tCOST CLASSIFICATION\tNAME OVERRIDE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectID,
			row.ProjectCostClassificationID,
			row.NameOverride,
		)
	}
	return writer.Flush()
}
