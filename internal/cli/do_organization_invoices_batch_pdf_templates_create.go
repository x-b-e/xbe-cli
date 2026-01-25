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

type doOrganizationInvoicesBatchPdfTemplatesCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Template         string
	Description      string
	IsActive         bool
	IsGlobal         bool
	Organization     string
	OrganizationType string
	OrganizationID   string
	Broker           string
	CreatedBy        string
}

func newDoOrganizationInvoicesBatchPdfTemplatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an organization invoices batch PDF template",
		Long: `Create an organization invoices batch PDF template.

Required flags:
  --template       Template content
  --organization   Organization in Type|ID format (required unless --is-global true)
  --broker         Broker ID (required unless --is-global true)

Optional flags:
  --description
  --is-active
  --is-global
  --organization-type / --organization-id
  --created-by`,
		Example: `  # Create a template for a broker organization
  xbe do organization-invoices-batch-pdf-templates create \\
    --organization Broker|123 \\
    --broker 123 \\
    --description "Default template" \\
    --template "{{invoice_number}}"

  # Create an inactive template
  xbe do organization-invoices-batch-pdf-templates create \\
    --organization Broker|123 \\
    --broker 123 \\
    --is-active=false \\
    --template "{{invoice_number}}"`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchPdfTemplatesCreate,
	}
	initDoOrganizationInvoicesBatchPdfTemplatesCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchPdfTemplatesCmd.AddCommand(newDoOrganizationInvoicesBatchPdfTemplatesCreateCmd())
}

func initDoOrganizationInvoicesBatchPdfTemplatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("template", "", "Template content (required)")
	cmd.Flags().String("description", "", "Template description")
	cmd.Flags().Bool("is-active", false, "Whether the template is active")
	cmd.Flags().Bool("is-global", false, "Whether the template is global")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (required unless --is-global true)")
	cmd.Flags().String("organization-type", "", "Organization type (optional if --organization is set)")
	cmd.Flags().String("organization-id", "", "Organization ID (optional if --organization is set)")
	cmd.Flags().String("broker", "", "Broker ID (required unless --is-global true)")
	cmd.Flags().String("created-by", "", "Created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationInvoicesBatchPdfTemplatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchPdfTemplatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.Template) == "" {
		err := fmt.Errorf("--template is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	orgType, orgID, err := resolveOrganizationInvoicesBatchPdfTemplateOrganization(cmd, opts.Organization, opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requireOrganization := true
	if cmd.Flags().Changed("is-global") && opts.IsGlobal {
		requireOrganization = false
	}

	if requireOrganization {
		if orgType == "" || orgID == "" {
			err := fmt.Errorf("--organization is required (format: Type|ID) or specify --organization-type and --organization-id")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if strings.TrimSpace(opts.Broker) == "" {
			err := fmt.Errorf("--broker is required unless --is-global true")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{
		"template": opts.Template,
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}
	if cmd.Flags().Changed("is-global") {
		attributes["is-global"] = opts.IsGlobal
	}

	relationships := map[string]any{}
	if orgType != "" && orgID != "" {
		relationships["organization"] = map[string]any{
			"data": map[string]string{
				"type": orgType,
				"id":   orgID,
			},
		}
	}
	if strings.TrimSpace(opts.Broker) != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]string{
				"type": "brokers",
				"id":   strings.TrimSpace(opts.Broker),
			},
		}
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   strings.TrimSpace(opts.CreatedBy),
			},
		}
	}

	data := map[string]any{
		"type":       "organization-invoices-batch-pdf-templates",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-pdf-templates", jsonBody)
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

	row := buildOrganizationInvoicesBatchPdfTemplateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch PDF template %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchPdfTemplatesCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchPdfTemplatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	template, _ := cmd.Flags().GetString("template")
	description, _ := cmd.Flags().GetString("description")
	isActive, _ := cmd.Flags().GetBool("is-active")
	isGlobal, _ := cmd.Flags().GetBool("is-global")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchPdfTemplatesCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Template:         template,
		Description:      description,
		IsActive:         isActive,
		IsGlobal:         isGlobal,
		Organization:     organization,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
		Broker:           broker,
		CreatedBy:        createdBy,
	}, nil
}

func resolveOrganizationInvoicesBatchPdfTemplateOrganization(cmd *cobra.Command, organization, orgType, orgID string) (string, string, error) {
	if cmd.Flags().Changed("organization") {
		return parseOrganization(organization)
	}
	if cmd.Flags().Changed("organization-type") || cmd.Flags().Changed("organization-id") {
		if strings.TrimSpace(orgType) == "" || strings.TrimSpace(orgID) == "" {
			return "", "", fmt.Errorf("--organization-type and --organization-id must be provided together")
		}
		return parseOrganization(fmt.Sprintf("%s|%s", orgType, orgID))
	}
	return "", "", nil
}
