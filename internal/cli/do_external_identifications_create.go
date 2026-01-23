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

type doExternalIdentificationsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ExternalIdentificationTypeID string
	IdentifiesType               string
	IdentifiesID                 string
	Value                        string
	SkipValueValidation          bool
}

func newDoExternalIdentificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new external identification",
		Long: `Create a new external identification.

Required flags:
  --external-identification-type   External identification type ID (required)
  --identifies-type                Type of entity to identify (e.g., truckers) (required)
  --identifies-id                  ID of entity to identify (required)
  --value                          The external identification value (required)

Optional flags:
  --skip-value-validation    Skip format and uniqueness validation`,
		Example: `  # Create an external identification for a trucker
  xbe do external-identifications create \
    --external-identification-type 123 \
    --identifies-type truckers \
    --identifies-id 456 \
    --value "DL-12345"

  # Create with validation skipped
  xbe do external-identifications create \
    --external-identification-type 123 \
    --identifies-type truckers \
    --identifies-id 456 \
    --value "TEMP-ID" \
    --skip-value-validation`,
		Args: cobra.NoArgs,
		RunE: runDoExternalIdentificationsCreate,
	}
	initDoExternalIdentificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doExternalIdentificationsCmd.AddCommand(newDoExternalIdentificationsCreateCmd())
}

func initDoExternalIdentificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-identification-type", "", "External identification type ID (required)")
	cmd.Flags().String("identifies-type", "", "Type of entity to identify (e.g., truckers) (required)")
	cmd.Flags().String("identifies-id", "", "ID of entity to identify (required)")
	cmd.Flags().String("value", "", "The external identification value (required)")
	cmd.Flags().Bool("skip-value-validation", false, "Skip format and uniqueness validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExternalIdentificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExternalIdentificationsCreateOptions(cmd)
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

	if opts.ExternalIdentificationTypeID == "" {
		err := fmt.Errorf("--external-identification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.IdentifiesType == "" {
		err := fmt.Errorf("--identifies-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.IdentifiesID == "" {
		err := fmt.Errorf("--identifies-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Value == "" {
		err := fmt.Errorf("--value is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"value": opts.Value,
	}

	if opts.SkipValueValidation {
		attributes["skip-value-validation"] = true
	}

	relationships := map[string]any{
		"external-identification-type": map[string]any{
			"data": map[string]any{
				"type": "external-identification-types",
				"id":   opts.ExternalIdentificationTypeID,
			},
		},
		"identifies": map[string]any{
			"data": map[string]any{
				"type": opts.IdentifiesType,
				"id":   opts.IdentifiesID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "external-identifications",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/external-identifications", jsonBody)
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

	row := buildExternalIdentificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created external identification %s\n", row.ID)
	return nil
}

func parseDoExternalIdentificationsCreateOptions(cmd *cobra.Command) (doExternalIdentificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalIdentificationTypeID, _ := cmd.Flags().GetString("external-identification-type")
	identifiesType, _ := cmd.Flags().GetString("identifies-type")
	identifiesID, _ := cmd.Flags().GetString("identifies-id")
	value, _ := cmd.Flags().GetString("value")
	skipValueValidation, _ := cmd.Flags().GetBool("skip-value-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExternalIdentificationsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ExternalIdentificationTypeID: externalIdentificationTypeID,
		IdentifiesType:               identifiesType,
		IdentifiesID:                 identifiesID,
		Value:                        value,
		SkipValueValidation:          skipValueValidation,
	}, nil
}

func buildExternalIdentificationRowFromSingle(resp jsonAPISingleResponse) externalIdentificationRow {
	attrs := resp.Data.Attributes

	row := externalIdentificationRow{
		ID:                  resp.Data.ID,
		Value:               stringAttr(attrs, "value"),
		SkipValueValidation: boolAttr(attrs, "skip-value-validation"),
	}

	if rel, ok := resp.Data.Relationships["external-identification-type"]; ok && rel.Data != nil {
		row.ExternalIdentificationTypeID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["identifies"]; ok && rel.Data != nil {
		row.IdentifiesType = rel.Data.Type
		row.IdentifiesID = rel.Data.ID
	}

	return row
}
