package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type lineupScenarioGeneratorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioGeneratorDetails struct {
	ID                                        string   `json:"id"`
	BrokerID                                  string   `json:"broker_id,omitempty"`
	CustomerID                                string   `json:"customer_id,omitempty"`
	Date                                      string   `json:"date,omitempty"`
	Window                                    string   `json:"window,omitempty"`
	CompletedAt                               string   `json:"completed_at,omitempty"`
	IncludeTruckerAssignmentsAsConstraints    bool     `json:"include_trucker_assignments_as_constraints"`
	TruckerAssignmentLimitsLookbackWindowDays int      `json:"trucker_assignment_limits_lookback_window_days,omitempty"`
	SkipMinimumAssignmentCount                bool     `json:"skip_minimum_assignment_count"`
	SkipCreateLineupScenarioSolution          bool     `json:"skip_create_lineup_scenario_solution"`
	UseMostRecentLineupScenarioConstraints    bool     `json:"use_most_recent_lineup_scenario_constraints"`
	LineupScenarioIDs                         []string `json:"lineup_scenario_ids,omitempty"`
}

func newLineupScenarioGeneratorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario generator details",
		Long: `Show the full details of a lineup scenario generator.

Output Fields:
  ID
  Broker
  Customer
  Date
  Window
  Completed At
  Include Trucker Assignments As Constraints
  Trucker Assignment Limits Lookback Window Days
  Skip Minimum Assignment Count
  Skip Create Lineup Scenario Solution
  Use Most Recent Lineup Scenario Constraints
  Lineup Scenario IDs

Arguments:
  <id>  The lineup scenario generator ID (required).`,
		Example: `  # Show a generator
  xbe view lineup-scenario-generators show 123

  # JSON output
  xbe view lineup-scenario-generators show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioGeneratorsShow,
	}
	initLineupScenarioGeneratorsShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioGeneratorsCmd.AddCommand(newLineupScenarioGeneratorsShowCmd())
}

func initLineupScenarioGeneratorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioGeneratorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupScenarioGeneratorsShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario generator id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-generators/"+id, nil)
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

	details := buildLineupScenarioGeneratorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioGeneratorDetails(cmd, details)
}

func parseLineupScenarioGeneratorsShowOptions(cmd *cobra.Command) (lineupScenarioGeneratorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioGeneratorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioGeneratorDetails(resp jsonAPISingleResponse) lineupScenarioGeneratorDetails {
	attrs := resp.Data.Attributes
	details := lineupScenarioGeneratorDetails{
		ID:                                     resp.Data.ID,
		Date:                                   formatDate(stringAttr(attrs, "date")),
		Window:                                 stringAttr(attrs, "window"),
		CompletedAt:                            formatDateTime(stringAttr(attrs, "completed-at")),
		IncludeTruckerAssignmentsAsConstraints: boolAttr(attrs, "include-trucker-assignments-as-constraints"),
		TruckerAssignmentLimitsLookbackWindowDays: intAttr(attrs, "trucker-assignment-limits-lookback-window-days"),
		SkipMinimumAssignmentCount:                boolAttr(attrs, "skip-minimum-assignment-count"),
		SkipCreateLineupScenarioSolution:          boolAttr(attrs, "skip-create-lineup-scenario-solution"),
		UseMostRecentLineupScenarioConstraints:    boolAttr(attrs, "use-most-recent-lineup-scenario-constraints"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["lineup-scenarios"]; ok {
		details.LineupScenarioIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderLineupScenarioGeneratorDetails(cmd *cobra.Command, details lineupScenarioGeneratorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer: %s\n", details.CustomerID)
	}
	if details.Date != "" {
		fmt.Fprintf(out, "Date: %s\n", details.Date)
	}
	if details.Window != "" {
		fmt.Fprintf(out, "Window: %s\n", details.Window)
	}
	if details.CompletedAt != "" {
		fmt.Fprintf(out, "Completed At: %s\n", details.CompletedAt)
	}
	fmt.Fprintf(out, "Include Trucker Assignments As Constraints: %s\n", formatBool(details.IncludeTruckerAssignmentsAsConstraints))
	fmt.Fprintf(out, "Trucker Assignment Limits Lookback Window Days: %d\n", details.TruckerAssignmentLimitsLookbackWindowDays)
	fmt.Fprintf(out, "Skip Minimum Assignment Count: %s\n", formatBool(details.SkipMinimumAssignmentCount))
	fmt.Fprintf(out, "Skip Create Lineup Scenario Solution: %s\n", formatBool(details.SkipCreateLineupScenarioSolution))
	fmt.Fprintf(out, "Use Most Recent Lineup Scenario Constraints: %s\n", formatBool(details.UseMostRecentLineupScenarioConstraints))
	if len(details.LineupScenarioIDs) > 0 {
		fmt.Fprintf(out, "Lineup Scenario IDs: %s\n", strings.Join(details.LineupScenarioIDs, ", "))
	}

	return nil
}
