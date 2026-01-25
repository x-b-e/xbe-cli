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

type organizationInvoicesBatchStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchStatusChangeDetails struct {
	ID                        string `json:"id"`
	OrganizationInvoicesBatch string `json:"organization_invoices_batch_id,omitempty"`
	Status                    string `json:"status,omitempty"`
	ChangedAt                 string `json:"changed_at,omitempty"`
	ChangedBy                 string `json:"changed_by_id,omitempty"`
	Comment                   string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch status change details",
		Long: `Show full details of an organization invoices batch status change.

Output Fields:
  ID         Status change identifier
  Status     Batch status
  Changed At When the status changed
  Changed By User who made the change
  Batch      Organization invoices batch ID
  Comment    Comment (if provided)

Arguments:
  <id>    Organization invoices batch status change ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch status change
  xbe view organization-invoices-batch-status-changes show 123

  # JSON output
  xbe view organization-invoices-batch-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchStatusChangesShow,
	}
	initOrganizationInvoicesBatchStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchStatusChangesCmd.AddCommand(newOrganizationInvoicesBatchStatusChangesShowCmd())
}

func initOrganizationInvoicesBatchStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchStatusChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOrganizationInvoicesBatchStatusChangesShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-status-changes]", "status,changed-at,comment,changed-by,organization-invoices-batch")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-status-changes/"+id, query)
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

	details := buildOrganizationInvoicesBatchStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchStatusChangeDetails(cmd, details)
}

func parseOrganizationInvoicesBatchStatusChangesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchStatusChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchStatusChangeDetails(resp jsonAPISingleResponse) organizationInvoicesBatchStatusChangeDetails {
	attrs := resp.Data.Attributes
	details := organizationInvoicesBatchStatusChangeDetails{
		ID:        resp.Data.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["organization-invoices-batch"]; ok && rel.Data != nil {
		details.OrganizationInvoicesBatch = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedBy = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchStatusChangeDetails(cmd *cobra.Command, details organizationInvoicesBatchStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Status: %s\n", formatOptional(details.Status))
	fmt.Fprintf(out, "Changed At: %s\n", formatOptional(details.ChangedAt))
	fmt.Fprintf(out, "Changed By: %s\n", formatOptional(details.ChangedBy))
	fmt.Fprintf(out, "Batch: %s\n", formatOptional(details.OrganizationInvoicesBatch))
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
