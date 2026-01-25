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

type doShiftSetTimeCardConstraintsUpdateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
	ID                                       string
	ConstrainedType                          string
	ConstrainedID                            string
	Name                                     string
	ConstrainedAmount                        string
	ConstraintType                           string
	ConstrainedAmountType                    string
	CurrencyCode                             string
	Status                                   string
	ServiceTypeUnitOfMeasuresConstraintType  string
	ShiftSetGroupedBy                        string
	CalculatedConstrainedAmountPricePerUnit  string
	CalculatedConstrainedAmountUnitOfMeasure string
	ShiftScope                               string
	ServiceTypeUnitOfMeasures                string
	TrailerClassifications                   string
	MaterialTypes                            string
}

func newDoShiftSetTimeCardConstraintsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a shift set time card constraint",
		Long: `Update a shift set time card constraint.

Provide the constraint ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --constrained-type                           Constrained type (tenders, rate-agreements)
  --constrained-id                             Constrained ID
  --name                                      Constraint name
  --constrained-amount                         Explicit constrained amount
  --constraint-type                            Constraint type (minimum, equality, maximum)
  --constrained-amount-type                    Constrained amount type (effective, base)
  --currency-code                              Currency code (USD)
  --status                                     Status (active, inactive)
  --service-type-unit-of-measures-constraint-type Service type unit of measures constraint type (applicable, not_applicable)
  --shift-set-grouped-by                       Shift set grouped by (customer, broker)
  --calculated-constrained-amount-price-per-unit  Calculated price per unit
  --calculated-constrained-amount-unit-of-measure Calculated unit of measure ID
  --shift-scope                               Shift scope ID (empty to clear)
  --service-type-unit-of-measures             Service type unit of measure IDs (comma-separated, empty to clear)
  --trailer-classifications                   Trailer classification IDs (comma-separated, empty to clear)
  --material-types                            Material type IDs (comma-separated, empty to clear)`,
		Example: `  # Update name and status
  xbe do shift-set-time-card-constraints update 123 --name "Updated" --status inactive

  # Update constrained amount
  xbe do shift-set-time-card-constraints update 123 --constrained-amount 750.00

  # Update service type unit of measures
  xbe do shift-set-time-card-constraints update 123 --service-type-unit-of-measures "1,2,3"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoShiftSetTimeCardConstraintsUpdate,
	}
	initDoShiftSetTimeCardConstraintsUpdateFlags(cmd)
	return cmd
}

func init() {
	doShiftSetTimeCardConstraintsCmd.AddCommand(newDoShiftSetTimeCardConstraintsUpdateCmd())
}

func initDoShiftSetTimeCardConstraintsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("constrained-type", "", "Constrained type (tenders, rate-agreements)")
	cmd.Flags().String("constrained-id", "", "Constrained ID")
	cmd.Flags().String("name", "", "Constraint name")
	cmd.Flags().String("constrained-amount", "", "Explicit constrained amount")
	cmd.Flags().String("constraint-type", "", "Constraint type (minimum, equality, maximum)")
	cmd.Flags().String("constrained-amount-type", "", "Constrained amount type (effective, base)")
	cmd.Flags().String("currency-code", "", "Currency code (USD)")
	cmd.Flags().String("status", "", "Status (active, inactive)")
	cmd.Flags().String("service-type-unit-of-measures-constraint-type", "", "Service type unit of measures constraint type (applicable, not_applicable)")
	cmd.Flags().String("shift-set-grouped-by", "", "Shift set grouped by (customer, broker)")
	cmd.Flags().String("calculated-constrained-amount-price-per-unit", "", "Calculated price per unit")
	cmd.Flags().String("calculated-constrained-amount-unit-of-measure", "", "Calculated unit of measure ID")
	cmd.Flags().String("shift-scope", "", "Shift scope ID (empty to clear)")
	cmd.Flags().String("service-type-unit-of-measures", "", "Service type unit of measure IDs (comma-separated, empty to clear)")
	cmd.Flags().String("trailer-classifications", "", "Trailer classification IDs (comma-separated, empty to clear)")
	cmd.Flags().String("material-types", "", "Material type IDs (comma-separated, empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftSetTimeCardConstraintsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftSetTimeCardConstraintsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
		hasChanges = true
	}
	if cmd.Flags().Changed("constrained-amount") {
		attributes["constrained-amount"] = opts.ConstrainedAmount
		hasChanges = true
	}
	if cmd.Flags().Changed("constraint-type") {
		attributes["constraint-type"] = opts.ConstraintType
		hasChanges = true
	}
	if cmd.Flags().Changed("constrained-amount-type") {
		attributes["constrained-amount-type"] = opts.ConstrainedAmountType
		hasChanges = true
	}
	if cmd.Flags().Changed("currency-code") {
		attributes["currency-code"] = opts.CurrencyCode
		hasChanges = true
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
		hasChanges = true
	}
	if cmd.Flags().Changed("service-type-unit-of-measures-constraint-type") {
		attributes["service-type-unit-of-measures-constraint-type"] = opts.ServiceTypeUnitOfMeasuresConstraintType
		hasChanges = true
	}
	if cmd.Flags().Changed("shift-set-grouped-by") {
		attributes["shift-set-grouped-by"] = opts.ShiftSetGroupedBy
		hasChanges = true
	}
	if cmd.Flags().Changed("calculated-constrained-amount-price-per-unit") {
		attributes["calculated-constrained-amount-price-per-unit"] = opts.CalculatedConstrainedAmountPricePerUnit
		hasChanges = true
	}

	constrainedTypeChanged := cmd.Flags().Changed("constrained-type")
	constrainedIDChanged := cmd.Flags().Changed("constrained-id")
	if constrainedTypeChanged || constrainedIDChanged {
		if opts.ConstrainedType == "" || opts.ConstrainedID == "" {
			err := fmt.Errorf("--constrained-type and --constrained-id must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["constrained"] = map[string]any{
			"data": map[string]any{
				"type": opts.ConstrainedType,
				"id":   opts.ConstrainedID,
			},
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("shift-scope") {
		if opts.ShiftScope == "" {
			relationships["shift-scope"] = map[string]any{"data": nil}
		} else {
			relationships["shift-scope"] = map[string]any{
				"data": map[string]any{
					"type": "shift-scopes",
					"id":   opts.ShiftScope,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("calculated-constrained-amount-unit-of-measure") {
		if opts.CalculatedConstrainedAmountUnitOfMeasure == "" {
			relationships["calculated-constrained-amount-unit-of-measure"] = map[string]any{"data": nil}
		} else {
			relationships["calculated-constrained-amount-unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.CalculatedConstrainedAmountUnitOfMeasure,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("service-type-unit-of-measures") {
		if opts.ServiceTypeUnitOfMeasures == "" {
			relationships["service-type-unit-of-measures"] = map[string]any{"data": []any{}}
		} else {
			ids := splitCommaSeparatedIDs(opts.ServiceTypeUnitOfMeasures)
			relationships["service-type-unit-of-measures"] = map[string]any{
				"data": buildRelationshipDataList(ids, "service-type-unit-of-measures"),
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("trailer-classifications") {
		if opts.TrailerClassifications == "" {
			relationships["trailer-classifications"] = map[string]any{"data": []any{}}
		} else {
			ids := splitCommaSeparatedIDs(opts.TrailerClassifications)
			relationships["trailer-classifications"] = map[string]any{
				"data": buildRelationshipDataList(ids, "trailer-classifications"),
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("material-types") {
		if opts.MaterialTypes == "" {
			relationships["material-types"] = map[string]any{"data": []any{}}
		} else {
			ids := splitCommaSeparatedIDs(opts.MaterialTypes)
			relationships["material-types"] = map[string]any{
				"data": buildRelationshipDataList(ids, "material-types"),
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("at least one attribute or relationship must be specified")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "shift-set-time-card-constraints",
		"id":         opts.ID,
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/shift-set-time-card-constraints/"+opts.ID, jsonBody)
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

	row := buildShiftSetTimeCardConstraintRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated shift set time card constraint %s\n", row.ID)
	return nil
}

func parseDoShiftSetTimeCardConstraintsUpdateOptions(cmd *cobra.Command, args []string) (doShiftSetTimeCardConstraintsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	constrainedType, _ := cmd.Flags().GetString("constrained-type")
	constrainedID, _ := cmd.Flags().GetString("constrained-id")
	name, _ := cmd.Flags().GetString("name")
	constrainedAmount, _ := cmd.Flags().GetString("constrained-amount")
	constraintType, _ := cmd.Flags().GetString("constraint-type")
	constrainedAmountType, _ := cmd.Flags().GetString("constrained-amount-type")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	status, _ := cmd.Flags().GetString("status")
	serviceTypeUnitOfMeasuresConstraintType, _ := cmd.Flags().GetString("service-type-unit-of-measures-constraint-type")
	shiftSetGroupedBy, _ := cmd.Flags().GetString("shift-set-grouped-by")
	calculatedConstrainedAmountPricePerUnit, _ := cmd.Flags().GetString("calculated-constrained-amount-price-per-unit")
	calculatedConstrainedAmountUnitOfMeasure, _ := cmd.Flags().GetString("calculated-constrained-amount-unit-of-measure")
	shiftScope, _ := cmd.Flags().GetString("shift-scope")
	serviceTypeUnitOfMeasures, _ := cmd.Flags().GetString("service-type-unit-of-measures")
	trailerClassifications, _ := cmd.Flags().GetString("trailer-classifications")
	materialTypes, _ := cmd.Flags().GetString("material-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftSetTimeCardConstraintsUpdateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		ID:                                       args[0],
		ConstrainedType:                          constrainedType,
		ConstrainedID:                            constrainedID,
		Name:                                     name,
		ConstrainedAmount:                        constrainedAmount,
		ConstraintType:                           constraintType,
		ConstrainedAmountType:                    constrainedAmountType,
		CurrencyCode:                             currencyCode,
		Status:                                   status,
		ServiceTypeUnitOfMeasuresConstraintType:  serviceTypeUnitOfMeasuresConstraintType,
		ShiftSetGroupedBy:                        shiftSetGroupedBy,
		CalculatedConstrainedAmountPricePerUnit:  calculatedConstrainedAmountPricePerUnit,
		CalculatedConstrainedAmountUnitOfMeasure: calculatedConstrainedAmountUnitOfMeasure,
		ShiftScope:                               shiftScope,
		ServiceTypeUnitOfMeasures:                serviceTypeUnitOfMeasures,
		TrailerClassifications:                   trailerClassifications,
		MaterialTypes:                            materialTypes,
	}, nil
}
