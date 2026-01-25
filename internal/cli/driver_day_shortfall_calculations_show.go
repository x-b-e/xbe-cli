package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type driverDayShortfallCalculationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverDayShortfallCalculationDetails struct {
	ID                                    string   `json:"id"`
	Quantity                              *float64 `json:"quantity,omitempty"`
	ServiceTypeUnitOfMeasureID            string   `json:"service_type_unit_of_measure_id,omitempty"`
	TimeCardIDs                           []string `json:"time_card_ids,omitempty"`
	UnallocatableTimeCardIDs              []string `json:"unallocatable_time_card_ids,omitempty"`
	AllocatableTimeCardIDs                []string `json:"allocatable_time_card_ids,omitempty"`
	DriverDayTimeCardConstraintIDs        []string `json:"driver_day_time_card_constraint_ids,omitempty"`
	TimeCardShortfallAllocationQuantities any      `json:"time_card_shortfall_allocation_quantities,omitempty"`
}

func newDriverDayShortfallCalculationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver day shortfall calculation details",
		Long: `Show the full details of a driver day shortfall calculation.

Output Fields:
  ID                        Calculation identifier
  Quantity                  Shortfall quantity
  Service Type UOM          Unit of measure for the calculated quantity
  Time Card IDs             Time cards included in the calculation
  Allocatable Time Card IDs Time cards eligible for allocation
  Unallocatable Time Card IDs  Time cards excluded from allocation
  Constraint IDs            Driver day time card constraint IDs
  Allocation Quantities     Per-time-card allocation details

Arguments:
  <id>    Calculation ID (required).`,
		Example: `  # Show a calculation
  xbe view driver-day-shortfall-calculations show 123

  # JSON output
  xbe view driver-day-shortfall-calculations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverDayShortfallCalculationsShow,
	}
	initDriverDayShortfallCalculationsShowFlags(cmd)
	return cmd
}

func init() {
	driverDayShortfallCalculationsCmd.AddCommand(newDriverDayShortfallCalculationsShowCmd())
}

func initDriverDayShortfallCalculationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayShortfallCalculationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverDayShortfallCalculationsShowOptions(cmd)
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
		return fmt.Errorf("driver day shortfall calculation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-shortfall-calculations/"+id, nil)
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

	details := buildDriverDayShortfallCalculationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverDayShortfallCalculationDetails(cmd, details)
}

func parseDriverDayShortfallCalculationsShowOptions(cmd *cobra.Command) (driverDayShortfallCalculationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayShortfallCalculationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverDayShortfallCalculationDetails(resp jsonAPISingleResponse) driverDayShortfallCalculationDetails {
	attrs := resp.Data.Attributes
	details := driverDayShortfallCalculationDetails{
		ID:                                    resp.Data.ID,
		TimeCardIDs:                           stringSliceAttr(attrs, "time-card-ids"),
		UnallocatableTimeCardIDs:              stringSliceAttr(attrs, "unallocatable-time-card-ids"),
		AllocatableTimeCardIDs:                stringSliceAttr(attrs, "allocatable-time-card-ids"),
		DriverDayTimeCardConstraintIDs:        stringSliceAttr(attrs, "driver-day-time-card-constraint-ids"),
		TimeCardShortfallAllocationQuantities: attrs["time-card-shortfall-allocation-quantities-attributes"],
	}

	if value, ok := attrs["quantity"]; ok && value != nil {
		qty := floatAttr(attrs, "quantity")
		details.Quantity = &qty
	}

	if rel, ok := resp.Data.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
		details.ServiceTypeUnitOfMeasureID = rel.Data.ID
	}

	return details
}

func renderDriverDayShortfallCalculationDetails(cmd *cobra.Command, details driverDayShortfallCalculationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Quantity != nil {
		fmt.Fprintf(out, "Quantity: %.2f\n", *details.Quantity)
	} else {
		fmt.Fprintln(out, "Quantity: (none)")
	}
	if details.ServiceTypeUnitOfMeasureID != "" {
		fmt.Fprintf(out, "Service Type UOM ID: %s\n", details.ServiceTypeUnitOfMeasureID)
	} else {
		fmt.Fprintln(out, "Service Type UOM ID: (none)")
	}

	fmt.Fprintf(out, "Time Card IDs: %s\n", strings.Join(details.TimeCardIDs, ", "))
	fmt.Fprintf(out, "Allocatable Time Card IDs: %s\n", strings.Join(details.AllocatableTimeCardIDs, ", "))
	fmt.Fprintf(out, "Unallocatable Time Card IDs: %s\n", strings.Join(details.UnallocatableTimeCardIDs, ", "))
	fmt.Fprintf(out, "Constraint IDs: %s\n", strings.Join(details.DriverDayTimeCardConstraintIDs, ", "))

	fmt.Fprintln(out, "Allocation Quantities:")
	allocations := formatShortfallAllocationDetails(details.TimeCardShortfallAllocationQuantities)
	if allocations == "" {
		fmt.Fprintln(out, "  (none)")
	} else {
		fmt.Fprintln(out, indentLines(allocations, "  "))
	}

	return nil
}

func formatShortfallAllocationDetails(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}

func indentLines(value, prefix string) string {
	if value == "" {
		return ""
	}
	return prefix + strings.ReplaceAll(value, "\n", "\n"+prefix)
}
