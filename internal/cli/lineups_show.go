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

type lineupsShowOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	NoAuth                             bool
	IncludeTruckerAssignmentStatistics bool
}

type lineupDetails struct {
	ID                                    string   `json:"id"`
	Name                                  string   `json:"name,omitempty"`
	StartAtMin                            string   `json:"start_at_min,omitempty"`
	StartAtMax                            string   `json:"start_at_max,omitempty"`
	CustomerID                            string   `json:"customer_id,omitempty"`
	LineupJobProductionPlanIDs            []string `json:"lineup_job_production_plan_ids,omitempty"`
	AliveLineupJobProductionPlanIDs       []string `json:"alive_lineup_job_production_plan_ids,omitempty"`
	UnabandonedLineupJobProductionPlanIDs []string `json:"unabandoned_lineup_job_production_plan_ids,omitempty"`
	JobProductionPlanIDs                  []string `json:"job_production_plan_ids,omitempty"`
	AliveJobProductionPlanIDs             []string `json:"alive_job_production_plan_ids,omitempty"`
	UnabandonedJobProductionPlanIDs       []string `json:"unabandoned_job_production_plan_ids,omitempty"`
	LineupJobScheduleShiftIDs             []string `json:"lineup_job_schedule_shift_ids,omitempty"`
	JobScheduleShiftIDs                   []string `json:"job_schedule_shift_ids,omitempty"`
	LineupDispatchIDs                     []string `json:"lineup_dispatch_ids,omitempty"`
	TruckerAssignmentStatistics           any      `json:"trucker_assignment_statistics,omitempty"`
}

func newLineupsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup details",
		Long: `Show the full details of a lineup.

Output Fields:
  ID, name, start time range, customer ID
  Related job production plans, job schedule shifts, and lineup dispatches

Arguments:
  <id>    The lineup ID (required). Use the list command to find IDs.

Flags:
  --include-trucker-assignment-statistics  Include trucker assignment statistics meta

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a lineup
  xbe view lineups show 123

  # Include trucker assignment statistics
  xbe view lineups show 123 --include-trucker-assignment-statistics

  # JSON output
  xbe view lineups show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupsShow,
	}
	initLineupsShowFlags(cmd)
	return cmd
}

func init() {
	lineupsCmd.AddCommand(newLineupsShowCmd())
}

func initLineupsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("include-trucker-assignment-statistics", false, "Include trucker assignment statistics meta")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("lineup id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	if opts.IncludeTruckerAssignmentStatistics {
		query.Set("meta[lineup]", "trucker-assignment-statistics")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/lineups/"+id, query)
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

	details := buildLineupDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupDetails(cmd, details)
}

func parseLineupsShowOptions(cmd *cobra.Command) (lineupsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	includeStats, _ := cmd.Flags().GetBool("include-trucker-assignment-statistics")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupsShowOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		NoAuth:                             noAuth,
		IncludeTruckerAssignmentStatistics: includeStats,
	}, nil
}

func buildLineupDetails(resp jsonAPISingleResponse) lineupDetails {
	resource := resp.Data
	attrs := resource.Attributes
	meta := resource.Meta

	details := lineupDetails{
		ID:         resource.ID,
		Name:       stringAttr(attrs, "name"),
		StartAtMin: formatDateTime(stringAttr(attrs, "start-at-min")),
		StartAtMax: formatDateTime(stringAttr(attrs, "start-at-max")),
	}

	details.CustomerID = relationshipIDFromMap(resource.Relationships, "customer")
	details.LineupJobProductionPlanIDs = relationshipIDsFromMap(resource.Relationships, "lineup-job-production-plans")
	details.AliveLineupJobProductionPlanIDs = relationshipIDsFromMap(resource.Relationships, "alive-lineup-job-production-plans")
	details.UnabandonedLineupJobProductionPlanIDs = relationshipIDsFromMap(resource.Relationships, "unabandoned-lineup-job-production-plans")
	details.JobProductionPlanIDs = relationshipIDsFromMap(resource.Relationships, "job-production-plans")
	details.AliveJobProductionPlanIDs = relationshipIDsFromMap(resource.Relationships, "alive-job-production-plans")
	details.UnabandonedJobProductionPlanIDs = relationshipIDsFromMap(resource.Relationships, "unabandoned-job-production-plans")
	details.LineupJobScheduleShiftIDs = relationshipIDsFromMap(resource.Relationships, "lineup-job-schedule-shifts")
	details.JobScheduleShiftIDs = relationshipIDsFromMap(resource.Relationships, "job-schedule-shifts")
	details.LineupDispatchIDs = relationshipIDsFromMap(resource.Relationships, "lineup-dispatches")

	if meta != nil {
		if stats, ok := meta["trucker_assignment_statistics"]; ok {
			details.TruckerAssignmentStatistics = stats
		} else if stats, ok := meta["trucker-assignment-statistics"]; ok {
			details.TruckerAssignmentStatistics = stats
		}
	}

	return details
}

func renderLineupDetails(cmd *cobra.Command, details lineupDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.StartAtMin != "" {
		fmt.Fprintf(out, "Start At Min: %s\n", details.StartAtMin)
	}
	if details.StartAtMax != "" {
		fmt.Fprintf(out, "Start At Max: %s\n", details.StartAtMax)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if len(details.LineupJobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Lineup Job Production Plan IDs: %s\n", strings.Join(details.LineupJobProductionPlanIDs, ", "))
	}
	if len(details.AliveLineupJobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Alive Lineup Job Production Plan IDs: %s\n", strings.Join(details.AliveLineupJobProductionPlanIDs, ", "))
	}
	if len(details.UnabandonedLineupJobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Unabandoned Lineup Job Production Plan IDs: %s\n", strings.Join(details.UnabandonedLineupJobProductionPlanIDs, ", "))
	}
	if len(details.JobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan IDs: %s\n", strings.Join(details.JobProductionPlanIDs, ", "))
	}
	if len(details.AliveJobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Alive Job Production Plan IDs: %s\n", strings.Join(details.AliveJobProductionPlanIDs, ", "))
	}
	if len(details.UnabandonedJobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Unabandoned Job Production Plan IDs: %s\n", strings.Join(details.UnabandonedJobProductionPlanIDs, ", "))
	}
	if len(details.LineupJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Lineup Job Schedule Shift IDs: %s\n", strings.Join(details.LineupJobScheduleShiftIDs, ", "))
	}
	if len(details.JobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Job Schedule Shift IDs: %s\n", strings.Join(details.JobScheduleShiftIDs, ", "))
	}
	if len(details.LineupDispatchIDs) > 0 {
		fmt.Fprintf(out, "Lineup Dispatch IDs: %s\n", strings.Join(details.LineupDispatchIDs, ", "))
	}
	if details.TruckerAssignmentStatistics != nil {
		payload, err := json.MarshalIndent(details.TruckerAssignmentStatistics, "", "  ")
		if err == nil {
			fmt.Fprintln(out, "Trucker Assignment Statistics:")
			fmt.Fprintln(out, string(payload))
		}
	}

	return nil
}
