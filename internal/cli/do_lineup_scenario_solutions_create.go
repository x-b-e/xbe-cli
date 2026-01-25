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

type doLineupScenarioSolutionsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	LineupScenario string
}

func newDoLineupScenarioSolutionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario solution",
		Long: `Create a lineup scenario solution.

Creating a solution triggers the lineup solver asynchronously. The solution
status will start as unsolved until the solver finishes.

Required flags:
  --lineup-scenario   Lineup scenario ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Solve a lineup scenario
  xbe do lineup-scenario-solutions create --lineup-scenario 123

  # JSON output
  xbe do lineup-scenario-solutions create --lineup-scenario 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupScenarioSolutionsCreate,
	}
	initDoLineupScenarioSolutionsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioSolutionsCmd.AddCommand(newDoLineupScenarioSolutionsCreateCmd())
}

func initDoLineupScenarioSolutionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-scenario", "", "Lineup scenario ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioSolutionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioSolutionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.LineupScenario) == "" {
		err := fmt.Errorf("--lineup-scenario is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"lineup-scenario": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenarios",
				"id":   opts.LineupScenario,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-scenario-solutions",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-solutions", jsonBody)
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

	row := buildLineupScenarioSolutionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario solution %s\n", row.ID)
	return nil
}

func parseDoLineupScenarioSolutionsCreateOptions(cmd *cobra.Command) (doLineupScenarioSolutionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupScenario, _ := cmd.Flags().GetString("lineup-scenario")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioSolutionsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		LineupScenario: lineupScenario,
	}, nil
}
