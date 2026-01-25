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

type tenderJobScheduleShiftsMaterialTransactionsChecksumsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderJobScheduleShiftsMaterialTransactionsChecksumRow struct {
	ID                  string `json:"id"`
	RawJobNumber        string `json:"raw_job_number,omitempty"`
	TransactionAtMin    string `json:"transaction_at_min,omitempty"`
	TransactionAtMax    string `json:"transaction_at_max,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
}

func newTenderJobScheduleShiftsMaterialTransactionsChecksumsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender job schedule shift material transaction checksums",
		Long: `List tender job schedule shift material transaction checksums.

Output Columns:
  ID         Checksum record identifier
  RAW JOB    Raw job number
  MIN AT     Transaction window start (UTC)
  MAX AT     Transaction window end (UTC)
  JPP ID     Job production plan ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List checksum records
  xbe view tender-job-schedule-shifts-material-transactions-checksums list

  # Limit results
  xbe view tender-job-schedule-shifts-material-transactions-checksums list --limit 5

  # Output as JSON
  xbe view tender-job-schedule-shifts-material-transactions-checksums list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderJobScheduleShiftsMaterialTransactionsChecksumsList,
	}
	initTenderJobScheduleShiftsMaterialTransactionsChecksumsListFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftsMaterialTransactionsChecksumsCmd.AddCommand(newTenderJobScheduleShiftsMaterialTransactionsChecksumsListCmd())
}

func initTenderJobScheduleShiftsMaterialTransactionsChecksumsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftsMaterialTransactionsChecksumsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderJobScheduleShiftsMaterialTransactionsChecksumsListOptions(cmd)
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
	query.Set("fields[tender-job-schedule-shifts-material-transactions-checksums]", "raw-job-number,transaction-at-min,transaction-at-max,job-production-plan-id")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shifts-material-transactions-checksums", query)
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

	rows := buildTenderJobScheduleShiftsMaterialTransactionsChecksumRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderJobScheduleShiftsMaterialTransactionsChecksumsTable(cmd, rows)
}

func parseTenderJobScheduleShiftsMaterialTransactionsChecksumsListOptions(cmd *cobra.Command) (tenderJobScheduleShiftsMaterialTransactionsChecksumsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftsMaterialTransactionsChecksumsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderJobScheduleShiftsMaterialTransactionsChecksumRows(resp jsonAPIResponse) []tenderJobScheduleShiftsMaterialTransactionsChecksumRow {
	rows := make([]tenderJobScheduleShiftsMaterialTransactionsChecksumRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, tenderJobScheduleShiftsMaterialTransactionsChecksumRow{
			ID:                  resource.ID,
			RawJobNumber:        stringAttr(resource.Attributes, "raw-job-number"),
			TransactionAtMin:    formatDateTime(stringAttr(resource.Attributes, "transaction-at-min")),
			TransactionAtMax:    formatDateTime(stringAttr(resource.Attributes, "transaction-at-max")),
			JobProductionPlanID: stringAttr(resource.Attributes, "job-production-plan-id"),
		})
	}
	return rows
}

func renderTenderJobScheduleShiftsMaterialTransactionsChecksumsTable(cmd *cobra.Command, rows []tenderJobScheduleShiftsMaterialTransactionsChecksumRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No checksum records found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tRAW JOB\tMIN AT\tMAX AT\tJPP ID")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RawJobNumber,
			row.TransactionAtMin,
			row.TransactionAtMax,
			row.JobProductionPlanID,
		)
	}
	return w.Flush()
}
