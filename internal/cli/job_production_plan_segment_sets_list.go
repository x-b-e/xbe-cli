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

type jobProductionPlanSegmentSetsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobProductionPlanID string
}

type jobProductionPlanSegmentSetRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	Name                string `json:"name,omitempty"`
	IsDefault           bool   `json:"is_default,omitempty"`
	StartOffsetMinutes  int    `json:"start_offset_minutes,omitempty"`
}

func newJobProductionPlanSegmentSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan segment sets",
		Long: `List job production plan segment sets.

Output Columns:
  ID            Segment set identifier
  PLAN          Job production plan ID
  NAME          Segment set name
  DEFAULT       Whether the set is default
  START OFFSET  Start offset (minutes)

Filters:
  --job-production-plan  Filter by job production plan ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan segment sets
  xbe view job-production-plan-segment-sets list

  # Filter by job production plan
  xbe view job-production-plan-segment-sets list --job-production-plan 123

  # Output as JSON
  xbe view job-production-plan-segment-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanSegmentSetsList,
	}
	initJobProductionPlanSegmentSetsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSegmentSetsCmd.AddCommand(newJobProductionPlanSegmentSetsListCmd())
}

func initJobProductionPlanSegmentSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSegmentSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSegmentSetsListOptions(cmd)
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
	query.Set("fields[job-production-plan-segment-sets]", "name,is-default,start-offset-minutes")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlanID)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-segment-sets", query)
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

	rows := buildJobProductionPlanSegmentSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSegmentSetsTable(cmd, rows)
}

func parseJobProductionPlanSegmentSetsListOptions(cmd *cobra.Command) (jobProductionPlanSegmentSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSegmentSetsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}

func buildJobProductionPlanSegmentSetRows(resp jsonAPIResponse) []jobProductionPlanSegmentSetRow {
	rows := make([]jobProductionPlanSegmentSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanSegmentSetRow{
			ID:                 resource.ID,
			Name:               stringAttr(attrs, "name"),
			IsDefault:          boolAttr(attrs, "is-default"),
			StartOffsetMinutes: intAttr(attrs, "start-offset-minutes"),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildJobProductionPlanSegmentSetRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanSegmentSetRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobProductionPlanSegmentSetRow{
		ID:                 resource.ID,
		Name:               stringAttr(attrs, "name"),
		IsDefault:          boolAttr(attrs, "is-default"),
		StartOffsetMinutes: intAttr(attrs, "start-offset-minutes"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}

func renderJobProductionPlanSegmentSetsTable(cmd *cobra.Command, rows []jobProductionPlanSegmentSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan segment sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tNAME\tDEFAULT\tSTART OFFSET")
	for _, row := range rows {
		startOffset := strconv.Itoa(row.StartOffsetMinutes)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			truncateString(row.Name, 30),
			formatYesNo(row.IsDefault),
			startOffset,
		)
	}
	return writer.Flush()
}
