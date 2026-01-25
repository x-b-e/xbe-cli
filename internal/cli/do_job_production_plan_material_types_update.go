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

type doJobProductionPlanMaterialTypesUpdateOptions struct {
	BaseURL                                           string
	Token                                             string
	JSON                                              bool
	ID                                                string
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

func newDoJobProductionPlanMaterialTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan material type",
		Long: `Update a job production plan material type.

Optional flags:
  --quantity              Planned quantity
  --is-quantity-unknown   Mark quantity as unknown
  --explicit-display-name Display name override
  --plan-requires-site-specific-material-types     Override plan site-specific requirement
  --plan-requires-supplier-specific-material-types Override plan supplier-specific requirement

Relationships:
  --job-production-plan                     Job production plan ID
  --material-type                           Material type ID
  --unit-of-measure                         Unit of measure ID
  --material-site                           Material site ID (empty to clear)
  --default-cost-code                       Default cost code ID (empty to clear)
  --explicit-material-mix-design            Explicit material mix design ID (empty to clear)
  --explicit-material-type-material-site-inventory-location Explicit inventory location ID (empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity
  xbe do job-production-plan-material-types update 123 --quantity 300

  # Update material type display name
  xbe do job-production-plan-material-types update 123 --explicit-display-name "Base Gravel"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanMaterialTypesUpdate,
	}
	initDoJobProductionPlanMaterialTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialTypesCmd.AddCommand(newDoJobProductionPlanMaterialTypesUpdateCmd())
}

func initDoJobProductionPlanMaterialTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("material-site", "", "Material site ID (empty to clear)")
	cmd.Flags().String("default-cost-code", "", "Default cost code ID (empty to clear)")
	cmd.Flags().String("explicit-material-mix-design", "", "Explicit material mix design ID (empty to clear)")
	cmd.Flags().String("explicit-material-type-material-site-inventory-location", "", "Explicit inventory location ID (empty to clear)")
	cmd.Flags().String("quantity", "", "Planned quantity")
	cmd.Flags().Bool("is-quantity-unknown", false, "Mark quantity as unknown")
	cmd.Flags().String("explicit-display-name", "", "Display name override")
	cmd.Flags().Bool("plan-requires-site-specific-material-types", false, "Override plan site-specific requirement")
	cmd.Flags().Bool("plan-requires-supplier-specific-material-types", false, "Override plan supplier-specific requirement")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanMaterialTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanMaterialTypesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("is-quantity-unknown") {
		attributes["is-quantity-unknown"] = opts.IsQuantityUnknown
	}
	if cmd.Flags().Changed("explicit-display-name") {
		attributes["explicit-display-name"] = opts.ExplicitDisplayName
	}
	if cmd.Flags().Changed("plan-requires-site-specific-material-types") {
		attributes["plan-requires-site-specific-material-types"] = opts.PlanRequiresSiteSpecificMaterialTypes
	}
	if cmd.Flags().Changed("plan-requires-supplier-specific-material-types") {
		attributes["plan-requires-supplier-specific-material-types"] = opts.PlanRequiresSupplierSpecificMaterialTypes
	}

	if cmd.Flags().Changed("job-production-plan") {
		if opts.JobProductionPlan == "" {
			relationships["job-production-plan"] = map[string]any{"data": nil}
		} else {
			relationships["job-production-plan"] = map[string]any{
				"data": map[string]any{
					"type": "job-production-plans",
					"id":   opts.JobProductionPlan,
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
	if cmd.Flags().Changed("unit-of-measure") {
		if opts.UnitOfMeasure == "" {
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
	if cmd.Flags().Changed("default-cost-code") {
		if opts.DefaultCostCode == "" {
			relationships["default-cost-code"] = map[string]any{"data": nil}
		} else {
			relationships["default-cost-code"] = map[string]any{
				"data": map[string]any{
					"type": "cost-codes",
					"id":   opts.DefaultCostCode,
				},
			}
		}
	}
	if cmd.Flags().Changed("explicit-material-mix-design") {
		if opts.ExplicitMaterialMixDesign == "" {
			relationships["explicit-material-mix-design"] = map[string]any{"data": nil}
		} else {
			relationships["explicit-material-mix-design"] = map[string]any{
				"data": map[string]any{
					"type": "material-mix-designs",
					"id":   opts.ExplicitMaterialMixDesign,
				},
			}
		}
	}
	if cmd.Flags().Changed("explicit-material-type-material-site-inventory-location") {
		if opts.ExplicitMaterialTypeMaterialSiteInventoryLocation == "" {
			relationships["explicit-material-type-material-site-inventory-location"] = map[string]any{"data": nil}
		} else {
			relationships["explicit-material-type-material-site-inventory-location"] = map[string]any{
				"data": map[string]any{
					"type": "material-type-material-site-inventory-locations",
					"id":   opts.ExplicitMaterialTypeMaterialSiteInventoryLocation,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "job-production-plan-material-types",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-material-types/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan material type %s\n", result.ID)
	return nil
}

func parseDoJobProductionPlanMaterialTypesUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanMaterialTypesUpdateOptions, error) {
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

	return doJobProductionPlanMaterialTypesUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
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
