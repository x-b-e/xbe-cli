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

type timeSheetLineItemsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TimeSheet              string
	CostCode               string
	CraftClass             string
	TimeCard               string
	MaintenanceRequirement string
	Broker                 string
	Trucker                string
	Customer               string
	CraftClassEffective    string
	CraftClassEffectiveID  string
	StartAtMin             string
	StartAtMax             string
	EndAtMin               string
	EndAtMax               string
	IsStartAt              string
	IsEndAt                string
}

type timeSheetLineItemRow struct {
	ID                                string `json:"id"`
	TimeSheetID                       string `json:"time_sheet_id,omitempty"`
	StartAt                           string `json:"start_at,omitempty"`
	EndAt                             string `json:"end_at,omitempty"`
	BreakMinutes                      int    `json:"break_minutes,omitempty"`
	DurationSeconds                   int    `json:"duration_seconds,omitempty"`
	IsNonJobLineItem                  bool   `json:"is_non_job_line_item"`
	CostCodeID                        string `json:"cost_code_id,omitempty"`
	CraftClassID                      string `json:"craft_class_id,omitempty"`
	TimeSheetLineItemClassificationID string `json:"time_sheet_line_item_classification_id,omitempty"`
	ProjectCostClassificationID       string `json:"project_cost_classification_id,omitempty"`
	EquipmentRequirementID            string `json:"equipment_requirement_id,omitempty"`
	MaintenanceRequirementID          string `json:"maintenance_requirement_id,omitempty"`
	TimeCardID                        string `json:"time_card_id,omitempty"`
	ExplicitJobProductionPlanID       string `json:"explicit_job_production_plan_id,omitempty"`
	ImplicitJobProductionPlanID       string `json:"implicit_job_production_plan_id,omitempty"`
	JobProductionPlanID               string `json:"job_production_plan_id,omitempty"`
	CraftClassEffectiveID             string `json:"craft_class_effective_id,omitempty"`
}

func newTimeSheetLineItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet line items",
		Long: `List time sheet line items with filtering and pagination.

Output Columns:
  ID            Time sheet line item identifier
  TIME_SHEET    Time sheet ID
  START_AT      Start timestamp
  END_AT        End timestamp
  BREAK_MIN     Break minutes
  DURATION_SEC  Duration in seconds
  NON_JOB       Non-job line item indicator
  COST_CODE     Cost code ID
  CRAFT_CLASS   Craft class ID

Filters:
  --time-sheet                Filter by time sheet ID (comma-separated for multiple)
  --cost-code                 Filter by cost code ID (comma-separated for multiple)
  --craft-class               Filter by craft class ID (comma-separated for multiple)
  --time-card                 Filter by time card ID (comma-separated for multiple)
  --maintenance-requirement   Filter by maintenance requirement ID (comma-separated for multiple)
  --broker                    Filter by broker ID (comma-separated for multiple)
  --trucker                   Filter by trucker ID (comma-separated for multiple)
  --customer                  Filter by customer ID (comma-separated for multiple)
  --craft-class-effective     Filter by craft class effective ID (comma-separated for multiple)
  --craft-class-effective-id  Filter by craft class effective ID (legacy filter)
  --start-at-min              Filter by minimum start timestamp (ISO 8601)
  --start-at-max              Filter by maximum start timestamp (ISO 8601)
  --end-at-min                Filter by minimum end timestamp (ISO 8601)
  --end-at-max                Filter by maximum end timestamp (ISO 8601)
  --is-start-at               Filter by presence of start timestamp (true/false)
  --is-end-at                 Filter by presence of end timestamp (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheet line items
  xbe view time-sheet-line-items list

  # Filter by time sheet
  xbe view time-sheet-line-items list --time-sheet 123

  # Filter by time range
  xbe view time-sheet-line-items list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view time-sheet-line-items list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetLineItemsList,
	}
	initTimeSheetLineItemsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetLineItemsCmd.AddCommand(newTimeSheetLineItemsListCmd())
}

func initTimeSheetLineItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-sheet", "", "Filter by time sheet ID (comma-separated for multiple)")
	cmd.Flags().String("cost-code", "", "Filter by cost code ID (comma-separated for multiple)")
	cmd.Flags().String("craft-class", "", "Filter by craft class ID (comma-separated for multiple)")
	cmd.Flags().String("time-card", "", "Filter by time card ID (comma-separated for multiple)")
	cmd.Flags().String("maintenance-requirement", "", "Filter by maintenance requirement ID (comma-separated for multiple)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("craft-class-effective", "", "Filter by craft class effective ID (comma-separated for multiple)")
	cmd.Flags().String("craft-class-effective-id", "", "Filter by craft class effective ID (legacy filter)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start timestamp (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start timestamp (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end timestamp (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end timestamp (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start timestamp (true/false)")
	cmd.Flags().String("is-end-at", "", "Filter by presence of end timestamp (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetLineItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetLineItemsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-line-items]", "start-at,end-at,break-minutes,description,skip-validate-overlap,is-non-job-line-item,duration-seconds,time-sheet,cost-code,craft-class,time-card,maintenance-requirement,time-sheet-line-item-classification,project-cost-classification,explicit-job-production-plan,implicit-job-production-plan,job-production-plan,equipment-requirement,craft-class-effective")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[time-sheet]", opts.TimeSheet)
	setFilterIfPresent(query, "filter[cost-code]", opts.CostCode)
	setFilterIfPresent(query, "filter[craft-class]", opts.CraftClass)
	setFilterIfPresent(query, "filter[time-card]", opts.TimeCard)
	setFilterIfPresent(query, "filter[maintenance-requirement]", opts.MaintenanceRequirement)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[craft-class-effective]", opts.CraftClassEffective)
	setFilterIfPresent(query, "filter[craft-class-effective-id]", opts.CraftClassEffectiveID)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end-at-min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[is-start-at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[is-end-at]", opts.IsEndAt)

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-line-items", query)
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

	rows := buildTimeSheetLineItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetLineItemsTable(cmd, rows)
}

func parseTimeSheetLineItemsListOptions(cmd *cobra.Command) (timeSheetLineItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	costCode, _ := cmd.Flags().GetString("cost-code")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	timeCard, _ := cmd.Flags().GetString("time-card")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	customer, _ := cmd.Flags().GetString("customer")
	craftClassEffective, _ := cmd.Flags().GetString("craft-class-effective")
	craftClassEffectiveID, _ := cmd.Flags().GetString("craft-class-effective-id")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	isStartAt, _ := cmd.Flags().GetString("is-start-at")
	isEndAt, _ := cmd.Flags().GetString("is-end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetLineItemsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TimeSheet:              timeSheet,
		CostCode:               costCode,
		CraftClass:             craftClass,
		TimeCard:               timeCard,
		MaintenanceRequirement: maintenanceRequirement,
		Broker:                 broker,
		Trucker:                trucker,
		Customer:               customer,
		CraftClassEffective:    craftClassEffective,
		CraftClassEffectiveID:  craftClassEffectiveID,
		StartAtMin:             startAtMin,
		StartAtMax:             startAtMax,
		EndAtMin:               endAtMin,
		EndAtMax:               endAtMax,
		IsStartAt:              isStartAt,
		IsEndAt:                isEndAt,
	}, nil
}

func buildTimeSheetLineItemRows(resp jsonAPIResponse) []timeSheetLineItemRow {
	rows := make([]timeSheetLineItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeSheetLineItemRow{
			ID:                                resource.ID,
			TimeSheetID:                       relationshipIDFromMap(resource.Relationships, "time-sheet"),
			StartAt:                           formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:                             formatDateTime(stringAttr(attrs, "end-at")),
			BreakMinutes:                      intAttr(attrs, "break-minutes"),
			DurationSeconds:                   intAttr(attrs, "duration-seconds"),
			IsNonJobLineItem:                  boolAttr(attrs, "is-non-job-line-item"),
			CostCodeID:                        relationshipIDFromMap(resource.Relationships, "cost-code"),
			CraftClassID:                      relationshipIDFromMap(resource.Relationships, "craft-class"),
			TimeSheetLineItemClassificationID: relationshipIDFromMap(resource.Relationships, "time-sheet-line-item-classification"),
			ProjectCostClassificationID:       relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
			EquipmentRequirementID:            relationshipIDFromMap(resource.Relationships, "equipment-requirement"),
			MaintenanceRequirementID:          relationshipIDFromMap(resource.Relationships, "maintenance-requirement"),
			TimeCardID:                        relationshipIDFromMap(resource.Relationships, "time-card"),
			ExplicitJobProductionPlanID:       relationshipIDFromMap(resource.Relationships, "explicit-job-production-plan"),
			ImplicitJobProductionPlanID:       relationshipIDFromMap(resource.Relationships, "implicit-job-production-plan"),
			JobProductionPlanID:               relationshipIDFromMap(resource.Relationships, "job-production-plan"),
			CraftClassEffectiveID:             relationshipIDFromMap(resource.Relationships, "craft-class-effective"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTimeSheetLineItemsTable(cmd *cobra.Command, rows []timeSheetLineItemRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No time sheet line items found.")
		return nil
	}

	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME_SHEET\tSTART_AT\tEND_AT\tBREAK_MIN\tDURATION_SEC\tNON_JOB\tCOST_CODE\tCRAFT_CLASS")
	for _, row := range rows {
		nonJob := ""
		if row.IsNonJobLineItem {
			nonJob = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\n",
			row.ID,
			row.TimeSheetID,
			row.StartAt,
			row.EndAt,
			row.BreakMinutes,
			row.DurationSeconds,
			nonJob,
			row.CostCodeID,
			row.CraftClassID,
		)
	}

	return writer.Flush()
}
