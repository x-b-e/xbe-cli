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

type doInventoryChangesCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	MaterialSiteID  string
	MaterialTypeID  string
	EstimateAt      string
	ForecastStartAt string
}

func newDoInventoryChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an inventory change",
		Long: `Create an inventory change.

Required flags:
  --material-site   Material site ID
  --material-type   Material type ID
  --estimate-at     Estimate timestamp (ISO 8601)

Optional flags:
  --forecast-start-at  Forecast start timestamp (ISO 8601, must be before estimate-at)`,
		Example: `  # Create an inventory change for a site/type
  xbe do inventory-changes create --material-site 123 --material-type 456 --estimate-at 2025-01-01T12:00:00Z

  # Create with a forecast window
  xbe do inventory-changes create --material-site 123 --material-type 456 --estimate-at 2025-01-01T12:00:00Z --forecast-start-at 2025-01-01T00:00:00Z`,
		RunE: runDoInventoryChangesCreate,
	}
	initDoInventoryChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doInventoryChangesCmd.AddCommand(newDoInventoryChangesCreateCmd())
}

func initDoInventoryChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("estimate-at", "", "Estimate timestamp (ISO 8601, required)")
	cmd.Flags().String("forecast-start-at", "", "Forecast start timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("material-site")
	cmd.MarkFlagRequired("material-type")
	cmd.MarkFlagRequired("estimate-at")
}

func runDoInventoryChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInventoryChangesCreateOptions(cmd)
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
		"estimate-at": opts.EstimateAt,
	}
	if cmd.Flags().Changed("forecast-start-at") {
		attributes["forecast-start-at"] = opts.ForecastStartAt
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSiteID,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialTypeID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "inventory-changes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/inventory-changes", jsonBody)
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

	row := inventoryChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created inventory change %s\n", row.ID)
	return nil
}

func parseDoInventoryChangesCreateOptions(cmd *cobra.Command) (doInventoryChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	materialTypeID, _ := cmd.Flags().GetString("material-type")
	estimateAt, _ := cmd.Flags().GetString("estimate-at")
	forecastStartAt, _ := cmd.Flags().GetString("forecast-start-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInventoryChangesCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		MaterialSiteID:  materialSiteID,
		MaterialTypeID:  materialTypeID,
		EstimateAt:      estimateAt,
		ForecastStartAt: forecastStartAt,
	}, nil
}

func inventoryChangeRowFromSingle(resp jsonAPISingleResponse) inventoryChangeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildInventoryChangeRowFromResource(resp.Data, included)
}
