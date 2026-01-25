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

type materialSiteMixingLotsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteMixingLotDetails struct {
	ID                                string   `json:"id"`
	MaterialSiteID                    string   `json:"material_site_id,omitempty"`
	MaterialSupplierID                string   `json:"material_supplier_id,omitempty"`
	BrokerID                          string   `json:"broker_id,omitempty"`
	MaterialSiteReadingMaterialTypeID string   `json:"material_site_reading_material_type_id,omitempty"`
	MaterialTypeID                    string   `json:"material_type_id,omitempty"`
	StartAt                           string   `json:"start_at,omitempty"`
	StartOn                           string   `json:"start_on,omitempty"`
	EndAt                             string   `json:"end_at,omitempty"`
	TimeZoneID                        string   `json:"time_zone_id,omitempty"`
	TonsPerHourAvg                    string   `json:"tons_per_hour_avg,omitempty"`
	AcTonsPerHourAvg                  string   `json:"ac_tons_per_hour_avg,omitempty"`
	AggTonsPerHourAvg                 string   `json:"agg_tons_per_hour_avg,omitempty"`
	TemperatureAvg                    string   `json:"temperature_avg,omitempty"`
	AcTemperatureAvg                  string   `json:"ac_temperature_avg,omitempty"`
	ReadingAtMax                      string   `json:"reading_at_max,omitempty"`
	CommentIDs                        []string `json:"comment_ids,omitempty"`
}

func newMaterialSiteMixingLotsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site mixing lot details",
		Long: `Show the full details of a material site mixing lot.

Output Fields:
  ID                 Mixing lot identifier
  MATERIAL SITE ID   Material site ID
  MATERIAL SUPPLIER ID Material supplier ID
  BROKER ID          Broker ID
  READING MATERIAL TYPE ID  Material site reading material type ID
  MATERIAL TYPE ID   Material type ID
  START AT           Start timestamp
  START ON           Start date (site-local)
  END AT             End timestamp
  TIME ZONE ID       Time zone ID
  TPH AVG            Tons per hour average
  AC TPH AVG         AC tons per hour average
  AGG TPH AVG        Aggregate tons per hour average
  TEMP AVG           Temperature average
  AC TEMP AVG        AC temperature average
  READING AT MAX     Last reading timestamp for the lot
  COMMENT IDS        Comment IDs

Arguments:
  <id>  Material site mixing lot ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a mixing lot
  xbe view material-site-mixing-lots show 123

  # JSON output
  xbe view material-site-mixing-lots show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteMixingLotsShow,
	}
	initMaterialSiteMixingLotsShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteMixingLotsCmd.AddCommand(newMaterialSiteMixingLotsShowCmd())
}

func initMaterialSiteMixingLotsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteMixingLotsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialSiteMixingLotsShowOptions(cmd)
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
		return fmt.Errorf("material site mixing lot id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-site-mixing-lots]", "start-at,start-on,end-at,time-zone-id,tons-per-hour-avg,ac-tons-per-hour-avg,agg-tons-per-hour-avg,temperature-avg,ac-temperature-avg,reading-at-max,material-site,material-supplier,material-site-reading-material-type,material-type,broker,comments")

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-mixing-lots/"+id, query)
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

	details := buildMaterialSiteMixingLotDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteMixingLotDetails(cmd, details)
}

func parseMaterialSiteMixingLotsShowOptions(cmd *cobra.Command) (materialSiteMixingLotsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteMixingLotsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteMixingLotDetails(resp jsonAPISingleResponse) materialSiteMixingLotDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := materialSiteMixingLotDetails{
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
		details.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-site-reading-material-type"]; ok && rel.Data != nil {
		details.MaterialSiteReadingMaterialTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["comments"]; ok {
		details.CommentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderMaterialSiteMixingLotDetails(cmd *cobra.Command, details materialSiteMixingLotDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site ID: %s\n", details.MaterialSiteID)
	}
	if details.MaterialSupplierID != "" {
		fmt.Fprintf(out, "Material Supplier ID: %s\n", details.MaterialSupplierID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.MaterialSiteReadingMaterialTypeID != "" {
		fmt.Fprintf(out, "Material Site Reading Material Type ID: %s\n", details.MaterialSiteReadingMaterialTypeID)
	}
	if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type ID: %s\n", details.MaterialTypeID)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if details.TonsPerHourAvg != "" {
		fmt.Fprintf(out, "Tons/Hour Avg: %s\n", details.TonsPerHourAvg)
	}
	if details.AcTonsPerHourAvg != "" {
		fmt.Fprintf(out, "AC Tons/Hour Avg: %s\n", details.AcTonsPerHourAvg)
	}
	if details.AggTonsPerHourAvg != "" {
		fmt.Fprintf(out, "Agg Tons/Hour Avg: %s\n", details.AggTonsPerHourAvg)
	}
	if details.TemperatureAvg != "" {
		fmt.Fprintf(out, "Temperature Avg: %s\n", details.TemperatureAvg)
	}
	if details.AcTemperatureAvg != "" {
		fmt.Fprintf(out, "AC Temperature Avg: %s\n", details.AcTemperatureAvg)
	}
	if details.ReadingAtMax != "" {
		fmt.Fprintf(out, "Reading At Max: %s\n", details.ReadingAtMax)
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}

	return nil
}
