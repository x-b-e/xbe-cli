package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type shiftCountersListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newShiftCountersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shift counters",
		Long: `List shift counters.

Shift counters are generated on demand and are not persisted, so this list is
typically empty.

Output Columns:
  ID            Counter identifier
  START_AT_MIN  Minimum shift start timestamp used for the count
  COUNT         Number of accepted shifts after start_at_min`,
		Example: `  # List shift counters
  xbe view shift-counters list

  # JSON output
  xbe view shift-counters list --json`,
		RunE: runShiftCountersList,
	}
	initShiftCountersListFlags(cmd)
	return cmd
}

func init() {
	shiftCountersCmd.AddCommand(newShiftCountersListCmd())
}

func initShiftCountersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftCountersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseShiftCountersListOptions(cmd)
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

	body, _, err := client.Get(cmd.Context(), "/v1/shift-counters", query)
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

	rows := buildShiftCounterRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderShiftCountersTable(cmd, rows)
}

func parseShiftCountersListOptions(cmd *cobra.Command) (shiftCountersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftCountersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

type shiftCounterRow struct {
	ID         string `json:"id"`
	StartAtMin string `json:"start_at_min,omitempty"`
	Count      int    `json:"count,omitempty"`
}

func buildShiftCounterRows(resp jsonAPIResponse) []shiftCounterRow {
	rows := make([]shiftCounterRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, shiftCounterRow{
			ID:         resource.ID,
			StartAtMin: formatDateTime(stringAttr(resource.Attributes, "start-at-min")),
			Count:      intAttr(resource.Attributes, "count"),
		})
	}
	return rows
}

func renderShiftCountersTable(cmd *cobra.Command, rows []shiftCounterRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift counters found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTART_AT_MIN\tCOUNT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%d\n",
			row.ID,
			row.StartAtMin,
			row.Count,
		)
	}
	return writer.Flush()
}
