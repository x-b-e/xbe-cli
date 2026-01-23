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

type materialSiteMixingLotsListOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	NoAuth                          bool
	Limit                           int
	Offset                          int
	Sort                            string
	MaterialSite                    string
	MaterialSupplierID              string
	MaterialSupplier                string
	MaterialSiteReadingMaterialType string
	MaterialTypeID                  string
	MaterialType                    string
	Broker                          string
	StartAt                         string
	EndAt                           string
	StartOnCached                   string
	TonsPerHourAvg                  string
	AcTonsPerHourAvg                string
	AggTonsPerHourAvg               string
	TemperatureAvg                  string
	AcTemperatureAvg                string
}

type materialSiteMixingLotRow struct {
	ID                                string `json:"id"`
	StartAt                           string `json:"start_at,omitempty"`
	StartOn                           string `json:"start_on,omitempty"`
	EndAt                             string `json:"end_at,omitempty"`
	TimeZoneID                        string `json:"time_zone_id,omitempty"`
	TonsPerHourAvg                    string `json:"tons_per_hour_avg,omitempty"`
	AcTonsPerHourAvg                  string `json:"ac_tons_per_hour_avg,omitempty"`
	AggTonsPerHourAvg                 string `json:"agg_tons_per_hour_avg,omitempty"`
	TemperatureAvg                    string `json:"temperature_avg,omitempty"`
	AcTemperatureAvg                  string `json:"ac_temperature_avg,omitempty"`
	ReadingAtMax                      string `json:"reading_at_max,omitempty"`
	MaterialSiteID                    string `json:"material_site_id,omitempty"`
	MaterialSupplierID                string `json:"material_supplier_id,omitempty"`
	BrokerID                          string `json:"broker_id,omitempty"`
	MaterialSiteReadingMaterialTypeID string `json:"material_site_reading_material_type_id,omitempty"`
	MaterialTypeID                    string `json:"material_type_id,omitempty"`
}

func newMaterialSiteMixingLotsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site mixing lots",
		Long: `List material site mixing lots.

Output Columns:
  ID         Mixing lot identifier
  START AT   Start timestamp
  END AT     End timestamp
  SITE       Material site ID
  TYPE       Material type ID
  TPH AVG    Tons per hour average
  TEMP AVG   Temperature average

Filters:
  --material-site                       Filter by material site ID
  --material-supplier-id                Filter by material supplier ID (join)
  --material-supplier                   Filter by material supplier ID
  --material-site-reading-material-type Filter by material site reading material type ID
  --material-type-id                    Filter by material type ID (join)
  --material-type                       Filter by material type ID
  --broker                              Filter by broker ID
  --start-at                            Filter by start-at timestamp (ISO 8601)
  --end-at                              Filter by end-at timestamp (ISO 8601)
  --start-on-cached                     Filter by start-on date (YYYY-MM-DD)
  --tons-per-hour-avg                   Filter by tons per hour average
  --ac-tons-per-hour-avg                Filter by AC tons per hour average
  --agg-tons-per-hour-avg               Filter by aggregate tons per hour average
  --temperature-avg                     Filter by temperature average
  --ac-temperature-avg                  Filter by AC temperature average

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List mixing lots
  xbe view material-site-mixing-lots list

  # Filter by material site and start date
  xbe view material-site-mixing-lots list --material-site 123 --start-on-cached 2025-01-15

  # Output as JSON
  xbe view material-site-mixing-lots list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialSiteMixingLotsList,
	}
	initMaterialSiteMixingLotsListFlags(cmd)
	return cmd
}

func init() {
	materialSiteMixingLotsCmd.AddCommand(newMaterialSiteMixingLotsListCmd())
}

func initMaterialSiteMixingLotsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("material-supplier-id", "", "Filter by material supplier ID (join)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("material-site-reading-material-type", "", "Filter by material site reading material type ID")
	cmd.Flags().String("material-type-id", "", "Filter by material type ID (join)")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("start-at", "", "Filter by start-at timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "Filter by end-at timestamp (ISO 8601)")
	cmd.Flags().String("start-on-cached", "", "Filter by start-on date (YYYY-MM-DD)")
	cmd.Flags().String("tons-per-hour-avg", "", "Filter by tons per hour average")
	cmd.Flags().String("ac-tons-per-hour-avg", "", "Filter by AC tons per hour average")
	cmd.Flags().String("agg-tons-per-hour-avg", "", "Filter by aggregate tons per hour average")
	cmd.Flags().String("temperature-avg", "", "Filter by temperature average")
	cmd.Flags().String("ac-temperature-avg", "", "Filter by AC temperature average")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteMixingLotsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteMixingLotsListOptions(cmd)
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
	query.Set("fields[material-site-mixing-lots]", "start-at,start-on,end-at,time-zone-id,tons-per-hour-avg,ac-tons-per-hour-avg,agg-tons-per-hour-avg,temperature-avg,ac-temperature-avg,reading-at-max,material-site,material-supplier,material-site-reading-material-type,material-type,broker")

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
	setFilterIfPresent(query, "filter[material-supplier-id]", opts.MaterialSupplierID)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[material-site-reading-material-type]", opts.MaterialSiteReadingMaterialType)
	setFilterIfPresent(query, "filter[material-type-id]", opts.MaterialTypeID)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[start-at]", opts.StartAt)
	setFilterIfPresent(query, "filter[end-at]", opts.EndAt)
	setFilterIfPresent(query, "filter[start-on-cached]", opts.StartOnCached)
	setFilterIfPresent(query, "filter[tons-per-hour-avg]", opts.TonsPerHourAvg)
	setFilterIfPresent(query, "filter[ac-tons-per-hour-avg]", opts.AcTonsPerHourAvg)
	setFilterIfPresent(query, "filter[agg-tons-per-hour-avg]", opts.AggTonsPerHourAvg)
	setFilterIfPresent(query, "filter[temperature-avg]", opts.TemperatureAvg)
	setFilterIfPresent(query, "filter[ac-temperature-avg]", opts.AcTemperatureAvg)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-mixing-lots", query)
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

	rows := buildMaterialSiteMixingLotRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteMixingLotsTable(cmd, rows)
}

func parseMaterialSiteMixingLotsListOptions(cmd *cobra.Command) (materialSiteMixingLotsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier-id")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	materialSiteReadingMaterialType, _ := cmd.Flags().GetString("material-site-reading-material-type")
	materialTypeID, _ := cmd.Flags().GetString("material-type-id")
	materialType, _ := cmd.Flags().GetString("material-type")
	broker, _ := cmd.Flags().GetString("broker")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	startOnCached, _ := cmd.Flags().GetString("start-on-cached")
	tonsPerHourAvg, _ := cmd.Flags().GetString("tons-per-hour-avg")
	acTonsPerHourAvg, _ := cmd.Flags().GetString("ac-tons-per-hour-avg")
	aggTonsPerHourAvg, _ := cmd.Flags().GetString("agg-tons-per-hour-avg")
	temperatureAvg, _ := cmd.Flags().GetString("temperature-avg")
	acTemperatureAvg, _ := cmd.Flags().GetString("ac-temperature-avg")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteMixingLotsListOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		NoAuth:                          noAuth,
		Limit:                           limit,
		Offset:                          offset,
		Sort:                            sort,
		MaterialSite:                    materialSite,
		MaterialSupplierID:              materialSupplierID,
		MaterialSupplier:                materialSupplier,
		MaterialSiteReadingMaterialType: materialSiteReadingMaterialType,
		MaterialTypeID:                  materialTypeID,
		MaterialType:                    materialType,
		Broker:                          broker,
		StartAt:                         startAt,
		EndAt:                           endAt,
		StartOnCached:                   startOnCached,
		TonsPerHourAvg:                  tonsPerHourAvg,
		AcTonsPerHourAvg:                acTonsPerHourAvg,
		AggTonsPerHourAvg:               aggTonsPerHourAvg,
		TemperatureAvg:                  temperatureAvg,
		AcTemperatureAvg:                acTemperatureAvg,
	}, nil
}

func buildMaterialSiteMixingLotRows(resp jsonAPIResponse) []materialSiteMixingLotRow {
	rows := make([]materialSiteMixingLotRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialSiteMixingLotRow(resource))
	}
	return rows
}

func buildMaterialSiteMixingLotRow(resource jsonAPIResource) materialSiteMixingLotRow {
	attrs := resource.Attributes
	row := materialSiteMixingLotRow{
		ID:                resource.ID,
		StartAt:           formatDateTime(stringAttr(attrs, "start-at")),
		StartOn:           formatDate(stringAttr(attrs, "start-on")),
		EndAt:             formatDateTime(stringAttr(attrs, "end-at")),
		TimeZoneID:        stringAttr(attrs, "time-zone-id"),
		TonsPerHourAvg:    stringAttr(attrs, "tons-per-hour-avg"),
		AcTonsPerHourAvg:  stringAttr(attrs, "ac-tons-per-hour-avg"),
		AggTonsPerHourAvg: stringAttr(attrs, "agg-tons-per-hour-avg"),
		TemperatureAvg:    stringAttr(attrs, "temperature-avg"),
		AcTemperatureAvg:  stringAttr(attrs, "ac-temperature-avg"),
		ReadingAtMax:      formatDateTime(stringAttr(attrs, "reading-at-max")),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.MaterialSupplierID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-site-reading-material-type"]; ok && rel.Data != nil {
		row.MaterialSiteReadingMaterialTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
	}

	return row
}

func renderMaterialSiteMixingLotsTable(cmd *cobra.Command, rows []materialSiteMixingLotRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site mixing lots found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTART AT\tEND AT\tSITE\tTYPE\tTPH AVG\tTEMP AVG")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.StartAt,
			row.EndAt,
			row.MaterialSiteID,
			row.MaterialTypeID,
			truncateString(row.TonsPerHourAvg, 10),
			truncateString(row.TemperatureAvg, 10),
		)
	}
	return writer.Flush()
}
