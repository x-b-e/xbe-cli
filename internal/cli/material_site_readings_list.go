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

type materialSiteReadingsListOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	NoAuth                          bool
	Limit                           int
	Offset                          int
	Sort                            string
	MaterialSite                    string
	MaterialSiteMeasure             string
	MaterialSiteReadingMaterialType string
	MaterialType                    string
	MaterialTypeID                  string
	Broker                          string
	ReadingAt                       string
	ReadingAtMin                    string
	ReadingAtMax                    string
	Value                           string
	ValueMin                        string
	ValueMax                        string
}

type materialSiteReadingRow struct {
	ID                                   string  `json:"id"`
	ReadingAt                            string  `json:"reading_at,omitempty"`
	Value                                float64 `json:"value,omitempty"`
	RawMaterialKind                      string  `json:"raw_material_kind,omitempty"`
	MaterialSiteID                       string  `json:"material_site_id,omitempty"`
	MaterialSiteName                     string  `json:"material_site,omitempty"`
	MaterialSiteMeasureID                string  `json:"material_site_measure_id,omitempty"`
	MaterialSiteMeasureName              string  `json:"material_site_measure,omitempty"`
	MaterialSiteMeasureSlug              string  `json:"material_site_measure_slug,omitempty"`
	MaterialSiteReadingMaterialTypeID    string  `json:"material_site_reading_material_type_id,omitempty"`
	MaterialSiteReadingRawMaterialTypeID string  `json:"material_site_reading_raw_material_type_id,omitempty"`
	MaterialTypeID                       string  `json:"material_type_id,omitempty"`
	MaterialTypeName                     string  `json:"material_type,omitempty"`
	RawMaterialTypeID                    string  `json:"raw_material_type_id,omitempty"`
	RawMaterialTypeName                  string  `json:"raw_material_type,omitempty"`
}

func newMaterialSiteReadingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site readings",
		Long: `List material site readings with filtering and pagination.

Output Columns:
  ID         Reading identifier
  READING AT Reading timestamp
  VALUE      Reading value
  MEASURE    Material site measure name/slug
  SITE       Material site name
  TYPE       Material type (if available)
  RAW KIND   Raw material kind (agg, rap, additive, filler)

Filters:
  --material-site                    Filter by material site ID (comma-separated for multiple)
  --material-site-measure            Filter by material site measure ID (comma-separated for multiple)
  --material-site-reading-material-type  Filter by material site reading material type ID (comma-separated for multiple)
  --material-type                    Filter by material type ID (comma-separated for multiple)
  --material-type-id                 Filter by material type ID (join filter)
  --broker                           Filter by broker ID (comma-separated for multiple)
  --reading-at                       Filter by reading timestamp (ISO 8601)
  --reading-at-min                   Filter by minimum reading timestamp (ISO 8601)
  --reading-at-max                   Filter by maximum reading timestamp (ISO 8601)
  --value                            Filter by exact reading value
  --value-min                        Filter by minimum reading value
  --value-max                        Filter by maximum reading value

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material site readings
  xbe view material-site-readings list

  # Filter by material site
  xbe view material-site-readings list --material-site 123

  # Filter by reading timestamp range
  xbe view material-site-readings list --reading-at-min 2025-01-01T00:00:00Z --reading-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view material-site-readings list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialSiteReadingsList,
	}
	initMaterialSiteReadingsListFlags(cmd)
	return cmd
}

func init() {
	materialSiteReadingsCmd.AddCommand(newMaterialSiteReadingsListCmd())
}

func initMaterialSiteReadingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("material-site-measure", "", "Filter by material site measure ID (comma-separated for multiple)")
	cmd.Flags().String("material-site-reading-material-type", "", "Filter by material site reading material type ID (comma-separated for multiple)")
	cmd.Flags().String("material-type", "", "Filter by material type ID (comma-separated for multiple)")
	cmd.Flags().String("material-type-id", "", "Filter by material type ID (join filter)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("reading-at", "", "Filter by reading timestamp (ISO 8601)")
	cmd.Flags().String("reading-at-min", "", "Filter by minimum reading timestamp (ISO 8601)")
	cmd.Flags().String("reading-at-max", "", "Filter by maximum reading timestamp (ISO 8601)")
	cmd.Flags().String("value", "", "Filter by exact reading value")
	cmd.Flags().String("value-min", "", "Filter by minimum reading value")
	cmd.Flags().String("value-max", "", "Filter by maximum reading value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteReadingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteReadingsListOptions(cmd)
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
	query.Set("fields[material-site-readings]", "reading-at,value,raw-material-kind,material-site,material-site-measure,material-site-reading-material-type,material-site-reading-raw-material-type,material-type,raw-material-type")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-site-measures]", "name,slug")
	query.Set("fields[material-types]", "display-name,name")
	query.Set("include", "material-site,material-site-measure,material-type,raw-material-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-site-measure]", opts.MaterialSiteMeasure)
	setFilterIfPresent(query, "filter[material-site-reading-material-type]", opts.MaterialSiteReadingMaterialType)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material-type-id]", opts.MaterialTypeID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[reading-at]", opts.ReadingAt)
	setFilterIfPresent(query, "filter[reading-at-min]", opts.ReadingAtMin)
	setFilterIfPresent(query, "filter[reading-at-max]", opts.ReadingAtMax)
	setFilterIfPresent(query, "filter[value]", opts.Value)
	setFilterIfPresent(query, "filter[value-min]", opts.ValueMin)
	setFilterIfPresent(query, "filter[value-max]", opts.ValueMax)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-readings", query)
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

	rows := buildMaterialSiteReadingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteReadingsTable(cmd, rows)
}

func parseMaterialSiteReadingsListOptions(cmd *cobra.Command) (materialSiteReadingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialSiteMeasure, _ := cmd.Flags().GetString("material-site-measure")
	materialSiteReadingMaterialType, _ := cmd.Flags().GetString("material-site-reading-material-type")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialTypeID, _ := cmd.Flags().GetString("material-type-id")
	broker, _ := cmd.Flags().GetString("broker")
	readingAt, _ := cmd.Flags().GetString("reading-at")
	readingAtMin, _ := cmd.Flags().GetString("reading-at-min")
	readingAtMax, _ := cmd.Flags().GetString("reading-at-max")
	value, _ := cmd.Flags().GetString("value")
	valueMin, _ := cmd.Flags().GetString("value-min")
	valueMax, _ := cmd.Flags().GetString("value-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteReadingsListOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		NoAuth:                          noAuth,
		Limit:                           limit,
		Offset:                          offset,
		Sort:                            sort,
		MaterialSite:                    materialSite,
		MaterialSiteMeasure:             materialSiteMeasure,
		MaterialSiteReadingMaterialType: materialSiteReadingMaterialType,
		MaterialType:                    materialType,
		MaterialTypeID:                  materialTypeID,
		Broker:                          broker,
		ReadingAt:                       readingAt,
		ReadingAtMin:                    readingAtMin,
		ReadingAtMax:                    readingAtMax,
		Value:                           value,
		ValueMin:                        valueMin,
		ValueMax:                        valueMax,
	}, nil
}

func buildMaterialSiteReadingRows(resp jsonAPIResponse) []materialSiteReadingRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialSiteReadingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := materialSiteReadingRow{
			ID:              resource.ID,
			ReadingAt:       formatDateTime(stringAttr(attrs, "reading-at")),
			Value:           floatAttr(attrs, "value"),
			RawMaterialKind: stringAttr(attrs, "raw-material-kind"),
		}

		row.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
		if row.MaterialSiteID != "" {
			if site, ok := included[resourceKey("material-sites", row.MaterialSiteID)]; ok {
				row.MaterialSiteName = stringAttr(site.Attributes, "name")
			}
		}

		row.MaterialSiteMeasureID = relationshipIDFromMap(resource.Relationships, "material-site-measure")
		if row.MaterialSiteMeasureID != "" {
			if measure, ok := included[resourceKey("material-site-measures", row.MaterialSiteMeasureID)]; ok {
				row.MaterialSiteMeasureName = stringAttr(measure.Attributes, "name")
				row.MaterialSiteMeasureSlug = stringAttr(measure.Attributes, "slug")
			}
		}

		row.MaterialSiteReadingMaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-site-reading-material-type")
		row.MaterialSiteReadingRawMaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-site-reading-raw-material-type")

		row.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
		if row.MaterialTypeID != "" {
			if mt, ok := included[resourceKey("material-types", row.MaterialTypeID)]; ok {
				row.MaterialTypeName = firstNonEmpty(
					stringAttr(mt.Attributes, "display-name"),
					stringAttr(mt.Attributes, "name"),
				)
			}
		}

		row.RawMaterialTypeID = relationshipIDFromMap(resource.Relationships, "raw-material-type")
		if row.RawMaterialTypeID != "" {
			if mt, ok := included[resourceKey("material-types", row.RawMaterialTypeID)]; ok {
				row.RawMaterialTypeName = firstNonEmpty(
					stringAttr(mt.Attributes, "display-name"),
					stringAttr(mt.Attributes, "name"),
				)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialSiteReadingsTable(cmd *cobra.Command, rows []materialSiteReadingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site readings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREADING AT\tVALUE\tMEASURE\tSITE\tTYPE\tRAW KIND")
	for _, row := range rows {
		measure := firstNonEmpty(row.MaterialSiteMeasureName, row.MaterialSiteMeasureSlug, row.MaterialSiteMeasureID)
		site := firstNonEmpty(row.MaterialSiteName, row.MaterialSiteID)
		materialType := firstNonEmpty(row.MaterialTypeName, row.RawMaterialTypeName, row.MaterialTypeID, row.RawMaterialTypeID)
		value := fmt.Sprintf("%.2f", row.Value)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ReadingAt,
			value,
			measure,
			site,
			materialType,
			row.RawMaterialKind,
		)
	}

	return writer.Flush()
}
