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

type equipmentMovementTripJobProductionPlansListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	EquipmentMovementTrip string
	JobProductionPlan     string
	CreatedAtMin          string
	CreatedAtMax          string
	UpdatedAtMin          string
	UpdatedAtMax          string
}

type equipmentMovementTripJobProductionPlanRow struct {
	ID                    string `json:"id"`
	EquipmentMovementTrip string `json:"equipment_movement_trip_id,omitempty"`
	JobProductionPlan     string `json:"job_production_plan_id,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newEquipmentMovementTripJobProductionPlansListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement trip job production plans",
		Long: `List equipment movement trip job production plans.

Output Columns:
  ID         Link identifier
  TRIP       Equipment movement trip ID
  JOB PLAN   Job production plan ID
  CREATED AT Creation timestamp
  UPDATED AT Last update timestamp

Filters:
  --equipment-movement-trip  Filter by equipment movement trip ID
  --job-production-plan      Filter by job production plan ID
  --created-at-min           Filter by created-at on/after (ISO 8601)
  --created-at-max           Filter by created-at on/before (ISO 8601)
  --updated-at-min           Filter by updated-at on/after (ISO 8601)
  --updated-at-max           Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List links
  xbe view equipment-movement-trip-job-production-plans list

  # Filter by equipment movement trip
  xbe view equipment-movement-trip-job-production-plans list --equipment-movement-trip 123

  # Filter by job production plan
  xbe view equipment-movement-trip-job-production-plans list --job-production-plan 456

  # Output as JSON
  xbe view equipment-movement-trip-job-production-plans list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentMovementTripJobProductionPlansList,
	}
	initEquipmentMovementTripJobProductionPlansListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripJobProductionPlansCmd.AddCommand(newEquipmentMovementTripJobProductionPlansListCmd())
}

func initEquipmentMovementTripJobProductionPlansListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("equipment-movement-trip", "", "Filter by equipment movement trip ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripJobProductionPlansList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementTripJobProductionPlansListOptions(cmd)
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
	query.Set("fields[equipment-movement-trip-job-production-plans]", "created-at,updated-at,equipment-movement-trip,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[equipment_movement_trip]", opts.EquipmentMovementTrip)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trip-job-production-plans", query)
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

	rows := buildEquipmentMovementTripJobProductionPlanRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementTripJobProductionPlansTable(cmd, rows)
}

func parseEquipmentMovementTripJobProductionPlansListOptions(cmd *cobra.Command) (equipmentMovementTripJobProductionPlansListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	equipmentMovementTrip, _ := cmd.Flags().GetString("equipment-movement-trip")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripJobProductionPlansListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		EquipmentMovementTrip: equipmentMovementTrip,
		JobProductionPlan:     jobProductionPlan,
		CreatedAtMin:          createdAtMin,
		CreatedAtMax:          createdAtMax,
		UpdatedAtMin:          updatedAtMin,
		UpdatedAtMax:          updatedAtMax,
	}, nil
}

func buildEquipmentMovementTripJobProductionPlanRows(resp jsonAPIResponse) []equipmentMovementTripJobProductionPlanRow {
	rows := make([]equipmentMovementTripJobProductionPlanRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := equipmentMovementTripJobProductionPlanRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
		}
		if rel, ok := resource.Relationships["equipment-movement-trip"]; ok && rel.Data != nil {
			row.EquipmentMovementTrip = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlan = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentMovementTripJobProductionPlansTable(cmd *cobra.Command, rows []equipmentMovementTripJobProductionPlanRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement trip job production plans found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRIP\tJOB PLAN\tCREATED AT\tUPDATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EquipmentMovementTrip,
			row.JobProductionPlan,
			row.CreatedAt,
			row.UpdatedAt,
		)
	}
	return writer.Flush()
}
