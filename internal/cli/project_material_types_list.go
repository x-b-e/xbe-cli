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

type projectMaterialTypesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Project          string
	MaterialType     string
	PickupLocation   string
	DeliveryLocation string
	PickupAtMinMin   string
	PickupAtMinMax   string
	PickupAtMaxMin   string
	PickupAtMaxMax   string
	DeliverAtMinMin  string
	DeliverAtMinMax  string
	DeliverAtMaxMin  string
	DeliverAtMaxMax  string
}

type projectMaterialTypeRow struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"display_name,omitempty"`
	Quantity           string `json:"quantity,omitempty"`
	ProjectID          string `json:"project_id,omitempty"`
	ProjectName        string `json:"project_name,omitempty"`
	MaterialTypeID     string `json:"material_type_id,omitempty"`
	MaterialTypeName   string `json:"material_type_name,omitempty"`
	UnitOfMeasureID    string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure      string `json:"unit_of_measure,omitempty"`
	MaterialSiteID     string `json:"material_site_id,omitempty"`
	MaterialSiteName   string `json:"material_site_name,omitempty"`
	JobSiteID          string `json:"job_site_id,omitempty"`
	JobSiteName        string `json:"job_site_name,omitempty"`
	PickupLocationID   string `json:"pickup_location_id,omitempty"`
	DeliveryLocationID string `json:"delivery_location_id,omitempty"`
}

func newProjectMaterialTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project material types",
		Long: `List project material types with filtering and pagination.

Project material types associate material types with projects and capture
optional quantities, units, and pickup/delivery windows.

Output Columns:
  ID             Project material type identifier
  DISPLAY        Display name
  PROJECT        Project name
  MATERIAL       Material type name
  QTY            Quantity
  UOM            Unit of measure
  MATERIAL SITE  Material site name
  JOB SITE       Job site name

