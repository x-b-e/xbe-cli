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

type tenderAcceptancesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderAcceptanceRow struct {
	ID                                string   `json:"id"`
	TenderType                        string   `json:"tender_type,omitempty"`
	TenderID                          string   `json:"tender_id,omitempty"`
	Comment                           string   `json:"comment,omitempty"`
	SkipCertificationValidation       bool     `json:"skip_certification_validation"`
	RejectedTenderJobScheduleShiftIDs []string `json:"rejected_tender_job_schedule_shift_ids,omitempty"`
}

func newTenderAcceptancesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender acceptances",
		Long: `List tender acceptances.

Output Columns:
  ID               Acceptance identifier
  TENDER TYPE      Tender type (broker-tenders, customer-tenders)
  TENDER ID        Tender ID
  SKIP CERT        Skip certification validation (Yes/No)
  REJECTED SHIFTS  Rejected tender job schedule shift IDs (truncated)
  COMMENT          Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tender acceptances
  xbe view tender-acceptances list

  # JSON output
  xbe view tender-acceptances list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderAcceptancesList,
	}
	initTenderAcceptancesListFlags(cmd)
	return cmd
}

func init() {
	tenderAcceptancesCmd.AddCommand(newTenderAcceptancesListCmd())
}

func initTenderAcceptancesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderAcceptancesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderAcceptancesListOptions(cmd)
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
	query.Set("fields[tender-acceptances]", "tender,comment,skip-certification-validation,rejected-tender-job-schedule-shift-ids")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/tender-acceptances", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTenderAcceptancesUnavailable(cmd, opts.JSON)
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

	rows := buildTenderAcceptanceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderAcceptancesTable(cmd, rows)
}

func renderTenderAcceptancesUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []tenderAcceptanceRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Tender acceptances are write-only; list is not available.")
	return nil
}

func parseTenderAcceptancesListOptions(cmd *cobra.Command) (tenderAcceptancesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderAcceptancesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderAcceptanceRows(resp jsonAPIResponse) []tenderAcceptanceRow {
	rows := make([]tenderAcceptanceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildTenderAcceptanceRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildTenderAcceptanceRow(resource jsonAPIResource) tenderAcceptanceRow {
	attrs := resource.Attributes
	row := tenderAcceptanceRow{
		ID:                                resource.ID,
		Comment:                           strings.TrimSpace(stringAttr(attrs, "comment")),
		SkipCertificationValidation:       boolAttr(attrs, "skip-certification-validation"),
		RejectedTenderJobScheduleShiftIDs: stringSliceAttr(attrs, "rejected-tender-job-schedule-shift-ids"),
	}

	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderType = rel.Data.Type
		row.TenderID = rel.Data.ID
	}

	return row
}

func renderTenderAcceptancesTable(cmd *cobra.Command, rows []tenderAcceptanceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender acceptances found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER TYPE\tTENDER ID\tSKIP CERT\tREJECTED SHIFTS\tCOMMENT")
	for _, row := range rows {
		skipCert := "No"
		if row.SkipCertificationValidation {
			skipCert = "Yes"
		}
		rejected := strings.Join(row.RejectedTenderJobScheduleShiftIDs, ",")
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderType,
			row.TenderID,
			skipCert,
			truncateString(rejected, 30),
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
