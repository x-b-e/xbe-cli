package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanMaterialTypesListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	JobProductionPlan           string
	MaterialType                string
	UnitOfMeasure               string
	MaterialSite                string
	DefaultCostCode             string
	ExplicitMaterialMixDesign   string
	Customer                    string
	Broker                      string
	MaterialSupplier            string
	StartOnMin                  string
	StartOnMax                  string
	Status                      string
	ExternalIdentificationValue string
}

type jobProductionPlanMaterialTypeRow struct {
	ID                  string  `json:"id"`
	JobProductionPlan   string  `json:"job_production_plan,omitempty"`
	JobProductionPlanID string  `json:"job_production_plan_id,omitempty"`
	MaterialType        string  `json:"material_type,omitempty"`
	MaterialTypeID      string  `json:"material_type_id,omitempty"`
	MaterialSite        string  `json:"material_site,omitempty"`
	MaterialSiteID      string  `json:"material_site_id,omitempty"`
	Quantity            float64 `json:"quantity,omitempty"`
	IsQuantityUnknown   bool    `json:"is_quantity_unknown,omitempty"`
	UnitOfMeasure       string  `json:"unit_of_measure,omitempty"`
	UnitOfMeasureID     string  `json:"unit_of_measure_id,omitempty"`
	DisplayName         string  `json:"display_name,omitempty"`
}

func newJobProductionPlanMaterialTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan material types",
		Long: `List job production plan material types with filtering and pagination.

Output Columns:
  ID         Material type identifier on the plan
  PLAN       Job production plan (job number/name)
  MATERIAL   Material type
  SITE       Material site
  QTY        Planned quantity (or unknown)
  UOM        Unit of measure
  DISPLAY    Display name

Filters:
  --job-production-plan            Filter by job production plan ID
  --material-type                  Filter by material type ID
  --unit-of-measure                Filter by unit of measure ID
  --material-site                  Filter by material site ID
  --default-cost-code              Filter by default cost code ID
  --explicit-material-mix-design   Filter by explicit material mix design ID
  --customer                       Filter by customer ID
  --broker                         Filter by broker ID
  --material-supplier              Filter by material supplier ID
  --start-on-min                   Filter by plan start date (min, YYYY-MM-DD)
  --start-on-max                   Filter by plan start date (max, YYYY-MM-DD)
  --status                         Filter by plan status (editing/submitted/rejected/approved/cancelled/complete/abandoned/scrapped)
  --external-identification-value  Filter by external identification value

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan material types
  xbe view job-production-plan-material-types list

  # Filter by job production plan
  xbe view job-production-plan-material-types list --job-production-plan 123

  # Filter by material site
  xbe view job-production-plan-material-types list --material-site 456

  # Filter by plan status
  xbe view job-production-plan-material-types list --status approved

  # JSON output
  xbe view job-production-plan-material-types list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanMaterialTypesList,
	}
	initJobProductionPlanMaterialTypesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialTypesCmd.AddCommand(newJobProductionPlanMaterialTypesListCmd())
}

func initJobProductionPlanMaterialTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("default-cost-code", "", "Filter by default cost code ID")
	cmd.Flags().String("explicit-material-mix-design", "", "Filter by explicit material mix design ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("start-on-min", "", "Filter by plan start date (min, YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by plan start date (max, YYYY-MM-DD)")
	cmd.Flags().String("status", "", "Filter by plan status")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanMaterialTypesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-material-types]", "quantity,is-quantity-unknown,explicit-display-name,display-name")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("include", "job-production-plan,material-type,material-site,unit-of-measure")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[default-cost-code]", opts.DefaultCostCode)
	setFilterIfPresent(query, "filter[explicit-material-mix-design]", opts.ExplicitMaterialMixDesign)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[start-on-min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start-on-max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-types", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildJobProductionPlanMaterialTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanMaterialTypesTable(cmd, rows)
}

func parseJobProductionPlanMaterialTypesListOptions(cmd *cobra.Command) (jobProductionPlanMaterialTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	materialType, _ := cmd.Flags().GetString("material-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	materialSite, _ := cmd.Flags().GetString("material-site")
	defaultCostCode, _ := cmd.Flags().GetString("default-cost-code")
	explicitMaterialMixDesign, _ := cmd.Flags().GetString("explicit-material-mix-design")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	status, _ := cmd.Flags().GetString("status")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanMaterialTypesListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		JobProductionPlan:           jobProductionPlan,
		MaterialType:                materialType,
		UnitOfMeasure:               unitOfMeasure,
		MaterialSite:                materialSite,
		DefaultCostCode:             defaultCostCode,
		ExplicitMaterialMixDesign:   explicitMaterialMixDesign,
		Customer:                    customer,
		Broker:                      broker,
		MaterialSupplier:            materialSupplier,
		StartOnMin:                  startOnMin,
		StartOnMax:                  startOnMax,
		Status:                      status,
		ExternalIdentificationValue: externalIdentificationValue,
	}, nil
}

func buildJobProductionPlanMaterialTypeRows(resp jsonAPIResponse) []jobProductionPlanMaterialTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanMaterialTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanMaterialTypeRow{
			ID:                resource.ID,
			Quantity:          floatAttr(attrs, "quantity"),
			IsQuantityUnknown: boolAttr(attrs, "is-quantity-unknown"),
			DisplayName:       stringAttr(attrs, "display-name"),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
			if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				jobNumber := stringAttr(plan.Attributes, "job-number")
				jobName := stringAttr(plan.Attributes, "job-name")
				if jobNumber != "" && jobName != "" {
					row.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
				} else {
					row.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
				}
			}
		}
		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialType = firstNonEmpty(
					stringAttr(materialType.Attributes, "display-name"),
					stringAttr(materialType.Attributes, "name"),
				)
			}
		}
		if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
			row.MaterialSiteID = rel.Data.ID
			if materialSite, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialSite = stringAttr(materialSite.Attributes, "name")
			}
		}
		if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
			row.UnitOfMeasureID = rel.Data.ID
			if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.UnitOfMeasure = firstNonEmpty(
					stringAttr(uom.Attributes, "abbreviation"),
					stringAttr(uom.Attributes, "name"),
				)
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanMaterialTypesTable(cmd *cobra.Command, rows []jobProductionPlanMaterialTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan material types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tMATERIAL\tSITE\tQTY\tUOM\tDISPLAY")
	for _, row := range rows {
		plan := row.JobProductionPlan
		if plan == "" {
			plan = row.JobProductionPlanID
		}
		material := row.MaterialType
		if material == "" {
			material = row.MaterialTypeID
		}
		site := row.MaterialSite
		if site == "" {
			site = row.MaterialSiteID
		}
		uom := row.UnitOfMeasure
		if uom == "" {
			uom = row.UnitOfMeasureID
		}
		qty := fmt.Sprintf("%.2f", row.Quantity)
		if row.IsQuantityUnknown {
			qty = "unknown"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(plan, 24),
			truncateString(material, 20),
			truncateString(site, 18),
			qty,
			truncateString(uom, 6),
			truncateString(row.DisplayName, 20),
		)
	}
	return writer.Flush()
}
