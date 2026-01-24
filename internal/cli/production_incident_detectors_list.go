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

type productionIncidentDetectorsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type productionIncidentDetectorRow struct {
	ID                string `json:"id"`
	JobProductionPlan string `json:"job_production_plan_id,omitempty"`
	LookaheadOffset   int    `json:"lookahead_offset,omitempty"`
	MinutesThreshold  int    `json:"minutes_threshold,omitempty"`
	QuantityThreshold int    `json:"quantity_threshold,omitempty"`
	IncidentCount     int    `json:"incident_count,omitempty"`
}

func newProductionIncidentDetectorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List production incident detector runs",
		Long: `List production incident detector runs.

Output Columns:
  ID            Detector run identifier
  JOB PLAN      Job production plan ID
  LOOKAHEAD     Lookahead offset (minutes)
  MINUTES       Minutes threshold (minutes)
  QUANTITY      Quantity threshold (units)
  INCIDENTS     Number of detected incidents

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List detector runs
  xbe view production-incident-detectors list

  # Paginate results
  xbe view production-incident-detectors list --limit 25 --offset 50

  # Output as JSON
  xbe view production-incident-detectors list --json`,
		Args: cobra.NoArgs,
		RunE: runProductionIncidentDetectorsList,
	}
	initProductionIncidentDetectorsListFlags(cmd)
	return cmd
}

func init() {
	productionIncidentDetectorsCmd.AddCommand(newProductionIncidentDetectorsListCmd())
}

func initProductionIncidentDetectorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProductionIncidentDetectorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProductionIncidentDetectorsListOptions(cmd)
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
	query.Set("fields[production-incident-detectors]", "job-production-plan,lookahead-offset,minutes-threshold,quantity-threshold,incidents")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/production-incident-detectors", query)
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

	rows := buildProductionIncidentDetectorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProductionIncidentDetectorsTable(cmd, rows)
}

func parseProductionIncidentDetectorsListOptions(cmd *cobra.Command) (productionIncidentDetectorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return productionIncidentDetectorsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildProductionIncidentDetectorRows(resp jsonAPIResponse) []productionIncidentDetectorRow {
	rows := make([]productionIncidentDetectorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProductionIncidentDetectorRow(resource))
	}
	return rows
}

func productionIncidentDetectorRowFromSingle(resp jsonAPISingleResponse) productionIncidentDetectorRow {
	return buildProductionIncidentDetectorRow(resp.Data)
}

func buildProductionIncidentDetectorRow(resource jsonAPIResource) productionIncidentDetectorRow {
	attrs := resource.Attributes
	row := productionIncidentDetectorRow{
		ID:                resource.ID,
		LookaheadOffset:   intAttr(attrs, "lookahead-offset"),
		MinutesThreshold:  intAttr(attrs, "minutes-threshold"),
		QuantityThreshold: intAttr(attrs, "quantity-threshold"),
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlan = rel.Data.ID
	}
	if incidents := anyAttr(attrs, "incidents"); incidents != nil {
		row.IncidentCount = countConstraintItems(incidents)
	}
	return row
}

func renderProductionIncidentDetectorsTable(cmd *cobra.Command, rows []productionIncidentDetectorRow) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tJOB PLAN\tLOOKAHEAD\tMINUTES\tQUANTITY\tINCIDENTS")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\n",
			row.ID,
			row.JobProductionPlan,
			row.LookaheadOffset,
			row.MinutesThreshold,
			row.QuantityThreshold,
			row.IncidentCount,
		)
	}
	return w.Flush()
}
