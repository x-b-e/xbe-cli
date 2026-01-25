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

type doShiftSetTimeCardConstraintsCreateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
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
	Parent                                   string
	ShiftScope                               string
	ServiceTypeUnitOfMeasures                string
	TrailerClassifications                   string
	MaterialTypes                            string
}

func newDoShiftSetTimeCardConstraintsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a shift set time card constraint",
		Long: `Create a shift set time card constraint.

Required flags:
  --constrained-type                           Constrained type (tenders, rate-agreements)
  --constrained-id                             Constrained ID
  --constraint-type                            Constraint type (minimum, equality, maximum)
  --constrained-amount-type                    Constrained amount type (effective, base)
  --currency-code                              Currency code (USD)
  --status                                     Status (active, inactive)
  --service-type-unit-of-measures-constraint-type Service type unit of measures constraint type (applicable, not_applicable)
  --shift-set-grouped-by                       Shift set grouped by (customer, broker)

Amount options (choose one):
  --constrained-amount                         Explicit constrained amount
  --calculated-constrained-amount-price-per-unit  Calculated price per unit
  --calculated-constrained-amount-unit-of-measure Calculated unit of measure ID (unit-of-measures)

Optional flags:
  --name                                      Constraint name
  --parent                                    Parent constraint ID (create-only)
  --shift-scope                               Shift scope ID
  --service-type-unit-of-measures             Service type unit of measure IDs (comma-separated)
  --trailer-classifications                   Trailer classification IDs (comma-separated)
  --material-types                            Material type IDs (comma-separated)`,
		Example: `  # Create a minimum constraint with explicit amount
  xbe do shift-set-time-card-constraints create \
    --constrained-type rate-agreements \
    --constrained-id 123 \
    --constraint-type minimum \
    --constrained-amount-type effective \
    --currency-code USD \
    --status active \
    --service-type-unit-of-measures-constraint-type applicable \
    --shift-set-grouped-by broker \
    --constrained-amount 500.00

  # Create a calculated constraint
  xbe do shift-set-time-card-constraints create \
    --constrained-type tenders \
    --constrained-id 456 \
    --constraint-type minimum \
    --constrained-amount-type effective \
    --currency-code USD \
    --status active \
    --service-type-unit-of-measures-constraint-type applicable \
    --shift-set-grouped-by customer \
    --calculated-constrained-amount-price-per-unit 45.00 \
    --calculated-constrained-amount-unit-of-measure 789`,
		Args: cobra.NoArgs,
		RunE: runDoShiftSetTimeCardConstraintsCreate,
	}
	initDoShiftSetTimeCardConstraintsCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftSetTimeCardConstraintsCmd.AddCommand(newDoShiftSetTimeCardConstraintsCreateCmd())
}

