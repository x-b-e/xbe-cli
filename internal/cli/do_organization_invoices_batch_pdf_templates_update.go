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

type doOrganizationInvoicesBatchPdfTemplatesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Template    string
	Description string
	IsActive    bool
	IsGlobal    bool
}

func newDoOrganizationInvoicesBatchPdfTemplatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an organization invoices batch PDF template",
		Long: `Update an existing organization invoices batch PDF template.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The template ID (required)

Flags:
  --template     Update template content
  --description  Update description
  --is-active    Update active status
  --is-global    Update global status`,
		Example: `  # Update the description
  xbe do organization-invoices-batch-pdf-templates update 456 --description "Updated description"

  # Update template content
  xbe do organization-invoices-batch-pdf-templates update 456 --template "{{invoice_number}}"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOrganizationInvoicesBatchPdfTemplatesUpdate,
	}
	initDoOrganizationInvoicesBatchPdfTemplatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchPdfTemplatesCmd.AddCommand(newDoOrganizationInvoicesBatchPdfTemplatesUpdateCmd())
}

func initDoOrganizationInvoicesBatchPdfTemplatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("template", "", "Template content")
	cmd.Flags().String("description", "", "Template description")
	cmd.Flags().Bool("is-active", false, "Whether the template is active")
	cmd.Flags().Bool("is-global", false, "Whether the template is global")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationInvoicesBatchPdfTemplatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOrganizationInvoicesBatchPdfTemplatesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("template") {
		attributes["template"] = opts.Template
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

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "organization-invoices-batch-pdf-templates",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/organization-invoices-batch-pdf-templates/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated organization invoices batch PDF template %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchPdfTemplatesUpdateOptions(cmd *cobra.Command, args []string) (doOrganizationInvoicesBatchPdfTemplatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	template, _ := cmd.Flags().GetString("template")
	description, _ := cmd.Flags().GetString("description")
	isActive, _ := cmd.Flags().GetBool("is-active")
	isGlobal, _ := cmd.Flags().GetBool("is-global")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchPdfTemplatesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Template:    template,
		Description: description,
		IsActive:    isActive,
		IsGlobal:    isGlobal,
	}, nil
}
