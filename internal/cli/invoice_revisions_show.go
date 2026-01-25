package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type invoiceRevisionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type invoiceRevisionDetails struct {
	ID                  string   `json:"id"`
	Revision            string   `json:"revision,omitempty"`
	InvoiceResourceType string   `json:"invoice_resource_type,omitempty"`
	InvoiceID           string   `json:"invoice_id,omitempty"`
	LineItems           any      `json:"line_items,omitempty"`
	LineItemMemberNames []string `json:"line_item_member_names,omitempty"`
}

func newInvoiceRevisionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show invoice revision details",
		Long: `Show the full details of an invoice revision.

Arguments:
  <id>  The invoice revision ID (required).`,
		Example: `  # Show an invoice revision
  xbe view invoice-revisions show 123

  # Output as JSON
  xbe view invoice-revisions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoiceRevisionsShow,
	}
	initInvoiceRevisionsShowFlags(cmd)
	return cmd
}

func init() {
	invoiceRevisionsCmd.AddCommand(newInvoiceRevisionsShowCmd())
}

func initInvoiceRevisionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceRevisionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInvoiceRevisionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("invoice revision id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[invoice-revisions]", strings.Join([]string{
		"revision",
		"line-items",
		"line-item-member-names",
		"invoice",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-revisions/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildInvoiceRevisionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInvoiceRevisionDetails(cmd, details)
}

func parseInvoiceRevisionsShowOptions(cmd *cobra.Command) (invoiceRevisionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceRevisionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInvoiceRevisionDetails(resp jsonAPISingleResponse) invoiceRevisionDetails {
	attrs := resp.Data.Attributes
	details := invoiceRevisionDetails{
		ID:                  resp.Data.ID,
		Revision:            stringAttr(attrs, "revision"),
		LineItems:           anyAttr(attrs, "line-items"),
		LineItemMemberNames: stringSliceAttr(attrs, "line-item-member-names"),
	}

	if rel, ok := resp.Data.Relationships["invoice"]; ok && rel.Data != nil {
		details.InvoiceResourceType = rel.Data.Type
		details.InvoiceID = rel.Data.ID
	}

	return details
}

func renderInvoiceRevisionDetails(cmd *cobra.Command, details invoiceRevisionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Revision != "" {
		fmt.Fprintf(out, "Revision: %s\n", details.Revision)
	}
	if details.InvoiceResourceType != "" || details.InvoiceID != "" {
		fmt.Fprintf(out, "Invoice: %s\n", formatPolymorphic(details.InvoiceResourceType, details.InvoiceID))
	}
	if len(details.LineItemMemberNames) > 0 {
		fmt.Fprintf(out, "Line Item Member Names: %s\n", strings.Join(details.LineItemMemberNames, ", "))
	}
	if details.LineItems != nil {
		fmt.Fprintf(out, "Line Items: %d\n", countConstraintItems(details.LineItems))
		if formatted := formatAnyJSON(details.LineItems); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Line Item Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
