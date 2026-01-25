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

type doMarketingMetricsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newDoMarketingMetricsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Refresh marketing metrics",
		Long: `Refresh marketing metrics.

Marketing metrics are cached aggregate counters. The create command refreshes
and returns the latest snapshot.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Refresh marketing metrics
  xbe do marketing-metrics create

  # Output as JSON
  xbe do marketing-metrics create --json`,
		Args: cobra.NoArgs,
		RunE: runDoMarketingMetricsCreate,
	}
	initDoMarketingMetricsCreateFlags(cmd)
	return cmd
}

func init() {
	doMarketingMetricsCmd.AddCommand(newDoMarketingMetricsCreateCmd())
}

func initDoMarketingMetricsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMarketingMetricsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMarketingMetricsCreateOptions(cmd)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Marketing metrics refreshed: %s\n", details.ID)
	return nil
}

func parseDoMarketingMetricsCreateOptions(cmd *cobra.Command) (doMarketingMetricsCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doMarketingMetricsCreateOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return doMarketingMetricsCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doMarketingMetricsCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doMarketingMetricsCreateOptions{}, err
	}

	return doMarketingMetricsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}
