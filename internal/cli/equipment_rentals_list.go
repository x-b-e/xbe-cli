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

type equipmentRentalsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Broker                  string
	EquipmentClassification string
	Equipment               string
	EquipmentSupplier       string
	StartOnMin              string
	StartOnMax              string
	EndOnMin                string
	EndOnMax                string
	Status                  string
}

type equipmentRentalRow struct {
	ID                        string `json:"id"`
	Description               string `json:"description,omitempty"`
	Status                    string `json:"status,omitempty"`
	StartOn                   string `json:"start_on,omitempty"`
	StartOnPlanned            string `json:"start_on_planned,omitempty"`
	EndOn                     string `json:"end_on,omitempty"`
	EndOnPlanned              string `json:"end_on_planned,omitempty"`
	ApproximateCostPerDay     string `json:"approximate_cost_per_day,omitempty"`
	CostPerHour               string `json:"cost_per_hour,omitempty"`
	TargetUtilizationHours    string `json:"target_utilization_hours,omitempty"`
	ActualRentalUsageHours    string `json:"actual_rental_usage_hours,omitempty"`
	SkipWeekend               bool   `json:"skip_weekend"`
	BrokerID                  string `json:"broker_id,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
	EquipmentID               string `json:"equipment_id,omitempty"`
	EquipmentSupplierID       string `json:"equipment_supplier_id,omitempty"`
}

func newEquipmentRentalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment rentals",
		Long: `List equipment rentals.

Output Columns:
  ID              Equipment rental identifier
  DESCRIPTION     Description
  STATUS          Rental status
  START           Start date
  END             End date
  BROKER          Broker ID

Filters:
  --broker                   Filter by broker ID
  --equipment-classification Filter by equipment classification ID
  --equipment                Filter by equipment ID
  --equipment-supplier       Filter by equipment supplier ID
  --start-on-min             Filter by minimum start date
  --start-on-max             Filter by maximum start date
  --end-on-min               Filter by minimum end date
  --end-on-max               Filter by maximum end date
  --status                   Filter by status`,
		Example: `  # List all equipment rentals
  xbe view equipment-rentals list

  # Filter by broker
  xbe view equipment-rentals list --broker 123

  # Filter by status
  xbe view equipment-rentals list --status active

  # Filter by date range
  xbe view equipment-rentals list --start-on-min "2024-01-01" --end-on-max "2024-12-31"

  # Output as JSON
  xbe view equipment-rentals list --json`,
		RunE: runEquipmentRentalsList,
	}
	initEquipmentRentalsListFlags(cmd)
	return cmd
}

func init() {
	equipmentRentalsCmd.AddCommand(newEquipmentRentalsListCmd())
}

func initEquipmentRentalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment-classification", "", "Filter by equipment classification ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("equipment-supplier", "", "Filter by equipment supplier ID")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date")
	cmd.Flags().String("end-on-min", "", "Filter by minimum end date")
	cmd.Flags().String("end-on-max", "", "Filter by maximum end date")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentRentalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentRentalsListOptions(cmd)
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
	query.Set("include", "broker,equipment-classification,equipment,equipment-supplier")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[equipment_classification]", opts.EquipmentClassification)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[equipment_supplier]", opts.EquipmentSupplier)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[end_on_min]", opts.EndOnMin)
	setFilterIfPresent(query, "filter[end_on_max]", opts.EndOnMax)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-rentals", query)
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

	rows := buildEquipmentRentalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentRentalsTable(cmd, rows)
}

func parseEquipmentRentalsListOptions(cmd *cobra.Command) (equipmentRentalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	equipment, _ := cmd.Flags().GetString("equipment")
	equipmentSupplier, _ := cmd.Flags().GetString("equipment-supplier")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	endOnMin, _ := cmd.Flags().GetString("end-on-min")
	endOnMax, _ := cmd.Flags().GetString("end-on-max")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentRentalsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Broker:                  broker,
		EquipmentClassification: equipmentClassification,
		Equipment:               equipment,
		EquipmentSupplier:       equipmentSupplier,
		StartOnMin:              startOnMin,
		StartOnMax:              startOnMax,
		EndOnMin:                endOnMin,
		EndOnMax:                endOnMax,
		Status:                  status,
	}, nil
}

func buildEquipmentRentalRows(resp jsonAPIResponse) []equipmentRentalRow {
	rows := make([]equipmentRentalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentRentalRow{
			ID:                     resource.ID,
			Description:            stringAttr(resource.Attributes, "description"),
			Status:                 stringAttr(resource.Attributes, "status"),
			StartOn:                stringAttr(resource.Attributes, "start-on"),
			StartOnPlanned:         stringAttr(resource.Attributes, "start-on-planned"),
			EndOn:                  stringAttr(resource.Attributes, "end-on"),
			EndOnPlanned:           stringAttr(resource.Attributes, "end-on-planned"),
			ApproximateCostPerDay:  stringAttr(resource.Attributes, "approximate-cost-per-day"),
			CostPerHour:            stringAttr(resource.Attributes, "cost-per-hour"),
			TargetUtilizationHours: stringAttr(resource.Attributes, "target-utilization-hours"),
			ActualRentalUsageHours: stringAttr(resource.Attributes, "actual-rental-usage-hours"),
			SkipWeekend:            boolAttr(resource.Attributes, "skip-weekend"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
			row.EquipmentClassificationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
			row.EquipmentID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["equipment-supplier"]; ok && rel.Data != nil {
			row.EquipmentSupplierID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentRentalsTable(cmd *cobra.Command, rows []equipmentRentalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment rentals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tSTATUS\tSTART\tEND\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Description, 30),
			row.Status,
			row.StartOn,
			row.EndOn,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
