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

type truckerInvoicePaymentsShowOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	TruckerID string
	Trucker   string
}

type truckerInvoicePaymentDetails struct {
	ID            string                          `json:"id"`
	PaymentAmount string                          `json:"payment_amount,omitempty"`
	PaymentDate   string                          `json:"payment_date,omitempty"`
	CreatedAt     string                          `json:"created_at,omitempty"`
	UpdatedAt     string                          `json:"updated_at,omitempty"`
	TruckerID     string                          `json:"trucker_id,omitempty"`
	TruckerName   string                          `json:"trucker_name,omitempty"`
	LineItems     []truckerInvoicePaymentLineItem `json:"line_items,omitempty"`
}

type truckerInvoicePaymentLineItem struct {
	Amount         string `json:"amount,omitempty"`
	QuickbooksType string `json:"quickbooks_type,omitempty"`
	QuickbooksID   string `json:"quickbooks_id,omitempty"`
	Type           string `json:"type,omitempty"`
}

func newTruckerInvoicePaymentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker invoice payment details",
		Long: `Show the full details of a trucker invoice payment.

Output Fields:
  ID
  Payment Amount
  Payment Date
  Created At
  Updated At
  Trucker
  Line Items

Arguments:
  <id>    Payment ID (required).

Filters:
  --trucker-id  Trucker ID (required; scopes the QuickBooks company)
  --trucker     Trucker relationship ID

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a payment
  xbe view trucker-invoice-payments show 123 --trucker-id 456

  # JSON output
  xbe view trucker-invoice-payments show 123 --trucker-id 456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerInvoicePaymentsShow,
	}
	initTruckerInvoicePaymentsShowFlags(cmd)
	return cmd
}

func init() {
	truckerInvoicePaymentsCmd.AddCommand(newTruckerInvoicePaymentsShowCmd())
}

func initTruckerInvoicePaymentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("trucker-id", "", "Trucker ID (required; scopes QuickBooks company)")
	cmd.Flags().String("trucker", "", "Trucker relationship ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerInvoicePaymentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTruckerInvoicePaymentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.TruckerID) == "" && strings.TrimSpace(opts.Trucker) == "" {
		err := fmt.Errorf("either --trucker-id or --trucker is required to scope QuickBooks payments")
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("trucker invoice payment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-invoice-payments]", "payment-amount,payment-date,line-items,created-at,updated-at,trucker")
	query.Set("fields[truckers]", "company-name")
	query.Set("include", "trucker")

	setFilterIfPresent(query, "filter[trucker_id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-invoice-payments/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTruckerInvoicePaymentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTruckerInvoicePaymentDetails(cmd, details)
}

func parseTruckerInvoicePaymentsShowOptions(cmd *cobra.Command) (truckerInvoicePaymentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerInvoicePaymentsShowOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		TruckerID: truckerID,
		Trucker:   trucker,
	}, nil
}

func buildTruckerInvoicePaymentDetails(resp jsonAPISingleResponse) truckerInvoicePaymentDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := truckerInvoicePaymentDetails{
		ID:            resp.Data.ID,
		PaymentAmount: stringAttr(attrs, "payment-amount"),
		PaymentDate:   formatDate(stringAttr(attrs, "payment-date")),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDateTime(stringAttr(attrs, "updated-at")),
		LineItems:     parseTruckerInvoicePaymentLineItems(attrs),
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = strings.TrimSpace(firstNonEmpty(
				stringAttr(trucker.Attributes, "company-name"),
				stringAttr(trucker.Attributes, "name"),
			))
		}
	}

	return details
}

func parseTruckerInvoicePaymentLineItems(attrs map[string]any) []truckerInvoicePaymentLineItem {
	if attrs == nil {
		return nil
	}
	value, ok := attrs["line-items"]
	if !ok || value == nil {
		return nil
	}

	items, ok := value.([]any)
	if !ok {
		return nil
	}

	results := make([]truckerInvoicePaymentLineItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		line := truckerInvoicePaymentLineItem{
			Amount:         stringFromAny(entry["amount"]),
			QuickbooksType: stringFromAny(entry["quickbooks-type"]),
			QuickbooksID:   stringFromAny(entry["quickbooks-id"]),
			Type:           stringFromAny(entry["type"]),
		}
		if line.Amount == "" && line.QuickbooksType == "" && line.QuickbooksID == "" && line.Type == "" {
			continue
		}
		results = append(results, line)
	}

	return results
}

func renderTruckerInvoicePaymentDetails(cmd *cobra.Command, details truckerInvoicePaymentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PaymentAmount != "" {
		fmt.Fprintf(out, "Payment Amount: %s\n", details.PaymentAmount)
	}
	if details.PaymentDate != "" {
		fmt.Fprintf(out, "Payment Date: %s\n", details.PaymentDate)
	}
	if details.TruckerID != "" {
		if details.TruckerName != "" {
			fmt.Fprintf(out, "Trucker: %s (ID: %s)\n", details.TruckerName, details.TruckerID)
		} else {
			fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
		}
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Line Items:")
	if len(details.LineItems) == 0 {
		fmt.Fprintln(out, "  (none)")
		return nil
	}

	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "AMOUNT\tQB TYPE\tQB ID\tTYPE")
	for _, item := range details.LineItems {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			item.Amount,
			item.QuickbooksType,
			item.QuickbooksID,
			item.Type,
		)
	}
	return writer.Flush()
}
