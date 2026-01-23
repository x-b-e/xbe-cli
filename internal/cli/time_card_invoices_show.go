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

type timeCardInvoicesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardInvoiceDetails struct {
	ID            string `json:"id"`
	InvoiceID     string `json:"invoice_id,omitempty"`
	InvoiceType   string `json:"invoice_type,omitempty"`
	InvoiceStatus string `json:"invoice_status,omitempty"`
	SellerID      string `json:"seller_id,omitempty"`
	TimeCardID    string `json:"time_card_id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func newTimeCardInvoicesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card invoice details",
		Long: `Show the full details of a time card invoice.

Output Fields:
  ID
  Time Card ID
  Invoice ID
  Invoice Type
  Invoice Status
  Seller ID
  Created At
  Updated At

Arguments:
  <id>    The time card invoice ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show time card invoice details
  xbe view time-card-invoices show 123

  # Get JSON output
  xbe view time-card-invoices show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardInvoicesShow,
	}
	initTimeCardInvoicesShowFlags(cmd)
	return cmd
}

func init() {
	timeCardInvoicesCmd.AddCommand(newTimeCardInvoicesShowCmd())
}

func initTimeCardInvoicesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardInvoicesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardInvoicesShowOptions(cmd)
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
		return fmt.Errorf("time card invoice id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-invoices]", "created-at,updated-at,invoice,time-card")
	query.Set("fields[invoices]", "status,seller")
	query.Set("include", "invoice")

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-invoices/"+id, query)
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

	details := buildTimeCardInvoiceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardInvoiceDetails(cmd, details)
}

func parseTimeCardInvoicesShowOptions(cmd *cobra.Command) (timeCardInvoicesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardInvoicesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardInvoiceDetails(resp jsonAPISingleResponse) timeCardInvoiceDetails {
	resource := resp.Data
	attrs := resource.Attributes
	included := map[string]jsonAPIResource{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	invoiceID := relationshipIDFromMap(resource.Relationships, "invoice")
	timeCardID := relationshipIDFromMap(resource.Relationships, "time-card")

	details := timeCardInvoiceDetails{
		ID:         resource.ID,
		InvoiceID:  invoiceID,
		TimeCardID: timeCardID,
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if invoiceID != "" {
		if invoice, ok := included[resourceKey("invoices", invoiceID)]; ok {
			details.InvoiceStatus = stringAttr(invoice.Attributes, "status")
			details.InvoiceType = resolveInvoiceType(invoice.Attributes, invoice.Type)
			details.SellerID = relationshipIDFromMap(invoice.Relationships, "seller")
		}
	}

	return details
}

func renderTimeCardInvoiceDetails(cmd *cobra.Command, details timeCardInvoiceDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card ID: %s\n", details.TimeCardID)
	}
	if details.InvoiceID != "" {
		fmt.Fprintf(out, "Invoice ID: %s\n", details.InvoiceID)
	}
	if details.InvoiceType != "" {
		fmt.Fprintf(out, "Invoice Type: %s\n", details.InvoiceType)
	}
	if details.InvoiceStatus != "" {
		fmt.Fprintf(out, "Invoice Status: %s\n", details.InvoiceStatus)
	}
	if details.SellerID != "" {
		fmt.Fprintf(out, "Seller ID: %s\n", details.SellerID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
