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

type projectsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Name         string
	Status       string
	CreatedAtMin string
	CreatedAtMax string
}

func newProjectsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Long: `List projects with filtering and pagination.

Returns a list of projects matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Date Filters:
  Use ISO 8601 format (YYYY-MM-DD) for date filters.`,
		Example: `  # List projects
  xbe view projects list

  # Search by project name
  xbe view projects list --name "Highway"

  # Filter by status
  xbe view projects list --status active

  # Filter by date range
  xbe view projects list --created-at-min 2024-01-01 --created-at-max 2024-12-31

  # Paginate results
  xbe view projects list --limit 20 --offset 40

  # Output as JSON
  xbe view projects list --json`,
		RunE: runProjectsList,
	}
	initProjectsListFlags(cmd)
	return cmd
}

func init() {
	projectsCmd.AddCommand(newProjectsListCmd())
}

func initProjectsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by project name (partial match)")
	cmd.Flags().String("status", "", "Filter by project status")
	cmd.Flags().String("created-at-min", "", "Filter by minimum created date (YYYY-MM-DD)")
	cmd.Flags().String("created-at-max", "", "Filter by maximum created date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[projects]", "name,status,created-at")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/projects", query)
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

	if opts.JSON {
		rows := buildProjectRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectsList(cmd, resp)
}

func parseProjectsListOptions(cmd *cobra.Command) (projectsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return projectsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return projectsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return projectsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return projectsListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return projectsListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return projectsListOptions{}, err
	}
	createdAtMin, err := cmd.Flags().GetString("created-at-min")
	if err != nil {
		return projectsListOptions{}, err
	}
	createdAtMax, err := cmd.Flags().GetString("created-at-max")
	if err != nil {
		return projectsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return projectsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return projectsListOptions{}, err
	}

	return projectsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Name:         name,
		Status:       status,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
	}, nil
}

type projectRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func buildProjectRows(resp jsonAPIResponse) []projectRow {
	rows := make([]projectRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, projectRow{
			ID:        resource.ID,
			Name:      strings.TrimSpace(stringAttr(resource.Attributes, "name")),
			Status:    strings.TrimSpace(stringAttr(resource.Attributes, "status")),
			CreatedAt: strings.TrimSpace(stringAttr(resource.Attributes, "created-at")),
		})
	}
	return rows
}

func renderProjectsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildProjectRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No projects found.")
		return nil
	}

	const nameMax = 50
	const statusMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSTATUS\tCREATED")
	for _, row := range rows {
		createdAt := row.CreatedAt
		if len(createdAt) > 10 {
			createdAt = createdAt[:10]
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Status, statusMax),
			createdAt,
		)
	}
	return writer.Flush()
}
