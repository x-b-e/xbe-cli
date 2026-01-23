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

type doLineupScenarioTruckersCreateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	LineupScenario                   string
	Trucker                          string
	MinimumAssignmentCount           int
	MaximumAssignmentCount           int
	MaximumMinutesToStartSite        int
	MaterialTypeConstraints          string
	TrailerClassificationConstraints string
}

func newDoLineupScenarioTruckersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario trucker",
		Long: `Create a lineup scenario trucker.

Required flags:
  --lineup-scenario  Lineup scenario ID (required)
  --trucker          Trucker ID (required)

Optional flags:
  --minimum-assignment-count         Minimum assignments for the trucker
  --maximum-assignment-count         Maximum assignments for the trucker
  --maximum-minutes-to-start-site    Maximum minutes to the start site
  --material-type-constraints        JSON array of constraints (snake_case keys)
  --trailer-classification-constraints  JSON array of constraints (snake_case keys)

Notes:
  - The trucker must belong to the same broker as the lineup scenario.
  - minimum-assignment-count must be less than or equal to maximum-assignment-count when both are set.`,
		Example: `  # Create a lineup scenario trucker
  xbe do lineup-scenario-truckers create --lineup-scenario 123 --trucker 456

  # Create with assignment limits
  xbe do lineup-scenario-truckers create --lineup-scenario 123 --trucker 456 \\
    --minimum-assignment-count 1 --maximum-assignment-count 3 --maximum-minutes-to-start-site 45

  # Create with constraints
  xbe do lineup-scenario-truckers create --lineup-scenario 123 --trucker 456 \\
    --material-type-constraints '[{\"material_type_base_qualified_name\":\"soil\",\"maximum_assignment_count\":2}]' \\
    --trailer-classification-constraints '[{\"trailer_classification_id\":\"789\",\"minimum_assignment_count\":1,\"maximum_assignment_count\":2}]'

  # Output as JSON
  xbe do lineup-scenario-truckers create --lineup-scenario 123 --trucker 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupScenarioTruckersCreate,
	}
	initDoLineupScenarioTruckersCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTruckersCmd.AddCommand(newDoLineupScenarioTruckersCreateCmd())
}

func initDoLineupScenarioTruckersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-scenario", "", "Lineup scenario ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().Int("minimum-assignment-count", 0, "Minimum assignments for the trucker")
	cmd.Flags().Int("maximum-assignment-count", 0, "Maximum assignments for the trucker")
	cmd.Flags().Int("maximum-minutes-to-start-site", 0, "Maximum minutes to the start site")
	cmd.Flags().String("material-type-constraints", "", "JSON array of material type constraints")
	cmd.Flags().String("trailer-classification-constraints", "", "JSON array of trailer classification constraints")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioTruckersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioTruckersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if opts.LineupScenario == "" {
		err := fmt.Errorf("--lineup-scenario is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Trucker == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("minimum-assignment-count") {
		attributes["minimum-assignment-count"] = opts.MinimumAssignmentCount
	}
	if cmd.Flags().Changed("maximum-assignment-count") {
		attributes["maximum-assignment-count"] = opts.MaximumAssignmentCount
	}
	if cmd.Flags().Changed("maximum-minutes-to-start-site") {
		attributes["maximum-minutes-to-start-site"] = opts.MaximumMinutesToStartSite
	}
	if cmd.Flags().Changed("material-type-constraints") {
		constraints, err := parseConstraintArray(opts.MaterialTypeConstraints, "material-type-constraints")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["material-type-constraints"] = constraints
	}
	if cmd.Flags().Changed("trailer-classification-constraints") {
		constraints, err := parseConstraintArray(opts.TrailerClassificationConstraints, "trailer-classification-constraints")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["trailer-classification-constraints"] = constraints
	}

	relationships := map[string]any{
		"lineup-scenario": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenarios",
				"id":   opts.LineupScenario,
			},
		},
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-scenario-truckers",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-truckers", jsonBody)
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

	row := lineupScenarioTruckerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario trucker %s\n", row.ID)
	return nil
}

func parseDoLineupScenarioTruckersCreateOptions(cmd *cobra.Command) (doLineupScenarioTruckersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupScenario, _ := cmd.Flags().GetString("lineup-scenario")
	trucker, _ := cmd.Flags().GetString("trucker")
	minimumAssignmentCount, _ := cmd.Flags().GetInt("minimum-assignment-count")
	maximumAssignmentCount, _ := cmd.Flags().GetInt("maximum-assignment-count")
	maximumMinutesToStartSite, _ := cmd.Flags().GetInt("maximum-minutes-to-start-site")
	materialTypeConstraints, _ := cmd.Flags().GetString("material-type-constraints")
	trailerClassificationConstraints, _ := cmd.Flags().GetString("trailer-classification-constraints")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTruckersCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		LineupScenario:                   lineupScenario,
		Trucker:                          trucker,
		MinimumAssignmentCount:           minimumAssignmentCount,
		MaximumAssignmentCount:           maximumAssignmentCount,
		MaximumMinutesToStartSite:        maximumMinutesToStartSite,
		MaterialTypeConstraints:          materialTypeConstraints,
		TrailerClassificationConstraints: trailerClassificationConstraints,
	}, nil
}

func parseConstraintArray(raw string, label string) ([]map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("%s requires a JSON array (use [] to clear)", label)
	}
	var data []map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("invalid %s JSON: %w", label, err)
	}
	return data, nil
}
