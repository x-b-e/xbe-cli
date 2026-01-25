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

type doMaterialSiteReadingMaterialTypesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	MaterialSite string
	MaterialType string
	ExternalID   string
}

func newDoMaterialSiteReadingMaterialTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material site reading material type",
		Long: `Create a material site reading material type.

Required flags:
  --material-site  Material site ID (required)
  --external-id    External identifier from the source system (required)

Optional flags:
  --material-type  Material type ID to associate (optional)`,
		Example: `  # Create a mapping for a material site reading material type
  xbe do material-site-reading-material-types create \\
    --material-site 123 \\
    --external-id \"EXT-100\" \\
    --material-type 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialSiteReadingMaterialTypesCreate,
	}
	initDoMaterialSiteReadingMaterialTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteReadingMaterialTypesCmd.AddCommand(newDoMaterialSiteReadingMaterialTypesCreateCmd())
}

func initDoMaterialSiteReadingMaterialTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("external-id", "", "External identifier (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("material-site")
	cmd.MarkFlagRequired("external-id")
}

func runDoMaterialSiteReadingMaterialTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSiteReadingMaterialTypesCreateOptions(cmd)
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

	attributes := map[string]any{
		"external-id": opts.ExternalID,
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		},
	}

	if opts.MaterialType != "" {
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-site-reading-material-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-reading-material-types", jsonBody)
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

	row := materialSiteReadingMaterialTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site reading material type %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteReadingMaterialTypesCreateOptions(cmd *cobra.Command) (doMaterialSiteReadingMaterialTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	externalID, _ := cmd.Flags().GetString("external-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteReadingMaterialTypesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		MaterialSite: materialSite,
		MaterialType: materialType,
		ExternalID:   externalID,
	}, nil
}

func materialSiteReadingMaterialTypeRowFromSingle(resp jsonAPISingleResponse) materialSiteReadingMaterialTypeRow {
	return materialSiteReadingMaterialTypeRow{
		ID:         resp.Data.ID,
		ExternalID: stringAttr(resp.Data.Attributes, "external-id"),
	}
}
