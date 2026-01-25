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

type tenderOffersListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderOfferRow struct {
	ID                          string `json:"id"`
	TenderType                  string `json:"tender_type,omitempty"`
	TenderID                    string `json:"tender_id,omitempty"`
	Comment                     string `json:"comment,omitempty"`
	SkipCertificationValidation bool   `json:"skip_certification_validation"`
}

func newTenderOffersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender offers",
		Long: `List tender offers.

Output Columns:
  ID           Offer identifier
  TENDER TYPE  Tender type (broker-tenders, customer-tenders)
  TENDER ID    Tender ID
  SKIP CERT    Skip certification validation (Yes/No)
  COMMENT      Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tender offers
  xbe view tender-offers list

  # JSON output
  xbe view tender-offers list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderOffersList,
	}
	initTenderOffersListFlags(cmd)
	return cmd
}

func init() {
	tenderOffersCmd.AddCommand(newTenderOffersListCmd())
}

func initTenderOffersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderOffersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderOffersListOptions(cmd)
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
	query.Set("fields[tender-offers]", "tender,comment,skip-certification-validation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/tender-offers", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTenderOffersUnavailable(cmd, opts.JSON)
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

	rows := buildTenderOfferRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderOffersTable(cmd, rows)
}

func renderTenderOffersUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []tenderOfferRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Tender offers are write-only; list is not available.")
	return nil
}

func parseTenderOffersListOptions(cmd *cobra.Command) (tenderOffersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderOffersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderOfferRows(resp jsonAPIResponse) []tenderOfferRow {
	rows := make([]tenderOfferRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildTenderOfferRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildTenderOfferRow(resource jsonAPIResource) tenderOfferRow {
	attrs := resource.Attributes
	row := tenderOfferRow{
		ID:                          resource.ID,
		Comment:                     strings.TrimSpace(stringAttr(attrs, "comment")),
		SkipCertificationValidation: boolAttr(attrs, "skip-certification-validation"),
	}

	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderType = rel.Data.Type
		row.TenderID = rel.Data.ID
	}

	return row
}

func renderTenderOffersTable(cmd *cobra.Command, rows []tenderOfferRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender offers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER TYPE\tTENDER ID\tSKIP CERT\tCOMMENT")
	for _, row := range rows {
		skipCert := "No"
		if row.SkipCertificationValidation {
			skipCert = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderType,
			row.TenderID,
			skipCert,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
