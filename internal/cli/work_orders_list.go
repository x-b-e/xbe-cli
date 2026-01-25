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

type workOrdersListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Broker                string
	ResponsibleParty      string
	ServiceSite           string
	CustomWorkOrderStatus string
	ServiceCode           string
	Priority              string
	Status                string
	SafetyTagStatus       string
	DueDateMin            string
	DueDateMax            string
}

type workOrderRow struct {
	ID                      string  `json:"id"`
	Priority                string  `json:"priority,omitempty"`
	Status                  string  `json:"status,omitempty"`
	ActualStatus            string  `json:"actual_status,omitempty"`
	EstimatedLaborHours     float64 `json:"estimated_labor_hours,omitempty"`
	EstimatedPartCost       float64 `json:"estimated_part_cost,omitempty"`
	DueDate                 string  `json:"due_date,omitempty"`
	SafetyTagStatus         string  `json:"safety_tag_status,omitempty"`
	Note                    string  `json:"note,omitempty"`
	BrokerID                string  `json:"broker_id,omitempty"`
	ResponsiblePartyID      string  `json:"responsible_party_id,omitempty"`
	ServiceSiteID           string  `json:"service_site_id,omitempty"`
	CustomWorkOrderStatusID string  `json:"custom_work_order_status_id,omitempty"`
	ServiceCodeID           string  `json:"service_code_id,omitempty"`
}

func newWorkOrdersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work orders",
		Long: `List work orders.

Output Columns:
  ID          Work order identifier
  PRIORITY    Priority level
  STATUS      Current status
  DUE DATE    Due date
  BROKER      Broker ID
  RESP PARTY  Responsible party (business unit) ID

Filters:
  --broker                    Filter by broker ID
  --responsible-party         Filter by responsible party (business unit) ID
  --service-site              Filter by service site ID
  --custom-work-order-status  Filter by custom work order status ID
  --service-code              Filter by service code ID
  --priority                  Filter by priority
  --status                    Filter by status
  --safety-tag-status         Filter by safety tag status
  --due-date-min              Filter by minimum due date
  --due-date-max              Filter by maximum due date`,
		Example: `  # List all work orders
  xbe view work-orders list

  # Filter by broker
  xbe view work-orders list --broker 123

  # Filter by status
  xbe view work-orders list --status open

  # Filter by priority
  xbe view work-orders list --priority high

  # Filter by due date range
  xbe view work-orders list --due-date-min 2024-01-01 --due-date-max 2024-12-31

  # Output as JSON
  xbe view work-orders list --json`,
		RunE: runWorkOrdersList,
	}
	initWorkOrdersListFlags(cmd)
	return cmd
}

func init() {
	workOrdersCmd.AddCommand(newWorkOrdersListCmd())
}

func initWorkOrdersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("responsible-party", "", "Filter by responsible party (business unit) ID")
	cmd.Flags().String("service-site", "", "Filter by service site ID")
	cmd.Flags().String("custom-work-order-status", "", "Filter by custom work order status ID")
	cmd.Flags().String("service-code", "", "Filter by service code ID")
	cmd.Flags().String("priority", "", "Filter by priority")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("safety-tag-status", "", "Filter by safety tag status")
	cmd.Flags().String("due-date-min", "", "Filter by minimum due date")
	cmd.Flags().String("due-date-max", "", "Filter by maximum due date")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runWorkOrdersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseWorkOrdersListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[responsible_party]", opts.ResponsibleParty)
	setFilterIfPresent(query, "filter[service_site]", opts.ServiceSite)
	setFilterIfPresent(query, "filter[custom_work_order_status]", opts.CustomWorkOrderStatus)
	setFilterIfPresent(query, "filter[service_code]", opts.ServiceCode)
	setFilterIfPresent(query, "filter[priority]", opts.Priority)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[safety_tag_status]", opts.SafetyTagStatus)
	setFilterIfPresent(query, "filter[due_date_min]", opts.DueDateMin)
	setFilterIfPresent(query, "filter[due_date_max]", opts.DueDateMax)

	body, _, err := client.Get(cmd.Context(), "/v1/work-orders", query)
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

	rows := buildWorkOrderRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderWorkOrdersTable(cmd, rows)
}

func parseWorkOrdersListOptions(cmd *cobra.Command) (workOrdersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	responsibleParty, _ := cmd.Flags().GetString("responsible-party")
	serviceSite, _ := cmd.Flags().GetString("service-site")
	customWorkOrderStatus, _ := cmd.Flags().GetString("custom-work-order-status")
	serviceCode, _ := cmd.Flags().GetString("service-code")
	priority, _ := cmd.Flags().GetString("priority")
	status, _ := cmd.Flags().GetString("status")
	safetyTagStatus, _ := cmd.Flags().GetString("safety-tag-status")
	dueDateMin, _ := cmd.Flags().GetString("due-date-min")
	dueDateMax, _ := cmd.Flags().GetString("due-date-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return workOrdersListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Broker:                broker,
		ResponsibleParty:      responsibleParty,
		ServiceSite:           serviceSite,
		CustomWorkOrderStatus: customWorkOrderStatus,
		ServiceCode:           serviceCode,
		Priority:              priority,
		Status:                status,
		SafetyTagStatus:       safetyTagStatus,
		DueDateMin:            dueDateMin,
		DueDateMax:            dueDateMax,
	}, nil
}

func buildWorkOrderRows(resp jsonAPIResponse) []workOrderRow {
	rows := make([]workOrderRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := workOrderRow{
			ID:                  resource.ID,
			Priority:            stringAttr(resource.Attributes, "priority"),
			Status:              stringAttr(resource.Attributes, "status"),
			ActualStatus:        stringAttr(resource.Attributes, "actual-status"),
			EstimatedLaborHours: floatAttr(resource.Attributes, "estimated-labor-hours"),
			EstimatedPartCost:   floatAttr(resource.Attributes, "estimated-part-cost"),
			DueDate:             stringAttr(resource.Attributes, "due-date"),
			SafetyTagStatus:     stringAttr(resource.Attributes, "safety-tag-status"),
			Note:                stringAttr(resource.Attributes, "note"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["responsible-party"]; ok && rel.Data != nil {
			row.ResponsiblePartyID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["service-site"]; ok && rel.Data != nil {
			row.ServiceSiteID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["custom-work-order-status"]; ok && rel.Data != nil {
			row.CustomWorkOrderStatusID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["service-code"]; ok && rel.Data != nil {
			row.ServiceCodeID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderWorkOrdersTable(cmd *cobra.Command, rows []workOrderRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No work orders found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPRIORITY\tSTATUS\tDUE DATE\tBROKER\tRESP PARTY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Priority,
			row.Status,
			row.DueDate,
			row.BrokerID,
			row.ResponsiblePartyID,
		)
	}
	return writer.Flush()
}
