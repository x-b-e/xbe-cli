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

type projectBidLocationMaterialTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectBidLocationMaterialTypeDetails struct {
	ID                        string `json:"id"`
	Quantity                  string `json:"quantity,omitempty"`
	Notes                     string `json:"notes,omitempty"`
	ProjectBidLocationID      string `json:"project_bid_location_id,omitempty"`
	ProjectBidLocationName    string `json:"project_bid_location_name,omitempty"`
	MaterialTypeID            string `json:"material_type_id,omitempty"`
	MaterialTypeName          string `json:"material_type_name,omitempty"`
	UnitOfMeasureID           string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName         string `json:"unit_of_measure_name,omitempty"`
	UnitOfMeasureAbbreviation string `json:"unit_of_measure_abbreviation,omitempty"`
	ProjectID                 string `json:"project_id,omitempty"`
	ProjectName               string `json:"project_name,omitempty"`
	ProjectNumber             string `json:"project_number,omitempty"`
}

func newProjectBidLocationMaterialTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project bid location material type details",
		Long: `Show the full details of a project bid location material type.

Project bid location material types define planned quantities and notes for a
material type at a specific project bid location.

Output Fields:
  ID                     Project bid location material type identifier
  Quantity               Planned quantity
  Notes                  Notes
  Project Bid Location   Project bid location name (or ID)
  Material Type          Material type name (or ID)
  Unit of Measure        Unit of measure
  Project                Project name (or ID)

Arguments:
  <id>                   The project bid location material type ID (required).`,
		Example: `  # Show a project bid location material type
  xbe view project-bid-location-material-types show 123

  # Show as JSON
  xbe view project-bid-location-material-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectBidLocationMaterialTypesShow,
	}
	initProjectBidLocationMaterialTypesShowFlags(cmd)
	return cmd
}

func init() {
	projectBidLocationMaterialTypesCmd.AddCommand(newProjectBidLocationMaterialTypesShowCmd())
}

func initProjectBidLocationMaterialTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectBidLocationMaterialTypesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectBidLocationMaterialTypesShowOptions(cmd)
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
		return fmt.Errorf("project bid location material type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-bid-location-material-types]", "quantity,notes,project-bid-location,material-type,unit-of-measure,project")
	query.Set("include", "project-bid-location,material-type,unit-of-measure,project")
	query.Set("fields[project-bid-locations]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[projects]", "name,number")

	body, _, err := client.Get(cmd.Context(), "/v1/project-bid-location-material-types/"+id, query)
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

	details := buildProjectBidLocationMaterialTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectBidLocationMaterialTypeDetails(cmd, details)
}

func parseProjectBidLocationMaterialTypesShowOptions(cmd *cobra.Command) (projectBidLocationMaterialTypesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectBidLocationMaterialTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectBidLocationMaterialTypeDetails(resp jsonAPISingleResponse) projectBidLocationMaterialTypeDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectBidLocationMaterialTypeDetails{
		ID:       resp.Data.ID,
		Quantity: stringAttr(attrs, "quantity"),
		Notes:    stringAttr(attrs, "notes"),
	}

	if rel, ok := resp.Data.Relationships["project-bid-location"]; ok && rel.Data != nil {
		details.ProjectBidLocationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectBidLocationName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasureName = stringAttr(inc.Attributes, "name")
			details.UnitOfMeasureAbbreviation = stringAttr(inc.Attributes, "abbreviation")
		}
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(inc.Attributes, "name")
			details.ProjectNumber = stringAttr(inc.Attributes, "number")
		}
	}

	return details
}

func renderProjectBidLocationMaterialTypeDetails(cmd *cobra.Command, details projectBidLocationMaterialTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}

	projectBidLocationLabel := firstNonEmpty(details.ProjectBidLocationName, details.ProjectBidLocationID)
	if projectBidLocationLabel != "" {
		fmt.Fprintf(out, "Project Bid Location: %s\n", projectBidLocationLabel)
	}

	materialTypeLabel := firstNonEmpty(details.MaterialTypeName, details.MaterialTypeID)
	if materialTypeLabel != "" {
		fmt.Fprintf(out, "Material Type: %s\n", materialTypeLabel)
	}

	unitLabel := firstNonEmpty(details.UnitOfMeasureAbbreviation, details.UnitOfMeasureName, details.UnitOfMeasureID)
	if unitLabel != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", unitLabel)
	}

	projectLabel := firstNonEmpty(details.ProjectName, details.ProjectNumber, details.ProjectID)
	if projectLabel != "" {
		fmt.Fprintf(out, "Project: %s\n", projectLabel)
	}

	return nil
}
