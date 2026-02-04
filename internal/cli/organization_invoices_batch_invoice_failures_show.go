package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type organizationInvoicesBatchInvoiceFailuresShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchInvoiceFailureDetails struct {
	ID                               string `json:"id"`
	OrganizationInvoicesBatchInvoice string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Comment                          string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchInvoiceFailuresShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch invoice failure details",
		Long: `Show full details of an organization invoices batch invoice failure.

Output Fields:
  ID             Failure identifier
  Batch Invoice  Organization invoices batch invoice ID
  Comment        Comment (if provided)

Arguments:
  <id>    Organization invoices batch invoice failure ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch invoice failure
  xbe view organization-invoices-batch-invoice-failures show 123

  # JSON output
  xbe view organization-invoices-batch-invoice-failures show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchInvoiceFailuresShow,
	}
	initOrganizationInvoicesBatchInvoiceFailuresShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoiceFailuresCmd.AddCommand(newOrganizationInvoicesBatchInvoiceFailuresShowCmd())
}

func initOrganizationInvoicesBatchInvoiceFailuresShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoiceFailuresShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseOrganizationInvoicesBatchInvoiceFailuresShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch invoice failure id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-invoice-failures]", "organization-invoices-batch-invoice,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoice-failures/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderOrganizationInvoicesBatchInvoiceFailuresShowUnavailable(cmd, opts.JSON)
		}
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

	details := buildOrganizationInvoicesBatchInvoiceFailureDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchInvoiceFailureDetails(cmd, details)
}

func renderOrganizationInvoicesBatchInvoiceFailuresShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), organizationInvoicesBatchInvoiceFailureDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Organization invoices batch invoice failures are write-only; show is not available.")
	return nil
}

func parseOrganizationInvoicesBatchInvoiceFailuresShowOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoiceFailuresShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoiceFailuresShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceFailureDetails(resp jsonAPISingleResponse) organizationInvoicesBatchInvoiceFailureDetails {
	attrs := resp.Data.Attributes
	details := organizationInvoicesBatchInvoiceFailureDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["organization-invoices-batch-invoice"]; ok && rel.Data != nil {
		details.OrganizationInvoicesBatchInvoice = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchInvoiceFailureDetails(cmd *cobra.Command, details organizationInvoicesBatchInvoiceFailureDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OrganizationInvoicesBatchInvoice != "" {
		fmt.Fprintf(out, "Batch Invoice: %s\n", details.OrganizationInvoicesBatchInvoice)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
