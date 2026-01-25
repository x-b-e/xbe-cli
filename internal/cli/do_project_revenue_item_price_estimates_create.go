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

type doProjectRevenueItemPriceEstimatesCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ProjectRevenueItem   string
	ProjectEstimateSet   string
	Kind                 string
	PricePerUnitExplicit string
	CostMultiplier       string
	CreatedBy            string
}

func newDoProjectRevenueItemPriceEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project revenue item price estimate",
		Long: `Create a project revenue item price estimate.

Required flags:
  --project-revenue-item   Project revenue item ID

Optional flags:
  --project-estimate-set   Project estimate set ID
  --kind                   Estimate kind (explicit or cost_multiplier)
  --price-per-unit-explicit Explicit price per unit (for kind=explicit)
  --cost-multiplier        Cost multiplier (for kind=cost_multiplier)
  --created-by             Created-by user ID (admin only)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an explicit price estimate
  xbe do project-revenue-item-price-estimates create \
    --project-revenue-item 123 \
    --kind explicit \
    --price-per-unit-explicit 45.50

  # Create a cost multiplier estimate
  xbe do project-revenue-item-price-estimates create \
    --project-revenue-item 123 \
    --kind cost_multiplier \
    --cost-multiplier 1.15

  # JSON output
  xbe do project-revenue-item-price-estimates create \
    --project-revenue-item 123 \
    --price-per-unit-explicit 45.50 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectRevenueItemPriceEstimatesCreate,
	}
	initDoProjectRevenueItemPriceEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueItemPriceEstimatesCmd.AddCommand(newDoProjectRevenueItemPriceEstimatesCreateCmd())
}

func initDoProjectRevenueItemPriceEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-revenue-item", "", "Project revenue item ID (required)")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID")
	cmd.Flags().String("kind", "", "Estimate kind (explicit or cost_multiplier)")
	cmd.Flags().String("price-per-unit-explicit", "", "Explicit price per unit")
	cmd.Flags().String("cost-multiplier", "", "Cost multiplier")
	cmd.Flags().String("created-by", "", "Created-by user ID (admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-revenue-item")
}

func runDoProjectRevenueItemPriceEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectRevenueItemPriceEstimatesCreateOptions(cmd)
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
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("price-per-unit-explicit") {
		attributes["price-per-unit-explicit"] = opts.PricePerUnitExplicit
	}
	if cmd.Flags().Changed("cost-multiplier") {
		attributes["cost-multiplier"] = opts.CostMultiplier
	}

	relationships := map[string]any{
		"project-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-revenue-items",
				"id":   opts.ProjectRevenueItem,
			},
		},
	}

	if opts.ProjectEstimateSet != "" {
		relationships["project-estimate-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   opts.ProjectEstimateSet,
			},
		}
	}
	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-revenue-item-price-estimates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-revenue-item-price-estimates", jsonBody)
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

	row := projectRevenueItemPriceEstimateRow{
		ID:                    resp.Data.ID,
		Kind:                  stringAttr(resp.Data.Attributes, "kind"),
		PricePerUnitExplicit:  stringAttr(resp.Data.Attributes, "price-per-unit-explicit"),
		CostMultiplier:        stringAttr(resp.Data.Attributes, "cost-multiplier"),
		PricePerUnit:          stringAttr(resp.Data.Attributes, "price-per-unit"),
		PricePerUnitEffective: stringAttr(resp.Data.Attributes, "price-per-unit-effective"),
	}
	if rel, ok := resp.Data.Relationships["project-revenue-item"]; ok && rel.Data != nil {
		row.ProjectRevenueItemID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		row.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project revenue item price estimate %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectRevenueItemPriceEstimatesCreateOptions(cmd *cobra.Command) (doProjectRevenueItemPriceEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectRevenueItem, _ := cmd.Flags().GetString("project-revenue-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	kind, _ := cmd.Flags().GetString("kind")
	pricePerUnitExplicit, _ := cmd.Flags().GetString("price-per-unit-explicit")
	costMultiplier, _ := cmd.Flags().GetString("cost-multiplier")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueItemPriceEstimatesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ProjectRevenueItem:   projectRevenueItem,
		ProjectEstimateSet:   projectEstimateSet,
		Kind:                 kind,
		PricePerUnitExplicit: pricePerUnitExplicit,
		CostMultiplier:       costMultiplier,
		CreatedBy:            createdBy,
	}, nil
}
