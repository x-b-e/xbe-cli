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

type organizationInvoicesBatchesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchDetails struct {
	ID                                  string   `json:"id"`
	Status                              string   `json:"status,omitempty"`
	InvoiceTypes                        []string `json:"invoice_types,omitempty"`
	CreatedAt                           string   `json:"created_at,omitempty"`
	UpdatedAt                           string   `json:"updated_at,omitempty"`
	OrganizationType                    string   `json:"organization_type,omitempty"`
	OrganizationID                      string   `json:"organization_id,omitempty"`
	BrokerID                            string   `json:"broker_id,omitempty"`
	CreatedByID                         string   `json:"created_by_id,omitempty"`
	UpdatedByID                         string   `json:"updated_by_id,omitempty"`
	InvoiceIDs                          []string `json:"invoice_ids,omitempty"`
	OrganizationInvoicesBatchFileIDs    []string `json:"organization_invoices_batch_file_ids,omitempty"`
	OrganizationInvoicesBatchInvoiceIDs []string `json:"organization_invoices_batch_invoice_ids,omitempty"`
	OrganizationInvoicesBatchStatusIDs  []string `json:"organization_invoices_batch_status_change_ids,omitempty"`
}

func newOrganizationInvoicesBatchesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch details",
		Long: `Show the full details of an organization invoices batch.

Output Fields:
  ID
  Status
  Invoice Types
  Created At
  Updated At
  Organization (type + ID)
  Broker ID
  Created By
  Updated By
  Invoice IDs
  Organization Invoices Batch File IDs
  Organization Invoices Batch Invoice IDs
  Organization Invoices Batch Status Change IDs

Arguments:
  <id>    The organization invoices batch ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch
  xbe view organization-invoices-batches show 123

  # JSON output
  xbe view organization-invoices-batches show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchesShow,
	}
	initOrganizationInvoicesBatchesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchesCmd.AddCommand(newOrganizationInvoicesBatchesShowCmd())
}

func initOrganizationInvoicesBatchesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseOrganizationInvoicesBatchesShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("organization invoices batch id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batches]", "status,invoice-types,created-at,updated-at,organization,broker,created-by,updated-by,invoices,organization-invoices-batch-files,organization-invoices-batch-invoices,organization-invoices-batch-status-changes")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batches/"+id, query)
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

	details := buildOrganizationInvoicesBatchDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchDetails(cmd, details)
}

func parseOrganizationInvoicesBatchesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return organizationInvoicesBatchesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return organizationInvoicesBatchesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return organizationInvoicesBatchesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return organizationInvoicesBatchesShowOptions{}, err
	}

	return organizationInvoicesBatchesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchDetails(resp jsonAPISingleResponse) organizationInvoicesBatchDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := organizationInvoicesBatchDetails{
		ID:                                  resource.ID,
		Status:                              stringAttr(attrs, "status"),
		InvoiceTypes:                        stringSliceAttr(attrs, "invoice-types"),
		CreatedAt:                           formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                           formatDateTime(stringAttr(attrs, "updated-at")),
		BrokerID:                            relationshipIDFromMap(resource.Relationships, "broker"),
		CreatedByID:                         relationshipIDFromMap(resource.Relationships, "created-by"),
		UpdatedByID:                         relationshipIDFromMap(resource.Relationships, "updated-by"),
		InvoiceIDs:                          relationshipIDsFromMap(resource.Relationships, "invoices"),
		OrganizationInvoicesBatchFileIDs:    relationshipIDsFromMap(resource.Relationships, "organization-invoices-batch-files"),
		OrganizationInvoicesBatchInvoiceIDs: relationshipIDsFromMap(resource.Relationships, "organization-invoices-batch-invoices"),
		OrganizationInvoicesBatchStatusIDs:  relationshipIDsFromMap(resource.Relationships, "organization-invoices-batch-status-changes"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchDetails(cmd *cobra.Command, details organizationInvoicesBatchDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if len(details.InvoiceTypes) > 0 {
		fmt.Fprintf(out, "Invoice Types: %s\n", strings.Join(details.InvoiceTypes, ", "))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s:%s\n", details.OrganizationType, details.OrganizationID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.UpdatedByID != "" {
		fmt.Fprintf(out, "Updated By: %s\n", details.UpdatedByID)
	}
	if len(details.InvoiceIDs) > 0 {
		fmt.Fprintf(out, "Invoice IDs: %s\n", strings.Join(details.InvoiceIDs, ", "))
	}
	if len(details.OrganizationInvoicesBatchFileIDs) > 0 {
		fmt.Fprintf(out, "Batch File IDs: %s\n", strings.Join(details.OrganizationInvoicesBatchFileIDs, ", "))
	}
	if len(details.OrganizationInvoicesBatchInvoiceIDs) > 0 {
		fmt.Fprintf(out, "Batch Invoice IDs: %s\n", strings.Join(details.OrganizationInvoicesBatchInvoiceIDs, ", "))
	}
	if len(details.OrganizationInvoicesBatchStatusIDs) > 0 {
		fmt.Fprintf(out, "Batch Status Change IDs: %s\n", strings.Join(details.OrganizationInvoicesBatchStatusIDs, ", "))
	}

	return nil
}
