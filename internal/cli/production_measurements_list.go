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

type productionMeasurementsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	JobProductionPlanSegment string
}

type productionMeasurementRow struct {
	ID                       string `json:"id"`
	JobProductionPlanSegment string `json:"job_production_plan_segment_id,omitempty"`
	WidthInches              string `json:"width_inches,omitempty"`
	DepthInches              string `json:"depth_inches,omitempty"`
	LengthFeet               string `json:"length_feet,omitempty"`
	SpeedFeetPerMinute       string `json:"speed_feet_per_minute,omitempty"`
	DensityLbsPerCubicFoot   string `json:"density_lbs_per_cubic_foot,omitempty"`
	PassCount                string `json:"pass_count,omitempty"`
}

func newProductionMeasurementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List production measurements",
		Long: `List production measurements with filtering and pagination.

Output Columns:
  ID            Measurement identifier
  SEGMENT       Job production plan segment ID
  WIDTH_IN      Width in inches
  DEPTH_IN      Depth in inches
  LENGTH_FT     Length in feet
  SPEED_FPM     Speed in feet per minute
  DENSITY_LB_CF Density in lbs per cubic foot
  PASS          Pass count

Filters:
  --job-production-plan-segment  Filter by job production plan segment ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List production measurements
  xbe view production-measurements list

  # Filter by job production plan segment
  xbe view production-measurements list --job-production-plan-segment 123

  # Output as JSON
  xbe view production-measurements list --json`,
		Args: cobra.NoArgs,
		RunE: runProductionMeasurementsList,
	}
	initProductionMeasurementsListFlags(cmd)
	return cmd
}

func init() {
	productionMeasurementsCmd.AddCommand(newProductionMeasurementsListCmd())
}

func initProductionMeasurementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan-segment", "", "Filter by job production plan segment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProductionMeasurementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProductionMeasurementsListOptions(cmd)
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
	query.Set("fields[production-measurements]", "width-inches,depth-inches,length-feet,speed-feet-per-minute,density-lbs-per-cubic-foot,pass-count,job-production-plan-segment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan-segment]", opts.JobProductionPlanSegment)

	body, _, err := client.Get(cmd.Context(), "/v1/production-measurements", query)
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

	rows := buildProductionMeasurementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProductionMeasurementsTable(cmd, rows)
}

func parseProductionMeasurementsListOptions(cmd *cobra.Command) (productionMeasurementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanSegment, _ := cmd.Flags().GetString("job-production-plan-segment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return productionMeasurementsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		JobProductionPlanSegment: jobProductionPlanSegment,
	}, nil
}

func buildProductionMeasurementRows(resp jsonAPIResponse) []productionMeasurementRow {
	rows := make([]productionMeasurementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := productionMeasurementRow{
			ID:                     resource.ID,
			WidthInches:            stringAttr(attrs, "width-inches"),
			DepthInches:            stringAttr(attrs, "depth-inches"),
			LengthFeet:             stringAttr(attrs, "length-feet"),
			SpeedFeetPerMinute:     stringAttr(attrs, "speed-feet-per-minute"),
			DensityLbsPerCubicFoot: stringAttr(attrs, "density-lbs-per-cubic-foot"),
			PassCount:              stringAttr(attrs, "pass-count"),
		}
		row.JobProductionPlanSegment = relationshipIDFromMap(resource.Relationships, "job-production-plan-segment")
		rows = append(rows, row)
	}
	return rows
}

func productionMeasurementRowFromSingle(resp jsonAPISingleResponse) productionMeasurementRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := productionMeasurementRow{
		ID:                     resource.ID,
		WidthInches:            stringAttr(attrs, "width-inches"),
		DepthInches:            stringAttr(attrs, "depth-inches"),
		LengthFeet:             stringAttr(attrs, "length-feet"),
		SpeedFeetPerMinute:     stringAttr(attrs, "speed-feet-per-minute"),
		DensityLbsPerCubicFoot: stringAttr(attrs, "density-lbs-per-cubic-foot"),
		PassCount:              stringAttr(attrs, "pass-count"),
	}
	row.JobProductionPlanSegment = relationshipIDFromMap(resource.Relationships, "job-production-plan-segment")
	return row
}

func renderProductionMeasurementsTable(cmd *cobra.Command, rows []productionMeasurementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No production measurements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSEGMENT\tWIDTH_IN\tDEPTH_IN\tLENGTH_FT\tSPEED_FPM\tDENSITY_LB_CF\tPASS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanSegment,
			row.WidthInches,
			row.DepthInches,
			row.LengthFeet,
			row.SpeedFeetPerMinute,
			row.DensityLbsPerCubicFoot,
			row.PassCount,
		)
	}
	return writer.Flush()
}
