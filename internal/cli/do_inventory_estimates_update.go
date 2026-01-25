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

type doInventoryEstimatesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	MaterialSite string
	MaterialType string
	EstimatedAt  string
	AmountTons   string
	Description  string
}

func newDoInventoryEstimatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an inventory estimate",
		Long: `Update an inventory estimate.

Optional flags:
  --material-site   Material site ID
  --material-type   Material type ID
  --estimated-at    Estimated timestamp (ISO 8601)
  --amount-tons     Estimated amount in tons
  --description     Description or notes`,
		Example: `  # Update amount and description
  xbe do inventory-estimates update 123 --amount-tons 140 --description "Adjusted"

  # Update estimated-at
  xbe do inventory-estimates update 123 --estimated-at 2025-01-06T08:00:00Z

  # Move to a different material site and type
  xbe do inventory-estimates update 123 --material-site 456 --material-type 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoInventoryEstimatesUpdate,
	}
	initDoInventoryEstimatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doInventoryEstimatesCmd.AddCommand(newDoInventoryEstimatesUpdateCmd())
}

func initDoInventoryEstimatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("estimated-at", "", "Estimated timestamp (ISO 8601)")
	cmd.Flags().String("amount-tons", "", "Estimated amount in tons")
	cmd.Flags().String("description", "", "Description or notes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInventoryEstimatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoInventoryEstimatesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("estimated-at") {
		attributes["estimated-at"] = opts.EstimatedAt
	}
	if cmd.Flags().Changed("amount-tons") {
		attributes["amount-tons"] = opts.AmountTons
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

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
			relationships["material-type"] = map[string]any{"data": nil}
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
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "inventory-estimates",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/inventory-estimates/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated inventory estimate %s\n", row.ID)
	return nil
}

func parseDoInventoryEstimatesUpdateOptions(cmd *cobra.Command, args []string) (doInventoryEstimatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	estimatedAt, _ := cmd.Flags().GetString("estimated-at")
	amountTons, _ := cmd.Flags().GetString("amount-tons")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInventoryEstimatesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		MaterialSite: materialSite,
		MaterialType: materialType,
		EstimatedAt:  estimatedAt,
		AmountTons:   amountTons,
		Description:  description,
	}, nil
}
