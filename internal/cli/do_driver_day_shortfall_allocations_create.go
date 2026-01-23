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

type doDriverDayShortfallAllocationsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	TimeCardIDs                    []string
	ShiftSetTimeCardConstraintIDs  []string
	ServiceTypeUnitOfMeasureID     string
	Quantity                       string
	AllocationQuantitiesAttributes string
}

type driverDayShortfallAllocationRow struct {
	ID                                  string   `json:"id"`
	Quantity                            string   `json:"quantity,omitempty"`
	TimeCardIDs                         []string `json:"time_card_ids,omitempty"`
	ShiftSetTimeCardConstraintIDs       []string `json:"shift_set_time_card_constraint_ids,omitempty"`
	ServiceTypeUnitOfMeasureID          string   `json:"service_type_unit_of_measure_id,omitempty"`
	JobIDs                              []string `json:"job_ids,omitempty"`
	TenderIDs                           []string `json:"tender_ids,omitempty"`
	RateIDs                             []string `json:"rate_ids,omitempty"`
	ServiceTypeUnitOfMeasureQuantityIDs []string `json:"service_type_unit_of_measure_quantity_ids,omitempty"`
	AllocationQuantities                any      `json:"allocation_quantities,omitempty"`
}

func newDoDriverDayShortfallAllocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Allocate driver day shortfall quantities",
		Long: `Allocate driver day shortfall quantities.

Distributes a shortfall quantity across one or more time cards and ensures the
related tenders have rates for the service type unit of measure.

Required flags:
  --shift-set-time-card-constraint-ids Shift set time card constraint IDs (comma-separated or repeated)
  --service-type-unit-of-measure      Service type unit of measure ID
  --quantity                          Total shortfall quantity to allocate

Required unless allocation-quantities supplies time_card_id values:
  --time-card-ids                     Time card IDs to allocate across (comma-separated or repeated)

Optional flags:
  --allocation-quantities             Allocation JSON array. Each entry should include
                                      time_card_id and quantity. Percentages must sum to 100% when provided.
                                      Example: '[{"time_card_id":"123","quantity":1.25}]'`,
		Example: `  # Allocate a shortfall across time cards
  xbe do driver-day-shortfall-allocations create \
    --time-card-ids 123,456 \
    --shift-set-time-card-constraint-ids 789 \
    --service-type-unit-of-measure 321 \
    --quantity 2.5

  # Allocate with explicit quantities
  xbe do driver-day-shortfall-allocations create \
    --time-card-ids 123 \
    --shift-set-time-card-constraint-ids 789 \
    --service-type-unit-of-measure 321 \
    --quantity 1 \
    --allocation-quantities '[{"time_card_id":"123","quantity":1}]'`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDayShortfallAllocationsCreate,
	}
	initDoDriverDayShortfallAllocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayShortfallAllocationsCmd.AddCommand(newDoDriverDayShortfallAllocationsCreateCmd())
}

func initDoDriverDayShortfallAllocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("time-card-ids", nil, "Time card IDs to allocate across (comma-separated or repeated)")
	cmd.Flags().StringSlice("shift-set-time-card-constraint-ids", nil, "Shift set time card constraint IDs (comma-separated or repeated)")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID")
	cmd.Flags().String("quantity", "", "Total shortfall quantity to allocate")
	cmd.Flags().String("allocation-quantities", "", "Allocation quantities JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayShortfallAllocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverDayShortfallAllocationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	timeCardIDs := compactStringSlice(opts.TimeCardIDs)
	constraintIDs := compactStringSlice(opts.ShiftSetTimeCardConstraintIDs)

	if opts.ServiceTypeUnitOfMeasureID == "" {
		err := fmt.Errorf("--service-type-unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Quantity == "" {
		err := fmt.Errorf("--quantity is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if len(constraintIDs) == 0 {
		err := fmt.Errorf("--shift-set-time-card-constraint-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var allocationQuantities []map[string]any
	if strings.TrimSpace(opts.AllocationQuantitiesAttributes) != "" {
		if err := json.Unmarshal([]byte(opts.AllocationQuantitiesAttributes), &allocationQuantities); err != nil {
			err := fmt.Errorf("invalid allocation-quantities JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if len(timeCardIDs) == 0 {
		if len(allocationQuantities) == 0 {
			err := fmt.Errorf("--time-card-ids is required when --allocation-quantities is not provided")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		derived, err := deriveTimeCardIDs(allocationQuantities)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		timeCardIDs = derived
	}

	attributes := map[string]any{
		"quantity": opts.Quantity,
	}
	if len(allocationQuantities) > 0 {
		attributes["time-card-shortfall-allocation-quantities-attributes"] = allocationQuantities
	}

	relationships := map[string]any{
		"time-cards": map[string]any{
			"data": buildRelationshipData("time-cards", timeCardIDs),
		},
		"shift-set-time-card-constraints": map[string]any{
			"data": buildRelationshipData("shift-set-time-card-constraints", constraintIDs),
		},
		"service-type-unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.ServiceTypeUnitOfMeasureID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-day-shortfall-allocations",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-shortfall-allocations", jsonBody)
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

	row := buildDriverDayShortfallAllocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver day shortfall allocation %s\n", row.ID)
	return nil
}

func parseDoDriverDayShortfallAllocationsCreateOptions(cmd *cobra.Command) (doDriverDayShortfallAllocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCardIDs, _ := cmd.Flags().GetStringSlice("time-card-ids")
	shiftSetTimeCardConstraintIDs, _ := cmd.Flags().GetStringSlice("shift-set-time-card-constraint-ids")
	serviceTypeUnitOfMeasureID, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	allocationQuantitiesAttributes, _ := cmd.Flags().GetString("allocation-quantities")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayShortfallAllocationsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		TimeCardIDs:                    timeCardIDs,
		ShiftSetTimeCardConstraintIDs:  shiftSetTimeCardConstraintIDs,
		ServiceTypeUnitOfMeasureID:     serviceTypeUnitOfMeasureID,
		Quantity:                       quantity,
		AllocationQuantitiesAttributes: allocationQuantitiesAttributes,
	}, nil
}

func buildDriverDayShortfallAllocationRowFromSingle(resp jsonAPISingleResponse) driverDayShortfallAllocationRow {
	attrs := resp.Data.Attributes
	relationships := resp.Data.Relationships

	row := driverDayShortfallAllocationRow{
		ID:                                  resp.Data.ID,
		Quantity:                            stringAttr(attrs, "quantity"),
		TimeCardIDs:                         relationshipIDsFromMap(relationships, "time-cards"),
		ShiftSetTimeCardConstraintIDs:       relationshipIDsFromMap(relationships, "shift-set-time-card-constraints"),
		ServiceTypeUnitOfMeasureID:          relationshipIDFromMap(relationships, "service-type-unit-of-measure"),
		JobIDs:                              relationshipIDsFromMap(relationships, "jobs"),
		TenderIDs:                           relationshipIDsFromMap(relationships, "tenders"),
		RateIDs:                             relationshipIDsFromMap(relationships, "rates"),
		ServiceTypeUnitOfMeasureQuantityIDs: relationshipIDsFromMap(relationships, "service-type-unit-of-measure-quantities"),
	}

	if len(row.TimeCardIDs) == 0 {
		row.TimeCardIDs = stringSliceAttr(attrs, "time-card-ids")
	}
	if len(row.ShiftSetTimeCardConstraintIDs) == 0 {
		row.ShiftSetTimeCardConstraintIDs = stringSliceAttr(attrs, "shift-set-time-card-constraint-ids")
	}
	if row.ServiceTypeUnitOfMeasureID == "" {
		row.ServiceTypeUnitOfMeasureID = stringAttr(attrs, "service-type-unit-of-measure-id")
	}
	if attrs != nil {
		if allocationQuantities, ok := attrs["time-card-shortfall-allocation-quantities-attributes"]; ok {
			row.AllocationQuantities = allocationQuantities
		}
	}

	return row
}

func compactStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func deriveTimeCardIDs(allocationQuantities []map[string]any) ([]string, error) {
	if len(allocationQuantities) == 0 {
		return nil, fmt.Errorf("allocation-quantities must include at least one entry")
	}
	seen := make(map[string]struct{}, len(allocationQuantities))
	ids := make([]string, 0, len(allocationQuantities))
	for index, entry := range allocationQuantities {
		var rawID any
		var ok bool
		if rawID, ok = entry["time_card_id"]; !ok {
			rawID, ok = entry["time-card-id"]
		}
		if !ok || rawID == nil {
			return nil, fmt.Errorf("allocation-quantities entry %d is missing time_card_id", index)
		}
		id := strings.TrimSpace(fmt.Sprintf("%v", rawID))
		if id == "" || id == "<nil>" {
			return nil, fmt.Errorf("allocation-quantities entry %d has an empty time_card_id", index)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func buildRelationshipData(resourceType string, ids []string) []map[string]any {
	data := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		data = append(data, map[string]any{
			"type": resourceType,
			"id":   id,
		})
	}
	return data
}

func relationshipIDFromMap(relationships map[string]jsonAPIRelationship, key string) string {
	if relationships == nil {
		return ""
	}
	rel, ok := relationships[key]
	if !ok || rel.Data == nil {
		return ""
	}
	return rel.Data.ID
}

func relationshipIDsFromMap(relationships map[string]jsonAPIRelationship, key string) []string {
	if relationships == nil {
		return nil
	}
	rel, ok := relationships[key]
	if !ok {
		return nil
	}
	if rel.Data != nil {
		return []string{rel.Data.ID}
	}
	if len(rel.raw) == 0 || string(rel.raw) == "null" {
		return nil
	}
	var identifiers []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &identifiers); err != nil {
		return nil
	}
	ids := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		if identifier.ID == "" {
			continue
		}
		ids = append(ids, identifier.ID)
	}
	return ids
}
