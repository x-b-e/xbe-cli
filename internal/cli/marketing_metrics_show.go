package cli

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type marketingMetricsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMarketingMetricsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show marketing metrics details",
		Long: `Show the full marketing metrics snapshot.

Marketing metrics are cached aggregate counters. The CLI refreshes the
snapshot and returns all metrics from the server.

Output Fields:
  ID                           Snapshot identifier
  Shift Count                  Total shift count
  Driver Day Count             Total driver day count
  Tons Sum                     Total tons sum
  Driver Count                 Total driver count
  User Count                   Total user count
  Foreman Count                Total foreman count
  Material Transaction Count   Total material transaction count
  Job Production Plan Count    Total job production plan count
  Notification Count           Total notification count
  Broadcast Message Count      Total broadcast message count
  Time Card Count              Total time card count
  Invoice Count                Total invoice count
  Trucker Count                Total trucker count
  Incident Count               Total incident count
  Trip Miles Avg               Average trip miles
  Active Branch Count          Active branch count
  Transportation Cost Per Ton  Transportation cost per ton
  Night Job Pct                Percent of night jobs

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Fetch the latest marketing metrics
  xbe view marketing-metrics show

  # Output as JSON
  xbe view marketing-metrics show --json`,
		Args: cobra.NoArgs,
		RunE: runMarketingMetricsShow,
	}
	initMarketingMetricsShowFlags(cmd)
	return cmd
}

func init() {
	marketingMetricsCmd.AddCommand(newMarketingMetricsShowCmd())
}

func initMarketingMetricsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMarketingMetricsShow(cmd *cobra.Command, _ []string) error {
	opts, err := parseMarketingMetricsShowOptions(cmd)
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
	query.Set("fields[marketing-metrics]", strings.Join(marketingMetricsDetailFields(), ","))

	details, err := fetchMarketingMetrics(cmd, client, query)
	if err != nil {
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMarketingMetricsDetails(cmd, details)
}

func parseMarketingMetricsShowOptions(cmd *cobra.Command) (marketingMetricsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return marketingMetricsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return marketingMetricsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return marketingMetricsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return marketingMetricsShowOptions{}, err
	}

	return marketingMetricsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderMarketingMetricsDetails(cmd *cobra.Command, details marketingMetricsDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Shift Count: %d\n", details.ShiftCount)
	fmt.Fprintf(out, "Driver Day Count: %d\n", details.DriverDayCount)
	fmt.Fprintf(out, "Tons Sum: %s\n", formatMetricFloat(details.TonsSum))
	fmt.Fprintf(out, "Driver Count: %d\n", details.DriverCount)
	fmt.Fprintf(out, "User Count: %d\n", details.UserCount)
	fmt.Fprintf(out, "Foreman Count: %d\n", details.ForemanCount)
	fmt.Fprintf(out, "Material Transaction Count: %d\n", details.MaterialTransactionCount)
	fmt.Fprintf(out, "Job Production Plan Count: %d\n", details.JobProductionPlanCount)
	fmt.Fprintf(out, "Notification Count: %d\n", details.NotificationCount)
	fmt.Fprintf(out, "Broadcast Message Count: %d\n", details.BroadcastMessageCount)
	fmt.Fprintf(out, "Time Card Count: %d\n", details.TimeCardCount)
	fmt.Fprintf(out, "Invoice Count: %d\n", details.InvoiceCount)
	fmt.Fprintf(out, "Trucker Count: %d\n", details.TruckerCount)
	fmt.Fprintf(out, "Incident Count: %d\n", details.IncidentCount)
	fmt.Fprintf(out, "Trip Miles Avg: %s\n", formatMetricFloat(details.TripMilesAvg))
	fmt.Fprintf(out, "Active Branch Count: %d\n", details.ActiveBranchCount)
	fmt.Fprintf(out, "Transportation Cost Per Ton: %s\n", formatMetricFloat(details.TransportationCostPerTon))
	fmt.Fprintf(out, "Night Job Pct: %s\n", formatMetricFloat(details.NightJobPct))

	return nil
}

func marketingMetricsDetailFields() []string {
	return []string{
		"shift-count",
		"driver-day-count",
		"tons-sum",
		"driver-count",
		"user-count",
		"foreman-count",
		"material-transaction-count",
		"job-production-plan-count",
		"notification-count",
		"broadcast-message-count",
		"time-card-count",
		"invoice-count",
		"trucker-count",
		"incident-count",
		"trip-miles-avg",
		"active-branch-count",
		"transportation-cost-per-ton",
		"night-job-pct",
	}
}
