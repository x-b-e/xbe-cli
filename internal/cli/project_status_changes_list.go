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

type projectStatusChangesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Project string
	Status  string
}

type projectStatusChangeRow struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	Comment     string `json:"comment,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
}

func newProjectStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project status changes",
		Long: `List project status changes with filtering and pagination.

Project status changes record the status history of projects.

Output Columns:
  ID         Status change ID
  PROJECT    Project ID
  STATUS     Project status
  CHANGED AT Timestamp when the status changed
  CHANGED BY User ID who changed the status (if present)
  COMMENT    Change comment (truncated)

Filters:
  --project  Filter by project ID
  --status   Filter by project status

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project status changes
  xbe view project-status-changes list

  # Filter by project
  xbe view project-status-changes list --project 123

  # Filter by status
  xbe view project-status-changes list --status active

  # Output as JSON
  xbe view project-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectStatusChangesList,
	}
	initProjectStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	projectStatusChangesCmd.AddCommand(newProjectStatusChangesListCmd())
}

func initProjectStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("status", "", "Filter by project status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectStatusChangesListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-status-changes]", "status,changed-at,comment,project,changed-by")
	query.Set("include", "project,changed-by")

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
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/project-status-changes", query)
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

	rows := buildProjectStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectStatusChangesTable(cmd, rows)
}

func parseProjectStatusChangesListOptions(cmd *cobra.Command) (projectStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectStatusChangesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Project: project,
		Status:  status,
	}, nil
}

func buildProjectStatusChangeRows(resp jsonAPIResponse) []projectStatusChangeRow {
	rows := make([]projectStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectStatusChangeRow{
			ID:        resource.ID,
			Status:    stringAttr(attrs, "status"),
			ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
			Comment:   stringAttr(attrs, "comment"),
		}

		row.ProjectID = relationshipIDFromMap(resource.Relationships, "project")
		row.ChangedByID = relationshipIDFromMap(resource.Relationships, "changed-by")

		rows = append(rows, row)
	}
	return rows
}

func renderProjectStatusChangesTable(cmd *cobra.Command, rows []projectStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectID,
			row.Status,
			row.ChangedAt,
			row.ChangedByID,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
