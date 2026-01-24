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

type missingRatesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type missingRateRow struct {
	ID                         string `json:"id"`
	JobID                      string `json:"job_id,omitempty"`
	ServiceTypeUnitOfMeasureID string `json:"service_type_unit_of_measure_id,omitempty"`
	CurrencyCode               string `json:"currency_code,omitempty"`
	CustomerPricePerUnit       string `json:"customer_price_per_unit,omitempty"`
	TruckerPricePerUnit        string `json:"trucker_price_per_unit,omitempty"`
}

func newMissingRatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List missing rates",
		Long: `List missing rates with filtering and pagination.

Missing rates are created to add rates to a job when a service type unit of
measure is missing. Each record ties a job to a service type unit of measure
and includes customer and trucker price per unit values.

Output Columns:
  ID             Missing rate identifier
  JOB ID         Job ID
  STUOM ID       Service type unit of measure ID
  CUSTOMER PPU   Customer price per unit
  TRUCKER PPU    Trucker price per unit
  CURRENCY       Currency code

Filters:
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --is-created-at    Filter by has created-at (true/false)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)
  --is-updated-at    Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List missing rates
  xbe view missing-rates list

  # Filter by created-at window
  xbe view missing-rates list --created-at-min 2024-01-01T00:00:00Z --created-at-max 2024-12-31T23:59:59Z

  # Output as JSON
  xbe view missing-rates list --json`,
		Args: cobra.NoArgs,
		RunE: runMissingRatesList,
	}
	initMissingRatesListFlags(cmd)
	return cmd
}

func init() {
	missingRatesCmd.AddCommand(newMissingRatesListCmd())
}

func initMissingRatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMissingRatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMissingRatesListOptions(cmd)
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
	query.Set("fields[missing-rates]", "job,service-type-unit-of-measure,currency-code,customer-price-per-unit,trucker-price-per-unit")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/missing-rates", query)
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

	rows := buildMissingRateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMissingRatesTable(cmd, rows)
}

func parseMissingRatesListOptions(cmd *cobra.Command) (missingRatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return missingRatesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildMissingRateRows(resp jsonAPIResponse) []missingRateRow {
	rows := make([]missingRateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMissingRateRow(resource))
	}
	return rows
}

func buildMissingRateRow(resource jsonAPIResource) missingRateRow {
	row := missingRateRow{
		ID:                   resource.ID,
		CurrencyCode:         stringAttr(resource.Attributes, "currency-code"),
		CustomerPricePerUnit: stringAttr(resource.Attributes, "customer-price-per-unit"),
		TruckerPricePerUnit:  stringAttr(resource.Attributes, "trucker-price-per-unit"),
	}

	if rel, ok := resource.Relationships["job"]; ok && rel.Data != nil {
		row.JobID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
		row.ServiceTypeUnitOfMeasureID = rel.Data.ID
	}

	return row
}

func missingRateRowFromSingle(resp jsonAPISingleResponse) missingRateRow {
	return buildMissingRateRow(resp.Data)
}

func renderMissingRatesTable(cmd *cobra.Command, rows []missingRateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No missing rates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB ID\tSTUOM ID\tCUSTOMER PPU\tTRUCKER PPU\tCURRENCY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobID,
			row.ServiceTypeUnitOfMeasureID,
			row.CustomerPricePerUnit,
			row.TruckerPricePerUnit,
			row.CurrencyCode,
		)
	}
	return writer.Flush()
}
