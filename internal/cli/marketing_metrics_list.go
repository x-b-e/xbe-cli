package cli

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type marketingMetricsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMarketingMetricsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List marketing metrics",
		Long: `Fetch the latest marketing metrics snapshot.

Marketing metrics are cached aggregate counters. The CLI issues a request to
refresh the snapshot and returns the current values.

Output Columns:
  ID                Snapshot identifier
  SHIFTS            Total shift count
  DRIVER DAYS       Total driver day count
  TONS              Total tons sum
  DRIVERS           Total driver count
  USERS             Total user count
  TRUCKERS          Total trucker count
  INCIDENTS         Total incident count
  ACTIVE BRANCHES   Active branch count

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Fetch the latest marketing metrics
  xbe view marketing-metrics list

  # Output as JSON
  xbe view marketing-metrics list --json`,
		Args: cobra.NoArgs,
		RunE: runMarketingMetricsList,
	}
	initMarketingMetricsListFlags(cmd)
	return cmd
}

func init() {
	marketingMetricsCmd.AddCommand(newMarketingMetricsListCmd())
}

func initMarketingMetricsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMarketingMetricsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMarketingMetricsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)
	query := url.Values{}
	query.Set("fields[marketing-metrics]", strings.Join(marketingMetricsFields(), ","))

	details, err := fetchMarketingMetrics(cmd, client, query)
	if err != nil {
		return err
	}

	row := buildMarketingMetricsListRow(details)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), []marketingMetricsListRow{row})
	}

	return renderMarketingMetricsTable(cmd, []marketingMetricsListRow{row})
}

func parseMarketingMetricsListOptions(cmd *cobra.Command) (marketingMetricsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return marketingMetricsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return marketingMetricsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return marketingMetricsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return marketingMetricsListOptions{}, err
	}

	return marketingMetricsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderMarketingMetricsTable(cmd *cobra.Command, rows []marketingMetricsListRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No marketing metrics found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFTS\tDRIVER DAYS\tTONS\tDRIVERS\tUSERS\tTRUCKERS\tINCIDENTS\tACTIVE BRANCHES")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%d\t%d\t%s\t%d\t%d\t%d\t%d\t%d\n",
			row.ID,
			row.ShiftCount,
			row.DriverDayCount,
			formatMetricFloat(row.TonsSum),
			row.DriverCount,
			row.UserCount,
			row.TruckerCount,
			row.IncidentCount,
			row.ActiveBranchCount,
		)
	}
	return writer.Flush()
}

func marketingMetricsFields() []string {
	return []string{
		"shift-count",
		"driver-day-count",
		"tons-sum",
		"driver-count",
		"user-count",
		"trucker-count",
		"incident-count",
		"active-branch-count",
	}
}

func formatMetricFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}
