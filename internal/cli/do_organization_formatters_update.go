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

type doOrganizationFormattersUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	FormatterFunction string
	Description       string
	Status            string
	IsLibrary         bool
	MimeTypes         string
}

func newDoOrganizationFormattersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an organization formatter",
		Long: `Update an existing organization formatter.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The organization formatter ID (required)

Flags:
  --formatter-function  Update the formatter function source
  --description         Update the description
  --status              Update the status (active/inactive)
  --is-library          Update the shared library flag
  --mime-types          Update supported MIME types (comma-separated)`,
		Example: `  # Update the description
  xbe do organization-formatters update 456 --description \"Updated formatter\"

  # Mark formatter as inactive
  xbe do organization-formatters update 456 --status inactive

  # Update formatter function
  xbe do organization-formatters update 456 --formatter-function 'function format(lineItemsJson, timestamp) { return lineItemsJson; }'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOrganizationFormattersUpdate,
	}
	initDoOrganizationFormattersUpdateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationFormattersCmd.AddCommand(newDoOrganizationFormattersUpdateCmd())
}

func initDoOrganizationFormattersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("formatter-function", "", "Formatter function source")
	cmd.Flags().String("description", "", "Formatter description")
	cmd.Flags().String("status", "", "Formatter status (active/inactive)")
	cmd.Flags().Bool("is-library", false, "Mark formatter as a shared library")
	cmd.Flags().String("mime-types", "", "Supported MIME types (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationFormattersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOrganizationFormattersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("formatter-function") {
		attributes["formatter-function"] = opts.FormatterFunction
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("is-library") {
		attributes["is-library"] = opts.IsLibrary
	}
	if cmd.Flags().Changed("mime-types") {
		mimeTypes := parseMimeTypes(opts.MimeTypes)
		attributes["mime-types"] = mimeTypes
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "organization-formatters",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/organization-formatters/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated organization formatter %s (%s)\n", row.ID, row.Description)
	return nil
}

func parseDoOrganizationFormattersUpdateOptions(cmd *cobra.Command, args []string) (doOrganizationFormattersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	formatterFunction, _ := cmd.Flags().GetString("formatter-function")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	isLibrary, _ := cmd.Flags().GetBool("is-library")
	mimeTypes, _ := cmd.Flags().GetString("mime-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationFormattersUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		FormatterFunction: formatterFunction,
		Description:       description,
		Status:            status,
		IsLibrary:         isLibrary,
		MimeTypes:         mimeTypes,
	}, nil
}
