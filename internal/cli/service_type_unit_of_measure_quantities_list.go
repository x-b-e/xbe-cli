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

type serviceTypeUnitOfMeasureQuantitiesListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	ServiceTypeUnitOfMeasure string
	Quantity                 string
	CalculatedQuantity       string
	ExplicitQuantity         string
	Broker                   string
	QuantifiesType           string
	QuantifiesID             string
	QuantifiesStartOnMin     string
	QuantifiesStartOnMax     string
}

type serviceTypeUnitOfMeasureQuantityRow struct {
	ID                           string `json:"id"`
	Quantity                     string `json:"quantity,omitempty"`
	ExplicitQuantity             string `json:"explicit_quantity,omitempty"`
	CalculatedQuantity           string `json:"calculated_quantity,omitempty"`
	ServiceTypeUnitOfMeasureID   string `json:"service_type_unit_of_measure_id,omitempty"`
	ServiceTypeUnitOfMeasureName string `json:"service_type_unit_of_measure_name,omitempty"`
	QuantifiesType               string `json:"quantifies_type,omitempty"`
	QuantifiesID                 string `json:"quantifies_id,omitempty"`
	MaterialTypeID               string `json:"material_type_id,omitempty"`
	MaterialTypeName             string `json:"material_type_name,omitempty"`
	TrailerClassificationID      string `json:"trailer_classification_id,omitempty"`
	TrailerClassificationName    string `json:"trailer_classification_name,omitempty"`
}

func newServiceTypeUnitOfMeasureQuantitiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service type unit of measure quantities",
		Long: `List service type unit of measure quantities with filtering and pagination.

Output Columns:
  ID             Quantity identifier
  QUANTITY       Quantity value
  EXPLICIT       Explicit quantity (if provided)
  CALCULATED     Calculated quantity (if available)
  SERVICE TYPE UOM  Service type/unit of measure name or ID
  QUANTIFIES     Quantified resource (type/id)
  MATERIAL TYPE  Material type name or ID
  TRAILER CLASS  Trailer classification name or ID

Filters:
  --service-type-unit-of-measure  Filter by service type unit of measure ID
  --quantity                      Filter by quantity
  --calculated-quantity            Filter by calculated quantity
  --explicit-quantity              Filter by explicit quantity
  --broker                        Filter by broker ID
  --quantifies-type               Filter by quantified resource type
  --quantifies-id                 Filter by quantified resource ID (requires --quantifies-type)
  --quantifies-start-on-min       Filter by quantifies start date on or after (YYYY-MM-DD)
  --quantifies-start-on-max       Filter by quantifies start date on or before (YYYY-MM-DD)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List service type unit of measure quantities
  xbe view service-type-unit-of-measure-quantities list

  # Filter by service type unit of measure
  xbe view service-type-unit-of-measure-quantities list --service-type-unit-of-measure 123

  # Filter by quantified time card
  xbe view service-type-unit-of-measure-quantities list --quantifies-type time-cards --quantifies-id 456

  # Filter by quantity range on quantifies start date
  xbe view service-type-unit-of-measure-quantities list --quantifies-start-on-min 2024-01-01 --quantifies-start-on-max 2024-12-31

  # Output as JSON
  xbe view service-type-unit-of-measure-quantities list --json`,
		Args: cobra.NoArgs,
		RunE: runServiceTypeUnitOfMeasureQuantitiesList,
	}
	initServiceTypeUnitOfMeasureQuantitiesListFlags(cmd)
	return cmd
}

func init() {
	serviceTypeUnitOfMeasureQuantitiesCmd.AddCommand(newServiceTypeUnitOfMeasureQuantitiesListCmd())
}

