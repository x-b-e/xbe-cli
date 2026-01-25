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

type doInventoryCapacitiesCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	MaterialSite    string
	MaterialType    string
	MaxCapacityTons string
	MinCapacityTons string
	ThresholdTons   string
}

func newDoInventoryCapacitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an inventory capacity",
		Long: `Create an inventory capacity.

Required flags:
  --material-site   Material site ID
  --material-type   Material type ID

Optional flags:
  --max-capacity-tons  Maximum capacity in tons
  --min-capacity-tons  Minimum capacity in tons
  --threshold-tons     Alert threshold in tons`,
		Example: `  # Create with min/max capacity
  xbe do inventory-capacities create --material-site 123 --material-type 456 \\
    --min-capacity-tons 50 --max-capacity-tons 500

  # Create with threshold only
  xbe do inventory-capacities create --material-site 123 --material-type 456 \\
    --threshold-tons 75`,
		RunE: runDoInventoryCapacitiesCreate,
	}
	initDoInventoryCapacitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doInventoryCapacitiesCmd.AddCommand(newDoInventoryCapacitiesCreateCmd())
}

func initDoInventoryCapacitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("max-capacity-tons", "", "Maximum capacity in tons")
	cmd.Flags().String("min-capacity-tons", "", "Minimum capacity in tons")
	cmd.Flags().String("threshold-tons", "", "Alert threshold in tons")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-site")
	_ = cmd.MarkFlagRequired("material-type")
}

func runDoInventoryCapacitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInventoryCapacitiesCreateOptions(cmd)
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
	if opts.MaxCapacityTons != "" {
		attributes["max-capacity-tons"] = opts.MaxCapacityTons
	}
	if opts.MinCapacityTons != "" {
		attributes["min-capacity-tons"] = opts.MinCapacityTons
	}
	if opts.ThresholdTons != "" {
		attributes["threshold-tons"] = opts.ThresholdTons
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "inventory-capacities",
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

	body, _, err := client.Post(cmd.Context(), "/v1/inventory-capacities", jsonBody)
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

	result := buildInventoryCapacityResult(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created inventory capacity %s\n", result["id"])
	return nil
}

func parseDoInventoryCapacitiesCreateOptions(cmd *cobra.Command) (doInventoryCapacitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	maxCapacityTons, _ := cmd.Flags().GetString("max-capacity-tons")
	minCapacityTons, _ := cmd.Flags().GetString("min-capacity-tons")
	thresholdTons, _ := cmd.Flags().GetString("threshold-tons")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInventoryCapacitiesCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		MaterialSite:    materialSite,
		MaterialType:    materialType,
		MaxCapacityTons: maxCapacityTons,
		MinCapacityTons: minCapacityTons,
		ThresholdTons:   thresholdTons,
	}, nil
}

func buildInventoryCapacityResult(resp jsonAPISingleResponse) map[string]string {
	result := map[string]string{
		"id":                resp.Data.ID,
		"max_capacity_tons": stringAttr(resp.Data.Attributes, "max-capacity-tons"),
		"min_capacity_tons": stringAttr(resp.Data.Attributes, "min-capacity-tons"),
		"threshold_tons":    stringAttr(resp.Data.Attributes, "threshold-tons"),
	}
	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		result["material_site_id"] = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		result["material_type_id"] = rel.Data.ID
	}
	return result
}
