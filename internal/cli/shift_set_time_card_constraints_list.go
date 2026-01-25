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

type shiftSetTimeCardConstraintsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	ConstrainedModelType      string
	ConstrainedModelID        string
	ConstrainedAmount         string
	ConstraintType            string
	Status                    string
	CurrencyCode              string
	Name                      string
	ShiftScope                string
	Search                    string
	ServiceTypeUnitOfMeasures string
	RateAgreement             string
	TrailerClassification     string
	MaterialType              string
	ScopedToShift             string
	ScopedToTender            string
}

type shiftSetTimeCardConstraintRow struct {
	ID                                         string `json:"id"`
	Name                                       string `json:"name,omitempty"`
	ConstrainedType                            string `json:"constrained_type,omitempty"`
	ConstrainedID                              string `json:"constrained_id,omitempty"`
	ConstraintType                             string `json:"constraint_type,omitempty"`
	ConstrainedAmount                          string `json:"constrained_amount,omitempty"`
	ConstrainedAmountType                      string `json:"constrained_amount_type,omitempty"`
	CurrencyCode                               string `json:"currency_code,omitempty"`
	Status                                     string `json:"status,omitempty"`
	ShiftScopeID                               string `json:"shift_scope_id,omitempty"`
	CalculatedConstrainedAmountPricePerUnit    string `json:"calculated_constrained_amount_price_per_unit,omitempty"`
	CalculatedConstrainedAmountUnitOfMeasureID string `json:"calculated_constrained_amount_unit_of_measure_id,omitempty"`
}

func newShiftSetTimeCardConstraintsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shift set time card constraints",
		Long: `List shift set time card constraints.

Output Columns:
  ID            Constraint identifier
  NAME          Constraint name
  CONSTRAINED   Constrained record type and ID
  TYPE          Constraint type
  AMOUNT        Constrained amount (or calculated rate)
  AMOUNT TYPE   Constrained amount type
  STATUS        Status
  SHIFT SCOPE   Shift scope ID

Filters:
  --constrained-model-type            Filter by constrained type (e.g., tenders, rate-agreements)
  --constrained-model-id              Filter by constrained ID (requires --constrained-model-type)
  --constrained-amount                Filter by constrained amount
  --constraint-type                   Filter by constraint type (minimum, equality, maximum)
  --status                            Filter by status (active, inactive)
  --currency-code                     Filter by currency code (USD)
  --name                              Filter by name
  --shift-scope                        Filter by shift scope ID
  --search                            Search by name
  --service-type-unit-of-measures     Filter by service type unit of measure IDs (comma-separated)
  --rate-agreement                     Filter by rate agreement ID
  --trailer-classification            Filter by trailer classification ID
  --material-type                     Filter by material type ID
  --scoped-to-shift                   Filter by tender job schedule shift ID
  --scoped-to-tender                  Filter by tender ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List constraints
  xbe view shift-set-time-card-constraints list

  # Filter by constrained record
  xbe view shift-set-time-card-constraints list --constrained-model-type rate-agreements --constrained-model-id 123

  # Filter by constraint type and status
  xbe view shift-set-time-card-constraints list --constraint-type minimum --status active

  # Filter by shift scope
  xbe view shift-set-time-card-constraints list --shift-scope 456

  # Search by name
  xbe view shift-set-time-card-constraints list --search "minimum pay"

  # Output as JSON
  xbe view shift-set-time-card-constraints list --json`,
		Args: cobra.NoArgs,
		RunE: runShiftSetTimeCardConstraintsList,
	}
	initShiftSetTimeCardConstraintsListFlags(cmd)
	return cmd
}

func init() {
	shiftSetTimeCardConstraintsCmd.AddCommand(newShiftSetTimeCardConstraintsListCmd())
}

func initShiftSetTimeCardConstraintsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("constrained-model-type", "", "Filter by constrained type (e.g., tenders, rate-agreements)")
	cmd.Flags().String("constrained-model-id", "", "Filter by constrained ID (requires --constrained-model-type)")
	cmd.Flags().String("constrained-amount", "", "Filter by constrained amount")
	cmd.Flags().String("constraint-type", "", "Filter by constraint type (minimum, equality, maximum)")
	cmd.Flags().String("status", "", "Filter by status (active, inactive)")
	cmd.Flags().String("currency-code", "", "Filter by currency code (USD)")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("shift-scope", "", "Filter by shift scope ID")
	cmd.Flags().String("search", "", "Search by name")
	cmd.Flags().String("service-type-unit-of-measures", "", "Filter by service type unit of measure IDs (comma-separated)")
	cmd.Flags().String("rate-agreement", "", "Filter by rate agreement ID")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("scoped-to-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("scoped-to-tender", "", "Filter by tender ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftSetTimeCardConstraintsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseShiftSetTimeCardConstraintsListOptions(cmd)
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

	if opts.ConstrainedModelID != "" && opts.ConstrainedModelType == "" {
		err := fmt.Errorf("--constrained-model-type is required when --constrained-model-id is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[shift-set-time-card-constraints]", "name,constrained-amount,constraint-type,constrained-amount-type,currency-code,status,constrained,shift-scope,calculated-constrained-amount-price-per-unit,calculated-constrained-amount-unit-of-measure")
	query.Set("include", "constrained,shift-scope,calculated-constrained-amount-unit-of-measure")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.ConstrainedModelType != "" {
		normalizedType := normalizeConstrainedModelType(opts.ConstrainedModelType)
		if opts.ConstrainedModelID != "" {
			query.Set("filter[constrained_model]", normalizedType+"|"+opts.ConstrainedModelID)
		} else {
			query.Set("filter[constrained_model_type]", normalizedType)
		}
	}
	setFilterIfPresent(query, "filter[constrained_amount]", opts.ConstrainedAmount)
	setFilterIfPresent(query, "filter[constraint_type]", opts.ConstraintType)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[currency_code]", opts.CurrencyCode)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[shift_scope]", opts.ShiftScope)
	setFilterIfPresent(query, "filter[q]", opts.Search)
	setFilterIfPresent(query, "filter[service_type_unit_of_measures]", opts.ServiceTypeUnitOfMeasures)
	setFilterIfPresent(query, "filter[rate_agreement]", opts.RateAgreement)
	setFilterIfPresent(query, "filter[trailer_classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[scoped_to_shift]", opts.ScopedToShift)
	setFilterIfPresent(query, "filter[scoped_to_tender]", opts.ScopedToTender)

	body, _, err := client.Get(cmd.Context(), "/v1/shift-set-time-card-constraints", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildShiftSetTimeCardConstraintRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderShiftSetTimeCardConstraintsTable(cmd, rows)
}

func parseShiftSetTimeCardConstraintsListOptions(cmd *cobra.Command) (shiftSetTimeCardConstraintsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	constrainedModelType, _ := cmd.Flags().GetString("constrained-model-type")
	constrainedModelID, _ := cmd.Flags().GetString("constrained-model-id")
	constrainedAmount, _ := cmd.Flags().GetString("constrained-amount")
	constraintType, _ := cmd.Flags().GetString("constraint-type")
	status, _ := cmd.Flags().GetString("status")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	name, _ := cmd.Flags().GetString("name")
	shiftScope, _ := cmd.Flags().GetString("shift-scope")
	search, _ := cmd.Flags().GetString("search")
	serviceTypeUnitOfMeasures, _ := cmd.Flags().GetString("service-type-unit-of-measures")
	rateAgreement, _ := cmd.Flags().GetString("rate-agreement")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	materialType, _ := cmd.Flags().GetString("material-type")
	scopedToShift, _ := cmd.Flags().GetString("scoped-to-shift")
	scopedToTender, _ := cmd.Flags().GetString("scoped-to-tender")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftSetTimeCardConstraintsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		ConstrainedModelType:      constrainedModelType,
		ConstrainedModelID:        constrainedModelID,
		ConstrainedAmount:         constrainedAmount,
		ConstraintType:            constraintType,
		Status:                    status,
		CurrencyCode:              currencyCode,
		Name:                      name,
		ShiftScope:                shiftScope,
		Search:                    search,
		ServiceTypeUnitOfMeasures: serviceTypeUnitOfMeasures,
		RateAgreement:             rateAgreement,
		TrailerClassification:     trailerClassification,
		MaterialType:              materialType,
		ScopedToShift:             scopedToShift,
		ScopedToTender:            scopedToTender,
	}, nil
}

