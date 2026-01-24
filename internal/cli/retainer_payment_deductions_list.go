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

type retainerPaymentDeductionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type retainerPaymentDeductionRow struct {
	ID                  string `json:"id"`
	RetainerPaymentID   string `json:"retainer_payment_id,omitempty"`
	RetainerDeductionID string `json:"retainer_deduction_id,omitempty"`
	AppliedAmount       string `json:"applied_amount,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
}

func newRetainerPaymentDeductionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List retainer payment deductions",
		Long: `List retainer payment deductions.

Output Columns:
  ID         Retainer payment deduction identifier
  PAYMENT    Retainer payment ID
  DEDUCTION  Retainer deduction ID
  APPLIED    Applied amount
  CREATED    When the deduction was created

Filters:
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List retainer payment deductions
  xbe view retainer-payment-deductions list

  # Filter by created date
  xbe view retainer-payment-deductions list --created-at-min 2025-01-01T00:00:00Z

  # JSON output
  xbe view retainer-payment-deductions list --json`,
		Args: cobra.NoArgs,
		RunE: runRetainerPaymentDeductionsList,
	}
	initRetainerPaymentDeductionsListFlags(cmd)
	return cmd
}

func init() {
	retainerPaymentDeductionsCmd.AddCommand(newRetainerPaymentDeductionsListCmd())
}

func initRetainerPaymentDeductionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerPaymentDeductionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRetainerPaymentDeductionsListOptions(cmd)
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
	query.Set("fields[retainer-payment-deductions]", "retainer-payment,retainer-deduction,applied-amount,created-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-payment-deductions", query)
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

	rows := buildRetainerPaymentDeductionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRetainerPaymentDeductionsTable(cmd, rows)
}

func parseRetainerPaymentDeductionsListOptions(cmd *cobra.Command) (retainerPaymentDeductionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerPaymentDeductionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildRetainerPaymentDeductionRows(resp jsonAPIResponse) []retainerPaymentDeductionRow {
	rows := make([]retainerPaymentDeductionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildRetainerPaymentDeductionRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildRetainerPaymentDeductionRow(resource jsonAPIResource) retainerPaymentDeductionRow {
	attrs := resource.Attributes
	row := retainerPaymentDeductionRow{
		ID:            resource.ID,
		AppliedAmount: stringAttr(attrs, "applied-amount"),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["retainer-payment"]; ok && rel.Data != nil {
		row.RetainerPaymentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["retainer-deduction"]; ok && rel.Data != nil {
		row.RetainerDeductionID = rel.Data.ID
	}

	return row
}

func renderRetainerPaymentDeductionsTable(cmd *cobra.Command, rows []retainerPaymentDeductionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No retainer payment deductions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPAYMENT\tDEDUCTION\tAPPLIED\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RetainerPaymentID,
			row.RetainerDeductionID,
			row.AppliedAmount,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
