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

type jobProductionPlanUnabandonmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlanUnabandonmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan unabandonment details",
		Long: `Show the full details of a job production plan unabandonment.

Output Fields:
  ID
  Job Production Plan ID
  Comment
  Suppress Status Change Notifications

Arguments:
  <id>    The unabandonment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an unabandonment
  xbe view job-production-plan-unabandonments show 123

  # Output as JSON
  xbe view job-production-plan-unabandonments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanUnabandonmentsShow,
	}
	initJobProductionPlanUnabandonmentsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanUnabandonmentsCmd.AddCommand(newJobProductionPlanUnabandonmentsShowCmd())
}

func initJobProductionPlanUnabandonmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanUnabandonmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanUnabandonmentsShowOptions(cmd)
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
		return fmt.Errorf("job production plan unabandonment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-unabandonments]", "comment,suppress-status-change-notifications,job-production-plan")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-unabandonments/"+id, query)
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

	details := buildJobProductionPlanUnabandonmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanUnabandonmentDetails(cmd, details)
}

func parseJobProductionPlanUnabandonmentsShowOptions(cmd *cobra.Command) (jobProductionPlanUnabandonmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanUnabandonmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobProductionPlanUnabandonmentDetails(cmd *cobra.Command, details jobProductionPlanUnabandonmentRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	fmt.Fprintf(out, "Suppress Status Change Notifications: %t\n", details.SuppressStatusChangeNotifications)

	return nil
}
