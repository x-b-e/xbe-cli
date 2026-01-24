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

type doOrganizationFormattersCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	FormatterType     string
	FormatterFunction string
	Description       string
	Status            string
	IsLibrary         bool
	MimeTypes         string
	Organization      string
}

func newDoOrganizationFormattersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new organization formatter",
		Long: `Create a new organization formatter.

Required flags:
  --formatter-type      Formatter type (STI class name, e.g., TimeSheetsExportFormatter)
  --organization        Organization in Type|ID format (required, e.g. Broker|123)
  --formatter-function  JavaScript formatter function source

Optional flags:
  --description  Formatter description
  --status       Formatter status (active/inactive)
  --is-library   Mark formatter as a shared library
  --mime-types   Supported MIME types (comma-separated, e.g., text/csv,application/json)`,
		Example: `  # Create a formatter for time sheets
  xbe do organization-formatters create --formatter-type TimeSheetsExportFormatter \\
    --organization Broker|123 \\
    --formatter-function 'function format(lineItemsJson, timestamp) { return lineItemsJson; }'

  # Create with description and mime types
  xbe do organization-formatters create --formatter-type TimeSheetsExportFormatter \\
    --organization Broker|123 \\
    --description \"Time sheets export\" \\
    --mime-types \"text/csv\" \\
    --formatter-function 'function format(lineItemsJson, timestamp) { return lineItemsJson; }'`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationFormattersCreate,
	}
	initDoOrganizationFormattersCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationFormattersCmd.AddCommand(newDoOrganizationFormattersCreateCmd())
}

func initDoOrganizationFormattersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("formatter-type", "", "Formatter type (required)")
	cmd.Flags().String("formatter-function", "", "Formatter function source (required)")
	cmd.Flags().String("description", "", "Formatter description")
	cmd.Flags().String("status", "", "Formatter status (active/inactive)")
	cmd.Flags().Bool("is-library", false, "Mark formatter as a shared library")
	cmd.Flags().String("mime-types", "", "Supported MIME types (comma-separated)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (required, e.g. Broker|123)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationFormattersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationFormattersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.FormatterType) == "" {
		err := fmt.Errorf("--formatter-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Organization) == "" {
		err := fmt.Errorf("--organization is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.FormatterFunction) == "" {
		err := fmt.Errorf("--formatter-function is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	orgType, orgID, err := parseOrganization(opts.Organization)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"formatter-type":     opts.FormatterType,
		"formatter-function": opts.FormatterFunction,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("is-library") {
		attributes["is-library"] = opts.IsLibrary
	}
	if cmd.Flags().Changed("mime-types") {
		mimeTypes := parseMimeTypes(opts.MimeTypes)
		if len(mimeTypes) > 0 {
			attributes["mime-types"] = mimeTypes
		}
	}

	data := map[string]any{
		"type":       "organization-formatters",
		"attributes": attributes,
		"relationships": map[string]any{
			"organization": map[string]any{
				"data": map[string]string{
					"type": orgType,
					"id":   orgID,
				},
			},
		},
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

	body, _, err := client.Post(cmd.Context(), "/v1/organization-formatters", jsonBody)
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

	row := buildOrganizationFormatterRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization formatter %s (%s)\n", row.ID, row.Description)
	return nil
}

func parseDoOrganizationFormattersCreateOptions(cmd *cobra.Command) (doOrganizationFormattersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	formatterType, _ := cmd.Flags().GetString("formatter-type")
	formatterFunction, _ := cmd.Flags().GetString("formatter-function")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	isLibrary, _ := cmd.Flags().GetBool("is-library")
	mimeTypes, _ := cmd.Flags().GetString("mime-types")
	organization, _ := cmd.Flags().GetString("organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationFormattersCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		FormatterType:     formatterType,
		FormatterFunction: formatterFunction,
		Description:       description,
		Status:            status,
		IsLibrary:         isLibrary,
		MimeTypes:         mimeTypes,
		Organization:      organization,
	}, nil
}

func parseMimeTypes(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
