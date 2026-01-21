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

type jobProductionPlanCancellationReasonTypesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Slug    string
}

func newJobProductionPlanCancellationReasonTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan cancellation reason types",
		Long: `List job production plan cancellation reason types with pagination.

Cancellation reason types define the reasons why a job production plan can be cancelled.

Output Columns:
  ID           Type identifier
  SLUG         URL-friendly identifier
  NAME         Display name
  DESCRIPTION  Description

Filters:
  --slug  Filter by slug`,
		Example: `  # List all cancellation reason types
  xbe view job-production-plan-cancellation-reason-types list

  # Filter by slug
  xbe view job-production-plan-cancellation-reason-types list --slug "weather"

  # Output as JSON
  xbe view job-production-plan-cancellation-reason-types list --json`,
		RunE: runJobProductionPlanCancellationReasonTypesList,
	}
	initJobProductionPlanCancellationReasonTypesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanCancellationReasonTypesCmd.AddCommand(newJobProductionPlanCancellationReasonTypesListCmd())
}

func initJobProductionPlanCancellationReasonTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("slug", "", "Filter by slug")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanCancellationReasonTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanCancellationReasonTypesListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[job-production-plan-cancellation-reason-types]", "slug,name,description")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[slug]", opts.Slug)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-cancellation-reason-types", query)
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

	rows := buildJobProductionPlanCancellationReasonTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanCancellationReasonTypesTable(cmd, rows)
}

func parseJobProductionPlanCancellationReasonTypesListOptions(cmd *cobra.Command) (jobProductionPlanCancellationReasonTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	slug, _ := cmd.Flags().GetString("slug")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanCancellationReasonTypesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Slug:    slug,
	}, nil
}

type jobProductionPlanCancellationReasonTypeRow struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func buildJobProductionPlanCancellationReasonTypeRows(resp jsonAPIResponse) []jobProductionPlanCancellationReasonTypeRow {
	rows := make([]jobProductionPlanCancellationReasonTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanCancellationReasonTypeRow{
			ID:          resource.ID,
			Slug:        stringAttr(resource.Attributes, "slug"),
			Name:        stringAttr(resource.Attributes, "name"),
			Description: stringAttr(resource.Attributes, "description"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanCancellationReasonTypesTable(cmd *cobra.Command, rows []jobProductionPlanCancellationReasonTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan cancellation reason types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSLUG\tNAME\tDESCRIPTION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Slug, 25),
			truncateString(row.Name, 30),
			truncateString(row.Description, 35),
		)
	}
	return writer.Flush()
}
