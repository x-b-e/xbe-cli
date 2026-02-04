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

type organizationInvoicesBatchProcessesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchProcessDetails struct {
	ID                        string `json:"id"`
	OrganizationInvoicesBatch string `json:"organization_invoices_batch_id,omitempty"`
	Comment                   string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchProcessesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch process details",
		Long: `Show full details of an organization invoices batch process.

Output Fields:
  ID      Process identifier
  Batch   Organization invoices batch ID
  Comment Comment (if provided)

Arguments:
  <id>    Organization invoices batch process ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch process
  xbe view organization-invoices-batch-processes show 123

  # JSON output
  xbe view organization-invoices-batch-processes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchProcessesShow,
	}
	initOrganizationInvoicesBatchProcessesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchProcessesCmd.AddCommand(newOrganizationInvoicesBatchProcessesShowCmd())
}

func initOrganizationInvoicesBatchProcessesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchProcessesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseOrganizationInvoicesBatchProcessesShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch process id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-processes]", "organization-invoices-batch,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-processes/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderOrganizationInvoicesBatchProcessesShowUnavailable(cmd, opts.JSON)
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

	details := buildOrganizationInvoicesBatchProcessDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchProcessDetails(cmd, details)
}

func renderOrganizationInvoicesBatchProcessesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), organizationInvoicesBatchProcessDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Organization invoices batch processes are write-only; show is not available.")
	return nil
}

func parseOrganizationInvoicesBatchProcessesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchProcessesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchProcessesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchProcessDetails(resp jsonAPISingleResponse) organizationInvoicesBatchProcessDetails {
	attrs := resp.Data.Attributes
	details := organizationInvoicesBatchProcessDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["organization-invoices-batch"]; ok && rel.Data != nil {
		details.OrganizationInvoicesBatch = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchProcessDetails(cmd *cobra.Command, details organizationInvoicesBatchProcessDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OrganizationInvoicesBatch != "" {
		fmt.Fprintf(out, "Batch: %s\n", details.OrganizationInvoicesBatch)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
