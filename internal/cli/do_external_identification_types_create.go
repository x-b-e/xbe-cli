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

type doExternalIdentificationTypesCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	Name                        string
	CanApplyTo                  []string
	FormatValidationRegex       string
	ValueShouldBeGloballyUnique bool
}

func newDoExternalIdentificationTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new external identification type",
		Long: `Create a new external identification type.

Required flags:
  --name          The identification type name (required)
  --can-apply-to  Entity types this can apply to (required, comma-separated)

Optional flags:
  --format-validation-regex        Regex pattern for validating values
  --value-should-be-globally-unique  Whether values must be globally unique`,
		Example: `  # Create a basic external identification type
  xbe do external-identification-types create --name "Employee ID" --can-apply-to User

  # Create with validation regex
  xbe do external-identification-types create --name "License Number" --can-apply-to Trucker --format-validation-regex "^[A-Z]{2}[0-9]{6}$"

  # Create with global uniqueness requirement
  xbe do external-identification-types create --name "Tax ID" --can-apply-to Broker --value-should-be-globally-unique

  # Get JSON output
  xbe do external-identification-types create --name "Test ID" --can-apply-to User --json`,
		Args: cobra.NoArgs,
		RunE: runDoExternalIdentificationTypesCreate,
	}
	initDoExternalIdentificationTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doExternalIdentificationTypesCmd.AddCommand(newDoExternalIdentificationTypesCreateCmd())
}

func initDoExternalIdentificationTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Identification type name (required)")
	cmd.Flags().StringSlice("can-apply-to", nil, "Entity types this can apply to (required, comma-separated)")
	cmd.Flags().String("format-validation-regex", "", "Regex pattern for validating values")
	cmd.Flags().Bool("value-should-be-globally-unique", false, "Whether values must be globally unique")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExternalIdentificationTypesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExternalIdentificationTypesCreateOptions(cmd)
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require can-apply-to
	if len(opts.CanApplyTo) == 0 {
		err := fmt.Errorf("--can-apply-to is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name":         opts.Name,
		"can-apply-to": opts.CanApplyTo,
	}
	if opts.FormatValidationRegex != "" {
		attributes["format-validation-regex"] = opts.FormatValidationRegex
	}
	if cmd.Flags().Changed("value-should-be-globally-unique") {
		attributes["value-should-be-globally-unique"] = opts.ValueShouldBeGloballyUnique
	}

	requestBody := map[string]any{
		"data": map[string]any{
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

	body, _, err := client.Post(cmd.Context(), "/v1/external-identification-types", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created external identification type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoExternalIdentificationTypesCreateOptions(cmd *cobra.Command) (doExternalIdentificationTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	canApplyTo, _ := cmd.Flags().GetStringSlice("can-apply-to")
	formatValidationRegex, _ := cmd.Flags().GetString("format-validation-regex")
	valueShouldBeGloballyUnique, _ := cmd.Flags().GetBool("value-should-be-globally-unique")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExternalIdentificationTypesCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		Name:                        name,
		CanApplyTo:                  canApplyTo,
		FormatValidationRegex:       formatValidationRegex,
		ValueShouldBeGloballyUnique: valueShouldBeGloballyUnique,
	}, nil
}

type externalIdentificationTypeRow struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name"`
	CanApplyTo                  []string `json:"can_apply_to,omitempty"`
	FormatValidationRegex       string   `json:"format_validation_regex,omitempty"`
	ValueShouldBeGloballyUnique bool     `json:"value_should_be_globally_unique"`
	CanDelete                   bool     `json:"can_delete"`
}

func buildExternalIdentificationTypeRow(resp jsonAPISingleResponse) externalIdentificationTypeRow {
	attrs := resp.Data.Attributes

	return externalIdentificationTypeRow{
		ID:                          resp.Data.ID,
		Name:                        stringAttr(attrs, "name"),
		CanApplyTo:                  stringSliceAttr(attrs, "can-apply-to"),
		FormatValidationRegex:       stringAttr(attrs, "format-validation-regex"),
		ValueShouldBeGloballyUnique: boolAttr(attrs, "value-should-be-globally-unique"),
		CanDelete:                   boolAttr(attrs, "can-delete"),
	}
}