func initServiceTypeUnitOfMeasureQuantitiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("service-type-unit-of-measure", "", "Filter by service type unit of measure ID")
	cmd.Flags().String("quantity", "", "Filter by quantity")
	cmd.Flags().String("calculated-quantity", "", "Filter by calculated quantity")
	cmd.Flags().String("explicit-quantity", "", "Filter by explicit quantity")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("quantifies-type", "", "Filter by quantified resource type")
	cmd.Flags().String("quantifies-id", "", "Filter by quantified resource ID")
	cmd.Flags().String("quantifies-start-on-min", "", "Filter by quantifies start date on or after (YYYY-MM-DD)")
	cmd.Flags().String("quantifies-start-on-max", "", "Filter by quantifies start date on or before (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypeUnitOfMeasureQuantitiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseServiceTypeUnitOfMeasureQuantitiesListOptions(cmd)
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
	query.Set("fields[service-type-unit-of-measure-quantities]", "quantity,explicit-quantity,calculated-quantity,service-type-unit-of-measure,quantifies,material-type,trailer-classification")
	query.Set("include", "service-type-unit-of-measure,material-type,trailer-classification")
	query.Set("fields[service-type-unit-of-measures]", "name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[trailer-classifications]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "id")
	}

	setFilterIfPresent(query, "filter[service_type_unit_of_measure]", opts.ServiceTypeUnitOfMeasure)
	setFilterIfPresent(query, "filter[quantity]", opts.Quantity)
	setFilterIfPresent(query, "filter[calculated_quantity]", opts.CalculatedQuantity)
	setFilterIfPresent(query, "filter[explicit_quantity]", opts.ExplicitQuantity)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[quantifies_start_on_min]", opts.QuantifiesStartOnMin)
	setFilterIfPresent(query, "filter[quantifies_start_on_max]", opts.QuantifiesStartOnMax)

	quantifiesType := normalizePolymorphicType(opts.QuantifiesType)
	if quantifiesType != "" && opts.QuantifiesID != "" {
		query.Set("filter[quantifies]", quantifiesType+"|"+opts.QuantifiesID)
	} else if quantifiesType != "" {
		query.Set("filter[quantifies_type]", quantifiesType)
	} else if opts.QuantifiesID != "" {
		return fmt.Errorf("--quantifies-id requires --quantifies-type")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/service-type-unit-of-measure-quantities", query)
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

	rows := buildServiceTypeUnitOfMeasureQuantityRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderServiceTypeUnitOfMeasureQuantitiesTable(cmd, rows)
}

func parseServiceTypeUnitOfMeasureQuantitiesListOptions(cmd *cobra.Command) (serviceTypeUnitOfMeasureQuantitiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	calculatedQuantity, _ := cmd.Flags().GetString("calculated-quantity")
	explicitQuantity, _ := cmd.Flags().GetString("explicit-quantity")
	broker, _ := cmd.Flags().GetString("broker")
	quantifiesType, _ := cmd.Flags().GetString("quantifies-type")
	quantifiesID, _ := cmd.Flags().GetString("quantifies-id")
	quantifiesStartOnMin, _ := cmd.Flags().GetString("quantifies-start-on-min")
	quantifiesStartOnMax, _ := cmd.Flags().GetString("quantifies-start-on-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypeUnitOfMeasureQuantitiesListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		Quantity:                 quantity,
		CalculatedQuantity:       calculatedQuantity,
		ExplicitQuantity:         explicitQuantity,
		Broker:                   broker,
		QuantifiesType:           quantifiesType,
		QuantifiesID:             quantifiesID,
		QuantifiesStartOnMin:     quantifiesStartOnMin,
		QuantifiesStartOnMax:     quantifiesStartOnMax,
	}, nil
}

func buildServiceTypeUnitOfMeasureQuantityRows(resp jsonAPIResponse) []serviceTypeUnitOfMeasureQuantityRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]serviceTypeUnitOfMeasureQuantityRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := serviceTypeUnitOfMeasureQuantityRow{
			ID:                 resource.ID,
			Quantity:           stringAttr(attrs, "quantity"),
			ExplicitQuantity:   stringAttr(attrs, "explicit-quantity"),
			CalculatedQuantity: stringAttr(attrs, "calculated-quantity"),
		}

		if rel, ok := resource.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
			row.ServiceTypeUnitOfMeasureID = rel.Data.ID
			if stuom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ServiceTypeUnitOfMeasureName = strings.TrimSpace(stringAttr(stuom.Attributes, "name"))
			}
		}

		if rel, ok := resource.Relationships["quantifies"]; ok && rel.Data != nil {
			row.QuantifiesType = rel.Data.Type
			row.QuantifiesID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialTypeName = materialTypeLabel(materialType.Attributes)
			}
		}

		if rel, ok := resource.Relationships["trailer-classification"]; ok && rel.Data != nil {
			row.TrailerClassificationID = rel.Data.ID
			if trailer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TrailerClassificationName = trailerClassificationLabel(trailer.Attributes)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderServiceTypeUnitOfMeasureQuantitiesTable(cmd *cobra.Command, rows []serviceTypeUnitOfMeasureQuantityRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No service type unit of measure quantities found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tQUANTITY\tEXPLICIT\tCALCULATED\tSERVICE TYPE UOM\tQUANTIFIES\tMATERIAL TYPE\tTRAILER CLASS")
	for _, row := range rows {
		stuomLabel := row.ServiceTypeUnitOfMeasureName
		if stuomLabel == "" {
			stuomLabel = row.ServiceTypeUnitOfMeasureID
		}
		quantifies := ""
		if row.QuantifiesType != "" && row.QuantifiesID != "" {
			quantifies = row.QuantifiesType + "/" + row.QuantifiesID
		}
		materialLabel := row.MaterialTypeName
		if materialLabel == "" {
			materialLabel = row.MaterialTypeID
		}
		trailerLabel := row.TrailerClassificationName
		if trailerLabel == "" {
			trailerLabel = row.TrailerClassificationID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Quantity, 12),
			truncateString(row.ExplicitQuantity, 12),
			truncateString(row.CalculatedQuantity, 12),
			truncateString(stuomLabel, 24),
			truncateString(quantifies, 24),
			truncateString(materialLabel, 20),
			truncateString(trailerLabel, 20),
		)
	}
	return writer.Flush()
}

func trailerClassificationLabel(attrs map[string]any) string {
	abbrev := strings.TrimSpace(stringAttr(attrs, "abbreviation"))
	if abbrev != "" {
		return abbrev
	}
	return strings.TrimSpace(stringAttr(attrs, "name"))
}

func normalizePolymorphicType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, "-_") {
		parts := strings.FieldsFunc(value, func(r rune) bool {
			return r == '-' || r == '_'
		})
		if len(parts) == 0 {
			return value
		}
		lastIdx := len(parts) - 1
		if strings.HasSuffix(parts[lastIdx], "s") {
			parts[lastIdx] = strings.TrimSuffix(parts[lastIdx], "s")
		}
		for i, part := range parts {
			if part == "" {
				continue
			}
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
		return strings.Join(parts, "")
	}
	return value
}
