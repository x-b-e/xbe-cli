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

type projectEstimateSetsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	Project    string
	CreatedBy  string
	IsBid      string
	IsActual   string
	IsPossible string
}

type projectEstimateSetRow struct {
	ID                  string `json:"id"`
	Name                string `json:"name,omitempty"`
	ProjectID           string `json:"project_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	BackupEstimateSetID string `json:"backup_estimate_set_id,omitempty"`
	IsBid               bool   `json:"is_bid,omitempty"`
	IsActual            bool   `json:"is_actual,omitempty"`
	IsPossible          bool   `json:"is_possible,omitempty"`
}

func newProjectEstimateSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project estimate sets",
		Long: `List project estimate sets with filtering and pagination.

Output Columns:
  ID        Project estimate set identifier
  NAME      Estimate set name
  PROJECT   Project ID
  CREATED   Created-by user ID
  BACKUP    Backup estimate set ID
  BID       Bid estimate set flag
  ACTUAL    Actual estimate set flag
  POSSIBLE  Possible estimate set flag

Filters:
  --project     Filter by project ID
  --created-by  Filter by creator user ID
  --is-bid      Filter by bid flag (true/false)
  --is-actual   Filter by actual flag (true/false)
  --is-possible Filter by possible flag (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project estimate sets
  xbe view project-estimate-sets list

  # Filter by project
  xbe view project-estimate-sets list --project 123

  # Filter to bid estimate sets
  xbe view project-estimate-sets list --is-bid true

  # Output as JSON
  xbe view project-estimate-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectEstimateSetsList,
	}
	initProjectEstimateSetsListFlags(cmd)
	return cmd
}

func init() {
	projectEstimateSetsCmd.AddCommand(newProjectEstimateSetsListCmd())
}

func initProjectEstimateSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("is-bid", "", "Filter by bid flag (true/false)")
	cmd.Flags().String("is-actual", "", "Filter by actual flag (true/false)")
	cmd.Flags().String("is-possible", "", "Filter by possible flag (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectEstimateSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectEstimateSetsListOptions(cmd)
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
	query.Set("fields[project-estimate-sets]", "name,is-bid,is-actual,is-possible,project,created-by,backup-estimate-set")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[is_bid]", opts.IsBid)
	setFilterIfPresent(query, "filter[is_actual]", opts.IsActual)
	setFilterIfPresent(query, "filter[is_possible]", opts.IsPossible)

	body, _, err := client.Get(cmd.Context(), "/v1/project-estimate-sets", query)
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

	rows := buildProjectEstimateSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectEstimateSetsTable(cmd, rows)
}

func parseProjectEstimateSetsListOptions(cmd *cobra.Command) (projectEstimateSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	createdBy, _ := cmd.Flags().GetString("created-by")
	isBid, _ := cmd.Flags().GetString("is-bid")
	isActual, _ := cmd.Flags().GetString("is-actual")
	isPossible, _ := cmd.Flags().GetString("is-possible")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectEstimateSetsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		Project:    project,
		CreatedBy:  createdBy,
		IsBid:      isBid,
		IsActual:   isActual,
		IsPossible: isPossible,
	}, nil
}

func buildProjectEstimateSetRows(resp jsonAPIResponse) []projectEstimateSetRow {
	rows := make([]projectEstimateSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectEstimateSetRow(resource))
	}
	return rows
}

func buildProjectEstimateSetRow(resource jsonAPIResource) projectEstimateSetRow {
	row := projectEstimateSetRow{
		ID:         resource.ID,
		Name:       stringAttr(resource.Attributes, "name"),
		IsBid:      boolAttr(resource.Attributes, "is-bid"),
		IsActual:   boolAttr(resource.Attributes, "is-actual"),
		IsPossible: boolAttr(resource.Attributes, "is-possible"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["backup-estimate-set"]; ok && rel.Data != nil {
		row.BackupEstimateSetID = rel.Data.ID
	}

	return row
}

func renderProjectEstimateSetsTable(cmd *cobra.Command, rows []projectEstimateSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project estimate sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tPROJECT\tCREATED\tBACKUP\tBID\tACTUAL\tPOSSIBLE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			row.ProjectID,
			row.CreatedByID,
			row.BackupEstimateSetID,
			formatYesNo(row.IsBid),
			formatYesNo(row.IsActual),
			formatYesNo(row.IsPossible),
		)
	}
	return writer.Flush()
}
