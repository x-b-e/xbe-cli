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

type doMaterialSiteReadingsUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	MaterialSiteReadingMaterialType    string
	MaterialSiteReadingRawMaterialType string
}

func newDoMaterialSiteReadingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material site reading",
		Long: `Update a material site reading.

Writable relationships:
  --material-site-reading-material-type     Material site reading material type ID
  --material-site-reading-raw-material-type Material site reading raw material type ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update the material type relationship
  xbe do material-site-readings update 123 --material-site-reading-material-type 456

  # Update the raw material type relationship
  xbe do material-site-readings update 123 --material-site-reading-raw-material-type 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSiteReadingsUpdate,
	}
	initDoMaterialSiteReadingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteReadingsCmd.AddCommand(newDoMaterialSiteReadingsUpdateCmd())
}

func initDoMaterialSiteReadingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site-reading-material-type", "", "Material site reading material type ID")
	cmd.Flags().String("material-site-reading-raw-material-type", "", "Material site reading raw material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSiteReadingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSiteReadingsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("material site reading id is required")
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("material-site-reading-material-type") {
		if strings.TrimSpace(opts.MaterialSiteReadingMaterialType) == "" {
			err := fmt.Errorf("--material-site-reading-material-type cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-site-reading-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-site-reading-material-types",
				"id":   opts.MaterialSiteReadingMaterialType,
			},
		}
	}
	if cmd.Flags().Changed("material-site-reading-raw-material-type") {
		if strings.TrimSpace(opts.MaterialSiteReadingRawMaterialType) == "" {
			err := fmt.Errorf("--material-site-reading-raw-material-type cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-site-reading-raw-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-site-reading-material-types",
				"id":   opts.MaterialSiteReadingRawMaterialType,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-site-readings",
			"id":            id,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-site-readings/"+id, jsonBody)
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

	if opts.JSON {
		row := materialSiteReadingRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material site reading %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialSiteReadingsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialSiteReadingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSiteReadingMaterialType, _ := cmd.Flags().GetString("material-site-reading-material-type")
	materialSiteReadingRawMaterialType, _ := cmd.Flags().GetString("material-site-reading-raw-material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteReadingsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		MaterialSiteReadingMaterialType:    materialSiteReadingMaterialType,
		MaterialSiteReadingRawMaterialType: materialSiteReadingRawMaterialType,
	}, nil
}
