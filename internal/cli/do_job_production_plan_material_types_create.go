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

type doJobProductionPlanMaterialTypesCreateOptions struct {
	BaseURL                                           string
	Token                                             string
	JSON                                              bool
	JobProductionPlan                                 string
	MaterialType                                      string
	UnitOfMeasure                                     string
	MaterialSite                                      string
	DefaultCostCode                                   string
	ExplicitMaterialMixDesign                         string
	ExplicitMaterialTypeMaterialSiteInventoryLocation string
	Quantity                                          string
	IsQuantityUnknown                                 bool
	ExplicitDisplayName                               string
	PlanRequiresSiteSpecificMaterialTypes             bool
	PlanRequiresSupplierSpecificMaterialTypes         bool
}

type jobProductionPlanMaterialTypeCreateResult struct {
	ID          string  `json:"id"`
	Quantity    float64 `json:"quantity,omitempty"`
	DisplayName string  `json:"display_name,omitempty"`
}

func newDoJobProductionPlanMaterialTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan material type",
		Long: `Create a job production plan material type.

Required flags:
  --job-production-plan   Job production plan ID
  --material-type         Material type ID
  --unit-of-measure       Unit of measure ID
  --quantity              Planned quantity (use 0 when unknown)

Optional flags:
  --is-quantity-unknown   Mark quantity as unknown
  --explicit-display-name Display name override
  --plan-requires-site-specific-material-types     Override plan site-specific requirement
  --plan-requires-supplier-specific-material-types Override plan supplier-specific requirement

Relationships:
  --material-site                           Material site ID (must be on the plan)
  --default-cost-code                       Default cost code ID
  --explicit-material-mix-design            Explicit material mix design ID
  --explicit-material-type-material-site-inventory-location Explicit inventory location ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a job production plan material type
  xbe do job-production-plan-material-types create \
    --job-production-plan 123 \
    --material-type 456 \
    --unit-of-measure 10 \
    --quantity 250

  # Create with display name override
  xbe do job-production-plan-material-types create \
    --job-production-plan 123 \
    --material-type 456 \
    --unit-of-measure 10 \
    --quantity 250 \
    --explicit-display-name "Base Gravel"`,
		RunE: runDoJobProductionPlanMaterialTypesCreate,
	}
	initDoJobProductionPlanMaterialTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialTypesCmd.AddCommand(newDoJobProductionPlanMaterialTypesCreateCmd())
}

func initDoJobProductionPlanMaterialTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("quantity", "", "Planned quantity (required)")
	cmd.Flags().Bool("is-quantity-unknown", false, "Mark quantity as unknown")
	cmd.Flags().String("explicit-display-name", "", "Display name override")
	cmd.Flags().Bool("plan-requires-site-specific-material-types", false, "Override plan site-specific requirement")
	cmd.Flags().Bool("plan-requires-supplier-specific-material-types", false, "Override plan supplier-specific requirement")
	cmd.Flags().String("material-site", "", "Material site ID")
	cmd.Flags().String("default-cost-code", "", "Default cost code ID")
	cmd.Flags().String("explicit-material-mix-design", "", "Explicit material mix design ID")
	cmd.Flags().String("explicit-material-type-material-site-inventory-location", "", "Explicit inventory location ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job-production-plan")
	cmd.MarkFlagRequired("material-type")
	cmd.MarkFlagRequired("unit-of-measure")
	cmd.MarkFlagRequired("quantity")
}

func runDoJobProductionPlanMaterialTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanMaterialTypesCreateOptions(cmd)
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
		"quantity": opts.Quantity,
	}

	if cmd.Flags().Changed("is-quantity-unknown") {
		attributes["is-quantity-unknown"] = opts.IsQuantityUnknown
	}
	if opts.ExplicitDisplayName != "" {
		attributes["explicit-display-name"] = opts.ExplicitDisplayName
	}
	if cmd.Flags().Changed("plan-requires-site-specific-material-types") {
		attributes["plan-requires-site-specific-material-types"] = opts.PlanRequiresSiteSpecificMaterialTypes
	}
	if cmd.Flags().Changed("plan-requires-supplier-specific-material-types") {
		attributes["plan-requires-supplier-specific-material-types"] = opts.PlanRequiresSupplierSpecificMaterialTypes
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
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
	if opts.DefaultCostCode != "" {
		relationships["default-cost-code"] = map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   opts.DefaultCostCode,
			},
		}
	}
	if opts.ExplicitMaterialMixDesign != "" {
		relationships["explicit-material-mix-design"] = map[string]any{
			"data": map[string]any{
				"type": "material-mix-designs",
				"id":   opts.ExplicitMaterialMixDesign,
			},
		}
	}
	if opts.ExplicitMaterialTypeMaterialSiteInventoryLocation != "" {
		relationships["explicit-material-type-material-site-inventory-location"] = map[string]any{
			"data": map[string]any{
				"type": "material-type-material-site-inventory-locations",
				"id":   opts.ExplicitMaterialTypeMaterialSiteInventoryLocation,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-material-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-material-types", jsonBody)
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

	result := jobProductionPlanMaterialTypeCreateResult{
		ID:          resp.Data.ID,
		Quantity:    floatAttr(resp.Data.Attributes, "quantity"),
		DisplayName: stringAttr(resp.Data.Attributes, "display-name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan material type %s\n", result.ID)
	return nil
}

func parseDoJobProductionPlanMaterialTypesCreateOptions(cmd *cobra.Command) (doJobProductionPlanMaterialTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	materialType, _ := cmd.Flags().GetString("material-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	materialSite, _ := cmd.Flags().GetString("material-site")
	defaultCostCode, _ := cmd.Flags().GetString("default-cost-code")
	explicitMaterialMixDesign, _ := cmd.Flags().GetString("explicit-material-mix-design")
	explicitInventoryLocation, _ := cmd.Flags().GetString("explicit-material-type-material-site-inventory-location")
	quantity, _ := cmd.Flags().GetString("quantity")
	isQuantityUnknown, _ := cmd.Flags().GetBool("is-quantity-unknown")
	explicitDisplayName, _ := cmd.Flags().GetString("explicit-display-name")
	planRequiresSiteSpecificMaterialTypes, _ := cmd.Flags().GetBool("plan-requires-site-specific-material-types")
	planRequiresSupplierSpecificMaterialTypes, _ := cmd.Flags().GetBool("plan-requires-supplier-specific-material-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialTypesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		JobProductionPlan:         jobProductionPlan,
		MaterialType:              materialType,
		UnitOfMeasure:             unitOfMeasure,
		MaterialSite:              materialSite,
		DefaultCostCode:           defaultCostCode,
		ExplicitMaterialMixDesign: explicitMaterialMixDesign,
		ExplicitMaterialTypeMaterialSiteInventoryLocation: explicitInventoryLocation,
		Quantity:                                  quantity,
		IsQuantityUnknown:                         isQuantityUnknown,
		ExplicitDisplayName:                       explicitDisplayName,
		PlanRequiresSiteSpecificMaterialTypes:     planRequiresSiteSpecificMaterialTypes,
		PlanRequiresSupplierSpecificMaterialTypes: planRequiresSupplierSpecificMaterialTypes,
	}, nil
}
