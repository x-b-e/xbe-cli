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

type jobProductionPlanInspectorsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobProductionPlanID string
	User                string
}

type jobProductionPlanInspectorRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
}

func newJobProductionPlanInspectorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan inspectors",
		Long: `List job production plan inspectors.

Output Columns:
  ID     Job production plan inspector identifier
  PLAN   Job production plan ID
  USER   User ID

Filters:
  --job-production-plan-id  Filter by job production plan ID
  --user                    Filter by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan inspectors
  xbe view job-production-plan-inspectors list

  # Filter by job production plan
  xbe view job-production-plan-inspectors list --job-production-plan-id 123

  # Filter by user
  xbe view job-production-plan-inspectors list --user 456

  # Output as JSON
  xbe view job-production-plan-inspectors list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanInspectorsList,
	}
	initJobProductionPlanInspectorsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanInspectorsCmd.AddCommand(newJobProductionPlanInspectorsListCmd())
}

func initJobProductionPlanInspectorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan-id", "", "Filter by job production plan ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanInspectorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanInspectorsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan-id]", opts.JobProductionPlanID)
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-inspectors", query)
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

	rows := buildJobProductionPlanInspectorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanInspectorsTable(cmd, rows)
}

func parseJobProductionPlanInspectorsListOptions(cmd *cobra.Command) (jobProductionPlanInspectorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan-id")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanInspectorsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobProductionPlanID: jobProductionPlanID,
		User:                user,
	}, nil
}

func buildJobProductionPlanInspectorRows(resp jsonAPIResponse) []jobProductionPlanInspectorRow {
	rows := make([]jobProductionPlanInspectorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanInspectorRow{
			ID:                  resource.ID,
			JobProductionPlanID: stringAttr(attrs, "job-production-plan-id"),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanInspectorsTable(cmd *cobra.Command, rows []jobProductionPlanInspectorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan inspectors found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tUSER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.UserID,
		)
	}
	return writer.Flush()
}
