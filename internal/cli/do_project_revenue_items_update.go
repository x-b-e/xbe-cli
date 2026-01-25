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

type doProjectRevenueItemsUpdateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	ID                             string
	Description                    string
	ExternalDeveloperRevenueItemID string
	DeveloperQuantityEstimate      string
	RevenueClassification          string
	UnitOfMeasure                  string
	QuantityEstimate               string
	PriceEstimate                  string
}

func newDoProjectRevenueItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project revenue item",
		Long: `Update a project revenue item.

Note: project cannot be changed after creation.

Updatable fields:
  --description                        Revenue item description (use empty to clear)
  --external-developer-revenue-item-id External developer revenue item ID (use empty to clear)
  --developer-quantity-estimate        Developer quantity estimate (use empty to clear)
  --revenue-classification             Revenue classification ID
  --unit-of-measure                     Unit of measure ID
  --quantity-estimate                   Quantity estimate ID (use empty to clear)
  --price-estimate                      Price estimate ID (use empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update description
  xbe do project-revenue-items update 123 --description "Updated description"

  # Update revenue classification and unit of measure
  xbe do project-revenue-items update 123 --revenue-classification 456 --unit-of-measure 789

  # Clear external developer revenue item ID
  xbe do project-revenue-items update 123 --external-developer-revenue-item-id ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectRevenueItemsUpdate,
	}
	initDoProjectRevenueItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueItemsCmd.AddCommand(newDoProjectRevenueItemsUpdateCmd())
}

func initDoProjectRevenueItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Revenue item description (use empty to clear)")
	cmd.Flags().String("external-developer-revenue-item-id", "", "External developer revenue item ID (use empty to clear)")
	cmd.Flags().String("developer-quantity-estimate", "", "Developer quantity estimate (use empty to clear)")
	cmd.Flags().String("revenue-classification", "", "Revenue classification ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("quantity-estimate", "", "Quantity estimate ID (use empty to clear)")
	cmd.Flags().String("price-estimate", "", "Price estimate ID (use empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRevenueItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectRevenueItemsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("external-developer-revenue-item-id") {
		attributes["external-developer-revenue-item-id"] = opts.ExternalDeveloperRevenueItemID
	}
	if cmd.Flags().Changed("developer-quantity-estimate") {
		attributes["developer-quantity-estimate"] = opts.DeveloperQuantityEstimate
	}

	if cmd.Flags().Changed("revenue-classification") {
		if strings.TrimSpace(opts.RevenueClassification) == "" {
			relationships["revenue-classification"] = map[string]any{"data": nil}
		} else {
			relationships["revenue-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-revenue-classifications",
					"id":   opts.RevenueClassification,
				},
			}
		}
	}

	if cmd.Flags().Changed("unit-of-measure") {
		if strings.TrimSpace(opts.UnitOfMeasure) == "" {
			relationships["unit-of-measure"] = map[string]any{"data": nil}
		} else {
			relationships["unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasure,
				},
			}
		}
	}

	if cmd.Flags().Changed("quantity-estimate") {
		if strings.TrimSpace(opts.QuantityEstimate) == "" {
			relationships["quantity-estimate"] = map[string]any{"data": nil}
		} else {
			relationships["quantity-estimate"] = map[string]any{
				"data": map[string]any{
					"type": "project-revenue-item-quantity-estimates",
					"id":   opts.QuantityEstimate,
				},
			}
		}
	}

	if cmd.Flags().Changed("price-estimate") {
		if strings.TrimSpace(opts.PriceEstimate) == "" {
			relationships["price-estimate"] = map[string]any{"data": nil}
		} else {
			relationships["price-estimate"] = map[string]any{
				"data": map[string]any{
					"type": "project-revenue-item-price-estimates",
					"id":   opts.PriceEstimate,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-revenue-items",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-revenue-items/"+opts.ID, jsonBody)
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

	row := buildProjectRevenueItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project revenue item %s\n", row.ID)
	return nil
}

func parseDoProjectRevenueItemsUpdateOptions(cmd *cobra.Command, args []string) (doProjectRevenueItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	externalDeveloperRevenueItemID, _ := cmd.Flags().GetString("external-developer-revenue-item-id")
	developerQuantityEstimate, _ := cmd.Flags().GetString("developer-quantity-estimate")
	revenueClassification, _ := cmd.Flags().GetString("revenue-classification")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantityEstimate, _ := cmd.Flags().GetString("quantity-estimate")
	priceEstimate, _ := cmd.Flags().GetString("price-estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueItemsUpdateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		ID:                             args[0],
		Description:                    description,
		ExternalDeveloperRevenueItemID: externalDeveloperRevenueItemID,
		DeveloperQuantityEstimate:      developerQuantityEstimate,
		RevenueClassification:          revenueClassification,
		UnitOfMeasure:                  unitOfMeasure,
		QuantityEstimate:               quantityEstimate,
		PriceEstimate:                  priceEstimate,
	}, nil
}
