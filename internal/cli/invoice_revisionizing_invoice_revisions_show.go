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

type invoiceRevisionizingInvoiceRevisionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type invoiceRevisionizingInvoiceRevisionDetails struct {
	ID                         string `json:"id"`
	InvoiceRevisionizingWorkID string `json:"invoice_revisionizing_work_id,omitempty"`
	InvoiceRevisionID          string `json:"invoice_revision_id,omitempty"`
	InvoiceID                  string `json:"invoice_id,omitempty"`
	RevisionNumber             string `json:"revision_number,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

func newInvoiceRevisionizingInvoiceRevisionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show invoice revisionizing invoice revision details",
		Long: `Show the full details of an invoice revisionizing invoice revision.

Output Fields:
  ID
  Invoice Revisionizing Work ID
  Invoice Revision ID
  Invoice ID
  Revision Number
  Created At
  Updated At

Arguments:
  <id>    The invoice revisionizing invoice revision ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an invoice revisionizing invoice revision
  xbe view invoice-revisionizing-invoice-revisions show 123

  # Output as JSON
  xbe view invoice-revisionizing-invoice-revisions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoiceRevisionizingInvoiceRevisionsShow,
	}
	initInvoiceRevisionizingInvoiceRevisionsShowFlags(cmd)
	return cmd
}

func init() {
	invoiceRevisionizingInvoiceRevisionsCmd.AddCommand(newInvoiceRevisionizingInvoiceRevisionsShowCmd())
}

func initInvoiceRevisionizingInvoiceRevisionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceRevisionizingInvoiceRevisionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseInvoiceRevisionizingInvoiceRevisionsShowOptions(cmd)
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
		return fmt.Errorf("invoice revisionizing invoice revision id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[invoice-revisionizing-invoice-revisions]", "revision-number,created-at,updated-at,invoice-revisionizing-work,invoice-revision,invoice")

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-revisionizing-invoice-revisions/"+id, query)
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

	details := buildInvoiceRevisionizingInvoiceRevisionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInvoiceRevisionizingInvoiceRevisionDetails(cmd, details)
}

func parseInvoiceRevisionizingInvoiceRevisionsShowOptions(cmd *cobra.Command) (invoiceRevisionizingInvoiceRevisionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceRevisionizingInvoiceRevisionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInvoiceRevisionizingInvoiceRevisionDetails(resp jsonAPISingleResponse) invoiceRevisionizingInvoiceRevisionDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := invoiceRevisionizingInvoiceRevisionDetails{
		ID:             resource.ID,
		RevisionNumber: stringAttr(attrs, "revision-number"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["invoice-revisionizing-work"]; ok && rel.Data != nil {
		details.InvoiceRevisionizingWorkID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["invoice-revision"]; ok && rel.Data != nil {
		details.InvoiceRevisionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		details.InvoiceID = rel.Data.ID
	}

	return details
}

func renderInvoiceRevisionizingInvoiceRevisionDetails(cmd *cobra.Command, details invoiceRevisionizingInvoiceRevisionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.InvoiceRevisionizingWorkID != "" {
		fmt.Fprintf(out, "Invoice Revisionizing Work ID: %s\n", details.InvoiceRevisionizingWorkID)
	}
	if details.InvoiceRevisionID != "" {
		fmt.Fprintf(out, "Invoice Revision ID: %s\n", details.InvoiceRevisionID)
	}
	if details.InvoiceID != "" {
		fmt.Fprintf(out, "Invoice ID: %s\n", details.InvoiceID)
	}
	if details.RevisionNumber != "" {
		fmt.Fprintf(out, "Revision Number: %s\n", details.RevisionNumber)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
