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

type retainerPeriodsListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Retainer string
}

type retainerPeriodRow struct {
	ID                  string `json:"id"`
	StartOn             string `json:"start_on,omitempty"`
	EndOn               string `json:"end_on,omitempty"`
	WeeklyPaymentAmount any    `json:"weekly_payment_amount,omitempty"`
	RetainerID          string `json:"retainer_id,omitempty"`
}

func newRetainerPeriodsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List retainer periods",
		Long: `List retainer periods with filtering and pagination.

Output Columns:
  ID              Retainer period identifier
  START ON        Start date
  END ON          End date
  WEEKLY PAYMENT  Weekly payment amount
  RETAINER        Retainer ID

Filters:
  --retainer  Filter by retainer ID`,
		Example: `  # List retainer periods
  xbe view retainer-periods list

  # Filter by retainer
  xbe view retainer-periods list --retainer 123

  # Output as JSON
  xbe view retainer-periods list --json`,
		RunE: runRetainerPeriodsList,
	}
	initRetainerPeriodsListFlags(cmd)
	return cmd
}

func init() {
	retainerPeriodsCmd.AddCommand(newRetainerPeriodsListCmd())
}

func initRetainerPeriodsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("retainer", "", "Filter by retainer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerPeriodsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRetainerPeriodsListOptions(cmd)
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
	query.Set("fields[retainer-periods]", "start-on,end-on,weekly-payment-amount,retainer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[retainer]", opts.Retainer)

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-periods", query)
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

	rows := buildRetainerPeriodRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRetainerPeriodsTable(cmd, rows)
}

func parseRetainerPeriodsListOptions(cmd *cobra.Command) (retainerPeriodsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	retainer, _ := cmd.Flags().GetString("retainer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerPeriodsListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Retainer: retainer,
	}, nil
}

func buildRetainerPeriodRows(resp jsonAPIResponse) []retainerPeriodRow {
	rows := make([]retainerPeriodRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRetainerPeriodRow(resource))
	}
	return rows
}

func buildRetainerPeriodRow(resource jsonAPIResource) retainerPeriodRow {
	row := retainerPeriodRow{
		ID:                  resource.ID,
		StartOn:             stringAttr(resource.Attributes, "start-on"),
		EndOn:               stringAttr(resource.Attributes, "end-on"),
		WeeklyPaymentAmount: resource.Attributes["weekly-payment-amount"],
	}

	if rel, ok := resource.Relationships["retainer"]; ok && rel.Data != nil {
		row.RetainerID = rel.Data.ID
	}

	return row
}

func renderRetainerPeriodsTable(cmd *cobra.Command, rows []retainerPeriodRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No retainer periods found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTART ON\tEND ON\tWEEKLY PAYMENT\tRETAINER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.StartOn,
			row.EndOn,
			formatAnyValue(row.WeeklyPaymentAmount),
			row.RetainerID,
		)
	}
	return writer.Flush()
}
