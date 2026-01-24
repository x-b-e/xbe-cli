package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectRevenueItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectRevenueItemDetails struct {
	ID                             string   `json:"id"`
	Description                    string   `json:"description,omitempty"`
	ExternalDeveloperRevenueItemID string   `json:"external_developer_revenue_item_id,omitempty"`
	DeveloperQuantityEstimate      string   `json:"developer_quantity_estimate,omitempty"`
	ActualRevenueAmount            string   `json:"actual_revenue_amount,omitempty"`
	ActualCostAmount               string   `json:"actual_cost_amount,omitempty"`
	ActualProfitAmount             string   `json:"actual_profit_amount,omitempty"`
	ProjectID                      string   `json:"project_id,omitempty"`
	ProjectName                    string   `json:"project_name,omitempty"`
	ProjectNumber                  string   `json:"project_number,omitempty"`
	RevenueClassificationID        string   `json:"revenue_classification_id,omitempty"`
	RevenueClassificationName      string   `json:"revenue_classification_name,omitempty"`
	UnitOfMeasureID                string   `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName              string   `json:"unit_of_measure_name,omitempty"`
	UnitOfMeasureAbbreviation      string   `json:"unit_of_measure_abbreviation,omitempty"`
	QuantityEstimateID             string   `json:"quantity_estimate_id,omitempty"`
	PriceEstimateID                string   `json:"price_estimate_id,omitempty"`
	QuantityEstimateIDs            []string `json:"quantity_estimate_ids,omitempty"`
	PriceEstimateIDs               []string `json:"price_estimate_ids,omitempty"`
	ProjectPhaseRevenueItemIDs     []string `json:"project_phase_revenue_item_ids,omitempty"`
}

func newProjectRevenueItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project revenue item details",
		Long: `Show the full details of a project revenue item.

Project revenue items define billable line items for a project and tie to
revenue classifications, units of measure, and revenue estimates.

Output Fields:
  ID                           Project revenue item identifier
  Description                  Revenue item description
  External Developer Item ID   External developer revenue item ID
  Developer Quantity Estimate  Developer quantity estimate
  Actual Revenue Amount        Actual revenue amount (if available)
  Actual Cost Amount           Actual cost amount (if available)
  Actual Profit Amount         Actual profit amount (if available)
  Project                      Project name/number (or ID)
  Revenue Classification       Revenue classification name (or ID)
  Unit of Measure              Unit of measure
  Quantity Estimate ID         Current quantity estimate ID
  Price Estimate ID            Current price estimate ID
  Quantity Estimate IDs        All quantity estimate IDs
  Price Estimate IDs           All price estimate IDs
  Project Phase Revenue Item IDs Project phase revenue item IDs

Arguments:
  <id>                         The project revenue item ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project revenue item
  xbe view project-revenue-items show 123

  # Show as JSON
  xbe view project-revenue-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectRevenueItemsShow,
	}
	initProjectRevenueItemsShowFlags(cmd)
	return cmd
}

func init() {
	projectRevenueItemsCmd.AddCommand(newProjectRevenueItemsShowCmd())
}

func initProjectRevenueItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectRevenueItemsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectRevenueItemsShowOptions(cmd)
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
		return fmt.Errorf("project revenue item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-revenue-items]", "description,external-developer-revenue-item-id,developer-quantity-estimate,actual-revenue-amount,actual-cost-amount,actual-profit-amount,project,revenue-classification,unit-of-measure,quantity-estimate,price-estimate,quantity-estimates,price-estimates,project-phase-revenue-items")
	query.Set("include", "project,revenue-classification,unit-of-measure")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[project-revenue-classifications]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")

	body, _, err := client.Get(cmd.Context(), "/v1/project-revenue-items/"+id, query)
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

	details := buildProjectRevenueItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectRevenueItemDetails(cmd, details)
}

