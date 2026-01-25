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

type projectPhaseDatesEstimatesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	ProjectPhase       string
	ProjectEstimateSet string
	CreatedBy          string
	Project            string
}

type projectPhaseDatesEstimateRow struct {
	ID                   string `json:"id"`
	ProjectPhaseID       string `json:"project_phase_id,omitempty"`
	ProjectEstimateSetID string `json:"project_estimate_set_id,omitempty"`
	CreatedByID          string `json:"created_by_id,omitempty"`
	StartDate            string `json:"start_date,omitempty"`
	EndDate              string `json:"end_date,omitempty"`
}

func newProjectPhaseDatesEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase dates estimates",
		Long: `List project phase date estimates.

Output Columns:
  ID            Date estimate identifier
  PHASE         Project phase ID
  ESTIMATE SET  Project estimate set ID
  START DATE    Estimated start date
  END DATE      Estimated end date
  CREATED BY    User ID who created the estimate

Filters:
  --project-phase        Filter by project phase ID
  --project-estimate-set Filter by project estimate set ID
  --created-by           Filter by creator user ID
  --project              Filter by project ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List date estimates
  xbe view project-phase-dates-estimates list

  # Filter by project phase
  xbe view project-phase-dates-estimates list --project-phase 123

  # Filter by estimate set
  xbe view project-phase-dates-estimates list --project-estimate-set 456

  # Output as JSON
  xbe view project-phase-dates-estimates list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseDatesEstimatesList,
	}
	initProjectPhaseDatesEstimatesListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseDatesEstimatesCmd.AddCommand(newProjectPhaseDatesEstimatesListCmd())
}

func initProjectPhaseDatesEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-phase", "", "Filter by project phase ID")
	cmd.Flags().String("project-estimate-set", "", "Filter by project estimate set ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseDatesEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseDatesEstimatesListOptions(cmd)
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
	query.Set("fields[project-phase-dates-estimates]", "start-date,end-date,project-phase,project-estimate-set,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-phase]", opts.ProjectPhase)
	setFilterIfPresent(query, "filter[project-estimate-set]", opts.ProjectEstimateSet)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[project]", opts.Project)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-dates-estimates", query)
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

	rows := buildProjectPhaseDatesEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseDatesEstimatesTable(cmd, rows)
}

func parseProjectPhaseDatesEstimatesListOptions(cmd *cobra.Command) (projectPhaseDatesEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectPhase, _ := cmd.Flags().GetString("project-phase")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	project, _ := cmd.Flags().GetString("project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseDatesEstimatesListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		ProjectPhase:       projectPhase,
		ProjectEstimateSet: projectEstimateSet,
		CreatedBy:          createdBy,
		Project:            project,
	}, nil
}

func buildProjectPhaseDatesEstimateRows(resp jsonAPIResponse) []projectPhaseDatesEstimateRow {
	rows := make([]projectPhaseDatesEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildProjectPhaseDatesEstimateRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func projectPhaseDatesEstimateRowFromSingle(resp jsonAPISingleResponse) projectPhaseDatesEstimateRow {
	return buildProjectPhaseDatesEstimateRow(resp.Data)
}

func buildProjectPhaseDatesEstimateRow(resource jsonAPIResource) projectPhaseDatesEstimateRow {
	row := projectPhaseDatesEstimateRow{
		ID:        resource.ID,
		StartDate: stringAttr(resource.Attributes, "start-date"),
		EndDate:   stringAttr(resource.Attributes, "end-date"),
	}

	if rel, ok := resource.Relationships["project-phase"]; ok && rel.Data != nil {
		row.ProjectPhaseID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		row.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderProjectPhaseDatesEstimatesTable(cmd *cobra.Command, rows []projectPhaseDatesEstimateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase dates estimates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPHASE\tESTIMATE SET\tSTART DATE\tEND DATE\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ProjectPhaseID, 20),
			truncateString(row.ProjectEstimateSetID, 20),
			truncateString(row.StartDate, 12),
			truncateString(row.EndDate, 12),
			truncateString(row.CreatedByID, 20),
		)
	}
	return writer.Flush()
}