func initDoShiftSetTimeCardConstraintsCreateFlags(cmd *cobra.Command) {
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
	cmd.Flags().String("calculated-constrained-amount-unit-of-measure", "", "Calculated unit of measure ID (unit-of-measures)")
	cmd.Flags().String("parent", "", "Parent constraint ID (create-only)")
	cmd.Flags().String("shift-scope", "", "Shift scope ID")
	cmd.Flags().String("service-type-unit-of-measures", "", "Service type unit of measure IDs (comma-separated)")
	cmd.Flags().String("trailer-classifications", "", "Trailer classification IDs (comma-separated)")
	cmd.Flags().String("material-types", "", "Material type IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("constrained-type")
	_ = cmd.MarkFlagRequired("constrained-id")
	_ = cmd.MarkFlagRequired("constraint-type")
	_ = cmd.MarkFlagRequired("constrained-amount-type")
	_ = cmd.MarkFlagRequired("currency-code")
	_ = cmd.MarkFlagRequired("status")
	_ = cmd.MarkFlagRequired("service-type-unit-of-measures-constraint-type")
	_ = cmd.MarkFlagRequired("shift-set-grouped-by")
}

func runDoShiftSetTimeCardConstraintsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoShiftSetTimeCardConstraintsCreateOptions(cmd)
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

	if opts.ConstrainedType == "" {
		err := fmt.Errorf("--constrained-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ConstrainedID == "" {
		err := fmt.Errorf("--constrained-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ConstraintType == "" {
		err := fmt.Errorf("--constraint-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ConstrainedAmountType == "" {
		err := fmt.Errorf("--constrained-amount-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CurrencyCode == "" {
		err := fmt.Errorf("--currency-code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Status == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ServiceTypeUnitOfMeasuresConstraintType == "" {
		err := fmt.Errorf("--service-type-unit-of-measures-constraint-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ShiftSetGroupedBy == "" {
		err := fmt.Errorf("--shift-set-grouped-by is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	usesCalculated := opts.CalculatedConstrainedAmountPricePerUnit != "" || opts.CalculatedConstrainedAmountUnitOfMeasure != ""
	if usesCalculated {
		if opts.CalculatedConstrainedAmountPricePerUnit == "" || opts.CalculatedConstrainedAmountUnitOfMeasure == "" {
			err := fmt.Errorf("calculated constraints require both --calculated-constrained-amount-price-per-unit and --calculated-constrained-amount-unit-of-measure")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if opts.ConstrainedAmount != "" {
			err := fmt.Errorf("--constrained-amount cannot be set when using calculated constraint fields")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	} else if opts.ConstrainedAmount == "" {
		err := fmt.Errorf("--constrained-amount is required when not using calculated constraint fields")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.ConstrainedAmount != "" {
		attributes["constrained-amount"] = opts.ConstrainedAmount
	}
	if opts.ConstraintType != "" {
		attributes["constraint-type"] = opts.ConstraintType
	}
	if opts.ConstrainedAmountType != "" {
		attributes["constrained-amount-type"] = opts.ConstrainedAmountType
	}
	if opts.CurrencyCode != "" {
		attributes["currency-code"] = opts.CurrencyCode
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.ServiceTypeUnitOfMeasuresConstraintType != "" {
		attributes["service-type-unit-of-measures-constraint-type"] = opts.ServiceTypeUnitOfMeasuresConstraintType
	}
	if opts.ShiftSetGroupedBy != "" {
		attributes["shift-set-grouped-by"] = opts.ShiftSetGroupedBy
	}
	if opts.CalculatedConstrainedAmountPricePerUnit != "" {
		attributes["calculated-constrained-amount-price-per-unit"] = opts.CalculatedConstrainedAmountPricePerUnit
	}

	relationships := map[string]any{
		"constrained": map[string]any{
			"data": map[string]any{
				"type": opts.ConstrainedType,
				"id":   opts.ConstrainedID,
			},
		},
	}

	if opts.Parent != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": "shift-set-time-card-constraints",
				"id":   opts.Parent,
			},
		}
	}
	if opts.ShiftScope != "" {
		relationships["shift-scope"] = map[string]any{
			"data": map[string]any{
				"type": "shift-scopes",
				"id":   opts.ShiftScope,
			},
		}
	}
	if opts.CalculatedConstrainedAmountUnitOfMeasure != "" {
		relationships["calculated-constrained-amount-unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.CalculatedConstrainedAmountUnitOfMeasure,
			},
		}
	}
	if opts.ServiceTypeUnitOfMeasures != "" {
		ids := splitCommaSeparatedIDs(opts.ServiceTypeUnitOfMeasures)
		if len(ids) > 0 {
			relationships["service-type-unit-of-measures"] = map[string]any{
				"data": buildRelationshipDataList("service-type-unit-of-measures", ids),
			}
		}
	}
	if opts.TrailerClassifications != "" {
		ids := splitCommaSeparatedIDs(opts.TrailerClassifications)
		if len(ids) > 0 {
			relationships["trailer-classifications"] = map[string]any{
				"data": buildRelationshipDataList("trailer-classifications", ids),
			}
		}
	}
	if opts.MaterialTypes != "" {
		ids := splitCommaSeparatedIDs(opts.MaterialTypes)
		if len(ids) > 0 {
			relationships["material-types"] = map[string]any{
				"data": buildRelationshipDataList("material-types", ids),
			}
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "shift-set-time-card-constraints",
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

	body, _, err := client.Post(cmd.Context(), "/v1/shift-set-time-card-constraints", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created shift set time card constraint %s\n", row.ID)
	return nil
}

func parseDoShiftSetTimeCardConstraintsCreateOptions(cmd *cobra.Command) (doShiftSetTimeCardConstraintsCreateOptions, error) {
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
	parent, _ := cmd.Flags().GetString("parent")
	shiftScope, _ := cmd.Flags().GetString("shift-scope")
	serviceTypeUnitOfMeasures, _ := cmd.Flags().GetString("service-type-unit-of-measures")
	trailerClassifications, _ := cmd.Flags().GetString("trailer-classifications")
	materialTypes, _ := cmd.Flags().GetString("material-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftSetTimeCardConstraintsCreateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
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
		Parent:                                   parent,
		ShiftScope:                               shiftScope,
		ServiceTypeUnitOfMeasures:                serviceTypeUnitOfMeasures,
		TrailerClassifications:                   trailerClassifications,
		MaterialTypes:                            materialTypes,
	}, nil
}

func splitCommaSeparatedIDs(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func buildRelationshipDataList(resourceType string, ids []string) []map[string]any {
	data := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		data = append(data, map[string]any{
			"type": resourceType,
			"id":   id,
		})
	}
	return data
}
