package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type organizationInvoicesBatchInvoiceUnbatchingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchInvoiceUnbatchingDetails struct {
	ID                                 string `json:"id"`
	OrganizationInvoicesBatchInvoiceID string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Comment                            string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchInvoiceUnbatchingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch invoice unbatching details",
		Long: `Show the full details of an organization invoices batch invoice unbatching.

Output Fields:
  ID
  Organization Invoices Batch Invoice ID
  Comment

Arguments:
  <id>    The organization invoices batch invoice unbatching ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch invoice unbatching
  xbe view organization-invoices-batch-invoice-unbatchings show 123

  # JSON output
  xbe view organization-invoices-batch-invoice-unbatchings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchInvoiceUnbatchingsShow,
	}
	initOrganizationInvoicesBatchInvoiceUnbatchingsShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoiceUnbatchingsCmd.AddCommand(newOrganizationInvoicesBatchInvoiceUnbatchingsShowCmd())
}

func initOrganizationInvoicesBatchInvoiceUnbatchingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoiceUnbatchingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoiceUnbatchingsShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch invoice unbatching id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoice-unbatchings/"+id, nil)
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

	details := buildOrganizationInvoicesBatchInvoiceUnbatchingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchInvoiceUnbatchingDetails(cmd, details)
}

func parseOrganizationInvoicesBatchInvoiceUnbatchingsShowOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoiceUnbatchingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoiceUnbatchingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceUnbatchingDetails(resp jsonAPISingleResponse) organizationInvoicesBatchInvoiceUnbatchingDetails {
	resource := resp.Data
	return organizationInvoicesBatchInvoiceUnbatchingDetails{
		ID:                                 resource.ID,
		OrganizationInvoicesBatchInvoiceID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch-invoice"),
		Comment:                            stringAttr(resource.Attributes, "comment"),
	}
}

func renderOrganizationInvoicesBatchInvoiceUnbatchingDetails(cmd *cobra.Command, details organizationInvoicesBatchInvoiceUnbatchingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OrganizationInvoicesBatchInvoiceID != "" {
		fmt.Fprintf(out, "Organization Invoices Batch Invoice ID: %s\n", details.OrganizationInvoicesBatchInvoiceID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
