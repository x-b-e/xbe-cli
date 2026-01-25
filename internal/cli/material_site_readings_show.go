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

type materialSiteReadingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteReadingDetails struct {
	ID                                           string  `json:"id"`
	ReadingAt                                    string  `json:"reading_at,omitempty"`
	Value                                        float64 `json:"value,omitempty"`
	RawMaterialKind                              string  `json:"raw_material_kind,omitempty"`
	RawMaterialDescription                       string  `json:"raw_material_description,omitempty"`
	RawMaterialFeederNumber                      string  `json:"raw_material_feeder_number,omitempty"`
	MaterialSiteID                               string  `json:"material_site_id,omitempty"`
	MaterialSiteName                             string  `json:"material_site,omitempty"`
	MaterialSiteMeasureID                        string  `json:"material_site_measure_id,omitempty"`
	MaterialSiteMeasureName                      string  `json:"material_site_measure,omitempty"`
	MaterialSiteMeasureSlug                      string  `json:"material_site_measure_slug,omitempty"`
	MaterialSiteMeasureValidMin                  float64 `json:"material_site_measure_valid_min,omitempty"`
	MaterialSiteMeasureValidMax                  float64 `json:"material_site_measure_valid_max,omitempty"`
	MaterialSiteReadingMaterialTypeID            string  `json:"material_site_reading_material_type_id,omitempty"`
	MaterialSiteReadingMaterialTypeExternalID    string  `json:"material_site_reading_material_type_external_id,omitempty"`
	MaterialSiteReadingRawMaterialTypeID         string  `json:"material_site_reading_raw_material_type_id,omitempty"`
	MaterialSiteReadingRawMaterialTypeExternalID string  `json:"material_site_reading_raw_material_type_external_id,omitempty"`
	MaterialTypeID                               string  `json:"material_type_id,omitempty"`
	MaterialTypeName                             string  `json:"material_type,omitempty"`
	RawMaterialTypeID                            string  `json:"raw_material_type_id,omitempty"`
	RawMaterialTypeName                          string  `json:"raw_material_type,omitempty"`
	MaterialSupplierID                           string  `json:"material_supplier_id,omitempty"`
	MaterialSupplierName                         string  `json:"material_supplier,omitempty"`
	BrokerID                                     string  `json:"broker_id,omitempty"`
	BrokerName                                   string  `json:"broker,omitempty"`
	CreatedAt                                    string  `json:"created_at,omitempty"`
	UpdatedAt                                    string  `json:"updated_at,omitempty"`
}

func newMaterialSiteReadingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site reading details",
		Long: `Show the full details of a material site reading.

Output Fields:
  ID
  Reading At
  Value
  Raw Material Kind
  Raw Material Description
  Raw Material Feeder Number
  Material Site
  Material Site Measure
  Material Site Reading Material Type
  Material Site Reading Raw Material Type
  Material Type
  Raw Material Type
  Material Supplier
  Broker
  Created At
  Updated At

Arguments:
  <id>    The material site reading ID (required). You can find IDs using the list command.`,
		Example: `  # Show a material site reading
  xbe view material-site-readings show 123

  # Get JSON output
  xbe view material-site-readings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteReadingsShow,
	}
	initMaterialSiteReadingsShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteReadingsCmd.AddCommand(newMaterialSiteReadingsShowCmd())
}

func initMaterialSiteReadingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteReadingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialSiteReadingsShowOptions(cmd)
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
		return fmt.Errorf("material site reading id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-site-readings]", "reading-at,value,raw-material-kind,raw-material-description,raw-material-feeder-number,material-site,material-site-measure,material-site-reading-material-type,material-site-reading-raw-material-type,material-type,raw-material-type,material-supplier,broker,created-at,updated-at")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-site-measures]", "name,slug,valid-reading-value-min,valid-reading-value-max")
	query.Set("fields[material-site-reading-material-types]", "external-id,material-type")
	query.Set("fields[material-types]", "display-name,name")
	query.Set("fields[material-suppliers]", "name,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "material-site,material-site-measure,material-site-reading-material-type,material-site-reading-material-type.material-type,material-site-reading-raw-material-type,material-site-reading-raw-material-type.material-type,material-type,raw-material-type,material-supplier,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-readings/"+id, query)
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

	details := buildMaterialSiteReadingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteReadingDetails(cmd, details)
}

func parseMaterialSiteReadingsShowOptions(cmd *cobra.Command) (materialSiteReadingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteReadingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteReadingDetails(resp jsonAPISingleResponse) materialSiteReadingDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := materialSiteReadingDetails{
		ID:                      resource.ID,
		ReadingAt:               formatDateTime(stringAttr(attrs, "reading-at")),
		Value:                   floatAttr(attrs, "value"),
		RawMaterialKind:         stringAttr(attrs, "raw-material-kind"),
		RawMaterialDescription:  stringAttr(attrs, "raw-material-description"),
		RawMaterialFeederNumber: stringAttr(attrs, "raw-material-feeder-number"),
		CreatedAt:               formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:               formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
	details.MaterialSiteMeasureID = relationshipIDFromMap(resource.Relationships, "material-site-measure")
	details.MaterialSiteReadingMaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-site-reading-material-type")
	details.MaterialSiteReadingRawMaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-site-reading-raw-material-type")
	details.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
	details.RawMaterialTypeID = relationshipIDFromMap(resource.Relationships, "raw-material-type")
	details.MaterialSupplierID = relationshipIDFromMap(resource.Relationships, "material-supplier")
	details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.MaterialSiteID != "" {
		if site, ok := included[resourceKey("material-sites", details.MaterialSiteID)]; ok {
			details.MaterialSiteName = stringAttr(site.Attributes, "name")
		}
	}

	if details.MaterialSiteMeasureID != "" {
		if measure, ok := included[resourceKey("material-site-measures", details.MaterialSiteMeasureID)]; ok {
			details.MaterialSiteMeasureName = stringAttr(measure.Attributes, "name")
			details.MaterialSiteMeasureSlug = stringAttr(measure.Attributes, "slug")
			details.MaterialSiteMeasureValidMin = floatAttr(measure.Attributes, "valid-reading-value-min")
			details.MaterialSiteMeasureValidMax = floatAttr(measure.Attributes, "valid-reading-value-max")
		}
	}

	if details.MaterialTypeID != "" {
		if mt, ok := included[resourceKey("material-types", details.MaterialTypeID)]; ok {
			details.MaterialTypeName = firstNonEmpty(
				stringAttr(mt.Attributes, "display-name"),
				stringAttr(mt.Attributes, "name"),
			)
		}
	}

	if details.RawMaterialTypeID != "" {
		if mt, ok := included[resourceKey("material-types", details.RawMaterialTypeID)]; ok {
			details.RawMaterialTypeName = firstNonEmpty(
				stringAttr(mt.Attributes, "display-name"),
				stringAttr(mt.Attributes, "name"),
			)
		}
	}

	if details.MaterialSiteReadingMaterialTypeID != "" {
		if msrmt, ok := included[resourceKey("material-site-reading-material-types", details.MaterialSiteReadingMaterialTypeID)]; ok {
			details.MaterialSiteReadingMaterialTypeExternalID = stringAttr(msrmt.Attributes, "external-id")
			if details.MaterialTypeName == "" {
				if rel, ok := msrmt.Relationships["material-type"]; ok && rel.Data != nil {
					details.MaterialTypeID = rel.Data.ID
					if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
						details.MaterialTypeName = firstNonEmpty(
							stringAttr(mt.Attributes, "display-name"),
							stringAttr(mt.Attributes, "name"),
						)
					}
				}
			}
		}
	}

	if details.MaterialSiteReadingRawMaterialTypeID != "" {
		if msrmt, ok := included[resourceKey("material-site-reading-material-types", details.MaterialSiteReadingRawMaterialTypeID)]; ok {
			details.MaterialSiteReadingRawMaterialTypeExternalID = stringAttr(msrmt.Attributes, "external-id")
			if details.RawMaterialTypeName == "" {
				if rel, ok := msrmt.Relationships["material-type"]; ok && rel.Data != nil {
					details.RawMaterialTypeID = rel.Data.ID
					if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
						details.RawMaterialTypeName = firstNonEmpty(
							stringAttr(mt.Attributes, "display-name"),
							stringAttr(mt.Attributes, "name"),
						)
					}
				}
			}
		}
	}

	if details.MaterialSupplierID != "" {
		if supplier, ok := included[resourceKey("material-suppliers", details.MaterialSupplierID)]; ok {
			details.MaterialSupplierName = stringAttr(supplier.Attributes, "name")
		}
	}

	if details.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", details.BrokerID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	return details
}

func renderMaterialSiteReadingDetails(cmd *cobra.Command, details materialSiteReadingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Reading At: %s\n", details.ReadingAt)
	fmt.Fprintf(out, "Value: %.2f\n", details.Value)

	if details.RawMaterialKind != "" {
		fmt.Fprintf(out, "Raw Material Kind: %s\n", details.RawMaterialKind)
	}
	if details.RawMaterialDescription != "" {
		fmt.Fprintf(out, "Raw Material Description: %s\n", details.RawMaterialDescription)
	}
	if details.RawMaterialFeederNumber != "" {
		fmt.Fprintf(out, "Raw Material Feeder Number: %s\n", details.RawMaterialFeederNumber)
	}

	if details.MaterialSiteID != "" {
		label := details.MaterialSiteID
		if details.MaterialSiteName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSiteName, details.MaterialSiteID)
		}
		fmt.Fprintf(out, "Material Site: %s\n", label)
	}

	if details.MaterialSiteMeasureID != "" {
		label := details.MaterialSiteMeasureID
		name := details.MaterialSiteMeasureName
		if name == "" {
			name = details.MaterialSiteMeasureSlug
		}
		if name != "" {
			label = fmt.Sprintf("%s (%s)", name, details.MaterialSiteMeasureID)
		}
		fmt.Fprintf(out, "Material Site Measure: %s\n", label)
		if details.MaterialSiteMeasureValidMin != 0 || details.MaterialSiteMeasureValidMax != 0 {
			fmt.Fprintf(out, "Material Site Measure Valid Range: %.2f - %.2f\n", details.MaterialSiteMeasureValidMin, details.MaterialSiteMeasureValidMax)
		}
	}

	if details.MaterialSiteReadingMaterialTypeID != "" {
		label := details.MaterialSiteReadingMaterialTypeID
		if details.MaterialSiteReadingMaterialTypeExternalID != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSiteReadingMaterialTypeExternalID, details.MaterialSiteReadingMaterialTypeID)
		}
		fmt.Fprintf(out, "Material Site Reading Material Type: %s\n", label)
	}

	if details.MaterialSiteReadingRawMaterialTypeID != "" {
		label := details.MaterialSiteReadingRawMaterialTypeID
		if details.MaterialSiteReadingRawMaterialTypeExternalID != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSiteReadingRawMaterialTypeExternalID, details.MaterialSiteReadingRawMaterialTypeID)
		}
		fmt.Fprintf(out, "Material Site Reading Raw Material Type: %s\n", label)
	}

	if details.MaterialTypeID != "" {
		label := details.MaterialTypeID
		if details.MaterialTypeName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialTypeName, details.MaterialTypeID)
		}
		fmt.Fprintf(out, "Material Type: %s\n", label)
	}

	if details.RawMaterialTypeID != "" {
		label := details.RawMaterialTypeID
		if details.RawMaterialTypeName != "" {
			label = fmt.Sprintf("%s (%s)", details.RawMaterialTypeName, details.RawMaterialTypeID)
		}
		fmt.Fprintf(out, "Raw Material Type: %s\n", label)
	}

	if details.MaterialSupplierID != "" {
		label := details.MaterialSupplierID
		if details.MaterialSupplierName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSupplierName, details.MaterialSupplierID)
		}
		fmt.Fprintf(out, "Material Supplier: %s\n", label)
	}

	if details.BrokerID != "" {
		label := details.BrokerID
		if details.BrokerName != "" {
			label = fmt.Sprintf("%s (%s)", details.BrokerName, details.BrokerID)
		}
		fmt.Fprintf(out, "Broker: %s\n", label)
	}

	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
