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

type lineupScenarioTruckersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioTruckerDetails struct {
	ID                               string `json:"id"`
	LineupScenarioID                 string `json:"lineup_scenario_id,omitempty"`
	LineupScenarioName               string `json:"lineup_scenario_name,omitempty"`
	LineupScenarioDate               string `json:"lineup_scenario_date,omitempty"`
	LineupScenarioWindow             string `json:"lineup_scenario_window,omitempty"`
	TruckerID                        string `json:"trucker_id,omitempty"`
	TruckerName                      string `json:"trucker_name,omitempty"`
	MinimumAssignmentCount           string `json:"minimum_assignment_count,omitempty"`
	MaximumAssignmentCount           string `json:"maximum_assignment_count,omitempty"`
	MaximumMinutesToStartSite        string `json:"maximum_minutes_to_start_site,omitempty"`
	MaterialTypeConstraints          any    `json:"material_type_constraints,omitempty"`
	TrailerClassificationConstraints any    `json:"trailer_classification_constraints,omitempty"`
}

func newLineupScenarioTruckersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario trucker details",
		Long: `Show the full details of a lineup scenario trucker.

Output Fields:
  ID
  Lineup Scenario (name + ID)
  Lineup Scenario Date / Window
  Trucker (name + ID)
  Minimum Assignment Count / Maximum Assignment Count
  Maximum Minutes To Start Site
  Material Type Constraints
  Trailer Classification Constraints

Arguments:
  <id>  The lineup scenario trucker ID (required).`,
		Example: `  # Show lineup scenario trucker details
  xbe view lineup-scenario-truckers show 123

  # Output as JSON
  xbe view lineup-scenario-truckers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioTruckersShow,
	}
	initLineupScenarioTruckersShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioTruckersCmd.AddCommand(newLineupScenarioTruckersShowCmd())
}

func initLineupScenarioTruckersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioTruckersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupScenarioTruckersShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario trucker id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineup-scenario-truckers]", "lineup-scenario,trucker,minimum-assignment-count,maximum-assignment-count,maximum-minutes-to-start-site,material-type-constraints,trailer-classification-constraints")
	query.Set("include", "lineup-scenario,trucker")
	query.Set("fields[lineup-scenarios]", "name,date,window")
	query.Set("fields[truckers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-truckers/"+id, query)
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

	details := buildLineupScenarioTruckerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioTruckerDetails(cmd, details)
}

func parseLineupScenarioTruckersShowOptions(cmd *cobra.Command) (lineupScenarioTruckersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioTruckersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioTruckerDetails(resp jsonAPISingleResponse) lineupScenarioTruckerDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := lineupScenarioTruckerDetails{
		ID:                               resp.Data.ID,
		MinimumAssignmentCount:           stringAttr(resp.Data.Attributes, "minimum-assignment-count"),
		MaximumAssignmentCount:           stringAttr(resp.Data.Attributes, "maximum-assignment-count"),
		MaximumMinutesToStartSite:        stringAttr(resp.Data.Attributes, "maximum-minutes-to-start-site"),
		MaterialTypeConstraints:          anyAttr(resp.Data.Attributes, "material-type-constraints"),
		TrailerClassificationConstraints: anyAttr(resp.Data.Attributes, "trailer-classification-constraints"),
	}

	if rel, ok := resp.Data.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		details.LineupScenarioID = rel.Data.ID
		if scenario, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LineupScenarioName = stringAttr(scenario.Attributes, "name")
			details.LineupScenarioDate = stringAttr(scenario.Attributes, "date")
			details.LineupScenarioWindow = stringAttr(scenario.Attributes, "window")
		}
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = stringAttr(trucker.Attributes, "company-name")
		}
	}

	return details
}

func renderLineupScenarioTruckerDetails(cmd *cobra.Command, details lineupScenarioTruckerDetails) error {
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
	if details.TruckerID != "" || details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", formatRelated(details.TruckerName, details.TruckerID))
	}
	if details.MinimumAssignmentCount != "" {
		fmt.Fprintf(out, "Minimum Assignment Count: %s\n", details.MinimumAssignmentCount)
	}
	if details.MaximumAssignmentCount != "" {
		fmt.Fprintf(out, "Maximum Assignment Count: %s\n", details.MaximumAssignmentCount)
	}
	if details.MaximumMinutesToStartSite != "" {
		fmt.Fprintf(out, "Maximum Minutes To Start Site: %s\n", details.MaximumMinutesToStartSite)
	}

	if details.MaterialTypeConstraints != nil {
		fmt.Fprintf(out, "Material Type Constraints: %d\n", countConstraintItems(details.MaterialTypeConstraints))
		if formatted := formatAnyJSON(details.MaterialTypeConstraints); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Material Type Constraint Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	if details.TrailerClassificationConstraints != nil {
		fmt.Fprintf(out, "Trailer Classification Constraints: %d\n", countConstraintItems(details.TrailerClassificationConstraints))
		if formatted := formatAnyJSON(details.TrailerClassificationConstraints); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Trailer Classification Constraint Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}

func anyAttr(attrs map[string]any, key string) any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	return value
}

func countConstraintItems(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}

func formatAnyJSON(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
