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

type materialSiteInventoryLocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteInventoryLocationDetails struct {
	ID                 string `json:"id"`
	QualifiedName      string `json:"qualified_name,omitempty"`
	DisplayName        string `json:"display_name_explicit,omitempty"`
	Latitude           string `json:"latitude,omitempty"`
	Longitude          string `json:"longitude,omitempty"`
	MaterialSiteID     string `json:"material_site_id,omitempty"`
	MaterialSite       string `json:"material_site,omitempty"`
	UnitOfMeasureID    string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure      string `json:"unit_of_measure,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
	Broker             string `json:"broker,omitempty"`
	MaterialSupplierID string `json:"material_supplier_id,omitempty"`
	MaterialSupplier   string `json:"material_supplier,omitempty"`
}

func newMaterialSiteInventoryLocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site inventory location details",
		Long: `Show the full details of a material site inventory location.

Arguments:
  <id>  Material site inventory location ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a material site inventory location
  xbe view material-site-inventory-locations show 123

  # JSON output
  xbe view material-site-inventory-locations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteInventoryLocationsShow,
	}
	initMaterialSiteInventoryLocationsShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteInventoryLocationsCmd.AddCommand(newMaterialSiteInventoryLocationsShowCmd())
}

func initMaterialSiteInventoryLocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteInventoryLocationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialSiteInventoryLocationsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("material site inventory location id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-site-inventory-locations]", "qualified-name,display-name-explicit,latitude,longitude,material-site,unit-of-measure,broker,material-supplier")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("include", "material-site,unit-of-measure,broker,material-supplier")

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-inventory-locations/"+id, query)
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

	details := buildMaterialSiteInventoryLocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteInventoryLocationDetails(cmd, details)
}

func parseMaterialSiteInventoryLocationsShowOptions(cmd *cobra.Command) (materialSiteInventoryLocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteInventoryLocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteInventoryLocationDetails(resp jsonAPISingleResponse) materialSiteInventoryLocationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialSiteInventoryLocationDetails{
		ID:            resource.ID,
		QualifiedName: stringAttr(attrs, "qualified-name"),
		DisplayName:   stringAttr(attrs, "display-name-explicit"),
		Latitude:      stringAttr(attrs, "latitude"),
		Longitude:     stringAttr(attrs, "longitude"),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(ms.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = firstNonEmpty(
				stringAttr(uom.Attributes, "abbreviation"),
				stringAttr(uom.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Broker = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
		if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSupplier = stringAttr(supplier.Attributes, "name")
		}
	}

	return details
}

func renderMaterialSiteInventoryLocationDetails(cmd *cobra.Command, details materialSiteInventoryLocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DisplayName != "" {
		fmt.Fprintf(out, "Display Name: %s\n", details.DisplayName)
	}
	if details.QualifiedName != "" {
		fmt.Fprintf(out, "Qualified Name: %s\n", details.QualifiedName)
	}
	if details.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.Latitude)
	}
	if details.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.Longitude)
	}
	if details.MaterialSite != "" {
		fmt.Fprintf(out, "Material Site: %s\n", details.MaterialSite)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site ID: %s\n", details.MaterialSiteID)
	}
	if details.UnitOfMeasure != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", details.UnitOfMeasure)
	}
	if details.UnitOfMeasureID != "" {
		fmt.Fprintf(out, "Unit of Measure ID: %s\n", details.UnitOfMeasureID)
	}
	if details.MaterialSupplier != "" {
		fmt.Fprintf(out, "Material Supplier: %s\n", details.MaterialSupplier)
	}
	if details.MaterialSupplierID != "" {
		fmt.Fprintf(out, "Material Supplier ID: %s\n", details.MaterialSupplierID)
	}
	if details.Broker != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.Broker)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}
