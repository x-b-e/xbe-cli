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

type retainerEarningStatusesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Retainer     string
	CalculatedOn string
}

type retainerEarningStatusRow struct {
	ID           string `json:"id"`
	RetainerID   string `json:"retainer_id,omitempty"`
	CalculatedOn string `json:"calculated_on,omitempty"`
	Expected     string `json:"expected,omitempty"`
	Actual       string `json:"actual,omitempty"`
}

func newRetainerEarningStatusesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List retainer earning statuses",
		Long: `List retainer earning statuses with filtering and pagination.

Retainer earning statuses capture expected and actual earnings for a retainer on a calculated date.

Output Columns:
  ID           Status identifier
  RETAINER     Retainer ID
  CALCULATED   Calculated date
  EXPECTED     Expected earnings
  ACTUAL       Actual earnings

Filters:
  --retainer       Filter by retainer ID
  --calculated-on  Filter by calculated date (YYYY-MM-DD)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List retainer earning statuses
  xbe view retainer-earning-statuses list

  # Filter by retainer
  xbe view retainer-earning-statuses list --retainer 123

  # Filter by calculated date
  xbe view retainer-earning-statuses list --calculated-on 2025-01-15

  # Output as JSON
  xbe view retainer-earning-statuses list --json`,
		Args: cobra.NoArgs,
		RunE: runRetainerEarningStatusesList,
	}
	initRetainerEarningStatusesListFlags(cmd)
	return cmd
}

func init() {
	retainerEarningStatusesCmd.AddCommand(newRetainerEarningStatusesListCmd())
}

func initRetainerEarningStatusesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("retainer", "", "Filter by retainer ID")
	cmd.Flags().String("calculated-on", "", "Filter by calculated date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerEarningStatusesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRetainerEarningStatusesListOptions(cmd)
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
	query.Set("fields[retainer-earning-statuses]", "expected,actual,calculated-on,retainer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[retainer]", opts.Retainer)
	setFilterIfPresent(query, "filter[calculated-on]", opts.CalculatedOn)

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-earning-statuses", query)
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

	rows := buildRetainerEarningStatusRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRetainerEarningStatusesTable(cmd, rows)
}

func parseRetainerEarningStatusesListOptions(cmd *cobra.Command) (retainerEarningStatusesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	retainer, _ := cmd.Flags().GetString("retainer")
	calculatedOn, _ := cmd.Flags().GetString("calculated-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerEarningStatusesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Retainer:     retainer,
		CalculatedOn: calculatedOn,
	}, nil
}

func buildRetainerEarningStatusRows(resp jsonAPIResponse) []retainerEarningStatusRow {
	rows := make([]retainerEarningStatusRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := retainerEarningStatusRow{
			ID:           resource.ID,
			CalculatedOn: formatDate(stringAttr(attrs, "calculated-on")),
			Expected:     stringAttr(attrs, "expected"),
			Actual:       stringAttr(attrs, "actual"),
		}

		if rel, ok := resource.Relationships["retainer"]; ok && rel.Data != nil {
			row.RetainerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderRetainerEarningStatusesTable(cmd *cobra.Command, rows []retainerEarningStatusRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No retainer earning statuses found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRETAINER\tCALCULATED\tEXPECTED\tACTUAL")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RetainerID,
			row.CalculatedOn,
			row.Expected,
			row.Actual,
		)
	}
	return writer.Flush()
}
