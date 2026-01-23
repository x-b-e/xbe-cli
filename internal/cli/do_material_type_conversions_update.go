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

type doMaterialTypeConversionsUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
	MaterialSite            string
	MaterialType            string
	ForeignMaterialSupplier string
	ForeignMaterialSite     string
	ForeignMaterialType     string
}

func newDoMaterialTypeConversionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material type conversion",
		Long: `Update a material type conversion.

Writable relationships:
  --material-type             Local material type ID
  --material-site             Local material site ID (optional; pass empty to clear)
  --foreign-material-supplier Foreign material supplier ID
  --foreign-material-site     Foreign material site ID (optional; pass empty to clear)
  --foreign-material-type     Foreign material type ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update local and foreign mapping
  xbe do material-type-conversions update 123 \
    --material-type 456 \
    --foreign-material-supplier 789 \
    --foreign-material-type 987

  # Clear optional material sites
  xbe do material-type-conversions update 123 \
    --material-site "" \
    --foreign-material-site ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTypeConversionsUpdate,
	}
	initDoMaterialTypeConversionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeConversionsCmd.AddCommand(newDoMaterialTypeConversionsUpdateCmd())
}

func initDoMaterialTypeConversionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Local material site ID (optional; pass empty to clear)")
	cmd.Flags().String("material-type", "", "Local material type ID")
	cmd.Flags().String("foreign-material-supplier", "", "Foreign material supplier ID")
	cmd.Flags().String("foreign-material-site", "", "Foreign material site ID (optional; pass empty to clear)")
	cmd.Flags().String("foreign-material-type", "", "Foreign material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTypeConversionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTypeConversionsUpdateOptions(cmd, args)
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
		return fmt.Errorf("material type conversion id is required")
	}

	relationships := map[string]any{}

	if cmd.Flags().Changed("material-site") {
		if opts.MaterialSite == "" {
			relationships["material-site"] = map[string]any{"data": nil}
		} else {
			relationships["material-site"] = map[string]any{
				"data": map[string]any{
					"type": "material-sites",
					"id":   opts.MaterialSite,
				},
			}
		}
	}

	if cmd.Flags().Changed("material-type") {
		if opts.MaterialType == "" {
			err := fmt.Errorf("--material-type cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		}
	}

	if cmd.Flags().Changed("foreign-material-supplier") {
		if opts.ForeignMaterialSupplier == "" {
			err := fmt.Errorf("--foreign-material-supplier cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["foreign-material-supplier"] = map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.ForeignMaterialSupplier,
			},
		}
	}

	if cmd.Flags().Changed("foreign-material-site") {
		if opts.ForeignMaterialSite == "" {
			relationships["foreign-material-site"] = map[string]any{"data": nil}
		} else {
			relationships["foreign-material-site"] = map[string]any{
				"data": map[string]any{
					"type": "material-sites",
					"id":   opts.ForeignMaterialSite,
				},
			}
		}
	}

	if cmd.Flags().Changed("foreign-material-type") {
		if opts.ForeignMaterialType == "" {
			err := fmt.Errorf("--foreign-material-type cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["foreign-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.ForeignMaterialType,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-type-conversions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-type-conversions/"+id, jsonBody)
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
		details := buildMaterialTypeConversionDetails(resp)
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material type conversion %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialTypeConversionsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTypeConversionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	foreignMaterialSupplier, _ := cmd.Flags().GetString("foreign-material-supplier")
	foreignMaterialSite, _ := cmd.Flags().GetString("foreign-material-site")
	foreignMaterialType, _ := cmd.Flags().GetString("foreign-material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeConversionsUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
		MaterialSite:            materialSite,
		MaterialType:            materialType,
		ForeignMaterialSupplier: foreignMaterialSupplier,
		ForeignMaterialSite:     foreignMaterialSite,
		ForeignMaterialType:     foreignMaterialType,
	}, nil
}
