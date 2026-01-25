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

type unitOfMeasuresListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Name              string
	Metric            string
	MeasurementSystem string
}

type unitOfMeasureRow struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Abbreviation      string `json:"abbreviation,omitempty"`
	Metric            string `json:"metric,omitempty"`
	MeasurementSystem string `json:"measurement_system,omitempty"`
	IsCalculated      bool   `json:"is_calculated"`
	CalculationType   string `json:"calculation_type,omitempty"`
	IsQuantified      bool   `json:"is_quantified"`
}

func newUnitOfMeasuresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List units of measure",
		Long: `List units of measure with filtering and pagination.

Units of measure define how quantities are measured for billing and tracking
(e.g., tons, cubic yards, hours).

Output Columns:
  ID             Unit of measure identifier
  NAME           Unit name (e.g., Tons, Cubic Yards)
  ABBREVIATION   Short code (e.g., tn, cy)
  METRIC         Metric type (e.g., mass, volume, time)
  SYSTEM         Measurement system (e.g., imperial, metric)
  CALCULATED     Whether the unit is calculated
  CALC TYPE      Calculation type (if calculated)
  QUANTIFIED     Whether the unit is quantified

Filters:
  --name                Filter by name (partial match, case-insensitive)
  --metric              Filter by metric type
  --measurement-system  Filter by measurement system`,
		Example: `  # List all units of measure
  xbe view unit-of-measures list

  # Filter by metric type
  xbe view unit-of-measures list --metric mass

  # Filter by measurement system
  xbe view unit-of-measures list --measurement-system imperial

  # Output as JSON
  xbe view unit-of-measures list --json`,
		RunE: runUnitOfMeasuresList,
	}
	initUnitOfMeasuresListFlags(cmd)
	return cmd
}

func init() {
	unitOfMeasuresCmd.AddCommand(newUnitOfMeasuresListCmd())
}

func initUnitOfMeasuresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("metric", "", "Filter by metric type")
	cmd.Flags().String("measurement-system", "", "Filter by measurement system")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUnitOfMeasuresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUnitOfMeasuresListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation,metric,measurement-system,is-calculated,calculation-type,is-quantified")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[metric]", opts.Metric)
	setFilterIfPresent(query, "filter[measurement-system]", opts.MeasurementSystem)

	body, _, err := client.Get(cmd.Context(), "/v1/unit-of-measures", query)
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

	rows := buildUnitOfMeasureRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUnitOfMeasuresTable(cmd, rows)
}

func parseUnitOfMeasuresListOptions(cmd *cobra.Command) (unitOfMeasuresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	metric, _ := cmd.Flags().GetString("metric")
	measurementSystem, _ := cmd.Flags().GetString("measurement-system")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return unitOfMeasuresListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Name:              name,
		Metric:            metric,
		MeasurementSystem: measurementSystem,
	}, nil
}

func buildUnitOfMeasureRows(resp jsonAPIResponse) []unitOfMeasureRow {
	rows := make([]unitOfMeasureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := unitOfMeasureRow{
			ID:                resource.ID,
			Name:              stringAttr(resource.Attributes, "name"),
			Abbreviation:      stringAttr(resource.Attributes, "abbreviation"),
			Metric:            stringAttr(resource.Attributes, "metric"),
			MeasurementSystem: stringAttr(resource.Attributes, "measurement-system"),
			IsCalculated:      boolAttr(resource.Attributes, "is-calculated"),
			CalculationType:   stringAttr(resource.Attributes, "calculation-type"),
			IsQuantified:      boolAttr(resource.Attributes, "is-quantified"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderUnitOfMeasuresTable(cmd *cobra.Command, rows []unitOfMeasureRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No units of measure found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tMETRIC\tSYSTEM\tCALCULATED\tCALC TYPE\tQUANTIFIED")
	for _, row := range rows {
		calculated := "no"
		if row.IsCalculated {
			calculated = "yes"
		}
		quantified := "no"
		if row.IsQuantified {
			quantified = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Abbreviation, 10),
			truncateString(row.Metric, 15),
			truncateString(row.MeasurementSystem, 10),
			calculated,
			truncateString(row.CalculationType, 15),
			quantified,
		)
	}
	return writer.Flush()
}
