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

type jobProductionPlanCancellationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlanCancellationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan cancellation details",
		Long: `Show the full details of a job production plan cancellation.

Output Fields:
  ID
  Job Production Plan ID
  Cancellation Reason Type ID
  Comment
  Suppress Status Change Notifications

Arguments:
  <id>    The cancellation ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a cancellation
  xbe view job-production-plan-cancellations show 123

  # Output as JSON
  xbe view job-production-plan-cancellations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanCancellationsShow,
	}
	initJobProductionPlanCancellationsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanCancellationsCmd.AddCommand(newJobProductionPlanCancellationsShowCmd())
}

func initJobProductionPlanCancellationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanCancellationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanCancellationsShowOptions(cmd)
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
		return fmt.Errorf("job production plan cancellation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-cancellations]", "comment,suppress-status-change-notifications,job-production-plan,job-production-plan-cancellation-reason-type")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-cancellations/"+id, query)
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

	details := buildJobProductionPlanCancellationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanCancellationDetails(cmd, details)
}

func parseJobProductionPlanCancellationsShowOptions(cmd *cobra.Command) (jobProductionPlanCancellationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanCancellationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobProductionPlanCancellationDetails(cmd *cobra.Command, details jobProductionPlanCancellationRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.CancellationReasonTypeID != "" {
		fmt.Fprintf(out, "Cancellation Reason Type ID: %s\n", details.CancellationReasonTypeID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	fmt.Fprintf(out, "Suppress Status Change Notifications: %t\n", details.SuppressStatusChangeNotifications)

	return nil
}
