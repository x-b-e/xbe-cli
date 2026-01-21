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

type doTransportSummaryCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	EntityType  string
	FiltersJSON string
	FilterPairs []string
}

func newDoTransportSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a transport summary",
		Long: `Create a transport summary.

Provides status breakdowns and counts for transport operations. Unlike other
summaries that aggregate metrics, this returns status distributions showing
how many orders, plans, or assignments are in each state (editing, approved,
in-transit, complete, etc.).

REQUIRED: You must specify an entity type with --entity-type.

Entity types:
  transport_order     Customer orders: status counts, lifecycle states, overdue/at-risk counts
  transport_plan      Dispatch plans: status breakdown for trucking assignments
  driver_assignment   Driver assignments: status of driver-to-load assignments
  live_loads          Active loads: real-time counts of loads in transit, at pickup, etc.

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Required filters:
    broker                   Broker ID (required for all entity types)

  Available filters by entity type:

  transport_order:
    customer                 Customer ID
    status                   Status (editing, submitted, approved, cancelled, complete, scrapped)
    start_date               Start date (YYYY-MM-DD)
    end_date                 End date (YYYY-MM-DD)
    project_office           Project office ID
    project_category         Project category ID
    (trucker is NOT valid for transport_order)

  transport_plan:
    customer                 Customer ID
    trucker                  Trucker ID
    status                   Status (editing, approved, complete)
    start_date               Start date (YYYY-MM-DD)
    end_date                 End date (YYYY-MM-DD)
    project_office           Project office ID
    project_category         Project category ID

  driver_assignment:
    trucker                  Trucker ID
    status                   Status (editing, pending, active)
    start_date               Start date (YYYY-MM-DD)
    end_date                 End date (YYYY-MM-DD)
    project_office           Project office ID
    project_category         Project category ID
    (customer is NOT valid for driver_assignment)

  live_loads:
    customer                 Customer ID
    status                   Status (editing, submitted, approved)
    start_date               Start date (YYYY-MM-DD)
    end_date                 End date (YYYY-MM-DD)
    project_office           Project office ID
    project_category         Project category ID
    (trucker is NOT valid for live_loads)`,
		Example: `  # Transport order summary
  xbe summarize transport-summary create --entity-type transport_order --filter broker=123 --filter start_date=2025-01-01 --filter end_date=2025-01-31

  # Transport plan summary
  xbe summarize transport-summary create --entity-type transport_plan --filter broker=123

  # Driver assignment summary
  xbe summarize transport-summary create --entity-type driver_assignment --filter broker=123 --filter trucker=456

  # Live loads summary
  xbe summarize transport-summary create --entity-type live_loads --filter broker=123 --filter customer=789

  # JSON output
  xbe summarize transport-summary create --entity-type transport_order --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTransportSummaryCreate,
	}
	initDoTransportSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportSummaryCmd.AddCommand(newDoTransportSummaryCreateCmd())
}

func initDoTransportSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("entity-type", "", "Entity type (transport_order, transport_plan, driver_assignment, live_loads)")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("entity-type")
}

func runDoTransportSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportSummaryCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.EntityType == "" {
		err := errors.New("--entity-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	validEntityTypes := map[string]bool{
		"transport_order":   true,
		"transport_plan":    true,
		"driver_assignment": true,
		"live_loads":        true,
	}
	if !validEntityTypes[opts.EntityType] {
		err := fmt.Errorf("invalid --entity-type %q (must be transport_order, transport_plan, driver_assignment, or live_loads)", opts.EntityType)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			err := errors.New("authentication required. Run 'xbe auth login' first")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	filters, err := parseTransportSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Add entity_type to filters
	filters["entity_type"] = opts.EntityType

	attributes := map[string]any{
		"filters": filters,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "transport-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/transport-summaries", jsonBody)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), resp.Data.Attributes)
	}

	return renderTransportSummary(cmd, opts.EntityType, resp.Data.Attributes)
}

func parseDoTransportSummaryCreateOptions(cmd *cobra.Command) (doTransportSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doTransportSummaryCreateOptions{}, err
	}
	entityType, err := cmd.Flags().GetString("entity-type")
	if err != nil {
		return doTransportSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doTransportSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doTransportSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doTransportSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doTransportSummaryCreateOptions{}, err
	}

	return doTransportSummaryCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		EntityType:  entityType,
		FiltersJSON: filtersJSON,
		FilterPairs: filterPairs,
	}, nil
}

func parseTransportSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
	filters := map[string]any{}

	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON != "" {
		if err := json.Unmarshal([]byte(rawJSON), &filters); err != nil {
			return nil, fmt.Errorf("invalid --filters JSON: %w", err)
		}
	}

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --filter %q (expected key=value)", pair)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid --filter %q (missing key)", pair)
		}
		filters[key] = value
	}

	return filters, nil
}

func renderTransportSummary(cmd *cobra.Command, entityType string, attrs map[string]any) error {
	if attrs == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport summary data found.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Transport Summary (%s)\n", entityType)
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 40))

	switch entityType {
	case "transport_order":
		renderTransportOrderSummary(cmd, attrs)
	case "transport_plan":
		renderTransportPlanSummary(cmd, attrs)
	case "driver_assignment":
		renderDriverAssignmentSummary(cmd, attrs)
	case "live_loads":
		renderLiveLoadsSummary(cmd, attrs)
	default:
		// Generic rendering for unknown types
		for key, value := range attrs {
			fmt.Fprintf(cmd.OutOrStdout(), "%-30s %v\n", key+":", value)
		}
	}

	return nil
}

func renderTransportOrderSummary(cmd *cobra.Command, attrs map[string]any) {
	// Status breakdown
	if statuses, ok := attrs["transport_orders"].(map[string]any); ok {
		fmt.Fprintln(cmd.OutOrStdout(), "\nStatus Breakdown:")
		for status, count := range statuses {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %v\n", status+":", count)
		}
	}

	// Other counts
	printSummaryField(cmd, attrs, "lifecycle_status", "Lifecycle Status")
	printSummaryField(cmd, attrs, "unplanned", "Unplanned")
	printSummaryField(cmd, attrs, "overdue", "Overdue")
	printSummaryField(cmd, attrs, "at_risk", "At Risk")
}

func renderTransportPlanSummary(cmd *cobra.Command, attrs map[string]any) {
	if statuses, ok := attrs["transport_plans"].(map[string]any); ok {
		fmt.Fprintln(cmd.OutOrStdout(), "\nStatus Breakdown:")
		for status, count := range statuses {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %v\n", status+":", count)
		}
	}
}

func renderDriverAssignmentSummary(cmd *cobra.Command, attrs map[string]any) {
	if statuses, ok := attrs["driver_assignments"].(map[string]any); ok {
		fmt.Fprintln(cmd.OutOrStdout(), "\nStatus Breakdown:")
		for status, count := range statuses {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %v\n", status+":", count)
		}
	}
}

func renderLiveLoadsSummary(cmd *cobra.Command, attrs map[string]any) {
	printSummaryField(cmd, attrs, "active", "Active")
	printSummaryField(cmd, attrs, "in_transit", "In Transit")
	printSummaryField(cmd, attrs, "at_pickup", "At Pickup")
	printSummaryField(cmd, attrs, "at_delivery", "At Delivery")
	printSummaryField(cmd, attrs, "on_schedule", "On Schedule")
	printSummaryField(cmd, attrs, "attention_needed", "Attention Needed")
	printSummaryField(cmd, attrs, "delayed", "Delayed")
}

func printSummaryField(cmd *cobra.Command, attrs map[string]any, key, label string) {
	if value, ok := attrs[key]; ok {
		fmt.Fprintf(cmd.OutOrStdout(), "%-30s %v\n", label+":", value)
	}
}