func parseProjectRevenueItemsShowOptions(cmd *cobra.Command) (projectRevenueItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectRevenueItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectRevenueItemDetails(resp jsonAPISingleResponse) projectRevenueItemDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectRevenueItemDetails{
		ID:                             resource.ID,
		Description:                    stringAttr(attrs, "description"),
		ExternalDeveloperRevenueItemID: stringAttr(attrs, "external-developer-revenue-item-id"),
		DeveloperQuantityEstimate:      stringAttr(attrs, "developer-quantity-estimate"),
		ActualRevenueAmount:            stringAttr(attrs, "actual-revenue-amount"),
		ActualCostAmount:               stringAttr(attrs, "actual-cost-amount"),
		ActualProfitAmount:             stringAttr(attrs, "actual-profit-amount"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(inc.Attributes, "name")
			details.ProjectNumber = stringAttr(inc.Attributes, "number")
		}
	}

	if rel, ok := resource.Relationships["revenue-classification"]; ok && rel.Data != nil {
		details.RevenueClassificationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.RevenueClassificationName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasureName = stringAttr(inc.Attributes, "name")
			details.UnitOfMeasureAbbreviation = stringAttr(inc.Attributes, "abbreviation")
		}
	}

	if rel, ok := resource.Relationships["quantity-estimate"]; ok && rel.Data != nil {
		details.QuantityEstimateID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["price-estimate"]; ok && rel.Data != nil {
		details.PriceEstimateID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["quantity-estimates"]; ok && rel.raw != nil {
		details.QuantityEstimateIDs = relationshipIDStrings(rel)
	}

	if rel, ok := resource.Relationships["price-estimates"]; ok && rel.raw != nil {
		details.PriceEstimateIDs = relationshipIDStrings(rel)
	}

	if rel, ok := resource.Relationships["project-phase-revenue-items"]; ok && rel.raw != nil {
		details.ProjectPhaseRevenueItemIDs = relationshipIDStrings(rel)
	}

	return details
}

func relationshipIDStrings(rel jsonAPIRelationship) []string {
	var refs []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &refs); err != nil {
		return nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}
	return ids
}

func renderProjectRevenueItemDetails(cmd *cobra.Command, details projectRevenueItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.ExternalDeveloperRevenueItemID != "" {
		fmt.Fprintf(out, "External Developer Item ID: %s\n", details.ExternalDeveloperRevenueItemID)
	}
	if details.DeveloperQuantityEstimate != "" {
		fmt.Fprintf(out, "Developer Quantity Estimate: %s\n", details.DeveloperQuantityEstimate)
	}
	if details.ActualRevenueAmount != "" {
		fmt.Fprintf(out, "Actual Revenue Amount: %s\n", details.ActualRevenueAmount)
	}
	if details.ActualCostAmount != "" {
		fmt.Fprintf(out, "Actual Cost Amount: %s\n", details.ActualCostAmount)
	}
	if details.ActualProfitAmount != "" {
		fmt.Fprintf(out, "Actual Profit Amount: %s\n", details.ActualProfitAmount)
	}

	projectLabel := firstNonEmpty(details.ProjectName, details.ProjectNumber, details.ProjectID)
	if projectLabel != "" {
		fmt.Fprintf(out, "Project: %s\n", projectLabel)
	}
	classificationLabel := firstNonEmpty(details.RevenueClassificationName, details.RevenueClassificationID)
	if classificationLabel != "" {
		fmt.Fprintf(out, "Revenue Classification: %s\n", classificationLabel)
	}
	unitLabel := firstNonEmpty(details.UnitOfMeasureAbbreviation, details.UnitOfMeasureName, details.UnitOfMeasureID)
	if unitLabel != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", unitLabel)
	}

	if details.QuantityEstimateID != "" {
		fmt.Fprintf(out, "Quantity Estimate ID: %s\n", details.QuantityEstimateID)
	}
	if details.PriceEstimateID != "" {
		fmt.Fprintf(out, "Price Estimate ID: %s\n", details.PriceEstimateID)
	}
	if len(details.QuantityEstimateIDs) > 0 {
		fmt.Fprintf(out, "Quantity Estimate IDs: %s\n", strings.Join(details.QuantityEstimateIDs, ", "))
	}
	if len(details.PriceEstimateIDs) > 0 {
		fmt.Fprintf(out, "Price Estimate IDs: %s\n", strings.Join(details.PriceEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseRevenueItemIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Revenue Item IDs: %s\n", strings.Join(details.ProjectPhaseRevenueItemIDs, ", "))
	}

	return nil
}
