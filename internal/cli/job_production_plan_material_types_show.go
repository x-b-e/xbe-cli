package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanMaterialTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanMaterialTypeDetails struct {
	ID                                                  string   `json:"id"`
	Quantity                                            float64  `json:"quantity"`
	IsQuantityUnknown                                   bool     `json:"is_quantity_unknown"`
	DisplayName                                         string   `json:"display_name,omitempty"`
	ExplicitDisplayName                                 string   `json:"explicit_display_name,omitempty"`
	PlanRequiresSiteSpecificMaterialTypes               bool     `json:"plan_requires_site_specific_material_types,omitempty"`
	PlanRequiresSupplierSpecificMaterialTypes           bool     `json:"plan_requires_supplier_specific_material_types,omitempty"`
	JobProductionPlanID                                 string   `json:"job_production_plan_id,omitempty"`
	JobProductionPlan                                   string   `json:"job_production_plan,omitempty"`
	MaterialTypeID                                      string   `json:"material_type_id,omitempty"`
	MaterialType                                        string   `json:"material_type,omitempty"`
	MaterialSiteID                                      string   `json:"material_site_id,omitempty"`
	MaterialSite                                        string   `json:"material_site,omitempty"`
	UnitOfMeasureID                                     string   `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure                                       string   `json:"unit_of_measure,omitempty"`
	DefaultCostCodeID                                   string   `json:"default_cost_code_id,omitempty"`
	DefaultCostCode                                     string   `json:"default_cost_code,omitempty"`
	ExplicitMaterialMixDesignID                         string   `json:"explicit_material_mix_design_id,omitempty"`
	ExplicitMaterialMixDesign                           string   `json:"explicit_material_mix_design,omitempty"`
	ExplicitMaterialTypeMaterialSiteInventoryLocationID string   `json:"explicit_material_type_material_site_inventory_location_id,omitempty"`
	MaterialTypeMaterialSiteInventoryLocationID         string   `json:"material_type_material_site_inventory_location_id,omitempty"`
	QualityControlRequirementIDs                        []string `json:"quality_control_requirement_ids,omitempty"`
	ExternalIdentificationIDs                           []string `json:"external_identification_ids,omitempty"`
}

func newJobProductionPlanMaterialTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan material type details",
		Long: `Show the full details of a job production plan material type.

Output Fields:
  ID                         Job production plan material type identifier
  Quantity                   Planned quantity
  Quantity Unknown           Whether quantity is unknown
  Display Name               Display name
  Explicit Display Name      Explicit display name override
  Plan Requires Site Specific Material Types      Plan requires site-specific material types
  Plan Requires Supplier Specific Material Types  Plan requires supplier-specific material types
  Job Production Plan        Job production plan
  Material Type              Material type
  Material Site              Material site
  Unit Of Measure            Unit of measure
  Default Cost Code          Default cost code
  Explicit Material Mix Design                Explicit material mix design
  Explicit Material Type Material Site Inventory Location  Explicit inventory location
  Material Type Material Site Inventory Location           Derived inventory location
  Quality Control Requirements                QC requirement IDs
  External Identifications                     External identification IDs

Arguments:
  <id>    Job production plan material type ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job production plan material type
  xbe view job-production-plan-material-types show 123

  # JSON output
  xbe view job-production-plan-material-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanMaterialTypesShow,
	}
	initJobProductionPlanMaterialTypesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialTypesCmd.AddCommand(newJobProductionPlanMaterialTypesShowCmd())
}

func initJobProductionPlanMaterialTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialTypesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanMaterialTypesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan material type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-material-types]", "quantity,is-quantity-unknown,explicit-display-name,display-name,plan-requires-site-specific-material-types,plan-requires-supplier-specific-material-types")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[cost-codes]", "code,description")
	query.Set("fields[material-mix-designs]", "description,mix")
	query.Set("include", "job-production-plan,material-type,material-site,unit-of-measure,default-cost-code,explicit-material-mix-design,explicit-material-type-material-site-inventory-location,material-type-material-site-inventory-location,job-production-plan-material-type-quality-control-requirements,external-identifications")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-types/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobProductionPlanMaterialTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanMaterialTypeDetails(cmd, details)
}

