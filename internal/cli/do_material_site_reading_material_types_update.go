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

type doMaterialSiteReadingMaterialTypesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	ExternalID   string
	MaterialType string
}

func newDoMaterialSiteReadingMaterialTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material site reading material type",
		Long: `Update a material site reading material type.

All flags are optional. Only provided flags will update the mapping.

Optional flags:
  --external-id    External identifier

Relationships:
  --material-type  Material type ID (set empty to clear)`,
		Example: `  # Update external ID
  xbe do material-site-reading-material-types update 123 --external-id \"EXT-200\"

  # Update material type
  xbe do material-site-reading-material-types update 123 --material-type 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSiteReadingMaterialTypesUpdate,
	}
	initDoMaterialSiteReadingMaterialTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteReadingMaterialTypesCmd.AddCommand(newDoMaterialSiteReadingMaterialTypesUpdateCmd())
}

func initDoMaterialSiteReadingMaterialTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-id", "", "External identifier")
	cmd.Flags().String("material-type", "", "Material type ID (empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSiteReadingMaterialTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSiteReadingMaterialTypesUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("external-id") {
		attributes["external-id"] = opts.ExternalID
	}

	if cmd.Flags().Changed("material-type") {
		if opts.MaterialType == "" {
			relationships["material-type"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["material-type"] = map[string]any{
				"data": map[string]any{
					"type": "material-types",
					"id":   opts.MaterialType,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-site-reading-material-types",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-site-reading-material-types/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material site reading material type %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteReadingMaterialTypesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialSiteReadingMaterialTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalID, _ := cmd.Flags().GetString("external-id")
	materialType, _ := cmd.Flags().GetString("material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteReadingMaterialTypesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		ExternalID:   externalID,
		MaterialType: materialType,
	}, nil
}
