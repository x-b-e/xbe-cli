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

type invoiceStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type invoiceStatusChangeDetails struct {
	ID          string `json:"id"`
	InvoiceID   string `json:"invoice_id,omitempty"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newInvoiceStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show invoice status change details",
		Long: `Show the full details of an invoice status change.

Output Fields:
  ID
  Invoice ID
  Status
  Changed At
  Changed By (user ID)
  Comment

Arguments:
  <id>    The status change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an invoice status change
  xbe view invoice-status-changes show 123

  # Output as JSON
  xbe view invoice-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoiceStatusChangesShow,
	}
	initInvoiceStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	invoiceStatusChangesCmd.AddCommand(newInvoiceStatusChangesShowCmd())
}

func initInvoiceStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceStatusChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInvoiceStatusChangesShowOptions(cmd)
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
		return fmt.Errorf("invoice status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[invoice-status-changes]", "status,changed-at,comment,invoice,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-status-changes/"+id, query)
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

	details := buildInvoiceStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInvoiceStatusChangeDetails(cmd, details)
}

func parseInvoiceStatusChangesShowOptions(cmd *cobra.Command) (invoiceStatusChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInvoiceStatusChangeDetails(resp jsonAPISingleResponse) invoiceStatusChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := invoiceStatusChangeDetails{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		details.InvoiceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
	}

	return details
}

func renderInvoiceStatusChangeDetails(cmd *cobra.Command, details invoiceStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.InvoiceID != "" {
		fmt.Fprintf(out, "Invoice ID: %s\n", details.InvoiceID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ChangedAt != "" {
		fmt.Fprintf(out, "Changed At: %s\n", details.ChangedAt)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
