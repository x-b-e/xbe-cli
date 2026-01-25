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

type doMaterialTypeConversionsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	MaterialSupplier        string
	MaterialSite            string
	MaterialType            string
	ForeignMaterialSupplier string
	ForeignMaterialSite     string
	ForeignMaterialType     string
}

func newDoMaterialTypeConversionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material type conversion",
		Long: `Create a material type conversion.

Required flags:
  --material-supplier         Local material supplier ID
  --material-type             Local material type ID
  --foreign-material-supplier Foreign material supplier ID
  --foreign-material-type     Foreign material type ID

Optional flags:
  --material-site         Local material site ID
  --foreign-material-site Foreign material site ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a conversion with supplier-level mapping
  xbe do material-type-conversions create \
    --material-supplier 123 \
    --material-type 456 \
    --foreign-material-supplier 789 \
    --foreign-material-type 987

  # Create a conversion with site-specific mapping
  xbe do material-type-conversions create \
    --material-supplier 123 \
    --material-site 321 \
    --material-type 456 \
    --foreign-material-supplier 789 \
    --foreign-material-site 654 \
    --foreign-material-type 987`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTypeConversionsCreate,
	}
	initDoMaterialTypeConversionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeConversionsCmd.AddCommand(newDoMaterialTypeConversionsCreateCmd())
}

func initDoMaterialTypeConversionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-supplier", "", "Local material supplier ID")
	cmd.Flags().String("material-site", "", "Local material site ID")
	cmd.Flags().String("material-type", "", "Local material type ID")
	cmd.Flags().String("foreign-material-supplier", "", "Foreign material supplier ID")
	cmd.Flags().String("foreign-material-site", "", "Foreign material site ID")
	cmd.Flags().String("foreign-material-type", "", "Foreign material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-supplier")
	_ = cmd.MarkFlagRequired("material-type")
	_ = cmd.MarkFlagRequired("foreign-material-supplier")
	_ = cmd.MarkFlagRequired("foreign-material-type")
}

func runDoMaterialTypeConversionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTypeConversionsCreateOptions(cmd)
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

	relationships := map[string]any{
		"material-supplier": map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.MaterialSupplier,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
		"foreign-material-supplier": map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.ForeignMaterialSupplier,
			},
		},
		"foreign-material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.ForeignMaterialType,
			},
		},
	}

	if opts.MaterialSite != "" {
		relationships["material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		}
	}
	if opts.ForeignMaterialSite != "" {
		relationships["foreign-material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.ForeignMaterialSite,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-type-conversions",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-type-conversions", jsonBody)
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

	details := buildMaterialTypeConversionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material type conversion %s\n", details.ID)
	return nil
}

func parseDoMaterialTypeConversionsCreateOptions(cmd *cobra.Command) (doMaterialTypeConversionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	foreignMaterialSupplier, _ := cmd.Flags().GetString("foreign-material-supplier")
	foreignMaterialSite, _ := cmd.Flags().GetString("foreign-material-site")
	foreignMaterialType, _ := cmd.Flags().GetString("foreign-material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeConversionsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		MaterialSupplier:        materialSupplier,
		MaterialSite:            materialSite,
		MaterialType:            materialType,
		ForeignMaterialSupplier: foreignMaterialSupplier,
		ForeignMaterialSite:     foreignMaterialSite,
		ForeignMaterialType:     foreignMaterialType,
	}, nil
}
