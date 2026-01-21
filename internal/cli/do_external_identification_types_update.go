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

type doExternalIdentificationTypesUpdateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	Name                        string
	CanApplyTo                  []string
	FormatValidationRegex       string
	ValueShouldBeGloballyUnique bool
}

func newDoExternalIdentificationTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an external identification type",
		Long: `Update an existing external identification type.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The identification type ID (required)

Flags:
  --name                             Update the name
  --can-apply-to                     Update entity types this can apply to
  --format-validation-regex          Update the validation regex
  --value-should-be-globally-unique  Update global uniqueness requirement`,
		Example: `  # Update just the name
  xbe do external-identification-types update 123 --name "Updated Name"

  # Update validation regex
  xbe do external-identification-types update 123 --format-validation-regex "^[0-9]+$"

  # Get JSON output
  xbe do external-identification-types update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoExternalIdentificationTypesUpdate,
	}
	initDoExternalIdentificationTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doExternalIdentificationTypesCmd.AddCommand(newDoExternalIdentificationTypesUpdateCmd())
}

func initDoExternalIdentificationTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().StringSlice("can-apply-to", nil, "New entity types this can apply to")
	cmd.Flags().String("format-validation-regex", "", "New validation regex")
	cmd.Flags().Bool("value-should-be-globally-unique", false, "New global uniqueness requirement")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExternalIdentificationTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExternalIdentificationTypesUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("external identification type id is required")
	}

	// Check if at least one field is being updated
	hasUpdate := opts.Name != "" || opts.FormatValidationRegex != "" ||
		cmd.Flags().Changed("can-apply-to") ||
		cmd.Flags().Changed("value-should-be-globally-unique")

	if !hasUpdate {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("can-apply-to") {
		attributes["can-apply-to"] = opts.CanApplyTo
	}
	if opts.FormatValidationRegex != "" {
		attributes["format-validation-regex"] = opts.FormatValidationRegex
	}
	if cmd.Flags().Changed("value-should-be-globally-unique") {
		attributes["value-should-be-globally-unique"] = opts.ValueShouldBeGloballyUnique
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "external-identification-types",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/external-identification-types/"+id, jsonBody)
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

	row := buildExternalIdentificationTypeRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated external identification type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoExternalIdentificationTypesUpdateOptions(cmd *cobra.Command) (doExternalIdentificationTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	canApplyTo, _ := cmd.Flags().GetStringSlice("can-apply-to")
	formatValidationRegex, _ := cmd.Flags().GetString("format-validation-regex")
	valueShouldBeGloballyUnique, _ := cmd.Flags().GetBool("value-should-be-globally-unique")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExternalIdentificationTypesUpdateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		Name:                        name,
		CanApplyTo:                  canApplyTo,
		FormatValidationRegex:       formatValidationRegex,
		ValueShouldBeGloballyUnique: valueShouldBeGloballyUnique,
	}, nil
}
