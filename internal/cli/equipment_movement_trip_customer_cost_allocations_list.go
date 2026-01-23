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

type equipmentMovementTripCustomerCostAllocationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Trip    string
}

type equipmentMovementTripCustomerCostAllocationRow struct {
	ID         string         `json:"id"`
	TripID     string         `json:"trip_id,omitempty"`
	IsExplicit bool           `json:"is_explicit"`
	Allocation map[string]any `json:"allocation,omitempty"`
}

func newEquipmentMovementTripCustomerCostAllocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement trip customer cost allocations",
		Long: `List customer cost allocations for equipment movement trips.

Output Columns:
  ID         Allocation identifier
  TRIP       Equipment movement trip ID
  EXPLICIT   Whether the allocation was explicitly set
  ALLOCATION Allocation summary (customer_id=percentage)

Filters:
  --trip    Filter by equipment movement trip ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List cost allocations
  xbe view equipment-movement-trip-customer-cost-allocations list

  # Filter by trip
  xbe view equipment-movement-trip-customer-cost-allocations list --trip 123

  # Output as JSON
  xbe view equipment-movement-trip-customer-cost-allocations list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentMovementTripCustomerCostAllocationsList,
	}
	initEquipmentMovementTripCustomerCostAllocationsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripCustomerCostAllocationsCmd.AddCommand(newEquipmentMovementTripCustomerCostAllocationsListCmd())
}

func initEquipmentMovementTripCustomerCostAllocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("trip", "", "Filter by equipment movement trip ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripCustomerCostAllocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementTripCustomerCostAllocationsListOptions(cmd)
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
	query.Set("fields[equipment-movement-trip-customer-cost-allocations]", "is-explicit,allocation,trip")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[trip]", opts.Trip)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trip-customer-cost-allocations", query)
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

	rows := buildEquipmentMovementTripCustomerCostAllocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementTripCustomerCostAllocationsTable(cmd, rows)
}

func parseEquipmentMovementTripCustomerCostAllocationsListOptions(cmd *cobra.Command) (equipmentMovementTripCustomerCostAllocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trip, _ := cmd.Flags().GetString("trip")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripCustomerCostAllocationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Trip:    trip,
	}, nil
}

func buildEquipmentMovementTripCustomerCostAllocationRows(resp jsonAPIResponse) []equipmentMovementTripCustomerCostAllocationRow {
	rows := make([]equipmentMovementTripCustomerCostAllocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildEquipmentMovementTripCustomerCostAllocationRow(resource)

		rows = append(rows, row)
	}
	return rows
}

func equipmentMovementTripCustomerCostAllocationRowFromSingle(resp jsonAPISingleResponse) equipmentMovementTripCustomerCostAllocationRow {
	return buildEquipmentMovementTripCustomerCostAllocationRow(resp.Data)
}

func buildEquipmentMovementTripCustomerCostAllocationRow(resource jsonAPIResource) equipmentMovementTripCustomerCostAllocationRow {
	row := equipmentMovementTripCustomerCostAllocationRow{
		ID:         resource.ID,
		IsExplicit: boolAttr(resource.Attributes, "is-explicit"),
		Allocation: allocationAttr(resource.Attributes, "allocation"),
	}

	if rel, ok := resource.Relationships["trip"]; ok && rel.Data != nil {
		row.TripID = rel.Data.ID
	}

	return row
}

func renderEquipmentMovementTripCustomerCostAllocationsTable(cmd *cobra.Command, rows []equipmentMovementTripCustomerCostAllocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement trip customer cost allocations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRIP\tEXPLICIT\tALLOCATION")
	for _, row := range rows {
		explicit := "no"
		if row.IsExplicit {
			explicit = "yes"
		}
		allocation := allocationSummary(row.Allocation)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.TripID,
			explicit,
			allocation,
		)
	}
	return writer.Flush()
}

func allocationAttr(attrs map[string]any, key string) map[string]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(text), &parsed); err == nil {
			return parsed
		}
	}
	return nil
}

func allocationSummary(allocation map[string]any) string {
	if len(allocation) == 0 {
		return ""
	}
	details, ok := allocation["details"]
	if !ok || details == nil {
		return ""
	}
	items, ok := details.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		detail, ok := item.(map[string]any)
		if !ok {
			continue
		}
		customerID := stringAttr(detail, "customer_id")
		if customerID == "" {
			customerID = stringAttr(detail, "customer-id")
		}
		percentage := stringAttr(detail, "percentage")
		if customerID != "" && percentage != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", customerID, percentage))
		} else if customerID != "" {
			parts = append(parts, customerID)
		} else if percentage != "" {
			parts = append(parts, percentage)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return truncateString(strings.Join(parts, ","), 60)
}
