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

type lineupScenarioLineupsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioLineupDetails struct {
	ID                   string `json:"id"`
	LineupScenarioID     string `json:"lineup_scenario_id,omitempty"`
	LineupScenarioName   string `json:"lineup_scenario_name,omitempty"`
	LineupScenarioDate   string `json:"lineup_scenario_date,omitempty"`
	LineupScenarioWindow string `json:"lineup_scenario_window,omitempty"`
	LineupID             string `json:"lineup_id,omitempty"`
	LineupName           string `json:"lineup_name,omitempty"`
	LineupStartAtMin     string `json:"lineup_start_at_min,omitempty"`
	LineupStartAtMax     string `json:"lineup_start_at_max,omitempty"`
}

func newLineupScenarioLineupsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario lineup details",
		Long: `Show the full details of a lineup scenario lineup.

Output Fields:
  ID                    Resource identifier
  Lineup Scenario        Lineup scenario name or ID
  Lineup Scenario Date   Scenario date
  Lineup Scenario Window Scenario window
  Lineup                Lineup name or ID
  Lineup Start At Min    Lineup start window minimum
  Lineup Start At Max    Lineup start window maximum

Arguments:
  <id>  The lineup scenario lineup ID (required).`,
		Example: `  # Show lineup scenario lineup details
  xbe view lineup-scenario-lineups show 123

  # Output as JSON
  xbe view lineup-scenario-lineups show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioLineupsShow,
	}
	initLineupScenarioLineupsShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioLineupsCmd.AddCommand(newLineupScenarioLineupsShowCmd())
}

func initLineupScenarioLineupsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioLineupsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupScenarioLineupsShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario lineup id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineup-scenario-lineups]", "lineup-scenario,lineup")
	query.Set("include", "lineup-scenario,lineup")
	query.Set("fields[lineup-scenarios]", "name,date,window")
	query.Set("fields[lineups]", "name,start-at-min,start-at-max")

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-lineups/"+id, query)
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

	details := buildLineupScenarioLineupDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioLineupDetails(cmd, details)
}

func parseLineupScenarioLineupsShowOptions(cmd *cobra.Command) (lineupScenarioLineupsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioLineupsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioLineupDetails(resp jsonAPISingleResponse) lineupScenarioLineupDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := lineupScenarioLineupDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		details.LineupScenarioID = rel.Data.ID
		if scenario, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LineupScenarioName = stringAttr(scenario.Attributes, "name")
			details.LineupScenarioDate = stringAttr(scenario.Attributes, "date")
			details.LineupScenarioWindow = stringAttr(scenario.Attributes, "window")
		}
	}

	if rel, ok := resp.Data.Relationships["lineup"]; ok && rel.Data != nil {
		details.LineupID = rel.Data.ID
		if lineup, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LineupName = stringAttr(lineup.Attributes, "name")
			details.LineupStartAtMin = stringAttr(lineup.Attributes, "start-at-min")
			details.LineupStartAtMax = stringAttr(lineup.Attributes, "start-at-max")
		}
	}

	return details
}

func renderLineupScenarioLineupDetails(cmd *cobra.Command, details lineupScenarioLineupDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupScenarioID != "" || details.LineupScenarioName != "" {
		fmt.Fprintf(out, "Lineup Scenario: %s\n", formatRelated(details.LineupScenarioName, details.LineupScenarioID))
	}
	if details.LineupScenarioDate != "" {
		fmt.Fprintf(out, "Lineup Scenario Date: %s\n", details.LineupScenarioDate)
	}
	if details.LineupScenarioWindow != "" {
		fmt.Fprintf(out, "Lineup Scenario Window: %s\n", details.LineupScenarioWindow)
	}
	if details.LineupID != "" || details.LineupName != "" {
		fmt.Fprintf(out, "Lineup: %s\n", formatRelated(details.LineupName, details.LineupID))
	}
	if details.LineupStartAtMin != "" {
		fmt.Fprintf(out, "Lineup Start At Min: %s\n", details.LineupStartAtMin)
	}
	if details.LineupStartAtMax != "" {
		fmt.Fprintf(out, "Lineup Start At Max: %s\n", details.LineupStartAtMax)
	}

	return nil
}
