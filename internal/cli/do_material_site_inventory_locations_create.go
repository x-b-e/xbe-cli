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

type doMaterialSiteInventoryLocationsCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	QualifiedName   string
	DisplayName     string
	Latitude        string
	Longitude       string
	UnitOfMeasureID string
	MaterialSiteID  string
}

func newDoMaterialSiteInventoryLocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material site inventory location",
		Long: `Create a material site inventory location.

Required flags:
  --material-site    Material site ID

Optional flags:
  --qualified-name         Qualified name (unique within the material site)
  --display-name-explicit  Display name override
  --latitude               Latitude coordinate (use with --longitude)
  --longitude              Longitude coordinate (use with --latitude)
  --unit-of-measure        Unit of measure ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create with required fields
  xbe do material-site-inventory-locations create --material-site 123 --qualified-name "Plant A Stockpile"

  # Create with display name and coordinates
  xbe do material-site-inventory-locations create --material-site 123 \
    --qualified-name "Stockpile A" \
    --display-name-explicit "Stockpile A" \
    --latitude 41.881 --longitude -87.623`,
		RunE: runDoMaterialSiteInventoryLocationsCreate,
	}
	initDoMaterialSiteInventoryLocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteInventoryLocationsCmd.AddCommand(newDoMaterialSiteInventoryLocationsCreateCmd())
}

func initDoMaterialSiteInventoryLocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("qualified-name", "", "Qualified name (unique within the material site)")
	cmd.Flags().String("display-name-explicit", "", "Display name override")
	cmd.Flags().String("latitude", "", "Latitude coordinate")
	cmd.Flags().String("longitude", "", "Longitude coordinate")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("material-site")
}

func runDoMaterialSiteInventoryLocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSiteInventoryLocationsCreateOptions(cmd)
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

	if (opts.Latitude != "" && opts.Longitude == "") || (opts.Latitude == "" && opts.Longitude != "") {
		err := fmt.Errorf("latitude and longitude must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.QualifiedName != "" {
		attributes["qualified-name"] = opts.QualifiedName
	}
	if opts.DisplayName != "" {
		attributes["display-name-explicit"] = opts.DisplayName
	}
	if opts.Latitude != "" && opts.Longitude != "" {
		attributes["latitude"] = opts.Latitude
		attributes["longitude"] = opts.Longitude
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSiteID,
			},
		},
	}

	if opts.UnitOfMeasureID != "" {
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasureID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-site-inventory-locations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-inventory-locations", jsonBody)
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

	row := materialSiteInventoryLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site inventory location %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteInventoryLocationsCreateOptions(cmd *cobra.Command) (doMaterialSiteInventoryLocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	qualifiedName, _ := cmd.Flags().GetString("qualified-name")
	displayName, _ := cmd.Flags().GetString("display-name-explicit")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	unitOfMeasureID, _ := cmd.Flags().GetString("unit-of-measure")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteInventoryLocationsCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		QualifiedName:   qualifiedName,
		DisplayName:     displayName,
		Latitude:        latitude,
		Longitude:       longitude,
		UnitOfMeasureID: unitOfMeasureID,
		MaterialSiteID:  materialSiteID,
	}, nil
}

func materialSiteInventoryLocationRowFromSingle(resp jsonAPISingleResponse) materialSiteInventoryLocationRow {
	resource := resp.Data
	attrs := resource.Attributes

	row := materialSiteInventoryLocationRow{
		ID:            resource.ID,
		QualifiedName: stringAttr(attrs, "qualified-name"),
		DisplayName:   stringAttr(attrs, "display-name-explicit"),
		Latitude:      stringAttr(attrs, "latitude"),
		Longitude:     stringAttr(attrs, "longitude"),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.MaterialSupplierID = rel.Data.ID
	}

	return row
}
