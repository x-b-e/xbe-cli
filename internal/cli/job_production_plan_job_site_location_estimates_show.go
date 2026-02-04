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

type jobProductionPlanJobSiteLocationEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlanJobSiteLocationEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan job site location estimate details",
		Long: `Show the full details of a job production plan job site location estimate.

Output Fields:
  ID
  Job Production Plan ID
  Estimate Count
  Estimates

Arguments:
  <id>    The job site location estimate ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job site location estimate
  xbe view job-production-plan-job-site-location-estimates show 123

  # Output as JSON
  xbe view job-production-plan-job-site-location-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanJobSiteLocationEstimatesShow,
	}
	initJobProductionPlanJobSiteLocationEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanJobSiteLocationEstimatesCmd.AddCommand(newJobProductionPlanJobSiteLocationEstimatesShowCmd())
}

func initJobProductionPlanJobSiteLocationEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanJobSiteLocationEstimatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanJobSiteLocationEstimatesShowOptions(cmd)
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
		return fmt.Errorf("job production plan job site location estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-job-site-location-estimates]", "estimates,job-production-plan")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-job-site-location-estimates/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobProductionPlanJobSiteLocationEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanJobSiteLocationEstimateDetails(cmd, details)
}

func parseJobProductionPlanJobSiteLocationEstimatesShowOptions(cmd *cobra.Command) (jobProductionPlanJobSiteLocationEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanJobSiteLocationEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobProductionPlanJobSiteLocationEstimateDetails(cmd *cobra.Command, details jobProductionPlanJobSiteLocationEstimateRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlan)
	}
	fmt.Fprintf(out, "Estimate Count: %d\n", details.EstimateCount)

	if details.Estimates != nil {
		estimates := formatJSONValue(details.Estimates)
		if estimates != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Estimates:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, estimates)
		}
	}

	return nil
}
