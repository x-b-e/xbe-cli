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

type equipmentUtilizationReadingsListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	Equipment     string
	BusinessUnit  string
	User          string
	ReportedAtMin string
	ReportedAtMax string
	Source        string
}

type equipmentUtilizationReadingRow struct {
	ID             string `json:"id"`
	EquipmentID    string `json:"equipment_id,omitempty"`
	BusinessUnitID string `json:"business_unit_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	ReportedAt     string `json:"reported_at,omitempty"`
	Odometer       string `json:"odometer,omitempty"`
	Hourmeter      string `json:"hourmeter,omitempty"`
	Source         string `json:"source,omitempty"`
}

func newEquipmentUtilizationReadingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment utilization readings",
		Long: `List equipment utilization readings.

Output Columns:
  ID          Reading identifier
  EQUIPMENT   Equipment ID
  REPORTED AT Reported timestamp
  ODOMETER    Odometer reading
  HOURMETER   Hourmeter reading
  SOURCE      Reading source

Filters:
  --equipment        Filter by equipment ID
  --business-unit    Filter by business unit ID
  --user             Filter by user ID
  --reported-at-min  Filter by reported-at on/after (ISO 8601)
  --reported-at-max  Filter by reported-at on/before (ISO 8601)
  --source           Filter by reading source (manual includes empty source)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List readings
  xbe view equipment-utilization-readings list

  # Filter by equipment
  xbe view equipment-utilization-readings list --equipment 123

  # Filter by reported-at range
  xbe view equipment-utilization-readings list --reported-at-min 2025-01-01T00:00:00Z --reported-at-max 2025-01-31T23:59:59Z

  # Filter by source
  xbe view equipment-utilization-readings list --source manual

  # Output as JSON
  xbe view equipment-utilization-readings list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentUtilizationReadingsList,
	}
	initEquipmentUtilizationReadingsListFlags(cmd)
	return cmd
}

func init() {
	equipmentUtilizationReadingsCmd.AddCommand(newEquipmentUtilizationReadingsListCmd())
}

func initEquipmentUtilizationReadingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("reported-at-min", "", "Filter by reported-at on/after (ISO 8601)")
	cmd.Flags().String("reported-at-max", "", "Filter by reported-at on/before (ISO 8601)")
	cmd.Flags().String("source", "", "Filter by reading source (manual includes empty source)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentUtilizationReadingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentUtilizationReadingsListOptions(cmd)
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
	query.Set("fields[equipment-utilization-readings]", "odometer,hourmeter,reported-at,source,equipment,business-unit,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[reported-at-min]", opts.ReportedAtMin)
	setFilterIfPresent(query, "filter[reported-at-max]", opts.ReportedAtMax)
	setFilterIfPresent(query, "filter[source]", opts.Source)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-utilization-readings", query)
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

	rows := buildEquipmentUtilizationReadingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentUtilizationReadingsTable(cmd, rows)
}

func parseEquipmentUtilizationReadingsListOptions(cmd *cobra.Command) (equipmentUtilizationReadingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	equipment, _ := cmd.Flags().GetString("equipment")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	user, _ := cmd.Flags().GetString("user")
	reportedAtMin, _ := cmd.Flags().GetString("reported-at-min")
	reportedAtMax, _ := cmd.Flags().GetString("reported-at-max")
	source, _ := cmd.Flags().GetString("source")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentUtilizationReadingsListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		Equipment:     equipment,
		BusinessUnit:  businessUnit,
		User:          user,
		ReportedAtMin: reportedAtMin,
		ReportedAtMax: reportedAtMax,
		Source:        source,
	}, nil
}

func buildEquipmentUtilizationReadingRows(resp jsonAPIResponse) []equipmentUtilizationReadingRow {
	rows := make([]equipmentUtilizationReadingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildEquipmentUtilizationReadingRow(resource))
	}
	return rows
}

func buildEquipmentUtilizationReadingRowFromSingle(resp jsonAPISingleResponse) equipmentUtilizationReadingRow {
	return buildEquipmentUtilizationReadingRow(resp.Data)
}

func buildEquipmentUtilizationReadingRow(resource jsonAPIResource) equipmentUtilizationReadingRow {
	attrs := resource.Attributes
	row := equipmentUtilizationReadingRow{
		ID:         resource.ID,
		ReportedAt: formatDateTime(stringAttr(attrs, "reported-at")),
		Odometer:   stringAttr(attrs, "odometer"),
		Hourmeter:  stringAttr(attrs, "hourmeter"),
		Source:     stringAttr(attrs, "source"),
	}

	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
		row.BusinessUnitID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}

func renderEquipmentUtilizationReadingsTable(cmd *cobra.Command, rows []equipmentUtilizationReadingRow) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEQUIPMENT\tREPORTED AT\tODOMETER\tHOURMETER\tSOURCE")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EquipmentID,
			row.ReportedAt,
			row.Odometer,
			row.Hourmeter,
			row.Source,
		)
	}
	return w.Flush()
}