func parseJobProductionPlanMaterialTypesShowOptions(cmd *cobra.Command) (jobProductionPlanMaterialTypesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanMaterialTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanMaterialTypeDetails(resp jsonAPISingleResponse) jobProductionPlanMaterialTypeDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := jobProductionPlanMaterialTypeDetails{
		ID:                                    resp.Data.ID,
		Quantity:                              floatAttr(attrs, "quantity"),
		IsQuantityUnknown:                     boolAttr(attrs, "is-quantity-unknown"),
		DisplayName:                           stringAttr(attrs, "display-name"),
		ExplicitDisplayName:                   stringAttr(attrs, "explicit-display-name"),
		PlanRequiresSiteSpecificMaterialTypes: boolAttr(attrs, "plan-requires-site-specific-material-types"),
		PlanRequiresSupplierSpecificMaterialTypes: boolAttr(attrs, "plan-requires-supplier-specific-material-types"),
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			jobNumber := stringAttr(plan.Attributes, "job-number")
			jobName := stringAttr(plan.Attributes, "job-name")
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialType = firstNonEmpty(
				stringAttr(materialType.Attributes, "display-name"),
				stringAttr(materialType.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if materialSite, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(materialSite.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = firstNonEmpty(
				stringAttr(uom.Attributes, "abbreviation"),
				stringAttr(uom.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["default-cost-code"]; ok && rel.Data != nil {
		details.DefaultCostCodeID = rel.Data.ID
		if costCode, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DefaultCostCode = stringAttr(costCode.Attributes, "code")
		}
	}

	if rel, ok := resp.Data.Relationships["explicit-material-mix-design"]; ok && rel.Data != nil {
		details.ExplicitMaterialMixDesignID = rel.Data.ID
		if mixDesign, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ExplicitMaterialMixDesign = firstNonEmpty(
				stringAttr(mixDesign.Attributes, "description"),
				stringAttr(mixDesign.Attributes, "mix"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["explicit-material-type-material-site-inventory-location"]; ok && rel.Data != nil {
		details.ExplicitMaterialTypeMaterialSiteInventoryLocationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["material-type-material-site-inventory-location"]; ok && rel.Data != nil {
		details.MaterialTypeMaterialSiteInventoryLocationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["job-production-plan-material-type-quality-control-requirements"]; ok && rel.raw != nil {
		var qcRefs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &qcRefs); err == nil {
			for _, ref := range qcRefs {
				details.QualityControlRequirementIDs = append(details.QualityControlRequirementIDs, ref.ID)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["external-identifications"]; ok && rel.raw != nil {
		var extRefs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &extRefs); err == nil {
			for _, ref := range extRefs {
				details.ExternalIdentificationIDs = append(details.ExternalIdentificationIDs, ref.ID)
			}
		}
	}

	return details
}

func renderJobProductionPlanMaterialTypeDetails(cmd *cobra.Command, details jobProductionPlanMaterialTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	quantity := fmt.Sprintf("%.2f", details.Quantity)
	if details.IsQuantityUnknown {
		quantity = "unknown"
	}
	fmt.Fprintf(out, "Quantity: %s\n", quantity)
	fmt.Fprintf(out, "Quantity Unknown: %t\n", details.IsQuantityUnknown)

	if details.DisplayName != "" {
		fmt.Fprintf(out, "Display Name: %s\n", details.DisplayName)
	}
	if details.ExplicitDisplayName != "" {
		fmt.Fprintf(out, "Explicit Display Name: %s\n", details.ExplicitDisplayName)
	}

	fmt.Fprintf(out, "Plan Requires Site Specific Material Types: %t\n", details.PlanRequiresSiteSpecificMaterialTypes)
	fmt.Fprintf(out, "Plan Requires Supplier Specific Material Types: %t\n", details.PlanRequiresSupplierSpecificMaterialTypes)

	writeLabelWithID(out, "Job Production Plan", details.JobProductionPlan, details.JobProductionPlanID)
	writeLabelWithID(out, "Material Type", details.MaterialType, details.MaterialTypeID)
	writeLabelWithID(out, "Material Site", details.MaterialSite, details.MaterialSiteID)
	writeLabelWithID(out, "Unit Of Measure", details.UnitOfMeasure, details.UnitOfMeasureID)
	writeLabelWithID(out, "Default Cost Code", details.DefaultCostCode, details.DefaultCostCodeID)
	writeLabelWithID(out, "Explicit Material Mix Design", details.ExplicitMaterialMixDesign, details.ExplicitMaterialMixDesignID)

	if details.ExplicitMaterialTypeMaterialSiteInventoryLocationID != "" {
		fmt.Fprintf(out, "Explicit Inventory Location: %s\n", details.ExplicitMaterialTypeMaterialSiteInventoryLocationID)
	}
	if details.MaterialTypeMaterialSiteInventoryLocationID != "" {
		fmt.Fprintf(out, "Derived Inventory Location: %s\n", details.MaterialTypeMaterialSiteInventoryLocationID)
	}

	if len(details.QualityControlRequirementIDs) > 0 {
		fmt.Fprintf(out, "Quality Control Requirements: %s\n", strings.Join(details.QualityControlRequirementIDs, ", "))
	}
	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintf(out, "External Identifications: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}

	return nil
}

func writeLabelWithID(out io.Writer, label, value, id string) {
	if value != "" && id != "" {
		fmt.Fprintf(out, "%s: %s (%s)\n", label, value, id)
		return
	}
	if value != "" {
		fmt.Fprintf(out, "%s: %s\n", label, value)
		return
	}
	if id != "" {
		fmt.Fprintf(out, "%s: %s\n", label, id)
	}
}
