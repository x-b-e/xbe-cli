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

type doMaterialSiteReadingsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	MaterialSite                       string
	MaterialSiteMeasure                string
	MaterialSiteReadingMaterialType    string
	MaterialSiteReadingRawMaterialType string
	ReadingAt                          string
	Value                              string
	RawMaterialKind                    string
	RawMaterialDescription             string
	RawMaterialFeederNumber            string
}

func newDoMaterialSiteReadingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material site reading",
		Long: `Create a material site reading.

Required flags:
  --material-site          Material site ID
  --material-site-measure  Material site measure ID
  --reading-at             Reading timestamp (ISO 8601)
  --value                  Reading value

Optional flags:
  --material-site-reading-material-type     Material site reading material type ID
  --material-site-reading-raw-material-type Material site reading raw material type ID
  --raw-material-kind         Raw material kind (agg, rap, additive, filler)
  --raw-material-description  Raw material description
  --raw-material-feeder-number Raw material feeder number`,
		Example: `  # Create a material site reading
  xbe do material-site-readings create \
    --material-site 123 \
    --material-site-measure 456 \
    --reading-at 2025-01-15T12:00:00Z \
    --value 12.5

  # Create a raw material reading
  xbe do material-site-readings create \
    --material-site 123 \
    --material-site-measure 456 \
    --reading-at 2025-01-15T12:00:00Z \
    --value 4.2 \
    --raw-material-kind agg \
    --raw-material-description "Aggregate Bin 1" \
    --raw-material-feeder-number 1`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialSiteReadingsCreate,
	}
	initDoMaterialSiteReadingsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteReadingsCmd.AddCommand(newDoMaterialSiteReadingsCreateCmd())
}

func initDoMaterialSiteReadingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("material-site-measure", "", "Material site measure ID (required)")
	cmd.Flags().String("material-site-reading-material-type", "", "Material site reading material type ID")
	cmd.Flags().String("material-site-reading-raw-material-type", "", "Material site reading raw material type ID")
	cmd.Flags().String("reading-at", "", "Reading timestamp (ISO 8601, required)")
	cmd.Flags().String("value", "", "Reading value (required)")
	cmd.Flags().String("raw-material-kind", "", "Raw material kind (agg, rap, additive, filler)")
	cmd.Flags().String("raw-material-description", "", "Raw material description")
	cmd.Flags().String("raw-material-feeder-number", "", "Raw material feeder number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("material-site")
	cmd.MarkFlagRequired("material-site-measure")
	cmd.MarkFlagRequired("reading-at")
	cmd.MarkFlagRequired("value")
}

func runDoMaterialSiteReadingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSiteReadingsCreateOptions(cmd)
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

	attributes := map[string]any{
		"reading-at": opts.ReadingAt,
		"value":      opts.Value,
	}
	if opts.RawMaterialKind != "" {
		attributes["raw-material-kind"] = opts.RawMaterialKind
	}
	if opts.RawMaterialDescription != "" {
		attributes["raw-material-description"] = opts.RawMaterialDescription
	}
	if opts.RawMaterialFeederNumber != "" {
		attributes["raw-material-feeder-number"] = opts.RawMaterialFeederNumber
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		},
		"material-site-measure": map[string]any{
			"data": map[string]any{
				"type": "material-site-measures",
				"id":   opts.MaterialSiteMeasure,
			},
		},
	}
	if opts.MaterialSiteReadingMaterialType != "" {
		relationships["material-site-reading-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-site-reading-material-types",
				"id":   opts.MaterialSiteReadingMaterialType,
			},
		}
	}
	if opts.MaterialSiteReadingRawMaterialType != "" {
		relationships["material-site-reading-raw-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-site-reading-material-types",
				"id":   opts.MaterialSiteReadingRawMaterialType,
			},
		}
	}

	data := map[string]any{
		"type":          "material-site-readings",
		"attributes":    attributes,
		"relationships": relationships,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-readings", jsonBody)
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

	row := materialSiteReadingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site reading %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteReadingsCreateOptions(cmd *cobra.Command) (doMaterialSiteReadingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialSiteMeasure, _ := cmd.Flags().GetString("material-site-measure")
	materialSiteReadingMaterialType, _ := cmd.Flags().GetString("material-site-reading-material-type")
	materialSiteReadingRawMaterialType, _ := cmd.Flags().GetString("material-site-reading-raw-material-type")
	readingAt, _ := cmd.Flags().GetString("reading-at")
	value, _ := cmd.Flags().GetString("value")
	rawMaterialKind, _ := cmd.Flags().GetString("raw-material-kind")
	rawMaterialDescription, _ := cmd.Flags().GetString("raw-material-description")
	rawMaterialFeederNumber, _ := cmd.Flags().GetString("raw-material-feeder-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteReadingsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		MaterialSite:                       materialSite,
		MaterialSiteMeasure:                materialSiteMeasure,
		MaterialSiteReadingMaterialType:    materialSiteReadingMaterialType,
		MaterialSiteReadingRawMaterialType: materialSiteReadingRawMaterialType,
		ReadingAt:                          readingAt,
		Value:                              value,
		RawMaterialKind:                    rawMaterialKind,
		RawMaterialDescription:             rawMaterialDescription,
		RawMaterialFeederNumber:            rawMaterialFeederNumber,
	}, nil
}

func materialSiteReadingRowFromSingle(resp jsonAPISingleResponse) materialSiteReadingRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := materialSiteReadingRow{
		ID:                                   resource.ID,
		ReadingAt:                            formatDateTime(stringAttr(attrs, "reading-at")),
		Value:                                floatAttr(attrs, "value"),
		RawMaterialKind:                      stringAttr(attrs, "raw-material-kind"),
		MaterialSiteID:                       relationshipIDFromMap(resource.Relationships, "material-site"),
		MaterialSiteMeasureID:                relationshipIDFromMap(resource.Relationships, "material-site-measure"),
		MaterialSiteReadingMaterialTypeID:    relationshipIDFromMap(resource.Relationships, "material-site-reading-material-type"),
		MaterialSiteReadingRawMaterialTypeID: relationshipIDFromMap(resource.Relationships, "material-site-reading-raw-material-type"),
		MaterialTypeID:                       relationshipIDFromMap(resource.Relationships, "material-type"),
		RawMaterialTypeID:                    relationshipIDFromMap(resource.Relationships, "raw-material-type"),
	}

	return row
}
