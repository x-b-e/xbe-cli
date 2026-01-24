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

type retainerDeductionsListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Retainer string
}

type retainerDeductionRow struct {
	ID         string `json:"id"`
	Amount     any    `json:"amount,omitempty"`
	Note       string `json:"note,omitempty"`
	RetainerID string `json:"retainer_id,omitempty"`
}

func newRetainerDeductionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List retainer deductions",
		Long: `List retainer deductions with filtering and pagination.

Output Columns:
  ID        Retainer deduction identifier
  AMOUNT    Deduction amount
  NOTE      Note (truncated)
  RETAINER  Retainer ID

Filters:
  --retainer  Filter by retainer ID`,
		Example: `  # List retainer deductions
  xbe view retainer-deductions list

  # Filter by retainer
  xbe view retainer-deductions list --retainer 123

  # Output as JSON
  xbe view retainer-deductions list --json`,
		RunE: runRetainerDeductionsList,
	}
	initRetainerDeductionsListFlags(cmd)
	return cmd
}

func init() {
	retainerDeductionsCmd.AddCommand(newRetainerDeductionsListCmd())
}

func initRetainerDeductionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("retainer", "", "Filter by retainer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerDeductionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRetainerDeductionsListOptions(cmd)
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
	query.Set("fields[retainer-deductions]", "amount,note,retainer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[retainer]", opts.Retainer)

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-deductions", query)
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

	rows := buildRetainerDeductionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRetainerDeductionsTable(cmd, rows)
}

func parseRetainerDeductionsListOptions(cmd *cobra.Command) (retainerDeductionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	retainer, _ := cmd.Flags().GetString("retainer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerDeductionsListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Retainer: retainer,
	}, nil
}

func buildRetainerDeductionRows(resp jsonAPIResponse) []retainerDeductionRow {
	rows := make([]retainerDeductionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRetainerDeductionRow(resource))
	}
	return rows
}

func buildRetainerDeductionRow(resource jsonAPIResource) retainerDeductionRow {
	row := retainerDeductionRow{
		ID:     resource.ID,
		Amount: resource.Attributes["amount"],
		Note:   stringAttr(resource.Attributes, "note"),
	}

	if rel, ok := resource.Relationships["retainer"]; ok && rel.Data != nil {
		row.RetainerID = rel.Data.ID
	}

	return row
}

func renderRetainerDeductionsTable(cmd *cobra.Command, rows []retainerDeductionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No retainer deductions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tAMOUNT\tNOTE\tRETAINER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			formatAnyValue(row.Amount),
			truncateString(row.Note, 30),
			row.RetainerID,
		)
	}
	return writer.Flush()
}
