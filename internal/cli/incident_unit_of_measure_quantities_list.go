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

type incidentUnitOfMeasureQuantitiesListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	IncidentType  string
	IncidentID    string
	UnitOfMeasure string
}

type incidentUnitOfMeasureQuantityRow struct {
	ID                 string `json:"id"`
	Quantity           string `json:"quantity,omitempty"`
	IsSetAutomatically bool   `json:"is_set_automatically"`
	UnitOfMeasureID    string `json:"unit_of_measure_id,omitempty"`
	IncidentType       string `json:"incident_type,omitempty"`
	IncidentID         string `json:"incident_id,omitempty"`
}

func newIncidentUnitOfMeasureQuantitiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident unit of measure quantities",
		Long: `List incident unit of measure quantities with filtering and pagination.

Output Columns:
  ID               Quantity identifier
  QUANTITY         Quantity value
  AUTO             Whether the quantity was set automatically
  UNIT OF MEASURE  Unit of measure ID
  INCIDENT         Incident type and ID

Filters:
  --incident-type      Filter by incident type (use with --incident-id)
  --incident-id        Filter by incident ID (requires --incident-type)
  --unit-of-measure    Filter by unit of measure ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident unit of measure quantities
  xbe view incident-unit-of-measure-quantities list

  # Filter by incident
  xbe view incident-unit-of-measure-quantities list --incident-type incidents --incident-id 123

  # Filter by unit of measure
  xbe view incident-unit-of-measure-quantities list --unit-of-measure 456

  # Output as JSON
  xbe view incident-unit-of-measure-quantities list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentUnitOfMeasureQuantitiesList,
	}
	initIncidentUnitOfMeasureQuantitiesListFlags(cmd)
	return cmd
}

func init() {
	incidentUnitOfMeasureQuantitiesCmd.AddCommand(newIncidentUnitOfMeasureQuantitiesListCmd())
}

func initIncidentUnitOfMeasureQuantitiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("incident-type", "", "Filter by incident type")
	cmd.Flags().String("incident-id", "", "Filter by incident ID (requires --incident-type)")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentUnitOfMeasureQuantitiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentUnitOfMeasureQuantitiesListOptions(cmd)
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
	query.Set("fields[incident-unit-of-measure-quantities]", "quantity,is-set-automatically,unit-of-measure,incident")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	normalizedIncidentType := normalizeIncidentTypeFilter(opts.IncidentType)
	if opts.IncidentType != "" && opts.IncidentID != "" {
		query.Set("filter[incident]", normalizedIncidentType+"|"+opts.IncidentID)
	} else if opts.IncidentType != "" {
		query.Set("filter[incident_type]", normalizedIncidentType)
	} else if opts.IncidentID != "" {
		err := fmt.Errorf("--incident-id requires --incident-type")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-unit-of-measure-quantities", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildIncidentUnitOfMeasureQuantityRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentUnitOfMeasureQuantitiesTable(cmd, rows)
}

func parseIncidentUnitOfMeasureQuantitiesListOptions(cmd *cobra.Command) (incidentUnitOfMeasureQuantitiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentUnitOfMeasureQuantitiesListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		IncidentType:  incidentType,
		IncidentID:    incidentID,
		UnitOfMeasure: unitOfMeasure,
	}, nil
}

func buildIncidentUnitOfMeasureQuantityRows(resp jsonAPIResponse) []incidentUnitOfMeasureQuantityRow {
	rows := make([]incidentUnitOfMeasureQuantityRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildIncidentUnitOfMeasureQuantityRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildIncidentUnitOfMeasureQuantityRow(resource jsonAPIResource) incidentUnitOfMeasureQuantityRow {
	row := incidentUnitOfMeasureQuantityRow{
		ID:                 resource.ID,
		Quantity:           stringAttr(resource.Attributes, "quantity"),
		IsSetAutomatically: boolAttr(resource.Attributes, "is-set-automatically"),
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["incident"]; ok && rel.Data != nil {
		row.IncidentType = rel.Data.Type
		row.IncidentID = rel.Data.ID
	}

	return row
}

func buildIncidentUnitOfMeasureQuantityRowFromSingle(resp jsonAPISingleResponse) incidentUnitOfMeasureQuantityRow {
	return buildIncidentUnitOfMeasureQuantityRow(resp.Data)
}

func renderIncidentUnitOfMeasureQuantitiesTable(cmd *cobra.Command, rows []incidentUnitOfMeasureQuantityRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident unit of measure quantities found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tQUANTITY\tAUTO\tUNIT OF MEASURE\tINCIDENT")
	for _, row := range rows {
		auto := "no"
		if row.IsSetAutomatically {
			auto = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Quantity,
			auto,
			row.UnitOfMeasureID,
			formatIncidentReference(row.IncidentType, row.IncidentID),
		)
	}
	return writer.Flush()
}

func formatIncidentReference(incidentType, incidentID string) string {
	if incidentType == "" && incidentID == "" {
		return ""
	}
	if incidentType == "" {
		return incidentID
	}
	if incidentID == "" {
		return incidentType
	}
	return incidentType + "/" + incidentID
}

func normalizeIncidentTypeFilter(incidentType string) string {
	normalized := strings.TrimSpace(incidentType)
	if normalized == "" {
		return normalized
	}
	switch strings.ToLower(normalized) {
	case "incident", "incidents":
		return "Incident"
	case "safety-incident", "safety-incidents", "safetyincident":
		return "SafetyIncident"
	case "production-incident", "production-incidents", "productionincident":
		return "ProductionIncident"
	case "efficiency-incident", "efficiency-incidents", "efficiencyincident":
		return "EfficiencyIncident"
	case "administrative-incident", "administrative-incidents", "administrativeincident":
		return "AdministrativeIncident"
	case "liability-incident", "liability-incidents", "liabilityincident":
		return "LiabilityIncident"
	default:
		return normalized
	}
}
