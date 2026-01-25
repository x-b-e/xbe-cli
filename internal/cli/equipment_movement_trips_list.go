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

type equipmentMovementTripsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Broker              string
	MaxOriginAtMinMin   string
	MaxOriginAtMinMax   string
	JobProductionPlan   string
	JobProductionPlanID string
}

type equipmentMovementTripRow struct {
	ID                  string `json:"id"`
	JobNumber           string `json:"job_number,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	TrailerClassID      string `json:"trailer_classification_id,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
}

func newEquipmentMovementTripsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement trips",
		Long: `List equipment movement trips.

Output Columns:
  ID            Trip identifier
  JOB NUMBER    Job number
  BROKER        Broker ID
  TRAILER CLASS Trailer classification ID
  JPP           Job production plan ID

Filters:
  --broker                 Filter by broker ID
  --max-origin-at-min-min  Filter by max origin-at-min on/after (ISO 8601)
  --max-origin-at-min-max  Filter by max origin-at-min on/before (ISO 8601)
  --job-production-plan    Filter by job production plan ID
  --job-production-plan-id Filter by job production plan ID (alias filter)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List equipment movement trips
  xbe view equipment-movement-trips list

  # Filter by broker
  xbe view equipment-movement-trips list --broker 123

  # Filter by max origin-at-min window
  xbe view equipment-movement-trips list --max-origin-at-min-min 2025-01-01T00:00:00Z

  # Filter by job production plan
  xbe view equipment-movement-trips list --job-production-plan 456

  # Output as JSON
  xbe view equipment-movement-trips list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentMovementTripsList,
	}
	initEquipmentMovementTripsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripsCmd.AddCommand(newEquipmentMovementTripsListCmd())
}

func initEquipmentMovementTripsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("max-origin-at-min-min", "", "Filter by max origin-at-min on/after (ISO 8601)")
	cmd.Flags().String("max-origin-at-min-max", "", "Filter by max origin-at-min on/before (ISO 8601)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("job-production-plan-id", "", "Filter by job production plan ID (alias filter)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementTripsListOptions(cmd)
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
	query.Set("include", "broker,trailer-classification,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[max_origin_at_min_min]", opts.MaxOriginAtMinMin)
	setFilterIfPresent(query, "filter[max_origin_at_min_max]", opts.MaxOriginAtMinMax)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[job_production_plan_id]", opts.JobProductionPlanID)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trips", query)
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

	rows := buildEquipmentMovementTripRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementTripsTable(cmd, rows)
}

func parseEquipmentMovementTripsListOptions(cmd *cobra.Command) (equipmentMovementTripsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	maxOriginAtMinMin, _ := cmd.Flags().GetString("max-origin-at-min-min")
	maxOriginAtMinMax, _ := cmd.Flags().GetString("max-origin-at-min-max")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Broker:              broker,
		MaxOriginAtMinMin:   maxOriginAtMinMin,
		MaxOriginAtMinMax:   maxOriginAtMinMax,
		JobProductionPlan:   jobProductionPlan,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}

func buildEquipmentMovementTripRows(resp jsonAPIResponse) []equipmentMovementTripRow {
	rows := make([]equipmentMovementTripRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildEquipmentMovementTripRow(resource))
	}
	return rows
}

func buildEquipmentMovementTripRow(resource jsonAPIResource) equipmentMovementTripRow {
	row := equipmentMovementTripRow{
		ID:        resource.ID,
		JobNumber: stringAttr(resource.Attributes, "job-number"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer-classification"]; ok && rel.Data != nil {
		row.TrailerClassID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}

func buildEquipmentMovementTripRowFromSingle(resp jsonAPISingleResponse) equipmentMovementTripRow {
	return buildEquipmentMovementTripRow(resp.Data)
}

func renderEquipmentMovementTripsTable(cmd *cobra.Command, rows []equipmentMovementTripRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tJOB NUMBER\tBROKER\tTRAILER CLASS\tJPP")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobNumber,
			row.BrokerID,
			row.TrailerClassID,
			row.JobProductionPlanID,
		)
	}

	return w.Flush()
}
