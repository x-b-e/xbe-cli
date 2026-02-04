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

type organizationInvoicesBatchPdfTemplatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchPdfTemplateDetails struct {
	ID               string `json:"id"`
	Description      string `json:"description,omitempty"`
	Template         string `json:"template,omitempty"`
	IsActive         bool   `json:"is_active"`
	IsGlobal         bool   `json:"is_global"`
	Organization     string `json:"organization,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerName       string `json:"broker_name,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedBy        string `json:"created_by,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func newOrganizationInvoicesBatchPdfTemplatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch PDF template details",
		Long: `Show the full details of an organization invoices batch PDF template.

Output Fields:
  ID            Template identifier
  Description   Template description
  Active        Whether the template is active
  Global        Whether the template is global
  Organization  Organization (Type/ID)
  Broker        Broker ID
  Created By    Created-by user ID
  Created At    Created timestamp
  Updated At    Updated timestamp
  Template      Template content

Arguments:
  <id>  Template ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a template
  xbe view organization-invoices-batch-pdf-templates show 123

  # Output as JSON
  xbe view organization-invoices-batch-pdf-templates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchPdfTemplatesShow,
	}
	initOrganizationInvoicesBatchPdfTemplatesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfTemplatesCmd.AddCommand(newOrganizationInvoicesBatchPdfTemplatesShowCmd())
}

func initOrganizationInvoicesBatchPdfTemplatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfTemplatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseOrganizationInvoicesBatchPdfTemplatesShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch PDF template id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-pdf-templates]", strings.Join([]string{
		"template",
		"description",
		"is-active",
		"is-global",
		"organization",
		"broker",
		"created-by",
		"created-at",
		"updated-at",
	}, ","))
	query.Set("include", "organization,broker,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developers]", "name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-pdf-templates/"+id, query)
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

	details := buildOrganizationInvoicesBatchPdfTemplateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchPdfTemplateDetails(cmd, details)
}

func parseOrganizationInvoicesBatchPdfTemplatesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfTemplatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchPdfTemplatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchPdfTemplateDetails(resp jsonAPISingleResponse) organizationInvoicesBatchPdfTemplateDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := organizationInvoicesBatchPdfTemplateDetails{
		ID:          resp.Data.ID,
		Description: stringAttr(attrs, "description"),
		Template:    stringAttr(attrs, "template"),
		IsActive:    boolAttr(attrs, "is-active"),
		IsGlobal:    boolAttr(attrs, "is-global"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Organization = organizationNameFromIncluded(inc)
		}
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = organizationNameFromIncluded(inc)
		}
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedBy = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	return details
}

func renderOrganizationInvoicesBatchPdfTemplateDetails(cmd *cobra.Command, details organizationInvoicesBatchPdfTemplateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	fmt.Fprintf(out, "Active: %t\n", details.IsActive)
	fmt.Fprintf(out, "Global: %t\n", details.IsGlobal)

	if details.OrganizationType != "" || details.OrganizationID != "" || details.Organization != "" {
		orgLabel := formatRelated(details.Organization, formatPolymorphic(details.OrganizationType, details.OrganizationID))
		if orgLabel != "" {
			fmt.Fprintf(out, "Organization: %s\n", orgLabel)
		}
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CreatedByID != "" || details.CreatedBy != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedBy, details.CreatedByID))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.Template != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Template:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Template)
	}

	return nil
}