func buildShiftSetTimeCardConstraintRows(resp jsonAPIResponse) []shiftSetTimeCardConstraintRow {
	rows := make([]shiftSetTimeCardConstraintRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildShiftSetTimeCardConstraintRow(resource))
	}
	return rows
}

func buildShiftSetTimeCardConstraintRow(resource jsonAPIResource) shiftSetTimeCardConstraintRow {
	attrs := resource.Attributes
	row := shiftSetTimeCardConstraintRow{
		ID:                                      resource.ID,
		Name:                                    stringAttr(attrs, "name"),
		ConstraintType:                          stringAttr(attrs, "constraint-type"),
		ConstrainedAmount:                       stringAttr(attrs, "constrained-amount"),
		ConstrainedAmountType:                   stringAttr(attrs, "constrained-amount-type"),
		CurrencyCode:                            stringAttr(attrs, "currency-code"),
		Status:                                  stringAttr(attrs, "status"),
		CalculatedConstrainedAmountPricePerUnit: stringAttr(attrs, "calculated-constrained-amount-price-per-unit"),
	}

	if rel, ok := resource.Relationships["constrained"]; ok && rel.Data != nil {
		row.ConstrainedType = rel.Data.Type
		row.ConstrainedID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["shift-scope"]; ok && rel.Data != nil {
		row.ShiftScopeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["calculated-constrained-amount-unit-of-measure"]; ok && rel.Data != nil {
		row.CalculatedConstrainedAmountUnitOfMeasureID = rel.Data.ID
	}

	return row
}

func buildShiftSetTimeCardConstraintRowFromSingle(resp jsonAPISingleResponse) shiftSetTimeCardConstraintRow {
	return buildShiftSetTimeCardConstraintRow(resp.Data)
}

func renderShiftSetTimeCardConstraintsTable(cmd *cobra.Command, rows []shiftSetTimeCardConstraintRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift set time card constraints found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tCONSTRAINED\tTYPE\tAMOUNT\tAMOUNT TYPE\tSTATUS\tSHIFT SCOPE")
	for _, row := range rows {
		constrained := ""
		if row.ConstrainedType != "" && row.ConstrainedID != "" {
			constrained = row.ConstrainedType + "/" + row.ConstrainedID
		}

		amount := ""
		if row.ConstrainedAmount != "" {
			amount = row.ConstrainedAmount
			if row.CurrencyCode != "" {
				amount = amount + " " + row.CurrencyCode
			}
		} else if row.CalculatedConstrainedAmountPricePerUnit != "" {
			amount = "calc " + row.CalculatedConstrainedAmountPricePerUnit
			if row.CalculatedConstrainedAmountUnitOfMeasureID != "" {
				amount = amount + "/" + row.CalculatedConstrainedAmountUnitOfMeasureID
			}
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Name,
			constrained,
			row.ConstraintType,
			amount,
			row.ConstrainedAmountType,
			row.Status,
			row.ShiftScopeID,
		)
	}
	writer.Flush()
	return nil
}

func normalizeConstrainedModelType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed
	}
	switch strings.ToLower(trimmed) {
	case "tender", "tenders":
		return "Tender"
	case "rate-agreement", "rate-agreements", "rate_agreement", "rateagreement":
		return "RateAgreement"
	default:
		return trimmed
	}
}
