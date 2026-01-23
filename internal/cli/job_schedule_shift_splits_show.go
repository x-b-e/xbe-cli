package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobScheduleShiftSplitsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobScheduleShiftSplitsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job schedule shift split details",
		Long: `Show the full details of a job schedule shift split.

Output Fields:
  ID
  Job Schedule Shift ID
  New Job Schedule Shift ID
  Expected Material Transaction Count
  Expected Material Transaction Tons
  New Start At

Arguments:
  <id>    The shift split ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a shift split
  xbe view job-schedule-shift-splits show 123

  # Output as JSON
  xbe view job-schedule-shift-splits show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobScheduleShiftSplitsShow,
	}
	initJobScheduleShiftSplitsShowFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftSplitsCmd.AddCommand(newJobScheduleShiftSplitsShowCmd())
}

func initJobScheduleShiftSplitsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftSplitsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobScheduleShiftSplitsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job schedule shift split id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-schedule-shift-splits]", "expected-material-transaction-count,expected-material-transaction-tons,new-start-at,job-schedule-shift,new-job-schedule-shift")

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shift-splits/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildJobScheduleShiftSplitRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobScheduleShiftSplitDetails(cmd, details)
}

func parseJobScheduleShiftSplitsShowOptions(cmd *cobra.Command) (jobScheduleShiftSplitsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobScheduleShiftSplitsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobScheduleShiftSplitDetails(cmd *cobra.Command, details jobScheduleShiftSplitRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobScheduleShiftID != "" {
		fmt.Fprintf(out, "Job Schedule Shift ID: %s\n", details.JobScheduleShiftID)
	}
	if details.NewJobScheduleShiftID != "" {
		fmt.Fprintf(out, "New Job Schedule Shift ID: %s\n", details.NewJobScheduleShiftID)
	}
	if details.ExpectedMaterialTransactionCount != 0 {
		fmt.Fprintf(out, "Expected Material Transaction Count: %s\n", formatOptionalInt(details.ExpectedMaterialTransactionCount))
	}
	if details.ExpectedMaterialTransactionTons != 0 {
		fmt.Fprintf(out, "Expected Material Transaction Tons: %s\n", formatOptionalFloat(details.ExpectedMaterialTransactionTons))
	}
	if details.NewStartAt != "" {
		fmt.Fprintf(out, "New Start At: %s\n", details.NewStartAt)
	}

	return nil
}
