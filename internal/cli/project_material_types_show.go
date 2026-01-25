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

type projectMaterialTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectMaterialTypeDetails struct {
	ID                   string `json:"id"`
	DisplayName          string `json:"display_name,omitempty"`
	ExplicitDisplayName  string `json:"explicit_display_name,omitempty"`
	Quantity             string `json:"quantity,omitempty"`
	PickupAtMin          string `json:"pickup_at_min,omitempty"`
	PickupAtMax          string `json:"pickup_at_max,omitempty"`
	DeliverAtMin         string `json:"deliver_at_min,omitempty"`
	DeliverAtMax         string `json:"deliver_at_max,omitempty"`
	ProjectID            string `json:"project_id,omitempty"`
	ProjectName          string `json:"project_name,omitempty"`
	MaterialTypeID       string `json:"material_type_id,omitempty"`
	MaterialTypeName     string `json:"material_type_name,omitempty"`
	UnitOfMeasureID      string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure        string `json:"unit_of_measure,omitempty"`
	MaterialSiteID       string `json:"material_site_id,omitempty"`
	MaterialSiteName     string `json:"material_site_name,omitempty"`
	JobSiteID            string `json:"job_site_id,omitempty"`
	JobSiteName          string `json:"job_site_name,omitempty"`
	PickupLocationID     string `json:"pickup_location_id,omitempty"`
	PickupLocationName   string `json:"pickup_location_name,omitempty"`
	DeliveryLocationID   string `json:"delivery_location_id,omitempty"`
	DeliveryLocationName string `json:"delivery_location_name,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}

func newProjectMaterialTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project material type details",
		Long: `Show the full details of a project material type.

Output Fields:
  ID
  Display Name
  Explicit Display Name
  Quantity
  Pickup At Min
  Pickup At Max
  Deliver At Min
  Deliver At Max
  Project
  Material Type
  Unit of Measure
  Material Site
  Job Site
  Pickup Location
  Delivery Location
  Created At
  Updated At

Arguments:
  <id>    The project material type ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project material type
  xbe view project-material-types show 123

  # Get JSON output
  xbe view project-material-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectMaterialTypesShow,
	}
	initProjectMaterialTypesShowFlags(cmd)
	return cmd
}

func init() {
	projectMaterialTypesCmd.AddCommand(newProjectMaterialTypesShowCmd())
}

func initProjectMaterialTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectMaterialTypesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectMaterialTypesShowOptions(cmd)
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
		return fmt.Errorf("project material type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-material-types]", "display-name,explicit-display-name,quantity,pickup-at-min,pickup-at-max,deliver-at-min,deliver-at-max,created-at,updated-at")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[material-types]", "display-name,name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[project-transport-locations]", "name")
	query.Set("include", "project,material-type,unit-of-measure,material-site,job-site,pickup-location,delivery-location")

	body, _, err := client.Get(cmd.Context(), "/v1/project-material-types/"+id, query)
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

	details := buildProjectMaterialTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectMaterialTypeDetails(cmd, details)
}

func parseProjectMaterialTypesShowOptions(cmd *cobra.Command) (projectMaterialTypesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectMaterialTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectMaterialTypeDetails(resp jsonAPISingleResponse) projectMaterialTypeDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := projectMaterialTypeDetails{
		ID:                  resp.Data.ID,
		DisplayName:         stringAttr(attrs, "display-name"),
		ExplicitDisplayName: stringAttr(attrs, "explicit-display-name"),
		Quantity:            stringAttr(attrs, "quantity"),
		PickupAtMin:         formatDateTime(stringAttr(attrs, "pickup-at-min")),
		PickupAtMax:         formatDateTime(stringAttr(attrs, "pickup-at-max")),
		DeliverAtMin:        formatDateTime(stringAttr(attrs, "deliver-at-min")),
		DeliverAtMax:        formatDateTime(stringAttr(attrs, "deliver-at-max")),
		CreatedAt:           formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:           formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = firstNonEmpty(
				stringAttr(project.Attributes, "name"),
				stringAttr(project.Attributes, "number"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = firstNonEmpty(
				stringAttr(materialType.Attributes, "display-name"),
				stringAttr(materialType.Attributes, "name"),
			)
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

	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSiteName = stringAttr(site.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobSiteName = stringAttr(site.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["pickup-location"]; ok && rel.Data != nil {
		details.PickupLocationID = rel.Data.ID
		if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.PickupLocationName = stringAttr(location.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["delivery-location"]; ok && rel.Data != nil {
		details.DeliveryLocationID = rel.Data.ID
		if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DeliveryLocationName = stringAttr(location.Attributes, "name")
		}
	}

	return details
}

func renderProjectMaterialTypeDetails(cmd *cobra.Command, details projectMaterialTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DisplayName != "" {
		fmt.Fprintf(out, "Display Name: %s\n", details.DisplayName)
	}
	if details.ExplicitDisplayName != "" {
		fmt.Fprintf(out, "Explicit Display Name: %s\n", details.ExplicitDisplayName)
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.PickupAtMin != "" {
		fmt.Fprintf(out, "Pickup At Min: %s\n", details.PickupAtMin)
	}
	if details.PickupAtMax != "" {
		fmt.Fprintf(out, "Pickup At Max: %s\n", details.PickupAtMax)
	}
	if details.DeliverAtMin != "" {
		fmt.Fprintf(out, "Deliver At Min: %s\n", details.DeliverAtMin)
	}
	if details.DeliverAtMax != "" {
		fmt.Fprintf(out, "Deliver At Max: %s\n", details.DeliverAtMax)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.ProjectName != "" {
		fmt.Fprintf(out, "Project Name: %s\n", details.ProjectName)
	}
	if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type ID: %s\n", details.MaterialTypeID)
	}
	if details.MaterialTypeName != "" {
		fmt.Fprintf(out, "Material Type Name: %s\n", details.MaterialTypeName)
	}
	if details.UnitOfMeasureID != "" {
		fmt.Fprintf(out, "Unit of Measure ID: %s\n", details.UnitOfMeasureID)
	}
	if details.UnitOfMeasure != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", details.UnitOfMeasure)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site ID: %s\n", details.MaterialSiteID)
	}
	if details.MaterialSiteName != "" {
		fmt.Fprintf(out, "Material Site Name: %s\n", details.MaterialSiteName)
	}
	if details.JobSiteID != "" {
		fmt.Fprintf(out, "Job Site ID: %s\n", details.JobSiteID)
	}
	if details.JobSiteName != "" {
		fmt.Fprintf(out, "Job Site Name: %s\n", details.JobSiteName)
	}
	if details.PickupLocationID != "" {
		fmt.Fprintf(out, "Pickup Location ID: %s\n", details.PickupLocationID)
	}
	if details.PickupLocationName != "" {
		fmt.Fprintf(out, "Pickup Location Name: %s\n", details.PickupLocationName)
	}
	if details.DeliveryLocationID != "" {
		fmt.Fprintf(out, "Delivery Location ID: %s\n", details.DeliveryLocationID)
	}
	if details.DeliveryLocationName != "" {
		fmt.Fprintf(out, "Delivery Location Name: %s\n", details.DeliveryLocationName)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
