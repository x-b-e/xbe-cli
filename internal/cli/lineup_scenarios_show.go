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

type lineupScenariosShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioDetails struct {
	ID                                      string   `json:"id"`
	Name                                    string   `json:"name,omitempty"`
	BrokerID                                string   `json:"broker_id,omitempty"`
	CustomerID                              string   `json:"customer_id,omitempty"`
	Date                                    string   `json:"date,omitempty"`
	Window                                  string   `json:"window,omitempty"`
	IncludeTruckerAssignmentsAsConstraints  bool     `json:"include_trucker_assignments_as_constraints"`
	AddLineupsAutomatically                 bool     `json:"add_lineups_automatically"`
	GeneratorID                             string   `json:"generator_id,omitempty"`
	LineupScenarioLineupIDs                 []string `json:"lineup_scenario_lineup_ids,omitempty"`
	LineupIDs                               []string `json:"lineup_ids,omitempty"`
	LineupScenarioTruckerIDs                []string `json:"lineup_scenario_trucker_ids,omitempty"`
	TruckerIDs                              []string `json:"trucker_ids,omitempty"`
	LineupScenarioLineupJobScheduleShiftIDs []string `json:"lineup_scenario_lineup_job_schedule_shift_ids,omitempty"`
	LineupJobScheduleShiftIDs               []string `json:"lineup_job_schedule_shift_ids,omitempty"`
	LineupScenarioSolutionIDs               []string `json:"lineup_scenario_solution_ids,omitempty"`
}

func newLineupScenariosShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario details",
		Long: `Show the full details of a lineup scenario.

Output Fields:
  ID
  Name
  Broker
  Customer
  Date
  Window
  Include Trucker Assignments As Constraints
  Add Lineups Automatically
  Generator
  Lineup Scenario Lineup IDs
  Lineup IDs
  Lineup Scenario Trucker IDs
  Trucker IDs
  Lineup Scenario Lineup Job Schedule Shift IDs
  Lineup Job Schedule Shift IDs
  Lineup Scenario Solution IDs

Arguments:
  <id>  The lineup scenario ID (required).`,
		Example: `  # Show a lineup scenario
  xbe view lineup-scenarios show 123

  # JSON output
  xbe view lineup-scenarios show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenariosShow,
	}
	initLineupScenariosShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenariosCmd.AddCommand(newLineupScenariosShowCmd())
}

func initLineupScenariosShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenariosShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupScenariosShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenarios/"+id, nil)
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

	details := buildLineupScenarioDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioDetails(cmd, details)
}

func parseLineupScenariosShowOptions(cmd *cobra.Command) (lineupScenariosShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenariosShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioDetails(resp jsonAPISingleResponse) lineupScenarioDetails {
	attrs := resp.Data.Attributes
	details := lineupScenarioDetails{
		ID:                                     resp.Data.ID,
		Name:                                   strings.TrimSpace(stringAttr(attrs, "name")),
		Date:                                   formatDate(stringAttr(attrs, "date")),
		Window:                                 stringAttr(attrs, "window"),
		IncludeTruckerAssignmentsAsConstraints: boolAttr(attrs, "include-trucker-assignments-as-constraints"),
		AddLineupsAutomatically:                boolAttr(attrs, "add-lineups-automatically"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["generator"]; ok && rel.Data != nil {
		details.GeneratorID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["lineup-scenario-lineups"]; ok {
		details.LineupScenarioLineupIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["lineups"]; ok {
		details.LineupIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["lineup-scenario-truckers"]; ok {
		details.LineupScenarioTruckerIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["truckers"]; ok {
		details.TruckerIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["lineup-scenario-lineup-job-schedule-shifts"]; ok {
		details.LineupScenarioLineupJobScheduleShiftIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["lineup-job-schedule-shifts"]; ok {
		details.LineupJobScheduleShiftIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["lineup-scenario-solutions"]; ok {
		details.LineupScenarioSolutionIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderLineupScenarioDetails(cmd *cobra.Command, details lineupScenarioDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
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
	fmt.Fprintf(out, "Include Trucker Assignments As Constraints: %s\n", formatBool(details.IncludeTruckerAssignmentsAsConstraints))
	fmt.Fprintf(out, "Add Lineups Automatically: %s\n", formatBool(details.AddLineupsAutomatically))
	if details.GeneratorID != "" {
		fmt.Fprintf(out, "Generator: %s\n", details.GeneratorID)
	}
	if len(details.LineupScenarioLineupIDs) > 0 {
		fmt.Fprintf(out, "Lineup Scenario Lineup IDs: %s\n", strings.Join(details.LineupScenarioLineupIDs, ", "))
	}
	if len(details.LineupIDs) > 0 {
		fmt.Fprintf(out, "Lineup IDs: %s\n", strings.Join(details.LineupIDs, ", "))
	}
	if len(details.LineupScenarioTruckerIDs) > 0 {
		fmt.Fprintf(out, "Lineup Scenario Trucker IDs: %s\n", strings.Join(details.LineupScenarioTruckerIDs, ", "))
	}
	if len(details.TruckerIDs) > 0 {
		fmt.Fprintf(out, "Trucker IDs: %s\n", strings.Join(details.TruckerIDs, ", "))
	}
	if len(details.LineupScenarioLineupJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Lineup Scenario Lineup Job Schedule Shift IDs: %s\n", strings.Join(details.LineupScenarioLineupJobScheduleShiftIDs, ", "))
	}
	if len(details.LineupJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Lineup Job Schedule Shift IDs: %s\n", strings.Join(details.LineupJobScheduleShiftIDs, ", "))
	}
	if len(details.LineupScenarioSolutionIDs) > 0 {
		fmt.Fprintf(out, "Lineup Scenario Solution IDs: %s\n", strings.Join(details.LineupScenarioSolutionIDs, ", "))
	}

	return nil
}
