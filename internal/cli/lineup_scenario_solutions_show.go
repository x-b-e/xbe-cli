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

type lineupScenarioSolutionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioSolutionDetails struct {
	ID                        string   `json:"id"`
	LineupScenarioID          string   `json:"lineup_scenario_id,omitempty"`
	Status                    string   `json:"status,omitempty"`
	Cost                      *float64 `json:"cost,omitempty"`
	SolvedAt                  string   `json:"solved_at,omitempty"`
	Assignments               any      `json:"assignments,omitempty"`
	TruckerAssignmentsSummary any      `json:"trucker_assignments_summary,omitempty"`
}

func newLineupScenarioSolutionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario solution details",
		Long: `Show the full details of a lineup scenario solution.

Output Fields:
  ID
  Lineup Scenario ID
  Status
  Cost
  Solved At
  Assignments
  Trucker Assignments Summary

Arguments:
  <id>    The lineup scenario solution ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a lineup scenario solution
  xbe view lineup-scenario-solutions show 123

  # Output as JSON
  xbe view lineup-scenario-solutions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioSolutionsShow,
	}
	initLineupScenarioSolutionsShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioSolutionsCmd.AddCommand(newLineupScenarioSolutionsShowCmd())
}

func initLineupScenarioSolutionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioSolutionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupScenarioSolutionsShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario solution id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineup-scenario-solutions]", "status,cost,solved-at,assignments,trucker-assignments-summary,lineup-scenario")
	query.Set("include", "lineup-scenario")

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-solutions/"+id, query)
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

	details := buildLineupScenarioSolutionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioSolutionDetails(cmd, details)
}

func parseLineupScenarioSolutionsShowOptions(cmd *cobra.Command) (lineupScenarioSolutionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioSolutionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioSolutionDetails(resp jsonAPISingleResponse) lineupScenarioSolutionDetails {
	attrs := resp.Data.Attributes

	details := lineupScenarioSolutionDetails{
		ID:       resp.Data.ID,
		Status:   stringAttr(attrs, "status"),
		SolvedAt: stringAttr(attrs, "solved-at"),
	}

	if value, ok := attrs["cost"]; ok && value != nil {
		cost := floatAttr(attrs, "cost")
		details.Cost = &cost
	}
	if value, ok := attrs["assignments"]; ok && value != nil {
		details.Assignments = value
	}
	if value, ok := attrs["trucker-assignments-summary"]; ok && value != nil {
		details.TruckerAssignmentsSummary = value
	}

	if rel, ok := resp.Data.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		details.LineupScenarioID = rel.Data.ID
	}

	return details
}

func renderLineupScenarioSolutionDetails(cmd *cobra.Command, details lineupScenarioSolutionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupScenarioID != "" {
		fmt.Fprintf(out, "Lineup Scenario ID: %s\n", details.LineupScenarioID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Cost != nil {
		fmt.Fprintf(out, "Cost: %.2f\n", *details.Cost)
	}
	if details.SolvedAt != "" {
		fmt.Fprintf(out, "Solved At: %s\n", details.SolvedAt)
	}

	if details.Assignments != nil {
		assignments := formatJSONValue(details.Assignments)
		if assignments != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Assignments:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, assignments)
		}
	}

	if details.TruckerAssignmentsSummary != nil {
		summary := formatJSONValue(details.TruckerAssignmentsSummary)
		if summary != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Trucker Assignments Summary:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, summary)
		}
	}

	return nil
}
