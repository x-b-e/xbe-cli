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

type ratesListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	ServiceTypeUnitOfMeasure string
	ParentRate               string
	Status                   string
	RatedType                string
	RatedID                  string
	NameLike                 string
	TrailerClassification    string
	MaterialType             string
	MaterialSite             string
	RateAgreement            string
}

type rateRow struct {
	ID                         string `json:"id"`
	Name                       string `json:"name,omitempty"`
	PricePerUnit               string `json:"price_per_unit,omitempty"`
	CurrencyCode               string `json:"currency_code,omitempty"`
	Status                     string `json:"status,omitempty"`
	Importance                 string `json:"importance,omitempty"`
	EffectivePricePerUnit      string `json:"effective_price_per_unit,omitempty"`
	RatedType                  string `json:"rated_type,omitempty"`
	RatedID                    string `json:"rated_id,omitempty"`
	ServiceTypeUnitOfMeasureID string `json:"service_type_unit_of_measure_id,omitempty"`
	ParentRateID               string `json:"parent_rate_id,omitempty"`
	ShiftScopeID               string `json:"shift_scope_id,omitempty"`
}

func newRatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rates",
		Long: `List rates.

Output Columns:
  ID              Rate identifier
  NAME            Rate name
  PRICE           Price per unit
  STATUS          Rate status
  RATED           Rated type and ID

Filters:
  --service-type-unit-of-measure  Filter by service type unit of measure ID
  --parent-rate                   Filter by parent rate ID
  --status                        Filter by status
  --rated-type                    Filter by rated type
  --rated-id                      Filter by rated ID
  --name-like                     Filter by name (partial match)
  --trailer-classification        Filter by trailer classification ID
  --material-type                 Filter by material type ID
  --material-site                 Filter by material site ID
  --rate-agreement                Filter by rate agreement ID`,
		Example: `  # List all rates
  xbe view rates list

  # Filter by status
  xbe view rates list --status active

  # Filter by rated type and ID
  xbe view rates list --rated-type broker-tenders --rated-id 123

  # Search by name
  xbe view rates list --name-like "hourly"

  # Output as JSON
  xbe view rates list --json`,
		RunE: runRatesList,
	}
	initRatesListFlags(cmd)
	return cmd
}

func init() {
	ratesCmd.AddCommand(newRatesListCmd())
}

func initRatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("service-type-unit-of-measure", "", "Filter by service type unit of measure ID")
	cmd.Flags().String("parent-rate", "", "Filter by parent rate ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("rated-type", "", "Filter by rated type")
	cmd.Flags().String("rated-id", "", "Filter by rated ID")
	cmd.Flags().String("name-like", "", "Filter by name (partial match)")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("rate-agreement", "", "Filter by rate agreement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRatesListOptions(cmd)
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
	query.Set("include", "rated,service-type-unit-of-measure")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[service_type_unit_of_measure]", opts.ServiceTypeUnitOfMeasure)
	setFilterIfPresent(query, "filter[parent_rate]", opts.ParentRate)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[name_like]", opts.NameLike)
	setFilterIfPresent(query, "filter[trailer_classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[rate_agreement]", opts.RateAgreement)

	// Handle polymorphic rated filter
	if opts.RatedType != "" && opts.RatedID != "" {
		query.Set("filter[rated]", opts.RatedType+"|"+opts.RatedID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/rates", query)
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

	rows := buildRateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRatesTable(cmd, rows)
}

func parseRatesListOptions(cmd *cobra.Command) (ratesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	parentRate, _ := cmd.Flags().GetString("parent-rate")
	status, _ := cmd.Flags().GetString("status")
	ratedType, _ := cmd.Flags().GetString("rated-type")
	ratedID, _ := cmd.Flags().GetString("rated-id")
	nameLike, _ := cmd.Flags().GetString("name-like")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSite, _ := cmd.Flags().GetString("material-site")
	rateAgreement, _ := cmd.Flags().GetString("rate-agreement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return ratesListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		ParentRate:               parentRate,
		Status:                   status,
		RatedType:                ratedType,
		RatedID:                  ratedID,
		NameLike:                 nameLike,
		TrailerClassification:    trailerClassification,
		MaterialType:             materialType,
		MaterialSite:             materialSite,
		RateAgreement:            rateAgreement,
	}, nil
}

func buildRateRows(resp jsonAPIResponse) []rateRow {
	rows := make([]rateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := rateRow{
			ID:                    resource.ID,
			Name:                  stringAttr(resource.Attributes, "name"),
			PricePerUnit:          stringAttr(resource.Attributes, "price-per-unit"),
			CurrencyCode:          stringAttr(resource.Attributes, "currency-code"),
			Status:                stringAttr(resource.Attributes, "status"),
			Importance:            stringAttr(resource.Attributes, "importance"),
			EffectivePricePerUnit: stringAttr(resource.Attributes, "effective-price-per-unit"),
		}

		if rel, ok := resource.Relationships["rated"]; ok && rel.Data != nil {
			row.RatedType = rel.Data.Type
			row.RatedID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
			row.ServiceTypeUnitOfMeasureID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["parent-rate"]; ok && rel.Data != nil {
			row.ParentRateID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["shift-scope"]; ok && rel.Data != nil {
			row.ShiftScopeID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderRatesTable(cmd *cobra.Command, rows []rateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tPRICE\tSTATUS\tRATED")
	for _, row := range rows {
		rated := ""
		if row.RatedType != "" && row.RatedID != "" {
			rated = row.RatedType + "/" + row.RatedID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			row.PricePerUnit,
			row.Status,
			truncateString(rated, 30),
		)
	}
	return writer.Flush()
}