Filters:
  --project               Filter by project ID
  --material-type         Filter by material type ID
  --pickup-location       Filter by pickup location ID
  --delivery-location     Filter by delivery location ID
  --pickup-at-min-min     Filter by pickup_at_min on/after timestamp (ISO 8601)
  --pickup-at-min-max     Filter by pickup_at_min on/before timestamp (ISO 8601)
  --pickup-at-max-min     Filter by pickup_at_max on/after timestamp (ISO 8601)
  --pickup-at-max-max     Filter by pickup_at_max on/before timestamp (ISO 8601)
  --deliver-at-min-min    Filter by deliver_at_min on/after timestamp (ISO 8601)
  --deliver-at-min-max    Filter by deliver_at_min on/before timestamp (ISO 8601)
  --deliver-at-max-min    Filter by deliver_at_max on/after timestamp (ISO 8601)
  --deliver-at-max-max    Filter by deliver_at_max on/before timestamp (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List project material types
  xbe view project-material-types list

  # Filter by project
  xbe view project-material-types list --project 123

  # Filter by material type
  xbe view project-material-types list --material-type 456

  # Filter by pickup window
  xbe view project-material-types list --pickup-at-min-min 2026-01-01T08:00:00Z --pickup-at-min-max 2026-01-01T10:00:00Z

  # Output as JSON
  xbe view project-material-types list --json`,
		RunE: runProjectMaterialTypesList,
	}
	initProjectMaterialTypesListFlags(cmd)
	return cmd
}

func init() {
	projectMaterialTypesCmd.AddCommand(newProjectMaterialTypesListCmd())
}

func initProjectMaterialTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("pickup-location", "", "Filter by pickup location ID")
	cmd.Flags().String("delivery-location", "", "Filter by delivery location ID")
	cmd.Flags().String("pickup-at-min-min", "", "Filter by pickup_at_min on/after timestamp (ISO 8601)")
	cmd.Flags().String("pickup-at-min-max", "", "Filter by pickup_at_min on/before timestamp (ISO 8601)")
	cmd.Flags().String("pickup-at-max-min", "", "Filter by pickup_at_max on/after timestamp (ISO 8601)")
	cmd.Flags().String("pickup-at-max-max", "", "Filter by pickup_at_max on/before timestamp (ISO 8601)")
	cmd.Flags().String("deliver-at-min-min", "", "Filter by deliver_at_min on/after timestamp (ISO 8601)")
	cmd.Flags().String("deliver-at-min-max", "", "Filter by deliver_at_min on/before timestamp (ISO 8601)")
	cmd.Flags().String("deliver-at-max-min", "", "Filter by deliver_at_max on/after timestamp (ISO 8601)")
	cmd.Flags().String("deliver-at-max-max", "", "Filter by deliver_at_max on/before timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectMaterialTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectMaterialTypesListOptions(cmd)
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
	query.Set("fields[project-material-types]", "display-name,quantity,pickup-at-min,pickup-at-max,deliver-at-min,deliver-at-max")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[material-types]", "display-name,name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[job-sites]", "name")
	query.Set("include", "project,material-type,unit-of-measure,material-site,job-site")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[pickup-location]", opts.PickupLocation)
	setFilterIfPresent(query, "filter[delivery-location]", opts.DeliveryLocation)
	setFilterIfPresent(query, "filter[pickup-at-min-min]", opts.PickupAtMinMin)
	setFilterIfPresent(query, "filter[pickup-at-min-max]", opts.PickupAtMinMax)
	setFilterIfPresent(query, "filter[pickup-at-max-min]", opts.PickupAtMaxMin)
	setFilterIfPresent(query, "filter[pickup-at-max-max]", opts.PickupAtMaxMax)
	setFilterIfPresent(query, "filter[deliver-at-min-min]", opts.DeliverAtMinMin)
	setFilterIfPresent(query, "filter[deliver-at-min-max]", opts.DeliverAtMinMax)
	setFilterIfPresent(query, "filter[deliver-at-max-min]", opts.DeliverAtMaxMin)
	setFilterIfPresent(query, "filter[deliver-at-max-max]", opts.DeliverAtMaxMax)

	body, _, err := client.Get(cmd.Context(), "/v1/project-material-types", query)
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

	rows := buildProjectMaterialTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectMaterialTypesTable(cmd, rows)
}

func parseProjectMaterialTypesListOptions(cmd *cobra.Command) (projectMaterialTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	project, _ := cmd.Flags().GetString("project")
	materialType, _ := cmd.Flags().GetString("material-type")
	pickupLocation, _ := cmd.Flags().GetString("pickup-location")
	deliveryLocation, _ := cmd.Flags().GetString("delivery-location")
	pickupAtMinMin, _ := cmd.Flags().GetString("pickup-at-min-min")
	pickupAtMinMax, _ := cmd.Flags().GetString("pickup-at-min-max")
	pickupAtMaxMin, _ := cmd.Flags().GetString("pickup-at-max-min")
	pickupAtMaxMax, _ := cmd.Flags().GetString("pickup-at-max-max")
	deliverAtMinMin, _ := cmd.Flags().GetString("deliver-at-min-min")
	deliverAtMinMax, _ := cmd.Flags().GetString("deliver-at-min-max")
	deliverAtMaxMin, _ := cmd.Flags().GetString("deliver-at-max-min")
	deliverAtMaxMax, _ := cmd.Flags().GetString("deliver-at-max-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectMaterialTypesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Project:          project,
		MaterialType:     materialType,
		PickupLocation:   pickupLocation,
		DeliveryLocation: deliveryLocation,
		PickupAtMinMin:   pickupAtMinMin,
		PickupAtMinMax:   pickupAtMinMax,
		PickupAtMaxMin:   pickupAtMaxMin,
		PickupAtMaxMax:   pickupAtMaxMax,
		DeliverAtMinMin:  deliverAtMinMin,
		DeliverAtMinMax:  deliverAtMinMax,
		DeliverAtMaxMin:  deliverAtMaxMin,
		DeliverAtMaxMax:  deliverAtMaxMax,
	}, nil
}

func buildProjectMaterialTypeRows(resp jsonAPIResponse) []projectMaterialTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectMaterialTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectMaterialTypeRow{
			ID:          resource.ID,
			DisplayName: stringAttr(resource.Attributes, "display-name"),
			Quantity:    stringAttr(resource.Attributes, "quantity"),
		}

		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
			if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectName = firstNonEmpty(
					stringAttr(project.Attributes, "name"),
					stringAttr(project.Attributes, "number"),
				)
			}
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialTypeName = firstNonEmpty(
					stringAttr(materialType.Attributes, "display-name"),
					stringAttr(materialType.Attributes, "name"),
				)
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

		if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
			row.MaterialSiteID = rel.Data.ID
			if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialSiteName = stringAttr(site.Attributes, "name")
			}
		}

		if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
			row.JobSiteID = rel.Data.ID
			if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.JobSiteName = stringAttr(site.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderProjectMaterialTypesTable(cmd *cobra.Command, rows []projectMaterialTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project material types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDISPLAY\tPROJECT\tMATERIAL\tQTY\tUOM\tMATERIAL SITE\tJOB SITE")
	for _, row := range rows {
		project := row.ProjectName
		if project == "" {
			project = row.ProjectID
		}
		material := row.MaterialTypeName
		if material == "" {
			material = row.MaterialTypeID
		}
		uom := row.UnitOfMeasure
		if uom == "" {
			uom = row.UnitOfMeasureID
		}
		materialSite := row.MaterialSiteName
		if materialSite == "" {
			materialSite = row.MaterialSiteID
		}
		jobSite := row.JobSiteName
		if jobSite == "" {
			jobSite = row.JobSiteID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.DisplayName,
			project,
			material,
			row.Quantity,
			uom,
			materialSite,
			jobSite,
		)
	}
	return writer.Flush()
}
