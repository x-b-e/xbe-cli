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
	ID                                      string   `json:"id"`
	DisplayName                             string   `json:"display_name,omitempty"`
	ExplicitDisplayName                     string   `json:"explicit_display_name,omitempty"`
	Quantity                                string   `json:"quantity,omitempty"`
	PickupAtMin                             string   `json:"pickup_at_min,omitempty"`
	PickupAtMax                             string   `json:"pickup_at_max,omitempty"`
	DeliverAtMin                            string   `json:"deliver_at_min,omitempty"`
	DeliverAtMax                            string   `json:"deliver_at_max,omitempty"`
	ProjectID                               string   `json:"project_id,omitempty"`
	ProjectName                             string   `json:"project_name,omitempty"`
	ProjectNumber                           string   `json:"project_number,omitempty"`
	MaterialTypeID                          string   `json:"material_type_id,omitempty"`
	MaterialTypeName                        string   `json:"material_type_name,omitempty"`
	UnitOfMeasureID                         string   `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName                       string   `json:"unit_of_measure_name,omitempty"`
	UnitOfMeasureAbbreviation               string   `json:"unit_of_measure_abbreviation,omitempty"`
	MaterialSiteID                          string   `json:"material_site_id,omitempty"`
	MaterialSiteName                        string   `json:"material_site_name,omitempty"`
	JobSiteID                               string   `json:"job_site_id,omitempty"`
	JobSiteName                             string   `json:"job_site_name,omitempty"`
	PickupLocationID                        string   `json:"pickup_location_id,omitempty"`
	PickupLocationName                      string   `json:"pickup_location_name,omitempty"`
	DeliveryLocationID                      string   `json:"delivery_location_id,omitempty"`
	DeliveryLocationName                    string   `json:"delivery_location_name,omitempty"`
	ProjectMaterialTypeQualityControlReqIDs []string `json:"project_material_type_quality_control_requirement_ids,omitempty"`
}

func newProjectMaterialTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project material type details",
		Long: `Show the full details of a project material type.

Project material types define material requirements for a project and can be
scoped to material sites, job sites, or transport-only pickup/delivery locations.

Output Fields:
  ID                                  Project material type identifier
  Display Name                        Display name (explicit or material type)
  Explicit Display Name               Explicit display name override
  Quantity                            Quantity
  Pickup At Min                       Earliest pickup time
  Pickup At Max                       Latest pickup time
  Deliver At Min                      Earliest delivery time
  Deliver At Max                      Latest delivery time
  Project                             Project name/number (or ID)
  Material Type                       Material type name (or ID)
  Unit of Measure                     Unit of measure
  Material Site                       Material site name (or ID)
  Job Site                            Job site name (or ID)
  Pickup Location                     Pickup location name (or ID)
  Delivery Location                   Delivery location name (or ID)
  Quality Control Requirement IDs     Associated QC requirement IDs

Arguments:
  <id>                                The project material type ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project material type
  xbe view project-material-types show 123

  # Show as JSON
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
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
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
	query.Set("fields[project-material-types]", "quantity,explicit-display-name,display-name,pickup-at-min,pickup-at-max,deliver-at-min,deliver-at-max,project,material-type,unit-of-measure,material-site,job-site,pickup-location,delivery-location,project-material-type-quality-control-requirements")
	query.Set("include", "project,material-type,unit-of-measure,material-site,job-site,pickup-location,delivery-location")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[material-types]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[project-transport-locations]", "name")

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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
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
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectMaterialTypeDetails{
		ID:                  resource.ID,
		DisplayName:         stringAttr(attrs, "display-name"),
		ExplicitDisplayName: stringAttr(attrs, "explicit-display-name"),
		Quantity:            stringAttr(attrs, "quantity"),
		PickupAtMin:         formatDateTime(stringAttr(attrs, "pickup-at-min")),
		PickupAtMax:         formatDateTime(stringAttr(attrs, "pickup-at-max")),
		DeliverAtMin:        formatDateTime(stringAttr(attrs, "deliver-at-min")),
		DeliverAtMax:        formatDateTime(stringAttr(attrs, "deliver-at-max")),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(inc.Attributes, "name")
			details.ProjectNumber = stringAttr(inc.Attributes, "number")
		}
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasureName = stringAttr(inc.Attributes, "name")
			details.UnitOfMeasureAbbreviation = stringAttr(inc.Attributes, "abbreviation")
		}
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSiteName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobSiteName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["pickup-location"]; ok && rel.Data != nil {
		details.PickupLocationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.PickupLocationName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["delivery-location"]; ok && rel.Data != nil {
		details.DeliveryLocationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DeliveryLocationName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["project-material-type-quality-control-requirements"]; ok {
		details.ProjectMaterialTypeQualityControlReqIDs = relationshipIDList(rel)
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

	projectLabel := firstNonEmpty(details.ProjectName, details.ProjectNumber, details.ProjectID)
	if projectLabel != "" {
		fmt.Fprintf(out, "Project: %s\n", projectLabel)
	}

	materialTypeLabel := firstNonEmpty(details.MaterialTypeName, details.MaterialTypeID)
	if materialTypeLabel != "" {
		fmt.Fprintf(out, "Material Type: %s\n", materialTypeLabel)
	}

	unitLabel := firstNonEmpty(details.UnitOfMeasureAbbreviation, details.UnitOfMeasureName, details.UnitOfMeasureID)
	if unitLabel != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", unitLabel)
	}

	materialSiteLabel := firstNonEmpty(details.MaterialSiteName, details.MaterialSiteID)
	if materialSiteLabel != "" {
		fmt.Fprintf(out, "Material Site: %s\n", materialSiteLabel)
	}

	jobSiteLabel := firstNonEmpty(details.JobSiteName, details.JobSiteID)
	if jobSiteLabel != "" {
		fmt.Fprintf(out, "Job Site: %s\n", jobSiteLabel)
	}

	pickupLabel := firstNonEmpty(details.PickupLocationName, details.PickupLocationID)
	if pickupLabel != "" {
		fmt.Fprintf(out, "Pickup Location: %s\n", pickupLabel)
	}

	deliveryLabel := firstNonEmpty(details.DeliveryLocationName, details.DeliveryLocationID)
	if deliveryLabel != "" {
		fmt.Fprintf(out, "Delivery Location: %s\n", deliveryLabel)
	}

	if len(details.ProjectMaterialTypeQualityControlReqIDs) > 0 {
		fmt.Fprintf(out, "Quality Control Requirement IDs: %s\n", strings.Join(details.ProjectMaterialTypeQualityControlReqIDs, ", "))
	}

	return nil
}
