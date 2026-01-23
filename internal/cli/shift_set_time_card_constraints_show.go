package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type shiftSetTimeCardConstraintsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type shiftSetTimeCardConstraintDetails struct {
	ID                                         string   `json:"id"`
	Name                                       string   `json:"name,omitempty"`
	ConstrainedType                            string   `json:"constrained_type,omitempty"`
	ConstrainedID                              string   `json:"constrained_id,omitempty"`
	ParentID                                   string   `json:"parent_id,omitempty"`
	ChildIDs                                   []string `json:"child_ids,omitempty"`
	ConstrainedAmount                          string   `json:"constrained_amount,omitempty"`
	ConstraintType                             string   `json:"constraint_type,omitempty"`
	ConstrainedAmountType                      string   `json:"constrained_amount_type,omitempty"`
	CurrencyCode                               string   `json:"currency_code,omitempty"`
	Status                                     string   `json:"status,omitempty"`
	ServiceTypeUnitOfMeasuresConstraintType    string   `json:"service_type_unit_of_measures_constraint_type,omitempty"`
	ShiftSetGroupedBy                          string   `json:"shift_set_grouped_by,omitempty"`
	CalculatedConstrainedAmountPricePerUnit    string   `json:"calculated_constrained_amount_price_per_unit,omitempty"`
	CalculatedConstrainedAmountUnitOfMeasureID string   `json:"calculated_constrained_amount_unit_of_measure_id,omitempty"`
	ShiftScopeID                               string   `json:"shift_scope_id,omitempty"`
	ServiceTypeUnitOfMeasureIDs                []string `json:"service_type_unit_of_measure_ids,omitempty"`
	TrailerClassificationIDs                   []string `json:"trailer_classification_ids,omitempty"`
	MaterialTypeIDs                            []string `json:"material_type_ids,omitempty"`
	CanDelete                                  bool     `json:"can_delete"`
}

func newShiftSetTimeCardConstraintsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show shift set time card constraint details",
		Long: `Show the full details of a shift set time card constraint.

Output Fields:
  ID
  Name
  Constrained Type/ID
  Parent ID
  Child IDs
  Constrained Amount
  Constraint Type
  Constrained Amount Type
  Currency Code
  Status
  Service Type Unit Of Measures Constraint Type
  Shift Set Grouped By
  Calculated Constrained Amount Price Per Unit
  Calculated Constrained Amount Unit Of Measure ID
  Shift Scope ID
  Service Type Unit Of Measure IDs
  Trailer Classification IDs
  Material Type IDs
  Can Delete

Arguments:
  <id>    The constraint ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a constraint
  xbe view shift-set-time-card-constraints show 123

  # Output as JSON
  xbe view shift-set-time-card-constraints show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runShiftSetTimeCardConstraintsShow,
	}
	initShiftSetTimeCardConstraintsShowFlags(cmd)
	return cmd
}

func init() {
	shiftSetTimeCardConstraintsCmd.AddCommand(newShiftSetTimeCardConstraintsShowCmd())
}

func initShiftSetTimeCardConstraintsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftSetTimeCardConstraintsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseShiftSetTimeCardConstraintsShowOptions(cmd)
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
		return fmt.Errorf("shift set time card constraint id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[shift-set-time-card-constraints]", "name,constrained-amount,constraint-type,constrained-amount-type,currency-code,status,service-type-unit-of-measures-constraint-type,shift-set-grouped-by,calculated-constrained-amount-price-per-unit,can-delete,constrained,parent,children,service-type-unit-of-measures,trailer-classifications,material-types,shift-scope,calculated-constrained-amount-unit-of-measure")
	query.Set("include", "constrained,parent,children,service-type-unit-of-measures,trailer-classifications,material-types,shift-scope,calculated-constrained-amount-unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/shift-set-time-card-constraints/"+id, query)
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

	details := buildShiftSetTimeCardConstraintDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderShiftSetTimeCardConstraintDetails(cmd, details)
}

func parseShiftSetTimeCardConstraintsShowOptions(cmd *cobra.Command) (shiftSetTimeCardConstraintsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftSetTimeCardConstraintsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildShiftSetTimeCardConstraintDetails(resp jsonAPISingleResponse) shiftSetTimeCardConstraintDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := shiftSetTimeCardConstraintDetails{
		ID:                                      resource.ID,
		Name:                                    stringAttr(attrs, "name"),
		ConstrainedAmount:                       stringAttr(attrs, "constrained-amount"),
		ConstraintType:                          stringAttr(attrs, "constraint-type"),
		ConstrainedAmountType:                   stringAttr(attrs, "constrained-amount-type"),
		CurrencyCode:                            stringAttr(attrs, "currency-code"),
		Status:                                  stringAttr(attrs, "status"),
		ServiceTypeUnitOfMeasuresConstraintType: stringAttr(attrs, "service-type-unit-of-measures-constraint-type"),
		ShiftSetGroupedBy:                       stringAttr(attrs, "shift-set-grouped-by"),
		CalculatedConstrainedAmountPricePerUnit: stringAttr(attrs, "calculated-constrained-amount-price-per-unit"),
		CanDelete:                               boolAttr(attrs, "can-delete"),
	}

	if rel, ok := resource.Relationships["constrained"]; ok && rel.Data != nil {
		details.ConstrainedType = rel.Data.Type
		details.ConstrainedID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["parent"]; ok && rel.Data != nil {
		details.ParentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["children"]; ok {
		details.ChildIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["service-type-unit-of-measures"]; ok {
		details.ServiceTypeUnitOfMeasureIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["trailer-classifications"]; ok {
		details.TrailerClassificationIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["material-types"]; ok {
		details.MaterialTypeIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["shift-scope"]; ok && rel.Data != nil {
		details.ShiftScopeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["calculated-constrained-amount-unit-of-measure"]; ok && rel.Data != nil {
		details.CalculatedConstrainedAmountUnitOfMeasureID = rel.Data.ID
	}

	return details
}

func renderShiftSetTimeCardConstraintDetails(cmd *cobra.Command, details shiftSetTimeCardConstraintDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.ConstrainedType != "" && details.ConstrainedID != "" {
		fmt.Fprintf(out, "Constrained: %s/%s\n", details.ConstrainedType, details.ConstrainedID)
	}
	if details.ParentID != "" {
		fmt.Fprintf(out, "Parent ID: %s\n", details.ParentID)
	}
	if len(details.ChildIDs) > 0 {
		fmt.Fprintf(out, "Child IDs: %s\n", strings.Join(details.ChildIDs, ", "))
	}
	if details.ConstrainedAmount != "" {
		fmt.Fprintf(out, "Constrained Amount: %s\n", details.ConstrainedAmount)
	}
	if details.ConstraintType != "" {
		fmt.Fprintf(out, "Constraint Type: %s\n", details.ConstraintType)
	}
	if details.ConstrainedAmountType != "" {
		fmt.Fprintf(out, "Constrained Amount Type: %s\n", details.ConstrainedAmountType)
	}
	if details.CurrencyCode != "" {
		fmt.Fprintf(out, "Currency Code: %s\n", details.CurrencyCode)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ServiceTypeUnitOfMeasuresConstraintType != "" {
		fmt.Fprintf(out, "Service Type Unit Of Measures Constraint Type: %s\n", details.ServiceTypeUnitOfMeasuresConstraintType)
	}
	if details.ShiftSetGroupedBy != "" {
		fmt.Fprintf(out, "Shift Set Grouped By: %s\n", details.ShiftSetGroupedBy)
	}
	if details.CalculatedConstrainedAmountPricePerUnit != "" {
		fmt.Fprintf(out, "Calculated Constrained Amount Price Per Unit: %s\n", details.CalculatedConstrainedAmountPricePerUnit)
	}
	if details.CalculatedConstrainedAmountUnitOfMeasureID != "" {
		fmt.Fprintf(out, "Calculated Constrained Amount Unit Of Measure ID: %s\n", details.CalculatedConstrainedAmountUnitOfMeasureID)
	}
	if details.ShiftScopeID != "" {
		fmt.Fprintf(out, "Shift Scope ID: %s\n", details.ShiftScopeID)
	}
	if len(details.ServiceTypeUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Service Type Unit Of Measure IDs: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureIDs, ", "))
	}
	if len(details.TrailerClassificationIDs) > 0 {
		fmt.Fprintf(out, "Trailer Classification IDs: %s\n", strings.Join(details.TrailerClassificationIDs, ", "))
	}
	if len(details.MaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "Material Type IDs: %s\n", strings.Join(details.MaterialTypeIDs, ", "))
	}
	fmt.Fprintf(out, "Can Delete: %t\n", details.CanDelete)

	return nil
}
