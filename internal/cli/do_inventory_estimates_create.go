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

type doInventoryEstimatesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	MaterialSite string
	MaterialType string
	EstimatedAt  string
	AmountTons   string
	Description  string
}

func newDoInventoryEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an inventory estimate",
		Long: `Create an inventory estimate.

Required flags:
  --material-site   Material site ID
  --material-type   Material type ID
  --estimated-at    Estimated timestamp (ISO 8601)
  --amount-tons     Estimated amount in tons (>= 0)

Optional flags:
  --description     Description or notes

Note: Material types must be generic or related to the material supplier of the site.`,
		Example: `  # Create an inventory estimate
  xbe do inventory-estimates create --material-site 123 --material-type 456 \
    --estimated-at 2025-01-05T08:00:00Z --amount-tons 125.5

  # Create with description
  xbe do inventory-estimates create --material-site 123 --material-type 456 \
    --estimated-at 2025-01-05T08:00:00Z --amount-tons 125.5 --description "Morning estimate"

  # Get JSON output
  xbe do inventory-estimates create --material-site 123 --material-type 456 \
    --estimated-at 2025-01-05T08:00:00Z --amount-tons 125.5 --json`,
		Args: cobra.NoArgs,
		RunE: runDoInventoryEstimatesCreate,
	}
	initDoInventoryEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doInventoryEstimatesCmd.AddCommand(newDoInventoryEstimatesCreateCmd())
}

func initDoInventoryEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("estimated-at", "", "Estimated timestamp (ISO 8601, required)")
	cmd.Flags().String("amount-tons", "", "Estimated amount in tons (required)")
	cmd.Flags().String("description", "", "Description or notes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("material-site")
	cmd.MarkFlagRequired("material-type")
	cmd.MarkFlagRequired("estimated-at")
	cmd.MarkFlagRequired("amount-tons")
}

func runDoInventoryEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInventoryEstimatesCreateOptions(cmd)
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
		"estimated-at": opts.EstimatedAt,
		"amount-tons":  opts.AmountTons,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
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
			"type":          "inventory-estimates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/inventory-estimates", jsonBody)
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

	row := buildInventoryEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created inventory estimate %s\n", row.ID)
	return nil
}

func parseDoInventoryEstimatesCreateOptions(cmd *cobra.Command) (doInventoryEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	estimatedAt, _ := cmd.Flags().GetString("estimated-at")
	amountTons, _ := cmd.Flags().GetString("amount-tons")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInventoryEstimatesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		MaterialSite: materialSite,
		MaterialType: materialType,
		EstimatedAt:  estimatedAt,
		AmountTons:   amountTons,
		Description:  description,
	}, nil
}
