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

type doProjectDuplicationsCreateOptions struct {
	BaseURL                                       string
	Token                                         string
	JSON                                          bool
	ProjectTemplate                               string
	NewDeveloper                                  string
	DerivedProjectTemplateName                    string
	DerivedDueOn                                  string
	DerivedProjectNumber                          string
	DerivedIsPrevailingWageApplicable             bool
	DerivedIsTimeCardPayrollCertificationRequired bool
	SkipProjectMaterialTypes                      bool
	SkipProjectCustomers                          bool
	SkipProjectTruckers                           bool
	SkipProjectTrailerClassifications             bool
	SkipProjectLaborClassifications               bool
	SkipCertificationRequirements                 bool
	SkipProjectCostCodes                          bool
	SkipProjectRevenueItems                       bool
	SkipProjectPhaseRevenueItems                  bool
}

type projectDuplicationCreateResult struct {
	ID                                            string `json:"id"`
	ProjectTemplateID                             string `json:"project_template_id,omitempty"`
	DerivedProjectID                              string `json:"derived_project_id,omitempty"`
	NewDeveloperID                                string `json:"new_developer_id,omitempty"`
	DerivedProjectTemplateName                    string `json:"derived_project_template_name,omitempty"`
	DerivedDueOn                                  string `json:"derived_due_on,omitempty"`
	DerivedProjectNumber                          string `json:"derived_project_number,omitempty"`
	DerivedIsPrevailingWageApplicable             bool   `json:"derived_is_prevailing_wage_applicable,omitempty"`
	DerivedIsTimeCardPayrollCertificationRequired bool   `json:"derived_is_time_card_payroll_certification_required,omitempty"`
	SkipProjectMaterialTypes                      bool   `json:"skip_project_material_types,omitempty"`
	SkipProjectCustomers                          bool   `json:"skip_project_customers,omitempty"`
	SkipProjectTruckers                           bool   `json:"skip_project_truckers,omitempty"`
	SkipProjectTrailerClassifications             bool   `json:"skip_project_trailer_classifications,omitempty"`
	SkipProjectLaborClassifications               bool   `json:"skip_project_labor_classifications,omitempty"`
	SkipCertificationRequirements                 bool   `json:"skip_certification_requirements,omitempty"`
	SkipProjectCostCodes                          bool   `json:"skip_project_cost_codes,omitempty"`
	SkipProjectRevenueItems                       bool   `json:"skip_project_revenue_items,omitempty"`
	SkipProjectPhaseRevenueItems                  bool   `json:"skip_project_phase_revenue_items,omitempty"`
}

func newDoProjectDuplicationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Duplicate a project",
		Long: `Duplicate a project from a template.

Required flags:
  --project-template   Template project ID

Optional flags:
  --derived-project-template-name                       Derived project template name
  --derived-project-number                              Derived project number
  --derived-due-on                                      Derived project due date (ISO 8601)
  --derived-is-prevailing-wage-applicable               Override prevailing wage setting
  --derived-is-time-card-payroll-certification-required Override time card payroll certification setting
  --new-developer                                       Developer ID for the derived project
  --skip-project-material-types                         Skip copying project material types
  --skip-project-customers                              Skip copying project customers
  --skip-project-truckers                               Skip copying project truckers
  --skip-project-trailer-classifications                Skip copying project trailer classifications
  --skip-project-labor-classifications                  Skip copying project labor classifications
  --skip-certification-requirements                     Skip copying certification requirements
  --skip-project-cost-codes                             Skip copying project cost codes
  --skip-project-revenue-items                          Skip copying project revenue items
  --skip-project-phase-revenue-items                    Skip copying project phase revenue items`,
		Example: `  # Duplicate a project
  xbe do project-duplications create --project-template 123

  # Duplicate with overrides and skip relations
  xbe do project-duplications create \
    --project-template 123 \
    --derived-project-template-name "Template Copy" \
    --derived-project-number "TMP-001" \
    --derived-due-on 2026-02-01 \
    --derived-is-prevailing-wage-applicable \
    --derived-is-time-card-payroll-certification-required \
    --skip-project-material-types \
    --skip-project-customers \
    --skip-project-truckers \
    --skip-project-trailer-classifications \
    --skip-project-labor-classifications \
    --skip-certification-requirements \
    --skip-project-cost-codes \
    --skip-project-revenue-items \
    --skip-project-phase-revenue-items`,
		Args: cobra.NoArgs,
		RunE: runDoProjectDuplicationsCreate,
	}
	initDoProjectDuplicationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectDuplicationsCmd.AddCommand(newDoProjectDuplicationsCreateCmd())
}

func initDoProjectDuplicationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-template", "", "Template project ID (required)")
	cmd.Flags().String("new-developer", "", "Developer ID for the derived project")
	cmd.Flags().String("derived-project-template-name", "", "Derived project template name")
	cmd.Flags().String("derived-project-number", "", "Derived project number")
	cmd.Flags().String("derived-due-on", "", "Derived project due date (ISO 8601)")
	cmd.Flags().Bool("derived-is-prevailing-wage-applicable", false, "Override prevailing wage setting")
	cmd.Flags().Bool("derived-is-time-card-payroll-certification-required", false, "Override time card payroll certification setting")
	cmd.Flags().Bool("skip-project-material-types", false, "Skip copying project material types")
	cmd.Flags().Bool("skip-project-customers", false, "Skip copying project customers")
	cmd.Flags().Bool("skip-project-truckers", false, "Skip copying project truckers")
	cmd.Flags().Bool("skip-project-trailer-classifications", false, "Skip copying project trailer classifications")
	cmd.Flags().Bool("skip-project-labor-classifications", false, "Skip copying project labor classifications")
	cmd.Flags().Bool("skip-certification-requirements", false, "Skip copying certification requirements")
	cmd.Flags().Bool("skip-project-cost-codes", false, "Skip copying project cost codes")
	cmd.Flags().Bool("skip-project-revenue-items", false, "Skip copying project revenue items")
	cmd.Flags().Bool("skip-project-phase-revenue-items", false, "Skip copying project phase revenue items")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("project-template")
}

func runDoProjectDuplicationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectDuplicationsCreateOptions(cmd)
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

	projectTemplateID := strings.TrimSpace(opts.ProjectTemplate)
	if projectTemplateID == "" {
		err := fmt.Errorf("--project-template is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.DerivedProjectTemplateName != "" {
		attributes["derived-project-template-name"] = opts.DerivedProjectTemplateName
	}
	if opts.DerivedProjectNumber != "" {
		attributes["derived-project-number"] = opts.DerivedProjectNumber
	}
	if opts.DerivedDueOn != "" {
		attributes["derived-due-on"] = opts.DerivedDueOn
	}
	if cmd.Flags().Changed("derived-is-prevailing-wage-applicable") {
		attributes["derived-is-prevailing-wage-applicable"] = opts.DerivedIsPrevailingWageApplicable
	}
	if cmd.Flags().Changed("derived-is-time-card-payroll-certification-required") {
		attributes["derived-is-time-card-payroll-certification-required"] = opts.DerivedIsTimeCardPayrollCertificationRequired
	}
	if cmd.Flags().Changed("skip-project-material-types") {
		attributes["skip-project-material-types"] = opts.SkipProjectMaterialTypes
	}
	if cmd.Flags().Changed("skip-project-customers") {
		attributes["skip-project-customers"] = opts.SkipProjectCustomers
	}
	if cmd.Flags().Changed("skip-project-truckers") {
		attributes["skip-project-truckers"] = opts.SkipProjectTruckers
	}
	if cmd.Flags().Changed("skip-project-trailer-classifications") {
		attributes["skip-project-trailer-classifications"] = opts.SkipProjectTrailerClassifications
	}
	if cmd.Flags().Changed("skip-project-labor-classifications") {
		attributes["skip-project-labor-classifications"] = opts.SkipProjectLaborClassifications
	}
	if cmd.Flags().Changed("skip-certification-requirements") {
		attributes["skip-certification-requirements"] = opts.SkipCertificationRequirements
	}
	if cmd.Flags().Changed("skip-project-cost-codes") {
		attributes["skip-project-cost-codes"] = opts.SkipProjectCostCodes
	}
	if cmd.Flags().Changed("skip-project-revenue-items") {
		attributes["skip-project-revenue-items"] = opts.SkipProjectRevenueItems
	}
	if cmd.Flags().Changed("skip-project-phase-revenue-items") {
		attributes["skip-project-phase-revenue-items"] = opts.SkipProjectPhaseRevenueItems
	}

	relationships := map[string]any{
		"project-template": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   projectTemplateID,
			},
		},
	}

	if opts.NewDeveloper != "" {
		relationships["new-developer"] = map[string]any{
			"data": map[string]any{
				"type": "developers",
				"id":   opts.NewDeveloper,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-duplications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-duplications", jsonBody)
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

	result := buildProjectDuplicationCreateResult(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	if result.DerivedProjectID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created project duplication %s (derived project %s)\n", result.ID, result.DerivedProjectID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project duplication %s\n", result.ID)
	return nil
}

func parseDoProjectDuplicationsCreateOptions(cmd *cobra.Command) (doProjectDuplicationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTemplate, _ := cmd.Flags().GetString("project-template")
	newDeveloper, _ := cmd.Flags().GetString("new-developer")
	derivedProjectTemplateName, _ := cmd.Flags().GetString("derived-project-template-name")
	derivedProjectNumber, _ := cmd.Flags().GetString("derived-project-number")
	derivedDueOn, _ := cmd.Flags().GetString("derived-due-on")
	derivedIsPrevailingWageApplicable, _ := cmd.Flags().GetBool("derived-is-prevailing-wage-applicable")
	derivedIsTimeCardPayrollCertificationRequired, _ := cmd.Flags().GetBool("derived-is-time-card-payroll-certification-required")
	skipProjectMaterialTypes, _ := cmd.Flags().GetBool("skip-project-material-types")
	skipProjectCustomers, _ := cmd.Flags().GetBool("skip-project-customers")
	skipProjectTruckers, _ := cmd.Flags().GetBool("skip-project-truckers")
	skipProjectTrailerClassifications, _ := cmd.Flags().GetBool("skip-project-trailer-classifications")
	skipProjectLaborClassifications, _ := cmd.Flags().GetBool("skip-project-labor-classifications")
	skipCertificationRequirements, _ := cmd.Flags().GetBool("skip-certification-requirements")
	skipProjectCostCodes, _ := cmd.Flags().GetBool("skip-project-cost-codes")
	skipProjectRevenueItems, _ := cmd.Flags().GetBool("skip-project-revenue-items")
	skipProjectPhaseRevenueItems, _ := cmd.Flags().GetBool("skip-project-phase-revenue-items")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectDuplicationsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		ProjectTemplate:                   projectTemplate,
		NewDeveloper:                      newDeveloper,
		DerivedProjectTemplateName:        derivedProjectTemplateName,
		DerivedProjectNumber:              derivedProjectNumber,
		DerivedDueOn:                      derivedDueOn,
		DerivedIsPrevailingWageApplicable: derivedIsPrevailingWageApplicable,
		DerivedIsTimeCardPayrollCertificationRequired: derivedIsTimeCardPayrollCertificationRequired,
		SkipProjectMaterialTypes:                      skipProjectMaterialTypes,
		SkipProjectCustomers:                          skipProjectCustomers,
		SkipProjectTruckers:                           skipProjectTruckers,
		SkipProjectTrailerClassifications:             skipProjectTrailerClassifications,
		SkipProjectLaborClassifications:               skipProjectLaborClassifications,
		SkipCertificationRequirements:                 skipCertificationRequirements,
		SkipProjectCostCodes:                          skipProjectCostCodes,
		SkipProjectRevenueItems:                       skipProjectRevenueItems,
		SkipProjectPhaseRevenueItems:                  skipProjectPhaseRevenueItems,
	}, nil
}

func buildProjectDuplicationCreateResult(resp jsonAPISingleResponse) projectDuplicationCreateResult {
	resource := resp.Data
	attrs := resource.Attributes

	result := projectDuplicationCreateResult{
		ID:                                resource.ID,
		DerivedProjectTemplateName:        stringAttr(attrs, "derived-project-template-name"),
		DerivedDueOn:                      stringAttr(attrs, "derived-due-on"),
		DerivedProjectNumber:              stringAttr(attrs, "derived-project-number"),
		DerivedIsPrevailingWageApplicable: boolAttr(attrs, "derived-is-prevailing-wage-applicable"),
		DerivedIsTimeCardPayrollCertificationRequired: boolAttr(attrs, "derived-is-time-card-payroll-certification-required"),
		SkipProjectMaterialTypes:                      boolAttr(attrs, "skip-project-material-types"),
		SkipProjectCustomers:                          boolAttr(attrs, "skip-project-customers"),
		SkipProjectTruckers:                           boolAttr(attrs, "skip-project-truckers"),
		SkipProjectTrailerClassifications:             boolAttr(attrs, "skip-project-trailer-classifications"),
		SkipProjectLaborClassifications:               boolAttr(attrs, "skip-project-labor-classifications"),
		SkipCertificationRequirements:                 boolAttr(attrs, "skip-certification-requirements"),
		SkipProjectCostCodes:                          boolAttr(attrs, "skip-project-cost-codes"),
		SkipProjectRevenueItems:                       boolAttr(attrs, "skip-project-revenue-items"),
		SkipProjectPhaseRevenueItems:                  boolAttr(attrs, "skip-project-phase-revenue-items"),
	}

	if rel, ok := resource.Relationships["project-template"]; ok && rel.Data != nil {
		result.ProjectTemplateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["derived-project"]; ok && rel.Data != nil {
		result.DerivedProjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["new-developer"]; ok && rel.Data != nil {
		result.NewDeveloperID = rel.Data.ID
	}

	return result
}
