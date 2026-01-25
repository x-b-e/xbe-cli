package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type tenderReRatesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderReRateListRow struct {
	ID                       string   `json:"id"`
	TenderIDs                []string `json:"tender_ids,omitempty"`
	ReRate                   bool     `json:"re_rate"`
	ReConstrain              bool     `json:"re_constrain"`
	UpdateTimeCardQuantities bool     `json:"update_time_card_quantities"`
	InvoiceIDs               []string `json:"invoice_ids,omitempty"`
}

func newTenderReRatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender re-rates",
		Long: `List tender re-rates.

Output Columns:
  ID                 Re-rate identifier
  TENDER IDS         Tender IDs (truncated)
  RE RATE            Re-rate tenders (Yes/No)
  RE CONSTRAIN       Re-constrain tenders (Yes/No)
  UPDATE TIME CARDS  Update time card quantities (Yes/No)
  INVOICE IDS        Invoice IDs (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tender re-rates
  xbe view tender-re-rates list

  # JSON output
  xbe view tender-re-rates list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderReRatesList,
	}
	initTenderReRatesListFlags(cmd)
	return cmd
}

func init() {
	tenderReRatesCmd.AddCommand(newTenderReRatesListCmd())
}

func initTenderReRatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderReRatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderReRatesListOptions(cmd)
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
	query.Set("fields[tender-re-rates]", "tender-ids,re-rate,re-constrain,update-time-card-quantities,invoice-ids")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/sombreros/tender-re-rates", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTenderReRatesUnavailable(cmd, opts.JSON)
		}
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

	rows := buildTenderReRateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderReRatesTable(cmd, rows)
}

func renderTenderReRatesUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []tenderReRateListRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Tender re-rates are write-only; list is not available.")
	return nil
}

func parseTenderReRatesListOptions(cmd *cobra.Command) (tenderReRatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderReRatesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderReRateRows(resp jsonAPIResponse) []tenderReRateListRow {
	rows := make([]tenderReRateListRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildTenderReRateRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildTenderReRateRow(resource jsonAPIResource) tenderReRateListRow {
	attrs := resource.Attributes
	row := tenderReRateListRow{
		ID:                       resource.ID,
		TenderIDs:                stringSliceAttr(attrs, "tender-ids"),
		ReRate:                   boolAttr(attrs, "re-rate"),
		ReConstrain:              boolAttr(attrs, "re-constrain"),
		UpdateTimeCardQuantities: boolAttr(attrs, "update-time-card-quantities"),
		InvoiceIDs:               stringSliceAttr(attrs, "invoice-ids"),
	}

	return row
}

func renderTenderReRatesTable(cmd *cobra.Command, rows []tenderReRateListRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender re-rates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER IDS\tRE RATE\tRE CONSTRAIN\tUPDATE TIME CARDS\tINVOICE IDS")
	for _, row := range rows {
		reRate := "No"
		if row.ReRate {
			reRate = "Yes"
		}
		reConstrain := "No"
		if row.ReConstrain {
			reConstrain = "Yes"
		}
		updateTimeCards := "No"
		if row.UpdateTimeCardQuantities {
			updateTimeCards = "Yes"
		}
		tenderIDs := truncateString(strings.Join(row.TenderIDs, ","), 30)
		invoiceIDs := truncateString(strings.Join(row.InvoiceIDs, ","), 30)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			tenderIDs,
			reRate,
			reConstrain,
			updateTimeCards,
			invoiceIDs,
		)
	}
	return writer.Flush()
}
