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

type projectTrailerClassificationsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	Project                    string
	TrailerClassification      string
	ProjectLaborClassification string
}

type projectTrailerClassificationRow struct {
	ID                           string `json:"id"`
	ProjectID                    string `json:"project_id,omitempty"`
	TrailerClassificationID      string `json:"trailer_classification_id,omitempty"`
	ProjectLaborClassificationID string `json:"project_labor_classification_id,omitempty"`
}

func newProjectTrailerClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project trailer classifications",
		Long: `List project trailer classifications with filtering and pagination.

Project trailer classifications associate trailer classifications with projects.
They can optionally link to project labor classifications.

Output Columns:
  ID                           Project trailer classification identifier
  PROJECT                      Project ID
  TRAILER CLASSIFICATION       Trailer classification ID
  PROJECT LABOR CLASSIFICATION Project labor classification ID (if set)

Filters:
  --project                      Filter by project ID
  --trailer-classification        Filter by trailer classification ID
  --project-labor-classification  Filter by project labor classification ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project trailer classifications
  xbe view project-trailer-classifications list

  # Filter by project
  xbe view project-trailer-classifications list --project 123

  # Filter by trailer classification
  xbe view project-trailer-classifications list --trailer-classification 456

  # Filter by project labor classification
  xbe view project-trailer-classifications list --project-labor-classification 789

  # Output as JSON
  xbe view project-trailer-classifications list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTrailerClassificationsList,
	}
	initProjectTrailerClassificationsListFlags(cmd)
	return cmd
}

func init() {
	projectTrailerClassificationsCmd.AddCommand(newProjectTrailerClassificationsListCmd())
}

func initProjectTrailerClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID")
	cmd.Flags().String("project-labor-classification", "", "Filter by project labor classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTrailerClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTrailerClassificationsListOptions(cmd)
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
	query.Set("fields[project-trailer-classifications]", "project,trailer-classification,project-labor-classification")

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
	setFilterIfPresent(query, "filter[trailer-classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[project-labor-classification]", opts.ProjectLaborClassification)

	body, _, err := client.Get(cmd.Context(), "/v1/project-trailer-classifications", query)
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

	rows := buildProjectTrailerClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTrailerClassificationsTable(cmd, rows)
}

func parseProjectTrailerClassificationsListOptions(cmd *cobra.Command) (projectTrailerClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	projectLaborClassification, _ := cmd.Flags().GetString("project-labor-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTrailerClassificationsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		Project:                    project,
		TrailerClassification:      trailerClassification,
		ProjectLaborClassification: projectLaborClassification,
	}, nil
}

func buildProjectTrailerClassificationRows(resp jsonAPIResponse) []projectTrailerClassificationRow {
	rows := make([]projectTrailerClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTrailerClassificationRow{
			ID:                           resource.ID,
			ProjectID:                    relationshipIDFromMap(resource.Relationships, "project"),
			TrailerClassificationID:      relationshipIDFromMap(resource.Relationships, "trailer-classification"),
			ProjectLaborClassificationID: relationshipIDFromMap(resource.Relationships, "project-labor-classification"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectTrailerClassificationsTable(cmd *cobra.Command, rows []projectTrailerClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project trailer classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tTRAILER_CLASSIFICATION\tPROJECT_LABOR_CLASSIFICATION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectID,
			row.TrailerClassificationID,
			row.ProjectLaborClassificationID,
		)
	}
	return writer.Flush()
}
