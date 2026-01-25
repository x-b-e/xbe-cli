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

type projectPhasesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Project string
}

type projectPhaseRow struct {
	ID               string `json:"id"`
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Sequence         string `json:"sequence,omitempty"`
	SequencePosition int    `json:"sequence_position,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`
}

func newProjectPhasesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phases",
		Long: `List project phases.

Output Columns:
  ID            Project phase identifier
  NAME          Phase name
  SEQUENCE      Sequence value
  PROJECT       Project ID

Filters:
  --project     Filter by project ID`,
		Example: `  # List all project phases
  xbe view project-phases list

  # Filter by project
  xbe view project-phases list --project 123

  # Output as JSON
  xbe view project-phases list --json`,
		RunE: runProjectPhasesList,
	}
	initProjectPhasesListFlags(cmd)
	return cmd
}

func init() {
	projectPhasesCmd.AddCommand(newProjectPhasesListCmd())
}

func initProjectPhasesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhasesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhasesListOptions(cmd)
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
	query.Set("include", "project")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phases", query)
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

	rows := buildProjectPhaseRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhasesTable(cmd, rows)
}

func parseProjectPhasesListOptions(cmd *cobra.Command) (projectPhasesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	project, _ := cmd.Flags().GetString("project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhasesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Project: project,
	}, nil
}

func buildProjectPhaseRows(resp jsonAPIResponse) []projectPhaseRow {
	rows := make([]projectPhaseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectPhaseRow{
			ID:               resource.ID,
			Name:             stringAttr(resource.Attributes, "name"),
			Description:      stringAttr(resource.Attributes, "description"),
			Sequence:         stringAttr(resource.Attributes, "sequence"),
			SequencePosition: intAttr(resource.Attributes, "sequence-position"),
		}

		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectPhasesTable(cmd *cobra.Command, rows []projectPhaseRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phases found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSEQUENCE\tPROJECT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.Name,
			row.Sequence,
			row.ProjectID,
		)
	}
	return writer.Flush()
}
