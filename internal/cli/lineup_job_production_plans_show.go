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

type lineupJobProductionPlansShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupJobProductionPlanDetails struct {
	ID                  string `json:"id"`
	LineupID            string `json:"lineup_id,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	IsDeletable         bool   `json:"is_deletable"`
}

func newLineupJobProductionPlansShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup job production plan details",
		Long: `Show the full details of a lineup job production plan.

Output Fields:
  ID                  Lineup job production plan identifier
  Lineup              Lineup ID
  Job Production Plan Job production plan ID
  Deletable           Whether the plan can be deleted

Arguments:
  <id>    The lineup job production plan ID (required). You can find IDs using the list command.`,
		Example: `  # Show a lineup job production plan
  xbe view lineup-job-production-plans show 123

  # Get JSON output
  xbe view lineup-job-production-plans show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupJobProductionPlansShow,
	}
	initLineupJobProductionPlansShowFlags(cmd)
	return cmd
}

func init() {
	lineupJobProductionPlansCmd.AddCommand(newLineupJobProductionPlansShowCmd())
}

func initLineupJobProductionPlansShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupJobProductionPlansShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupJobProductionPlansShowOptions(cmd)
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
		return fmt.Errorf("lineup job production plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-job-production-plans/"+id, query)
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

	details := buildLineupJobProductionPlanDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupJobProductionPlanDetails(cmd, details)
}

func parseLineupJobProductionPlansShowOptions(cmd *cobra.Command) (lineupJobProductionPlansShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupJobProductionPlansShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupJobProductionPlanDetails(resp jsonAPISingleResponse) lineupJobProductionPlanDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := lineupJobProductionPlanDetails{
		ID:          resource.ID,
		IsDeletable: boolAttr(attrs, "is-deletable"),
	}

	if rel, ok := resource.Relationships["lineup"]; ok && rel.Data != nil {
		details.LineupID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}

	return details
}

func renderLineupJobProductionPlanDetails(cmd *cobra.Command, details lineupJobProductionPlanDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupID != "" {
		fmt.Fprintf(out, "Lineup: %s\n", details.LineupID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	deletable := "no"
	if details.IsDeletable {
		deletable = "yes"
	}
	fmt.Fprintf(out, "Deletable: %s\n", deletable)

	return nil
}
